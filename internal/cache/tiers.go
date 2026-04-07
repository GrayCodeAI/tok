package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/compression"
)

// L1Cache is the fastest in-memory cache with LRU eviction
type L1Cache struct {
	maxSize     int64
	maxEntries  int
	ttl         time.Duration
	entries     map[string]*list.Element
	order       *list.List
	currentSize int64
	mu          sync.RWMutex
	hits        int64
	misses      int64
	evictions   int64
}

type l1Entry struct {
	key       string
	value     *CacheEntry
	createdAt time.Time
}

// NewL1Cache creates a new L1 cache
func NewL1Cache(config L1Config) (*L1Cache, error) {
	return &L1Cache{
		maxSize:    config.MaxSize,
		maxEntries: config.MaxEntries,
		ttl:        config.TTL,
		entries:    make(map[string]*list.Element),
		order:      list.New(),
	}, nil
}

// Get retrieves an entry from L1 cache
func (c *L1Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, found := c.entries[key]
	if !found {
		c.misses++
		return nil, false
	}

	entry := elem.Value.(*l1Entry)
	if entry.value.IsExpired() {
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	entry.value.Touch()
	c.order.MoveToFront(elem)
	c.hits++
	return entry.value, true
}

// Set stores an entry in L1 cache
func (c *L1Cache) Set(key string, entry *CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if elem, found := c.entries[key]; found {
		c.order.MoveToFront(elem)
		elem.Value.(*l1Entry).value = entry
		return nil
	}

	// Evict if necessary
	for c.currentSize+entry.Size > c.maxSize || len(c.entries) >= c.maxEntries {
		if !c.evictLRU() {
			break
		}
	}

	// Add new entry
	l1Entry := &l1Entry{
		key:       key,
		value:     entry,
		createdAt: time.Now(),
	}
	elem := c.order.PushFront(l1Entry)
	c.entries[key] = elem
	c.currentSize += entry.Size

	return nil
}

// Delete removes an entry
func (c *L1Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.entries[key]; found {
		c.removeElement(elem)
	}
}

// Invalidate removes entries matching a pattern
func (c *L1Cache) Invalidate(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, elem := range c.entries {
		if matchPattern(key, pattern) {
			c.removeElement(elem)
			delete(c.entries, key)
		}
	}
}

// Clear clears all entries
func (c *L1Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.order = list.New()
	c.currentSize = 0
}

// EvictExpired removes expired entries
func (c *L1Cache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, elem := range c.entries {
		entry := elem.Value.(*l1Entry)
		if entry.value.IsExpired() {
			c.removeElement(elem)
			delete(c.entries, key)
		}
	}
}

// GetStats returns tier statistics
func (c *L1Cache) GetStats() TierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return TierStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.currentSize,
		Entries:   int64(len(c.entries)),
		HitRate:   hitRate,
	}
}

// Close closes the cache
func (c *L1Cache) Close() error {
	return nil
}

func (c *L1Cache) evictLRU() bool {
	elem := c.order.Back()
	if elem == nil {
		return false
	}

	c.removeElement(elem)
	c.evictions++
	return true
}

func (c *L1Cache) removeElement(elem *list.Element) {
	entry := elem.Value.(*l1Entry)
	c.currentSize -= entry.value.Size
	c.order.Remove(elem)
	delete(c.entries, entry.key)
}

// L2Cache is the persistent disk-based cache with LFU eviction
type L2Cache struct {
	maxSize     int64
	maxEntries  int
	ttl         time.Duration
	compression bool
	dir         string
	entries     map[string]*l2Entry
	mu          sync.RWMutex
	hits        int64
	misses      int64
	evictions   int64
	currentSize int64
}

type l2Entry struct {
	key       string
	value     *CacheEntry
	createdAt time.Time
	frequency int
}

// NewL2Cache creates a new L2 cache
func NewL2Cache(config L2Config) (*L2Cache, error) {
	dataDir := os.Getenv("TOKMAN_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "tokman")
	}

	cacheDir := filepath.Join(dataDir, "cache", "l2")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create L2 cache directory: %w", err)
	}

	cache := &L2Cache{
		maxSize:     config.MaxSize,
		maxEntries:  config.MaxEntries,
		ttl:         config.TTL,
		compression: config.Compression,
		dir:         cacheDir,
		entries:     make(map[string]*l2Entry),
	}

	// Load existing entries
	cache.loadEntries()

	return cache, nil
}

// Get retrieves an entry from L2 cache
func (c *L2Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.entries[key]
	if !found {
		c.misses++
		return nil, false
	}

	if entry.value.IsExpired() {
		c.deleteEntry(key)
		c.misses++
		return nil, false
	}

	entry.frequency++
	entry.value.Touch()
	c.hits++
	return entry.value, true
}

// Set stores an entry in L2 cache
func (c *L2Cache) Set(key string, entry *CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if necessary
	for c.currentSize+entry.Size > c.maxSize || len(c.entries) >= c.maxEntries {
		if !c.evictLFU() {
			break
		}
	}

	// Add entry
	c.entries[key] = &l2Entry{
		key:       key,
		value:     entry,
		createdAt: time.Now(),
		frequency: 1,
	}
	c.currentSize += entry.Size

	// Persist to disk
	return c.persistEntry(key, c.entries[key])
}

// Delete removes an entry
func (c *L2Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteEntry(key)
}

// Invalidate removes entries matching a pattern
func (c *L2Cache) Invalidate(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if matchPattern(key, pattern) {
			c.deleteEntry(key)
			_ = entry
		}
	}
}

// Clear clears all entries
func (c *L2Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.entries {
		c.deleteEntry(key)
	}
	c.currentSize = 0
}

// EvictExpired removes expired entries
func (c *L2Cache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.value.IsExpired() {
			c.deleteEntry(key)
		}
	}
}

// GetStats returns tier statistics
func (c *L2Cache) GetStats() TierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return TierStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.currentSize,
		Entries:   int64(len(c.entries)),
		HitRate:   hitRate,
	}
}

// Close closes the cache
func (c *L2Cache) Close() error {
	return nil
}

func (c *L2Cache) evictLFU() bool {
	var minKey string
	minFreq := int(^uint(0) >> 1)

	for key, entry := range c.entries {
		if entry.frequency < minFreq {
			minFreq = entry.frequency
			minKey = key
		}
	}

	if minKey != "" {
		c.deleteEntry(minKey)
		c.evictions++
		return true
	}

	return false
}

func (c *L2Cache) deleteEntry(key string) {
	if entry, found := c.entries[key]; found {
		c.currentSize -= entry.value.Size
		delete(c.entries, key)
		os.Remove(c.entryPath(key))
	}
}

func (c *L2Cache) entryPath(key string) string {
	return filepath.Join(c.dir, key+".json")
}

func (c *L2Cache) persistEntry(key string, entry *l2Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if c.compression {
		compressor := compression.NewBrotliCompressor()
		data, err = compressor.Compress(data)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(c.entryPath(key), data, 0644)
}

func (c *L2Cache) loadEntries() {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.dir, file.Name()))
		if err != nil {
			continue
		}

		var entry l2Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		c.entries[entry.key] = &entry
		c.currentSize += entry.value.Size
	}
}

// L3Cache is the remote/distributed cache with FIFO eviction
type L3Cache struct {
	maxSize     int64
	maxEntries  int
	ttl         time.Duration
	remoteURL   string
	entries     map[string]*l3Entry
	order       []string
	mu          sync.RWMutex
	hits        int64
	misses      int64
	evictions   int64
	currentSize int64
}

type l3Entry struct {
	key       string
	value     *CacheEntry
	createdAt time.Time
}

// NewL3Cache creates a new L3 cache
func NewL3Cache(config L3Config) (*L3Cache, error) {
	return &L3Cache{
		maxSize:    config.MaxSize,
		maxEntries: config.MaxEntries,
		ttl:        config.TTL,
		remoteURL:  config.RemoteURL,
		entries:    make(map[string]*l3Entry),
		order:      make([]string, 0),
	}, nil
}

// Get retrieves an entry from L3 cache
func (c *L3Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.entries[key]
	if !found {
		c.misses++
		return nil, false
	}

	if entry.value.IsExpired() {
		c.deleteEntry(key)
		c.misses++
		return nil, false
	}

	entry.value.Touch()
	c.hits++
	return entry.value, true
}

// Set stores an entry in L3 cache
func (c *L3Cache) Set(key string, entry *CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if necessary
	for c.currentSize+entry.Size > c.maxSize || len(c.entries) >= c.maxEntries {
		if !c.evictFIFO() {
			break
		}
	}

	c.entries[key] = &l3Entry{
		key:       key,
		value:     entry,
		createdAt: time.Now(),
	}
	c.order = append(c.order, key)
	c.currentSize += entry.Size

	return nil
}

// Delete removes an entry
func (c *L3Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteEntry(key)
}

// Invalidate removes entries matching a pattern
func (c *L3Cache) Invalidate(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.entries {
		if matchPattern(key, pattern) {
			c.deleteEntry(key)
		}
	}
}

// Clear clears all entries
func (c *L3Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.entries {
		c.deleteEntry(key)
	}
	c.currentSize = 0
}

// EvictExpired removes expired entries
func (c *L3Cache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.value.IsExpired() {
			c.deleteEntry(key)
		}
	}
}

// GetStats returns tier statistics
func (c *L3Cache) GetStats() TierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return TierStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.currentSize,
		Entries:   int64(len(c.entries)),
		HitRate:   hitRate,
	}
}

// Close closes the cache
func (c *L3Cache) Close() error {
	return nil
}

func (c *L3Cache) evictFIFO() bool {
	if len(c.order) == 0 {
		return false
	}

	key := c.order[0]
	c.deleteEntry(key)
	c.order = c.order[1:]
	c.evictions++
	return true
}

func (c *L3Cache) deleteEntry(key string) {
	if entry, found := c.entries[key]; found {
		c.currentSize -= entry.value.Size
		delete(c.entries, key)
	}
}

// matchPattern checks if a key matches a pattern (simple glob)
func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == key {
		return true
	}
	// Simple prefix match
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	return false
}
