package cache

import (
	"sync"
	"time"
)

type CacheStrategy string

const (
	StrategyLRU  CacheStrategy = "lru"
	StrategyLFU  CacheStrategy = "lfu"
	StrategyFIFO CacheStrategy = "fifo"
)

type StringCacheItem struct {
	key       string
	value     string
	frequency int
}

type MultiLayerCache struct {
	lru  *LRUCache
	lfu  *StringLFUCache
	fifo *StringFIFOCache
	mu   sync.RWMutex
}

func NewMultiLayerCache(maxSize int) *MultiLayerCache {
	return &MultiLayerCache{
		lru:  NewLRUCache(maxSize, 5*time.Minute),
		lfu:  NewStringLFUCache(maxSize),
		fifo: NewStringFIFOCache(maxSize),
	}
}

func (c *MultiLayerCache) GetString(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val := c.lru.Get(key); val != nil {
		if s, ok := val.(string); ok {
			return s, true
		}
	}
	if val, ok := c.lfu.Get(key); ok {
		return val, true
	}
	if val, ok := c.fifo.Get(key); ok {
		return val, true
	}
	return "", false
}

func (c *MultiLayerCache) SetString(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Set(key, value)
	c.lfu.Set(key, value)
	c.fifo.Set(key, value)
}

func (c *MultiLayerCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Delete(key)
	c.lfu.Delete(key)
	c.fifo.Delete(key)
}

func (c *MultiLayerCache) Stats() map[string]interface{} {
	lruStats := c.lru.Stats()
	return map[string]interface{}{
		"lru_entries": lruStats.Entries,
		"lru_hits":    lruStats.Hits,
		"lru_hitrate": lruStats.HitRate,
		"lfu_size":    c.lfu.Len(),
		"fifo_size":   c.fifo.Len(),
	}
}

type StringLFUCache struct {
	capacity int
	items    map[string]*StringCacheItem
}

func NewStringLFUCache(capacity int) *StringLFUCache {
	return &StringLFUCache{
		capacity: capacity,
		items:    make(map[string]*StringCacheItem),
	}
}

func (c *StringLFUCache) Get(key string) (string, bool) {
	if item, ok := c.items[key]; ok {
		item.frequency++
		return item.value, true
	}
	return "", false
}

func (c *StringLFUCache) Set(key, value string) {
	if item, ok := c.items[key]; ok {
		item.value = value
		item.frequency++
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	c.items[key] = &StringCacheItem{key: key, value: value, frequency: 1}
}

func (c *StringLFUCache) Delete(key string) {
	delete(c.items, key)
}

func (c *StringLFUCache) Len() int {
	return len(c.items)
}

func (c *StringLFUCache) evict() {
	var minKey string
	minFreq := int(^uint(0) >> 1)
	for key, item := range c.items {
		if item.frequency < minFreq {
			minFreq = item.frequency
			minKey = key
		}
	}
	if minKey != "" {
		delete(c.items, minKey)
	}
}

type StringFIFOCache struct {
	capacity int
	items    map[string]string
	order    []string
}

func NewStringFIFOCache(capacity int) *StringFIFOCache {
	return &StringFIFOCache{
		capacity: capacity,
		items:    make(map[string]string),
	}
}

func (c *StringFIFOCache) Get(key string) (string, bool) {
	val, ok := c.items[key]
	return val, ok
}

func (c *StringFIFOCache) Set(key, value string) {
	if _, ok := c.items[key]; !ok {
		if len(c.items) >= c.capacity {
			oldest := c.order[0]
			delete(c.items, oldest)
			c.order = c.order[1:]
		}
		c.order = append(c.order, key)
	}
	c.items[key] = value
}

func (c *StringFIFOCache) Delete(key string) {
	delete(c.items, key)
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

func (c *StringFIFOCache) Len() int {
	return len(c.items)
}
