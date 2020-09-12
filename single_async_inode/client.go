// Example client of inode that has a single inode and doesn't share the
// allocator with anything else.
package single_async_inode

import (
	"github.com/tchajed/goose/machine/disk"

	"github.com/mit-pdos/perennial-examples/alloc"
	"github.com/mit-pdos/perennial-examples/async_inode"
)

type SingleInode struct {
	i     *async_inode.Inode
	alloc *alloc.Allocator
}

// Restore the SingleInode from disk
//
// sz should be the size of the disk to use
func Open(d disk.Disk, sz uint64) *SingleInode {
	i := async_inode.Open(d, 0)
	used := make(alloc.AddrSet)
	alloc.SetAdd(used, i.UsedBlocks())
	allocator := alloc.New(1, sz-1, used)
	return &SingleInode{i: i, alloc: allocator}
}

func (i *SingleInode) Read(off uint64) disk.Block {
	return i.i.Read(off)
}

func (i *SingleInode) Append(b disk.Block) bool {
	return i.i.Append(b)
}

func (i *SingleInode) Flush() bool {
	return i.i.Flush(i.alloc)
}
