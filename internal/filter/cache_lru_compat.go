// Package filter provides LRU caching using the unified cache package.
// This file provides backward compatibility for existing code.
package filter

import (
	"time"

	"github.com/lakshmanpatel/tok/internal/cache"
)

// LRUCache is an alias to the unified LRU cache implementation.
// Note: The cache stores *CachedResult values defined locally in manager.go.
type LRUCache = cache.LRUCache

// NewLRUCache creates an LRU cache with given max size and TTL.
func NewLRUCache(maxSize int, ttl time.Duration) *cache.LRUCache {
	return cache.NewLRUCache(maxSize, ttl)
}
