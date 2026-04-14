package pkgmgr

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var pipJSON bool

func formatAsJSONpip(output string) string {
	return fmt.Sprintf(`{"output": %s}`, strconv.Quote(output))
}

var pipCmd = &cobra.Command{
	Use:   "pip [args...]",
	Short: "Pip package manager with compact output",
	Long: `Pip package manager with token-optimized output.

Auto-detects uv if available for faster operations.
Specialized filters for common commands:
  - pip list: Compact package listing
  - pip show: Compact package info
  - pip install: Install summary
  - pip audit: Security audit summary

Examples:
  tokman pip list
  tokman pip install package
  tokman pip show package
  tokman pip outdated`,
	DisableFlagParsing: true,
	RunE:               runPip,
}

func init() {
	registry.Add(func() { registry.Register(pipCmd) })
	pipCmd.Flags().BoolVarP(&pipJSON, "json", "j", false, "Output as JSON")
}

func runPip(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	var c *exec.Cmd
	if _, err := exec.LookPath("uv"); err == nil {
		uvArgs := append([]string{"pip"}, args...)
		c = exec.Command("uv", uvArgs...)
	} else {
		c = exec.Command("pip", args...)
	}

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	if len(args) > 0 {
		switch args[0] {
		case "list":
			filtered = filterPipList(output)
		case "show":
			filtered = filterPipShow(output)
		case "install":
			filtered = filterPipInstall(output)
		case "audit":
			filtered = filterPipAudit(output)
		default:
			filtered = filterPipOutput(output)
		}
	} else {
		filtered = filterPipOutput(output)
	}

	if pipJSON {
		fmt.Println(formatAsJSONpip(output))
		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track(fmt.Sprintf("pip %s", strings.Join(args, " ")), "tokman pip", originalTokens, filteredTokens)
		return err
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("pip %s", strings.Join(args, " ")), "tokman pip", originalTokens, filteredTokens)

	return err
}

func filterPipOutput(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Collecting") && strings.Contains(trimmed, "Downloading") {
			continue
		}
		if strings.HasPrefix(trimmed, "Requirement already satisfied") {
			continue
		}
		if trimmed == "" {
			continue
		}

		result.WriteString(line + "\n")
	}
	return result.String()
}

func filterPipList(output string) string {
	lines := strings.Split(output, "\n")
	var packages []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Package") || strings.HasPrefix(trimmed, "---") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 {
			packages = append(packages, fmt.Sprintf("%s (%s)", fields[0], fields[1]))
		} else if len(fields) == 1 {
			packages = append(packages, fields[0])
		}
	}

	if len(packages) == 0 {
		return output
	}

	if shared.UltraCompact {
		result := fmt.Sprintf("%d packages:", len(packages))
		for i, pkg := range packages {
			if i >= 5 {
				break
			}
			result += " " + pkg
		}
		if len(packages) > 5 {
			result += fmt.Sprintf(" +%d", len(packages)-5)
		}
		return result + "\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d packages:\n", len(packages)))
	for i, pkg := range packages {
		if i >= 30 {
			sb.WriteString(fmt.Sprintf("  ... +%d more\n", len(packages)-30))
			break
		}
		sb.WriteString(fmt.Sprintf("  %s\n", pkg))
	}
	return sb.String()
}

func filterPipShow(output string) string {
	var result strings.Builder
	keyFields := map[string]bool{
		"Name:":        true,
		"Version:":     true,
		"Summary:":     true,
		"Location:":    true,
		"Requires:":    true,
		"Required-by:": true,
		"Home-page:":   true,
		"Author:":      true,
		"License:":     true,
	}

	for _, line := range strings.Split(output, "\n") {
		for key := range keyFields {
			if strings.HasPrefix(line, key) {
				result.WriteString(shared.TruncateLine(line, 100) + "\n")
				break
			}
		}
	}

	if result.Len() == 0 {
		return output
	}
	return result.String()
}

func filterPipInstall(output string) string {
	var added, removed int
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Successfully installed") {
			parts := strings.Split(trimmed, " ")
			for _, p := range parts {
				if strings.Contains(p, "-") {
					added++
				}
			}
		}
		if strings.Contains(trimmed, "Successfully uninstalled") {
			removed++
		}
		if strings.Contains(trimmed, "ERROR") || strings.Contains(trimmed, "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	var result []string
	if added > 0 {
		result = append(result, fmt.Sprintf("Installed %d package(s)", added))
	}
	if removed > 0 {
		result = append(result, fmt.Sprintf("Uninstalled %d package(s)", removed))
	}
	if len(errors) > 0 {
		result = append(result, fmt.Sprintf("Errors (%d):", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result = append(result, fmt.Sprintf("  ... +%d more", len(errors)-5))
				break
			}
			result = append(result, fmt.Sprintf("  %s", e))
		}
	}

	if len(result) == 0 {
		return "Install complete"
	}
	return strings.Join(result, "\n")
}

func filterPipAudit(output string) string {
	var critical, high, moderate, low int
	var vulnerabilities []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "CRITICAL") || strings.Contains(trimmed, "critical") {
			critical++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 100))
		} else if strings.Contains(trimmed, "HIGH") || strings.Contains(trimmed, "high") {
			high++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 100))
		} else if strings.Contains(trimmed, "MODERATE") || strings.Contains(trimmed, "moderate") {
			moderate++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 100))
		} else if strings.Contains(trimmed, "LOW") || strings.Contains(trimmed, "low severity") {
			low++
		}
	}

	total := critical + high + moderate + low
	if total == 0 {
		if strings.Contains(strings.ToLower(strings.Join(strings.Fields(output), " ")), "no known vulnerabilities") {
			return "No known vulnerabilities found"
		}
		return filterPipOutput(output)
	}

	var result []string
	result = append(result, fmt.Sprintf("Audit: %d vulnerabilities", total))
	if critical > 0 {
		result = append(result, fmt.Sprintf("  critical: %d", critical))
	}
	if high > 0 {
		result = append(result, fmt.Sprintf("  high: %d", high))
	}
	if moderate > 0 {
		result = append(result, fmt.Sprintf("  moderate: %d", moderate))
	}
	if low > 0 {
		result = append(result, fmt.Sprintf("  low: %d", low))
	}

	if len(vulnerabilities) > 0 {
		result = append(result, "")
		for i, v := range vulnerabilities {
			if i >= 10 {
				result = append(result, fmt.Sprintf("  ... +%d more", len(vulnerabilities)-10))
				break
			}
			result = append(result, fmt.Sprintf("  %s", v))
		}
	}

	return strings.Join(result, "\n")
}
