// Append-only, sequential, crash-safe log.
//
// The main interesting feature is that the log supports multi-block atomic
// appends, which are implemented by atomically updating an on-disk header with
// the number of valid blocks in the log.
package wal

import (
	"github.com/tchajed/marshal"

	"github.com/tchajed/goose/machine/disk"
)

const maxLogSize uint64 = 511

type update struct {
	addr uint64
	b    disk.Block
}

type appender struct {
	d     disk.Disk
	addrs []uint64
}

func (app *appender) mkHdr() disk.Block {
	enc := marshal.NewEnc(disk.BlockSize)
	enc.PutInt(uint64(len(app.addrs)))
	enc.PutInts(app.addrs)
	return enc.Finish()
}

func (app *appender) writeHdr() {
	app.d.Write(0, app.mkHdr())
}

func openAppender(d disk.Disk) (*appender, []update) {
	hdr := d.Read(0)
	dec := marshal.NewDec(hdr)
	sz := dec.GetInt()
	addrs := dec.GetInts(sz)
	var upds = make([]update, 0)
	for i, addr := range addrs {
		upds = append(upds, update{
			addr: addr,
			b:    d.Read(1 + uint64(i)),
		})
	}
	return &appender{d: d, addrs: addrs}, upds
}

func writeAll(d disk.Disk, upds []update, off uint64) {
	for i, u := range upds {
		d.Write(off+uint64(i), u.b)
	}
}

func (app *appender) Append(upds []update) bool {
	sz := uint64(len(app.addrs))
	if sz+uint64(len(upds)) > maxLogSize {
		return false
	}
	writeAll(app.d, upds, 1+sz)
	for _, u := range upds {
		app.addrs = append(app.addrs, u.addr)
	}
	app.writeHdr()
	return true
}

func (app *appender) Reset() {
	app.addrs = nil
	app.writeHdr()
}
