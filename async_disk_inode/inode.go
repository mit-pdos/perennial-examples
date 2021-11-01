package inode

import (
	"sync"

	"github.com/tchajed/goose/machine/async_disk"
	"github.com/tchajed/marshal"

	"github.com/mit-pdos/perennial-examples/async_alloc"
)

// Maximum size of inode, in blocks.
const MaxBlocks uint64 = 511

type Inode struct {
	// read-only
	d    async_disk.Disk
	m    *sync.Mutex
	addr uint64 // address on disk where inode is stored

	// mutable
	addrs []uint64 // addresses of data blocks
}

func Open(d async_disk.Disk, addr uint64) *Inode {
	b := d.Read(addr)
	dec := marshal.NewDec(b)
	numAddrs := dec.GetInt()
	addrs := dec.GetInts(numAddrs)
	return &Inode{
		d:     d,
		m:     new(sync.Mutex),
		addr:  addr,
		addrs: addrs,
	}
}

// UsedBlocks returns the addresses allocated to the inode for the purposes
// of recovery. Assumes full ownership of the inode, so does not lock,
// and expects the caller to need only temporary access to the returned slice.
func (i *Inode) UsedBlocks() []uint64 {
	return i.addrs
}

func (i *Inode) read(off uint64) async_disk.Block {
	if off >= uint64(len(i.addrs)) {
		return nil
	}
	a := i.addrs[off]
	return i.d.Read(a)
}

func (i *Inode) Read(off uint64) async_disk.Block {
	i.m.Lock()
	b := i.read(off)
	i.m.Unlock()
	return b
}

func (i *Inode) Size() uint64 {
	i.m.Lock()
	sz := uint64(len(i.addrs))
	i.m.Unlock()
	return sz
}

func (i *Inode) mkHdr() async_disk.Block {
	enc := marshal.NewEnc(async_disk.BlockSize)
	enc.PutInt(uint64(len(i.addrs)))
	enc.PutInts(i.addrs)
	hdr := enc.Finish()
	return hdr
}

// append adds address a (and whatever data is stored there) to the inode
//
// Requires the lock to be held.
//
// In this simple design with only direct blocks, appending never requires
// internal allocation, so we don't take an allocator.
//
// This method can only fail due to running out of space in the inode. In this
// case, append returns ownership of the allocated block.
func (i *Inode) append(a uint64) bool {
	if uint64(len(i.addrs)) >= MaxBlocks {
		return false
	}

	i.addrs = append(i.addrs, a)
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
	i.d.Barrier()
	return true
}

// Append adds a block to the inode.
//
// Returns false on failure (if the allocator or inode are out of space)
func (i *Inode) Append(b async_disk.Block, allocator *async_alloc.Alloc) bool {
	// allocate lock-free
	a := allocator.AllocNum()
	// prepare lock-free
	i.d.Write(a, b)
	allocator.Flush()
	i.d.Barrier()

	i.m.Lock()
	ok2 := i.append(a)
	i.m.Unlock()
	if !ok2 {
		allocator.FreeNum(a)
	}
	return ok2
}
