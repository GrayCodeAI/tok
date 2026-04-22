package filter

import (
	"strings"
	"testing"
)

func TestShouldEarlyExit_NoBudget(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)
	stats := &PipelineStats{}

	if coord.shouldEarlyExit(stats) {
		t.Error("expected no early exit when budget is 0")
	}
}

func TestShouldEarlyExit_WithBudget(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal, Budget: 100}
	coord := NewPipelineCoordinator(&cfg)
	stats := &PipelineStats{FinalTokens: 200}

	if !coord.shouldEarlyExit(stats) {
		t.Error("expected early exit when final tokens exceed budget")
	}
}

func TestShouldSkipEntropy(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipEntropy("hi") {
		t.Error("expected skip entropy for very short content")
	}
	// Content with low character diversity (< 30 unique) should be skipped
	if !coord.shouldSkipEntropy(strings.Repeat("a", 100)) {
		t.Error("expected skip entropy for low-diversity content")
	}
	// Content with high character diversity (> 30 unique) should NOT be skipped
	diverse := "The quick brown fox jumps over the lazy dog. 1234567890 !@#$%^&*()"
	if coord.shouldSkipEntropy(diverse) {
		t.Error("expected not to skip entropy for high-diversity content")
	}
}

func TestShouldSkipPerplexity(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipPerplexity("one line") {
		t.Error("expected skip perplexity for single-line content")
	}
	if coord.shouldSkipPerplexity("line1\nline2\nline3\nline4\nline5\nline6") {
		t.Error("expected not to skip perplexity for multi-line content")
	}
}

func TestShouldSkipQueryDependent(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipQueryDependent() {
		t.Error("expected skip query-dependent layers when no query intent")
	}
}

func TestShouldSkipNgram(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipNgram("tiny") {
		t.Error("expected skip ngram for very short content")
	}
}

func TestShouldSkipCompaction(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipCompaction("short") {
		t.Error("expected skip compaction for short content")
	}
}

func TestShouldSkipH2O(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipH2O("hi") {
		t.Error("expected skip H2O for very short content")
	}
}

func TestShouldSkipAttentionSink(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipAttentionSink("one\ntwo") {
		t.Error("expected skip attention sink for short content")
	}
}

func TestShouldSkipMetaToken(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipMetaToken("short") {
		t.Error("expected skip meta-token for short content")
	}
}

func TestShouldSkipSemanticChunk(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipSemanticChunk("short text") {
		t.Error("expected skip semantic chunk for short content")
	}
}

func TestShouldSkipBudgetDependent(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	if !coord.shouldSkipBudgetDependent() {
		t.Error("expected skip budget-dependent layers when budget is 0")
	}
}

func TestFinalizeStats(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)
	stats := &PipelineStats{OriginalTokens: 100}

	result := coord.finalizeStats(stats, "output")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FinalTokens == 0 {
		t.Error("expected non-zero final tokens")
	}
	if result.TotalSaved < 0 {
		t.Error("expected non-negative total saved")
	}
	if result.ReductionPercent < 0 || result.ReductionPercent > 100 {
		t.Errorf("expected reduction percent in [0,100], got %f", result.ReductionPercent)
	}
}

func TestProcessLayer(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)
	stats := &PipelineStats{LayerStats: make(map[string]LayerStat)}

	// Process with a nil filter layer (should return input unchanged)
	layer := filterLayer{filter: nil, name: "test"}
	output := coord.processLayer(layer, "input", stats)
	if output != "input" {
		t.Errorf("expected passthrough for nil filter, got %q", output)
	}
}
