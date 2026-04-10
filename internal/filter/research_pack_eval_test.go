package filter

import (
	"strings"
	"testing"
)

func TestResearchPackExtensions_EvalHarness(t *testing.T) {
	type evalCase struct {
		name     string
		input    string
		mustKeep string
		apply    func(string) (string, int)
	}

	cases := []evalCase{
		{
			name: "swe_adaptive_loop",
			input: strings.Join([]string{
				"Planner: investigate panic in auth handler",
				"noise 1", "noise 2", "noise 3", "noise 4",
				"ERROR: nil pointer at internal/auth/handler.go:88",
				"Executor: add guard and run go test ./...",
				"noise 5", "noise 6", "noise 7", "noise 8",
			}, "\n"),
			mustKeep: "handler.go",
			apply: func(input string) (string, int) {
				return NewSWEAdaptiveLoopFilter().Apply(input, ModeMinimal)
			},
		},
		{
			name: "agent_ocr_history",
			input: strings.Join([]string{
				"Planner: evaluate migration rollback safety",
				"verbose line a", "verbose line b", "verbose line c",
				"Critic: repeated concerns", "repeat 1", "repeat 2",
				"Executor: apply migration in db/migrate.go",
				"Reviewer: approve and merge",
			}, "\n"),
			mustKeep: "Executor:",
			apply: func(input string) (string, int) {
				return NewAgentOCRHistoryFilter().Apply(input, ModeMinimal)
			},
		},
		{
			name: "plan_budget",
			input: strings.Join([]string{
				"Step 1: read stack trace",
				"stack trace: service panicked in payment path",
				"therefore adjust plan and budget",
				"filler one", "filler two", "filler three", "filler four",
				"ERROR: payment validation failed",
			}, "\n"),
			mustKeep: "[plan-budget:",
			apply: func(input string) (string, int) {
				return NewPlanBudgetFilter().Apply(input, ModeMinimal)
			},
		},
		{
			name: "lightmem",
			input: strings.Join([]string{
				"ERROR: migration failed at file db/migrate.go line 88",
				"context",
				"ERROR: migration failed at file db/migrate.go line 88",
				"path: internal/service/payment.go",
				"path: internal/service/payment.go",
				"WARN: retrying migration",
				"WARN: retrying migration",
				"context final",
			}, "\n"),
			mustKeep: "[lightmem: reuse",
			apply: func(input string) (string, int) {
				return NewLightMemFilter().Apply(input, ModeMinimal)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, saved := tc.apply(tc.input)
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tc.mustKeep)) {
				t.Fatalf("expected output to retain %q", tc.mustKeep)
			}
			if saved < 0 {
				t.Fatalf("unexpected negative token savings")
			}
		})
	}
}

func BenchmarkResearchPackExtensions(b *testing.B) {
	input := strings.Join([]string{
		"Planner: investigate flaky migration in production",
		"noise one", "noise two", "noise three", "noise four", "noise five",
		"ERROR: migration failed at file db/migrate.go line 88",
		"Executor: apply fix and rerun integration suite",
		"path: internal/service/payment.go",
		"path: internal/service/payment.go",
		"WARN: retrying migration",
		"WARN: retrying migration",
		"Reviewer: approve patch",
	}, "\n")

	filters := []Filter{
		NewSWEAdaptiveLoopFilter(),
		NewAgentOCRHistoryFilter(),
		NewPlanBudgetFilter(),
		NewLightMemFilter(),
	}

	for _, f := range filters {
		b.Run(f.Name(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = f.Apply(input, ModeMinimal)
			}
		})
	}
}
