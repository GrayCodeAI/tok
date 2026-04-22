package cache

import (
	"sync"
	"testing"
	"time"
)

func TestFingerprintCache_GetSet(t *testing.T) {
	c := NewFingerprintCache()
	c.Set("key1", "value1")

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if v != "value1" {
		t.Errorf("expected value1, got %q", v)
	}
}

func TestFingerprintCache_GetMiss(t *testing.T) {
	c := NewFingerprintCache()
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected miss for nonexistent key")
	}
}

func TestFingerprintCache_Overwrite(t *testing.T) {
	c := NewFingerprintCache()
	c.Set("key1", "value1")
	c.Set("key1", "value2")

	v, ok := c.Get("key1")
	if !ok || v != "value2" {
		t.Errorf("expected value2, got %q (ok=%v)", v, ok)
	}
}

func TestFingerprintCache_Expiration(t *testing.T) {
	c := NewFingerprintCache()
	c.ttl = 50 * time.Millisecond
	c.Set("key1", "value1")

	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist before expiration")
	}

	time.Sleep(100 * time.Millisecond)
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestFingerprintCache_Eviction(t *testing.T) {
	c := NewFingerprintCache()
	c.maxSize = 3

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")
	c.Set("d", "4") // should evict "a"

	_, ok := c.Get("a")
	if ok {
		t.Error("expected 'a' to be evicted")
	}
	_, ok = c.Get("d")
	if !ok {
		t.Error("expected 'd' to exist")
	}
}

func TestFingerprintCache_ProactiveSweep(t *testing.T) {
	c := NewFingerprintCache()
	c.ttl = 10 * time.Millisecond
	c.maxSize = 2

	c.Set("old", "1")
	time.Sleep(50 * time.Millisecond)
	c.Set("new", "2") // triggers proactive sweep of expired "old"

	_, ok := c.Get("old")
	if ok {
		t.Error("expected 'old' to be swept during Set")
	}
}

func TestFingerprintCache_Concurrent(t *testing.T) {
	c := NewFingerprintCache()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			c.Set("key", "value")
		}(i)
		go func(i int) {
			defer wg.Done()
			c.Get("key")
		}(i)
	}
	wg.Wait()
}

func TestLRUCache_GetSet(t *testing.T) {
	c := NewLRUCache(10, time.Minute)
	c.Set("key1", "value1")

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if v.(string) != "value1" {
		t.Errorf("expected value1, got %q", v)
	}
}

func TestLRUCache_GetMiss(t *testing.T) {
	c := NewLRUCache(10, time.Minute)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected miss")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	c := NewLRUCache(10, time.Minute)
	c.Set("key1", "value1")
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestLRUCache_DeleteNonExistent(t *testing.T) {
	c := NewLRUCache(10, time.Minute)
	c.Delete("nonexistent") // should not panic
}

func TestLRUCache_Len(t *testing.T) {
	c := NewLRUCache(10, time.Minute)
	if c.Len() != 0 {
		t.Errorf("expected len 0, got %d", c.Len())
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Errorf("expected len 2, got %d", c.Len())
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache(2, time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // evicts "a"

	_, ok := c.Get("a")
	if ok {
		t.Error("expected 'a' to be evicted")
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	c := NewLRUCache(10, 50*time.Millisecond)
	c.Set("key1", "value1")

	time.Sleep(100 * time.Millisecond)
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	c := NewLRUCache(100, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			c.Set("key", i)
		}(i)
		go func() {
			defer wg.Done()
			c.Get("key")
		}()
		go func() {
			defer wg.Done()
			c.Delete("key")
		}()
	}
	wg.Wait()
}

func TestGetGlobalCache(t *testing.T) {
	c1 := GetGlobalCache()
	c2 := GetGlobalCache()
	if c1 != c2 {
		t.Error("expected same singleton instance")
	}
}
