package filecache

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	Fingerprint string    `json:"fingerprint"`
	Compressed  string    `json:"compressed"`
	Tokens      int       `json:"tokens"`
	HitCount    int       `json:"hit_count"`
	CreatedAt   time.Time `json:"created_at"`
	LastAccess  time.Time `json:"last_access"`
}

type FileCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	maxSize int
	hits    int64
	misses  int64
}

func NewFileCache(maxSize int) *FileCache {
	if maxSize == 0 {
		maxSize = 1000
	}
	return &FileCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

func (c *FileCache) fingerprint(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8])
}

func (c *FileCache) Get(content string) (string, bool) {
	fp := c.fingerprint(content)
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[fp]; ok {
		entry.HitCount++
		entry.LastAccess = time.Now()
		c.hits++
		return entry.Compressed, true
	}
	c.misses++
	return "", false
}

func (c *FileCache) Set(content, compressed string) {
	fp := c.fingerprint(content)
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evict()
	}

	c.entries[fp] = &CacheEntry{
		Fingerprint: fp,
		Compressed:  compressed,
		Tokens:      len(content) / 4,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
	}
}

func (c *FileCache) evict() {
	var oldestKey string
	var oldestTime time.Time
	for k, v := range c.entries {
		if oldestKey == "" || v.LastAccess.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.LastAccess
		}
	}
	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *FileCache) HitRate() float64 {
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total) * 100
}

func (c *FileCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"entries":  len(c.entries),
		"hits":     c.hits,
		"misses":   c.misses,
		"hit_rate": c.HitRate(),
		"max_size": c.maxSize,
	}
}

func (c *FileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
	c.hits = 0
	c.misses = 0
}

func (c *FileCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
