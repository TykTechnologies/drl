package drl

import (
	"sync"
	"sync/atomic"
	"time"
)

// Cache is a synchronized map of items that auto-expire once stale
type Cache struct {
	mutex sync.RWMutex
	ttl   time.Duration
	items map[string]*Item
	stopC chan struct{}
	open  atomic.Value
}

// IsOpen returns true if cache is open. If true this means the cache is
// operational, since the cache uses a background goroutine to manage ttl, this
// will be false when that background process has been terminated marking this
// cache unsuitable for use.
func (c *Cache) IsOpen() bool {
	return c.open.Load().(bool)
}

// Set is a thread-safe way to add new items to the map
func (c *Cache) Set(key string, data Server) {
	c.mutex.Lock()
	item := &Item{data: data}
	item.touch(c.ttl)
	c.items[key] = item
	c.mutex.Unlock()
}

// Get is a thread-safe way to lookup items
// Every lookup, also touches the item, hence extending it's life
func (c *Cache) Get(key string) (data Server, found bool) {
	c.mutex.Lock()
	item, exists := c.items[key]
	if !exists || item.expired() {
		data = Server{}
		found = false
	} else {
		item.touch(c.ttl)
		data = item.data
		found = true
	}
	c.mutex.Unlock()
	return
}

// GetNoExtend is a thread-safe way to lookup items
// Every lookup, also touches the item, hence extending it's life
func (c *Cache) GetNoExtend(key string) (data Server, found bool) {
	c.mutex.Lock()
	item, exists := c.items[key]
	if !exists || item.expired() {
		data = Server{}
		found = false
	} else {
		data = item.data
		found = true
	}
	c.mutex.Unlock()
	return
}

// Count returns the number of items in the cache
// (helpful for tracking memory leaks)
func (c *Cache) Count() int {
	c.mutex.RLock()
	count := len(c.items)
	c.mutex.RUnlock()
	return count
}

// Close frees up resources used by the cache.
func (c *Cache) Close() {
	if c.IsOpen() {
		c.stopC <- struct{}{}
		c.items = nil
		close(c.stopC)
	}
}

func (c *Cache) cleanup() {
	c.mutex.Lock()
	for key, item := range c.items {
		if item.expired() {
			delete(c.items, key)
		}
	}
	c.mutex.Unlock()
}

func (c *Cache) startCleanupTimer() {
	duration := c.ttl
	if duration < time.Second {
		duration = time.Second
	}
	t := time.NewTicker(duration)
	defer t.Stop()
	for {
		select {
		case <-c.stopC:
			c.open.Store(false)
			return
		case <-t.C:
			c.cleanup()
		}
	}
}

// NewCache is a helper to create instance of the Cache struct
func NewCache(duration time.Duration) *Cache {
	cache := &Cache{
		ttl:   duration,
		items: map[string]*Item{},
		stopC: make(chan struct{}),
	}
	cache.open.Store(true)
	go cache.startCleanupTimer()
	return cache
}
