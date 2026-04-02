package kvcache

import "testing"

func TestKVCacheAligner(t *testing.T) {
	a := NewKVCacheAligner()
	if a == nil {
		t.Fatal("Expected non-nil aligner")
	}

	msgs := []string{
		"system prompt",
		"user query 1",
		"assistant response 1",
		"user query 2",
	}

	result := a.AnalyzePrefix(msgs)
	if result == nil {
		t.Error("Expected analysis result")
	}

	optimized := a.OptimizeMessageOrder(msgs)
	if len(optimized) != len(msgs) {
		t.Errorf("Expected %d messages, got %d", len(msgs), len(optimized))
	}
}

func TestKVCacheCheckCache(t *testing.T) {
	a := NewKVCacheAligner()

	if a.CheckCache("content1") {
		t.Error("First check should be miss")
	}
	if !a.CheckCache("content1") {
		t.Error("Second check should be hit")
	}

	if a.HitRate() == 0 {
		t.Error("Expected non-zero hit rate")
	}
}

func TestKVCacheStats(t *testing.T) {
	a := NewKVCacheAligner()
	a.CheckCache("test1")
	a.CheckCache("test1")
	a.CheckCache("test2")

	stats := a.Stats()
	if stats["hits"].(int) != 1 {
		t.Errorf("Expected 1 hit, got %v", stats["hits"])
	}
}
