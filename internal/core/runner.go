package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
)

// maxOutputSize is the maximum bytes of command output to capture.
// Beyond this, output is truncated with a [truncated] marker.
const maxOutputSize = 100 * 1024 * 1024 // 100 MiB

// shellMetaCharsPattern matches characters that should not appear in a raw command binary name.
// Note: exec.CommandContext does NOT invoke a shell, so shell injection is not possible
// through the binary name or arguments. This regex serves two purposes:
//  1. UX: fail fast with a clear message when the user accidentally includes shell syntax
//  2. Defense-in-depth: if the code is ever changed to use exec.Command("sh", "-c", ...)
//     this check would prevent injection (but that change should never be made).
var shellMetaCharsPattern = regexp.MustCompile("[;&|`\\x01-\\x1f]")

// OSCommandRunner executes real shell commands using os/exec.
type OSCommandRunner struct {
	Env []string
}

// NewOSCommandRunner creates a command runner with the current environment.
func NewOSCommandRunner() *OSCommandRunner {
	return &OSCommandRunner{
		Env: os.Environ(),
	}
}

// validateCommandName checks that the binary name doesn't contain shell
// meta-characters that would indicate the caller intended shell execution.
// exec.CommandContext does NOT use a shell, so these characters will simply
// fail to find the binary -- but we fail fast with a clear error message.
func validateCommandName(name string) error {
	if shellMetaCharsPattern.MatchString(name) {
		return fmt.Errorf("command name %q contains shell meta-characters", name)
	}
	return nil
}

// Run executes a command and captures combined stdout+stderr.
// The binary name is validated for shell meta-characters to provide a clear
// error message; arguments are passed directly to exec.CommandContext which
// does not invoke a shell, making shell injection impossible.
// Output is capped at maxOutputSize to prevent OOM from runaway commands.
func (r *OSCommandRunner) Run(ctx context.Context, args []string) (string, int, error) {
	if len(args) == 0 {
		return "", 0, nil
	}

	if err := validateCommandName(args[0]); err != nil {
		return err.Error(), 126, err
	}

	cmdPath, err := exec.LookPath(args[0])
	if err != nil {
		hint := fmt.Sprintf("command not found: %s", args[0])
		return hint, 127, err
	}

	cmd := exec.CommandContext(ctx, cmdPath, args[1:]...)
	cmd.Env = r.Env

	var buf bytes.Buffer
	lw := &LimitedWriter{W: &buf, N: maxOutputSize}
	cmd.Stdout = lw
	cmd.Stderr = lw

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	output := buf.String()
	if lw.Dropped > 0 {
		output += "\n[truncated: output exceeded 100 MiB]"
	}

	return output, exitCode, err
}

// LookPath resolves a command name to its full path.
func (r *OSCommandRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// LimitedWriter wraps an io.Writer and drops writes after N bytes.
type LimitedWriter struct {
	W       io.Writer
	N       int64
	Dropped int64
}

func (lw *LimitedWriter) Write(p []byte) (n int, err error) {
	if lw.N <= 0 {
		lw.Dropped += int64(len(p))
		return len(p), nil
	}
	if int64(len(p)) > lw.N {
		p = p[:lw.N]
	}
	n, err = lw.W.Write(p)
	lw.N -= int64(n)
	return n, err
}
