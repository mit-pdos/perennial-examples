package replicated_block

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

func TestRepBlock(t *testing.T) {
	assert := assert.New(t)
	d := disk.NewMemDisk(2)
	rb := Open(d, 0)
	assert.Equal(byte(0), rb.Read(true)[0],
		"initial value should be zero block")
	assert.Equal(byte(0), rb.Read(false)[0])

	rb.Write(mkBlock(1))
	assert.Equal(byte(1), rb.Read(true)[0])
	assert.Equal(byte(1), rb.Read(false)[0])

	rb = Open(d, 0)
	assert.Equal(byte(1), rb.Read(true)[0],
		"after crash should have same value")
	assert.Equal(byte(1), rb.Read(false)[0])
	rb.Write(mkBlock(2))
	assert.Equal(byte(2), rb.Read(true)[0],
		"writes should work after recovery")
	assert.Equal(byte(2), rb.Read(false)[0])
}
