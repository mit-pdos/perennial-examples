package inode

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"
)

// Maximum size of inode, in blocks.
const MaxBlocks uint64 = 511

type Inode struct {
	d     disk.Disk
	m     *sync.Mutex
	addr  uint64   // address on disk where inode is stored
	addrs []uint64 // addresses of data blocks
}

func Open(d disk.Disk, addr uint64) *Inode {
	b := d.Read(addr)
	dec := marshal.NewDec(b)
	numAddrs := dec.GetInt()
	addrs := dec.GetInts(numAddrs)
	return &Inode{d: d, m: new(sync.Mutex), addr: addr, addrs: addrs}
}

// UsedBlocks returns the addresses allocated to the inode for the purposes
// of recovery. Assumes full ownership of the inode, so does not lock,
// and expects the caller to need only temporary access to the returned slice.
func (i *Inode) UsedBlocks() []uint64 {
	return i.addrs
}

func (i *Inode) Read(off uint64) disk.Block {
	i.m.Lock()
	if off >= uint64(len(i.addrs)) {
		i.m.Unlock()
		return nil
	}
	a := i.addrs[off]
	b := i.d.Read(a)
	i.m.Unlock()
	// TODO: can we prove an optimization that unlocks early? It means all
	//  disk operations happen lock-free.
	return b
}

func (i *Inode) Size() uint64 {
	i.m.Lock()
	sz := uint64(len(i.addrs))
	i.m.Unlock()
	return sz
}

func (i *Inode) mkHdr() disk.Block {
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInt(uint64(len(i.addrs)))
	enc.PutInts(i.addrs)
	hdr := enc.Finish()
	return hdr
}

type AppendStatus byte

const (
	AppendOk    AppendStatus = 0
	AppendAgain AppendStatus = 1
	AppendFull  AppendStatus = 2
)

// Append adds a block to the inode.
//
// Takes ownership of the disk at a on success.
//
// Returns:
// - AppendOk on success and takes ownership of the allocated block.
// - AppendFull if inode is out of space (and returns the allocated block)
// - AppendAgain if inode needs a metadata block. Call i.Alloc and try again.
// 	 Returns the allocated block.
func (i *Inode) Append(a uint64) AppendStatus {
	i.m.Lock()
	if uint64(len(i.addrs)) >= MaxBlocks {
		i.m.Unlock()
		return AppendFull
	}
	i.addrs = append(i.addrs, a)
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
	i.m.Unlock()
	return AppendOk
}

// Give a block to the inode for metadata purposes.
//
// Returns true if the block was consumed.
func (i *Inode) Alloc(a uint64) bool {
	return false
}
