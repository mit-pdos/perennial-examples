package dir

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"
)

func makeBlock(x byte) disk.Block {
	b := make(disk.Block, disk.BlockSize)
	b[0] = x
	return b
}

func TestDirAppendRead(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(100)
	dir := OpenDir(theDisk, theDisk.Size())
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
	dir := OpenDir(theDisk, theDisk.Size())
	ok := dir.Append(1, makeBlock(1))
	assert.True(ok, "append should succeed")
	dir.Append(1, makeBlock(2))

	dir = OpenDir(theDisk, theDisk.Size())
	dir.Append(2, makeBlock(3))
	assert.Equal(makeBlock(1), dir.Read(1, 0))
	assert.Equal(makeBlock(2), dir.Read(1, 1))
	assert.Equal(makeBlock(3), dir.Read(2, 0))
}

func TestDirRecoverFull(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(NumInodes + 2)
	dir := OpenDir(theDisk, theDisk.Size())
	dir.Append(1, makeBlock(1))
	dir.Append(1, makeBlock(2))

	dir = OpenDir(theDisk, theDisk.Size())
	ok := dir.Append(2, makeBlock(3))
	assert.False(ok, "should be no space to add more blocks")
}
