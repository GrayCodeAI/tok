package filter

import "testing"

// Auto-generated tests for coverage

func TestAutoCoverageAllFilters(t *testing.T) {
	filters := []struct {
		name string
		f    Filter
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

	for _, ff := range filters {
		t.Run(ff.name, func(t *testing.T) {
			out, saved := ff.f.Apply("test", ModeAggressive)
			if saved < 0 {
				t.Error("negative savings")
			}
			_ = out
		})
	}
}
