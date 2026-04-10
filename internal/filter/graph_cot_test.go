package filter

import (
	"strings"
	"testing"
)

func TestGraphCoTFilter_PreservesCausalEdges(t *testing.T) {
	f := NewGraphCoTFilter()
	input := strings.Join([]string{
		"Step 1: identify failing endpoint",
		"Because the payload is nil, decoding fails.",
		"therefore we need a nil check before parse",
		"random filler line one",
		"random filler line two",
		"As a result, tests should pass after guard",
		"Step 2: rerun unit tests",
		"PASS",
	}, "\n")

	out, _ := f.Apply(input, ModeMinimal)
	if !strings.Contains(strings.ToLower(out), "therefore") {
		t.Fatalf("expected causal edge line to be preserved")
	}
}
