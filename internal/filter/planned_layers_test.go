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
	if len(p.plannedLayers) != 12 {
		t.Fatalf("planned layers = %d, want 12 canonical layers", len(p.plannedLayers))
	}
}

func TestPlannedLayerCanonicalID(t *testing.T) {
	tests := map[string]string{
		"36_stacktrace_focus": "31_trace_preserve",
		"41_error_window":     "31_trace_preserve",
		"40_log_cluster":      "30_salience_graph",
		"46_context_cache":    "30_salience_graph",
		"43_symbolic_patch":   "32_ast_diff_focus",
		"49_repair_pass":      "39_recall_booster",
		"47_confidence_gate":  "48_loss_guard",
		"33_unit_test_focus":  "33_unit_test_focus",
	}
	for in, want := range tests {
		if got := plannedLayerCanonicalID(in); got != want {
			t.Fatalf("canonical(%s)=%s, want %s", in, got, want)
		}
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
