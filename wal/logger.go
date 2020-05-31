package wal

import "github.com/tchajed/goose/machine/disk"

func install(d disk.Disk, txn []update) {
	// TODO: we need threads to either not observe these writes or see them
	//  all atomically. Not observing them is hard,
	//  since we don't have the old values.
	//  Committing them requires that we atomically write the header and
	//  cause threads to start reading from the wal,
	//  which we can do with a reverse search in the wal.
	//  This means we should prepare the wal, _lock_,
	//  and then write the header, which breaks the atomic_append API.
	for _, u := range txn {
		d.Write(u.addr, u.b)
	}
}

func absorb(txn []update) []update {
	addrs := make(map[uint64]uint64)
	var absorbed []update
	for _, u := range txn {
		i, ok := addrs[u.addr]
		if ok {
			absorbed[i].b = u.b
		} else {
			newIndex := uint64(len(absorbed))
			addrs[u.addr] = newIndex
			absorbed = append(absorbed, u)
		}
	}
	return absorbed
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

	absorbed := absorb(txn)
	app.Append(absorbed)
	// now txn (via absorbed) is durable
	install(l.d, absorbed)
	app.Reset()
	// and now it's fully installed

	l.m.Lock()
	// note that there might be new pending transactions which we missed
	// NOTE: diskEnd does not measure physical number of updates in log; do
	//  we really need it? can it be a transaction count?
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
