package test

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	out "github.com/lakshmanpatel/tok/internal/output"
	"github.com/spf13/cobra"
)

func TestRunCompressTestReturnsWrappedExitError(t *testing.T) {
	var buf bytes.Buffer
	old := out.SetGlobal(out.NewTest(&buf, &buf))
	defer out.SetGlobal(old)

	err := runCompressTest(&cobra.Command{}, []string{"sh", "-c", "printf 'compress output\\n'; exit 5"})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exec.ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 5 {
		t.Fatalf("ExitCode() = %d, want 5", exitErr.ExitCode())
	}
	if buf.Len() == 0 {
		t.Fatal("expected compression test output")
	}
}
