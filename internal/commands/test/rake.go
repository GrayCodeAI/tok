package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var rakeCmd = &cobra.Command{
	Use:   "rake [args...]",
	Short: "Rake/Rails test runner with filtered output",
	Long: `Minitest output filter for rake test and rails test.

Parses standard Minitest output, filtering down to failures/errors
and the summary line. Auto-detects rails test for positional files.

Examples:
  tokman rake test
  tokman rake test TEST=test/models/user_test.rb
  tokman rake test test/models/user_test.rb:15`,
	DisableFlagParsing: true,
	RunE:               runRake,
}

func init() {
	registry.Add(func() { registry.Register(rakeCmd) })
}

var reAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m`)
var reFailureHeader = regexp.MustCompile(`^\d+\)\s+(Failure|Error):$`)

func runRake(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	tool, effectiveArgs := selectRunner(args)

	c := rakeRubyExec(tool)
	c.Args = append(c.Args, effectiveArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterMinitestOutput(output)

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("rake %s", strings.Join(args, " ")), "tokman rake", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// selectRunner decides whether to use rake or rails based on args.
func selectRunner(args []string) (string, []string) {
	hasTestSubcommand := len(args) > 0 && args[0] == "test"
	if !hasTestSubcommand {
		return "rake", args
	}

	afterTest := args[1:]
	needsRails := false

	for _, a := range afterTest {
		if strings.Contains(a, "=") || strings.HasPrefix(a, "-") {
			continue
		}
		if looksLikeTestPath(a) {
			needsRails = true
			break
		}
	}

	if needsRails {
		return "rails", args
	}
	return "rake", args
}

func looksLikeTestPath(arg string) bool {
	path := arg
	if idx := strings.Index(arg, ":"); idx >= 0 {
		path = arg[:idx]
	}
	return strings.HasSuffix(path, ".rb") ||
		strings.HasPrefix(path, "test/") ||
		strings.HasPrefix(path, "spec/") ||
		strings.Contains(path, "_test.rb") ||
		strings.Contains(path, "_spec.rb")
}

func rakeRubyExec(tool string) *exec.Cmd {
	if _, err := os.Stat("Gemfile"); err == nil {
		if bundlePath, err := exec.LookPath("bundle"); err == nil {
			return exec.Command(bundlePath, "exec", tool)
		}
	}
	return exec.Command(tool)
}

type minitestParseState int

const (
	minitestHeader minitestParseState = iota
	minitestRunning
	minitestFailures
	minitestSummaryState
)

func filterMinitestOutput(output string) string {
	clean := reAnsi.ReplaceAllString(output, "")
	state := minitestHeader
	var failures []string
	var currentFailure []string
	summaryLine := ""

	for _, line := range strings.Split(clean, "\n") {
		trimmed := strings.TrimSpace(line)

		// Detect summary line anywhere
		if (strings.Contains(trimmed, " runs,") || strings.Contains(trimmed, " tests,")) &&
			strings.Contains(trimmed, " assertions,") {
			summaryLine = trimmed
			continue
		}

		// State transitions
		if trimmed == "# Running:" || strings.HasPrefix(trimmed, "Started with run options") {
			state = minitestRunning
			continue
		}
		if strings.HasPrefix(trimmed, "Finished in ") {
			state = minitestFailures
			continue
		}

		switch state {
		case minitestHeader, minitestRunning:
			continue
		case minitestFailures:
			if reFailureHeader.MatchString(trimmed) {
				if len(currentFailure) > 0 {
					failures = append(failures, strings.Join(currentFailure, "\n"))
					currentFailure = nil
				}
				currentFailure = append(currentFailure, trimmed)
			} else if trimmed == "" && len(currentFailure) > 0 {
				failures = append(failures, strings.Join(currentFailure, "\n"))
				currentFailure = nil
			} else if trimmed != "" {
				currentFailure = append(currentFailure, line)
			}
		case minitestSummaryState:
			// done
		}
	}

	// Save last failure
	if len(currentFailure) > 0 {
		failures = append(failures, strings.Join(currentFailure, "\n"))
	}

	return buildMinitestSummary(summaryLine, failures)
}

func buildMinitestSummary(summary string, failures []string) string {
	runs, _, failCount, errorCount, skips := parseMinitestSummary(summary)

	if runs == 0 && summary == "" {
		return "rake test: no tests ran\n"
	}

	if failCount == 0 && errorCount == 0 {
		msg := fmt.Sprintf("ok rake test: %d runs, 0 failures", runs)
		if skips > 0 {
			msg += fmt.Sprintf(", %d skips", skips)
		}
		return msg + "\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("rake test: %d runs, %d failures, %d errors", runs, failCount, errorCount))
	if skips > 0 {
		result.WriteString(fmt.Sprintf(", %d skips", skips))
	}
	result.WriteString("\n")

	if len(failures) == 0 {
		return result.String()
	}

	result.WriteString("\n")

	for i, failure := range failures {
		if i >= 10 {
			break
		}
		lines := strings.Split(failure, "\n")
		if len(lines) > 0 {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(lines[0])))
		}
		for j := 1; j < len(lines) && j <= 4; j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				continue
			}
			if len(trimmed) > 120 {
				trimmed = trimmed[:117] + "..."
			}
			result.WriteString(fmt.Sprintf("   %s\n", trimmed))
		}
		if i < min(len(failures), 10)-1 {
			result.WriteString("\n")
		}
	}

	if len(failures) > 10 {
		result.WriteString(fmt.Sprintf("\n... +%d more failures\n", len(failures)-10))
	}

	return result.String()
}

func parseMinitestSummary(summary string) (runs, assertions, failures, errors, skips int) {
	for _, part := range strings.Split(summary, ",") {
		part = strings.TrimSpace(part)
		words := strings.Fields(part)
		if len(words) >= 2 {
			n, err := strconv.Atoi(words[0])
			if err != nil {
				continue
			}
			keyword := strings.TrimSuffix(words[1], ",")
			switch keyword {
			case "runs", "run", "tests", "test":
				runs = n
			case "assertions", "assertion":
				assertions = n
			case "failures", "failure":
				failures = n
			case "errors", "error":
				errors = n
			case "skips", "skip":
				skips = n
			}
		}
	}
	return
}
