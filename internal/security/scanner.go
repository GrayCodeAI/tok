// Package security provides content scanning and PII redaction capabilities
// for detecting sensitive information in command output.
package security

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// Severity levels for findings
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Finding represents a security finding from content scanning
type Finding struct {
	Rule     string
	Severity string
	Message  string
	Match    string
	Position int
}

// Scanner provides security scanning capabilities
type Scanner struct {
	rules []ScanRule
}

// ContextValidator is an optional interface for rules that need additional
// context validation beyond regex matching (e.g., to reduce false positives).
type ContextValidator interface {
	// ValidateContext returns true if the match at position pos in content is a true positive.
	ValidateContext(content string, pos int, match string) bool
}

// awsSecretContextValidator checks that a 40-char base64 string is likely an
// AWS secret by looking for AWS-related keywords within a 60-char window.
type awsSecretContextValidator struct{}

var awsContextKeywords = []string{
	"aws", "secret", "access", "credential", "key", "token",
	"AKIA", "ASIA", "AROA", "AIDA",
}

func (awsSecretContextValidator) ValidateContext(content string, pos int, match string) bool {
	// Check 60 chars before the match for AWS-related context
	start := pos - 60
	if start < 0 {
		start = 0
	}
	window := strings.ToLower(content[start:pos])
	for _, kw := range awsContextKeywords {
		if strings.Contains(window, kw) {
			return true
		}
	}
	// Also check after the match (in case of "key = <secret>" format)
	end := pos + len(match) + 60
	if end > len(content) {
		end = len(content)
	}
	afterWindow := strings.ToLower(content[pos:end])
	for _, kw := range awsContextKeywords {
		if strings.Contains(afterWindow, kw) {
			return true
		}
	}
	return false
}

// ScanRule defines a single scanning rule
type ScanRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    string
	Description string
	// ContextValidator, if non-nil, is called for each match to reduce false positives.
	ContextValidator ContextValidator
}

// Pre-compiled regexes for performance (PERF-6)
var (
	scannerRulesOnce sync.Once
	scannerRules     []ScanRule
)

// initScanner initializes the scanner rules once
func initScanner() []ScanRule {
	scannerRulesOnce.Do(func() {
		scannerRules = []ScanRule{
			{
				Name:        "aws_access_key",
				Pattern:     regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
				Severity:    SeverityCritical,
				Description: "AWS Access Key ID",
			},
			{
				Name:             "aws_secret_key",
				Pattern:          regexp.MustCompile(`(?:^|[^a-zA-Z0-9/+=])([A-Za-z0-9/+=]{40})(?:[^a-zA-Z0-9/+=]|$)`),
				Severity:         SeverityCritical,
				Description:      "AWS Secret Access Key",
				ContextValidator: awsSecretContextValidator{},
			},
			{
				Name:        "github_token",
				Pattern:     regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
				Severity:    SeverityCritical,
				Description: "GitHub Personal Access Token",
			},
			{
				Name:        "github_oauth",
				Pattern:     regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),
				Severity:    SeverityCritical,
				Description: "GitHub OAuth Token",
			},
			{
				Name:        "slack_token",
				Pattern:     regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[a-zA-Z0-9]{24})?`),
				Severity:    SeverityCritical,
				Description: "Slack Token",
			},
			{
				Name:        "private_key",
				Pattern:     regexp.MustCompile(`-----BEGIN (RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`),
				Severity:    SeverityCritical,
				Description: "Private Key Header",
			},
			{
				Name:        "private_key_content",
				Pattern:     regexp.MustCompile(`MII[A-Za-z0-9+/]{10,}={0,2}`),
				Severity:    SeverityCritical,
				Description: "Private Key Content",
			},
			{
				Name:        "credit_card",
				Pattern:     regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12}|(?:2131|1800|35\d{3})\d{11})\b`),
				Severity:    SeverityHigh,
				Description: "Credit Card Number",
			},
			{
				Name:        "ssn",
				Pattern:     regexp.MustCompile(`\b[0-9]{3}-[0-9]{2}-[0-9]{4}\b`),
				Severity:    SeverityHigh,
				Description: "Social Security Number",
			},
			{
				Name:        "email",
				Pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
				Severity:    SeverityMedium,
				Description: "Email Address",
			},
			{
				Name:        "database_url",
				Pattern:     regexp.MustCompile(`(?:postgres|mysql|mongodb|redis)://[^:]+:[^@]+@`),
				Severity:    SeverityHigh,
				Description: "Database Connection String",
			},
			{
				Name:        "jwt_token",
				Pattern:     regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
				Severity:    SeverityHigh,
				Description: "JWT Token",
			},
			{
				Name:        "url_password",
				Pattern:     regexp.MustCompile(`[a-zA-Z]+://[^:]+:[^@]+@`),
				Severity:    SeverityCritical,
				Description: "Password in URL",
			},
			{
				Name:        "bearer_token",
				Pattern:     regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_\-\.]+`),
				Severity:    SeverityHigh,
				Description: "Bearer Token",
			},
			{
				Name:        "api_key",
				Pattern:     regexp.MustCompile(`(?i)(?:api[_-]?key|apikey)[:=\s]+["']?[a-zA-Z0-9]{32,}["']?`),
				Severity:    SeverityHigh,
				Description: "API Key",
			},
			{
				Name:        "google_api_key",
				Pattern:     regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
				Severity:    SeverityCritical,
				Description: "Google API Key",
			},
			{
				Name:        "stripe_key",
				Pattern:     regexp.MustCompile(`(?:sk_live|pk_live|sk_test|pk_test)_[0-9a-zA-Z]{24,}`),
				Severity:    SeverityCritical,
				Description: "Stripe API Key",
			},
		}
	})
	return scannerRules
}

// NewScanner creates a new security scanner with all rules configured
func NewScanner() *Scanner {
	return &Scanner{
		rules: initScanner(),
	}
}

// Scan analyzes content for sensitive information and returns findings.
// If a rule has a ContextValidator, matches that fail validation are skipped.
func (s *Scanner) Scan(content string) []Finding {
	var findings []Finding

	for _, rule := range s.rules {
		matches := rule.Pattern.FindAllStringIndex(content, -1)
		for _, match := range matches {
			m := content[match[0]:match[1]]
			if rule.ContextValidator != nil {
				if !rule.ContextValidator.ValidateContext(content, match[0], m) {
					continue
				}
			}
			findings = append(findings, Finding{
				Rule:     rule.Name,
				Severity: rule.Severity,
				Message:  rule.Description,
				Match:    m,
				Position: match[0],
			})
		}
	}

	return findings
}

// ScanWithRedaction scans content and returns redacted version along with findings
func (s *Scanner) ScanWithRedaction(content string) (string, []Finding) {
	findings := s.Scan(content)
	redacted := RedactPII(content)
	return redacted, findings
}

// HasCriticalFindings checks if content contains any critical severity findings
func (s *Scanner) HasCriticalFindings(content string) bool {
	findings := s.Scan(content)
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// RedactPII removes personally identifiable information from content
func RedactPII(content string) string {
	if len(content) == 0 {
		return content
	}

	scanner := NewScanner()
	findings := scanner.Scan(content)

	if len(findings) == 0 {
		return content
	}

	// Sort findings by position so overlapping redactions are deterministic
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Position < findings[j].Position
	})

	// Build result by iterating through content and replacing matches
	var result strings.Builder
	lastEnd := 0

	for i := 0; i < len(findings); i++ {
		f := findings[i]
		// Skip if this finding overlaps with previous
		if f.Position < lastEnd {
			continue
		}

		// Write content before this finding
		if f.Position > lastEnd {
			result.WriteString(content[lastEnd:f.Position])
		}

		// Write redaction marker
		result.WriteString("[REDACTED]")
		lastEnd = f.Position + len(f.Match)
	}

	// Write remaining content
	if lastEnd < len(content) {
		result.WriteString(content[lastEnd:])
	}

	return result.String()
}

// RedactWithMask redacts sensitive data but keeps structure with mask character
func RedactWithMask(content string, mask rune) string {
	scanner := NewScanner()
	findings := scanner.Scan(content)

	if len(findings) == 0 {
		return content
	}

	var result strings.Builder
	lastEnd := 0

	for _, f := range findings {
		// Skip if this finding overlaps with previous
		if f.Position < lastEnd {
			continue
		}

		// Write content before this finding
		if f.Position > lastEnd {
			result.WriteString(content[lastEnd:f.Position])
		}

		// Create masked version keeping first and last 2 chars if possible
		var masked string
		if len(f.Match) > 8 {
			masked = f.Match[:2] + strings.Repeat(string(mask), len(f.Match)-4) + f.Match[len(f.Match)-2:]
		} else {
			masked = strings.Repeat(string(mask), len(f.Match))
		}

		result.WriteString(masked)
		lastEnd = f.Position + len(f.Match)
	}

	// Write remaining content
	if lastEnd < len(content) {
		result.WriteString(content[lastEnd:])
	}

	return result.String()
}

// ValidateContent checks if content is safe and returns findings
// Returns true if safe (no critical findings), false otherwise
func ValidateContent(content string) (bool, []Finding) {
	scanner := NewScanner()
	findings := scanner.Scan(content)

	// Content is considered safe if no critical findings
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return false, findings
		}
	}

	return true, findings
}

// SanitizeForLogging prepares content for logging by redacting PII and truncating if needed
func SanitizeForLogging(content string) string {
	const maxLength = 10000

	// First redact PII
	content = RedactPII(content)

	// Truncate if too long (respecting UTF-8 rune boundaries)
	if utf8.RuneCountInString(content) > maxLength {
		runes := []rune(content)
		content = string(runes[:maxLength]) + "\n[truncated]"
	}

	return content
}

var (
	suspiciousPatternsOnce sync.Once
	suspiciousPatterns     []*regexp.Regexp
)

func initSuspiciousPatterns() {
	suspiciousPatternsOnce.Do(func() {
		suspiciousPatterns = []*regexp.Regexp{
			// Shell injection
			regexp.MustCompile(`[;&|]\s*(rm|mv|cp|chmod|chown|sudo|su)\s`),
			regexp.MustCompile(`\$\(.*\)`),
			regexp.MustCompile("`.*`"),
			// SQL injection
			regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter)\s+.*--`),
			regexp.MustCompile(`(?i)union\s+select|select\s+\*\s+from`),
			// Path traversal
			regexp.MustCompile(`\.\./.*\.\./`),
			regexp.MustCompile(`\.\./.*etc/passwd`),
			// Null byte injection
			regexp.MustCompile(`\x00`),
			// XSS
			regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`),
			regexp.MustCompile(`(?i)javascript\s*:`),
			regexp.MustCompile(`(?i)on\w+\s*=\s*["']?[^"'>]+`),
		}
	})
}

// IsSuspiciousContent checks for potentially malicious content patterns
func IsSuspiciousContent(content string) bool {
	initSuspiciousPatterns()
	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
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

// hiddenUnicodeSet contains invisible/zero-width Unicode characters.
var hiddenUnicodeSet = map[rune]bool{
	'\u200B': true, // Zero-width space
	'\u200C': true, // Zero-width non-joiner
	'\u200D': true, // Zero-width joiner
	'\uFEFF': true, // Byte order mark (zero-width no-break space)
	'\u2060': true, // Word joiner
	'\u180E': true, // Mongolian vowel separator
	'\u200E': true, // Left-to-right mark
	'\u200F': true, // Right-to-left mark
}

// HasHiddenUnicode checks for hidden/invisible Unicode characters
func HasHiddenUnicode(s string) bool {
	for _, r := range s {
		if hiddenUnicodeSet[r] {
			return true
		}
	}
	return false
}

// ValidateUTF8 checks if content is valid UTF-8
func ValidateUTF8(content string) bool {
	return utf8.ValidString(content)
}

// Validator provides input validation for API requests
type Validator struct{}

// NewValidator creates a new input validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidatePreset validates a compression preset value
func (v *Validator) ValidatePreset(preset string) error {
	validPresets := map[string]bool{
		"fast":     true,
		"balanced": true,
		"full":     true,
		"":         true, // empty is valid (uses default)
	}
	if !validPresets[preset] {
		return fmt.Errorf("invalid preset: %s (must be 'fast', 'balanced', or 'full')", preset)
	}
	return nil
}

// ValidateMode validates a compression mode value
func (v *Validator) ValidateMode(mode string) error {
	validModes := map[string]bool{
		"minimal":    true,
		"aggressive": true,
		"none":       true,
		"":           true, // empty is valid
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid mode: %s (must be 'minimal', 'aggressive', or 'none')", mode)
	}
	return nil
}

// ValidateBudget validates a token budget value
func (v *Validator) ValidateBudget(budget int) error {
	if budget < 0 {
		return fmt.Errorf("budget must be non-negative, got %d", budget)
	}
	if budget > 10000000 { // 10M tokens max
		return fmt.Errorf("budget exceeds maximum of 10,000,000 tokens")
	}
	return nil
}

// ValidatePath validates a file path for security
func (v *Validator) ValidatePath(path string) error {
	if path == "" {
		return nil // empty path is valid (uses default)
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains invalid sequence '..'")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	return nil
}
