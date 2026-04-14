package pkgmgr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var pnpmCmd = &cobra.Command{
	Use:   "pnpm [args...]",
	Short: "pnpm with ultra-compact output",
	Long: `Execute pnpm commands with token-optimized output.

Provides compact output for list, outdated, install, and other pnpm commands.

Examples:
  tokman pnpm list
  tokman pnpm list --depth 1
  tokman pnpm outdated
  tokman pnpm install`,
	DisableFlagParsing: true,
	RunE:               runPnpm,
}

func init() {
	registry.Add(func() { registry.Register(pnpmCmd) })
}

func runPnpm(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: pnpm %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("pnpm", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterPnpmOutput(raw, args)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("pnpm %s", strings.Join(args, " ")), "tokman pnpm", originalTokens, filteredTokens)

	if err != nil {
		return err
	}
	return nil
}

func filterPnpmOutput(output string, args []string) string {
	if len(args) == 0 {
		return output
	}

	switch args[0] {
	case "list", "ls":
		return filterPnpmList(output)
	case "outdated":
		return filterPnpmOutdated(output)
	case "install", "add", "update":
		return filterPnpmInstall(output)
	case "typecheck":
		return filterPnpmTypecheck(output)
	case "audit":
		return filterPnpmAudit(output)
	case "run":
		return filterPnpmRun(output)
	case "test":
		return filterPnpmTest(output)
	default:
		return output
	}
}

func filterPnpmList(output string) string {
	lines := strings.Split(output, "\n")
	var packages []string
	var devDeps []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "├──") || strings.HasPrefix(line, "└──") {
			pkg := strings.TrimPrefix(line, "├── ")
			pkg = strings.TrimPrefix(pkg, "└── ")
			pkg = strings.TrimSpace(pkg)
			if pkg != "" && len(pkg) < 60 {
				if strings.Contains(line, "dev:") || strings.Contains(line, "(dev)") {
					devDeps = append(devDeps, pkg)
				} else {
					packages = append(packages, pkg)
				}
			}
		}
	}

	var result []string
	if len(packages) > 0 {
		result = append(result, fmt.Sprintf("📦 Dependencies (%d):", len(packages)))
		for i, pkg := range packages {
			if i >= 15 {
				result = append(result, fmt.Sprintf("   ... +%d more", len(packages)-15))
				break
			}
			result = append(result, fmt.Sprintf("   %s", pkg))
		}
	}

	if len(devDeps) > 0 {
		result = append(result, fmt.Sprintf("📦 DevDependencies (%d):", len(devDeps)))
		for i, pkg := range devDeps {
			if i >= 10 {
				result = append(result, fmt.Sprintf("   ... +%d more", len(devDeps)-10))
				break
			}
			result = append(result, fmt.Sprintf("   %s", pkg))
		}
	}

	if len(result) == 0 {
		return output
	}
	return strings.Join(result, "\n")
}

func filterPnpmOutdated(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Package") || strings.HasPrefix(line, "─") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			pkg := fields[0]
			current := fields[1]
			latest := fields[2]
			if len(fields) >= 4 {
				latest = fields[3]
			}
			result = append(result, fmt.Sprintf("📦 %s: %s → %s", pkg, current, latest))
			count++
		}
	}

	if count == 0 {
		return "✅ All packages up to date"
	}
	return strings.Join(result, "\n")
}

func filterPnpmInstall(output string) string {
	lines := strings.Split(output, "\n")
	var added, removed, changed int
	var warnings []string

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "added") {
			if _, err := fmt.Sscanf(line, "added %d", &added); err != nil {
				added = 0
			}
		}
		if strings.Contains(lower, "removed") {
			if _, err := fmt.Sscanf(line, "removed %d", &removed); err != nil {
				removed = 0
			}
		}
		if strings.Contains(lower, "changed") {
			if _, err := fmt.Sscanf(line, "changed %d", &changed); err != nil {
				changed = 0
			}
		}
		if strings.Contains(lower, "warn") {
			warnings = append(warnings, shared.TruncateLine(line, 80))
		}
	}

	var result []string
	result = append(result, "📦 Install Summary:")
	if added > 0 {
		result = append(result, fmt.Sprintf("   ✅ %d added", added))
	}
	if removed > 0 {
		result = append(result, fmt.Sprintf("   🗑️  %d removed", removed))
	}
	if changed > 0 {
		result = append(result, fmt.Sprintf("   🔄 %d changed", changed))
	}

	if len(warnings) > 0 {
		result = append(result, "   ⚠️  Warnings:")
		for _, w := range warnings {
			if len(w) > 10 {
				result = append(result, fmt.Sprintf("   • %s", w))
			}
		}
	}

	if len(result) == 1 {
		return "✅ Install complete"
	}
	return strings.Join(result, "\n")
}

func filterPnpmTypecheck(output string) string {
	lines := strings.Split(output, "\n")
	var errors []string
	var warnings []string
	var summary string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Capture error lines with file paths
		if strings.Contains(line, "error TS") || strings.Contains(line, ": error ") {
			errors = append(errors, line)
			continue
		}

		// Capture warning lines
		if strings.Contains(line, ": warning ") || strings.Contains(line, "warning TS") {
			warnings = append(warnings, line)
			continue
		}

		// Capture summary line
		if strings.Contains(line, "Found") && (strings.Contains(line, "error") || strings.Contains(line, "warning")) {
			summary = trimmed
		}
	}

	var result []string
	result = append(result, "🔍 TypeScript Type Check:")

	if len(errors) > 0 {
		result = append(result, "")
		result = append(result, fmt.Sprintf("   ❌ %d error(s):", len(errors)))
		for i, err := range errors {
			if i >= 10 {
				result = append(result, fmt.Sprintf("   ... +%d more errors", len(errors)-10))
				break
			}
			result = append(result, fmt.Sprintf("   • %s", shared.TruncateLine(err, 80)))
		}
	}

	if len(warnings) > 0 {
		result = append(result, "")
		result = append(result, fmt.Sprintf("   ⚠️  %d warning(s):", len(warnings)))
		for i, warn := range warnings {
			if i >= 5 {
				result = append(result, fmt.Sprintf("   ... +%d more warnings", len(warnings)-5))
				break
			}
			result = append(result, fmt.Sprintf("   • %s", shared.TruncateLine(warn, 80)))
		}
	}

	if summary != "" {
		result = append(result, "")
		result = append(result, fmt.Sprintf("   📊 %s", summary))
	}

	if len(errors) == 0 && len(warnings) == 0 {
		result = append(result, "   ✅ No type errors found")
	}

	return strings.Join(result, "\n")
}

func filterPnpmAudit(output string) string {
	var critical, high, moderate, low int
	var vulnerabilities []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "critical") {
			critical++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 80))
		} else if strings.Contains(trimmed, "high") && !strings.Contains(trimmed, "higher") {
			high++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 80))
		} else if strings.Contains(trimmed, "moderate") {
			moderate++
			vulnerabilities = append(vulnerabilities, shared.TruncateLine(trimmed, 80))
		} else if strings.Contains(trimmed, "low") {
			low++
		}
	}

	if critical == 0 && high == 0 && moderate == 0 && low == 0 {
		return "No known vulnerabilities found"
	}

	var result []string
	result = append(result, "Audit Results:")
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

func filterPnpmRun(output string) string {
	var result strings.Builder
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "ERR") || strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			errors = append(errors, shared.TruncateLine(trimmed, 120))
			continue
		}
		if strings.HasPrefix(trimmed, ">") || strings.Contains(trimmed, "ready") || strings.Contains(trimmed, "watching") || strings.Contains(trimmed, "compiled") {
			result.WriteString(trimmed + "\n")
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return output
	}
	return result.String()
}

func filterPnpmTest(output string) string {
	var passed, failed, skipped int
	var failures []string
	var summary string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Tests:") || strings.Contains(trimmed, "passed") {
			fields := strings.Fields(trimmed)
			for i, f := range fields {
				if f == "passed" || f == "passing" {
					if i > 0 {
						fmt.Sscanf(fields[i-1], "%d", &passed)
					}
				}
				if f == "failed" || f == "failing" {
					if i > 0 {
						fmt.Sscanf(fields[i-1], "%d", &failed)
					}
				}
				if f == "skipped" || f == "pending" {
					if i > 0 {
						fmt.Sscanf(fields[i-1], "%d", &skipped)
					}
				}
			}
		}

		if strings.Contains(trimmed, "FAIL") || strings.Contains(trimmed, "Error:") {
			failures = append(failures, shared.TruncateLine(trimmed, 80))
		}

		if strings.Contains(trimmed, "Test Files") || strings.Contains(trimmed, "Tests") {
			summary = trimmed
		}
	}

	var result []string
	result = append(result, "Test Results:")
	if passed > 0 {
		result = append(result, fmt.Sprintf("  %d passed", passed))
	}
	if failed > 0 {
		result = append(result, fmt.Sprintf("  %d failed", failed))
	}
	if skipped > 0 {
		result = append(result, fmt.Sprintf("  %d skipped", skipped))
	}
	if summary != "" {
		result = append(result, fmt.Sprintf("  %s", summary))
	}

	if len(failures) > 0 {
		result = append(result, "")
		result = append(result, "Failures:")
		for i, f := range failures {
			if i >= 10 {
				result = append(result, fmt.Sprintf("  ... +%d more", len(failures)-10))
				break
			}
			result = append(result, fmt.Sprintf("  %s", f))
		}
	}

	if passed == 0 && failed == 0 && len(result) <= 1 {
		return output
	}

	return strings.Join(result, "\n")
}
