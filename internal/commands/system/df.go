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

var dfCmd = &cobra.Command{
	Use:   "df [options]",
	Short: "Disk free space with compact output",
	Long: `Execute df commands with token-optimized output.

Specialized filters for:
  - Compact disk usage table
  - Percentage-based alerts

Examples:
  tok df -h
  tok df -k
  tok df /`,
	DisableFlagParsing: true,
	RunE:               runDf,
}

func init() {
	registry.Add(func() { registry.Register(dfCmd) })
}

func runDf(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: df %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("df", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterDfOutput(raw)

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("df", "tok df", originalTokens, filteredTokens)

	return err
}

func filterDfOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Keep header line
		if i == 0 {
			if shared.UltraCompact {
				// Compact header
				result = append(result, "Filesystem  Size  Used  Avail  Use%  Mounted")
			} else {
				result = append(result, line)
			}
			continue
		}

		if shared.UltraCompact {
			// Parse and compact the output
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				// Extract key fields: filesystem, size, used, avail, use%, mount
				fs := fields[0]
				size := fields[1]
				used := fields[2]
				avail := fields[3]
				usePct := fields[4]
				mount := fields[5]

				// Truncate filesystem name
				if len(fs) > 15 {
					fs = fs[:12] + "..."
				}

				// Truncate mount point
				if len(mount) > 20 {
					mount = mount[:17] + "..."
				}

				result = append(result, fmt.Sprintf("%-15s %5s %5s %5s %4s  %s", fs, size, used, avail, usePct, mount))
			}
		} else {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	return strings.Join(result, "\n")
}
