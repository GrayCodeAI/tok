package cache_test

import (
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/cache"
)

func TestFingerprintCache_Basic(t *testing.T) {
	c := cache.NewFingerprintCache(10, 5*time.Minute)

	// Cache miss
	got := c.Get("test content")
	if got == nil || got.Hit {
		t.Error("expected cache miss for uncached content")
	}

	// Set and get
	c.Set("test content", "compressed data here", 42)
	got = c.Get("test content")
	if got == nil || !got.Hit {
		t.Fatal("expected cache hit after Set")
	}
	if got.Cached == nil {
		t.Fatal("expected Cached result")
	}
	if got.Cached.Compressed != "compressed data here" {
		t.Errorf("Compressed = %q, want %q", got.Cached.Compressed, "compressed data here")
	}
	if got.Cached.TokensSaved != 42 {
		t.Errorf("TokensSaved = %d, want 42", got.Cached.TokensSaved)
	}
}

func TestFingerprintCache_ByFingerprint(t *testing.T) {
	c := cache.NewFingerprintCache(10, 5*time.Minute)

	fp := cache.ComputeFingerprint("hello")
	if fp == "" {
		t.Fatal("ComputeFingerprint returned empty string")
	}

	c.SetByFingerprint(fp, "hello", "compressed", 10)
	got := c.GetByFingerprint(fp)
	if got == nil || got.Cached == nil {
		t.Errorf("GetByFingerprint failed, got %+v", got)
	}
}

func TestFingerprintCache_MissAfterTTL(t *testing.T) {
	c := cache.NewFingerprintCache(10, 1*time.Millisecond)
	c.Set("content", "compressed", 10)
	time.Sleep(10 * time.Millisecond)

	got := c.Get("content")
	if got != nil && got.Hit {
		t.Error("expected cache miss after TTL expired")
	}
}

func TestComputeFingerprint_Consistency(t *testing.T) {
	fp1 := cache.ComputeFingerprint("test")
	fp2 := cache.ComputeFingerprint("test")
	fp3 := cache.ComputeFingerprint("different")

	if fp1 != fp2 {
		t.Error("same content should produce same fingerprint")
	}
	if fp1 == fp3 {
		t.Error("different content should produce different fingerprint")
	}
}
