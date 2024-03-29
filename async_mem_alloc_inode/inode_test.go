package async_mem_alloc_inode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/async_disk"

	"github.com/mit-pdos/perennial-examples/alloc"
)

func makeBlock(x byte) async_disk.Block {
	b := make(async_disk.Block, async_disk.BlockSize)
	b[0] = x
	return b
}

func TestInodeAppendRead(t *testing.T) {
	assert := assert.New(t)
	d := async_disk.NewMemDisk(10)
	allocator := alloc.New(3, 4, alloc.AddrSet{})
	i := Open(d, 0)
	assert.Equal(true, i.Append(makeBlock(1), allocator),
		"should be enough space for append")
	i.Append(makeBlock(2), allocator)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
}

func TestInodeAppendFill(t *testing.T) {
	assert := assert.New(t)
	d := async_disk.NewMemDisk(1000)
	allocator := alloc.New(1, 999, alloc.AddrSet{})
	ino := Open(d, 0)
	for i := uint64(0); i < MaxBlocks; i++ {
		assert.Equal(true,
			ino.Append(makeBlock(byte(i)), allocator),
			"should be enough space for InodeMaxBlocks")
	}
	assert.Equal(false,
		ino.Append(makeBlock(0), allocator),
		"should not allow appending past InodeMaxBlocks")
}

func TestInodeRecover(t *testing.T) {
	assert := assert.New(t)
	d := async_disk.NewMemDisk(10)
	allocator := alloc.New(1, 9, alloc.AddrSet{})
	i := Open(d, 0)
	i.Append(makeBlock(1), allocator)
	i.Append(makeBlock(2), allocator)
	i = Open(d, 0)
	assert.Equal(makeBlock(1), i.Read(0))
	assert.Equal(makeBlock(2), i.Read(1))
	assert.Len(i.UsedBlocks(), 2)
}
