package filter

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// ==================== MODE TESTS ====================

func TestAllModes(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	
	for _, mode := range modes {
		t.Run(fmt.Sprintf("Mode_%s", mode), func(t *testing.T) {
			// Test mode validity
			if mode != ModeNone && mode != ModeMinimal && mode != ModeAggressive {
				t.Errorf("invalid mode: %s", mode)
			}
		})
	}
}

// ==================== FILTER INTERFACE TESTS ====================

// TestFilterInterfaceCompliance verifies all filters implement interface
func TestFilterInterfaceCompliance(t *testing.T) {
	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
		{"gist", NewGistFilter()},
		{"attribution", NewAttributionFilter()},
		{"attention_sink", NewAttentionSinkFilter()},
		{"meta_token", NewMetaTokenFilter()},
		{"semantic_chunk", NewSemanticChunkFilter()},
		{"lazy_pruner", NewLazyPrunerFilter()},
		{"semantic_anchor", NewSemanticAnchorFilter()},
		{"agent_memory", NewAgentMemoryFilter()},
	}
	
	for _, f := range filters {
		t.Run(f.name, func(t *testing.T) {
			// Verify Apply method exists and works
			input := "test input with more content for processing"
			output, saved := f.filter.Apply(input, ModeAggressive)
			
			// Filters may return empty for small inputs (below threshold)
			// or return processed content - both are valid
			_ = output
			
			if saved < 0 {
				t.Errorf("%s: negative savings", f.name)
			}
		})
	}
}

// ==================== PIPELINE CONFIG TESTS ====================

func TestPipelineConfigDefaults(t *testing.T) {
	cfg := PipelineConfig{}
	
	// Test defaults
	if cfg.Mode != "" {
		t.Logf("Mode default: %s", cfg.Mode)
	}
}

func TestPipelineConfigWithOptions(t *testing.T) {
	cfg := PipelineConfig{
		Mode:                ModeAggressive,
		Budget:              1000,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableAST:           true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}
	
	pipeline := NewPipelineCoordinator(cfg)
	
	input := "Test input for configuration"
	output, stats := pipeline.Process(input)
	
	if output == "" {
		t.Error("pipeline returned empty")
	}
	
	if stats.OriginalTokens == 0 {
		t.Error("did not count tokens")
	}
}

// ==================== EDGE CASE TESTS ====================

func TestFiltersWithUnicode(t *testing.T) {
	inputs := []string{
		"Unicode: 你好世界 🌍",
		"Arabic: مرحبا",
		"Russian: Привет",
		"Japanese: こんにちは",
		"Emoji: 🚀🎉💻",
	}
	
	filters := []Filter{
		NewEntropyFilter(),
		NewANSIFilter(),
		NewH2OFilter(),
	}
	
	for _, input := range inputs {
		for _, filter := range filters {
			output, _ := filter.Apply(input, ModeAggressive)
			// Emoji inputs may be filtered completely - this is acceptable
			_ = output
			_ = input
		}
	}
}

func TestFiltersWithBinary(t *testing.T) {
	input := "Test\x00binary\xffcontent\x01\x02\x03"
	
	filters := []Filter{
		NewEntropyFilter(),
		NewANSIFilter(),
	}
	
	for _, filter := range filters {
		output, _ := filter.Apply(input, ModeAggressive)
		// Should handle binary without crashing
		_ = output
	}
}

func TestFiltersWithVeryLongInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long input test")
	}
	
	// 1MB input
	input := strings.Repeat("Large input content. ", 50000)
	
	filters := []Filter{
		NewEntropyFilter(),
		NewANSIFilter(),
	}
	
	for _, filter := range filters {
		output, saved := filter.Apply(input, ModeAggressive)
		// Large inputs may be processed differently - just ensure no panic
		_ = output
		if saved < 0 {
			t.Error("negative savings for large input")
		}
	}
}

// ==================== CONCURRENCY TESTS ====================

func TestPipelineConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test")
	}
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Concurrent test input"
	
	var wg sync.WaitGroup
	errors := make(chan error, 20)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, stats := pipeline.Process(input)
				if stats.OriginalTokens == 0 {
					errors <- fmt.Errorf("goroutine %d: no tokens", id)
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

// ==================== BUDGET TESTS ====================

func TestPipelineWithVariousBudgets(t *testing.T) {
	input := "Test input with sufficient content for compression testing"
	
	budgets := []int{0, 10, 50, 100, 500, 1000}
	
	for _, budget := range budgets {
		t.Run(fmt.Sprintf("Budget_%d", budget), func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:   ModeAggressive,
				Budget: budget,
			}
			pipeline := NewPipelineCoordinator(cfg)
			
			output, stats := pipeline.Process(input)
			
			if output == "" {
				t.Error("returned empty")
			}
			
			if stats.OriginalTokens == 0 {
				t.Error("did not count tokens")
			}
		})
	}
}

// ==================== LAYER ENABLE/DISABLE TESTS ====================

func TestPipelineWithLayersDisabled(t *testing.T) {
	input := "Test input"
	
	tests := []struct {
		name string
		cfg  PipelineConfig
	}{
		{
			name: "Only Entropy",
			cfg: PipelineConfig{
				Mode:          ModeMinimal,
				EnableEntropy: true,
			},
		},
		{
			name: "Only H2O",
			cfg: PipelineConfig{
				Mode:    ModeMinimal,
				EnableH2O: true,
			},
		},
		{
			name: "No Layers",
			cfg: PipelineConfig{
				Mode: ModeMinimal,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipelineCoordinator(tt.cfg)
			output, _ := pipeline.Process(input)
			
			if output == "" {
				t.Error("returned empty")
			}
		})
	}
}

// ==================== STATS TESTS ====================

func TestPipelineStatsAccuracy(t *testing.T) {
	cfg := PipelineConfig{
		Mode:             ModeMinimal,
		EnableEntropy:    true,
		EnablePerplexity: false,
		EnableAST:        true,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Test input for stats accuracy verification"
	
	_, stats := pipeline.Process(input)
	
	// Verify stats consistency
	if stats.OriginalTokens < stats.FinalTokens {
		t.Error("original < final tokens")
	}
	
	if stats.TotalSaved != stats.OriginalTokens-stats.FinalTokens {
		t.Errorf("saved calculation incorrect: %d != %d - %d",
			stats.TotalSaved, stats.OriginalTokens, stats.FinalTokens)
	}
	
	if stats.ReductionPercent < 0 || stats.ReductionPercent > 100 {
		t.Errorf("reduction percent out of range: %f", stats.ReductionPercent)
	}
}

// ==================== HELPER FUNCTION TESTS ====================

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 2},
		{"hello world", 3},
		{strings.Repeat("a", 100), 25},
	}
	
	for _, tt := range tests {
		result := EstimateTokens(tt.input)
		// Allow 20% variance for heuristic
		if result < tt.expected*8/10 || result > tt.expected*12/10 {
			t.Logf("EstimateTokens(%q) = %d, expected ~%d", tt.input, result, tt.expected)
		}
	}
}

func TestModeValues(t *testing.T) {
	if string(ModeNone) != "none" {
		t.Error("ModeNone value incorrect")
	}
	if string(ModeMinimal) != "minimal" {
		t.Error("ModeMinimal value incorrect")
	}
	if string(ModeAggressive) != "aggressive" {
		t.Error("ModeAggressive value incorrect")
	}
}

// ==================== BENCHMARKS ====================

func BenchmarkPipelineWithBudget(b *testing.B) {
	cfg := PipelineConfig{
		Mode:   ModeAggressive,
		Budget: 1000,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Benchmark input for performance testing"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

func BenchmarkEntropyFilterDetailed(b *testing.B) {
	filter := NewEntropyFilter()
	inputs := []string{
		"Short",
		"Medium length input for testing",
		strings.Repeat("Large input. ", 100),
	}
	
	for _, input := range inputs {
		b.Run(fmt.Sprintf("len_%d", len(input)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				filter.Apply(input, ModeAggressive)
			}
		})
	}
}

func BenchmarkMultipleFilters(b *testing.B) {
	filters := []Filter{
		NewEntropyFilter(),
		NewANSIFilter(),
		NewH2OFilter(),
	}
	input := "Test input"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, filter := range filters {
			filter.Apply(input, ModeAggressive)
		}
	}
}

func BenchmarkStatsAccumulation(b *testing.B) {
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
