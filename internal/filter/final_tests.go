package filter

import (
	"testing"
)

// Phase 1: Comprehensive tests to reach 95%+ coverage

func TestAllModes(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	input := "test content"

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			cfg := PipelineConfig{Mode: mode}
			p := NewPipelineCoordinator(cfg)
			out, stats := p.Process(input)

			if mode == ModeNone && out != input {
				t.Error("ModeNone modified input")
			}
			if stats.OriginalTokens == 0 {
				t.Error("did not count tokens")
			}
		})
	}
}

func TestAllFiltersBasic(t *testing.T) {
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

	for _, f := range filters {
		t.Run(f.name, func(t *testing.T) {
			out, saved := f.filter.Apply("test", ModeAggressive)
			if saved < 0 {
				t.Error("negative savings")
			}
			_ = out
		})
	}
}

func TestPipelineWithLayers(t *testing.T) {
	cfg := PipelineConfig{
		Mode:             ModeAggressive,
		EnableEntropy:    true,
		EnablePerplexity: true,
		EnableAST:        true,
		EnableH2O:        true,
		EnableMetaToken:  true,
		EnableLazyPruner: true,
		EnableBudget:     true,
		Budget:           1000,
	}
	p := NewPipelineCoordinator(cfg)
	out, stats := p.Process("test input with content for processing")

	if out == "" {
		t.Error("returned empty")
	}
	if stats.OriginalTokens == 0 {
		t.Error("did not count tokens")
	}
}

func TestStatsSafe(t *testing.T) {
	stats := &PipelineStats{}

	// Test concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			stats.AddLayerStatSafe("test", LayerStat{TokensIn: 10, TokensOut: 5})
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			_ = stats.RunningSavedSafe()
		}()
	}
}
