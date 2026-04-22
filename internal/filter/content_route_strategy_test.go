package filter

import (
	"strings"
	"testing"
)

func TestApplyAdaptiveRouting_JSON(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	input := `{"key": "value", "list": [1, 2, 3]}`
	stats := &PipelineStats{}
	output := coord.applyAdaptiveRouting(input, stats)

	if output == "" {
		t.Error("expected non-empty output after routing")
	}
}

func TestApplyAdaptiveRouting_PlainText(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	coord := NewPipelineCoordinator(&cfg)

	input := "This is just plain text without any code or JSON."
	stats := &PipelineStats{}
	output := coord.applyAdaptiveRouting(input, stats)

	if output == "" {
		t.Error("expected non-empty output after routing")
	}
}

func TestApplyExtractivePrefilter(t *testing.T) {
	cfg := PipelineConfig{
		Mode:                      ModeMinimal,
		EnableExtractivePrefilter: true,
		ExtractiveMaxLines:        10,
		ExtractiveHeadLines:       3,
		ExtractiveTailLines:       3,
		ExtractiveSignalLines:     4,
	}
	coord := NewPipelineCoordinator(&cfg)

	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line content"
	}
	input := strings.Join(lines, "\n")

	output, saved := coord.applyExtractivePrefilter(input)
	if output == "" {
		t.Error("expected non-empty output after prefilter")
	}
	if saved < 0 {
		t.Error("expected non-negative saved count")
	}
}

func TestApplyExtractivePrefilter_ShortInput(t *testing.T) {
	cfg := PipelineConfig{
		Mode:                      ModeMinimal,
		EnableExtractivePrefilter: true,
		ExtractiveMaxLines:        100,
		ExtractiveHeadLines:       10,
		ExtractiveTailLines:       10,
		ExtractiveSignalLines:     10,
	}
	coord := NewPipelineCoordinator(&cfg)

	input := "short input"
	output, saved := coord.applyExtractivePrefilter(input)
	if output != input {
		t.Errorf("expected passthrough for short input, got %q", output)
	}
	if saved != 0 {
		t.Errorf("expected 0 saved for short input, got %d", saved)
	}
}
