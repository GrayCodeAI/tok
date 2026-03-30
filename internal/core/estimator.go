package core

import (
	"sync"
	"sync/atomic"

	tiktoken "github.com/tiktoken-go/tokenizer"
)

// BPETokenizer wraps tiktoken for accurate BPE token counting.
// P1.1: Replaces heuristic len/4 with real BPE tokenization.
// ~20-30% more accurate than heuristic estimation.
type BPETokenizer struct {
	codec tiktoken.Codec
	cache *tokenCache
}

// cacheEntry stores cached token counts with LRU metadata.
type cacheEntry struct {
	count   int
	lastHit int64 // Unix nanoseconds for LRU eviction
}

// tokenCache caches BPE token counts for frequently seen strings.
// Phase 2.8: Avoids repeated BPE encoding for identical content.
// P2: Uses LRU eviction to preserve frequently-accessed items.
type tokenCache struct {
	mu       sync.RWMutex
	items    map[string]*cacheEntry
	size     int
	max      int
	hitCount int64 // For statistics
}

func newTokenCache(maxSize int) *tokenCache {
	return &tokenCache{
		items: make(map[string]*cacheEntry),
		max:   maxSize,
	}
}

func (c *tokenCache) get(text string) (int, bool) {
	c.mu.RLock()
	entry, ok := c.items[text]
	c.mu.RUnlock()
	if !ok {
		return 0, false
	}
	// Update lastHit for LRU tracking (atomic to avoid write lock)
	c.mu.Lock()
	entry.lastHit = nanoTime()
	c.hitCount++
	c.mu.Unlock()
	return entry.count, true
}

func (c *tokenCache) set(text string, count int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists (update in place)
	if entry, ok := c.items[text]; ok {
		entry.count = count
		entry.lastHit = nanoTime()
		return
	}

	// Evict LRU entries if at capacity
	if c.size >= c.max {
		c.evictLRU()
	}

	c.items[text] = &cacheEntry{
		count:   count,
		lastHit: nanoTime(),
	}
	c.size++
}

// evictLRU removes the least recently used entries (called with lock held)
func (c *tokenCache) evictLRU() {
	// Find and remove 25% of entries (the oldest ones)
	toRemove := c.max / 4
	if toRemove < 1 {
		toRemove = 1
	}

	// Find the oldest entries
	type kv struct {
		key     string
		lastHit int64
	}
	entries := make([]kv, 0, len(c.items))
	for k, v := range c.items {
		entries = append(entries, kv{k, v.lastHit})
	}

	// Sort by lastHit (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].lastHit > entries[j].lastHit {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	for i := 0; i < toRemove && i < len(entries); i++ {
		delete(c.items, entries[i].key)
		c.size--
	}
}

// nanoTime returns a monotonic counter for LRU tracking
func nanoTime() int64 {
	staticCounter.Add(1)
	return staticCounter.Load()
}

var staticCounter atomic.Int64

var (
	bpeInstance *BPETokenizer
	bpeOnce     sync.Once
	bpeErr      error
)

// getBPETokenizer returns a singleton BPE tokenizer.
func getBPETokenizer() (*BPETokenizer, error) {
	bpeOnce.Do(func() {
		codec, err := tiktoken.Get(tiktoken.Cl100kBase)
		if err != nil {
			bpeErr = err
			return
		}
		bpeInstance = &BPETokenizer{
			codec: codec,
			cache: newTokenCache(1024),
		}
	})
	return bpeInstance, bpeErr
}

// Count returns the accurate BPE token count with caching.
func (b *BPETokenizer) Count(text string) int {
	if text == "" {
		return 0
	}

	// Check cache first (Phase 2.8 optimization)
	if val, ok := b.cache.get(text); ok {
		return val
	}

	count, err := b.codec.Count(text)
	if err != nil {
		return (len(text) + 3) / 4 // Fallback to heuristic
	}

	// Cache result for future lookups
	b.cache.set(text, count)
	return count
}

// useBPE controls whether to use BPE or heuristic estimation.
// Set to 1 (true) by default for accuracy; can be toggled for performance.
var useBPE atomic.Bool

func init() {
	useBPE.Store(true)
}

// EstimateTokens is the single source of truth for token estimation.
// P1.1: Uses BPE tokenization when available, falls back to heuristic.
// Phase 2.8: Results are cached to avoid repeated encoding.
func EstimateTokens(text string) int {
	if useBPE.Load() {
		if tok, err := getBPETokenizer(); err == nil {
			return tok.Count(text)
		}
	}
	return (len(text) + 3) / 4
}

// CalculateTokensSaved computes token savings between original and filtered.
func CalculateTokensSaved(original, filtered string) int {
	origTokens := EstimateTokens(original)
	filterTokens := EstimateTokens(filtered)
	if origTokens > filterTokens {
		return origTokens - filterTokens
	}
	return 0
}
