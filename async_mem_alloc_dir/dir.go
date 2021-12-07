package dir

import (
	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/mit-pdos/perennial-examples/async_mem_alloc_inode"
	"github.com/tchajed/goose/machine/async_disk"
)

const NumInodes uint64 = 5

type Dir struct {
	d         async_disk.Disk
	allocator *alloc.Allocator
	inodes    []*async_mem_alloc_inode.Inode
}

func openInodes(d async_disk.Disk) []*async_mem_alloc_inode.Inode {
	var inodes []*async_mem_alloc_inode.Inode
	for addr := uint64(0); addr < NumInodes; addr++ {
		inodes = append(inodes, async_mem_alloc_inode.Open(d, addr))
	}
	return inodes
}

func inodeUsedBlocks(inodes []*async_mem_alloc_inode.Inode) alloc.AddrSet {
	used := make(alloc.AddrSet)
	for _, i := range inodes {
		alloc.SetAdd(used, i.UsedBlocks())
	}
	return used
}

func Open(d async_disk.Disk, sz uint64) *Dir {
	inodes := openInodes(d)
	used := inodeUsedBlocks(inodes)
	allocator := alloc.New(NumInodes, sz-NumInodes, used)
	return &Dir{
		d:         d,
		allocator: allocator,
		inodes:    inodes,
	}
}

func (d *Dir) Read(ino uint64, off uint64) async_disk.Block {
	i := d.inodes[ino]
	return i.Read(off)
}

func (d *Dir) Size(ino uint64) uint64 {
	i := d.inodes[ino]
	return i.Size()
}

func (d *Dir) Append(ino uint64, b async_disk.Block) bool {
	i := d.inodes[ino]
	return i.Append(b, d.allocator)
}
