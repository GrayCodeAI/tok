package filter

import (
	"testing"
)

// TestAllFiltersEmptyInput tests all filters with empty input
func TestAllFiltersEmptyInput(t *testing.T) {
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
			output, saved := f.filter.Apply("", ModeAggressive)
			if output != "" {
				t.Errorf("%s: empty input should return empty", f.name)
			}
			if saved < 0 {
				t.Errorf("%s: saved should not be negative", f.name)
			}
		})
	}
}

// TestAllFiltersAllModes tests all filters with all modes
func TestAllFiltersAllModes(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	input := "Test input content for filter verification"

	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
		{"gist", NewGistFilter()},
	}

	for _, f := range filters {
		for _, mode := range modes {
			t.Run(f.name+"_"+string(mode), func(t *testing.T) {
				output, saved := f.filter.Apply(input, mode)
				
				if mode == ModeNone && output != input {
					t.Errorf("%s ModeNone should not modify", f.name)
				}
				
				if saved < 0 {
					t.Errorf("%s: negative savings", f.name)
				}
			})
		}
	}
}

// TestAllFiltersLargeInput tests all filters with large input
func TestAllFiltersLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large input test")
	}

	// 10KB input
	input := "Large test content. " + string(make([]byte, 10000))

	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
	}

	for _, f := range filters {
		t.Run(f.name, func(t *testing.T) {
			output, saved := f.filter.Apply(input, ModeAggressive)
			
			if len(output) == 0 && len(input) > 0 {
				t.Errorf("%s: returned empty for large input", f.name)
			}
			
			if saved < 0 {
				t.Errorf("%s: negative savings for large input", f.name)
			}
		})
	}
}

// TestAllFiltersConsistency tests filters return consistent results
func TestAllFiltersConsistency(t *testing.T) {
	input := "Test input for consistency"

	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
	}

	for _, f := range filters {
		t.Run(f.name, func(t *testing.T) {
			// Run 5 times
			var results []string
			for i := 0; i < 5; i++ {
				output, _ := f.filter.Apply(input, ModeAggressive)
				results = append(results, output)
			}
			
			// All should be identical
			for i := 1; i < len(results); i++ {
				if results[i] != results[0] {
					t.Errorf("%s: inconsistent results", f.name)
				}
			}
		})
	}
}

// TestAllFiltersUnicode tests filters with unicode content
func TestAllFiltersUnicode(t *testing.T) {
	inputs := []string{
		"Unicode: 你好世界 🌍",
		"Arabic: مرحبا",
		"Russian: Привет",
		"Japanese: こんにちは",
	}

	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
	}

	for _, f := range filters {
		for _, input := range inputs {
			t.Run(f.name+"_"+input[:10], func(t *testing.T) {
				output, _ := f.filter.Apply(input, ModeAggressive)
				if output == "" && input != "" {
					t.Errorf("%s: failed on unicode", f.name)
				}
			})
		}
	}
}

// BenchmarkAllFilters benchmarks all filters
func BenchmarkAllFilters(b *testing.B) {
	input := "Benchmark input content for performance testing"

	filters := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"ansi", NewANSIFilter()},
		{"h2o", NewH2OFilter()},
	}

	for _, f := range filters {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f.filter.Apply(input, ModeAggressive)
			}
		})
	}
}
