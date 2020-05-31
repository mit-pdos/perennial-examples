package wal

import "github.com/tchajed/goose/machine/disk"

func install(d disk.Disk, txn []update) {
	for _, u := range txn {
		// TODO: for lock-free reads to work,
		//  need to absorb the logged group txn
		d.Write(u.addr, u.b)
	}
}

// logOne takes the current pending transaction and logs and installs it
//
// Assumes lock is held initially.
func (l *Log) logAndInstallOne(app *appender) {
	txn := l.pending
	if uint64(len(txn)) == 0 {
		return
	}
	l.m.Unlock()

	app.Append(txn)
	// now txn is durable
	install(l.d, txn)
	app.Reset()
	// and now it's fully installed

	l.m.Lock()
	// note that there might be new pending transactions which we missed
	l.diskEnd = l.diskEnd + uint64(len(txn))
	l.pending = l.pending[len(txn):]
	// once we unlock, then other threads will know that txn is durable
}

func (l *Log) logger(app *appender) {
	l.m.Lock()
	for {
		l.logAndInstallOne(app)
		// TODO: replace with cond var
		l.m.Unlock()
		l.m.Lock()
	}
}
