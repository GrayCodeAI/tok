package filter

import (
	"strings"
	"testing"
)

func TestLogCrunchFilter_FoldsRepeatingLogs(t *testing.T) {
	f := NewLogCrunchFilter()
	input := strings.Join([]string{
		"INFO request completed path=/api/v1/items duration=45ms",
		"INFO request completed path=/api/v1/items duration=45ms",
		"INFO request completed path=/api/v1/items duration=45ms",
		"WARN retrying request",
		"ERROR request failed code=500",
		"tail1", "tail2", "tail3", "tail4", "tail5", "tail6", "tail7", "tail8", "tail9", "tail10", "tail11", "tail12", "tail13", "tail14", "tail15", "tail16",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings")
	}
	if !strings.Contains(out, "log-crunch") {
		t.Fatalf("expected log-crunch marker")
	}
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected errors preserved")
	}
}
