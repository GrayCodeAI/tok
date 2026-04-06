package toml

import (
	"fmt"
	"strings"
	"unicode"
)

// SafetyCheck represents a safety check result.
type SafetyCheck struct {
	Passed bool
	Issues []SafetyIssue
}

// SafetyIssue represents a safety issue found in a filter.
type SafetyIssue struct {
	Severity string // "critical", "warning", "info"
	Message  string
	Line     int
}

// CheckFilterSafety performs safety checks on a TOML filter definition.
// Inspired by tokf's filter safety checks.
func CheckFilterSafety(content string) SafetyCheck {
	result := SafetyCheck{Passed: true}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Check for prompt injection patterns
		if containsPromptInjection(trimmed) {
			result.Issues = append(result.Issues, SafetyIssue{
				Severity: "critical",
				Message:  "Potential prompt injection pattern detected",
				Line:     lineNum,
			})
			result.Passed = false
		}

		// Check for shell injection patterns
		if containsShellInjection(trimmed) {
			result.Issues = append(result.Issues, SafetyIssue{
				Severity: "critical",
				Message:  "Potential shell injection pattern detected",
				Line:     lineNum,
			})
			result.Passed = false
		}

		// Check for hidden Unicode characters
		if containsHiddenUnicode(trimmed) {
			result.Issues = append(result.Issues, SafetyIssue{
				Severity: "warning",
				Message:  "Hidden Unicode characters detected",
				Line:     lineNum,
			})
		}

		// Check for overly broad patterns
		if containsCatchAllPattern(trimmed) {
			result.Issues = append(result.Issues, SafetyIssue{
				Severity: "warning",
				Message:  "Overly broad pattern may match unintended content",
				Line:     lineNum,
			})
		}

		// Check for dangerous replace patterns
		if containsDangerousReplace(trimmed) {
			result.Issues = append(result.Issues, SafetyIssue{
				Severity: "warning",
				Message:  "Replace pattern may remove critical content",
				Line:     lineNum,
			})
		}
	}

	return result
}

// TestSuite represents a filter test suite.
type TestSuite struct {
	Name  string
	Tests []TestCase
}

// TestCase represents a single test case.
type TestCase struct {
	Name     string
	Input    string
	Expected string
	Filter   string
}

// RunTestSuite runs a test suite against a filter.
func RunTestSuite(suite TestSuite, filterFunc func(string, string) string) (int, int, []string) {
	passed := 0
	failed := 0
	var failures []string

	for _, tc := range suite.Tests {
		result := filterFunc(tc.Input, tc.Filter)
		if result == tc.Expected {
			passed++
		} else {
			failed++
			failures = append(failures, fmt.Sprintf("%s: expected %q, got %q", tc.Name, tc.Expected, result))
		}
	}

	return passed, failed, failures
}

// ValidateFilterConfig validates a filter configuration for correctness.
func ValidateFilterConfig(content string) []string {
	var errors []string

	if !strings.Contains(content, "[filters.") {
		errors = append(errors, "No filter sections found")
	}

	if !strings.Contains(content, "pattern") && !strings.Contains(content, "match") {
		errors = append(errors, "No pattern or match rules found")
	}

	// Check for unclosed brackets
	openBrackets := strings.Count(content, "[")
	closeBrackets := strings.Count(content, "]")
	if openBrackets != closeBrackets {
		errors = append(errors, "Unclosed brackets detected")
	}

	// Count unescaped quotes by subtracting escaped occurrences.
	// TOML uses \' and \" as escape sequences inside strings.
	unescapedSingle := strings.Count(content, "'") - strings.Count(content, "\\'")
	unescapedDouble := strings.Count(content, "\"") - strings.Count(content, "\\\"")
	if unescapedSingle%2 != 0 {
		errors = append(errors, "Unclosed single quote detected")
	}
	if unescapedDouble%2 != 0 {
		errors = append(errors, "Unclosed double quote detected")
	}

	return errors
}

func containsPromptInjection(line string) bool {
	injectionPatterns := []string{
		"ignore all",
		"ignore previous",
		"disregard",
		"you are now",
		"act as",
		"pretend to",
		"system prompt",
		"developer mode",
	}
	lower := strings.ToLower(line)
	for _, p := range injectionPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func containsShellInjection(line string) bool {
	shellPatterns := []string{
		"$(rm -rf",
		"$(curl",
		"$(wget",
		"`rm -rf",
		"`curl",
		"`wget",
		"; rm -rf",
		"| rm -rf",
		"&& rm -rf",
	}
	for _, p := range shellPatterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

func containsHiddenUnicode(line string) bool {
	for _, r := range line {
		// Zero-width characters, invisible format controls
		if (r >= 0x200B && r <= 0x200F) || // zero-width space, joiners, LTR/RTL marks
			(r >= 0x202A && r <= 0x202E) || // LRE, RLE, PDF, LRO, RLO
			(r >= 0x2060 && r <= 0x2064) || // word joiner, invisible separators
			(r >= 0x2066 && r <= 0x2069) || // LRI, RLI, FSI, PDI
			r == 0x00AD || // soft hyphen
			r == 0xFEFF || // BOM / zero-width no-break space
			(r >= 0xFFF0 && r <= 0xFFF8) { // specials
			return true
		}
	}
	return false
}

func containsCatchAllPattern(line string) bool {
	catchAllPatterns := []string{
		`pattern = "."`,
		`pattern = ".*"`,
		`pattern = '.+'`,
		`match = "*"`,
		`match = "**"`,
	}
	for _, p := range catchAllPatterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

func containsDangerousReplace(line string) bool {
	dangerousPatterns := []string{
		`replace = ""`,
		`replace = ''`,
		`strip = "*"`,
		`strip = "**"`,
		`remove = "all"`,
	}
	for _, p := range dangerousPatterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

// FormatSafetyReport returns a human-readable safety report.
func FormatSafetyReport(check SafetyCheck) string {
	if check.Passed && len(check.Issues) == 0 {
		return "✅ All safety checks passed"
	}

	var sb strings.Builder
	if check.Passed {
		sb.WriteString("⚠️  Safety checks passed with warnings:\n")
	} else {
		sb.WriteString("❌ Safety checks failed:\n")
	}

	for _, issue := range check.Issues {
		severity := strings.ToUpper(issue.Severity[:1]) + issue.Severity[1:]
		sb.WriteString(fmt.Sprintf("  [%s] Line %d: %s\n", severity, issue.Line, issue.Message))
	}

	return sb.String()
}

// IsPrintableASCII checks if a string contains only printable ASCII.
func IsPrintableASCII(s string) bool {
	for _, r := range s {
		if r > 127 && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
