package toml

import (
	"fmt"
	"strings"
)

// TOMLFilterTest represents an inline test case for a filter
type TOMLFilterTest struct {
	Name     string `toml:"name"`
	Input    string `toml:"input"`
	Expected string `toml:"expected"`
	Skip     bool   `toml:"skip"`
	Reason   string `toml:"reason"`
}

// TestResult represents the result of running a single test
type TestResult struct {
	FilterName string
	TestName   string
	Passed     bool
	Expected   string
	Got        string
	Error      error
	Skipped    bool
	Reason     string
}

// FilterTestSuite represents a collection of tests for filters
type FilterTestSuite struct {
	Tests map[string][]TOMLFilterTest // map[filterName][]tests
}

// NewFilterTestSuite creates a new filter test suite
func NewFilterTestSuite() *FilterTestSuite {
	return &FilterTestSuite{
		Tests: make(map[string][]TOMLFilterTest),
	}
}

// AddTest adds a test to the suite
func (ts *FilterTestSuite) AddTest(filterName string, test TOMLFilterTest) {
	ts.Tests[filterName] = append(ts.Tests[filterName], test)
}

// GetTests returns all tests for a specific filter
func (ts *FilterTestSuite) GetTests(filterName string) []TOMLFilterTest {
	return ts.Tests[filterName]
}

// TotalTests returns the total number of tests
func (ts *FilterTestSuite) TotalTests() int {
	total := 0
	for _, tests := range ts.Tests {
		total += len(tests)
	}
	return total
}

// ParseTests extracts test cases from TOML raw content
func ParseTests(raw map[string]any) (*FilterTestSuite, error) {
	suite := NewFilterTestSuite()

	// Look for "tests" section
	testsSection, ok := raw["tests"]
	if !ok {
		return suite, nil // No tests defined
	}

	testsMap, ok := testsSection.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("tests section must be a map")
	}

	// Parse each filter's tests
	for filterName, testsData := range testsMap {
		testsList, ok := testsData.([]map[string]any)
		if !ok {
			return nil, fmt.Errorf("tests for '%s' must be an array", filterName)
		}

		for i, testData := range testsList {
			test, err := parseTest(testData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse test %d for '%s': %w", i, filterName, err)
			}
			suite.AddTest(filterName, test)
		}
	}

	return suite, nil
}

// parseTest parses a single test case
func parseTest(data map[string]any) (TOMLFilterTest, error) {
	test := TOMLFilterTest{}

	if name, ok := data["name"].(string); ok {
		test.Name = name
	} else {
		return test, fmt.Errorf("test must have a 'name' field")
	}

	if input, ok := data["input"].(string); ok {
		test.Input = input
	} else {
		return test, fmt.Errorf("test must have an 'input' field")
	}

	if expected, ok := data["expected"].(string); ok {
		test.Expected = expected
	} else {
		return test, fmt.Errorf("test must have an 'expected' field")
	}

	if skip, ok := data["skip"].(bool); ok {
		test.Skip = skip
	}

	if reason, ok := data["reason"].(string); ok {
		test.Reason = reason
	}

	return test, nil
}

// RunTest executes a single test
func (ts *FilterTestSuite) RunTest(filterName string, test TOMLFilterTest, filterFunc func(string) string) TestResult {
	result := TestResult{
		FilterName: filterName,
		TestName:   test.Name,
		Expected:   test.Expected,
		Skipped:    test.Skip,
		Reason:     test.Reason,
	}

	if test.Skip {
		return result
	}

	// Run the filter
	got := filterFunc(test.Input)
	result.Got = got

	// Compare output (trim whitespace for comparison)
	expectedTrimmed := strings.TrimSpace(test.Expected)
	gotTrimmed := strings.TrimSpace(got)

	if expectedTrimmed == gotTrimmed {
		result.Passed = true
	} else {
		result.Passed = false
	}

	return result
}

// RunAllTests executes all tests in the suite
func (ts *FilterTestSuite) RunAllTests(filterFuncs map[string]func(string) string) []TestResult {
	var results []TestResult

	for filterName, tests := range ts.Tests {
		filterFunc, ok := filterFuncs[filterName]
		if !ok {
			// No filter function provided, skip tests for this filter
			for _, test := range tests {
				results = append(results, TestResult{
					FilterName: filterName,
					TestName:   test.Name,
					Passed:     false,
					Error:      fmt.Errorf("filter function not provided"),
				})
			}
			continue
		}

		for _, test := range tests {
			result := ts.RunTest(filterName, test, filterFunc)
			results = append(results, result)
		}
	}

	return results
}

// Summary returns a summary of test results
type TestSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Results []TestResult
}

// Summarize creates a summary from test results
func Summarize(results []TestResult) TestSummary {
	summary := TestSummary{
		Total:   len(results),
		Results: results,
	}

	for _, result := range results {
		if result.Skipped {
			summary.Skipped++
		} else if result.Passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}

	return summary
}

// FormatSummary formats a test summary for display
func (s TestSummary) FormatSummary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total: %d, Passed: %d, Failed: %d, Skipped: %d\n",
		s.Total, s.Passed, s.Failed, s.Skipped))

	if s.Failed > 0 {
		sb.WriteString("\nFailed tests:\n")
		for _, result := range s.Results {
			if !result.Passed && !result.Skipped {
				sb.WriteString(fmt.Sprintf("  ❌ %s / %s\n", result.FilterName, result.TestName))
				if result.Error != nil {
					sb.WriteString(fmt.Sprintf("     Error: %s\n", result.Error))
				} else {
					sb.WriteString(fmt.Sprintf("     Expected: %q\n", result.Expected))
					sb.WriteString(fmt.Sprintf("     Got:      %q\n", result.Got))
				}
			}
		}
	}

	if s.Skipped > 0 {
		sb.WriteString("\nSkipped tests:\n")
		for _, result := range s.Results {
			if result.Skipped {
				sb.WriteString(fmt.Sprintf("  ⏭️  %s / %s", result.FilterName, result.TestName))
				if result.Reason != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", result.Reason))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}
