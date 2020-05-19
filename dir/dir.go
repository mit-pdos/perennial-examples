package lvm

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"
)

type unit struct{}

// Allocator manages free disk blocks. It does not store its state durably, so
// the caller is responsible for returning its set of free disk blocks on
// recovery.
type Allocator struct {
	m    *sync.Mutex
	free map[uint64]unit
}

func NewAllocator(free map[uint64]unit) *Allocator {
	return &Allocator{m: new(sync.Mutex), free: free}
}

func findKey(m map[uint64]unit) (uint64, bool) {
	var found uint64 = 0
	var ok bool = false
	for k := range m {
		if !ok {
			found = k
			ok = true
		}
		// TODO: goose doesn't support break in map iteration
	}
	return found, ok
}

// Reserve transfers ownership of a free block from the allocator to the caller
func (a *Allocator) Reserve() (uint64, bool) {
	a.m.Lock()
	k, ok := findKey(a.free)
	delete(a.free, k)
	a.m.Unlock()
	return k, ok
}

type Inode struct {
	d     disk.Disk
	m     *sync.Mutex
	addr  uint64   // address on disk where inode is stored
	addrs []uint64 // addresses of data blocks
}

func OpenInode(d disk.Disk, addr uint64) Inode {
	b := disk.Read(addr)
	dec := marshal.NewDec(b)
	numAddrs := dec.GetInt()
	addrs := dec.GetInts(numAddrs)
	return Inode{d: d, m: new(sync.Mutex), addr: addr, addrs: addrs}
}

func (i Inode) UsedBlocks() []uint64 {
	i.m.Lock()
	addrs := i.addrs
	i.m.Unlock()
	return addrs
}

func (i Inode) Read(off uint64) disk.Block {
	i.m.Lock()
	a := i.addrs[off]
	b := i.d.Read(a)
	i.m.Unlock()
	return b
}

func (i Inode) Append(a uint64) {
	i.m.Lock()
	i.addrs = append(i.addrs, a)
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInt(uint64(len(i.addrs)))
	enc.PutInts(i.addrs)
	hdr := enc.Finish()
	disk.Write(i.addr, hdr)
	i.m.Unlock()
}

const NumInodes uint64 = 5

type Dir struct {
	d         disk.Disk
	allocator *Allocator
	inodes    []Inode
}

func OpenDir(d disk.Disk, free map[uint64]unit) *Dir {
	var inodes []Inode
	for addr := uint64(0); addr < NumInodes; addr++ {
		inodes = append(inodes, OpenInode(d, addr))
	}
	for _, i := range inodes {
		for _, a := range i.UsedBlocks() {
			delete(free, a)
		}
	}
	allocator := NewAllocator(free)
	return &Dir{
		d:         d,
		allocator: allocator,
		inodes:    inodes,
	}
}

func (d *Dir) Read(ino uint64, off uint64) disk.Block {
	i := d.inodes[ino]
	return i.Read(off)
}

func (d *Dir) Append(ino uint64, b disk.Block) bool {
	a, ok := d.allocator.Reserve()
	if !ok {
		return false
	}
	disk.Write(a, b)
	d.inodes[ino].Append(a)
	return true
}
