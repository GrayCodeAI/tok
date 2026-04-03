package output

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func captureStdout(t *testing.T, fn func()) string {
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

func testExitErr(t *testing.T, err error, want int) {
	t.Helper()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exec.ExitError, got %v", err)
	}
	if exitErr.ExitCode() != want {
		t.Fatalf("ExitCode() = %d, want %d", exitErr.ExitCode(), want)
	}
}

func TestRunSummaryReturnsWrappedExitError(t *testing.T) {
	out := captureStdout(t, func() {
		err := runSummary(&cobra.Command{}, []string{"sh", "-c", "printf 'summary failed\\n'; exit 7"})
		testExitErr(t, err, 7)
	})
	if out == "" {
		t.Fatal("expected summary output")
	}
}

func TestRunSummarySummarizesFileInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out := captureStdout(t, func() {
		if err := runSummary(&cobra.Command{}, []string{path}); err != nil {
			t.Fatalf("runSummary() error = %v", err)
		}
	})
	if out == "" {
		t.Fatal("expected summary output")
	}
}

func TestRunSummaryStripsKnownGlobalFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out := captureStdout(t, func() {
		err := runSummary(&cobra.Command{}, []string{"--preset", "fast", "--mode", "minimal", "--budget", "50", "--query", "errors", path})
		if err != nil {
			t.Fatalf("runSummary() error = %v", err)
		}
	})
	if out == "" {
		t.Fatal("expected summary output")
	}
}

func TestRunErrReturnsWrappedExitError(t *testing.T) {
	out := captureStdout(t, func() {
		err := runErr([]string{"sh", "-c", "printf 'fatal error\\n' >&2; exit 9"}, false)
		testExitErr(t, err, 9)
	})
	if out == "" {
		t.Fatal("expected filtered error output")
	}
}

func TestRunExplainReturnsWrappedExitError(t *testing.T) {
	out := captureStdout(t, func() {
		err := runExplain(&cobra.Command{}, []string{"sh", "-c", "printf 'explain me\\n'; exit 6"})
		testExitErr(t, err, 6)
	})
	if out == "" {
		t.Fatal("expected explain output")
	}
}

func TestRunSummaryHonorsCanceledContext(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(canceledContext())

	err := runSummary(cmd, []string{"sh", "-c", "sleep 5"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runSummary() error = %v, want context.Canceled", err)
	}
}

func TestRunExplainHonorsCanceledContext(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(canceledContext())

	err := runExplain(cmd, []string{"sh", "-c", "sleep 5"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runExplain() error = %v, want context.Canceled", err)
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
