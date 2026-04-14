package cache

import (
	"container/list"
	"sync"
	"time"
)

var (
	globalCache     *FingerprintCache
	globalCacheOnce sync.Once
)

func GetGlobalCache() *FingerprintCache {
	globalCacheOnce.Do(func() {
		globalCache = NewFingerprintCache()
	})
	return globalCache
}

type FingerprintCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewFingerprintCache() *FingerprintCache {
	return &FingerprintCache{
		cache: make(map[string]string),
	}
}

func (c *FingerprintCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

func (c *FingerprintCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}

// Cache interface defines a generic cache
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Len() int
}

// LRUCache implements a true LRU cache with O(1) operations
type LRUCache struct {
	mu    sync.Mutex
	cap   int
	ttl   time.Duration
	cache map[string]*cacheEntry
	order *list.List
}

type cacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
	element   *list.Element
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cap:   capacity,
		ttl:   ttl,
		cache: make(map[string]*cacheEntry),
		order: list.New(),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Check TTL
	if time.Now().After(entry.expiresAt) {
		c.removeEntry(entry)
		return nil, false
	}

	// Move to front (true LRU promotion)
	c.order.MoveToBack(entry.element)

	return entry.value, true
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing key
	if entry, ok := c.cache[key]; ok {
		entry.value = value
		entry.expiresAt = time.Now().Add(c.ttl)
		c.order.MoveToBack(entry.element)
		return
	}

	// Evict if at capacity
	if len(c.cache) >= c.cap {
		c.evictOldest()
	}

	// Add new entry
	element := c.order.PushBack(key)
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
		element:   element,
	}
	c.cache[key] = entry
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.cache[key]; ok {
		c.removeEntry(entry)
	}
}

func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cache)
}

func (c *LRUCache) evictOldest() {
	front := c.order.Front()
	if front == nil {
		return
	}

	key := front.Value.(string)
	if entry, ok := c.cache[key]; ok {
		c.removeEntry(entry)
	}
}

func (c *LRUCache) removeEntry(entry *cacheEntry) {
	c.order.Remove(entry.element)
	delete(c.cache, entry.key)
}
