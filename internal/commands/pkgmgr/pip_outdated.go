package pkgmgr

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// pipOutdatedCmd represents the pip outdated command
var pipOutdatedCmd = &cobra.Command{
	Use:   "pip-outdated",
	Short: "Show outdated Python packages (condensed format)",
	Long: `Show outdated Python packages in a compact, token-optimized format.

Auto-detects uv and uses it if available for faster results.
Shows package name, current version, and latest version in a condensed format.

Examples:
  tok pip-outdated
  tok pip-outdated --format json`,
	RunE: runPipOutdated,
}

var pipOutdatedFormat string

func init() {
	registry.Add(func() { registry.Register(pipOutdatedCmd) })
	pipOutdatedCmd.Flags().StringVarP(&pipOutdatedFormat, "format", "f", "text", "Output format (text, json, csv)")
}

func runPipOutdated(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Try uv first (faster), fall back to pip
	var output []byte
	var err error

	if _, uvErr := exec.LookPath("uv"); uvErr == nil {
		if shared.Verbose > 0 {
			out.Global().Errorf("Using uv for faster package check")
		}
		output, err = exec.Command("uv", "pip", "list", "--outdated").Output()
	} else {
		output, err = exec.Command("pip", "list", "--outdated", "--format=columns").Output()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("pip list failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("pip list failed: %w", err)
	}

	raw := string(output)

	// Parse and filter
	packages := parsePipOutdated(raw)
	filtered := formatPipOutdated(packages, pipOutdatedFormat)

	out.Global().Println(filtered)

	// Track metrics
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("pip list --outdated", "tok pip-outdated", originalTokens, filteredTokens)

	return nil
}

// PackageInfo holds information about an outdated package
type PackageInfo struct {
	Name    string
	Current string
	Latest  string
	Type    string // wheel, sdist, etc.
}

func parsePipOutdated(raw string) []PackageInfo {
	var packages []PackageInfo
	lines := strings.Split(raw, "\n")

	// Try to detect format and parse
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Package") || strings.HasPrefix(line, "-") {
			continue
		}

		// Try various formats
		// Format 1: uv pip list --outdated
		// Package    Version    Latest     Type
		// django     4.2.0      5.0.0      wheel
		re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s*(\S*)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			pkg := PackageInfo{
				Name:    matches[1],
				Current: matches[2],
				Latest:  matches[3],
			}
			if len(matches) >= 5 {
				pkg.Type = matches[4]
			}
			packages = append(packages, pkg)
		}
	}

	return packages
}

func formatPipOutdated(packages []PackageInfo, format string) string {
	if len(packages) == 0 {
		return "All packages are up to date"
	}

	switch format {
	case "json":
		return formatJSON(packages)
	case "csv":
		return formatCSV(packages)
	default:
		return formatText(packages)
	}
}

func formatText(packages []PackageInfo) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("📦 %d outdated packages\n", len(packages)))
	buf.WriteString("\n")

	for _, pkg := range packages {
		buf.WriteString(fmt.Sprintf("%s: %s → %s\n", pkg.Name, pkg.Current, pkg.Latest))
	}

	return buf.String()
}

func formatJSON(packages []PackageInfo) string {
	// Simple JSON output
	var parts []string
	for _, pkg := range packages {
		parts = append(parts, fmt.Sprintf(`  {"name": "%s", "current": "%s", "latest": "%s"}`,
			pkg.Name, pkg.Current, pkg.Latest))
	}
	return "[\n" + strings.Join(parts, ",\n") + "\n]"
}

func formatCSV(packages []PackageInfo) string {
	var buf bytes.Buffer
	buf.WriteString("name,current,latest\n")
	for _, pkg := range packages {
		buf.WriteString(fmt.Sprintf("%s,%s,%s\n", pkg.Name, pkg.Current, pkg.Latest))
	}
	return buf.String()
}
