package web

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestRunProxyReturnsExitError(t *testing.T) {
	err := runProxy([]string{"sh", "-c", "printf 'proxy output\\n'; exit 8"})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exec.ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 8 {
		t.Fatalf("ExitCode() = %d, want 8", exitErr.ExitCode())
	}
}

func TestRunProxyHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runProxyContext(ctx, []string{"sh", "-c", "sleep 5"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runProxyContext() error = %v, want context.Canceled", err)
	}
}
