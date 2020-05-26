package alloc

import "sync"

type unit struct{}

type AddrSet = map[uint64]unit

// Allocator manages free disk blocks. It does not store its state durably, so
// the caller is responsible for returning its set of free disk blocks on
// recovery.
type Allocator struct {
	m    *sync.Mutex
	free map[uint64]unit
}

func freeRange(start, sz uint64) AddrSet {
	m := make(AddrSet)
	end := start + sz
	for i := start; i < end; i++ {
		m[i] = unit{}
	}
	return m
}

// mapRemove deletes addresses in remove from m
//
// like m -= remove
func mapRemove(m AddrSet, remove AddrSet) {
	for k := range remove {
		delete(m, k)
	}
}

// SetAdd adds addresses in add to m
//
// like m += add
func SetAdd(m AddrSet, add []uint64) {
	for _, k := range add {
		m[k] = unit{}
	}
}

func New(start, sz uint64, used AddrSet) *Allocator {
	free := freeRange(start, sz)
	mapRemove(free, used)
	return &Allocator{m: new(sync.Mutex), free: free}
}

func findKey(m map[uint64]unit) (uint64, bool) {
	var found uint64 = 0
	var ok bool = false
	for k := range m {
		if !ok {
			found = k
			ok = true
		}
		// TODO: goose doesn't support break in map iteration
	}
	return found, ok
}

// Reserve transfers ownership of a free block from the Allocator to the caller
func (a *Allocator) Reserve() (uint64, bool) {
	a.m.Lock()
	k, ok := findKey(a.free)
	delete(a.free, k)
	a.m.Unlock()
	return k, ok
}

func (a *Allocator) Free(addr uint64) {
	a.m.Lock()
	a.free[addr] = unit{}
	a.m.Unlock()
}
