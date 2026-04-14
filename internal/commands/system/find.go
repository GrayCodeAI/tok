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

var findCmd = &cobra.Command{
	Use:   "find [args...]",
	Short: "Find files with compact tree output",
	Long: `Find files with token-optimized output.

Accepts native find flags like -name, -type.
Filters and formats output for minimal token usage.

Examples:
  tokman find . -name "*.go"
  tokman find . -type f -mtime -1`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE:               runFind,
}

func init() {
	registry.Add(func() { registry.Register(findCmd) })
}

func runFind(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	findArgs := append([]string{}, args...)
	if len(findArgs) == 0 {
		findArgs = []string{"."}
	}

	output, _, err := shared.RunAndCapture("find", findArgs)

	// Apply filtering - compact output
	engine := filter.NewEngine(filter.ModeMinimal)
	filtered, _ := engine.Process(output)

	// Further compact: one file per line, strip common prefix
	filtered = compactFindOutput(filtered)

	if err != nil {
		if hint := shared.TeeOnFailure(output, "find", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("find %s", strings.Join(args, " ")), "tokman find", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func compactFindOutput(output string) string {
	if shared.UltraCompact {
		lines := strings.Split(output, "\n")
		dirCount := 0
		fileCount := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || trimmed == "." {
				continue
			}
			if strings.HasSuffix(line, "/") || !strings.Contains(line, ".") {
				dirCount++
			} else {
				fileCount++
			}
		}
		if dirCount == 0 && fileCount == 0 {
			return "0 entries\n"
		}
		return fmt.Sprintf("%d dirs, %d files\n", dirCount, fileCount)
	}

	lines := strings.Split(output, "\n")
	var files []string
	var dirs []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip current directory entry
		if line == "." {
			continue
		}
		// Check if directory (ends with / or is a dir path)
		if strings.HasSuffix(line, "/") || !strings.Contains(line, ".") {
			dirs = append(dirs, line)
		} else {
			files = append(files, line)
		}
	}

	// Ultra-compact: show counts first, then inline list with truncation
	var result strings.Builder

	maxShow := 30 // Show first 30 items inline

	if len(dirs) > 0 {
		result.WriteString(fmt.Sprintf("%dD:", len(dirs)))
		shown := 0
		for i, d := range dirs {
			if shown >= maxShow {
				result.WriteString(fmt.Sprintf("\n+%d more", len(dirs)-maxShow))
				break
			}
			if i > 0 && i < maxShow {
				result.WriteString(" ")
			}
			if i < maxShow {
				// Strip leading ./
				d = strings.TrimPrefix(d, "./")
				result.WriteString(d)
				shown++
			}
		}
		result.WriteString("\n")
	}

	if len(files) > 0 {
		result.WriteString(fmt.Sprintf("%dF:", len(files)))
		shown := 0
		for i, f := range files {
			if shown >= maxShow {
				result.WriteString(fmt.Sprintf("\n+%d more", len(files)-maxShow))
				break
			}
			if i > 0 && i < maxShow {
				result.WriteString(" ")
			}
			if i < maxShow {
				// Strip leading ./
				f = strings.TrimPrefix(f, "./")
				result.WriteString(f)
				shown++
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}
