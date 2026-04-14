package test

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var vitestCmd = &cobra.Command{
	Use:   "vitest [args...]",
	Short: "Vitest with filtered output (90% token reduction)",
	Long: `Execute Vitest with token-optimized output.

Shows only test failures and summary with accurate count extraction.

Examples:
  tokman vitest run
  tokman vitest run --coverage`,
	DisableFlagParsing: true,
	RunE:               runVitest,
}

func init() {
	registry.Add(func() { registry.Register(vitestCmd) })
}

func runVitest(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"run"}
	}

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: vitest %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("vitest", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterVitestOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "vitest", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("vitest %s", strings.Join(args, " ")), "tokman vitest", originalTokens, filteredTokens)

	return err
}

var vitestSummaryRe = regexp.MustCompile(`(\d+)\s+(passed|failed|skipped)`)
var vitestSuiteRe = regexp.MustCompile(`(✓|×|FAIL|PASS)\s+(.+)`)
var vitestFileRe = regexp.MustCompile(`(✓|×|✗)\s+(.+?)(?:\s+\(\d+\s*ms\))`)

func filterVitestOutput(raw string) string {
	var passed, failed, skipped int
	var suitePassed, suiteFailed, suiteTotal int
	var failures []string
	var testFiles []string
	suiteTotal++

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		matches := vitestSummaryRe.FindAllStringSubmatch(trimmed, -1)
		for _, m := range matches {
			count := 0
			fmt.Sscanf(m[1], "%d", &count)
			switch m[2] {
			case "passed":
				passed += count
			case "failed":
				failed += count
			case "skipped":
				skipped += count
			}
		}

		if strings.Contains(trimmed, "Tests") && strings.Contains(trimmed, "passed") {
			suiteTotal++
		}

		if strings.Contains(trimmed, "FAIL") || strings.Contains(trimmed, "✗") || strings.Contains(trimmed, "×") {
			if !strings.Contains(trimmed, "0 failed") {
				failures = append(failures, shared.TruncateLine(trimmed, 80))
			}
		}

		if strings.Contains(trimmed, "✓") || strings.Contains(trimmed, "PASS") {
			suitePassed++
		}
		if strings.Contains(trimmed, "✗") || strings.Contains(trimmed, "×") || strings.Contains(trimmed, "FAIL") {
			if !strings.Contains(trimmed, "0 failed") {
				suiteFailed++
			}
		}

		if strings.Contains(trimmed, "Test Files") || strings.Contains(trimmed, "Tests") {
			testFiles = append(testFiles, trimmed)
		}
	}

	if shared.UltraCompact {
		var parts []string
		parts = append(parts, fmt.Sprintf("P:%d F:%d", passed, failed))
		if skipped > 0 {
			parts = append(parts, fmt.Sprintf("S:%d", skipped))
		}
		if suiteFailed > 0 {
			parts = append(parts, fmt.Sprintf("SuitesF:%d", suiteFailed))
		}
		if len(failures) > 0 {
			for i, f := range failures {
				if i >= 3 {
					parts = append(parts, fmt.Sprintf("+%d more", len(failures)-3))
					break
				}
				parts = append(parts, shared.TruncateLine(f, 50))
			}
		}
		return strings.Join(parts, " ")
	}

	var result []string
	result = append(result, "Vitest Results:")
	if suitePassed > 0 || suiteFailed > 0 {
		result = append(result, fmt.Sprintf("  %d suites passed, %d suites failed", suitePassed, suiteFailed))
	}
	if passed > 0 {
		result = append(result, fmt.Sprintf("  %d passed", passed))
	}
	if failed > 0 {
		result = append(result, fmt.Sprintf("  %d failed", failed))
	}
	if skipped > 0 {
		result = append(result, fmt.Sprintf("  %d skipped", skipped))
	}

	if len(testFiles) > 0 {
		result = append(result, "")
		for _, tf := range testFiles {
			result = append(result, fmt.Sprintf("  %s", tf))
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
			result = append(result, fmt.Sprintf("  %s", f))
		}
	}

	if passed == 0 && failed == 0 && len(result) <= 1 {
		for _, line := range strings.Split(raw, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				result = append(result, shared.TruncateLine(trimmed, 100))
				if len(result) > 20 {
					result = append(result, fmt.Sprintf("  ... (%d more lines)", len(strings.Split(raw, "\n"))-20))
					break
				}
			}
		}
	}

	return strings.Join(result, "\n")
}
