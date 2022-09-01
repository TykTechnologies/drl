package drl

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// Modify the minimumCleanupInterval to make expiry faster
	minimumCleanupInterval = 100 * time.Millisecond
}

func TestCache(t *testing.T) {
	t.Parallel()

	t.Run("IsOpen", func(t *testing.T) {
		t.Parallel()

		c := NewCache(0)
		defer c.Close()

		assert.True(t, c.IsOpen(), "expected the cache to be open")
	})

	t.Run("Close", func(t *testing.T) {
		t.Parallel()

		c := NewCache(0)
		c.Close()

		assert.False(t, c.IsOpen(), "expected the cache to be closed")
	})

	t.Run("Set, Get", func(t *testing.T) {
		t.Parallel()

		c := NewCache(10 * time.Millisecond)
		defer c.Close()

		key := "key"
		c.Set(key, Server{})
		_, ok := c.Get(key)
		assert.True(t, ok, "expected valid key")
	})

	t.Run("Set, Get", func(t *testing.T) {
		t.Parallel()

		c := NewCache(0)
		defer c.Close()

		key := "key"
		c.Set(key, Server{})
		_, ok := c.Get(key)
		assert.False(t, ok, "expected key to be expired")
	})

	t.Run("Item evictions", func(t *testing.T) {
		t.Parallel()

		// Test with 10 items, calculate tick to hit cleanup interval.

		addItemInterval := minimumCleanupInterval / 10

		// This test ensures that we are correctly evicting expired items.
		// The expiry is very short, so we know that all the items added until
		// the first expiry will immediately expire.

		c := NewCache(time.Microsecond)
		defer c.Close()

		// Offset adding items by half of an addTicketInterval, so we can
		// reliably predict how many items will be left in the cache at expiry.

		time.Sleep(addItemInterval / 2)

		// Start a ticker so we may add a new cache item every addItemInterval.
		// Add 14 items, 5 of which will be added in the next expiry interval.

		tick := time.NewTicker(addItemInterval)
		defer tick.Stop()

		var n int64
		for range tick.C {
			n++
			if n == 15 {
				break
			}
			c.Set(strconv.FormatInt(n, 10), Server{})
		}

		// Assert the expected cache item count.

		got := c.Count()
		want := 5
		assert.Equal(t, want, got, "expected %d items to remain in the cache got %d instead", want, got)
	})
}

func BenchmarkCache_Set(b *testing.B) {
	c := NewCache(time.Millisecond)
	s := Server{}
	key := "1"
	defer c.Close()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(key, s)
		}
	})
}
