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
	i.Append(7)
	d.Write(6, makeBlock(2))
	i.Append(6)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
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
