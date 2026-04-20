package system

import (
	"fmt"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var (
	grepMaxLen   int
	grepMax      int
	grepFileType string
	grepGroup    bool
	grepContext  int
)

var grepCmd = &cobra.Command{
	Use:   "grep [args...]",
	Short: "Compact grep - strips whitespace, truncates, groups by file",
	Long: `Compact grep with token-optimized output.

Strips whitespace, truncates long lines, and groups results by file.
Passes native grep/ripgrep flags through.

Examples:
  tok grep -r "TODO" .
  tok grep "func " . -t go
  tok grep -r "error" . --max-len 60 --max 20 --group
  tok grep -C 2 "func main" main.go`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE:               runGrep,
}

func init() {
	registry.Add(func() { registry.Register(grepCmd) })
	grepCmd.Flags().IntVarP(&grepMaxLen, "max-len", "l", 80, "Max line length")
	grepCmd.Flags().IntVarP(&grepMax, "max", "m", 50, "Max results to show")
	grepCmd.Flags().StringVarP(&grepFileType, "type", "t", "", "Filter by file type (go, py, js, rust, ts, java)")
	grepCmd.Flags().BoolVarP(&grepGroup, "group", "g", true, "Group results by file")
	grepCmd.Flags().IntVarP(&grepContext, "context", "C", 0, "Lines of context around matches")
}

var fileTypeExtensions = map[string][]string{
	"go":   {".go"},
	"py":   {".py", ".pyi", ".pyx"},
	"js":   {".js", ".jsx", ".mjs", ".cjs"},
	"ts":   {".ts", ".tsx", ".mts", ".cts"},
	"rs":   {".rs"},
	"java": {".java"},
	"rb":   {".rb"},
	"cpp":  {".cpp", ".cc", ".cxx", ".hpp"},
	"c":    {".c", ".h"},
	"css":  {".css", ".scss", ".less", ".sass"},
	"html": {".html", ".htm", ".tmpl"},
	"json": {".json"},
	"yaml": {".yaml", ".yml"},
	"toml": {".toml"},
	"md":   {".md", ".mdx"},
	"sh":   {".sh", ".bash", ".zsh"},
}

func runGrep(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	grepArgs := append([]string{}, args...)
	grepArgs = append([]string{"--color=never"}, grepArgs...)

	if grepContext > 0 {
		grepArgs = append([]string{"-C", fmt.Sprintf("%d", grepContext)}, grepArgs...)
	}

	if grepFileType != "" {
		if exts, ok := fileTypeExtensions[grepFileType]; ok {
			includeArgs := []string{}
			for _, ext := range exts {
				includeArgs = append(includeArgs, "--include=*"+ext)
			}
			grepArgs = append(includeArgs, grepArgs...)
		}
	}

	output, exitCode, err := shared.RunAndCapture("grep", grepArgs)

	if err != nil && exitCode == 1 && output == "" {
		out.Global().Println("(no matches)")
		return nil
	}

	var filtered string
	if grepGroup {
		filtered = compactGrepOutputGrouped(output, grepMaxLen, grepMax)
	} else {
		filtered = compactGrepOutputSimple(output, grepMaxLen, grepMax)
	}

	if err != nil && exitCode != 1 {
		if hint := shared.TeeOnFailure(output, "grep", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("grep %s", strings.Join(args, " ")), "tok grep", originalTokens, filteredTokens)

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
		if len(line) > maxLen {
			line = line[:maxLen] + "..."
		}
		result.WriteString(line + "\n")
		count++
	}

	return result.String()
}

func compactGrepOutputGrouped(output string, maxLen, maxResults int) string {
	lines := strings.Split(output, "\n")

	type fileGroup struct {
		filename string
		matches  []string
	}
	groups := []fileGroup{}
	currentFile := ""
	currentMatches := []string{}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		idx := strings.Index(line, ":")
		if idx > 0 {
			filename := line[:idx]
			matchLine := line[idx+1:]

			if filename != currentFile {
				if currentFile != "" && len(currentMatches) > 0 {
					groups = append(groups, fileGroup{filename: currentFile, matches: currentMatches})
				}
				currentFile = filename
				currentMatches = []string{}
			}

			if len(matchLine) > maxLen {
				matchLine = matchLine[:maxLen] + "..."
			}
			currentMatches = append(currentMatches, matchLine)
		} else {
			if len(line) > maxLen {
				line = line[:maxLen] + "..."
			}
			if currentFile == "" {
				currentFile = "(unknown)"
			}
			currentMatches = append(currentMatches, line)
		}
	}
	if currentFile != "" && len(currentMatches) > 0 {
		groups = append(groups, fileGroup{filename: currentFile, matches: currentMatches})
	}

	if len(groups) == 0 {
		return "(no matches)\n"
	}

	totalMatches := 0
	for _, g := range groups {
		totalMatches += len(g.matches)
	}

	if shared.UltraCompact {
		var fileNames []string
		for i, g := range groups {
			if i >= 5 {
				break
			}
			fileNames = append(fileNames, fmt.Sprintf("%s(%d)", shortFilename(g.filename), len(g.matches)))
		}
		result := fmt.Sprintf("%d matches in %d files: %s", totalMatches, len(groups), strings.Join(fileNames, " "))
		if len(groups) > 5 {
			result += fmt.Sprintf(" +%d", len(groups)-5)
		}
		return result + "\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d matches in %d files\n", totalMatches, len(groups)))

	filesShown := 0
	matchesShown := 0
	for _, g := range groups {
		if filesShown >= 20 || matchesShown >= maxResults {
			remaining := len(groups) - filesShown
			if remaining > 0 {
				result.WriteString(fmt.Sprintf("\n... +%d more files\n", remaining))
			}
			break
		}

		result.WriteString(fmt.Sprintf("\n%s (%d):\n", g.filename, len(g.matches)))
		filesShown++

		for _, m := range g.matches {
			if matchesShown >= maxResults {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", totalMatches-matchesShown))
				return result.String()
			}
			result.WriteString(fmt.Sprintf("  %s\n", m))
			matchesShown++
		}
	}

	return result.String()
}

func shortFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
