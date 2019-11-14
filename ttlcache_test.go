package drl

import (
	"context"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Run("IsOpen", func(ts *testing.T) {
		c := NewCache(context.Background(), 0)
		if !c.IsOpen() {
			t.Error("expected the cache to be open")
		}
		c.Close()
	})
	t.Run("Close", func(ts *testing.T) {
		c := NewCache(context.Background(), 0)
		c.Close()
		if c.IsOpen() {
			t.Error("expected the cache to be closed")
		}
	})
	t.Run("Item evictions", func(ts *testing.T) {
		c := NewCache(context.Background(), 0)
		defer c.Close()
		c.Set("key", Server{})
		time.Sleep(2 * time.Second)
		count := c.Count()
		if count != 0 {
			ts.Errorf("expected 0 items to remain in the cache got %d instead", count)
		}
	})
}

func BenchmarkCache_Set(b *testing.B) {
	c := NewCache(context.Background(), time.Millisecond)
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
