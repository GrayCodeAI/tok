package filter

import (
	"strings"
	"testing"
)

func TestLatentCollabFilter_MergesEquivalentTurns(t *testing.T) {
	f := NewLatentCollabFilter()
	input := strings.Join([]string{
		"Planner: We should inspect the failing handler path and nil guard.",
		"check handler path and nil guard first",
		"Planner: We should inspect the failing handler path and nil guard.",
		"inspect handler path and nil guard first",
		"Executor: apply fix to service.go and rerun tests",
	}, "\n")

	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative token savings")
	}
	if !strings.Contains(out, "latent-collab") {
		t.Fatalf("expected latent-collab marker in output")
	}
	if strings.Contains(out, "inspect handler path and nil guard first") {
		t.Fatalf("expected equivalent duplicate turn to be merged")
	}
}
