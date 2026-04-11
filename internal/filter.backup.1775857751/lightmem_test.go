package filter

import (
	"strings"
	"testing"
)

func TestLightMemFilter_ReusesRepeatedFacts(t *testing.T) {
	f := NewLightMemFilter()
	input := strings.Join([]string{
		"ERROR: migration failed at file db/migrate.go line 88",
		"context line",
		"ERROR: migration failed at file db/migrate.go line 88",
		"path: internal/service/payment.go",
		"path: internal/service/payment.go",
		"another context line",
		"WARN: retrying migration",
		"WARN: retrying migration",
	}, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved <= 0 {
		t.Fatalf("expected savings, got %d", saved)
	}
	if !strings.Contains(out, "[lightmem: reuse") {
		t.Fatalf("expected lightmem reuse marker")
	}
}
