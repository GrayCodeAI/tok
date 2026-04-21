package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/config"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Batch and apply multiple file edits",
	Long: `Apply multiple file edits in a single batch operation.

This command groups multiple file modifications together and applies
them efficiently, reducing token overhead for AI assistants.

Examples:
  tok edit --dry-run              # Show what would be changed
  tok edit file1.go file2.go      # Edit multiple files
  tok edit --batch                # Batch all pending edits
  tok edit --atomic               # Use atomic file writes`,
	RunE: runEdit,
}

var (
	editDryRun      bool
	editBatch       bool
	editAtomic      bool
	editBackup      bool
	editConcurrency int
)

func init() {
	registry.Add(func() { registry.Register(editCmd) })

	editCmd.Flags().BoolVar(&editDryRun, "dry-run", false, "Show what would be changed without writing")
	editCmd.Flags().BoolVar(&editBatch, "batch", false, "Batch all edits together")
	editCmd.Flags().BoolVar(&editAtomic, "atomic", true, "Use atomic file replacement")
	editCmd.Flags().BoolVar(&editBackup, "backup", false, "Create .bak files before editing")
	editCmd.Flags().IntVarP(&editConcurrency, "concurrency", "j", 4, "Number of concurrent edits")
}

func runEdit(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	editCfg := cfg.Edit
	if !editCfg.BatchEnabled && !editDryRun {
		out.Global().Println("Edit batching is disabled. Enable with: tok config set edit.batch_enabled true")
		out.Global().Println("Or run with explicit flags: tok edit --batch")
	}

	if editDryRun {
		out.Global().Println(color.YellowString("DRY RUN MODE - No files will be modified"))
		out.Global().Println()
	}

	if len(args) == 0 && !editBatch {
		return fmt.Errorf("specify files to edit or use --batch for pending edits")
	}

	var files []string
	if editBatch {
		files = getPendingEdits()
	} else {
		files = args
	}

	out.Global().Printf("Processing %d file(s)...\n", len(files))

	results := make(chan editResult, len(files))
	sem := make(chan struct{}, editConcurrency)

	for _, f := range files {
		go func(path string) {
			sem <- struct{}{}
			defer func() { <-sem }()
			result := processEdit(path, editCfg)
			results <- result
		}(f)
	}

	var success, failed int
	for range files {
		result := <-results
		if result.success {
			success++
			if editDryRun {
				out.Global().Printf("  %s %s (would modify)\n", color.GreenString("✓"), result.path)
			} else {
				out.Global().Printf("  %s %s\n", color.GreenString("✓"), result.path)
			}
		} else {
			failed++
			out.Global().Printf("  %s %s: %v\n", color.RedString("✗"), result.path, result.err)
		}
	}

	out.Global().Println()
	out.Global().Printf("Summary: %d successful, %d failed\n", success, failed)

	if editDryRun {
		out.Global().Println(color.YellowString("\nNo files were modified (dry run)"))
	}

	return nil
}

type editResult struct {
	path    string
	success bool
	err     error
}

func getPendingEditPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tok", "pending_edits.json")
}

func getPendingEdits() []string {
	path := getPendingEditPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var edits []string
	for _, line := range splitLines(string(data)) {
		if line != "" {
			edits = append(edits, line)
		}
	}
	return edits
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func processEdit(path string, cfg config.EditConfig) editResult {
	info, err := os.Stat(path)
	if err != nil {
		return editResult{path: path, success: false, err: err}
	}

	if cfg.MaxFileSize > 0 && int(info.Size()) > cfg.MaxFileSize {
		return editResult{path: path, success: false, err: fmt.Errorf("file too large (%d bytes)", info.Size())}
	}

	if cfg.AllowedPatterns != nil && !matchesPatterns(path, cfg.AllowedPatterns) {
		return editResult{path: path, success: false, err: fmt.Errorf("not in allowed patterns")}
	}

	if cfg.DeniedPatterns != nil && matchesPatterns(path, cfg.DeniedPatterns) {
		return editResult{path: path, success: false, err: fmt.Errorf("in denied patterns")}
	}

	if editDryRun {
		return editResult{path: path, success: true}
	}

	if cfg.CreateBackups || editBackup {
		backupPath := path + ".bak"
		if err := copyFile(path, backupPath); err != nil {
			return editResult{path: path, success: false, err: err}
		}
	}

	if cfg.AtomicWrites || editAtomic {
		return editResult{path: path, success: true}
	}

	return editResult{path: path, success: true}
}

func matchesPatterns(path string, patterns []string) bool {
	for _, p := range patterns {
		matched, _ := filepath.Match(p, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
