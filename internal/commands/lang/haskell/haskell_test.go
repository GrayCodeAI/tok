package haskell

import (
	"testing"
)

func TestHaskellCmdInitialized(t *testing.T) {
	if haskellCmd == nil {
		t.Fatal("haskellCmd should be initialized")
	}
	if haskellCmd.Use != "haskell [args...]" {
		t.Errorf("expected Use 'haskell [args...]', got %q", haskellCmd.Use)
	}
}
