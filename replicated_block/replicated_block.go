package replicated_block

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
)

type RepBlock struct {
	// read-only
	d    disk.Disk
	addr uint64

	// protects disk addresses addr and addr+1
	m *sync.Mutex
}

// Open initializes a replicated block,
// either after a crash or from two disk blocks.
//
// Takes ownership of addr and addr+1 on disk.
func Open(d disk.Disk, addr uint64) *RepBlock {
	b := d.Read(addr)
	d.Write(addr+1, b)
	return &RepBlock{
		d:    d,
		addr: addr,
		m:    new(sync.Mutex),
	}
}

// readAddr returns the address to read from
//
// gives ownership of a disk block, so requires the lock to be held
func (rb *RepBlock) readAddr(primary bool) uint64 {
	if primary {
		return rb.addr
	} else {
		return rb.addr + 1
	}
}

func (rb *RepBlock) Read(primary bool) disk.Block {
	rb.m.Lock()
	b := rb.d.Read(rb.readAddr(primary))
	rb.m.Unlock()
	return b
}

func (rb *RepBlock) Write(b disk.Block) {
	rb.m.Lock()
	rb.d.Write(rb.addr, b)
	rb.d.Write(rb.addr+1, b)
	rb.m.Unlock()
}
