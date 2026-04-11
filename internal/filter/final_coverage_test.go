package filter

// Final comprehensive test suite to reach 95%+ coverage
// This file adds tests for all uncovered functions

import (
	"testing"
)

// ==================== FILTER MODE TESTS ====================

func TestModeStringValues(t *testing.T) {
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

// ==================== ALL FILTER CONSTRUCTORS ====================

func TestAllFilterConstructors(t *testing.T) {
	// Test that all filters can be created
	filters := map[string]func() Filter{
		"entropy":       func() Filter { return NewEntropyFilter() },
		"ansi":          func() Filter { return NewANSIFilter() },
		"h2o":           func() Filter { return NewH2OFilter() },
		"gist":          func() Filter { return NewGistFilter() },
		"attribution":   func() Filter { return NewAttributionFilter() },
		"attention_sink": func() Filter { return NewAttentionSinkFilter() },
		"meta_token":    func() Filter { return NewMetaTokenFilter() },
		"semantic_chunk": func() Filter { return NewSemanticChunkFilter() },
		"lazy_pruner":   func() Filter { return NewLazyPrunerFilter() },
		"semantic_anchor": func() Filter { return NewSemanticAnchorFilter() },
		"agent_memory":  func() Filter { return NewAgentMemoryFilter() },
		"budget":        func() Filter { return NewBudgetEnforcer(DefaultBudgetConfig()) },
		"ngram":         func() Filter { return NewNgramAbbreviator(DefaultNgramConfig()) },
		"evaluator":     func() Filter { return NewEvaluatorHeadsFilter(DefaultEvaluatorConfig()) },
		"contrastive":   func() Filter { return NewContrastiveFilter(DefaultContrastiveConfig()) },
	}

	for name, constructor := range filters {
		t.Run(name, func(t *testing.T) {
			filter := constructor()
			if filter == nil {
				t.Errorf("%s: constructor returned nil", name)
			}
		})
	}
}

// ==================== FILTER APPLY ALL MODES ====================

func TestAllFiltersAllModesComprehensive(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	input := "Test input content for comprehensive filter testing"

	filters := map[string]Filter{
		"entropy":     NewEntropyFilter(),
		"ansi":        NewANSIFilter(),
		"h2o":         NewH2OFilter(),
		"gist":        NewGistFilter(),
		"attribution": NewAttributionFilter(),
		"meta_token":  NewMetaTokenFilter(),
	}

	for name, filter := range filters {
		for _, mode := range modes {
			testName := name + "_" + string(mode)
			t.Run(testName, func(t *testing.T) {
				output, saved := filter.Apply(input, mode)
				
				// ModeNone should not modify
				if mode == ModeNone && output != input {
					t.Errorf("%s ModeNone modified input", name)
				}
				
				// Saved should never be negative
				if saved < 0 {
					t.Errorf("%s: negative savings", name)
				}
			})
		}
	}
}

// ==================== EMPTY INPUT TESTS ====================

func TestAllFiltersEmptyInput(t *testing.T) {
	filters := map[string]Filter{
		"entropy":     NewEntropyFilter(),
		"ansi":        NewANSIFilter(),
		"h2o":         NewH2OFilter(),
		"gist":        NewGistFilter(),
		"attribution": NewAttributionFilter(),
		"meta_token":  NewMetaTokenFilter(),
	}

	for name, filter := range filters {
		t.Run(name, func(t *testing.T) {
			output, saved := filter.Apply("", ModeAggressive)
			
			if output != "" {
				t.Errorf("%s: empty input should return empty", name)
			}
			
			if saved != 0 {
				t.Errorf("%s: empty input should save 0", name)
			}
		})
	}
}

// ==================== LARGE INPUT TESTS ====================

func TestAllFiltersLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large input test")
	}

	// 100KB input
	input := "Test content. " + string(make([]byte, 100000))

	filters := map[string]Filter{
		"entropy": NewEntropyFilter(),
		"ansi":    NewANSIFilter(),
		"h2o":     NewH2OFilter(),
	}

	for name, filter := range filters {
		t.Run(name, func(t *testing.T) {
			output, saved := filter.Apply(input, ModeAggressive)
			
			// Should not crash
			_ = output
			_ = saved
		})
	}
}

// ==================== CONCURRENT SAFETY TESTS ====================

func TestAllFiltersConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test")
	}

	filters := map[string]Filter{
		"entropy": NewEntropyFilter(),
		"ansi":    NewANSIFilter(),
		"h2o":     NewH2OFilter(),
	}

	input := "Concurrent test input"

	for name, filter := range filters {
		t.Run(name, func(t *testing.T) {
			// Run from multiple goroutines
			done := make(chan bool, 10)
			
			for i := 0; i < 10; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						filter.Apply(input, ModeAggressive)
					}
					done <- true
				}()
			}
			
			// Wait for all
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	}
}

// ==================== PIPELINE CONFIG TESTS ====================

func TestPipelineConfigVariations(t *testing.T) {
	configs := []PipelineConfig{
		{Mode: ModeNone},
		{Mode: ModeMinimal},
		{Mode: ModeAggressive},
		{Mode: ModeAggressive, Budget: 100},
		{Mode: ModeAggressive, Budget: 1000},
		{
			Mode:             ModeAggressive,
			EnableEntropy:    true,
			EnablePerplexity: true,
			EnableAST:        true,
		},
		{
			Mode:            ModeAggressive,
			EnableH2O:       true,
			EnableMetaToken: true,
		},
	}

	input := "Test input for config variations"

	for i, cfg := range configs {
		t.Run(fmt.Sprintf("Config_%d", i), func(t *testing.T) {
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

// ==================== STATS CONSISTENCY TESTS ====================

func TestStatsConsistency(t *testing.T) {
	cfg := PipelineConfig{
		Mode:          ModeMinimal,
		EnableEntropy: true,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Test input for stats consistency"

	_, stats := pipeline.Process(input)

	// Verify calculations
	if stats.OriginalTokens < stats.FinalTokens {
		t.Error("original < final")
	}

	expectedSaved := stats.OriginalTokens - stats.FinalTokens
	if stats.TotalSaved != expectedSaved {
		t.Errorf("saved mismatch: %d != %d", stats.TotalSaved, expectedSaved)
	}

	if stats.ReductionPercent < 0 || stats.ReductionPercent > 100 {
		t.Errorf("reduction out of range: %f", stats.ReductionPercent)
	}
}

// ==================== BENCHMARKS ====================

func BenchmarkAllFilters(b *testing.B) {
	input := "Benchmark input content"
	
	filters := map[string]Filter{
		"entropy": NewEntropyFilter(),
		"ansi":    NewANSIFilter(),
		"h2o":     NewH2OFilter(),
	}

	for name, filter := range filters {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				filter.Apply(input, ModeAggressive)
			}
		})
	}
}

func BenchmarkPipelineModes(b *testing.B) {
	modes := []Mode{ModeMinimal, ModeAggressive}
	input := "Benchmark input"

	for _, mode := range modes {
		b.Run(string(mode), func(b *testing.B) {
			cfg := PipelineConfig{Mode: mode}
			pipeline := NewPipelineCoordinator(cfg)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pipeline.Process(input)
			}
		})
	}
}
