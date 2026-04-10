package filter

import (
	"strings"
	"testing"
)

type suiteScenario struct {
	name     string
	input    string
	mustKeep string
}

func TestBenchmarkSuiteScenarios(t *testing.T) {
	scenarios := []suiteScenario{
		{
			name:     "single_shot_error_log",
			input:    strings.Repeat("INFO build step\n", 200) + "ERROR: linker failed at cmd/main.go:42\n",
			mustKeep: "error",
		},
		{
			name:     "multi_turn_context",
			input:    "User: fix failing test\nAssistant: checking stack trace\npanic: runtime error in service.go:88\n",
			mustKeep: "panic",
		},
		{
			name:     "diff_review",
			input:    "diff --git a/a.go b/a.go\n@@ -10,2 +10,3 @@\n- old\n+ new\n",
			mustKeep: "diff --git",
		},
		{
			name:     "test_output",
			input:    "=== RUN TestX\n--- FAIL: TestX\nexpected 1 got 2\n",
			mustKeep: "fail",
		},
	}

	cfg := TierConfig(TierAdaptive, ModeMinimal)
	cfg.EnableQualityGuardrail = true
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			p := NewPipelineCoordinator(cfg)
			out, stats := p.Process(sc.input)
			if stats.OriginalTokens == 0 {
				t.Fatalf("invalid stats")
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(sc.mustKeep)) {
				t.Fatalf("expected output to retain %q", sc.mustKeep)
			}
		})
	}
}
