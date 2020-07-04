package toy

import (
	"testing"

	"github.com/tchajed/goose/machine/disk"
)

func TestTransferBlock(t *testing.T) {
	d := disk.NewMemDisk(1)
	b := make(disk.Block, disk.BlockSize)
	b[0] = 2
	d.Write(0, b)
	TransferEvenBlock(d, 0)
	b = d.Read(0)
	if b[0]%2 != 0 {
		t.Errorf("expected even value on disk, got %v", b[0])
	}
}
