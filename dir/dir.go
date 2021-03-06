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

func inodeUsedBlocks(inodes []*inode.Inode) alloc.AddrSet {
	used := make(alloc.AddrSet)
	for _, i := range inodes {
		alloc.SetAdd(used, i.UsedBlocks())
	}
	return used
}

func Open(d disk.Disk, sz uint64) *Dir {
	inodes := openInodes(d)
	used := inodeUsedBlocks(inodes)
	allocator := alloc.New(NumInodes, sz-NumInodes, used)
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
	i := d.inodes[ino]
	return i.Append(b, d.allocator)
}
