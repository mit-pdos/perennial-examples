package async_alloc

import (
	"sync"

	"github.com/tchajed/goose/machine/async_disk"
)

// Allocator uses a bit map to allocate and free numbers. Bit 0
// corresponds to number 0, bit 1 to 1, and so on.
type Alloc struct {
	// read only
	d    async_disk.Disk
	mu   *sync.Mutex
	addr uint64 // starting on disk address of bitmap

	// mutable
	next   uint64 // first number to try
	bitmap []byte
	dirty  bool
}

// MkAlloc initializes with a bitmap.
//
func MkAlloc(d async_disk.Disk, addr uint64) *Alloc {
	bitmap := d.Read(addr)
	a := &Alloc{
		d:      d,
		mu:     new(sync.Mutex),
		addr:   addr,
		next:   0,
		bitmap: bitmap,
		dirty:  false,
	}
	return a
}

func (a *Alloc) MarkUsed(bn uint64) {
	a.mu.Lock()
	byte := bn / 8
	bit := bn % 8
	a.bitmap[byte] = a.bitmap[byte] | (1 << bit)
	a.dirty = true
	a.mu.Unlock()
}

func (a *Alloc) incNext() uint64 {
	a.next = a.next + 1
	if a.next >= uint64(len(a.bitmap)*8) {
		a.next = 0
	}
	return a.next
}

// Returns a free number in the bitmap
func (a *Alloc) allocBit() uint64 {
	var num uint64
	a.mu.Lock()
	num = a.incNext()
	start := num
	for {
		bit := num % 8
		byte := num / 8
		// util.DPrintf(10, "allocBit: s %d num %d\n", start, num)
		if a.bitmap[byte]&(1<<bit) == 0 {
			a.bitmap[byte] = a.bitmap[byte] | (1 << bit)
			a.dirty = true
			break
		}
		num = a.incNext()
		if num == start { // looped around?
			num = 0
			break
		}
		continue
	}
	a.mu.Unlock()
	return num
}

func (a *Alloc) freeBit(bn uint64) {
	a.mu.Lock()
	byte := bn / 8
	bit := bn % 8
	a.bitmap[byte] = a.bitmap[byte] & ^(1 << bit)
	a.dirty = true
	a.mu.Unlock()
}

func (a *Alloc) AllocNum() uint64 {
	num := a.allocBit()
	return num
}

func (a *Alloc) FreeNum(num uint64) {
	if num == 0 {
		panic("FreeNum")
	}
	a.freeBit(num)
}

func (a *Alloc) Flush() {
	a.mu.Lock()
	if a.dirty {
		a.d.Write(a.addr, a.bitmap)
		a.dirty = false
	}
	a.mu.Unlock()
}

func popCnt(b byte) uint64 {
	var count uint64
	var x = b
	for i := uint64(0); i < 8; i++ {
		count += uint64(x & 1)
		x = x >> 1
	}
	return count
}

func (a *Alloc) NumFree() uint64 {
	a.mu.Lock()
	total := 8 * uint64(len(a.bitmap))
	var count uint64
	for _, b := range a.bitmap {
		count += popCnt(b)
	}
	a.mu.Unlock()
	return total - count
}
