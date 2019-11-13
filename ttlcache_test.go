package drl

import (
	"strconv"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Run("IsOpen", func(ts *testing.T) {
		c := NewCache(0)
		if !c.IsOpen() {
			t.Error("expected the cache to be open")
		}
		c.Close()
	})
	t.Run("Close", func(ts *testing.T) {
		c := NewCache(0)
		c.Close()
		if c.IsOpen() {
			t.Error("expected the cache to be closed")
		}
	})
	t.Run("Item evictions", func(ts *testing.T) {
		// This test ensures that we are correctly evicting expired items. We set ttl
		// to me 1ms, this will force the eviction loop to use 1s interval for
		// eviction.
		c := NewCache(time.Millisecond)
		defer c.Close()

		// Ticking at 100ms we will reach 1s after 10 ticks, meaning any items added
		// after 10th tick will be available in the cache because tey will be evicted
		// on 20th tick.
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		var n int64
		for range tick.C {
			n++
			if n == 15 {
				break
			}
			c.Set(strconv.FormatInt(n, 10), Server{})
		}
		count := c.Count()
		if count != 5 {
			ts.Errorf("expected 5 items to remain in the cache got %d instead", count)
		}
		_, ok := c.Get("14")
		if ok {
			t.Error("expected the key to have expired")
		}
		key := "key"
		c.Set(key, Server{})
		_, ok = c.Get(key)
		if !ok {
			t.Error("expected the key to exist")
		}
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
