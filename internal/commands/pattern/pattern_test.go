package pattern

import (
	"testing"
)

func TestPackageCompiles(t *testing.T) {
	// Smoke test: ensures the package compiles and init() runs without panic.
	if patternCmd == nil {
		t.Fatal("patternCmd should be initialized")
	}
	if patternCmd.Use != "pattern" {
		t.Errorf("expected command name 'pattern', got %q", patternCmd.Use)
	}
}
