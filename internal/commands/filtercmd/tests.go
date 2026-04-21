package filtercmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/toml"
)

var (
	testVerbose bool
	testFilter  string
)

var testsCmd = &cobra.Command{
	Use:   "tests [filter-name]",
	Short: "Run inline tests for TOML filters",
	Long: `Run inline tests defined in TOML filter files.

Tests are defined in the filter file using [[tests.filtername]] sections:

Example filter with tests (git.toml):

  [git_status]
  match_command = "^git status"
  strip_lines_matching = ["^On branch"]
  
  [[tests.git_status]]
  name = "strips branch line"
  input = """
  On branch main
  nothing to commit
  """
  expected = "nothing to commit"
  
  [[tests.git_status]]
  name = "preserves status messages"
  input = "nothing to commit, working tree clean"
  expected = "nothing to commit, working tree clean"

Usage:
  tok filter tests              # Run all tests in all filters
  tok filter tests git_status   # Run tests for specific filter
  tok filter tests -v           # Verbose output`,
	RunE: runFilterTests,
}

func init() {
	testsCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "verbose output")
	testsCmd.Flags().StringVarP(&testFilter, "filter", "f", "", "test specific filter only")
	registry.Add(func() { registry.Register(testsCmd) })
}

func runFilterTests(cmd *cobra.Command, args []string) error {
	// Determine which filter to test
	filterName := ""
	if len(args) > 0 {
		filterName = args[0]
	} else if testFilter != "" {
		filterName = testFilter
	}

	// Find all TOML filter files
	filterFiles, err := findFilterFiles()
	if err != nil {
		return fmt.Errorf("failed to find filter files: %w", err)
	}

	if len(filterFiles) == 0 {
		out.Global().Println("No filter files found")
		return nil
	}

	// Parse all filters and collect tests
	allTests := toml.NewFilterTestSuite()
	filterRules := make(map[string]toml.TOMLFilterRule)

	for _, file := range filterFiles {
		filter, err := parseFilterFile(file)
		if err != nil {
			if testVerbose {
				out.Global().Errorf("Warning: failed to parse %s: %v\n", file, err)
			}
			continue
		}

		// Collect filter rules
		for name, rule := range filter.Filters {
			filterRules[name] = rule
		}

		// Parse tests from raw content
		tests, err := toml.ParseTests(filter.RawContent)
		if err != nil {
			if testVerbose {
				out.Global().Errorf("Warning: failed to parse tests from %s: %v\n", file, err)
			}
			continue // Skip this file
		}

		// Merge tests into suite
		for name, testList := range tests.Tests {
			for _, test := range testList {
				allTests.AddTest(name, test)
			}
		}
	}

	// Filter tests if specific filter requested
	if filterName != "" {
		tests := allTests.GetTests(filterName)
		if len(tests) == 0 {
			return fmt.Errorf("no tests found for filter '%s'", filterName)
		}

		// Create new suite with only requested filter
		filteredSuite := toml.NewFilterTestSuite()
		for _, test := range tests {
			filteredSuite.AddTest(filterName, test)
		}
		allTests = filteredSuite
	}

	totalTests := allTests.TotalTests()
	if totalTests == 0 {
		out.Global().Println("No tests found in filter files")
		out.Global().Println("\nTo add tests, edit your filter TOML file and add [[tests.filtername]] sections.")
		return nil
	}

	// Print header
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	out.Global().Printf("%s\n", cyan("Running filter tests..."))
	out.Global().Printf("Found %d test(s)\n\n", totalTests)

	// Create filter functions for testing
	filterFuncs := make(map[string]func(string) string)
	for name, rule := range filterRules {
		rule := rule // Capture loop variable
		filterFuncs[name] = func(input string) string {
			return applyFilterRule(rule, input)
		}
	}

	// Run tests
	results := allTests.RunAllTests(filterFuncs)
	summary := toml.Summarize(results)

	// Print results
	if testVerbose {
		printVerboseResults(results)
	} else {
		printCompactResults(results)
	}

	// Print summary
	out.Global().Println()
	if summary.Failed == 0 && summary.Total > 0 {
		out.Global().Printf("%s\n", green(fmt.Sprintf("✓ All tests passed! (%d/%d)", summary.Passed, summary.Total)))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		out.Global().Printf("%s\n", red(fmt.Sprintf("✗ %d test(s) failed", summary.Failed)))
		out.Global().Print(summary.FormatSummary())
	}

	if summary.Failed > 0 {
		os.Exit(1)
	}

	return nil
}

func printVerboseResults(results []toml.TestResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for _, result := range results {
		if result.Skipped {
			out.Global().Printf("%s %s / %s", yellow("⏭"), result.FilterName, result.TestName)
			if result.Reason != "" {
				out.Global().Printf(" (%s)", result.Reason)
			}
			out.Global().Println()
			continue
		}

		if result.Passed {
			out.Global().Printf("%s %s / %s\n", green("✓"), result.FilterName, result.TestName)
		} else {
			out.Global().Printf("%s %s / %s\n", red("✗"), result.FilterName, result.TestName)
			if result.Error != nil {
				out.Global().Printf("  Error: %s\n", result.Error)
			} else {
				out.Global().Printf("  Expected: %q\n", result.Expected)
				out.Global().Printf("  Got:      %q\n", result.Got)
			}
		}
	}
}

func printCompactResults(results []toml.TestResult) {
	for _, result := range results {
		if result.Skipped {
			out.Global().Print("⏭")
		} else if result.Passed {
			out.Global().Print("✓")
		} else {
			out.Global().Print("✗")
		}
	}
	out.Global().Println()
}

func findFilterFiles() ([]string, error) {
	var files []string

	// Check builtin filters
	builtinPath := "internal/toml/builtin"
	if _, err := os.Stat(builtinPath); err == nil {
		matches, err := filepath.Glob(filepath.Join(builtinPath, "*.toml"))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}

	// Check user config directory
	configPath := config.ConfigPath()
	filtersPath := filepath.Join(filepath.Dir(configPath), "filters")
	if _, err := os.Stat(filtersPath); err == nil {
		matches, err := filepath.Glob(filepath.Join(filtersPath, "*.toml"))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}

	return files, nil
}

func parseFilterFile(path string) (*toml.TOMLFilter, error) {
	parser := toml.NewParser()
	return parser.ParseFile(path)
}

func applyFilterRule(rule toml.TOMLFilterRule, input string) string {
	output := input

	// Strip ANSI codes if requested
	if rule.StripANSI {
		// Simple ANSI stripping (could be more sophisticated)
		output = stripANSI(output)
	}

	// Apply line filtering
	lines := splitLines(output)
	filtered := filterLines(lines, rule)

	// Apply head/tail/max_lines
	if rule.Head > 0 && len(filtered) > rule.Head {
		filtered = filtered[:rule.Head]
	}
	if rule.Tail > 0 && len(filtered) > rule.Tail {
		filtered = filtered[len(filtered)-rule.Tail:]
	}
	if rule.MaxLines > 0 && len(filtered) > rule.MaxLines {
		filtered = filtered[:rule.MaxLines]
	}

	return joinLines(filtered)
}

// ansiEscapePattern matches ANSI escape sequences including:
// - Standard escape sequences: ESC[<params><letter>
// - Extended sequences: ESC[<params>;<params><letter>
// - OSC sequences: ESC]<params><BEL> or ESC]<params>ESC\
// - Cursor positioning, colors, styles, etc.
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]|\x1b\][0-9;]*(?:\x07|\x1b\\)|\x1b\[[\?0-9]*[hl]`)

func stripANSI(s string) string {
	return ansiEscapePattern.ReplaceAllString(s, "")
}

func filterLines(lines []string, rule toml.TOMLFilterRule) []string {
	var result []string

	for _, line := range lines {
		// Check strip patterns
		shouldStrip := false
		for _, pattern := range rule.StripLinesMatching {
			if matched, _ := matchPattern(pattern, line); matched {
				shouldStrip = true
				break
			}
		}

		if shouldStrip {
			continue
		}

		// Check keep patterns (if any)
		if len(rule.KeepLinesMatching) > 0 {
			shouldKeep := false
			for _, pattern := range rule.KeepLinesMatching {
				if matched, _ := matchPattern(pattern, line); matched {
					shouldKeep = true
					break
				}
			}

			if !shouldKeep {
				continue
			}
		}

		result = append(result, line)
	}

	return result
}

// regexCache stores compiled regex patterns for reuse
var regexCache = &sync.Map{}

func matchPattern(pattern, text string) (bool, error) {
	// Check if pattern contains regex metacharacters
	if !containsRegexMetachars(pattern) {
		// Use simple glob matching for simple patterns
		return filepath.Match(pattern, text)
	}

	// Try to get cached regex
	if cached, ok := regexCache.Load(pattern); ok {
		re := cached.(*regexp.Regexp)
		return re.MatchString(text), nil
	}

	// Compile and cache new regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fall back to glob matching for invalid regex
		return filepath.Match(pattern, text)
	}

	regexCache.Store(pattern, re)
	return re.MatchString(text), nil
}

func containsRegexMetachars(s string) bool {
	// Check for regex-specific metacharacters not used in glob patterns
	metachars := []string{"^", "$", "+", "?", "|", "(", ")", "[", "]", "{", "}"}
	for _, char := range metachars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := []string{}
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}
