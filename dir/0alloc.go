package dir

import "sync"

type unit struct{}

// allocator manages free disk blocks. It does not store its state durably, so
// the caller is responsible for returning its set of free disk blocks on
// recovery.
type allocator struct {
	m    *sync.Mutex
	free map[uint64]unit
}

func FreeRange(start, sz uint64) map[uint64]unit {
	m := make(map[uint64]unit)
	for i := start; i < start+sz; i++ {
		m[i] = unit{}
	}
	return m
}

func newAllocator(free map[uint64]unit) *allocator {
	return &allocator{m: new(sync.Mutex), free: free}
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

// Reserve transfers ownership of a free block from the allocator to the caller
func (a *allocator) Reserve() (uint64, bool) {
	a.m.Lock()
	k, ok := findKey(a.free)
	delete(a.free, k)
	a.m.Unlock()
	return k, ok
}
