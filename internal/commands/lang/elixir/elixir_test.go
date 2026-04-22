package elixir

import (
	"testing"
)

func TestElixirCmdInitialized(t *testing.T) {
	if elixirCmd == nil {
		t.Fatal("elixirCmd should be initialized")
	}
	if elixirCmd.Use != "elixir [args...]" {
		t.Errorf("expected Use 'elixir [args...]', got %q", elixirCmd.Use)
	}
}
