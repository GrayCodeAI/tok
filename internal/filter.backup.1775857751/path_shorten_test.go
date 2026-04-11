package filter

import (
	"strings"
	"testing"
)

func TestPathShortenFilter_AliasesRepeatedPaths(t *testing.T) {
	f := NewPathShortenFilter()
	input := strings.Join([]string{
		"ERROR at internal/services/payment/handler/process.go:82",
		"check internal/services/payment/handler/process.go for guard",
		"identifier verylongidentifierwithmanycharactersandnumbers1234 failed",
		"retry verylongidentifierwithmanycharactersandnumbers1234 after fix",
		"tail", "tail2", "tail3", "tail4", "tail5",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings")
	}
	if !strings.Contains(out, "@p") && !strings.Contains(out, "@id") {
		t.Fatalf("expected alias replacement markers")
	}
}
