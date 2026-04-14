package ttlcache

import (
	"sync"
	"time"
)

type entry struct {
	value  any
	expiry time.Time
	size   int
}

// Cache implements TTL-based cache with memory limits
type Cache struct {
	mu          sync.RWMutex
	items       map[string]*entry
	maxSize     int
	currentSize int
	ttl         time.Duration
}

// New creates a cache with TTL and max size
func New(ttl time.Duration, maxSize int) *Cache {
	c := &Cache{
		items:   make(map[string]*entry),
		maxSize: maxSize,
		ttl:     ttl,
	}
	go c.cleanup()
	return c
}

// Set adds item to cache
func (c *Cache) Set(key string, value any, size int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove old entry if exists
	if old, exists := c.items[key]; exists {
		c.currentSize -= old.size
	}

	// Evict if needed
	for c.currentSize+size > c.maxSize && len(c.items) > 0 {
		c.evictOldest()
	}

	c.items[key] = &entry{
		value:  value,
		expiry: time.Now().Add(c.ttl),
		size:   size,
	}
	c.currentSize += size
}

// Get retrieves item from cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(e.expiry) {
		return nil, false
	}

	return e.value, true
}

// Delete removes item from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, exists := c.items[key]; exists {
		c.currentSize -= e.size
		delete(c.items, key)
	}
}

// Clear removes all items
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*entry)
	c.currentSize = 0
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for k, e := range c.items {
		if oldestKey == "" || e.expiry.Before(oldestTime) {
			oldestKey = k
			oldestTime = e.expiry
		}
	}

	if oldestKey != "" {
		c.currentSize -= c.items[oldestKey].size
		delete(c.items, oldestKey)
	}
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.items {
			if now.After(e.expiry) {
				c.currentSize -= e.size
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// Stats returns cache statistics
func (c *Cache) Stats() (items, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items), c.currentSize
}
