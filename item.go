package drl

import (
	"sync"
	"time"
)

// Item represents a record in the cache map
type Item struct {
	mu      sync.RWMutex
	data    Server
	expires *time.Time
}

func (item *Item) touch(duration time.Duration) {
	item.mu.Lock()
	expiration := time.Now().Add(duration)
	item.expires = &expiration
	item.mu.Unlock()
}

func (item *Item) expired() bool {
	var value bool
	item.mu.RLock()
	if item.expires == nil {
		value = true
	} else {
		value = item.expires.Before(time.Now())
	}
	item.mu.RUnlock()
	return value
}
