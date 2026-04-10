package filter

import (
	"strings"
	"testing"
)

type agenticEvalScenario struct {
	name     string
	input    string
	mustKeep string
	apply    func(string) (string, int)
}

func TestAgenticEvalHarness(t *testing.T) {
	scenarios := []agenticEvalScenario{
		{
			name: "latent_collab_repeated_plans",
			input: strings.Join([]string{
				"Planner: analyze null pointer in payment handler and add guard",
				"plan detail a",
				"Planner: analyze null pointer in payment handler and add guard",
				"plan detail b",
				"Executor: patch payment_handler.go line 122 and rerun tests",
			}, "\n"),
			mustKeep: "payment_handler.go",
			apply: func(input string) (string, int) {
				return NewLatentCollabFilter().Apply(input, ModeMinimal)
			},
		},
		{
			name: "graph_cot_reasoning_trace",
			input: strings.Join([]string{
				"Step 1: parse failing logs",
				"because request body is empty, decoder fails",
				"therefore add nil check before decode",
				"random filler one",
				"random filler two",
				"as a result tests should pass",
				"random filler three",
				"Step 2: rerun tests",
			}, "\n"),
			mustKeep: "therefore",
			apply: func(input string) (string, int) {
				return NewGraphCoTFilter().Apply(input, ModeMinimal)
			},
		},
		{
			name: "role_budget_multi_agent",
			input: strings.Join([]string{
				"Critic: long repetition and restatement",
				"restate 1", "restate 2", "restate 3",
				"Planner: decide rollback-safe migration strategy",
				"Executor: apply migration and verify checksum",
				"Tool: INFO info info info info",
			}, "\n"),
			mustKeep: "executor:",
			apply: func(input string) (string, int) {
				return NewRoleBudgetFilter().Apply(input, ModeAggressive)
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			out, saved := sc.apply(sc.input)
			if !strings.Contains(strings.ToLower(out), strings.ToLower(sc.mustKeep)) {
				t.Fatalf("expected output to retain %q", sc.mustKeep)
			}
			if saved < 0 {
				t.Fatalf("unexpected negative savings")
			}
		})
	}
}
