package toml

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

// SafetyCheck represents the result of a safety validation
type SafetyCheck struct {
	Passed bool
	Issues []SafetyIssue
}

// SafetyIssue represents a single safety issue found
type SafetyIssue struct {
	Severity string
	Message  string
	Line     int
}

// CheckFilterSafety performs safety checks on TOML filter content
func CheckFilterSafety(content string) SafetyCheck {
	var issues []SafetyIssue

	// Check for prompt injection attempts
	promptInjectionPatterns := []string{
		`ignore all previous instructions`,
		`ignore previous instructions`,
		`disregard all prior`,
		`forget everything`,
		`system prompt`,
		`you are now`,
		`new role`,
	}
	lowerContent := strings.ToLower(content)
	for _, pattern := range promptInjectionPatterns {
		if strings.Contains(lowerContent, pattern) {
			issues = append(issues, SafetyIssue{
				Severity: "critical",
				Message:  "Potential prompt injection detected: " + pattern,
			})
		}
	}

	// Check for shell injection patterns (using separate patterns for clarity)
	patterns := []string{
		`\$\([^)]+\)`,   // $(cmd)
		"`[^`]+`",       // `cmd`
		`\|[\s]*[a-z]+`, // |cmd
	}
	shellInjectionPattern := regexp.MustCompile(strings.Join(patterns, "|"))
	if shellInjectionPattern.MatchString(content) {
		issues = append(issues, SafetyIssue{
			Severity: "critical",
			Message:  "Shell command injection detected",
		})
	}

	// Check for hidden Unicode characters
	if hasHiddenUnicode(content) {
		issues = append(issues, SafetyIssue{
			Severity: "warning",
			Message:  "Hidden Unicode characters detected",
		})
	}

	// Check for non-printable characters (except common whitespace)
	if hasNonPrintableChars(content) {
		issues = append(issues, SafetyIssue{
			Severity: "warning",
			Message:  "Non-printable characters detected",
		})
	}

	return SafetyCheck{
		Passed: len(issues) == 0,
		Issues: issues,
	}
}

// ValidateFilterConfig validates TOML filter configuration syntax
func ValidateFilterConfig(content string) []error {
	var errors []error

	// Try to parse as TOML
	var config map[string]interface{}
	if _, err := toml.Decode(content, &config); err != nil {
		errors = append(errors, fmt.Errorf("invalid TOML syntax: %w", err))
		return errors
	}

	// Check for required filter structure
	filters, ok := config["filters"].(map[string]interface{})
	if !ok {
		errors = append(errors, fmt.Errorf("missing 'filters' section"))
		return errors
	}

	// Validate each filter entry
	for name, filterData := range filters {
		filter, ok := filterData.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("filter '%s' is not a valid table", name))
			continue
		}

		// Check for pattern field
		if _, hasPattern := filter["pattern"]; !hasPattern {
			errors = append(errors, fmt.Errorf("filter '%s' missing required 'pattern' field", name))
		}

		// Validate pattern is a string
		if pattern, ok := filter["pattern"].(string); ok {
			if strings.TrimSpace(pattern) == "" {
				errors = append(errors, fmt.Errorf("filter '%s' has empty pattern", name))
			}
			// Try to compile as regex
			if _, err := regexp.Compile(pattern); err != nil {
				errors = append(errors, fmt.Errorf("filter '%s' has invalid regex pattern: %w", name, err))
			}
		}
	}

	return errors
}

// FormatSafetyReport formats a safety check result as a human-readable string
func FormatSafetyReport(check SafetyCheck) string {
	if check.Passed {
		return "Safety check passed: No issues found"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Safety check failed: %d issue(s) found", len(check.Issues)))

	for i, issue := range check.Issues {
		parts = append(parts, fmt.Sprintf("  %d. [%s] %s", i+1, issue.Severity, issue.Message))
	}

	return strings.Join(parts, "\n")
}

// IsPrintableASCII checks if a string contains only printable ASCII characters
func IsPrintableASCII(s string) bool {
	for _, r := range s {
		if r > 127 || (r < 32 && r != '\t' && r != '\n' && r != '\r') {
			return false
		}
	}
	return true
}

// hasHiddenUnicode checks for hidden/invisible Unicode characters
func hasHiddenUnicode(s string) bool {
	// Zero-width characters and other invisible Unicode
	hiddenChars := []rune{
		'\u200B', // Zero-width space
		'\u200C', // Zero-width non-joiner
		'\u200D', // Zero-width joiner
		'\uFEFF', // Byte order mark
		'\u2060', // Word joiner
		'\u180E', // Mongolian vowel separator
		'\u200E', // Left-to-right mark
		'\u200F', // Right-to-left mark
	}

	for _, r := range s {
		for _, hidden := range hiddenChars {
			if r == hidden {
				return true
			}
		}
	}
	return false
}

// hasNonPrintableChars checks for non-printable characters
func hasNonPrintableChars(s string) bool {
	for _, r := range s {
		// Allow printable ASCII and common whitespace
		if !utf8.ValidRune(r) {
			return true
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
		if r == 127 { // DEL character
			return true
		}
	}
	return false
}
