package dir

import (
	"github.com/tchajed/goose/machine/disk"
)

const NumInodes uint64 = 5

type Dir struct {
	d         disk.Disk
	allocator *allocator
	inodes    []*inode
}

func openInodes(d disk.Disk) []*inode {
	var inodes []*inode
	for addr := uint64(0); addr < NumInodes; addr++ {
		inodes = append(inodes, openInode(d, addr))
	}
	return inodes
}

func deleteInodeBlocks(numInodes uint64, free map[uint64]unit) {
	for i := uint64(0); i < numInodes; i++ {
		delete(free, i)
	}
}

func OpenDir(d disk.Disk, free map[uint64]unit) *Dir {
	inodes := openInodes(d)
	for _, i := range inodes {
		for _, a := range i.UsedBlocks() {
			delete(free, a)
		}
	}
	deleteInodeBlocks(NumInodes, free)
	allocator := newAllocator(free)
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
