package swift

import (
	"testing"
)

func TestSwiftCmdInitialized(t *testing.T) {
	if swiftCmd == nil {
		t.Fatal("swiftCmd should be initialized")
	}
	if swiftCmd.Use != "swift [args...]" {
		t.Errorf("expected Use 'swift [args...]', got %q", swiftCmd.Use)
	}
}
