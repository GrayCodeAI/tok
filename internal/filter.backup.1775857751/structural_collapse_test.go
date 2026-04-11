package filter

import (
	"strings"
	"testing"
)

func TestStructuralCollapseFilter_PrunesRepeatedBoilerplate(t *testing.T) {
	f := NewStructuralCollapseFilter()
	input := strings.Join([]string{
		"import os",
		"import os",
		"import os",
		"module payments",
		"module payments",
		"module payments",
		"section setup",
		"section setup",
		"section setup",
		"ERROR: failed to load config",
		"detail 1", "detail 2", "detail 3", "detail 4", "detail 5", "detail 6", "detail 7", "detail 8",
	}, "\n")
	out, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Fatalf("expected non-negative savings")
	}
	if !strings.Contains(out, "structural-collapse") {
		t.Fatalf("expected structural-collapse marker")
	}
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected errors preserved")
	}
}
