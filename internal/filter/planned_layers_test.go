package filter

import (
	"strings"
	"testing"
)

func TestPlannedLayersInit(t *testing.T) {
	p := NewPipelineCoordinator(PipelineConfig{
		Mode:                ModeMinimal,
		EnablePlannedLayers: true,
	})
	if len(p.plannedLayers) != 20 {
		t.Fatalf("planned layers = %d, want 20", len(p.plannedLayers))
	}
}

func TestPlannedLayersProcess(t *testing.T) {
	input := strings.Repeat("INFO line\n", 100) + "ERROR: failed at main.go:42\n" + strings.Repeat("INFO line\n", 100)
	p := NewPipelineCoordinator(PipelineConfig{
		Mode:                   ModeMinimal,
		EnablePlannedLayers:    true,
		EnableQualityGuardrail: true,
	})
	out, stats := p.Process(input)
	if out == "" {
		t.Fatal("output should not be empty")
	}
	if stats.OriginalTokens <= 0 {
		t.Fatal("invalid stats")
	}
}

func TestCorePresetEnablesPlannedLayers(t *testing.T) {
	cfg := TierConfig(TierCore, ModeMinimal)
	if !cfg.EnablePlannedLayers {
		t.Fatal("core preset should enable planned layers")
	}
}
