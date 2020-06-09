// Example client of inode that has a single inode and doesn't share the
// allocator with anything else.
package single_inode

import (
	"github.com/tchajed/goose/machine/disk"

	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/mit-pdos/perennial-examples/inode"
)

type SingleInode struct {
	i     *inode.Inode
	alloc *alloc.Allocator
}

func Open(d disk.Disk) *SingleInode {
	i := inode.Open(d, 0)
	used := make(alloc.AddrSet)
	alloc.SetAdd(used, i.UsedBlocks())
	allocator := alloc.New(1, d.Size()-1, used)
	return &SingleInode{i: i, alloc: allocator}
}

func (i *SingleInode) Read(off uint64) disk.Block {
	return i.i.Read(off)
}

func (i *SingleInode) Append(b disk.Block) bool {
	return i.i.Append(b, i.alloc)
}
