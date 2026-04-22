package filter

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"
	"time"
)

// LayerCacheEntry stores cached filter results.
type LayerCacheEntry struct {
	InputHash   string
	Output      string
	TokensSaved int
	LayerName   string
	Mode        Mode
	Timestamp   time.Time
	HitCount    int
}

// LayerCache provides content-addressable caching for filter results.
// Uses SHA-256 hashing for cache keys with LRU eviction.
type LayerCache struct {
	mu        sync.RWMutex
	items     map[string]*LayerCacheEntry
	maxSize   int
	ttl       time.Duration
	hits      int64
	misses    int64
	evictions int64
}

// NewLayerCache creates a new layer cache.
func NewLayerCache(maxSize int, ttl time.Duration) *LayerCache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &LayerCache{
		items:   make(map[string]*LayerCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// computeHash generates a SHA-256 hash of the input.
func computeHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// makeKey creates a cache key from layer name, input hash, and mode.
func makeKey(layerName, inputHash string, mode Mode) string {
	return layerName + ":" + inputHash + ":" + string(mode)
}

// Get retrieves a cached result if available and not expired.
// Uses a single write lock to avoid the race window between RUnlock and Lock
// during hit-count promotion.
func (c *LayerCache) Get(layerName, input string, mode Mode) (*LayerCacheEntry, bool) {
	if c == nil {
		return nil, false
	}

	inputHash := computeHash(input)
	key := makeKey(layerName, inputHash, mode)

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.items[key]
	if !found {
		c.misses++
		return nil, false
	}

	// Check TTL
	if time.Since(entry.Timestamp) > c.ttl {
		delete(c.items, key)
		c.misses++
		return nil, false
	}

	// Update hit count
	entry.HitCount++
	c.hits++
	return entry, true
}

// Put stores a result in the cache.
func (c *LayerCache) Put(layerName, input string, mode Mode, output string, tokensSaved int) {
	if c == nil || len(input) < 100 {
		// Don't cache very small inputs
		return
	}

	inputHash := computeHash(input)
	key := makeKey(layerName, inputHash, mode)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at capacity
	if len(c.items) >= c.maxSize {
		c.evictOldest(100) // Evict 10% of entries
	}

	c.items[key] = &LayerCacheEntry{
		InputHash:   inputHash,
		Output:      output,
		TokensSaved: tokensSaved,
		LayerName:   layerName,
		Mode:        mode,
		Timestamp:   time.Now(),
		HitCount:    1,
	}
}

// evictOldest removes the oldest N entries from the cache.
func (c *LayerCache) evictOldest(n int) {
	type kv struct {
		key   string
		value *LayerCacheEntry
	}

	// Convert map to slice for sorting
	kvs := make([]kv, 0, len(c.items))
	for k, v := range c.items {
		kvs = append(kvs, kv{k, v})
	}

	// Sort by timestamp (oldest first) — O(n log n)
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].value.Timestamp.Before(kvs[j].value.Timestamp)
	})

	// Remove oldest n entries
	toRemove := n
	if toRemove > len(kvs) {
		toRemove = len(kvs)
	}
	for i := 0; i < toRemove; i++ {
		delete(c.items, kvs[i].key)
		c.evictions++
	}
}

// Stats returns cache statistics.
func (c *LayerCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return CacheStats{
		Size:      len(c.items),
		MaxSize:   c.maxSize,
		Hits:      c.hits,
		Misses:    c.misses,
		HitRate:   hitRate,
		Evictions: c.evictions,
	}
}

// Clear removes all entries from the cache.
func (c *LayerCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*LayerCacheEntry)
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// CacheStats holds cache performance statistics.
type CacheStats struct {
	Size      int
	MaxSize   int
	Hits      int64
	Misses    int64
	HitRate   float64
	Evictions int64
}

// Global layer cache instance (optional, per-pipeline caches also supported)
var (
	globalLayerCache *LayerCache
	initGlobalCache  sync.Once
)

// GetGlobalLayerCache returns the global layer cache instance.
func GetGlobalLayerCache() *LayerCache {
	initGlobalCache.Do(func() {
		globalLayerCache = NewLayerCache(5000, 5*time.Minute)
	})
	return globalLayerCache
}
