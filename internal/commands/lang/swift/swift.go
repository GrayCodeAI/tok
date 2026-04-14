package swift

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func init() {
	registry.Add(func() {
		registry.Register(swiftCmd)
	})
}

var swiftCmd = &cobra.Command{
	Use:   "swift [args...]",
	Short: "Swift/Xcodebuild with compact output",
	Long:  `Swift toolchain with token-optimized output.`,
}

func runSwift(args []string) error {
	if len(args) == 0 {
		return swiftCmd.Help()
	}

	timer := tracking.Start()

	cmd := exec.Command("swift", args...)
	output, err := cmd.CombinedOutput()
	raw := string(output)

	filtered := filterSwift(raw)
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)

	timer.Track("swift", "tokman swift", originalTokens, filteredTokens)

	if shared.IsUltraCompact() {
		filtered = compactOutput(filtered)
	}

	fmt.Print(filtered)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "swift", err); hint != "" {
			fmt.Print("\n" + hint)
		}
	}

	return err
}

func filterSwift(raw string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "error.") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "warning:") {
			filtered = append(filtered, line)
		} else if strings.HasPrefix(line, "Compiling") || strings.HasPrefix(line, "Linking") || strings.HasPrefix(line, "Build complete") {
			filtered = append(filtered, line)
		} else if !shared.IsUltraCompact() && line != "" {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func compactOutput(filtered string) string {
	lines := strings.Split(filtered, "\n")
	if len(lines) > 8 {
		return strings.Join(lines[:8], "\n") + fmt.Sprintf("\n... (+%d lines)", len(lines)-8)
	}
	return filtered
}
