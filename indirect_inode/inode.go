package inode

import (
	"sync"

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
	numDirect := min(size, maxDirect)
	direct := dec.GetInts(maxDirect)
	numIndirect := dec.GetInt()
	indirect := dec.GetInts(maxIndirect)
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
	i.m.Lock()
	direct := i.direct
	indirect := i.indirect
	i.m.Unlock()
	for _, a := range direct {
		addrs = append(addrs, a)
	}
	for _, blkAddr := range indirect {
		addrs = append(addrs, blkAddr)
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
	enc.PutInt(i.size)
	direct := i.direct
	enc.PutInts(direct)
	padInts(enc, maxDirect-uint64(len(direct)))
	enc.PutInt(uint64(len(i.indirect)))
	enc.PutInts(i.indirect)
	padInts(enc, maxIndirect-uint64(len(i.indirect)))
	hdr := enc.Finish()
	return hdr
}

type AppendStatus byte

const (
	AppendOk    AppendStatus = 0
	AppendAgain AppendStatus = 1
	AppendFull  AppendStatus = 2
)

func (i *Inode) inSize() {
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
}

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

	if i.size >= MaxBlocks {
		i.m.Unlock()
		return AppendFull
	}

	if uint64(len(i.direct)) < maxDirect {
		i.direct = append(i.direct, a)
		i.size += 1
		hdr := i.mkHdr()
		i.d.Write(i.addr, hdr)
		i.m.Unlock()
		return AppendOk
	}

	if indNum(i.size) < uint64(len(i.indirect)) {
		indAddr := i.indirect[indNum(i.size)]
		addrs := readIndirect(i.d, indAddr)
		addrs[indOff(i.size)] = a
		diskBlk := prepIndirect(addrs)
		i.d.Write(indAddr, diskBlk)

		i.size += 1
		hdr := i.mkHdr()
		i.d.Write(i.addr, hdr)
		i.m.Unlock()
		return AppendOk
	}

	i.indirect = append(i.indirect, a)
	hdr := i.mkHdr()
	i.d.Write(i.addr, hdr)
	i.m.Unlock()
	return AppendAgain
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
