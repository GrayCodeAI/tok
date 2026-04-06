package filtercmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/toml"
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
  tokman filter tests              # Run all tests in all filters
  tokman filter tests git_status   # Run tests for specific filter
  tokman filter tests -v           # Verbose output`,
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
		fmt.Println("No filter files found")
		return nil
	}

	// Parse all filters and collect tests
	allTests := toml.NewFilterTestSuite()
	filterRules := make(map[string]toml.TOMLFilterRule)

	for _, file := range filterFiles {
		filter, err := parseFilterFile(file)
		if err != nil {
			if testVerbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file, err)
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
				fmt.Fprintf(os.Stderr, "Warning: failed to parse tests from %s: %v\n", file, err)
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
		fmt.Println("No tests found in filter files")
		fmt.Println("\nTo add tests, edit your filter TOML file and add [[tests.filtername]] sections.")
		return nil
	}

	// Print header
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s\n", cyan("Running filter tests..."))
	fmt.Printf("Found %d test(s)\n\n", totalTests)

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
	fmt.Println()
	if summary.Failed == 0 && summary.Total > 0 {
		fmt.Printf("%s\n", green(fmt.Sprintf("✓ All tests passed! (%d/%d)", summary.Passed, summary.Total)))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s\n", red(fmt.Sprintf("✗ %d test(s) failed", summary.Failed)))
		fmt.Print(summary.FormatSummary())
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
			fmt.Printf("%s %s / %s", yellow("⏭"), result.FilterName, result.TestName)
			if result.Reason != "" {
				fmt.Printf(" (%s)", result.Reason)
			}
			fmt.Println()
			continue
		}

		if result.Passed {
			fmt.Printf("%s %s / %s\n", green("✓"), result.FilterName, result.TestName)
		} else {
			fmt.Printf("%s %s / %s\n", red("✗"), result.FilterName, result.TestName)
			if result.Error != nil {
				fmt.Printf("  Error: %s\n", result.Error)
			} else {
				fmt.Printf("  Expected: %q\n", result.Expected)
				fmt.Printf("  Got:      %q\n", result.Got)
			}
		}
	}
}

func printCompactResults(results []toml.TestResult) {
	for _, result := range results {
		if result.Skipped {
			fmt.Print("⏭")
		} else if result.Passed {
			fmt.Print("✓")
		} else {
			fmt.Print("✗")
		}
	}
	fmt.Println()
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

func stripANSI(s string) string {
	// Simple ANSI escape sequence removal
	// TODO: Use more robust regex
	result := ""
	inEscape := false
	for _, ch := range s {
		if ch == '\x1b' {
			inEscape = true
		} else if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
		} else {
			result += string(ch)
		}
	}
	return result
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

func matchPattern(pattern, text string) (bool, error) {
	// Simple pattern matching for now
	// TODO: Use compiled regex cache
	return filepath.Match(pattern, text)
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
