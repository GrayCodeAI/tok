package core

import (
	"container/list"
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
	count int
	elem  *list.Element // pointer to list element for O(1) removal
}

// lruItem holds the key for list element tracking.
type lruItem struct {
	key string
}

// tokenCache caches BPE token counts for frequently seen strings.
// Phase 2.8: Avoids repeated BPE encoding for identical content.
// P2: Uses LRU eviction with doubly-linked list for O(1) operations.
type tokenCache struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
	ll    *list.List // doubly-linked list for O(1) LRU eviction
	max   int
	hits  int64 // For statistics
}

func newTokenCache(maxSize int) *tokenCache {
	return &tokenCache{
		items: make(map[string]*cacheEntry),
		ll:    list.New(),
		max:   maxSize,
	}
}

func (c *tokenCache) get(text string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[text]
	if !ok {
		return 0, false
	}
	c.ll.MoveToFront(entry.elem) // O(1) move to front
	atomic.AddInt64(&c.hits, 1)
	return entry.count, true
}

func (c *tokenCache) set(text string, count int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists (update in place)
	if entry, ok := c.items[text]; ok {
		entry.count = count
		c.ll.MoveToFront(entry.elem) // O(1) move to front
		return
	}

	// Evict LRU entries if at capacity
	if c.ll.Len() >= c.max {
		c.evictLRU()
	}

	elem := c.ll.PushFront(&lruItem{key: text}) // O(1) push front
	c.items[text] = &cacheEntry{
		count: count,
		elem:  elem,
	}
}

// evictLRU removes the least recently used entry (called with lock held).
// O(1) operation using doubly-linked list.
func (c *tokenCache) evictLRU() {
	// Remove 25% of entries (the oldest ones)
	toRemove := c.max / 4
	if toRemove < 1 {
		toRemove = 1
	}

	for i := 0; i < toRemove && c.ll.Len() > 0; i++ {
		back := c.ll.Back() // O(1) get oldest
		if back == nil {
			break
		}
		item := back.Value.(*lruItem)
		c.ll.Remove(back)         // O(1) remove from list
		delete(c.items, item.key) // O(1) remove from map
	}
}

var (
	bpeInstance *BPETokenizer
	bpeOnce     sync.Once
	bpeReady    atomic.Bool
	bpeErr      error
)

// getBPETokenizer returns the singleton BPE tokenizer, loading it if needed.
// The codec is loaded lazily but subsequent calls block only on initialization.
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
		bpeReady.Store(true)
	})
	return bpeInstance, bpeErr
}

// WarmupBPETokenizer preloads the codec in a background goroutine.
// Call this during application startup to avoid latency on the first
// token estimation. Safe to call multiple times.
func WarmupBPETokenizer() {
	go func() { _, _ = getBPETokenizer() }()
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
// Uses BPE tokenization when available and loaded, falls back to heuristic.
// Returns immediately with heuristic if BPE codec is still loading.
func EstimateTokens(text string) int {
	if useBPE.Load() && bpeReady.Load() {
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
