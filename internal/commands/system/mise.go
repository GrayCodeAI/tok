package system

import (
	"fmt"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var miseCmd = &cobra.Command{
	Use:   "mise [command] [args...]",
	Short: "Mise version manager with compact output",
	Long: `Execute Mise (formerly rtx) commands with token-optimized output.

Specialized filters for:
  - install: Compact installation summary
  - ls: Compact tool listing
  - outdated: Compact outdated tools

Examples:
  tok mise install
  tok mise ls
  tok mise outdated`,
	DisableFlagParsing: true,
	RunE:               runMise,
}

func init() {
	registry.Add(func() { registry.Register(miseCmd) })
}

func runMise(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"help"}
	}

	switch args[0] {
	case "install":
		return runMiseInstall(args[1:])
	case "ls", "list":
		return runMiseLs(args[1:])
	case "outdated":
		return runMiseOutdated(args[1:])
	default:
		return runMisePassthrough(args)
	}
}

func runMiseInstall(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: mise install %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("mise", append([]string{"install"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterMiseInstallOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "mise_install", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("mise install", "tok mise install", originalTokens, filteredTokens)

	return err
}

func filterMiseInstallOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string
	var installed int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Count installations
		if strings.Contains(trimmed, "installed") || strings.Contains(trimmed, "Downloading") {
			installed++
			if shared.UltraCompact {
				continue
			}
			result = append(result, shared.TruncateLine(line, 100))
			continue
		}

		// Keep errors
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "ERROR") {
			result = append(result, line)
			continue
		}

		// Skip verbose download progress
		if shared.UltraCompact {
			continue
		}

		result = append(result, shared.TruncateLine(line, 100))
	}

	if shared.UltraCompact && installed > 0 {
		return fmt.Sprintf("OK Installed %d tools", installed)
	}

	return strings.Join(result, "\n")
}

func runMiseLs(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: mise ls %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("mise", append([]string{"ls"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterMiseLsOutput(raw)

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("mise ls", "tok mise ls", originalTokens, filteredTokens)

	return err
}

func filterMiseLsOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip header lines
		if strings.HasPrefix(trimmed, "Name") || strings.HasPrefix(trimmed, "----") {
			if !shared.UltraCompact {
				result = append(result, line)
			}
			continue
		}

		if shared.UltraCompact {
			// Just show tool name and version
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				result = append(result, fmt.Sprintf("%s@%s", parts[0], parts[1]))
			}
		} else {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	return strings.Join(result, "\n")
}

func runMiseOutdated(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: mise outdated %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("mise", append([]string{"outdated"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterMiseOutdatedOutput(raw)

	if filtered == "" && err == nil {
		out.Global().Println("OK All tools are up to date")
	} else {
		out.Global().Println(filtered)
	}

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("mise outdated", "tok mise outdated", originalTokens, filteredTokens)

	return err
}

func filterMiseOutdatedOutput(raw string) string {
	if raw == "" {
		return ""
	}

	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if shared.UltraCompact {
			// Show tool name and version delta
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				result = append(result, fmt.Sprintf("%s: %s → %s", parts[0], parts[1], parts[2]))
			}
		} else {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	return strings.Join(result, "\n")
}

func runMisePassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: mise %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("mise", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterMiseBasicOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "mise", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("mise", "tok mise", originalTokens, filteredTokens)

	return err
}

func filterMiseBasicOutput(raw string) string {
	if shared.UltraCompact {
		lines := strings.Split(raw, "\n")
		var result []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "ERROR") {
				result = append(result, shared.TruncateLine(line, 100))
			}
		}
		return strings.Join(result, "\n")
	}
	return raw
}
