package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

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

// Run executes a command and captures combined stdout+stderr.
func (r *OSCommandRunner) Run(ctx context.Context, args []string) (string, int, error) {
	if len(args) == 0 {
		return "", 0, nil
	}

	cmdPath, err := exec.LookPath(args[0])
	if err != nil {
		hint := fmt.Sprintf("command not found: %s\n\nDid you install it? Try:\n  apt install %s  # Debian/Ubuntu\n  brew install %s  # macOS\n  dnf install %s  # Fedora", args[0], args[0], args[0], args[0])
		return hint, 127, err
	}

	cmd := exec.CommandContext(ctx, cmdPath, args[1:]...)
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
