package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/telemetry"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

// genericTestCmd provides a generic test wrapper
// Auto-detects and runs appropriate test runner for the project
var genericTestCmd = &cobra.Command{
	Use:   "test-runner <command> [args...]",
	Short: "Auto-detect and run project tests",
	Long: `Run tests with automatic test runner detection.

This command detects the appropriate test runner for your project
and runs it with optimized output. Supports:
- Rust (cargo test)
- Go (go test)
- Node.js (npm test, pnpm test, vitest, jest)
- Python (pytest)
- Ruby (rspec, rake test)

Examples:
  tok test-runner cargo test      # Run Rust tests
  tok test-runner npm test        # Run npm tests
  tok test-runner                 # Auto-detect and run tests`,
	DisableFlagParsing: true,
	RunE:               runGenericTest,
}

func init() {
	registry.Add(func() { registry.Register(genericTestCmd) })
}

// TestRunner represents a detected test runner
type TestRunner struct {
	Name        string
	Command     string
	Args        []string
	DetectFiles []string
	Priority    int
}

// DetectedRunners returns list of detected test runners for current project
func DetectedRunners() []TestRunner {
	var runners []TestRunner
	cwd, _ := os.Getwd()

	// Define test runners with their detection criteria
	runnerDefs := []TestRunner{
		{
			Name:        "Vitest",
			Command:     "npx",
			Args:        []string{"vitest", "run"},
			DetectFiles: []string{"vitest.config.ts", "vitest.config.js"},
			Priority:    110, // Highest - very specific
		},
		{
			Name:        "Playwright",
			Command:     "npx",
			Args:        []string{"playwright", "test"},
			DetectFiles: []string{"playwright.config.ts", "playwright.config.js"},
			Priority:    105, // Very specific
		},
		{
			Name:        "Cargo",
			Command:     "cargo",
			Args:        []string{"test"},
			DetectFiles: []string{"Cargo.toml"},
			Priority:    100,
		},
		{
			Name:        "Go",
			Command:     "go",
			Args:        []string{"test", "./..."},
			DetectFiles: []string{"go.mod"},
			Priority:    100,
		},
		{
			Name:        "Jest",
			Command:     "npx",
			Args:        []string{"jest"},
			DetectFiles: []string{"jest.config.js", "jest.config.ts"},
			Priority:    80,
		},
		{
			Name:        "npm",
			Command:     "npm",
			Args:        []string{"test"},
			DetectFiles: []string{"package.json"},
			Priority:    70,
		},
		{
			Name:        "pnpm",
			Command:     "pnpm",
			Args:        []string{"test"},
			DetectFiles: []string{"pnpm-lock.yaml"},
			Priority:    75,
		},
		{
			Name:        "Pytest",
			Command:     "pytest",
			Args:        []string{},
			DetectFiles: []string{"pytest.ini", "setup.py", "pyproject.toml"},
			Priority:    100,
		},
		{
			Name:        "Python (unittest)",
			Command:     "python",
			Args:        []string{"-m", "unittest"},
			DetectFiles: []string{"test_*.py", "*_test.py"},
			Priority:    60,
		},
		{
			Name:        "RSpec",
			Command:     "rspec",
			Args:        []string{},
			DetectFiles: []string{"spec", ".rspec"},
			Priority:    100,
		},
		{
			Name:        "Rake Test",
			Command:     "rake",
			Args:        []string{"test"},
			DetectFiles: []string{"Rakefile"},
			Priority:    80,
		},
		{
			Name:        "Playwright",
			Command:     "npx",
			Args:        []string{"playwright", "test"},
			DetectFiles: []string{"playwright.config.ts", "playwright.config.js"},
			Priority:    95,
		},
	}

	for _, runner := range runnerDefs {
		for _, file := range runner.DetectFiles {
			fullPath := filepath.Join(cwd, file)
			if _, err := os.Stat(fullPath); err == nil {
				runners = append(runners, runner)
				break
			}
		}
	}

	// Sort by priority (higher first)
	for i := 0; i < len(runners)-1; i++ {
		for j := i + 1; j < len(runners); j++ {
			if runners[i].Priority < runners[j].Priority {
				runners[i], runners[j] = runners[j], runners[i]
			}
		}
	}

	return runners
}

func runGenericTest(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	var command string
	var commandArgs []string
	var autoDetected bool

	if len(args) > 0 {
		// User specified a test command
		command = args[0]
		commandArgs = args[1:]
	} else {
		// Auto-detect test runner
		runners := DetectedRunners()
		if len(runners) == 0 {
			return fmt.Errorf("no test runner detected in current directory")
		}

		runner := runners[0]
		command = runner.Command
		commandArgs = runner.Args
		autoDetected = true

		// Track telemetry
		telemetry.TrackTestRunnerUsage(runner.Name, true)

		if shared.Verbose > 0 {
			fmt.Fprintf(os.Stderr, "Detected test runner: %s\n", runner.Name)
		}
	}

	// Track telemetry for explicit test commands
	if !autoDetected && len(args) > 0 {
		telemetry.TrackTestRunnerUsage(command, false)
	}

	// Route to appropriate handler based on command
	switch command {
	case "cargo":
		return runCargoTestWithArgs(commandArgs, timer)
	case "go":
		return runGoTestWithArgs(commandArgs, timer)
	case "npm":
		return runNpmTestWithArgs(commandArgs, timer)
	case "pnpm":
		return runPnpmTestWithArgs(commandArgs, timer)
	case "pytest":
		return runPytestWithArgs(commandArgs, timer)
	case "vitest":
		return runVitestWithArgs(commandArgs, timer)
	case "jest":
		return runJestWithArgs(commandArgs, timer)
	case "rspec":
		return runRspecWithArgs(commandArgs, timer)
	case "rake":
		return runRakeWithArgs(commandArgs, timer)
	case "playwright":
		return runPlaywrightWithArgs(commandArgs, timer)
	default:
		// Unknown command - pass through
		return runPassthroughTest(command, commandArgs, timer)
	}
}

func runCargoTestWithArgs(args []string, timer *tracking.TimedExecution) error {
	// Delegate to existing cargo test implementation
	out.Global().Println(color.CyanString("Running Cargo tests..."))

	// This would call the existing cargo test filter
	// For now, pass through to cargo test
	return runPassthroughTest("cargo", append([]string{"test"}, args...), timer)
}

func runGoTestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running Go tests..."))
	return runPassthroughTest("go", append([]string{"test"}, args...), timer)
}

func runNpmTestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running npm tests..."))
	return runPassthroughTest("npm", append([]string{"test"}, args...), timer)
}

func runPnpmTestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running pnpm tests..."))
	return runPassthroughTest("pnpm", append([]string{"test"}, args...), timer)
}

func runPytestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running pytest..."))
	return runPassthroughTest("pytest", args, timer)
}

func runVitestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running Vitest..."))
	return runPassthroughTest("npx", append([]string{"vitest", "run"}, args...), timer)
}

func runJestWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running Jest..."))
	return runPassthroughTest("npx", append([]string{"jest"}, args...), timer)
}

func runRspecWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running RSpec..."))
	return runPassthroughTest("rspec", args, timer)
}

func runRakeWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running Rake tests..."))
	return runPassthroughTest("rake", append([]string{"test"}, args...), timer)
}

func runPlaywrightWithArgs(args []string, timer *tracking.TimedExecution) error {
	out.Global().Println(color.CyanString("Running Playwright..."))
	return runPassthroughTest("npx", append([]string{"playwright", "test"}, args...), timer)
}

func runPassthroughTest(command string, args []string, timer *tracking.TimedExecution) error {
	fullCommand := command + " " + strings.Join(args, " ")

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Executing: %s %s\n", command, strings.Join(args, " "))
	}

	// Execute the command
	execCmd := exec.Command(command, args...)
	outputBytes, err := execCmd.CombinedOutput()
	output := string(outputBytes)

	// Apply generic test filtering
	filtered := filterGenericTestOutput(output)

	if err != nil {
		if hint := shared.TeeOnFailure(output, "test", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	// Track usage
	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fullCommand, "tok test", originalTokens, filteredTokens)

	return err
}

// filterGenericTestOutput applies generic filtering for test output
func filterGenericTestOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result []string

	var failures []string
	var summary string
	inFailure := false
	var currentFailure []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect failure sections (but not "0 failed" or similar success indicators)
		isFailureLine := false
		if strings.Contains(trimmed, "FAIL") && !strings.Contains(trimmed, "0 FAIL") && !strings.Contains(trimmed, "0 failed") {
			isFailureLine = true
		}
		if strings.Contains(trimmed, "failed") && !strings.Contains(trimmed, "0 failed") && !strings.Contains(trimmed, "no failed") {
			isFailureLine = true
		}
		if strings.Contains(trimmed, "error:") || strings.Contains(trimmed, "Error:") {
			isFailureLine = true
		}

		if isFailureLine {
			inFailure = true
			if len(currentFailure) > 0 {
				failures = append(failures, strings.Join(currentFailure, "\n"))
				currentFailure = nil
			}
			currentFailure = append(currentFailure, line)
			continue
		}

		// Detect summary lines
		if strings.Contains(trimmed, "passed") || strings.Contains(trimmed, "failed") ||
			strings.Contains(trimmed, "skipped") || strings.Contains(trimmed, "tests") ||
			strings.Contains(trimmed, "Test Suites:") || strings.Contains(trimmed, "Tests:") {
			summary = line
			inFailure = false
			if len(currentFailure) > 0 {
				failures = append(failures, strings.Join(currentFailure, "\n"))
				currentFailure = nil
			}
			continue
		}

		// Collect failure details
		if inFailure && trimmed != "" {
			currentFailure = append(currentFailure, line)
		}
	}

	// Don't forget the last failure
	if len(currentFailure) > 0 {
		failures = append(failures, strings.Join(currentFailure, "\n"))
	}

	// Format output
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(failures) > 0 {
		result = append(result, bold("\nTest Failures:"))
		result = append(result, strings.Repeat("─", 40))
		for _, failure := range failures {
			result = append(result, red(failure))
			result = append(result, "")
		}
	}

	if summary != "" {
		if len(failures) == 0 {
			result = append(result, green("✓ "+summary))
		} else {
			result = append(result, red("✗ "+summary))
		}
	} else if len(failures) == 0 && len(result) == 0 {
		// No clear summary, show truncated output
		if len(lines) > 20 {
			result = append(result, lines[:20]...)
			result = append(result, fmt.Sprintf("... (%d more lines)", len(lines)-20))
		} else {
			result = lines
		}
	}

	return strings.Join(result, "\n")
}
