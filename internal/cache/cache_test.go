package cache

import (
	"testing"
	"time"
)

func TestNewFingerprintCache(t *testing.T) {
	c := NewFingerprintCache(100, time.Hour)
	if c.maxEntries != 100 {
		t.Errorf("maxEntries = %d, want 100", c.maxEntries)
	}
	if c.ttl != time.Hour {
		t.Errorf("ttl = %v, want 1h", c.ttl)
	}
	if c.Size() != 0 {
		t.Errorf("initial size = %d, want 0", c.Size())
	}
}

func TestComputeFingerprint(t *testing.T) {
	h1 := ComputeFingerprint("hello")
	h2 := ComputeFingerprint("hello")
	h3 := ComputeFingerprint("world")

	if h1 != h2 {
		t.Error("same content should produce same fingerprint")
	}
	if h1 == h3 {
		t.Error("different content should produce different fingerprint")
	}
	if len(h1) != 16 {
		t.Errorf("fingerprint length = %d, want 16", len(h1))
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := NewFingerprintCache(10, time.Hour)

	c.Set("test content", "compressed", 42)

	result := c.Get("test content")
	if !result.Hit {
		t.Fatal("expected cache hit")
	}
	if result.Cached.TokensSaved != 42 {
		t.Errorf("tokensSaved = %d, want 42", result.Cached.TokensSaved)
	}
	if result.Cached.Compressed != "compressed" {
		t.Errorf("compressed = %q, want %q", result.Cached.Compressed, "compressed")
	}
}

func TestCacheMiss(t *testing.T) {
	c := NewFingerprintCache(10, time.Hour)

	result := c.Get("nonexistent")
	if result.Hit {
		t.Error("expected cache miss")
	}
}

func TestCacheEviction(t *testing.T) {
	c := NewFingerprintCache(3, time.Hour)

	// Fill cache to capacity
	c.Set("content1", "compressed1", 10)
	c.Set("content2", "compressed2", 20)
	c.Set("content3", "compressed3", 30)

	if c.Size() != 3 {
		t.Fatalf("size = %d, want 3", c.Size())
	}

	// Add one more - should evict oldest
	c.Set("content4", "compressed4", 40)

	if c.Size() != 3 {
		t.Errorf("size = %d, want 3 (after eviction)", c.Size())
	}

	// First entry should be evicted
	result := c.Get("content1")
	if result.Hit {
		t.Error("oldest entry should have been evicted")
	}

	// Last entry should still exist
	result = c.Get("content4")
	if !result.Hit {
		t.Error("newest entry should still exist")
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewFingerprintCache(10, 50*time.Millisecond)

	c.Set("expiring", "compressed", 10)

	// Should hit immediately
	result := c.Get("expiring")
	if !result.Hit {
		t.Fatal("expected hit before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	result = c.Get("expiring")
	if result.Hit {
		t.Error("expected miss after expiration")
	}
}

func TestCachePrune(t *testing.T) {
	c := NewFingerprintCache(10, 50*time.Millisecond)

	c.Set("item1", "c1", 1)
	c.Set("item2", "c2", 2)
	c.Set("item3", "c3", 3)

	time.Sleep(100 * time.Millisecond)

	pruned := c.Prune()
	if pruned != 3 {
		t.Errorf("pruned = %d, want 3", pruned)
	}
	if c.Size() != 0 {
		t.Errorf("size after prune = %d, want 0", c.Size())
	}
}

func TestCacheClear(t *testing.T) {
	c := NewFingerprintCache(10, time.Hour)

	c.Set("item1", "c1", 1)
	c.Set("item2", "c2", 2)

	if c.Size() != 2 {
		t.Fatalf("size = %d, want 2", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("size after clear = %d, want 0", c.Size())
	}

	stats := c.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("stats should be reset after clear")
	}
}

func TestCacheStats(t *testing.T) {
	c := NewFingerprintCache(10, time.Hour)

	c.Set("item1", "c1", 1)
	c.Get("item1") // hit
	c.Get("item1") // hit
	c.Get("missing") // miss

	stats := c.Stats()
	if stats.Entries != 1 {
		t.Errorf("entries = %d, want 1", stats.Entries)
	}
	if stats.Hits != 2 {
		t.Errorf("hits = %d, want 2", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("misses = %d, want 1", stats.Misses)
	}
	if stats.HitRate != 2.0/3.0 {
		t.Errorf("hitRate = %f, want %f", stats.HitRate, 2.0/3.0)
	}
}

func TestCacheZeroEntries(t *testing.T) {
	c := NewFingerprintCache(0, time.Hour)

	c.Set("item1", "c1", 1)

	// With maxEntries=0, the entry should still be stored
	// (eviction just won't happen until capacity is reached)
	result := c.Get("item1")
	if !result.Hit {
		t.Error("expected hit even with maxEntries=0")
	}
}

func TestGetGlobalCache(t *testing.T) {
	c1 := GetGlobalCache()
	c2 := GetGlobalCache()
	if c1 != c2 {
		t.Error("GetGlobalCache should return the same instance")
	}
}
