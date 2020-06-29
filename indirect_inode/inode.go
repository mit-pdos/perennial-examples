package inode

import (
	"sync"

	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"
)

// on-disk layout of inode:
// [ size: u64 -- number of blocks in inode |
//   direct: [500]u64 -- number valid determined by size |
//   numIndirect: u64 --number of indirect blocks
//   indirect: [10]u64 -- number valid determined by numIndirect ]
//
// indirect block:
// [ direct: [512]u64 -- direct blocks ]
//
// note that a "direct block" means the address of a block of data

// Maximum size of inode, in blocks.
const MaxBlocks uint64 = 500 + 10*512

const maxDirect uint64 = 500
const maxIndirect uint64 = 10
const indirectNumBlocks uint64 = 512

type Inode struct {
	d        disk.Disk
	m        *sync.Mutex
	addr     uint64 // address on disk where inode is stored
	size     uint64
	direct   []uint64 // addresses of data blocks
	indirect []uint64 // addresses of indirect blocks
}

func min(a, b uint64) uint64 {
	if a <= b {
		return a
	}
	return b
}

func Open(d disk.Disk, addr uint64) *Inode {
	b := d.Read(addr)
	dec := marshal.NewDec(b)
	size := dec.GetInt()
	direct := dec.GetInts(maxDirect)
	indirect := dec.GetInts(maxIndirect)
	numIndirect := dec.GetInt()
	numDirect := min(size, maxDirect)
	return &Inode{
		d:        d,
		m:        new(sync.Mutex),
		size:     size,
		addr:     addr,
		direct:   direct[:numDirect],
		indirect: indirect[:numIndirect],
	}
}

func readIndirect(d disk.Disk, a uint64) []uint64 {
	b := d.Read(a)
	dec := marshal.NewDec(b)
	return dec.GetInts(indirectNumBlocks)
}

func prepIndirect(addrs []uint64) disk.Block {
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInts(addrs)
	return enc.Finish()
}

func (i *Inode) UsedBlocks() []uint64 {
	var addrs []uint64
	addrs = make([]uint64, 0)
	direct := i.direct
	indirect := i.indirect
	for _, a := range direct {
		addrs = append(addrs, a)
	}
	// append all addrs pointing to indirect blocks
	for _, blkAddr := range indirect {
		addrs = append(addrs, blkAddr)
	}
	// append all addrs inside indirect blocks pointing to blocks
	for _, blkAddr := range indirect {
		addrs = append(addrs, readIndirect(i.d, blkAddr)...)
	}
	return addrs
}

func indNum(off uint64) uint64 {
	return (off - maxDirect) / indirectNumBlocks
}

func indOff(off uint64) uint64 {
	return (off - maxDirect) % indirectNumBlocks
}

func (i *Inode) Read(off uint64) disk.Block {
	i.m.Lock()
	if off >= i.size {
		i.m.Unlock()
		return nil
	}
	if off < maxDirect {
		a := i.direct[off]
		b := i.d.Read(a)
		i.m.Unlock()
		return b
	}
	addrs := readIndirect(i.d, i.indirect[indNum(off)])
	b := i.d.Read(addrs[indOff(off)])
	i.m.Unlock()
	return b
}

func (i *Inode) Size() uint64 {
	i.m.Lock()
	sz := i.size
	i.m.Unlock()
	return sz
}

func padInts(enc marshal.Enc, num uint64) {
	for i := uint64(0); i < num; i++ {
		enc.PutInt(0)
	}
}

func (i *Inode) mkHdr() disk.Block {
	enc := marshal.NewEnc(disk.BlockSize)
	// sz
	enc.PutInt(i.size)
	// direct_s
	enc.PutInts(i.direct)
	padInts(enc, maxDirect-uint64(len(i.direct)))
	// indirect_s
	enc.PutInts(i.indirect)
	padInts(enc, maxIndirect-uint64(len(i.indirect)))
	// numIndirect
	enc.PutInt(uint64(len(i.indirect)))

	hdr := enc.Finish()
	return hdr
}

func (i *Inode) inSize() {
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
}

// checkTotalSize determines that the inode is not already at maximum size
//
// Requires the lock to be held.
//
func (i *Inode) checkTotalSize() bool {
	if i.size >= MaxBlocks {
		return false
	}
	return true
}

// appendDirect adds address a (and whatever data is stored there) to the inode
//
// Requires the lock to be held.
//
// In this simple design with only direct blocks, appending never requires
// internal allocation, so we don't take an allocator.
//
// This method can only fail due to running out of space in the inode. In this
// case, append returns ownership of the allocated block.
func (i *Inode) appendDirect(a uint64) bool {
	if i.size < maxDirect {
		i.direct = append(i.direct, a)
		i.size += 1
		hdr := i.mkHdr()
		i.d.Write(i.addr, hdr)
		return true
	}
	return false
}

// appendIndirect adds address a (and whatever data is stored there) to the inode
//
// Requires the lock to be held.
//
// In this simple design with only direct blocks, appending never requires
// internal allocation, so we don't take an allocator.
//
// This method can only fail due to running out of space in the inode. In this
// case, append returns ownership of the allocated block.
func (i *Inode) appendIndirect(a uint64) bool {
	if indNum(i.size) >= uint64(len(i.indirect)) {
		return false
	}
	indAddr := i.indirect[indNum(i.size)]
	addrs := readIndirect(i.d, indAddr)
	addrs[indOff(i.size)] = a
	i.writeIndirect(indAddr, addrs)
	return true
}

// writeIndirect preps the block of addrs and
// adds writes the new indirect block to disk
//
// Requires the lock to be held.
func (i *Inode) writeIndirect(indAddr uint64, addrs []uint64) {
	diskBlk := prepIndirect(addrs)
	i.d.Write(indAddr, diskBlk)
	i.size += 1
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
}

// Append adds a block to the inode.
//
// Takes ownership of the disk at a on success.
//
// Returns false on failure (if the allocator or inode are out of space)
func (i *Inode) Append(b disk.Block, allocator *alloc.Allocator) bool {
	i.m.Lock()

	ok := i.checkTotalSize()
	if !ok {
		i.m.Unlock()
		return false
	}

	a, ok2 := allocator.Reserve()
	if !ok2 {
		i.m.Unlock()
		return false
	}
	i.d.Write(a, b)

	ok3 := i.appendDirect(a)
	if ok3 {
		i.m.Unlock()
		return true
	}

	ok4 := i.appendIndirect(a)
	if ok4 {
		i.m.Unlock()
		return true
	}

	// we need to allocate a new indirect block
	// and put the data there
	indAddr, ok := allocator.Reserve()
	if !ok {
		i.m.Unlock()
		allocator.Free(a)
		return false
	}

	i.indirect = append(i.indirect, indAddr)
	i.writeIndirect(indAddr, []uint64{a})
	i.m.Unlock()
	return true
}

// Give a block to the inode for metadata purposes.
// Precondition: Block at addr a should be zeroed
//
// Returns true if the block was consumed.
func (i *Inode) Alloc(a uint64) bool {
	i.m.Lock()
	if uint64(len(i.indirect)) >= maxIndirect {
		i.m.Unlock()
		return false
	}
	i.indirect = append(i.indirect, a)
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
	i.m.Unlock()
	return true
}
