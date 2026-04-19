// Package input handles input compression commands for human-written text.
package input

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lakshmanpatel/tok/internal/compressor"
	"github.com/lakshmanpatel/tok/internal/git"
	"github.com/lakshmanpatel/tok/internal/hooks"
	"github.com/lakshmanpatel/tok/internal/review"
)

//go:embed agents
var agentsFS embed.FS

const maxInputSize = 50 << 20 // 50 MB

// validateScriptPath checks that a script path is safe to execute.
// It verifies the path is absolute, exists as a regular file, is within the
// expected hooks directory (preventing path traversal), and is not writable
// by others (preventing tampering).
func validateScriptPath(scriptPath, hooksDir string) error {
	// Check that scriptPath is absolute
	if !filepath.IsAbs(scriptPath) {
		return fmt.Errorf("script path must be absolute: %s", scriptPath)
	}

	// Check that scriptPath exists and is a regular file
	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("script path not accessible: %s: %w", scriptPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("script path is not a regular file: %s", scriptPath)
	}

	// Check that scriptPath is within the expected hooks directory
	absHooksDir, err := filepath.Abs(hooksDir)
	if err != nil {
		return fmt.Errorf("cannot resolve hooks directory: %w", err)
	}
	absScriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("cannot resolve script path: %w", err)
	}
	rel, err := filepath.Rel(absHooksDir, absScriptPath)
	if err != nil {
		return fmt.Errorf("cannot compute relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("script path is outside hooks directory (path traversal): %s", scriptPath)
	}

	// Check that scriptPath is not writable by others
	perm := info.Mode().Perm()
	if perm&0002 != 0 {
		return fmt.Errorf("script path is world-writable (potential tampering): %s", scriptPath)
	}

	return nil
}

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
		limited := io.LimitReader(os.Stdin, maxInputSize+1)
		bytes, err := io.ReadAll(limited)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		if len(bytes) > maxInputSize {
			return fmt.Errorf("input exceeds maximum size of %d bytes (%.0f MB)", maxInputSize, float64(maxInputSize)/1024/1024)
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
	entries, err := agentsFS.ReadDir("agents")
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	fmt.Println("Available agent templates:")
	for _, entry := range entries {
		if entry.IsDir() {
			subEntries, _ := agentsFS.ReadDir(filepath.Join("agents", entry.Name()))
			for _, sub := range subEntries {
				fmt.Printf("  %s/%s\n", entry.Name(), sub.Name())
			}
		}
	}
	fmt.Println()
	fmt.Println("Usage: tok install-agents <agent>  (e.g., tok install-agents cursor)")
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
	if len(args) == 0 {
		return errors.New("file required: tok restore <path>")
	}

	filename := args[0]
	if err := compressor.RestoreFile(filename); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}
	return nil
}

// InstallAgents installs agents.
func (h *Handler) InstallAgents() error {
	agentDirs, err := agentsFS.ReadDir("agents")
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, agentDir := range agentDirs {
		if !agentDir.IsDir() {
			continue
		}

		agentName := agentDir.Name()
		targetDir := getAgentTargetDir(agentName)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create dir for %s: %v\n", agentName, err)
			continue
		}

		subEntries, _ := agentsFS.ReadDir(filepath.Join("agents", agentName))
		for _, sub := range subEntries {
			if sub.IsDir() {
				continue
			}
			srcPath := filepath.Join("agents", agentName, sub.Name())
			data, err := agentsFS.ReadFile(srcPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", srcPath, err)
				continue
			}
			dstPath := filepath.Join(targetDir, sub.Name())
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", dstPath, err)
				continue
			}
			fmt.Printf("Installed: %s/%s → %s\n", agentName, sub.Name(), dstPath)
		}
	}

	fmt.Println("All agent rules installed.")
	return nil
}

// UninstallAgents uninstalls agents.
func (h *Handler) UninstallAgents() error {
	agentDirs, err := agentsFS.ReadDir("agents")
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, agentDir := range agentDirs {
		if !agentDir.IsDir() {
			continue
		}

		agentName := agentDir.Name()
		targetDir := getAgentTargetDir(agentName)

		subEntries, _ := agentsFS.ReadDir(filepath.Join("agents", agentName))
		for _, sub := range subEntries {
			if sub.IsDir() {
				continue
			}
			dstPath := filepath.Join(targetDir, sub.Name())
			if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", dstPath, err)
				continue
			}
			fmt.Printf("Removed: %s\n", dstPath)
		}
	}

	fmt.Println("All agent rules removed.")
	return nil
}

// HooksInstall installs hooks.
func (h *Handler) HooksInstall() error {
	hooksDir := getHooksDir()
	scriptPath := filepath.Join(hooksDir, "hooks", "install.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hooks/install.sh not found at %s", scriptPath)
	}
	if err := validateScriptPath(scriptPath, hooksDir); err != nil {
		return fmt.Errorf("hook script validation failed: %w", err)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run install.sh: %w", err)
	}
	return nil
}

// HooksUninstall uninstalls hooks.
func (h *Handler) HooksUninstall() error {
	hooksDir := getHooksDir()
	scriptPath := filepath.Join(hooksDir, "hooks", "uninstall.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hooks/uninstall.sh not found at %s", scriptPath)
	}
	if err := validateScriptPath(scriptPath, hooksDir); err != nil {
		return fmt.Errorf("hook script validation failed: %w", err)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run uninstall.sh: %w", err)
	}
	return nil
}

// HooksInstallPwsh installs PowerShell hooks.
func (h *Handler) HooksInstallPwsh() error {
	hooksDir := getHooksDir()
	scriptPath := filepath.Join(hooksDir, "hooks", "install.ps1")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hooks/install.ps1 not found at %s", scriptPath)
	}
	if err := validateScriptPath(scriptPath, hooksDir); err != nil {
		return fmt.Errorf("hook script validation failed: %w", err)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	} else {
		cmd = exec.Command("pwsh", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run install.ps1: %w", err)
	}
	return nil
}

// HooksUninstallPwsh uninstalls PowerShell hooks.
func (h *Handler) HooksUninstallPwsh() error {
	hooksDir := getHooksDir()
	scriptPath := filepath.Join(hooksDir, "hooks", "uninstall.ps1")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hooks/uninstall.ps1 not found at %s", scriptPath)
	}
	if err := validateScriptPath(scriptPath, hooksDir); err != nil {
		return fmt.Errorf("hook script validation failed: %w", err)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	} else {
		cmd = exec.Command("pwsh", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run uninstall.ps1: %w", err)
	}
	return nil
}

func getAgentTargetDir(agentName string) string {
	switch agentName {
	case "cursor":
		return filepath.Join(os.Getenv("HOME"), ".cursor", "rules")
	case "windsurf":
		return filepath.Join(os.Getenv("HOME"), ".codeium", "windsurf")
	case "cline":
		return filepath.Join(os.Getenv("HOME"), ".claude", "rules")
	case "copilot":
		return filepath.Join(os.Getenv("HOME"), ".github", "copilot-instructions")
	case "claude-code":
		return filepath.Join(os.Getenv("HOME"), ".claude")
	case "aider":
		return "."
	case "continue":
		return filepath.Join(os.Getenv("HOME"), ".continue")
	case "roo-code":
		return filepath.Join(os.Getenv("HOME"), ".roo", "rules")
	case "cody":
		return filepath.Join(os.Getenv("HOME"), ".sourcegraph", "cody")
	case "code-whisperer":
		return filepath.Join(os.Getenv("HOME"), ".aws", "codewhisperer")
	case "tabnine":
		return filepath.Join(os.Getenv("HOME"), ".tabnine")
	case "codeium":
		return filepath.Join(os.Getenv("HOME"), ".codeium")
	default:
		return filepath.Join(os.Getenv("HOME"), ".config", "tok", "agents", agentName)
	}
}

func getHooksDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}
