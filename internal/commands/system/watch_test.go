package system

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
)

func captureWatchStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	return string(out)
}

func TestRunWatchReturnsWrappedExitError(t *testing.T) {
	out := captureWatchStdout(t, func() {
		err := runWatch(&cobra.Command{}, []string{"sh", "-c", "printf 'watch output\\n'; exit 4"})
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}
		if exitErr.ExitCode() != 4 {
			t.Fatalf("ExitCode() = %d, want 4", exitErr.ExitCode())
		}
	})
	if out == "" {
		t.Fatal("expected watch output")
	}
}

func TestRunWatchHonorsCanceledContext(t *testing.T) {
	cmd := &cobra.Command{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := runWatch(cmd, []string{"sh", "-c", "sleep 5"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runWatch() error = %v, want context.Canceled", err)
	}
}
