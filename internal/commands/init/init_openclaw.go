package initpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const openclawRulesContent = `<!-- tokman-openclaw-rules v1 -->
# TokMan — Token-Optimized Commands

## Golden Rule

**Always prefix shell commands with ` + "`tokman`" + `** for 60-90% token savings.

## Key Commands

` + "```bash" + `
tokman git status / log / diff / show   # Git (59-80% savings)
tokman cargo build / test / clippy      # Rust (80-90%)
tokman tsc / lint / prettier            # JS/TS (70-87%)
tokman vitest / playwright              # Tests (90-99%)
tokman rspec / rubocop / rake test      # Ruby (60-90%)
tokman docker / kubectl                 # Infra (85%)
tokman ls / grep / find / tree          # Files (60-75%)
tokman gain                             # View savings stats
tokman discover                         # Find missed savings
tokman proxy <cmd>                      # Run without filtering
` + "```" + `

## Examples

Instead of:
` + "```bash" + `
git status
cargo test
rspec spec/
` + "```" + `

Use:
` + "```bash" + `
tokman git status
tokman cargo test
tokman rspec spec/
` + "```" + `

<!-- /tokman-openclaw-rules -->
`

// runOpenClawInit sets up OpenClaw integration (project-scoped)
func runOpenClawInit(global bool) {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var rulesPath string

	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			return
		}
		openclawDir := filepath.Join(homeDir, ".openclaw")
		if err := os.MkdirAll(openclawDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating OpenClaw directory: %v\n", err)
			return
		}
		rulesPath = filepath.Join(openclawDir, "tokman-rules.md")
	} else {
		openclawDir := "openclaw"
		if err := os.MkdirAll(openclawDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating openclaw/ directory: %v\n", err)
			return
		}
		rulesPath = filepath.Join(openclawDir, "tokman-rules.md")
	}

	if err := writeIfChanged(rulesPath, openclawRulesContent, "tokman-rules.md"); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing OpenClaw rules: %v\n", err)
		return
	}

	fmt.Printf("\n%s\n\n", green("TokMan configured for OpenClaw."))
	fmt.Printf("  Rules: %s\n", cyan(rulesPath))
	if global {
		fmt.Println("  Installed globally.")
	} else {
		fmt.Println("  Installed in project openclaw/ directory.")
	}
	fmt.Println()
}

// uninstallOpenClaw removes OpenClaw artifacts
func uninstallOpenClaw() []string {
	var removed []string

	// Check project-local
	localPath := filepath.Join("openclaw", "tokman-rules.md")
	if data, err := os.ReadFile(localPath); err == nil {
		if strings.Contains(string(data), "tokman") {
			if err := os.Remove(localPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", localPath, err)
			}
			removed = append(removed, fmt.Sprintf("OpenClaw rules: %s", localPath))
		}
	}

	// Check global
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return removed
	}
	globalPath := filepath.Join(homeDir, ".openclaw", "tokman-rules.md")
	if data, err := os.ReadFile(globalPath); err == nil {
		if strings.Contains(string(data), "tokman") {
			if err := os.Remove(globalPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", globalPath, err)
			}
			removed = append(removed, fmt.Sprintf("OpenClaw rules: %s", globalPath))
		}
	}

	return removed
}
