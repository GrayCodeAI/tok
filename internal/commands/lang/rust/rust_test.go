package rust

import (
	"testing"
)

func TestRustCmdInitialized(t *testing.T) {
	if rustCmd == nil {
		t.Fatal("rustCmd should be initialized")
	}
	if rustCmd.Use != "rust [subcommand] [args...]" {
		t.Errorf("expected Use 'rust [subcommand] [args...]', got %q", rustCmd.Use)
	}
}
