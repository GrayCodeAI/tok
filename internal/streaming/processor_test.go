package streaming

import (
	"strings"
	"testing"
)

func TestNewStreamingProcessor(t *testing.T) {
	sp := NewStreamingProcessor()
	if sp == nil {
		t.Fatal("NewStreamingProcessor returned nil")
	}
	if sp.chunkSize != 4096 {
		t.Errorf("default chunkSize = %d, want 4096", sp.chunkSize)
	}
	if sp.threshold != 500000 {
		t.Errorf("default threshold = %d, want 500000", sp.threshold)
	}
	if !sp.adaptive {
		t.Error("default should be adaptive")
	}
}

func TestShouldStream(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.adaptive = false // disable adaptive for predictable tests

	// Small input should NOT stream
	small := strings.Repeat("word ", 10) // ~50 tokens
	if sp.ShouldStream(small) {
		t.Error("ShouldStream(small) should be false")
	}

	// Large input should stream
	large := strings.Repeat("word ", 600000) // ~3M tokens
	if !sp.ShouldStream(large) {
		t.Error("ShouldStream(large) should be true")
	}
}

func TestShouldStream_Adaptive(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.adaptive = true

	// In adaptive mode, threshold = tokens/2, so any input > tokens/2 should stream
	// which means adaptive always returns true (threshold < input)
	input := strings.Repeat("word ", 1000)
	if !sp.ShouldStream(input) {
		t.Error("adaptive ShouldStream should return true")
	}
}

func TestProcess_BelowThreshold(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.adaptive = false
	sp.threshold = 1000000 // high threshold

	input := "hello world"
	result := sp.Process(input, func(s string) string {
		return strings.ToUpper(s)
	})

	if strings.TrimSpace(result) != "HELLO WORLD" {
		t.Errorf("Process = %q, want %q", strings.TrimSpace(result), "HELLO WORLD")
	}
}

func TestProcess_AboveThreshold(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.adaptive = false
	sp.threshold = 1 // very low threshold to force streaming

	input := "hello\nworld\nmultiline"
	result := sp.Process(input, func(s string) string {
		return strings.ToUpper(s)
	})

	expected := "HELLO\nWORLD\nMULTILINE\n"
	if result != expected {
		t.Errorf("Process = %q, want %q", result, expected)
	}
}

func TestSetChunkSize(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.SetChunkSize(8192)
	if sp.chunkSize != 8192 {
		t.Errorf("chunkSize = %d, want 8192", sp.chunkSize)
	}
}

func TestSetThreshold(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.SetThreshold(100)
	if sp.threshold != 100 {
		t.Errorf("threshold = %d, want 100", sp.threshold)
	}
}

func TestSetAdaptive(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.SetAdaptive(false)
	if sp.adaptive {
		t.Error("SetAdaptive(false) didn't work")
	}
}

func TestProcessWithMetrics(t *testing.T) {
	sp := NewStreamingProcessor()
	sp.adaptive = false
	sp.threshold = 1

	input := "line1\nline2\nline3"
	result, metrics := sp.ProcessWithMetrics(input, func(s string) string {
		return strings.ToUpper(s)
	})

	if !strings.Contains(strings.ToUpper(result), "LINE1") {
		t.Error("result should contain LINE1")
	}
	if metrics.TotalChunks < 1 {
		t.Errorf("TotalChunks = %d, want >= 1", metrics.TotalChunks)
	}
	if metrics.ProcessedBytes <= 0 {
		t.Error("ProcessedBytes should be positive")
	}
}

func TestBackpressureController(t *testing.T) {
	c := NewBackpressureController(2)

	if !c.TryAcquire() {
		t.Error("first acquire should succeed")
	}
	if !c.TryAcquire() {
		t.Error("second acquire should succeed")
	}
	if c.TryAcquire() {
		t.Error("third acquire should fail (max 2)")
	}
	if !c.IsBlocked() {
		t.Error("should be blocked at limit")
	}

	c.Release() // pending=1
	if !c.TryAcquire() {
		t.Error("should succeed after release")
	}
	// Now pending=2 again (at max)

	// Release twice to get below max
	c.Release() // pending=1
	c.Release() // pending=0
	if c.IsBlocked() {
		t.Error("should not be blocked with pending=0, max=2")
	}
}

func TestBackpressureController_ReleaseBelow(t *testing.T) {
	c := NewBackpressureController(2)
	c.Release() // release without acquire - should be safe
	if c.pending < 0 {
		t.Error("pending should not go negative")
	}
}
