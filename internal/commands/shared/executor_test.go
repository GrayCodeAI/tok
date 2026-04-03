package shared

import "testing"

func TestRunAndCaptureRequiresCommand(t *testing.T) {
	_, exitCode, err := RunAndCapture("", nil)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}
