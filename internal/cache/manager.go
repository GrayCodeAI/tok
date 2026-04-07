package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CacheManager coordinates all cache tiers
type CacheManager struct {
	l1       *L1Cache
	l2       *L2Cache
	l3       *L3Cache
	config   CacheConfig
	stats    *CacheStats
	stopChan chan struct{}
	mu       sync.RWMutex
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled bool
	L1      L1Config
	L2      L2Config
	L3      L3Config
}

// L1Config holds L1 cache configuration
type L1Config struct {
	Enabled    bool
	MaxSize    int64
	MaxEntries int
	TTL        time.Duration
}

// L2Config holds L2 cache configuration
type L2Config struct {
	Enabled     bool
	MaxSize     int64
	MaxEntries  int
	TTL         time.Duration
	Compression bool
}

// L3Config holds L3 cache configuration
type L3Config struct {
	Enabled    bool
	MaxSize    int64
	MaxEntries int
	TTL        time.Duration
	RemoteURL  string
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled: true,
		L1: L1Config{
			Enabled:    true,
			MaxSize:    100 * 1024 * 1024, // 100MB
			MaxEntries: 10000,
			TTL:        5 * time.Minute,
		},
		L2: L2Config{
			Enabled:     true,
			MaxSize:     1024 * 1024 * 1024, // 1GB
			MaxEntries:  100000,
			TTL:         time.Hour,
			Compression: true,
		},
		L3: L3Config{
			Enabled:    false,
			MaxSize:    10 * 1024 * 1024 * 1024, // 10GB
			MaxEntries: 1000000,
			TTL:        24 * time.Hour,
		},
	}
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config CacheConfig) (*CacheManager, error) {
	cm := &CacheManager{
		config:   config,
		stats:    &CacheStats{},
		stopChan: make(chan struct{}),
	}

	// Initialize L1 cache
	if config.L1.Enabled {
		l1, err := NewL1Cache(config.L1)
		if err != nil {
			return nil, fmt.Errorf("failed to create L1 cache: %w", err)
		}
		cm.l1 = l1
	}

	// Initialize L2 cache
	if config.L2.Enabled {
		l2, err := NewL2Cache(config.L2)
		if err != nil {
			return nil, fmt.Errorf("failed to create L2 cache: %w", err)
		}
		cm.l2 = l2
	}

	// Initialize L3 cache
	if config.L3.Enabled {
		l3, err := NewL3Cache(config.L3)
		if err != nil {
			return nil, fmt.Errorf("failed to create L3 cache: %w", err)
		}
		cm.l3 = l3
	}

	// Start background tasks
	go cm.backgroundTasks()

	slog.Info("Cache manager initialized",
		"l1_enabled", config.L1.Enabled,
		"l2_enabled", config.L2.Enabled,
		"l3_enabled", config.L3.Enabled)

	return cm, nil
}

// Get retrieves an item from cache
func (cm *CacheManager) Get(ctx context.Context, key string) (*CacheEntry, bool) {
	// Try L1 first (fastest)
	if cm.l1 != nil {
		if entry, found := cm.l1.Get(key); found {
			cm.stats.RecordHit("l1")
			return entry, true
		}
	}

	// Try L2 (persistent)
	if cm.l2 != nil {
		if entry, found := cm.l2.Get(key); found {
			cm.stats.RecordHit("l2")
			// Promote to L1
			if cm.l1 != nil {
				cm.l1.Set(key, entry)
			}
			return entry, true
		}
	}

	// Try L3 (remote)
	if cm.l3 != nil {
		if entry, found := cm.l3.Get(key); found {
			cm.stats.RecordHit("l3")
			// Promote to L2 and L1
			if cm.l2 != nil {
				cm.l2.Set(key, entry)
			}
			if cm.l1 != nil {
				cm.l1.Set(key, entry)
			}
			return entry, true
		}
	}

	cm.stats.RecordMiss()
	return nil, false
}

// Set stores an item in cache
func (cm *CacheManager) Set(ctx context.Context, key string, entry *CacheEntry) error {
	// Store in all enabled tiers
	if cm.l1 != nil {
		if err := cm.l1.Set(key, entry); err != nil {
			return fmt.Errorf("L1 set failed: %w", err)
		}
	}

	if cm.l2 != nil {
		if err := cm.l2.Set(key, entry); err != nil {
			return fmt.Errorf("L2 set failed: %w", err)
		}
	}

	if cm.l3 != nil {
		if err := cm.l3.Set(key, entry); err != nil {
			return fmt.Errorf("L3 set failed: %w", err)
		}
	}

	return nil
}

// Delete removes an item from all cache tiers
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	if cm.l1 != nil {
		cm.l1.Delete(key)
	}

	if cm.l2 != nil {
		cm.l2.Delete(key)
	}

	if cm.l3 != nil {
		cm.l3.Delete(key)
	}

	return nil
}

// Invalidate removes items matching a pattern
func (cm *CacheManager) Invalidate(ctx context.Context, pattern string) error {
	if cm.l1 != nil {
		cm.l1.Invalidate(pattern)
	}

	if cm.l2 != nil {
		cm.l2.Invalidate(pattern)
	}

	if cm.l3 != nil {
		cm.l3.Invalidate(pattern)
	}

	return nil
}

// Clear clears all caches
func (cm *CacheManager) Clear(ctx context.Context) error {
	if cm.l1 != nil {
		cm.l1.Clear()
	}

	if cm.l2 != nil {
		cm.l2.Clear()
	}

	if cm.l3 != nil {
		cm.l3.Clear()
	}

	return nil
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() CacheStats {
	return cm.stats.GetSnapshot()
}

// GetTierStats returns statistics for each tier
func (cm *CacheManager) GetTierStats() map[string]TierStats {
	stats := make(map[string]TierStats)

	if cm.l1 != nil {
		stats["l1"] = cm.l1.GetStats()
	}

	if cm.l2 != nil {
		stats["l2"] = cm.l2.GetStats()
	}

	if cm.l3 != nil {
		stats["l3"] = cm.l3.GetStats()
	}

	return stats
}

// WarmCache pre-populates cache with frequently accessed items
func (cm *CacheManager) WarmCache(ctx context.Context, items map[string]*CacheEntry) error {
	for key, entry := range items {
		if err := cm.Set(ctx, key, entry); err != nil {
			slog.Warn("Failed to warm cache", "key", key, "error", err)
		}
	}

	slog.Info("Cache warming complete", "items", len(items))
	return nil
}

// Prefetch predicts and loads likely-needed items
func (cm *CacheManager) Prefetch(ctx context.Context, keys []string) {
	// This would typically use ML to predict which keys to prefetch
	// For now, just log the request
	slog.Debug("Prefetch requested", "keys", len(keys))
}

// backgroundTasks runs background maintenance tasks
func (cm *CacheManager) backgroundTasks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.performMaintenance()
		case <-cm.stopChan:
			return
		}
	}
}

// performMaintenance performs cache maintenance
func (cm *CacheManager) performMaintenance() {
	// Evict expired entries
	if cm.l1 != nil {
		cm.l1.EvictExpired()
	}

	if cm.l2 != nil {
		cm.l2.EvictExpired()
	}

	if cm.l3 != nil {
		cm.l3.EvictExpired()
	}

	// Log statistics
	stats := cm.GetStats()
	slog.Debug("Cache maintenance complete",
		"total_hits", stats.TotalHits,
		"total_misses", stats.TotalMisses,
		"hit_rate", stats.HitRate())
}

// Close closes the cache manager
func (cm *CacheManager) Close() error {
	close(cm.stopChan)

	if cm.l1 != nil {
		cm.l1.Close()
	}

	if cm.l2 != nil {
		cm.l2.Close()
	}

	if cm.l3 != nil {
		cm.l3.Close()
	}

	return nil
}
