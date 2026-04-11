package filter

import (
	"strings"
	"testing"
)

func TestPlanBudgetFilter_AnnotatesAndCompresses(t *testing.T) {
	f := NewPlanBudgetFilter()
	input := strings.Join([]string{
		"Step 1: parse stack trace",
		"stack trace shows panic in auth middleware",
		"therefore adjust plan and budget for deeper retention",
		"filler one", "filler two", "filler three", "filler four",
		"Step 2: apply patch",
		"ERROR: test_login fails with 401",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if !strings.Contains(out, "[plan-budget:") {
		t.Fatalf("expected plan-budget marker")
	}
	if saved < 0 {
		t.Fatalf("expected non-negative savings, got %d", saved)
	}
}
