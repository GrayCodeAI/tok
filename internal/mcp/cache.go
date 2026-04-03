// Package mcp provides hash-based caching for MCP context server.
package mcp

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultMaxSize is the default maximum cache entries.
	DefaultMaxSize = 10000

	// DefaultMaxMemory is the default maximum memory in bytes (100MB).
	DefaultMaxMemory = 100 * 1024 * 1024
)

// HashCache provides SHA-256 based content caching with LRU eviction.
type HashCache struct {
	mu         sync.RWMutex
	entries    map[string]*list.Element
	lru        *list.List
	maxSize    int
	maxMemory  int64
	currentMem int64

	// Statistics
	hits      int64
	misses    int64
	evictions int64

	// Callbacks
	onEvict func(entry *CacheEntry)
	onHit   func(entry *CacheEntry)
}

// cacheItem is the internal storage for cache entries.
type cacheItem struct {
	key   string
	entry *CacheEntry
	size  int64
}

// NewHashCache creates a new hash cache.
func NewHashCache(maxSize int, maxMemory int64) *HashCache {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	if maxMemory <= 0 {
		maxMemory = DefaultMaxMemory
	}

	return &HashCache{
		entries:   make(map[string]*list.Element),
		lru:       list.New(),
		maxSize:   maxSize,
		maxMemory: maxMemory,
	}
}

// SetEvictCallback sets a callback for evictions.
func (c *HashCache) SetEvictCallback(fn func(entry *CacheEntry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = fn
}

// SetHitCallback sets a callback for cache hits.
func (c *HashCache) SetHitCallback(fn func(entry *CacheEntry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onHit = fn
}

// Get retrieves an entry by hash.
func (c *HashCache) Get(hash string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[hash]
	if !ok {
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)

	item := elem.Value.(*cacheItem)
	item.entry.Accessed = time.Now()
	item.entry.HitCount++
	c.hits++

	if c.onHit != nil {
		c.onHit(item.entry)
	}

	return item.entry, true
}

// GetByContent retrieves an entry by computing hash of content.
func (c *HashCache) GetByContent(content string) (*CacheEntry, bool) {
	hash := ComputeHash(content)
	return c.Get(hash)
}

// Set adds or updates a cache entry.
func (c *HashCache) Set(entry *CacheEntry) error {
	if entry.Hash == "" {
		entry.Hash = ComputeHash(entry.Content)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size
	size := int64(len(entry.Content) + len(entry.FilePath) + 256) // overhead

	// Check if updating existing entry
	if elem, ok := c.entries[entry.Hash]; ok {
		item := elem.Value.(*cacheItem)
		c.currentMem -= item.size
		c.currentMem += size
		item.entry = entry
		item.size = size
		c.lru.MoveToFront(elem)
		return nil
	}

	// Evict if necessary
	for (len(c.entries) >= c.maxSize || c.currentMem+size > c.maxMemory) && c.lru.Len() > 0 {
		c.evictLRU()
	}

	// Add new entry
	item := &cacheItem{
		key:   entry.Hash,
		entry: entry,
		size:  size,
	}
	elem := c.lru.PushFront(item)
	c.entries[entry.Hash] = elem
	c.currentMem += size

	return nil
}

// evictLRU removes the least recently used entry.
func (c *HashCache) evictLRU() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}

	item := elem.Value.(*cacheItem)
	delete(c.entries, item.key)
	c.lru.Remove(elem)
	c.currentMem -= item.size
	c.evictions++

	if c.onEvict != nil {
		c.onEvict(item.entry)
	}
}

// Delete removes an entry by hash.
func (c *HashCache) Delete(hash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[hash]
	if !ok {
		return false
	}

	item := elem.Value.(*cacheItem)
	delete(c.entries, hash)
	c.lru.Remove(elem)
	c.currentMem -= item.size

	return true
}

// InvalidateByPattern removes entries matching a glob pattern.
func (c *HashCache) InvalidateByPattern(pattern string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple glob matching - could be enhanced
	var toDelete []string
	for hash, elem := range c.entries {
		item := elem.Value.(*cacheItem)
		if matchGlob(pattern, item.entry.FilePath) {
			toDelete = append(toDelete, hash)
		}
	}

	for _, hash := range toDelete {
		if elem, ok := c.entries[hash]; ok {
			item := elem.Value.(*cacheItem)
			delete(c.entries, hash)
			c.lru.Remove(elem)
			c.currentMem -= item.size
		}
	}

	return len(toDelete)
}

// Clear removes all entries.
func (c *HashCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lru = list.New()
	c.currentMem = 0
}

// Stats returns cache statistics.
func (c *HashCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var oldest, newest time.Time
	if c.lru.Len() > 0 {
		back := c.lru.Back()
		if back != nil {
			oldest = back.Value.(*cacheItem).entry.Timestamp
		}
		front := c.lru.Front()
		if front != nil {
			newest = front.Value.(*cacheItem).entry.Timestamp
		}
	}

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return Stats{
		TotalEntries: int64(len(c.entries)),
		TotalSize:    c.currentMem,
		HitRate:      hitRate,
		HitCount:     c.hits,
		MissCount:    c.misses,
		OldestEntry:  oldest,
		NewestEntry:  newest,
	}
}

// Len returns the number of entries.
func (c *HashCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// MemoryUsage returns current memory usage in bytes.
func (c *HashCache) MemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMem
}

// ComputeHash computes SHA-256 hash of content.
func ComputeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

// ComputeHashShort computes short SHA-256 hash (8 chars).
func ComputeHashShort(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:8])
}

// matchGlob performs simple glob matching.
func matchGlob(pattern, s string) bool {
	// Simple implementation - could use filepath.Match
	// For now, just check if pattern is contained
	return len(pattern) > 0 && len(s) > 0 &&
		(pattern == "*" || s == pattern ||
			(len(pattern) > 1 && pattern[0] == '*' && len(s) >= len(pattern)-1 &&
				s[len(s)-len(pattern)+1:] == pattern[1:]))
}

// BatchGet retrieves multiple entries efficiently.
func (c *HashCache) BatchGet(hashes []string) map[string]*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]*CacheEntry)
	for _, hash := range hashes {
		if elem, ok := c.entries[hash]; ok {
			item := elem.Value.(*cacheItem)
			results[hash] = item.entry
		}
	}
	return results
}

// BatchSet sets multiple entries efficiently.
func (c *HashCache) BatchSet(entries []*CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate total size
	var totalSize int64
	for _, entry := range entries {
		if entry.Hash == "" {
			entry.Hash = ComputeHash(entry.Content)
		}
		totalSize += int64(len(entry.Content) + len(entry.FilePath) + 256)
	}

	// Evict if necessary
	for (len(c.entries)+len(entries) > c.maxSize || c.currentMem+totalSize > c.maxMemory) && c.lru.Len() > 0 {
		c.evictLRU()
	}

	// Add all entries
	for _, entry := range entries {
		size := int64(len(entry.Content) + len(entry.FilePath) + 256)

		if elem, ok := c.entries[entry.Hash]; ok {
			// Update existing
			item := elem.Value.(*cacheItem)
			c.currentMem -= item.size
			c.currentMem += size
			item.entry = entry
			item.size = size
			c.lru.MoveToFront(elem)
		} else {
			// Add new
			item := &cacheItem{
				key:   entry.Hash,
				entry: entry,
				size:  size,
			}
			elem := c.lru.PushFront(item)
			c.entries[entry.Hash] = elem
			c.currentMem += size
		}
	}

	return nil
}

// GetOrCompute retrieves entry or computes and stores it.
func (c *HashCache) GetOrCompute(hash string, compute func() (*CacheEntry, error)) (*CacheEntry, error) {
	// Try to get from cache
	if entry, ok := c.Get(hash); ok {
		return entry, nil
	}

	// Compute
	entry, err := compute()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(entry); err != nil {
		return nil, fmt.Errorf("failed to cache entry: %w", err)
	}

	return entry, nil
}

// Touch updates access time for an entry without returning it.
func (c *HashCache) Touch(hash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[hash]
	if !ok {
		return false
	}

	c.lru.MoveToFront(elem)
	item := elem.Value.(*cacheItem)
	item.entry.Accessed = time.Now()
	item.entry.HitCount++

	return true
}

// Entries returns a snapshot of all cache entries.
// Use with caution on large caches.
func (c *HashCache) Entries() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CacheEntry, 0, len(c.entries))
	for elem := c.lru.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*cacheItem)
		entries = append(entries, item.entry)
	}
	return entries
}

// Resize changes the cache size limits.
func (c *HashCache) Resize(maxSize int, maxMemory int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxSize = maxSize
	c.maxMemory = maxMemory

	// Evict if over new limits
	for (len(c.entries) > c.maxSize || c.currentMem > c.maxMemory) && c.lru.Len() > 0 {
		c.evictLRU()
	}
}

// Persist saves cache metadata (not content) for analysis.
func (c *HashCache) Persist() ([]*CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return metadata only - content would be too large
	metadata := make([]*CacheEntry, 0, len(c.entries))
	for elem := c.lru.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*cacheItem)
		// Create copy without content
		meta := &CacheEntry{
			Hash:      item.entry.Hash,
			FilePath:  item.entry.FilePath,
			Timestamp: item.entry.Timestamp,
			Accessed:  item.entry.Accessed,
			HitCount:  item.entry.HitCount,
			// Content is empty - will be loaded on demand
		}
		metadata = append(metadata, meta)
	}

	return metadata, nil
}
