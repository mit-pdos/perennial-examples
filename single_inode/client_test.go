package single_inode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"
)

func mkBlock(b0 byte) disk.Block {
	b := make(disk.Block, disk.BlockSize)
	b[0] = b0
	return b
}

func TestSingleInode(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(1000)
	i := Open(d, d.Size())
	i.Append(mkBlock(1))
	i.Append(mkBlock(2))
	assert.Nil(i.Read(2), "out-of-bound read")
	i.Append(mkBlock(2))
	assert.Equal(byte(2), i.Read(2)[0])

	i = Open(d, d.Size())
	assert.Equal(byte(1), i.Read(0)[0])
	i.Append(mkBlock(3))
	assert.Equal(byte(3), i.Read(3)[0])
}
