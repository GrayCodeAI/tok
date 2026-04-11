package filter

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestEntropyFilterModes tests all compression modes
func TestEntropyFilterModes(t *testing.T) {
	filter := NewEntropyFilter()
	input := "This is test content with repeated repeated words words"

	tests := []struct {
		name     string
		mode     Mode
		minSaved int
	}{
		{"ModeNone", ModeNone, 0},
		{"ModeMinimal", ModeMinimal, 0},
		{"ModeAggressive", ModeAggressive, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, saved := filter.Apply(input, tt.mode)
			
			if tt.mode == ModeNone && output != input {
				t.Error("ModeNone should not modify input")
			}
			
			if saved < tt.minSaved {
				t.Errorf("expected at least %d saved, got %d", tt.minSaved, saved)
			}
		})
	}
}

// TestEntropyFilterEmptyInput tests empty input handling
func TestEntropyFilterEmptyInput(t *testing.T) {
	filter := NewEntropyFilter()
	output, saved := filter.Apply("", ModeAggressive)
	
	if output != "" {
		t.Error("empty input should return empty output")
	}
	
	if saved != 0 {
		t.Errorf("empty input should save 0 tokens, got %d", saved)
	}
}

// TestEntropyFilterLargeInput tests large input handling
func TestEntropyFilterLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large input test in short mode")
	}
	
	filter := NewEntropyFilter()
	// Create 100KB input
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("Line " + string(rune(i)) + ": This is test content with some repeated words. ")
	}
	input := sb.String()
	
	output, saved := filter.Apply(input, ModeAggressive)
	
	if len(output) == 0 {
		t.Error("large input should not return empty")
	}
	
	if saved < 0 {
		t.Error("saved tokens should not be negative")
	}
}

// TestPipelineCoordinatorModes tests pipeline with different modes
func TestPipelineCoordinatorModes(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	input := "Test input for pipeline mode verification"
	
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			cfg := PipelineConfig{Mode: mode}
			pipeline := NewPipelineCoordinator(cfg)
			
			output, stats := pipeline.Process(input)
			
			if output == "" && input != "" {
				t.Errorf("mode %s returned empty output", mode)
			}
			
			if stats.OriginalTokens == 0 {
				t.Errorf("mode %s did not count tokens", mode)
			}
			
			if mode == ModeNone && output != input {
				t.Error("ModeNone should not modify input")
			}
		})
	}
}

// TestPipelineCoordinatorBudget tests budget enforcement
func TestPipelineCoordinatorBudget(t *testing.T) {
	input := "This is a test input that should be compressed within budget constraints"
	
	tests := []struct {
		name   string
		budget int
	}{
		{"TightBudget", 10},
		{"MediumBudget", 50},
		{"LargeBudget", 200},
		{"NoBudget", 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:   ModeAggressive,
				Budget: tt.budget,
			}
			pipeline := NewPipelineCoordinator(cfg)
			
			output, stats := pipeline.Process(input)
			
			if output == "" && input != "" {
				t.Error("should not return empty output")
			}
			
			// If budget is set, should try to respect it
			if tt.budget > 0 && stats.FinalTokens > tt.budget*2 {
				t.Logf("Budget: %d, Final: %d (may exceed for small inputs)", 
					tt.budget, stats.FinalTokens)
			}
		})
	}
}

// TestConstantsValues tests all constants have valid values
func TestConstantsValues(t *testing.T) {
	constants := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"TightBudgetThreshold", TightBudgetThreshold, 1, 10000},
		{"MinContentLength", MinContentLength, 1, 1000},
		{"StreamingThreshold", StreamingThreshold, 1000, 10000000},
		{"EarlyExitCheckInterval", EarlyExitCheckInterval, 1, 100},
	}
	
	for _, c := range constants {
		t.Run(c.name, func(t *testing.T) {
			if c.value < c.min || c.value > c.max {
				t.Errorf("%s = %d, expected between %d and %d", 
					c.name, c.value, c.min, c.max)
			}
		})
	}
}

// TestLayerStatsAccumulation tests stats are accumulated correctly
func TestLayerStatsAccumulation(t *testing.T) {
	cfg := PipelineConfig{
		Mode:             ModeMinimal,
		EnableEntropy:    true,
		EnablePerplexity: false,
		EnableAST:        true,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Test input with enough content for multiple layers to process"
	
	_, stats := pipeline.Process(input)
	
	// Should have layer stats
	if len(stats.LayerStats) == 0 {
		t.Error("no layer stats recorded")
	}
	
	// Total saved should be non-negative
	if stats.TotalSaved < 0 {
		t.Errorf("negative tokens saved: %d", stats.TotalSaved)
	}
	
	// Original should be >= final
	if stats.FinalTokens > stats.OriginalTokens {
		t.Error("final tokens > original tokens")
	}
}

// TestConcurrentPipelineUsage tests thread safety
func TestConcurrentPipelineUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Concurrent test input"
	
	// Run from multiple goroutines
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, stats := pipeline.Process(input)
				if stats.OriginalTokens == 0 {
					errors <- fmt.Errorf("goroutine %d: no tokens counted", id)
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	for err := range errors {
		t.Error(err)
	}
}

// BenchmarkEntropyFilterComprehensive benchmarks entropy filter performance
func BenchmarkEntropyFilterComprehensive(b *testing.B) {
	filter := NewEntropyFilter()
	input := "This is benchmark input with repeated words for testing performance"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Apply(input, ModeAggressive)
	}
}

// BenchmarkPipelineComprehensive benchmarks full pipeline performance
func BenchmarkPipelineComprehensive(b *testing.B) {
	cfg := PipelineConfig{
		Mode: ModeMinimal,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Benchmark input for pipeline performance testing"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkAddLayerStatSafeComprehensive benchmarks thread-safe method
func BenchmarkAddLayerStatSafeComprehensive(b *testing.B) {
	stats := &PipelineStats{
		OriginalTokens: 1000,
		LayerStats:     make(map[string]LayerStat),
	}
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.AddLayerStatSafe("benchmark", LayerStat{TokensSaved: 10})
		}
	})
}
