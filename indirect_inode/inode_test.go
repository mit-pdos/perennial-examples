package inode

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

func TestInodeAppendRead(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	i := Open(d, 0)
	d.Write(7, makeBlock(1))
	assert.Equal(AppendOk, i.Append(7), "should be enough space for append")
	d.Write(6, makeBlock(2))
	i.Append(6)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
}

func TestInodeAppendFill(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(MaxBlocks)
	ino := Open(d, 0)
	for i := uint64(0); i < MaxBlocks; i++ {
		res := ino.Append(1 + i)
		if res == AppendAgain {
			assert.Equal(AppendOk,
				ino.Append(1+i),
				"should be enough space for InodeMaxBlocks")
		} else {
			assert.Equal(AppendOk,
				res,
				"should be enough space for InodeMaxBlocks")

		}
	}
	assert.Equal(AppendFull,
		ino.Append(1+MaxBlocks),
		"should not allow appending past InodeMaxBlocks")
}

func TestInodeRecover(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	i := Open(d, 0)
	d.Write(7, makeBlock(1))
	i.Append(7)
	d.Write(6, makeBlock(2))
	i.Append(6)
	i = Open(d, 0)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
	assert.Equal([]uint64{7, 6}, i.UsedBlocks())
}
