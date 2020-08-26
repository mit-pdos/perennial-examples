package dynamic_dir

import (
	"sync"

	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/mit-pdos/perennial-examples/inode"
	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"
)

// MaxInodes = 512 (the number of inode addresses that fit into a single root
// inode)
const MaxInodes uint64 = disk.BlockSize / 8

const rootInode uint64 = 0

type Dir struct {
	d         disk.Disk
	allocator *alloc.Allocator

	m      *sync.Mutex
	inodes map[uint64]*inode.Inode
}

func (d *Dir) mkHdr() disk.Block {
	var inode_addrs []uint64
	for a, _ := range d.inodes {
		inode_addrs = append(inode_addrs, a)
	}
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInt(uint64(len(inode_addrs)))
	enc.PutInts(inode_addrs)
	return enc.Finish()
}

func (d *Dir) writeHdr() {
	hdr := d.mkHdr()
	d.d.Write(rootInode, hdr)
}

// read header (which has addresses for inodes in use)
func parseHdr(b disk.Block) []uint64 {
	dec := marshal.NewDec(b)
	num := dec.GetInt()
	return dec.GetInts(num)
}

func openInodes(d disk.Disk) map[uint64]*inode.Inode {
	inode_addrs := parseHdr(d.Read(rootInode))
	inodes := make(map[uint64]*inode.Inode)
	for _, a := range inode_addrs {
		inodes[a] = inode.Open(d, a)
	}
	return inodes
}

func inodeUsedBlocks(inodes map[uint64]*inode.Inode) alloc.AddrSet {
	used := make(alloc.AddrSet)
	for a, i := range inodes {
		alloc.SetAdd(used, []uint64{a})
		alloc.SetAdd(used, i.UsedBlocks())
	}
	return used
}

func Open(d disk.Disk, sz uint64) *Dir {
	inodes := openInodes(d)
	used := inodeUsedBlocks(inodes)
	// reserve 1 block for root inode
	allocator := alloc.New(1, sz-1, used)
	return &Dir{
		d:         d,
		allocator: allocator,
		m:         new(sync.Mutex),
		inodes:    inodes,
	}
}

func (d *Dir) Create() (uint64, bool) {
	a, ok := d.allocator.Reserve()
	if !ok {
		return 0, false
	}
	empty := make(disk.Block, disk.BlockSize)
	d.d.Write(a, empty)
	d.m.Lock()
	d.inodes[a] = inode.Open(d.d, a)
	d.writeHdr()
	d.m.Unlock()
	return a, true
}

func (d *Dir) delete(ino uint64) {
	i := d.inodes[ino]
	delete(d.inodes, ino)
	d.writeHdr() // crash commit point
	// now we can free all the used addresses for other threads to use
	// (somewhat optional - restarting the system would also free these
	// addresses)
	d.allocator.Free(ino)
	for _, inode_a := range i.UsedBlocks() {
		d.allocator.Free(inode_a)
	}
}

func (d *Dir) Delete(ino uint64) {
	d.m.Lock()
	d.delete(ino)
	d.m.Unlock()
}

func (d *Dir) Read(ino uint64, off uint64) disk.Block {
	d.m.Lock()
	i := d.inodes[ino]
	if i == nil {
		panic("invalid inode")
	}
	b := i.Read(off)
	d.m.Unlock()
	return b
}

func (d *Dir) Size(ino uint64) uint64 {
	d.m.Lock()
	i := d.inodes[ino]
	if i == nil {
		panic("invalid inode")
	}
	sz := i.Size()
	d.m.Unlock()
	return sz
}

func (d *Dir) Append(ino uint64, b disk.Block) bool {
	d.m.Lock()
	i := d.inodes[ino]
	if i == nil {
		panic("invalid inode")
	}
	ok := i.Append(b, d.allocator)
	d.m.Unlock()
	return ok
}
