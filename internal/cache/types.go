package cache

import (
	"time"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key        string
	Value      []byte
	CreatedAt  time.Time
	AccessedAt time.Time
	ExpiresAt  time.Time
	Hits       int
	Size       int64
	Tags       []string
}

// IsExpired checks if the entry has expired
func (e *CacheEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// Touch updates access time and hit count
func (e *CacheEntry) Touch() {
	e.AccessedAt = time.Now()
	e.Hits++
}

// CacheStats tracks cache statistics
type CacheStats struct {
	TotalHits   int64
	TotalMisses int64
	L1Hits      int64
	L2Hits      int64
	L3Hits      int64
	Evictions   int64
	Size        int64
	Entries     int64
}

// RecordHit records a cache hit
func (cs *CacheStats) RecordHit(tier string) {
	cs.TotalHits++
	switch tier {
	case "l1":
		cs.L1Hits++
	case "l2":
		cs.L2Hits++
	case "l3":
		cs.L3Hits++
	}
}

// RecordMiss records a cache miss
func (cs *CacheStats) RecordMiss() {
	cs.TotalMisses++
}

// HitRate returns the overall hit rate
func (cs *CacheStats) HitRate() float64 {
	total := cs.TotalHits + cs.TotalMisses
	if total == 0 {
		return 0
	}
	return float64(cs.TotalHits) / float64(total)
}

// GetSnapshot returns a copy of the stats
func (cs *CacheStats) GetSnapshot() CacheStats {
	return *cs
}

// TierStats holds statistics for a single tier
type TierStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int64
	Entries   int64
	HitRate   float64
}
