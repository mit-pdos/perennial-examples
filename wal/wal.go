package wal

import (
	"sync"

	"github.com/tchajed/goose/machine/disk"
)

type Log struct {
	// read-only state
	d disk.Disk

	m       *sync.Mutex
	diskEnd uint64
	pending []update
}

func open(d disk.Disk) (*Log, *appender) {
	m := new(sync.Mutex)
	app, upds := openAppender(d)
	install(d, upds)
	return &Log{d: d, m: m, diskEnd: 0, pending: []update{}}, app
}

func Open(d disk.Disk) *Log {
	l, app := open(d)
	go func() { l.logger(app) }()
	return l
}

// waitForSpaceAndLock waits until the log has space for numUpdates
//
// Requires numUpdates < maxLogSpace both to avoid int overflow and also for
// progress (there will never be space otherwise).
//
// Acquires the lock in the process.
//
// Looks exactly like a lock acquire with an extra pure postcondition that
// there is space in the log.
func (l *Log) waitForSpaceAndLock(numUpdates uint64) {
	l.m.Lock()
	for {
		if uint64(len(l.pending))+numUpdates <= maxLogSize {
			break
		}
		l.m.Unlock()
		l.m.Lock()
		continue
	}
	// establishes len(l.pending) + numUpdates <= maxLogSize
	return
}

func (l *Log) writePrepare(upds []update) (uint64, bool) {
	if uint64(len(upds)) > maxLogSize {
		return 0, false
	}
	l.waitForSpaceAndLock(uint64(len(upds)))
	l.pending = append(l.pending, upds...)
	txnId := l.diskEnd + uint64(len(l.pending))
	l.m.Unlock()
	return txnId, true
}

func (l *Log) writeWait(txnId uint64) {
	l.m.Lock()
	for {
		if l.diskEnd >= txnId {
			// this establishes that the transaction has been committed durably
			l.m.Unlock()
			break
		}
		// TODO: use a cond var
		l.m.Unlock()
		l.m.Lock()
	}
}

func (l *Log) Write(upds []update) bool {
	txnId, ok := l.writePrepare(upds)
	if !ok {
		return false
	}
	l.writeWait(txnId)
	return true
}

func (l *Log) Read(a uint64) disk.Block {
	return l.d.Read(a)
}
