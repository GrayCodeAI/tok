// Package output handles output filtering commands for terminal/tool output.
package output

import (
	"fmt"

	tokcli "github.com/lakshmanpatel/tok/pkg/cli"
)

// Handler processes output commands.
type Handler struct {
	version string
}

// New creates a new output handler.
func New(version string) *Handler {
	return &Handler{version: version}
}

// Run executes the output command with args.
func (h *Handler) Run(args []string) error {
	if code := tokcli.Run(args, h.version); code != 0 {
		return fmt.Errorf("output command failed with exit code %d", code)
	}
	return nil
}

// Status shows output engine status.
func (h *Handler) Status() error {
	// Status is handled by the output engine itself
	return h.Run([]string{"status"})
}
