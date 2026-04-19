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

func atoi(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		n = 0
	}
	return n
}

var (
	treeDepth      int
	treePruneNoise bool
)

var noiseDirs = []string{
	"node_modules", ".git", ".svn", ".hg",
	"target", "build", "dist", "out", "bin",
	"__pycache__", ".pytest_cache", ".mypy_cache",
	".next", ".nuxt", ".cache",
	"vendor", "vendor/bundle",
	".terraform", ".terragrunt-cache",
	".gradle", ".m2", ".idea", ".vs",
	"coverage", ".nyc_output", "htmlcov",
	".tox", ".nox", ".eggs",
}

var treeCmd = &cobra.Command{
	Use:   "tree [args...]",
	Short: "Directory tree with token-optimized output",
	Long: `Directory tree with token-optimized output.

Automatically prunes noise directories (node_modules, .git, etc.)
and provides extension summary. Supports all native tree flags.

Examples:
  tok tree
  tok tree -L 2
  tok tree --depth 3 --no-noise
  tok tree -a`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE:               runTree,
}

func init() {
	registry.Add(func() { registry.Register(treeCmd) })
	treeCmd.Flags().IntVarP(&treeDepth, "depth", "L", 0, "Max display depth (0=unlimited)")
	treeCmd.Flags().BoolVar(&treePruneNoise, "no-noise", true, "Prune noise directories")
}

func runTree(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	treeArgs := append([]string{}, args...)

	if treePruneNoise {
		pattern := strings.Join(noiseDirs, "|")
		treeArgs = append([]string{"-I", pattern}, treeArgs...)
	}

	if treeDepth > 0 {
		hasDepth := false
		for _, a := range treeArgs {
			if a == "-L" {
				hasDepth = true
				break
			}
		}
		if !hasDepth {
			treeArgs = append([]string{"-L", fmt.Sprintf("%d", treeDepth)}, treeArgs...)
		}
	}

	c := exec.Command("tree", treeArgs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTreeOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "tree", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("tree %s", strings.Join(args, " ")), "tok tree", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("tree failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("tree failed: %w", err)
	}
	return nil
}

func filterTreeOutput(output string) string {
	lines := strings.Split(output, "\n")

	var dirs, files int
	var extensions = make(map[string]int)
	var treeLines []string
	var summaryLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "directories") && strings.Contains(trimmed, "file") {
			summaryLine = trimmed
			dirs = atoi(trimmed)
			parts := strings.Split(trimmed, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if strings.Contains(p, "file") {
					files = atoi(p)
				}
			}
			continue
		}

		if strings.Contains(trimmed, "directories,") {
			summaryLine = trimmed
			continue
		}

		name := extractTreeName(trimmed)
		if name != "" && !strings.HasPrefix(name, ".") {
			ext := getExtension(name)
			if ext != "" {
				extensions[ext]++
			}
		}

		treeLines = append(treeLines, line)
	}

	if shared.UltraCompact {
		result := fmt.Sprintf("%d dirs %d files", dirs, files)
		if len(extensions) > 0 {
			var topExts []string
			for ext, count := range extensions {
				topExts = append(topExts, fmt.Sprintf("%s:%d", ext, count))
			}
			if len(topExts) > 5 {
				topExts = topExts[:5]
			}
			result += " " + strings.Join(topExts, " ")
		}
		return result + "\n"
	}

	var result strings.Builder

	if len(treeLines) > 50 {
		for i, line := range treeLines {
			if i >= 30 {
				result.WriteString(fmt.Sprintf("... (%d more entries)\n", len(treeLines)-30))
				break
			}
			result.WriteString(line + "\n")
		}
	} else {
		for _, line := range treeLines {
			result.WriteString(line + "\n")
		}
	}

	if summaryLine != "" {
		result.WriteString(summaryLine + "\n")
	} else if dirs > 0 || files > 0 {
		result.WriteString(fmt.Sprintf("%d directories, %d files\n", dirs, files))
	}

	if len(extensions) > 0 {
		result.WriteString("\nExtensions:\n")
		sortedExts := sortExtensions(extensions)
		for i, ext := range sortedExts {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(extensions)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  .%-8s %d\n", ext.ext, ext.count))
		}
	}

	return result.String()
}

func extractTreeName(line string) string {
	trimmed := strings.TrimSpace(line)

	trimmed = strings.TrimPrefix(trimmed, "├── ")
	trimmed = strings.TrimPrefix(trimmed, "└── ")
	trimmed = strings.TrimPrefix(trimmed, "│   ")
	trimmed = strings.TrimPrefix(trimmed, "│")

	for i, c := range trimmed {
		if c == ' ' || c == '─' || c == '│' || c == '├' || c == '└' {
			continue
		}
		trimmed = trimmed[i:]
		break
	}

	if idx := strings.Index(trimmed, " "); idx > 0 {
		trimmed = trimmed[:idx]
	}

	return trimmed
}

func getExtension(name string) string {
	if idx := strings.LastIndex(name, "."); idx > 0 {
		return name[idx+1:]
	}
	return ""
}

type extCount struct {
	ext   string
	count int
}

func sortExtensions(extensions map[string]int) []extCount {
	sorted := make([]extCount, 0, len(extensions))
	for ext, count := range extensions {
		sorted = append(sorted, extCount{ext: ext, count: count})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
