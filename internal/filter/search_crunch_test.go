package filter

import (
	"strings"
	"testing"
)

func TestSearchCrunchFilter_PrunesDuplicateHits(t *testing.T) {
	f := NewSearchCrunchFilter()
	input := strings.Join([]string{
		"1. internal/service/auth.go:88 panic path",
		"2. internal/service/auth.go:88 panic path",
		"3. internal/service/auth.go:88 panic path",
		"4. internal/service/payments.go:42 validation failed",
		"5. internal/service/payments.go:42 validation failed",
		"WARN partial results",
		"x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9", "x10", "x11", "x12",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings")
	}
	if !strings.Contains(out, "search-crunch") {
		t.Fatalf("expected search-crunch marker")
	}
}
