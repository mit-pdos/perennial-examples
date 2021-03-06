package async_inode

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"

	"github.com/mit-pdos/perennial-examples/alloc"
)

// Maximum size of inode, in blocks.
const MaxBlocks uint64 = 511

type Inode struct {
	// read-only
	d    disk.Disk
	m    *sync.Mutex
	addr uint64 // address on disk where inode is stored

	// mutable
	addrs    []uint64     // addresses of data blocks
	buffered []disk.Block // buffered data
}

func Open(d disk.Disk, addr uint64) *Inode {
	b := d.Read(addr)
	dec := marshal.NewDec(b)
	numAddrs := dec.GetInt()
	addrs := dec.GetInts(numAddrs)
	return &Inode{
		d:        d,
		m:        new(sync.Mutex),
		addr:     addr,
		addrs:    addrs,
		buffered: nil,
	}
}

// UsedBlocks returns the addresses allocated to the inode for the purposes
// of recovery. Assumes full ownership of the inode, so does not lock,
// and expects the caller to need only temporary access to the returned slice.
func (i *Inode) UsedBlocks() []uint64 {
	return i.addrs
}

func (i *Inode) read(off uint64) disk.Block {
	if off >= uint64(len(i.addrs))+uint64(len(i.buffered)) {
		return nil
	}
	if off < uint64(len(i.addrs)) {
		a := i.addrs[off]
		return i.d.Read(a)
	}
	return i.buffered[off-uint64(len(i.addrs))]
}

func (i *Inode) Read(off uint64) disk.Block {
	i.m.Lock()
	b := i.read(off)
	i.m.Unlock()
	return b
}

func (i *Inode) Size() uint64 {
	i.m.Lock()
	sz := uint64(len(i.addrs)) + uint64(len(i.buffered))
	i.m.Unlock()
	return sz
}

func (i *Inode) mkHdr() disk.Block {
	enc := marshal.NewEnc(disk.BlockSize)
	// buffered is not involved since they will be lost on crash
	enc.PutInt(uint64(len(i.addrs)))
	enc.PutInts(i.addrs)
	hdr := enc.Finish()
	return hdr
}

// appendOne durably extends the inode with the data in some address
func (i *Inode) appendOne(a uint64) {
	i.addrs = append(i.addrs, a)
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
}

// flushOne extends the on-disk inode with the next buffered write
//
// assumes lock is held and that there is at least one buffered write
func (i *Inode) flushOne(allocator *alloc.Allocator) bool {
	a, ok := allocator.Reserve()
	if !ok {
		return false
	}
	b := i.buffered[0]
	i.buffered = i.buffered[1:]
	i.d.Write(a, b)
	i.appendOne(a)
	return true
}

// critical section for Flush
//
// assumes lock is held
func (i *Inode) flush(allocator *alloc.Allocator) bool {
	for len(i.buffered) > 0 {
		ok := i.flushOne(allocator)
		if !ok {
			break
		}
	}
	if len(i.buffered) > 0 {
		return false
	}
	return true
}

// Flush persists all allocated data atomically
//
// returns false on allocator failure
func (i *Inode) Flush(allocator *alloc.Allocator) bool {
	i.m.Lock()
	ok := i.flush(allocator)
	i.m.Unlock()
	return ok
}

// assumes lock is held
func (i *Inode) append(b disk.Block) bool {
	if uint64(len(i.addrs))+uint64(len(i.buffered)) >= MaxBlocks {
		return false
	}

	i.buffered = append(i.buffered, b)
	return true
}

// Append adds a block to the inode, without making it persistent.
//
// Returns false on failure (if the allocator or inode are out of space)
func (i *Inode) Append(b disk.Block) bool {
	i.m.Lock()
	ok := i.append(b)
	i.m.Unlock()
	return ok
}
