package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var duCmd = &cobra.Command{
	Use:   "du [path...]",
	Short: "Disk usage with compact output",
	Long: `Execute du commands with token-optimized output.

Specialized filters for:
  - Large directory highlighting
  - Sorted usage summary

Examples:
  tok du -sh *
  tok du -h --max-depth=1
  tok du -k /var`,
	DisableFlagParsing: true,
	RunE:               runDu,
}

func init() {
	registry.Add(func() { registry.Register(duCmd) })
}

func runDu(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: du %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("du", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterDuOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "du", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("du", "tok du", originalTokens, filteredTokens)

	return err
}

func filterDuOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	// In ultra-compact mode, show only top entries by size
	if shared.UltraCompact && len(lines) > 20 {
		// Sort by size (du output has size first)
		// Just show top 15 entries
		count := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			if count < 15 {
				// Compact the line
				fields := strings.Fields(trimmed)
				if len(fields) >= 2 {
					size := fields[0]
					path := strings.Join(fields[1:], " ")
					path = shared.TruncateLine(path, 50)
					result = append(result, fmt.Sprintf("%8s  %s", size, path))
				}
			}
			count++
		}

		if len(lines) > 15 {
			result = append(result, fmt.Sprintf("... (%d more entries)", len(lines)-15))
		}
		return strings.Join(result, "\n")
	}

	// Normal mode - truncate long paths
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Truncate long paths
		result = append(result, shared.TruncateLine(line, 100))
	}

	return strings.Join(result, "\n")
}
