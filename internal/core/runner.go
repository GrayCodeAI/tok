package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// shellMetaCharsPattern matches characters that should not appear in a raw command binary name.
// These characters indicate shell interpretation and would bypass argument sanitization.
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

// sanitizeArgs removes control characters from arguments to prevent
// command injection when arguments come from untrusted sources.
func sanitizeArgs(arg string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' {
			return -1
		}
		return r
	}, arg)
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
// Arguments are sanitized to prevent command injection from untrusted sources.
func (r *OSCommandRunner) Run(ctx context.Context, args []string) (string, int, error) {
	if len(args) == 0 {
		return "", 0, nil
	}

	if err := validateCommandName(args[0]); err != nil {
		return err.Error(), 126, err
	}

	// Sanitize all arguments
	safeArgs := make([]string, len(args))
	safeArgs[0] = args[0]
	for i, arg := range args[1:] {
		safeArgs[i+1] = sanitizeArgs(arg)
	}

	cmdPath, err := exec.LookPath(safeArgs[0])
	if err != nil {
		hint := fmt.Sprintf("command not found: %s", args[0])
		return hint, 127, err
	}

	cmd := exec.CommandContext(ctx, cmdPath, safeArgs[1:]...)
	cmd.Env = r.Env

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return string(output), exitCode, err
}

// LookPath resolves a command name to its full path.
func (r *OSCommandRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
