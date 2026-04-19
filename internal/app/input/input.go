// Package input handles input compression commands for human-written text.
package input

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lakshmanpatel/tok/internal/compressor"
	"github.com/lakshmanpatel/tok/internal/git"
	"github.com/lakshmanpatel/tok/internal/hooks"
	"github.com/lakshmanpatel/tok/internal/review"
)

// Handler processes input commands.
type Handler struct{}

// New creates a new input handler.
func New() *Handler {
	return &Handler{}
}

// Run executes the input command with args.
func (h *Handler) Run(args []string) error {
	if len(args) == 0 {
		return errors.New("input command required. Use: tok input compress -mode ultra -input \"text\"")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "compress", "c":
		return h.Compress(subargs)
	case "terse", "t":
		return h.Terse(subargs)
	case "compact", "cmp":
		return h.Compact(subargs)
	case "activate", "on":
		return h.Activate(subargs)
	case "mode":
		return h.Mode()
	case "deactivate", "off", "normal", "stop", "stop-terse":
		return h.Deactivate()
	case "statusline":
		return h.Statusline()
	case "template", "agent-template":
		return h.Template()
	case "commit", "cm":
		return h.Commit(subargs)
	case "terse-commit", "tc":
		return h.TerseCommit(subargs)
	case "compact-commit", "cc":
		return h.CompactCommit(subargs)
	case "review", "rv":
		return h.Review(subargs)
	case "terse-review", "tr":
		return h.TerseReview(subargs)
	case "compact-review", "cr":
		return h.CompactReview(subargs)
	case "compress-file":
		return h.CompressFile(subargs)
	case "compress-memory", "tf", "compact-file", "cf":
		return h.CompressMemory(subargs)
	case "restore":
		return h.Restore(subargs)
	case "install-agents", "agents-install":
		return h.InstallAgents()
	case "uninstall-agents", "agents-uninstall":
		return h.UninstallAgents()
	case "hooks-install", "hook-install":
		return h.HooksInstall()
	case "hooks-uninstall", "hook-uninstall":
		return h.HooksUninstall()
	case "hooks-install-pwsh", "hook-install-pwsh":
		return h.HooksInstallPwsh()
	case "hooks-uninstall-pwsh", "hook-uninstall-pwsh":
		return h.HooksUninstallPwsh()
	default:
		return fmt.Errorf("unknown input command: %s", subcmd)
	}
}

// Compress compresses input text.
func (h *Handler) Compress(args []string) error {
	fs := flag.NewFlagSet("compress", flag.ContinueOnError)
	input := fs.String("input", "", "Input text (or use stdin)")
	mode := fs.String("mode", "full", "Compression mode: lite, full, ultra")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var text string
	if *input != "" {
		text = *input
	} else {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		text = string(bytes)
	}

	compressed, err := compressor.Compress(text, *mode)
	if err != nil {
		return err
	}

	fmt.Print(compressed)
	return nil
}

// Terse activates terse mode.
func (h *Handler) Terse(args []string) error {
	mode := "full"
	if len(args) > 0 {
		mode = args[0]
	}

	if err := hooks.Activate(mode); err != nil {
		return fmt.Errorf("failed to activate terse mode: %w", err)
	}

	fmt.Printf("Terse mode activated: %s\n", mode)
	fmt.Println("Run 'tok off' to deactivate")
	return nil
}

// Compact activates compact mode.
func (h *Handler) Compact(args []string) error {
	if err := hooks.Activate("full"); err != nil {
		return fmt.Errorf("failed to activate compact mode: %w", err)
	}
	fmt.Println("Compact mode activated")
	return nil
}

// Activate activates tok with mode.
func (h *Handler) Activate(args []string) error {
	mode := "full"
	if len(args) > 0 {
		mode = args[0]
	}

	if err := hooks.Activate(mode); err != nil {
		return fmt.Errorf("failed to activate: %w", err)
	}

	fmt.Printf("tok activated: %s mode\n", mode)
	return nil
}

// Mode shows current mode.
func (h *Handler) Mode() error {
	if !hooks.IsActive() {
		fmt.Println("tok mode: inactive")
		return nil
	}

	mode := hooks.GetMode()
	if mode == "" {
		mode = "full"
	}
	fmt.Printf("tok mode: %s\n", mode)
	return nil
}

// Deactivates tok.
func (h *Handler) Deactivate() error {
	if err := hooks.Deactivate(); err != nil {
		return fmt.Errorf("failed to deactivate: %w", err)
	}
	fmt.Println("tok deactivated")
	return nil
}

// Status returns input status.
func (h *Handler) Status() error {
	if hooks.IsActive() {
		fmt.Printf("tok input status: active (%s mode)\n", hooks.GetMode())
	} else {
		fmt.Println("tok input status: inactive")
	}
	return nil
}

// Statusline returns statusline badge.
func (h *Handler) Statusline() error {
	badge := hooks.GetStatusLine()
	if badge == "" {
		badge = "[TOK:off]"
	}
	fmt.Println(badge)
	return nil
}

// Template shows agent template.
func (h *Handler) Template() error {
	fmt.Println("Template: not implemented yet")
	return nil
}

// Commit generates commit message.
func (h *Handler) Commit(args []string) error {
	msg, err := git.GenerateCommitMessage()
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}
	fmt.Println(msg)
	return nil
}

// TerseCommit generates terse commit.
func (h *Handler) TerseCommit(args []string) error {
	return h.Commit(args)
}

// CompactCommit generates compact commit.
func (h *Handler) CompactCommit(args []string) error {
	return h.Commit(args)
}

// Review generates code review.
func (h *Handler) Review(args []string) error {
	results, err := review.GenerateReview()
	if err != nil {
		return fmt.Errorf("failed to generate review: %w", err)
	}

	formatted := review.FormatReview(results)
	fmt.Println(formatted)
	return nil
}

// TerseReview generates terse review.
func (h *Handler) TerseReview(args []string) error {
	return h.Review(args)
}

// CompactReview generates compact review.
func (h *Handler) CompactReview(args []string) error {
	return h.Review(args)
}

// CompressFile compresses a file.
func (h *Handler) CompressFile(args []string) error {
	fs := flag.NewFlagSet("compress-file", flag.ContinueOnError)
	file := fs.String("file", "", "File to compress")
	mode := fs.String("mode", "full", "Compression mode")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return errors.New("file required: -file <path>")
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	compressed, err := compressor.Compress(string(content), *mode)
	if err != nil {
		return err
	}

	fmt.Print(compressed)
	return nil
}

// CompressMemory compresses file to memory format.
func (h *Handler) CompressMemory(args []string) error {
	fs := flag.NewFlagSet("compress-memory", flag.ContinueOnError)
	file := fs.String("file", "", "File to compress to memory format")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" && len(fs.Args()) > 0 {
		*file = fs.Args()[0]
	}

	if *file == "" {
		return errors.New("file required")
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	compressed, err := compressor.Compress(string(content), "ultra")
	if err != nil {
		return err
	}

	fmt.Printf("// %s (compressed)\n", *file)
	fmt.Print(compressed)
	return nil
}

// Restore restores compressed content.
func (h *Handler) Restore(args []string) error {
	fmt.Println("Restore: not implemented yet")
	return nil
}

// InstallAgents installs agents.
func (h *Handler) InstallAgents() error {
	fmt.Println("Install agents: not implemented yet")
	return nil
}

// UninstallAgents uninstalls agents.
func (h *Handler) UninstallAgents() error {
	fmt.Println("Uninstall agents: not implemented yet")
	return nil
}

// HooksInstall installs hooks.
func (h *Handler) HooksInstall() error {
	fmt.Println("Hooks install: not implemented yet")
	return nil
}

// HooksUninstall uninstalls hooks.
func (h *Handler) HooksUninstall() error {
	fmt.Println("Hooks uninstall: not implemented yet")
	return nil
}

// HooksInstallPwsh installs PowerShell hooks.
func (h *Handler) HooksInstallPwsh() error {
	fmt.Println("Hooks install pwsh: not implemented yet")
	return nil
}

// HooksUninstallPwsh uninstalls PowerShell hooks.
func (h *Handler) HooksUninstallPwsh() error {
	fmt.Println("Hooks uninstall pwsh: not implemented yet")
	return nil
}
