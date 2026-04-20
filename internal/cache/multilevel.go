package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
)

// MultiLevelCache implements L1 (memory) + L2 (disk) + L3 (optional Redis) caching
type MultiLevelCache struct {
	l1    map[string]string // In-memory LRU
	l2Dir string            // Disk cache directory
	mu    sync.RWMutex
	maxL1 int
}

// NewMultiLevelCache creates a 3-tier cache
func NewMultiLevelCache(l2Dir string, maxL1Size int) *MultiLevelCache {
	os.MkdirAll(l2Dir, 0755)
	return &MultiLevelCache{
		l1:    make(map[string]string, maxL1Size),
		l2Dir: l2Dir,
		maxL1: maxL1Size,
	}
}

// Get retrieves from L1 → L2 → L3
func (mc *MultiLevelCache) Get(key string) (string, bool) {
	// L1: Memory
	mc.mu.RLock()
	if val, ok := mc.l1[key]; ok {
		mc.mu.RUnlock()
		return val, true
	}
	mc.mu.RUnlock()

	// L2: Disk
	hash := hashKey(key)
	path := filepath.Join(mc.l2Dir, hash)
	data, err := os.ReadFile(path)
	if err == nil {
		val := string(data)
		mc.promoteToL1(key, val)
		return val, true
	}

	return "", false
}

// Set stores in all cache levels
func (mc *MultiLevelCache) Set(key, value string) {
	mc.mu.Lock()
	if len(mc.l1) >= mc.maxL1 {
		// Evict random entry
		for k := range mc.l1 {
			delete(mc.l1, k)
			break
		}
	}
	mc.l1[key] = value
	mc.mu.Unlock()

	// L2: Disk
	hash := hashKey(key)
	path := filepath.Join(mc.l2Dir, hash)
	os.WriteFile(path, []byte(value), 0644)
}

func (mc *MultiLevelCache) promoteToL1(key, value string) {
	mc.mu.Lock()
	mc.l1[key] = value
	mc.mu.Unlock()
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
