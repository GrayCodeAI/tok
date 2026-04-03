package filtercmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
)

func capturePipelineStdout(t *testing.T, fn func()) string {
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

func TestRunPipelineReturnsWrappedExitError(t *testing.T) {
	out := capturePipelineStdout(t, func() {
		err := runPipeline(&cobra.Command{}, []string{"sh", "-c", "printf 'pipeline output\\n'; exit 3"})
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}
		if exitErr.ExitCode() != 3 {
			t.Fatalf("ExitCode() = %d, want 3", exitErr.ExitCode())
		}
	})
	if out == "" {
		t.Fatal("expected pipeline output")
	}
}
