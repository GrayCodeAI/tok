package test

import (
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var playwrightCmd = &cobra.Command{
	Use:   "playwright [args...]",
	Short: "Playwright E2E tests with compact output",
	Long: `Execute Playwright with token-optimized output.

Shows only test results, failures, and per-browser breakdown.

Examples:
  tok playwright test
  tok playwright test --project=chromium
  tok playwright test --reporter=json`,
	DisableFlagParsing: true,
	RunE:               runPlaywright,
}

func init() {
	registry.Add(func() { registry.Register(playwrightCmd) })
}

func runPlaywright(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"test"}
	}

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: playwright %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"playwright"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterPlaywrightOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "playwright", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("playwright %s", strings.Join(args, " ")), "tok playwright", originalTokens, filteredTokens)

	return err
}

func filterPlaywrightOutput(raw string) string {
	var passed, failed, skipped, flaky int
	var failures []string
	var inFailure bool
	var currentFailure []string
	var browsers []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "passed") {
			fields := strings.Fields(trimmed)
			for i, f := range fields {
				if f == "passed" && i > 0 {
					var v int
					if _, err := fmt.Sscanf(fields[i-1], "%d", &v); err == nil {
						passed += v
					}
				}
			}
		}
		if strings.Contains(trimmed, "failed") && !strings.Contains(trimmed, "0 failed") {
			fields := strings.Fields(trimmed)
			for i, f := range fields {
				if f == "failed" && i > 0 {
					var v int
					if _, err := fmt.Sscanf(fields[i-1], "%d", &v); err == nil {
						failed += v
					}
				}
			}
		}
		if strings.Contains(trimmed, "skipped") {
			fields := strings.Fields(trimmed)
			for i, f := range fields {
				if (f == "skipped" || f == "pending") && i > 0 {
					var v int
					if _, err := fmt.Sscanf(fields[i-1], "%d", &v); err == nil {
						skipped += v
					}
				}
			}
		}
		if strings.Contains(trimmed, "flaky") {
			flaky++
		}

		if strings.Contains(trimmed, "workers") || strings.Contains(trimmed, "project") || strings.Contains(trimmed, "browser") {
			browsers = append(browsers, shared.TruncateLine(trimmed, 80))
		}

		if strings.Contains(trimmed, "✘") || strings.Contains(trimmed, "FAIL") || strings.Contains(trimmed, "Error:") {
			if !strings.Contains(trimmed, "0 failed") {
				inFailure = true
				currentFailure = []string{shared.TruncateLine(trimmed, 80)}
			}
		}

		if inFailure {
			if strings.HasPrefix(trimmed, "   at ") || trimmed == "" || strings.Contains(trimmed, "passed") {
				if len(currentFailure) > 0 {
					failures = append(failures, strings.Join(currentFailure, "\n"))
				}
				inFailure = false
				currentFailure = nil
			} else {
				currentFailure = append(currentFailure, shared.TruncateLine(trimmed, 80))
			}
		}
	}

	if inFailure && len(currentFailure) > 0 {
		failures = append(failures, strings.Join(currentFailure, "\n"))
	}

	if shared.UltraCompact {
		var parts []string
		parts = append(parts, fmt.Sprintf("P:%d F:%d", passed, failed))
		if skipped > 0 {
			parts = append(parts, fmt.Sprintf("S:%d", skipped))
		}
		if flaky > 0 {
			parts = append(parts, fmt.Sprintf("Flaky:%d", flaky))
		}
		if len(failures) > 0 {
			for i, f := range failures {
				if i >= 3 {
					parts = append(parts, fmt.Sprintf("+%d", len(failures)-3))
					break
				}
				lines := strings.Split(f, "\n")
				if len(lines) > 0 && len(lines[0]) > 5 {
					parts = append(parts, shared.TruncateLine(lines[0], 60))
				}
			}
		}
		return strings.Join(parts, " ")
	}

	var result []string
	result = append(result, "Playwright Results:")
	if passed > 0 {
		result = append(result, fmt.Sprintf("  %d passed", passed))
	}
	if failed > 0 {
		result = append(result, fmt.Sprintf("  %d failed", failed))
	}
	if skipped > 0 {
		result = append(result, fmt.Sprintf("  %d skipped", skipped))
	}
	if flaky > 0 {
		result = append(result, fmt.Sprintf("  %d flaky", flaky))
	}

	if len(browsers) > 0 {
		result = append(result, "")
		result = append(result, "Browsers:")
		for i, b := range browsers {
			if i >= 5 {
				result = append(result, fmt.Sprintf("  ... +%d more", len(browsers)-5))
				break
			}
			result = append(result, fmt.Sprintf("  %s", b))
		}
	}

	if len(failures) > 0 {
		result = append(result, "")
		result = append(result, "Failures:")
		for i, f := range failures {
			if i >= 10 {
				result = append(result, fmt.Sprintf("  ... +%d more failures", len(failures)-10))
				break
			}
			for _, l := range strings.Split(f, "\n") {
				if len(l) > 3 {
					result = append(result, fmt.Sprintf("  %s", l))
				}
			}
		}
	}

	if len(result) <= 1 {
		return filterTestOutput(raw)
	}

	return strings.Join(result, "\n")
}

func filterTestOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string
	for i, line := range lines {
		if i > 30 {
			result = append(result, fmt.Sprintf("... (%d more lines)", len(lines)-30))
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, shared.TruncateLine(trimmed, 100))
		}
	}
	return strings.Join(result, "\n")
}
