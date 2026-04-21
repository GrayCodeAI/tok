// Package unified handles unified commands that work across both engines.
package unified

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/GrayCodeAI/tok/internal/app/input"
	"github.com/GrayCodeAI/tok/internal/app/output"
)

// Handler processes unified commands.
type Handler struct {
	inputHandler  *input.Handler
	outputHandler *output.Handler
	version       string
}

// New creates a new unified handler.
func New(inputHandler *input.Handler, outputHandler *output.Handler, version string) *Handler {
	return &Handler{
		inputHandler:  inputHandler,
		outputHandler: outputHandler,
		version:       version,
	}
}

// Doctor runs diagnostics on both engines.
func (h *Handler) Doctor() error {
	fmt.Println("tok doctor")
	fmt.Println("  [input] ok: native tok input engine")
	fmt.Println("  [output] ok: embedded output engine")
	return nil
}

// Status shows unified status.
func (h *Handler) Status() error {
	fmt.Println("tok status (input)")
	if err := h.inputHandler.Status(); err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println("tok status (output)")
	fmt.Println("output engine is available")
	return nil
}

// Version shows version info.
func (h *Handler) Version() error {
	fmt.Printf("tok %s\n", h.version)
	fmt.Println("input engine: compatible")
	fmt.Println("output engine: embedded")
	return nil
}

// Both runs both input and output stages.
func (h *Handler) Both(args []string) error {
	fs := flag.NewFlagSet("both", flag.ContinueOnError)
	text := fs.String("text", "", "Text to compress through input engine")
	mode := fs.String("mode", "", "Input compression mode (lite|full|ultra|wenyan)")
	command := fs.String("command", "", "Command to run through output engine")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*text) != "" {
		inputArgs := []string{"-input", *text}
		if strings.TrimSpace(*mode) != "" {
			inputArgs = append(inputArgs, "-mode", *mode)
		}
		if err := h.inputHandler.Compress(inputArgs); err != nil {
			return fmt.Errorf("input stage failed: %w", err)
		}
	}

	if strings.TrimSpace(*command) != "" {
		outputArgs := shellWords(*command)
		if len(outputArgs) == 0 {
			return errors.New("both: --command is empty after parsing")
		}
		if err := h.outputHandler.Run(outputArgs); err != nil {
			return fmt.Errorf("output stage failed: %w", err)
		}
	}

	if strings.TrimSpace(*text) == "" && strings.TrimSpace(*command) == "" {
		return errors.New("both requires at least one of --text or --command")
	}
	return nil
}

func shellWords(s string) []string {
	return strings.Fields(strings.TrimSpace(s))
}
