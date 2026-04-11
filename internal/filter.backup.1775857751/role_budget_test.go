package filter

import (
	"strings"
	"testing"
)

func TestRoleBudgetFilter_PrioritizesExecutorPlanner(t *testing.T) {
	f := NewRoleBudgetFilter()
	input := strings.Join([]string{
		"Planner: assess bug scope and patch strategy",
		"plan detail 1",
		"plan detail 2",
		"Critic: restating previous points in long form",
		"restate line 1",
		"restate line 2",
		"Executor: apply patch in service.go and run tests",
		"patch detail 1",
		"patch detail 2",
		"Tool: INFO output repeated repeated repeated",
		"INFO repeated repeated repeated",
	}, "\n")

	out, saved := f.Apply(input, ModeAggressive)
	if saved <= 0 {
		t.Fatalf("expected savings in aggressive mode")
	}
	if !strings.Contains(out, "Executor:") || !strings.Contains(out, "Planner:") {
		t.Fatalf("expected higher-priority roles to be retained")
	}
}
