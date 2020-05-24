package alloc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllocatorReservationUnique(t *testing.T) {
	assert := assert.New(t)
	alloc := New(5, 10, AddrSet{})
	a1, ok := alloc.Reserve()
	assert.GreaterOrEqual(a1, uint64(5), "allocated address %d should be in range", a1)
	assert.True(ok)
	a2, ok := alloc.Reserve()
	assert.True(ok)
	assert.NotEqual(a1, a2, "reserved same block twice")
}

func TestAllocatorAll(t *testing.T) {
	assert := assert.New(t)
	alloc := New(5, 10, AddrSet{})
	for i := 0; i < 10; i++ {
		_, ok := alloc.Reserve()
		assert.True(ok, "reservation failed early: %d", i)
	}
	_, ok := alloc.Reserve()
	assert.False(ok, "all addresses should be allocatd")
}
