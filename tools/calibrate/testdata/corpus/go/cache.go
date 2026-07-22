package ledger

import (
	"sync"
	"time"
)

type cacheEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache stores values until their individual deadlines.
type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]cacheEntry[V]
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{entries: make(map[K]cacheEntry[V])}
}

func (c *Cache[K, V]) Put(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry[V]{value: value, expiresAt: time.Now().Add(ttl)}
}

func (c *Cache[K, V]) Get(key K, now time.Time) (V, bool) {
	c.mu.RLock()
	entry, found := c.entries[key]
	c.mu.RUnlock()
	if found && now.Before(entry.expiresAt) {
		return entry.value, true
	}
	var zero V
	return zero, false
}
