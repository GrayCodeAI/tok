package toml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilterTestSuite represents a declarative test suite for TOML filters.
// Inspired by tokf's filter test suites.
type FilterTestSuite struct {
	Name  string
	Tests []FilterTestCase
}

// FilterTestCase represents a single filter test case.
type FilterTestCase struct {
	Name     string
	Input    string
	Expected string
	Filter   string
}

// NewFilterTestSuite creates a new filter test suite from a directory.
func NewFilterTestSuite(dir string) (*FilterTestSuite, error) {
	suite := &FilterTestSuite{Name: filepath.Base(dir)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".toml") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			// Parse TOML test case
			tc := parseTestCase(string(data))
			if tc.Name != "" {
				suite.Tests = append(suite.Tests, tc)
			}
		}
	}

	return suite, nil
}

// Run runs the test suite against a filter function.
func (suite *FilterTestSuite) Run(filterFunc func(string, string) string) (int, int, []string) {
	passed := 0
	failed := 0
	var failures []string

	for _, tc := range suite.Tests {
		result := filterFunc(tc.Input, tc.Filter)
		if strings.TrimSpace(result) == strings.TrimSpace(tc.Expected) {
			passed++
		} else {
			failed++
			failures = append(failures, fmt.Sprintf("%s: expected %q, got %q", tc.Name, tc.Expected, result))
		}
	}

	return passed, failed, failures
}

// FormatResults formats test results.
func FormatResults(passed, failed int, failures []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Test Results: %d passed, %d failed\n", passed, failed))
	for _, f := range failures {
		sb.WriteString(fmt.Sprintf("  ❌ %s\n", f))
	}
	if failed == 0 {
		sb.WriteString("✅ All tests passed\n")
	}
	return sb.String()
}

func parseTestCase(data string) FilterTestCase {
	tc := FilterTestCase{}
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name = ") {
			tc.Name = strings.Trim(strings.TrimPrefix(line, "name = "), "\"")
		} else if strings.HasPrefix(line, "input = ") {
			tc.Input = strings.Trim(strings.TrimPrefix(line, "input = "), "\"")
		} else if strings.HasPrefix(line, "expected = ") {
			tc.Expected = strings.Trim(strings.TrimPrefix(line, "expected = "), "\"")
		} else if strings.HasPrefix(line, "filter = ") {
			tc.Filter = strings.Trim(strings.TrimPrefix(line, "filter = "), "\"")
		}
	}
	return tc
}
