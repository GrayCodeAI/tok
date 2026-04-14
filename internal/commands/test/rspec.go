package test

import (
	"bytes"
	"encoding/json"
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

var rspecCmd = &cobra.Command{
	Use:   "rspec [args...]",
	Short: "RSpec test runner with filtered output",
	Long: `RSpec test runner with token-optimized output.

Injects --format json for structured output, shows only failures.
Falls back to text parsing when JSON is unavailable.

Examples:
  tokman rspec
  tokman rspec spec/models/
  tokman rspec spec/models/user_spec.rb:15`,
	DisableFlagParsing: true,
	RunE:               runRspec,
}

func init() {
	registry.Add(func() { registry.Register(rspecCmd) })
}

// Noise-stripping patterns
var (
	reSpring      = regexp.MustCompile(`(?i)running via spring preloader`)
	reSimplecov   = regexp.MustCompile(`(?i)(coverage report|simplecov|coverage/|\.simplecov|All Files.*Lines)`)
	reDeprecation = regexp.MustCompile(`^DEPRECATION WARNING:`)
	reFinishedIn  = regexp.MustCompile(`^Finished in \d`)
	reScreenshot  = regexp.MustCompile(`saved screenshot to (.+)`)
	reRspecSumm   = regexp.MustCompile(`(\d+) examples?, (\d+) failures?`)
)

// JSON structures for RSpec --format json output

type RspecOutput struct {
	Examples []RspecExample `json:"examples"`
	Summary  RspecSummary   `json:"summary"`
}

type RspecExample struct {
	FullDescription string          `json:"full_description"`
	Status          string          `json:"status"`
	FilePath        string          `json:"file_path"`
	LineNumber      int             `json:"line_number"`
	Exception       *RspecException `json:"exception"`
}

type RspecException struct {
	Class     string   `json:"class"`
	Message   string   `json:"message"`
	Backtrace []string `json:"backtrace"`
}

type RspecSummary struct {
	Duration                     float64 `json:"duration"`
	ExampleCount                 int     `json:"example_count"`
	FailureCount                 int     `json:"failure_count"`
	PendingCount                 int     `json:"pending_count"`
	ErrorsOutsideOfExamplesCount int     `json:"errors_outside_of_examples_count"`
}

func runRspec(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Detect if user already specified a format
	hasFormat := false
	for _, a := range args {
		if a == "--format" || a == "-f" || strings.HasPrefix(a, "--format=") {
			hasFormat = true
			break
		}
		if strings.HasPrefix(a, "-f") && len(a) > 2 && !strings.HasPrefix(a, "--") {
			hasFormat = true
			break
		}
	}

	c := rspecRubyExec("rspec")
	if !hasFormat {
		c.Args = append(c.Args, "--format", "json")
	}
	c.Args = append(c.Args, args...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	if stdout.String() == "" && err != nil {
		filtered = "RSpec: FAILED (no stdout, see stderr below)\n"
		if stderr.String() != "" {
			filtered += stderr.String()
		}
	} else if hasFormat {
		stripped := stripRspecNoise(stdout.String())
		filtered = filterRspecText(stripped)
	} else {
		filtered = filterRspecOutput(stdout.String())
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("rspec %s", strings.Join(args, " ")), "tokman rspec", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// rspecRubyExec returns a command with bundle exec if Gemfile exists.
func rspecRubyExec(tool string) *exec.Cmd {
	if _, err := os.Stat("Gemfile"); err == nil {
		if bundlePath, err := exec.LookPath("bundle"); err == nil {
			return exec.Command(bundlePath, "exec", tool)
		}
	}
	return exec.Command(tool)
}

// stripRspecNoise removes noise lines: Spring, SimpleCov, DEPRECATION, etc.
func stripRspecNoise(output string) string {
	var result []string
	inSimplecovBlock := false

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if reSpring.MatchString(trimmed) {
			continue
		}
		if reDeprecation.MatchString(trimmed) {
			continue
		}
		if reFinishedIn.MatchString(trimmed) {
			continue
		}
		if reSimplecov.MatchString(trimmed) {
			inSimplecovBlock = true
			continue
		}
		if inSimplecovBlock {
			if trimmed == "" {
				inSimplecovBlock = false
			}
			continue
		}
		if m := reScreenshot.FindStringSubmatch(trimmed); len(m) > 1 {
			result = append(result, fmt.Sprintf("[screenshot: %s]", strings.TrimSpace(m[1])))
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func filterRspecOutput(output string) string {
	if strings.TrimSpace(output) == "" {
		return "RSpec: No output\n"
	}

	// Try parsing as JSON
	var rspec RspecOutput
	if err := json.Unmarshal([]byte(output), &rspec); err == nil {
		return buildRspecSummary(&rspec)
	}

	// Strip noise and retry JSON
	stripped := stripRspecNoise(output)
	if err := json.Unmarshal([]byte(stripped), &rspec); err == nil {
		return buildRspecSummary(&rspec)
	}

	return filterRspecText(stripped)
}

func buildRspecSummary(rspec *RspecOutput) string {
	s := rspec.Summary

	if s.ExampleCount == 0 && s.ErrorsOutsideOfExamplesCount == 0 {
		return "RSpec: No examples found\n"
	}

	if s.ExampleCount == 0 && s.ErrorsOutsideOfExamplesCount > 0 {
		return fmt.Sprintf("RSpec: %d errors outside of examples (%.2fs)\n", s.ErrorsOutsideOfExamplesCount, s.Duration)
	}

	if s.FailureCount == 0 && s.ErrorsOutsideOfExamplesCount == 0 {
		passed := s.ExampleCount - s.PendingCount
		result := fmt.Sprintf("✓ RSpec: %d passed", passed)
		if s.PendingCount > 0 {
			result += fmt.Sprintf(", %d pending", s.PendingCount)
		}
		result += fmt.Sprintf(" (%.2fs)\n", s.Duration)
		return result
	}

	passed := s.ExampleCount - s.FailureCount - s.PendingCount
	if passed < 0 {
		passed = 0
	}
	var result strings.Builder
	result.WriteString(fmt.Sprintf("RSpec: %d passed, %d failed", passed, s.FailureCount))
	if s.PendingCount > 0 {
		result.WriteString(fmt.Sprintf(", %d pending", s.PendingCount))
	}
	result.WriteString(fmt.Sprintf(" (%.2fs)\n", s.Duration))
	result.WriteString("═══════════════════════════════════════\n")

	var failures []RspecExample
	for _, ex := range rspec.Examples {
		if ex.Status == "failed" {
			failures = append(failures, ex)
		}
	}

	if len(failures) == 0 {
		return result.String()
	}

	result.WriteString("\nFailures:\n")

	for i, example := range failures {
		if i >= 5 {
			break
		}
		result.WriteString(fmt.Sprintf("%d. FAIL %s\n   %s:%d\n", i+1, example.FullDescription, example.FilePath, example.LineNumber))

		if example.Exception != nil {
			exc := example.Exception
			shortClass := exc.Class
			if idx := strings.LastIndex(shortClass, "::"); idx >= 0 {
				shortClass = shortClass[idx+2:]
			}
			firstMsg := exc.Message
			if idx := strings.IndexByte(firstMsg, '\n'); idx >= 0 {
				firstMsg = firstMsg[:idx]
			}
			if len(firstMsg) > 120 {
				firstMsg = firstMsg[:117] + "..."
			}
			result.WriteString(fmt.Sprintf("   %s: %s\n", shortClass, firstMsg))

			// First backtrace line not from gems/rspec internals
			for _, bt := range exc.Backtrace {
				if !strings.Contains(bt, "/gems/") && !strings.Contains(bt, "lib/rspec") {
					btLine := bt
					if len(btLine) > 120 {
						btLine = btLine[:117] + "..."
					}
					result.WriteString(fmt.Sprintf("   %s\n", btLine))
					break
				}
			}
		}

		if i < min(len(failures), 5)-1 {
			result.WriteString("\n")
		}
	}

	if len(failures) > 5 {
		result.WriteString(fmt.Sprintf("\n... +%d more failures\n", len(failures)-5))
	}

	return result.String()
}

// filterRspecText is a state-machine text fallback parser.
func filterRspecText(output string) string {
	type rspecState int
	const (
		stateHeader rspecState = iota
		stateFailures
		stateFailedExamples
		stateSummary
	)

	state := stateHeader
	var failures []string
	var currentFailure strings.Builder
	var summaryLine string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		switch state {
		case stateHeader:
			if trimmed == "Failures:" {
				state = stateFailures
			} else if trimmed == "Failed examples:" {
				state = stateFailedExamples
			} else if reRspecSumm.MatchString(trimmed) {
				summaryLine = trimmed
				state = stateSummary
			}
		case stateFailures:
			if isNumberedFailure(trimmed) {
				if currentFailure.Len() > 0 {
					failures = append(failures, compactFailureBlock(currentFailure.String()))
					currentFailure.Reset()
				}
				currentFailure.WriteString(trimmed)
				currentFailure.WriteString("\n")
			} else if trimmed == "Failed examples:" {
				if currentFailure.Len() > 0 {
					failures = append(failures, compactFailureBlock(currentFailure.String()))
					currentFailure.Reset()
				}
				state = stateFailedExamples
			} else if reRspecSumm.MatchString(trimmed) {
				if currentFailure.Len() > 0 {
					failures = append(failures, compactFailureBlock(currentFailure.String()))
					currentFailure.Reset()
				}
				summaryLine = trimmed
				state = stateSummary
			} else if trimmed != "" {
				// Skip gem-internal backtrace lines
				if isGemBacktrace(trimmed) {
					continue
				}
				currentFailure.WriteString(trimmed)
				currentFailure.WriteString("\n")
			}
		case stateFailedExamples:
			if reRspecSumm.MatchString(trimmed) {
				summaryLine = trimmed
				state = stateSummary
			}
		case stateSummary:
			// done
		}
	}

	// Capture remaining failure
	if currentFailure.Len() > 0 && state == stateFailures {
		failures = append(failures, compactFailureBlock(currentFailure.String()))
	}

	if summaryLine != "" {
		if len(failures) == 0 {
			return fmt.Sprintf("RSpec: %s\n", summaryLine)
		}
		var result strings.Builder
		result.WriteString(fmt.Sprintf("RSpec: %s\n", summaryLine))
		result.WriteString("═══════════════════════════════════════\n\n")

		for i, failure := range failures {
			if i >= 5 {
				break
			}
			result.WriteString(fmt.Sprintf("%d. FAIL %s\n", i+1, failure))
			if i < min(len(failures), 5)-1 {
				result.WriteString("\n")
			}
		}
		if len(failures) > 5 {
			result.WriteString(fmt.Sprintf("\n... +%d more failures\n", len(failures)-5))
		}
		return result.String()
	}

	// Fallback: look for summary anywhere
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if strings.Contains(t, "example") && (strings.Contains(t, "failure") || strings.Contains(t, "pending")) {
			return fmt.Sprintf("RSpec: %s\n", t)
		}
	}

	// Last resort: last 5 lines
	return rspecFallbackTail(output, 5)
}

func isNumberedFailure(line string) bool {
	trimmed := strings.TrimSpace(line)
	idx := strings.Index(trimmed, ")")
	if idx <= 0 {
		return false
	}
	prefix := trimmed[:idx]
	for _, ch := range prefix {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isGemBacktrace(line string) bool {
	return strings.Contains(line, "/gems/") ||
		strings.Contains(line, "lib/rspec") ||
		strings.Contains(line, "lib/ruby/") ||
		strings.Contains(line, "vendor/bundle")
}

func compactFailureBlock(block string) string {
	var lines []string
	for _, l := range strings.Split(block, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}

	var specFile string
	var keptLines []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ./spec/") || strings.HasPrefix(t, "# ./test/") {
			specFile = strings.TrimPrefix(t, "# ")
		} else if strings.HasPrefix(t, "#") && (strings.Contains(t, "/gems/") || strings.Contains(t, "lib/rspec")) {
			continue
		} else {
			keptLines = append(keptLines, t)
		}
	}

	result := strings.Join(keptLines, "\n   ")
	if specFile != "" {
		result += "\n   " + specFile
	}
	return result
}

func rspecFallbackTail(output string, n int) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= n {
		return strings.TrimSpace(output) + "\n"
	}
	return strings.Join(lines[len(lines)-n:], "\n") + "\n"
}
