package filtercmd

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	out "github.com/lakshmanpatel/tok/internal/output"
	"github.com/spf13/cobra"
)

func TestRunPipelineReturnsWrappedExitError(t *testing.T) {
	var buf bytes.Buffer
	testPrinter := out.NewTest(&buf, &buf)
	prev := out.SetGlobal(testPrinter)
	defer out.SetGlobal(prev)

	err := runPipeline(&cobra.Command{}, []string{"sh", "-c", "printf 'pipeline output\\n'; exit 3"})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exec.ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 3 {
		t.Fatalf("ExitCode() = %d, want 3", exitErr.ExitCode())
	}
	if buf.Len() == 0 {
		t.Fatal("expected pipeline output")
	}
}
