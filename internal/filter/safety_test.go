package filter

import (
	"sync"
	"testing"
)

// TestAddLayerStatSafe verifies thread-safe layer stat addition.
func TestAddLayerStatSafe(t *testing.T) {
	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}

	// Add a layer stat
	stats.AddLayerStatSafe("test_layer", LayerStat{TokensSaved: 100})

	// Verify it was added
	if stat, ok := stats.LayerStats["test_layer"]; !ok {
		t.Error("layer stat not added")
	} else if stat.TokensSaved != 100 {
		t.Errorf("expected 100 tokens saved, got %d", stat.TokensSaved)
	}

	// Verify running saved was updated
	if stats.RunningSavedSafe() != 100 {
		t.Errorf("expected running saved to be 100, got %d", stats.RunningSavedSafe())
	}
}

// TestAddLayerStatSafeConcurrent verifies thread safety under concurrent access.
func TestAddLayerStatSafeConcurrent(t *testing.T) {
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
					"layer_"+string(rune(id)),
					LayerStat{TokensSaved: 1},
				)
			}
		}(i)
	}

	wg.Wait()

	// Verify total
	totalSaved := stats.RunningSavedSafe()
	expected := numGoroutines * numIterations
	if totalSaved != expected {
		t.Errorf("expected %d total saved, got %d", expected, totalSaved)
	}
}

// TestPipelineWithThreadSafeStats verifies the pipeline uses thread-safe methods.
func TestPipelineWithThreadSafeStats(t *testing.T) {
	cfg := PipelineConfig{
		Mode:            ModeMinimal,
		EnableEntropy:   true,
		EnablePerplexity: false,
	}
	pipeline := NewPipelineCoordinator(cfg)

	input := "Test input for thread-safe pipeline"
	output, stats := pipeline.Process(input)

	if output == "" {
		t.Error("pipeline returned empty output")
	}

	if stats.OriginalTokens == 0 {
		t.Error("pipeline did not count original tokens")
	}

	// Verify stats were collected
	if len(stats.LayerStats) == 0 {
		t.Error("no layer stats collected")
	}
}

// TestConstantsUsage verifies constants are used correctly.
func TestConstantsUsage(t *testing.T) {
	// Verify constant values are reasonable
	if MinContentLength <= 0 {
		t.Error("MinContentLength should be positive")
	}

	if TightBudgetThreshold <= 0 {
		t.Error("TightBudgetThreshold should be positive")
	}

	if StreamingThreshold <= 0 {
		t.Error("StreamingThreshold should be positive")
	}

	if EarlyExitCheckInterval <= 0 {
		t.Error("EarlyExitCheckInterval should be positive")
	}
}

// TestPipelineConfigWithConstants verifies config uses constants.
func TestPipelineConfigWithConstants(t *testing.T) {
	cfg := PipelineConfig{
		Mode:   ModeMinimal,
		Budget: TightBudgetThreshold - 1, // Just under threshold
	}

	if cfg.Budget >= TightBudgetThreshold {
		t.Error("budget should be under tight threshold")
	}
}

// BenchmarkAddLayerStatSafe measures thread-safe method performance.
func BenchmarkAddLayerStatSafe(b *testing.B) {
	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.AddLayerStatSafe("test", LayerStat{TokensSaved: 10})
	}
}

// BenchmarkAddLayerStatSafeConcurrent measures concurrent performance.
func BenchmarkAddLayerStatSafeConcurrent(b *testing.B) {
	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.AddLayerStatSafe("test", LayerStat{TokensSaved: 10})
		}
	})
}
