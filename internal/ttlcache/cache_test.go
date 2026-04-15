package ttlcache

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cache := New(time.Minute, 1000)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.ttl != time.Minute {
		t.Errorf("expected ttl=1m, got %v", cache.ttl)
	}
	if cache.maxSize != 1000 {
		t.Errorf("expected maxSize=1000, got %d", cache.maxSize)
	}
	if cache.items == nil {
		t.Error("expected initialized items map")
	}
}

func TestSetAndGet(t *testing.T) {
	cache := New(time.Minute, 1000)

	cache.Set("key1", "value1", 100)

	val, found := cache.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestGet_NotFound(t *testing.T) {
	cache := New(time.Minute, 1000)

	val, found := cache.Get("nonexistent")
	if found {
		t.Error("expected not to find nonexistent key")
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestGet_Expired(t *testing.T) {
	cache := New(1*time.Millisecond, 1000)

	cache.Set("key1", "value1", 100)
	time.Sleep(2 * time.Millisecond)

	val, found := cache.Get("key1")
	if found {
		t.Error("expected expired key to not be found")
	}
	if val != nil {
		t.Errorf("expected nil for expired key, got %v", val)
	}
}

func TestSet_Update(t *testing.T) {
	cache := New(time.Minute, 1000)

	cache.Set("key1", "value1", 100)
	cache.Set("key1", "value2", 200)

	val, found := cache.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != "value2" {
		t.Errorf("expected 'value2', got %v", val)
	}

	// Check size was updated correctly
	_, size := cache.Stats()
	if size != 200 {
		t.Errorf("expected size=200, got %d", size)
	}
}

func TestDelete(t *testing.T) {
	cache := New(time.Minute, 1000)

	cache.Set("key1", "value1", 100)
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key to be deleted")
	}

	// Check size was updated
	_, size := cache.Stats()
	if size != 0 {
		t.Errorf("expected size=0 after delete, got %d", size)
	}
}

func TestDelete_Nonexistent(t *testing.T) {
	cache := New(time.Minute, 1000)

	// Should not panic
	cache.Delete("nonexistent")

	_, size := cache.Stats()
	if size != 0 {
		t.Errorf("expected size=0, got %d", size)
	}
}

func TestClear(t *testing.T) {
	cache := New(time.Minute, 1000)

	cache.Set("key1", "value1", 100)
	cache.Set("key2", "value2", 200)
	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("expected all keys to be cleared")
	}

	items, size := cache.Stats()
	if items != 0 {
		t.Errorf("expected items=0, got %d", items)
	}
	if size != 0 {
		t.Errorf("expected size=0, got %d", size)
	}
}

func TestEviction(t *testing.T) {
	cache := New(time.Minute, 100)

	// Add items that exceed maxSize
	cache.Set("key1", "value1", 40)
	cache.Set("key2", "value2", 40)
	cache.Set("key3", "value3", 40) // Should trigger eviction

	items, size := cache.Stats()
	if items > 2 {
		t.Errorf("expected at most 2 items after eviction, got %d", items)
	}
	if size > 100 {
		t.Errorf("expected size <= 100 after eviction, got %d", size)
	}
}

func TestStats(t *testing.T) {
	cache := New(time.Minute, 1000)

	items, size := cache.Stats()
	if items != 0 {
		t.Errorf("expected items=0, got %d", items)
	}
	if size != 0 {
		t.Errorf("expected size=0, got %d", size)
	}

	cache.Set("key1", "value1", 100)
	cache.Set("key2", "value2", 200)

	items, size = cache.Stats()
	if items != 2 {
		t.Errorf("expected items=2, got %d", items)
	}
	if size != 300 {
		t.Errorf("expected size=300, got %d", size)
	}
}

func TestCleanup(t *testing.T) {
	cache := New(50*time.Millisecond, 1000)

	cache.Set("key1", "value1", 100)
	time.Sleep(100 * time.Millisecond)

	// Cleanup runs every minute by default, but we can test expired retrieval
	val, found := cache.Get("key1")
	if found {
		t.Error("expected expired item to be cleaned up")
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestConcurrency(t *testing.T) {
	cache := New(time.Minute, 10000)

	// Run concurrent operations
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			cache.Set(string(rune(i)), i, 10)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Get(string(rune(i)))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			cache.Delete(string(rune(i)))
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for goroutines")
		}
	}

	// Cache should still be functional
	cache.Set("final", "value", 10)
	val, found := cache.Get("final")
	if !found || val != "value" {
		t.Error("cache not functional after concurrent access")
	}
}

func TestDifferentValueTypes(t *testing.T) {
	cache := New(time.Minute, 1000)

	tests := []struct {
		key   string
		value any
		size  int
	}{
		{"string", "hello", 10},
		{"int", 42, 8},
		{"slice", []int{1, 2, 3}, 24},
		{"map", map[string]int{"a": 1}, 16},
		{"struct", struct{ Name string }{Name: "test"}, 16},
	}

	for _, tt := range tests {
		cache.Set(tt.key, tt.value, tt.size)
		val, found := cache.Get(tt.key)
		if !found {
			t.Errorf("expected to find %s", tt.key)
		}
		if val == nil {
			t.Errorf("expected non-nil value for %s", tt.key)
		}
	}
}
