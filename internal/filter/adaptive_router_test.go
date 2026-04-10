package filter

import (
	"strings"
	"testing"
)

func TestAdaptiveRouting_InferQueryIntent(t *testing.T) {
	p := NewPipelineCoordinator(PipelineConfig{
		Mode:               ModeMinimal,
		EnablePolicyRouter: true,
	})
	stats := &PipelineStats{OriginalTokens: 1000, LayerStats: map[string]LayerStat{}}
	_ = p.applyAdaptiveRouting("panic: nil pointer\nstack trace", stats)
	if p.runtimeQueryIntent != "debug" {
		t.Fatalf("expected debug intent, got %q", p.runtimeQueryIntent)
	}
}

func TestAdaptiveRouting_ExtractivePrefilter(t *testing.T) {
	input := strings.Repeat("INFO normal line\n", 200) + "ERROR: critical failure\n" + strings.Repeat("tail line\n", 200)
	p := NewPipelineCoordinator(PipelineConfig{
		Mode:                      ModeMinimal,
		EnableExtractivePrefilter: true,
		ExtractiveMaxLines:        80,
		ExtractiveHeadLines:       10,
		ExtractiveTailLines:       10,
		ExtractiveSignalLines:     10,
	})
	stats := &PipelineStats{OriginalTokens: 2000, LayerStats: map[string]LayerStat{}}
	out := p.applyAdaptiveRouting(input, stats)
	if !strings.Contains(out, "extractive-prefilter") {
		t.Fatalf("expected prefilter marker")
	}
	if !strings.Contains(out, "ERROR: critical failure") {
		t.Fatalf("expected critical signal preserved")
	}
}
