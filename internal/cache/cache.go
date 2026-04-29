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

const (
	defaultFingerprintCacheMaxSize = 10000
	defaultFingerprintCacheTTL     = 10 * time.Minute
)

type fingerprintEntry struct {
	key       string
	value     string
	expiresAt time.Time
	element   *list.Element
}

type FingerprintCache struct {
	mu      sync.RWMutex
	cache   map[string]*fingerprintEntry
	order   *list.List
	maxSize int
	ttl     time.Duration
}

func NewFingerprintCache() *FingerprintCache {
	return &FingerprintCache{
		cache:   make(map[string]*fingerprintEntry),
		order:   list.New(),
		maxSize: defaultFingerprintCacheMaxSize,
		ttl:     defaultFingerprintCacheTTL,
	}
}

func (c *FingerprintCache) Get(key string) (string, bool) {
	// Fast path: read-only check under RLock.
	c.mu.RLock()
	entry, ok := c.cache[key]
	if !ok {
		c.mu.RUnlock()
		return "", false
	}
	now := time.Now()
	if now.After(entry.expiresAt) {
		c.mu.RUnlock()
		// Expired — need write lock to remove.
		c.mu.Lock()
		if e, stillThere := c.cache[key]; stillThere && now.After(e.expiresAt) {
			c.removeEntry(e)
		}
		c.mu.Unlock()
		return "", false
	}
	value := entry.value
	c.mu.RUnlock()

	// Promotion requires mutation — acquire write lock.
	c.mu.Lock()
	if e, stillThere := c.cache[key]; stillThere && !now.After(e.expiresAt) {
		c.order.MoveToBack(e.element)
	}
	c.mu.Unlock()
	return value, true
}

func (c *FingerprintCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Update existing key
	if entry, ok := c.cache[key]; ok {
		entry.value = value
		entry.expiresAt = now.Add(c.ttl)
		c.order.MoveToBack(entry.element)
		return
	}

	// Proactive expiration: remove expired entries at the front (oldest)
	for front := c.order.Front(); front != nil; {
		k := front.Value.(string)
		e, ok := c.cache[k]
		if !ok || now.After(e.expiresAt) {
			next := front.Next()
			c.removeEntry(e)
			front = next
			continue
		}
		break
	}

	// Evict oldest entry if still at capacity
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	element := c.order.PushBack(key)
	c.cache[key] = &fingerprintEntry{
		key:       key,
		value:     value,
		expiresAt: now.Add(c.ttl),
		element:   element,
	}
}

func (c *FingerprintCache) removeEntry(entry *fingerprintEntry) {
	c.order.Remove(entry.element)
	delete(c.cache, entry.key)
}

func (c *FingerprintCache) evictOldest() {
	front := c.order.Front()
	if front == nil {
		return
	}
	key := front.Value.(string)
	if entry, ok := c.cache[key]; ok {
		c.removeEntry(entry)
	}
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
	mu    sync.RWMutex
	cap   int
	ttl   time.Duration
	cache map[string]*lruCacheEntry
	order *list.List
}

type lruCacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
	element   *list.Element
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cap:   capacity,
		ttl:   ttl,
		cache: make(map[string]*lruCacheEntry),
		order: list.New(),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	// Fast path: read-only check under RLock.
	c.mu.RLock()
	entry, ok := c.cache[key]
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}
	now := time.Now()
	if now.After(entry.expiresAt) {
		c.mu.RUnlock()
		// Expired — need write lock to remove.
		c.mu.Lock()
		if e, stillThere := c.cache[key]; stillThere && now.After(e.expiresAt) {
			c.removeEntry(e)
		}
		c.mu.Unlock()
		return nil, false
	}
	value := entry.value
	c.mu.RUnlock()

	// Promotion requires mutation — acquire write lock.
	c.mu.Lock()
	if e, stillThere := c.cache[key]; stillThere && !now.After(e.expiresAt) {
		c.order.MoveToBack(e.element)
	}
	c.mu.Unlock()
	return value, true
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
	entry := &lruCacheEntry{
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

func (c *LRUCache) removeEntry(entry *lruCacheEntry) {
	c.order.Remove(entry.element)
	delete(c.cache, entry.key)
}
