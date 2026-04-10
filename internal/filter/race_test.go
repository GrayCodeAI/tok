package filter

import (
	"sync"
	"testing"
)

// TestPipelineStatsThreadSafety verifies no race conditions.
func TestPipelineStatsThreadSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race test in short mode")
	}

	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	numIterations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				stats.AddLayerStatSafe(
					"test_layer",
					LayerStat{TokensSaved: 10},
				)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				_ = stats.RunningSavedSafe()
				_ = stats.LayerStats
			}
		}(i)
	}

	wg.Wait()

	// Verify results
	totalSaved := stats.RunningSavedSafe()
	expected := numGoroutines * numIterations * 10
	if totalSaved != expected {
		t.Errorf("expected %d tokens saved, got %d", expected, totalSaved)
	}
}

// TestSafeFilterNilHandling verifies nil filters don't panic.
func TestSafeFilterNilHandling(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		input  string
	}{
		{
			name:   "nil filter",
			filter: nil,
			input:  "test input",
		},
		{
			name:   "valid filter",
			filter: NewEntropyFilter(),
			input:  "test input with some content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safeFilter := NewSafeFilter(tt.filter, "test")
			
			// Should not panic
			output, saved := safeFilter.Apply(tt.input, ModeMinimal)
			
			// With nil filter, output should equal input
			if tt.filter == nil && output != tt.input {
				t.Errorf("nil filter should return input unchanged, got %q", output)
			}
			
			if tt.filter == nil && saved != 0 {
				t.Errorf("nil filter should return 0 saved, got %d", saved)
			}
		})
	}
}

// TestSafeFilterPanicRecovery verifies panics are recovered.
func TestSafeFilterPanicRecovery(t *testing.T) {
	panicFilter := &panicFilter{}
	safeFilter := NewSafeFilter(panicFilter, "panic_test")
	
	input := "test input"
	output, saved := safeFilter.Apply(input, ModeMinimal)
	
	// Should recover and return input unchanged
	if output != input {
		t.Errorf("expected input after panic recovery, got %q", output)
	}
	
	if saved != 0 {
		t.Errorf("expected 0 saved after panic, got %d", saved)
	}
}

// panicFilter is a test filter that always panics.
type panicFilter struct{}

func (p *panicFilter) Apply(input string, mode Mode) (string, int) {
	panic("intentional panic for testing")
}

// TestConstantsAreUsed verifies constants are used correctly.
func TestConstantsAreUsed(t *testing.T) {
	// Verify constants have reasonable values
	if MinContentLength <= 0 {
		t.Error("MinContentLength should be positive")
	}
	
	if StreamingThreshold <= 0 {
		t.Error("StreamingThreshold should be positive")
	}
	
	if TightBudgetThreshold <= 0 {
		t.Error("TightBudgetThreshold should be positive")
	}
	
	// Verify mode constants
	if !ModeNone.IsValid() {
		t.Error("ModeNone should be valid")
	}
	
	if !ModeMinimal.IsValid() {
		t.Error("ModeMinimal should be valid")
	}
	
	if !ModeAggressive.IsValid() {
		t.Error("ModeAggressive should be valid")
	}
	
	invalidMode := Mode("invalid")
	if invalidMode.IsValid() {
		t.Error("invalid mode should not be valid")
	}
}

// BenchmarkSafeFilter measures safe filter overhead.
func BenchmarkSafeFilter(b *testing.B) {
	filter := NewEntropyFilter()
	safeFilter := NewSafeFilter(filter, "entropy")
	input := "test input with enough content for entropy calculation"
	
	b.Run("direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			filter.Apply(input, ModeMinimal)
		}
	})
	
	b.Run("safe_wrapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			safeFilter.Apply(input, ModeMinimal)
		}
	})
}

// BenchmarkPipelineStatsSafe measures thread-safe stats performance.
func BenchmarkPipelineStatsSafe(b *testing.B) {
	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}
	
	b.Run("unsafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stats.LayerStats["test"] = LayerStat{TokensSaved: 10}
			stats.runningSaved += 10
		}
	})
	
	b.Run("safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stats.AddLayerStatSafe("test", LayerStat{TokensSaved: 10})
		}
	})
}
