package dir

import (
	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/mit-pdos/perennial-examples/inode"
	"github.com/tchajed/goose/machine/disk"
)

const NumInodes uint64 = 5

type Dir struct {
	d         disk.Disk
	allocator *alloc.Allocator
	inodes    []*inode.Inode
}

func openInodes(d disk.Disk) []*inode.Inode {
	var inodes []*inode.Inode
	for addr := uint64(0); addr < NumInodes; addr++ {
		inodes = append(inodes, inode.Open(d, addr))
	}
	return inodes
}

func deleteInodeBlocks(numInodes uint64, free alloc.AddrSet) {
	for i := uint64(0); i < numInodes; i++ {
		delete(free, i)
	}
}

func OpenDir(d disk.Disk, free alloc.AddrSet) *Dir {
	inodes := openInodes(d)
	for _, i := range inodes {
		for _, a := range i.UsedBlocks() {
			delete(free, a)
		}
	}
	deleteInodeBlocks(NumInodes, free)
	allocator := alloc.New(free)
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

func (d *Dir) Size(ino uint64) uint64 {
	i := d.inodes[ino]
	return i.Size()
}

func (d *Dir) Append(ino uint64, b disk.Block) bool {
	a, ok := d.allocator.Reserve()
	if !ok {
		return false
	}
	d.d.Write(a, b)
	ok2 := d.inodes[ino].Append(a)
	if !ok2 {
		return false
	}
	return true
}
