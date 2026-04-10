package filter

import (
	"strings"
	"testing"
)

func TestDiffCrunchFilter_FoldsContext(t *testing.T) {
	f := NewDiffCrunchFilter()
	input := strings.Join([]string{
		"diff --git a/a.go b/a.go",
		"--- a/a.go",
		"+++ b/a.go",
		"@@ -1,10 +1,10 @@",
		" context 1",
		" context 2",
		" context 3",
		" context 4",
		"-old line",
		"+new line",
		" context 5",
		" context 6",
		" context 7",
		" context 8",
		" context 9",
		" context 10",
		"@@ -20,6 +20,6 @@",
		" context 11",
		" context 12",
		" context 13",
		"-old 2",
		"+new 2",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings")
	}
	if !strings.Contains(out, "diff-crunch") {
		t.Fatalf("expected diff-crunch marker")
	}
}
