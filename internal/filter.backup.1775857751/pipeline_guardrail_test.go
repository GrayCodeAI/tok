package filter

import (
	"strings"
	"testing"
)

func TestPipelineGuardrailFallbackTriggered(t *testing.T) {
	input := strings.Repeat("INFO line\n", 600) + "panic: nil pointer dereference\n" + strings.Repeat("tail\n", 600)

	p := NewPipelineCoordinator(PipelineConfig{
		Mode:                      ModeAggressive,
		SessionTracking:           true,
		EnableExtractivePrefilter: true,
		ExtractiveMaxLines:        50,
		ExtractiveHeadLines:       5,
		ExtractiveTailLines:       5,
		ExtractiveSignalLines:     0,
		EnableQualityGuardrail:    true,
	})

	out, stats := p.Process(input)
	_ = stats
	if !strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("expected quality guardrail path to preserve panic line")
	}
}
