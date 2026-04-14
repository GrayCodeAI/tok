package system

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var (
	grepMaxLen   int
	grepMax      int
	grepFileType string
)

var grepCmd = &cobra.Command{
	Use:   "grep [args...]",
	Short: "Compact grep - strips whitespace, truncates, groups by file",
	Long: `Compact grep with token-optimized output.

Strips whitespace, truncates long lines, and groups results by file.
Passes native grep/ripgrep flags through.

Examples:
  tokman grep -r "TODO" .
  tokman grep "func " . -t go
  tokman grep -r "error" . --max-len 60 --max 20`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE:               runGrep,
}

func init() {
	registry.Add(func() { registry.Register(grepCmd) })
	grepCmd.Flags().IntVarP(&grepMaxLen, "max-len", "l", 80, "Max line length")
	grepCmd.Flags().IntVarP(&grepMax, "max", "m", 50, "Max results to show")
	grepCmd.Flags().StringVarP(&grepFileType, "type", "t", "", "Filter by file type (go, py, js, rust)")
}

func runGrep(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Use standard grep with all args passed through
	grepArgs := append([]string{}, args...)

	// Add --color=never to avoid ANSI codes
	grepArgs = append([]string{"--color=never"}, grepArgs...)

	output, exitCode, err := shared.RunAndCapture("grep", grepArgs)

	// Grep returns exit code 1 when no matches - that's not an error for us
	if err != nil && exitCode == 1 && output == "" {
		fmt.Println("(no matches)")
		return nil
	}

	// Compact output for minimal tokens
	filtered := compactGrepOutputSimple(output, grepMaxLen, grepMax)

	if err != nil && exitCode != 1 {
		if hint := shared.TeeOnFailure(output, "grep", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("grep %s", strings.Join(args, " ")), "tokman grep", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil && exitCode != 1 {
		return fmt.Errorf("grep failed: %w", err)
	}
	return nil
}

func compactGrepOutputSimple(output string, maxLen, maxResults int) string {
	if shared.UltraCompact {
		lines := strings.Split(output, "\n")
		matchCount := 0
		fileSet := make(map[string]bool)
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			matchCount++
			if idx := strings.Index(line, ":"); idx > 0 {
				fileSet[line[:idx]] = true
			}
		}
		if matchCount == 0 {
			return "0 matches\n"
		}
		return fmt.Sprintf("%d matches in %d files\n", matchCount, len(fileSet))
	}

	var result strings.Builder
	count := 0

	for _, line := range strings.Split(output, "\n") {
		if count >= maxResults {
			result.WriteString(fmt.Sprintf("... (%d more)\n", count-maxResults+1))
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Truncate long lines
		if len(line) > maxLen {
			line = line[:maxLen] + "..."
		}
		result.WriteString(line + "\n")
		count++
	}

	return result.String()
}
