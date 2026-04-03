// Package tokftest provides the `tokman verify` command.
package tokftest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/fatih/color"
)

// CLI provides the verify command interface.
type CLI struct {
	parser  *Parser
	runner  *Runner
	verbose bool
	color   bool
}

// NewCLI creates a new CLI.
func NewCLI(fixtureDir string, verbose, useColor bool) *CLI {
	parser := NewParser(fixtureDir)
	runner := NewRunner(&defaultFilterLoader{})
	return &CLI{
		parser:  parser,
		runner:  runner,
		verbose: verbose,
		color:   useColor,
	}
}

// RunTests runs all test files in a directory.
func (c *CLI) RunTests(testDir string) error {
	specs, err := c.parser.ParseDir(testDir)
	if err != nil {
		return fmt.Errorf("failed to parse tests: %w", err)
	}

	if len(specs) == 0 {
		fmt.Println("No test files found")
		return nil
	}

	var totalPassed, totalFailed, totalSkipped int
	start := time.Now()

	for _, spec := range specs {
		if c.verbose {
			fmt.Printf("\nRunning: %s\n", spec.Name)
			if spec.Description != "" {
				fmt.Printf("  %s\n", spec.Description)
			}
		}

		result := c.runner.Run(spec)

		if result.Skipped {
			totalSkipped++
			c.printSkipped(spec.Name)
			continue
		}

		if result.Passed {
			totalPassed++
			c.printPassed(spec.Name)
		} else {
			totalFailed++
			c.printFailed(spec.Name)
		}

		// Print case details on failure or verbose
		if c.verbose || !result.Passed {
			for _, caseResult := range result.Cases {
				if !caseResult.Passed || c.verbose {
					c.printCaseResult(caseResult)
				}
			}
		}

		// Print errors
		for _, err := range result.Errors {
			c.printError(err)
		}
	}

	duration := time.Since(start)

	// Summary
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Results: %d passed, %d failed, %d skipped\n", totalPassed, totalFailed, totalSkipped)
	fmt.Printf("Duration: %v\n", duration)

	if totalFailed > 0 {
		return fmt.Errorf("%d test(s) failed", totalFailed)
	}
	return nil
}

// RunTest runs a single test file.
func (c *CLI) RunTest(testFile string) error {
	spec, err := c.parser.ParseFile(testFile)
	if err != nil {
		return err
	}

	result := c.runner.Run(spec)

	if result.Skipped {
		c.printSkipped(spec.Name)
		return nil
	}

	if result.Passed {
		c.printPassed(spec.Name)
	} else {
		c.printFailed(spec.Name)
	}

	// Always print case details for single test
	for _, caseResult := range result.Cases {
		c.printCaseResult(caseResult)
	}

	for _, err := range result.Errors {
		c.printError(err)
	}

	if !result.Passed {
		return fmt.Errorf("test failed: %s", spec.Name)
	}
	return nil
}

// RunTestsByTag runs tests matching a tag.
func (c *CLI) RunTestsByTag(testDir, tag string) error {
	specs, err := c.parser.ParseDir(testDir)
	if err != nil {
		return err
	}

	var matching []*TestSpec
	for _, spec := range specs {
		for _, t := range spec.Tags {
			if t == tag {
				matching = append(matching, spec)
				break
			}
		}
	}

	if len(matching) == 0 {
		fmt.Printf("No tests found with tag: %s\n", tag)
		return nil
	}

	fmt.Printf("Running %d tests with tag '%s'\n\n", len(matching), tag)

	var passed, failed int
	for _, spec := range matching {
		result := c.runner.Run(spec)
		if result.Passed && !result.Skipped {
			passed++
			c.printPassed(spec.Name)
		} else if !result.Skipped {
			failed++
			c.printFailed(spec.Name)
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
	if failed > 0 {
		return fmt.Errorf("%d test(s) failed", failed)
	}
	return nil
}

// GenerateFixture generates a fixture file from a test.
func (c *CLI) GenerateFixture(spec *TestSpec, outputDir string) error {
	for _, fixture := range spec.Fixtures {
		if fixture.Name == "" {
			continue
		}

		filename := fmt.Sprintf("%s.fixture.toml", fixture.Name)
		path := filepath.Join(outputDir, filename)

		content := fmt.Sprintf("# Fixture: %s\n# From test: %s\n\ncontent = '''\n%s\n'''\n",
			fixture.Name, spec.Name, fixture.Content)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write fixture %s: %w", filename, err)
		}

		fmt.Printf("Generated: %s\n", path)
	}

	return nil
}

// Print helpers

func (c *CLI) printPassed(name string) {
	if c.color {
		color.Green("✓ %s", name)
	} else {
		fmt.Printf("PASS: %s\n", name)
	}
}

func (c *CLI) printFailed(name string) {
	if c.color {
		color.Red("✗ %s", name)
	} else {
		fmt.Printf("FAIL: %s\n", name)
	}
}

func (c *CLI) printSkipped(name string) {
	if c.color {
		color.Yellow("⊘ %s (skipped)", name)
	} else {
		fmt.Printf("SKIP: %s\n", name)
	}
}

func (c *CLI) printCaseResult(result CaseResult) {
	indent := "  "
	if result.Passed {
		fmt.Printf("%s✓ %s\n", indent, result.Name)
	} else {
		fmt.Printf("%s✗ %s\n", indent, result.Name)
		for _, err := range result.Errors {
			fmt.Printf("%s  Error: %s\n", indent, err)
		}
	}

	if c.verbose {
		fmt.Printf("%s  Input tokens: ~%d\n", indent, len(result.Input)/4)
		fmt.Printf("%s  Output tokens: %d\n", indent, result.Tokens)
		fmt.Printf("%s  Saved: %d\n", indent, result.Saved)
	}
}

func (c *CLI) printError(err string) {
	if c.color {
		color.Red("  Error: %s", err)
	} else {
		fmt.Printf("  Error: %s\n", err)
	}
}

// defaultFilterLoader implements FilterLoader using TokMan's filter package.
type defaultFilterLoader struct{}

func (l *defaultFilterLoader) Load(name string) (*filter.Engine, error) {
	switch name {
	case "minimal":
		return filter.NewEngine(filter.ModeMinimal), nil
	case "aggressive":
		return filter.NewEngine(filter.ModeAggressive), nil
	case "none":
		return filter.NewEngine(filter.ModeNone), nil
	default:
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
}

func (l *defaultFilterLoader) LoadFromFile(path string) (*filter.Engine, error) {
	// For now, just return minimal mode
	// In production, this would load a TOML filter definition
	return filter.NewEngine(filter.ModeMinimal), nil
}

// Report generates a test report.
type Report struct {
	TotalTests  int
	Passed      int
	Failed      int
	Skipped     int
	TotalCases  int
	TotalTokens int
	TotalSaved  int
	Duration    time.Duration
	TestResults []*TestResult
}

// GenerateReport generates a comprehensive report.
func GenerateReport(results []*TestResult) *Report {
	report := &Report{
		TestResults: results,
	}

	for _, result := range results {
		report.TotalTests++
		if result.Skipped {
			report.Skipped++
		} else if result.Passed {
			report.Passed++
		} else {
			report.Failed++
		}

		for _, c := range result.Cases {
			report.TotalCases++
			report.TotalTokens += c.Tokens
			report.TotalSaved += c.Saved
		}
	}

	return report
}

// Print prints the report.
func (r *Report) Print() {
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("Test Report")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Printf("Total Tests:  %d\n", r.TotalTests)
	fmt.Printf("Passed:       %d (%.1f%%)\n", r.Passed, float64(r.Passed)/float64(r.TotalTests)*100)
	fmt.Printf("Failed:       %d (%.1f%%)\n", r.Failed, float64(r.Failed)/float64(r.TotalTests)*100)
	fmt.Printf("Skipped:      %d\n", r.Skipped)
	fmt.Println("-" + strings.Repeat("-", 60))
	fmt.Printf("Total Cases:  %d\n", r.TotalCases)
	fmt.Printf("Total Tokens: %d\n", r.TotalTokens)
	fmt.Printf("Total Saved:  %d (%.1f%%)\n", r.TotalSaved, float64(r.TotalSaved)/float64(r.TotalTokens+r.TotalSaved)*100)
	fmt.Println("=" + strings.Repeat("=", 60))
}
