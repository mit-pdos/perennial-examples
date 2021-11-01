package async_alloc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/async_disk"
)

func TestPopCnt(t *testing.T) {
	assert.Equal(t, uint64(0), popCnt(0))
	assert.Equal(t, uint64(1), popCnt(1))
	assert.Equal(t, uint64(1), popCnt(2))
	assert.Equal(t, uint64(2), popCnt(3))
	assert.Equal(t, uint64(8), popCnt(255))
}

func TestAlloc(t *testing.T) {
	assert := assert.New(t)
	max := uint64(8 * 4096)
	d := async_disk.NewMemDisk(8 * 4096)
	a := MkAlloc(d, 0)
	a.MarkUsed(0)

	assert.Equal(max-1, a.NumFree(), "everything (but 0) should be initially free")

	n := a.AllocNum()
	assert.NotEqual(uint64(0), n, "should not allocate 0")

	a.MarkUsed(n + 1)
	n2 := a.AllocNum()
	assert.NotEqual(n+1, n2, "should not allocate something marked used")

	assert.Equal(max-4, a.NumFree(), "should have used 4 items")

	a.FreeNum(n)
	a.FreeNum(n2)
	assert.Equal(max-2, a.NumFree(), "should have freed")

	// Make a new allocator with the same disk before flushing.
	b := MkAlloc(d, 0)
	assert.Equal(max, b.NumFree(), "everything should be free")

	a.Flush()

	// Make a new allocator with the same disk after flushing.
	c := MkAlloc(d, 0)
	assert.Equal(max-2, c.NumFree(), "should have some used")
	n3 := a.AllocNum()
	assert.NotEqual(n+1, n3, "should not allocate something marked used")

}
