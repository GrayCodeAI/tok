package cache

import (
	"sync"
	"time"
)

var globalCache *FingerprintCache

func GetGlobalCache() *FingerprintCache {
	if globalCache == nil {
		globalCache = NewFingerprintCache()
	}
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

type LRUCache struct {
	mu    sync.Mutex
	cap   int
	ttl   time.Duration
	cache map[string]interface{}
	order []string
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cap:   capacity,
		ttl:   ttl,
		cache: make(map[string]interface{}),
		order: make([]string, 0, capacity),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.cache[key]
	return val, ok
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[key]; ok {
		return
	}
	if len(c.cache) >= c.cap {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, oldest)
	}
	c.cache[key] = value
	c.order = append(c.order, key)
}


