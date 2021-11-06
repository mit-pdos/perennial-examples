package dir

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/async_disk"
)

func makeBlock(x byte) async_disk.Block {
	b := make(async_disk.Block, async_disk.BlockSize)
	b[0] = x
	return b
}

func TestDirAppendRead(t *testing.T) {
	assert := assert.New(t)
	theDisk := async_disk.NewMemDisk(NumInodes + 100)
	dir := Open(theDisk, theDisk.Size())
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
	theDisk := async_disk.NewMemDisk(NumInodes + 3)
	dir := Open(theDisk, theDisk.Size())
	ok := dir.Append(1, makeBlock(1))
	assert.True(ok, "append should succeed")
	assert.True(dir.Append(1, makeBlock(2)),
		"append should succeed")

	dir = Open(theDisk, theDisk.Size())
	assert.True(dir.Append(2, makeBlock(3)),
		"append of last block should succeed")
	assert.Equal(makeBlock(1), dir.Read(1, 0))
	assert.Equal(makeBlock(2), dir.Read(1, 1))
	assert.Equal(makeBlock(3), dir.Read(2, 0))
}

func TestDirRecoverFull(t *testing.T) {
	assert := assert.New(t)
	theDisk := async_disk.NewMemDisk(NumInodes + 2)
	dir := Open(theDisk, theDisk.Size())
	dir.Append(1, makeBlock(1))
	dir.Append(1, makeBlock(2))

	dir = Open(theDisk, theDisk.Size())
	ok := dir.Append(2, makeBlock(3))
	assert.False(ok, "should be no space to add more blocks")
}
