package filter

import (
	"strings"
	"testing"
)

func TestSWEAdaptiveLoopFilter_CompressesVerboseTrace(t *testing.T) {
	f := NewSWEAdaptiveLoopFilter()
	input := strings.Join([]string{
		"Planner: diagnose failing request handler",
		"noise line 1", "noise line 2", "noise line 3", "noise line 4",
		"ERROR: nil pointer dereference in handler.go:74",
		"Executor: add nil guard and rerun tests",
		"noise line 5", "noise line 6", "noise line 7", "noise line 8",
		"Result: tests passed",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved <= 0 {
		t.Fatalf("expected token savings, got %d", saved)
	}
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected critical error line to be preserved")
	}
}
