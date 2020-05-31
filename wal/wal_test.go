package wal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"
)

func mkUpdate(addr uint64, b0 byte) update {
	return update{addr: addr, b: mkBlock(b0)}
}

func TestLogBasic(t *testing.T) {
	d := disk.NewMemDisk(1000)
	log := Open(d)
	log.Write([]update{
		mkUpdate(2, 0),
		mkUpdate(3, 1),
	})
	log.Write([]update{
		mkUpdate(4, 2),
		mkUpdate(2, 3),
	})
	assert.Equal(t, byte(3), log.Read(2)[0])
}

func TestLogRecover(t *testing.T) {
	d := disk.NewMemDisk(1000)
	log := Open(d)
	log.Write([]update{
		mkUpdate(2, 0),
		mkUpdate(3, 1),
	})
	log.Write([]update{
		mkUpdate(4, 2),
		mkUpdate(2, 3),
	})

	log = Open(d)
	assert.Equal(t, byte(3), log.Read(2)[0])
}
