package dynamic_dir

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
	theDisk := disk.NewMemDisk(MaxInodes + 100)
	dir := Open(theDisk, theDisk.Size())
	ino0, ok := dir.Create()
	assert.True(ok, "creating inode should succeed")
	ino1, _ := dir.Create()
	assert.Equal(uint64(0), dir.Size(ino1))
	dir.Append(ino1, makeBlock(1))
	ino2, _ := dir.Create()
	dir.Append(ino1, makeBlock(2))
	dir.Append(ino2, makeBlock(3))
	assert.Equal(uint64(2), dir.Size(ino1))
	assert.Equal(makeBlock(2), dir.Read(ino1, 1))
	assert.Equal(makeBlock(3), dir.Read(ino2, 0))
	assert.Equal(uint64(0), dir.Size(ino0))
}

func TestDirCreateDelete(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(MaxInodes + 100)
	dir := Open(theDisk, theDisk.Size())
	ino0, _ := dir.Create()
	ino1, _ := dir.Create()
	assert.Equal(uint64(0), dir.Size(ino1))
	dir.Append(ino1, makeBlock(1))
	ino2, _ := dir.Create()
	dir.Append(ino1, makeBlock(2))
	dir.Append(ino2, makeBlock(3))
	assert.Equal(uint64(2), dir.Size(ino1))
	assert.Equal(makeBlock(2), dir.Read(ino1, 1))
	dir.Delete(ino1)
	assert.Equal(makeBlock(3), dir.Read(ino2, 0))
	dir.Delete(ino2)
	assert.Equal(uint64(0), dir.Size(ino0))
	ino3, _ := dir.Create()
	dir.Append(ino3, makeBlock(1))
	assert.Equal(makeBlock(1), dir.Read(ino3, 0))
}

func TestDirRecover(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(MaxInodes + 3)
	dir := Open(theDisk, theDisk.Size())
	ino1, _ := dir.Create()
	ok := dir.Append(ino1, makeBlock(1))
	assert.True(ok, "append should succeed")
	assert.True(dir.Append(ino1, makeBlock(2)),
		"append should succeed")

	dir = Open(theDisk, theDisk.Size())
	ino2, ok := dir.Create()
	assert.True(ok, "create should succeed")
	assert.True(dir.Append(ino2, makeBlock(3)),
		"append of last block should succeed")
	assert.Equal(makeBlock(1), dir.Read(ino1, 0))
	assert.Equal(makeBlock(2), dir.Read(ino1, 1))
	assert.Equal(makeBlock(3), dir.Read(ino2, 0))
}

func TestDirRecoverFull(t *testing.T) {
	assert := assert.New(t)
	theDisk := disk.NewMemDisk(1 + 2 + 2)
	dir := Open(theDisk, theDisk.Size())
	ino1, _ := dir.Create()
	ino2, _ := dir.Create()
	dir.Append(ino1, makeBlock(1))
	dir.Append(ino1, makeBlock(2))

	dir = Open(theDisk, theDisk.Size())
	ok := dir.Append(ino2, makeBlock(3))
	assert.False(ok, "should be no space to add more blocks")
}
