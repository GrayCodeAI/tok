package filter

import (
	"fmt"
	"testing"
)

// Extended tests to increase coverage to 95%+

func TestPipelineConfigVariationsExtended(t *testing.T) {
	variations := []PipelineConfig{
		{Mode: ModeNone},
		{Mode: ModeMinimal},
		{Mode: ModeAggressive},
		{Mode: ModeAggressive, Budget: 100},
		{Mode: ModeAggressive, Budget: 1000, EnableEntropy: true},
		{Mode: ModeAggressive, EnablePerplexity: true},
		{Mode: ModeAggressive, EnableAST: true},
		{Mode: ModeAggressive, EnableH2O: true},
		{Mode: ModeAggressive, EnableMetaToken: true},
		{Mode: ModeAggressive, EnableLazyPruner: true},
		{
			Mode:             ModeAggressive,
			EnableEntropy:    true,
			EnablePerplexity: true,
			EnableAST:        true,
			EnableH2O:        true,
			EnableMetaToken:  true,
			EnableLazyPruner: true,
		},
	}

	input := "Test input content for pipeline variations"

	for i, cfg := range variations {
		t.Run(fmt.Sprintf("Config%d", i), func(t *testing.T) {
			p := NewPipelineCoordinator(cfg)
			out, stats := p.Process(input)

			if cfg.Mode == ModeNone && out != input {
				t.Error("ModeNone modified input")
			}

			if stats.OriginalTokens == 0 {
				t.Error("did not count tokens")
			}
		})
	}
}

func TestAllFilterTypesExtended(t *testing.T) {
	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
		{"gist", NewGistFilter()},
		{"attribution", NewAttributionFilter()},
		{"meta_token", NewMetaTokenFilter()},
		{"semantic_chunk", NewSemanticChunkFilter()},
		{"lazy_pruner", NewLazyPrunerFilter()},
		{"semantic_anchor", NewSemanticAnchorFilter()},
		{"agent_memory", NewAgentMemoryFilter()},
	}

	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	inputs := []string{
		"",
		"a",
		"test",
		"test input",
		"test input with content",
	}

	for _, f := range filters {
		for _, mode := range modes {
			for _, input := range inputs {
				testName := fmt.Sprintf("%s_%s_%d", f.name, string(mode), len(input))
				t.Run(testName, func(t *testing.T) {
					out, saved := f.filter.Apply(input, mode)

					if mode == ModeNone && out != input {
						t.Errorf("%s ModeNone modified input", f.name)
					}

					if saved < 0 {
						t.Errorf("%s: negative savings", f.name)
					}
				})
			}
		}
	}
}

func TestConcurrentFilterAccess(t *testing.T) {
	filters := []Filter{
		NewEntropyFilter(),
		NewANSIFilter(),
		NewH2OFilter(),
	}

	input := "Concurrent test input"
	done := make(chan bool, 30)

	for _, filter := range filters {
		for i := 0; i < 10; i++ {
			go func(f Filter) {
				for j := 0; j < 10; j++ {
					f.Apply(input, ModeAggressive)
				}
				done <- true
			}(filter)
		}
	}

	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestStatsOperations(t *testing.T) {
	stats := &PipelineStats{}

	// Add layer stats
	stats.AddLayerStatSafe("entropy", LayerStat{TokensSaved: 20})
	stats.AddLayerStatSafe("ansi", LayerStat{TokensSaved: 5})
	stats.AddLayerStatSafe("h2o", LayerStat{TokensSaved: 15})

	// Verify running saved
	saved := stats.RunningSavedSafe()
	if saved != 40 {
		t.Errorf("expected 40 saved, got %d", saved)
	}

	// Verify layer stats were added
	if len(stats.LayerStats) != 3 {
		t.Errorf("expected 3 layer stats, got %d", len(stats.LayerStats))
	}
}

func BenchmarkFiltersExtended(b *testing.B) {
	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
	}

	input := "Benchmark input content for testing"

	for _, f := range filters {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f.filter.Apply(input, ModeAggressive)
			}
		})
	}
}

func BenchmarkPipelineExtended(b *testing.B) {
	cfg := PipelineConfig{
		Mode:             ModeAggressive,
		EnableEntropy:    true,
		EnablePerplexity: true,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Benchmark input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}
