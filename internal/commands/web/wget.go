package web

import (
	"fmt"
	"os/exec"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var wgetCmd = &cobra.Command{
	Use:   "wget [flags...] <URL>",
	Short: "wget with compact output",
	Long: `Execute wget commands with compact output formatting.

Shows download progress in a condensed format while preserving
important information like file size and download speed.

Examples:
  tok wget https://example.com/file.zip
  tok wget -O output.txt https://example.com/data`,
	DisableFlagParsing: true,
	RunE:               runWget,
}

func init() {
	registry.Add(func() { registry.Register(wgetCmd) })
}

func runWget(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		return fmt.Errorf("wget requires a URL")
	}

	// Build wget command
	wgetArgs := append([]string{"-q", "--show-progress"}, args...)
	execCmd := exec.Command("wget", wgetArgs...)

	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Filter wget output
	filtered := filterWgetOutput(raw)

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("wget %s", strings.Join(args, " ")), "tok wget", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func filterWgetOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Keep important lines
		if strings.Contains(line, "saved") ||
			strings.Contains(line, "error") ||
			strings.Contains(line, "Error") ||
			strings.Contains(line, "complete") {
			result = append(result, line)
		}
	}

	if len(result) == 0 && output != "" {
		return "✓ Download complete"
	}

	return strings.Join(result, "\n")
}
