package system

import (
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var justCmd = &cobra.Command{
	Use:   "just [recipe] [args...]",
	Short: "Just command runner with compact output",
	Long: `Execute Just recipes with token-optimized output.

Specialized filters for:
  - Recipe execution: Compact output
  - List: Compact recipe listing

Examples:
  tok just build
  tok just test
  tok just --list`,
	DisableFlagParsing: true,
	RunE:               runJust,
}

func init() {
	registry.Add(func() { registry.Register(justCmd) })
}

func runJust(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runJustList()
	}

	// Check for list flag
	for _, arg := range args {
		if arg == "-l" || arg == "--list" {
			return runJustList()
		}
	}

	return runJustRecipe(args)
}

func runJustRecipe(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: just %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("just", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterJustOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "just", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("just", "tok just", originalTokens, filteredTokens)

	return err
}

func filterJustOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Keep errors
		if strings.Contains(trimmed, "error:") || strings.HasPrefix(trimmed, "error:") {
			result = append(result, line)
			continue
		}

		// In ultra-compact mode, be very selective
		if shared.UltraCompact {
			// Keep only error lines
			if strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "FAILED") {
				result = append(result, shared.TruncateLine(line, 100))
			}
			continue
		}

		result = append(result, shared.TruncateLine(line, 120))
	}

	return strings.Join(result, "\n")
}

func runJustList() error {
	timer := tracking.Start()

	execCmd := exec.Command("just", "--list")
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterJustListOutput(raw)

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("just list", "tok just list", originalTokens, filteredTokens)

	return err
}

func filterJustListOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip "Available recipes:" header
		if strings.HasPrefix(trimmed, "Available") {
			if !shared.UltraCompact {
				result = append(result, line)
			}
			continue
		}

		if shared.UltraCompact {
			// Just show recipe names
			if strings.HasPrefix(trimmed, "just") {
				// Skip the "just" prefix in listing
				continue
			}
			// Extract recipe name (first word)
			recipe := strings.Fields(trimmed)[0]
			result = append(result, recipe)
		} else {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	return strings.Join(result, "\n")
}
