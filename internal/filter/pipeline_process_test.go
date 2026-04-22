package filter

import (
	"strings"
	"testing"
)

func TestPipelineCoordinator_Process_EmptyInput(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	output, stats := coord.Process("")
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.OriginalTokens != 0 {
		t.Errorf("expected 0 original tokens, got %d", stats.OriginalTokens)
	}
}

func TestPipelineCoordinator_Process_SmallInput(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	input := "hello world"
	output, stats := coord.Process(input)
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.OriginalTokens == 0 {
		t.Error("expected non-zero original tokens")
	}
	// Small input may pass through unchanged or be lightly filtered
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestPipelineCoordinator_Process_LargeInput(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	input := strings.Repeat("hello world\n", 1000)
	output, stats := coord.Process(input)
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.OriginalTokens == 0 {
		t.Error("expected non-zero original tokens")
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestPipelineCoordinator_runGuardrailFallback(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal, EnableQualityGuardrail: true}
	coord := NewPipelineCoordinator(&cfg)

	input := "important error message here"
	output, stats := coord.runGuardrailFallback(input)
	if output == "" {
		t.Error("expected non-empty fallback output")
	}
	if stats == nil {
		t.Fatal("expected non-nil stats from fallback")
	}
}

func TestPipelineStats_ThreadSafety(t *testing.T) {
	stats := &PipelineStats{LayerStats: make(map[string]LayerStat)}
	stats.AddLayerStatSafe("test", LayerStat{TokensSaved: 10})
	if stats.RunningSavedSafe() != 10 {
		t.Errorf("expected running saved 10, got %d", stats.RunningSavedSafe())
	}
}
