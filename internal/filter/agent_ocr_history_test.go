package filter

import (
	"strings"
	"testing"
)

func TestAgentOCRHistoryFilter_CompactsOldTurns(t *testing.T) {
	f := NewAgentOCRHistoryFilter()
	input := strings.Join([]string{
		"Planner: investigate flaky migration",
		"long low value discussion line a", "long low value discussion line b", "long low value discussion line c",
		"Critic: restating known risks",
		"restate x", "restate y", "restate z",
		"Executor: apply migration in db/migrate.go",
		"run migration checksum verification",
		"Reviewer: approve patch and merge",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings, got %d", saved)
	}
	if !strings.Contains(out, "agent-ocr-history") {
		t.Fatalf("expected compact marker in output")
	}
	if !strings.Contains(out, "Executor:") {
		t.Fatalf("expected recent turn to be preserved")
	}
}
