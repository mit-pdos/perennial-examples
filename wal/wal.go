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

func (l *Log) open(d disk.Disk) (*Log, *appender) {
	m := new(sync.Mutex)
	app, upds := openAppender(d)
	install(l.d, upds)
	return &Log{d: d, m: m, diskEnd: 0, pending: []update{}}, app
}

func (l *Log) Open(d disk.Disk) *Log {
	l, app := l.open(d)
	go func() { l.logger(app) }()
	return l
}

func (l *Log) Write(upds []update) {
	l.m.Lock()
	if uint64(len(l.pending))+uint64(len(upds)) > maxLogSize {
		// TODO: wait for space
	}
	l.pending = append(l.pending, upds...)
	txnId := l.diskEnd + uint64(len(l.pending))
	l.m.Unlock()

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

func (l *Log) Read(a uint64) disk.Block {
	l.m.Lock()
	b := l.d.Read(a)
	l.m.Unlock()
	return b
}
