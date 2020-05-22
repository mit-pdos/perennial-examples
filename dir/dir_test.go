package dir

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"
)

func TestAllocatorReservationUnique(t *testing.T) {
	assert := assert.New(t)
	free := FreeRange(5, 10)
	alloc := newAllocator(free)
	a1, ok := alloc.Reserve()
	assert.GreaterOrEqual(a1, uint64(5), "allocated address %d should be in range", a1)
	assert.True(ok)
	a2, ok := alloc.Reserve()
	assert.True(ok)
	assert.NotEqual(a1, a2, "reserved same block twice")
}

func TestAllocatorAll(t *testing.T) {
	assert := assert.New(t)
	free := FreeRange(5, 10)
	alloc := newAllocator(free)
	for i := 0; i < 10; i++ {
		_, ok := alloc.Reserve()
		assert.True(ok, "reservation failed early: %d", i)
	}
	_, ok := alloc.Reserve()
	assert.False(ok, "all addresses should be allocatd")
}

func makeBlock(x byte) disk.Block {
	b := make(disk.Block, disk.BlockSize)
	b[0] = x
	return b
}

func TestInodeAppendRead(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	i := openInode(d, 0)
	d.Write(7, makeBlock(1))
	assert.True(i.Append(7), "should be enough space for append")
	d.Write(6, makeBlock(2))
	i.Append(6)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
}

func TestInodeAppendFill(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(1000)
	ino := openInode(d, 0)
	for i := uint64(0); i < InodeMaxBlocks; i++ {
		assert.True(ino.Append(1+i), "should be enough space for InodeMaxBlocks")
	}
	assert.False(ino.Append(1+InodeMaxBlocks),
		"should not allow appending past InodeMaxBlocks")
}

func TestInodeRecover(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	i := openInode(d, 0)
	d.Write(7, makeBlock(1))
	i.Append(7)
	d.Write(6, makeBlock(2))
	i.Append(6)
	i = openInode(d, 0)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
	assert.Equal([]uint64{7, 6}, i.UsedBlocks())
}

func TestDirAppendRead(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(100)
	// note that we supply [0,5) as blocks, but dir will correctly avoid
	// allocating them
	blocks := FreeRange(0, 100)
	dir := OpenDir(theDisk, blocks)
	assert.Equal(uint64(0), dir.Size(1))
	dir.Append(1, makeBlock(1))
	dir.Append(1, makeBlock(2))
	dir.Append(2, makeBlock(3))
	assert.Equal(uint64(2), dir.Size(1))
	assert.Equal(makeBlock(2), dir.Read(1, 1))
	assert.Equal(makeBlock(3), dir.Read(2, 0))
	assert.Equal(uint64(0), dir.Size(0))
}

func TestDirRecover(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(NumInodes + 3)
	dir := OpenDir(theDisk, FreeRange(0, theDisk.Size()))
	ok := dir.Append(1, makeBlock(1))
	assert.True(ok, "append should succeed")
	dir.Append(1, makeBlock(2))

	dir = OpenDir(theDisk, FreeRange(0, theDisk.Size()))
	dir.Append(2, makeBlock(3))
	assert.Equal(makeBlock(1), dir.Read(1, 0))
	assert.Equal(makeBlock(2), dir.Read(1, 1))
	assert.Equal(makeBlock(3), dir.Read(2, 0))
}

func TestDirRecoverFull(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(NumInodes + 2)
	dir := OpenDir(theDisk, FreeRange(0, theDisk.Size()))
	dir.Append(1, makeBlock(1))
	dir.Append(1, makeBlock(2))

	dir = OpenDir(theDisk, FreeRange(0, theDisk.Size()))
	ok := dir.Append(2, makeBlock(3))
	assert.False(ok, "should be no space to add more blocks")
}
