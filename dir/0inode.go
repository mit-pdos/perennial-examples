package dir

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
	"github.com/tchajed/marshal"
)

type inode struct {
	d     disk.Disk
	m     *sync.Mutex
	addr  uint64   // address on disk where inode is stored
	addrs []uint64 // addresses of data blocks
}

func openInode(d disk.Disk, addr uint64) *inode {
	b := d.Read(addr)
	dec := marshal.NewDec(b)
	numAddrs := dec.GetInt()
	addrs := dec.GetInts(numAddrs)
	return &inode{d: d, m: new(sync.Mutex), addr: addr, addrs: addrs}
}

func (i *inode) UsedBlocks() []uint64 {
	i.m.Lock()
	addrs := i.addrs
	i.m.Unlock()
	return addrs
}

func (i *inode) Read(off uint64) disk.Block {
	i.m.Lock()
	a := i.addrs[off]
	b := i.d.Read(a)
	i.m.Unlock()
	return b
}

func (i *inode) Size() uint64 {
	i.m.Lock()
	sz := uint64(len(i.addrs))
	i.m.Unlock()
	return sz
}

func (i *inode) Append(a uint64) {
	i.m.Lock()
	i.addrs = append(i.addrs, a)
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInt(uint64(len(i.addrs)))
	enc.PutInts(i.addrs)
	hdr := enc.Finish()
	i.d.Write(i.addr, hdr)
	i.m.Unlock()
}
