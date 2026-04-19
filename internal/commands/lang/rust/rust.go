package rust

import (
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

func init() {
	registry.Add(func() {
		registry.Register(rustCmd)
	})
}

var rustCmd = &cobra.Command{
	Use:   "rust [subcommand] [args...]",
	Short: "Rust toolchain with compact output",
	Long:  `Rust toolchain commands with token-optimized output. Supports cargo, rustc, rustup, etc.`,
}

func runRust(args []string) error {
	if len(args) == 0 {
		return rustCmd.Help()
	}

	subcommand := args[0]
	timer := tracking.Start()

	cmd := exec.Command("rustc", args...)
	output, err := cmd.CombinedOutput()
	raw := string(output)

	filtered := raw

	switch subcommand {
	case "build", "b":
		filtered = filterRustBuild(raw)
	case "run":
		filtered = filterRustRun(raw)
	case "test":
		filtered = filterRustTest(raw)
	case "check":
		filtered = filterRustCheck(raw)
	case "doc":
		filtered = filterRustDoc(raw)
	case "clippy":
		filtered = filterRustClippy(raw)
	case "fmt":
		filtered = filterRustFmt(raw)
	default:
		filtered = raw
	}

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("rust %s", subcommand), "tok rust", originalTokens, filteredTokens)

	if shared.IsUltraCompact() {
		filtered = compactRustOutput(filtered, subcommand)
	}

	out.Global().Print(filtered)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "rust_"+subcommand, err); hint != "" {
			out.Global().Print("\n" + hint)
		}
		return err
	}

	return nil
}

func filterRustBuild(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string
	showedCount := 0

	for _, line := range lines {
		if strings.Contains(line, "Compiling") {
			if shared.IsUltraCompact() {
				parts := strings.Split(line, "Compiling ")
				if len(parts) > 1 {
					name := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if showedCount < 3 {
						filtered = append(filtered, "[+] "+name)
						showedCount++
					}
				}
			} else {
				filtered = append(filtered, line)
			}
		} else if strings.Contains(line, "Finished") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "error") || strings.Contains(line, "warning:") {
			filtered = append(filtered, line)
		} else if strings.HasPrefix(line, "   ") || strings.HasPrefix(line, "  ") {
		} else if strings.Contains(line, "=") && len(line) < 80 {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func filterRustRun(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "Running") || strings.HasPrefix(line, "Finished") {
			continue
		}
		if strings.Contains(line, "error") || strings.Contains(line, "thread ") {
			filtered = append(filtered, line)
		} else if line != "" {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func filterRustTest(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string
	passed := 0
	failed := 0

	for _, line := range lines {
		if strings.Contains(line, "test result: ok") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "FAILED") {
			failed++
			filtered = append(filtered, line)
		} else if strings.Contains(line, "running") {
			filtered = append(filtered, line)
		} else if shared.IsUltraCompact() && strings.Contains(line, "test ") && strings.Contains(line, " ... ") {
			parts := strings.Split(line, " ... ")
			if len(parts) == 2 {
				status := strings.TrimSpace(parts[1])
				if strings.HasPrefix(status, "ok") {
					passed++
				}
			}
		} else if !shared.IsUltraCompact() {
			if strings.Contains(line, "test ") || strings.Contains(line, "running") || strings.Contains(line, "PASSED") || strings.Contains(line, "FAILED") {
				filtered = append(filtered, line)
			}
		}
	}

	if shared.IsUltraCompact() && (passed > 0 || failed > 0) {
		return fmt.Sprintf("tests: %d passed, %d failed", passed, failed)
	}

	return strings.Join(filtered, "\n")
}

func filterRustCheck(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, "error") || strings.Contains(line, "warning:") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "Checking") || strings.Contains(line, "Finished") {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func filterRustDoc(raw string) string {
	if shared.IsUltraCompact() {
		return "rust doc generated"
	}
	return raw
}

func filterRustClippy(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string
	warnCount := 0

	for _, line := range lines {
		if strings.Contains(line, "warning:") {
			warnCount++
			if shared.IsUltraCompact() && warnCount <= 3 {
				parts := strings.Split(line, "warning:")
				if len(parts) > 1 {
					msg := strings.TrimSpace(parts[1])
					if len(msg) > 60 {
						msg = msg[:60]
					}
					filtered = append(filtered, "warn: "+msg)
				}
			} else if !shared.IsUltraCompact() {
				filtered = append(filtered, line)
			}
		} else if strings.Contains(line, "Finished") || strings.Contains(line, "error") {
			filtered = append(filtered, line)
		}
	}

	if shared.IsUltraCompact() {
		return fmt.Sprintf("clippy: %d warnings", warnCount)
	}

	return strings.Join(filtered, "\n")
}

func filterRustFmt(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, "error") || strings.Contains(line, "warning:") {
			filtered = append(filtered, line)
		}
	}

	if len(filtered) == 0 {
		return "fmt: no issues"
	}

	return strings.Join(filtered, "\n")
}

func compactRustOutput(filtered, subcommand string) string {
	lines := strings.Split(filtered, "\n")
	if len(lines) > 10 {
		return strings.Join(lines[:10], "\n") + fmt.Sprintf("\n... (+%d lines)", len(lines)-10)
	}
	return filtered
}
