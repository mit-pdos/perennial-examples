package async_inode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"

	"github.com/mit-pdos/perennial-examples/alloc"
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
	assert.Equal(true, i.Append(makeBlock(1)),
		"should be enough space for append")
	i.Append(makeBlock(2))
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
}

func TestInodeAppendReadFlushSome(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	allocator := alloc.New(1, 9, alloc.AddrSet{})
	i := Open(d, 0)
	i.Append(makeBlock(1))
	i.Append(makeBlock(2))
	i.Flush(allocator)
	i.Append(makeBlock(3))
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
	assert.Equal(makeBlock(3), i.Read(2))
}

func TestInodeAppendFill(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(1000)
	ino := Open(d, 0)
	for i := uint64(0); i < MaxBlocks; i++ {
		assert.Equal(true,
			ino.Append(makeBlock(byte(i))),
			"should be enough space for InodeMaxBlocks")
	}
	assert.Equal(false,
		ino.Append(makeBlock(0)),
		"should not allow appending past InodeMaxBlocks")
}

func TestInodeRecover(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(10)
	allocator := alloc.New(1, 9, alloc.AddrSet{})
	i := Open(d, 0)
	i.Append(makeBlock(1))
	i.Append(makeBlock(2))
	i.Flush(allocator)
	i = Open(d, 0)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
	assert.Len(i.UsedBlocks(), 2)
}
