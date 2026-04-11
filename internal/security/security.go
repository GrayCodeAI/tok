package security

import (
	"regexp"
	"strings"
)

// Finding represents a security issue found in content
type Finding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Position int    `json:"position"`
}

// Severity levels
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Scanner detects security issues in text content
type Scanner struct {
	rules []scanRule
}

// scanRule defines a single security scanning rule
type scanRule struct {
	name        string
	severity    string
	pattern     *regexp.Regexp
	message     string
	maxMatches  int // Limit matches to prevent DoS
}

// NewScanner creates a new security scanner with default rules
func NewScanner() *Scanner {
	return &Scanner{
		rules: defaultScanRules(),
	}
}

// defaultScanRules returns the default set of security scanning rules
func defaultScanRules() []scanRule {
	return []scanRule{
		// API Keys and Tokens
		{
			name:       "aws_access_key",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			message:    "AWS Access Key ID detected",
			maxMatches: 10,
		},
		{
			name:       "aws_secret_key",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`["'][0-9a-zA-Z/+]{40}["']`),
			message:    "Potential AWS Secret Access Key detected",
			maxMatches: 10,
		},
		{
			name:       "github_token",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
			message:    "GitHub token detected",
			maxMatches: 10,
		},
		{
			name:       "slack_token",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z-]+`),
			message:    "Slack token detected",
			maxMatches: 10,
		},
		{
			name:       "generic_api_key",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret)[\s]*[=:]+[\s]*["'][a-z0-9_-]{16,}["']`),
			message:    "Generic API key detected",
			maxMatches: 10,
		},
		// Private Keys
		{
			name:       "private_key",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`-----BEGIN (RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`),
			message:    "Private key detected",
			maxMatches: 5,
		},
		{
			name:       "ssh_private_key",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`-----BEGIN OPENSSH PRIVATE KEY-----`),
			message:    "SSH private key detected",
			maxMatches: 5,
		},
		// PII - Credit Cards
		{
			name:       "credit_card_visa",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`\b4[0-9]{12}(?:[0-9]{3})?\b`),
			message:    "Visa credit card number detected",
			maxMatches: 10,
		},
		{
			name:       "credit_card_mastercard",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`\b5[1-5][0-9]{14}\b`),
			message:    "Mastercard number detected",
			maxMatches: 10,
		},
		{
			name:       "credit_card_amex",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`\b3[47][0-9]{13}\b`),
			message:    "American Express card number detected",
			maxMatches: 10,
		},
		// PII - Social Security Numbers
		{
			name:       "ssn_us",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			message:    "US Social Security Number detected",
			maxMatches: 10,
		},
		// PII - Email addresses
		{
			name:       "email_address",
			severity:   SeverityMedium,
			pattern:    regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			message:    "Email address detected",
			maxMatches: 20,
		},
		// Database Connection Strings
		{
			name:       "db_connection_string",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`(?i)(mongodb|postgres|mysql)://[^:]+:[^@]+@`),
			message:    "Database connection string with credentials detected",
			maxMatches: 5,
		},
		// JWT Tokens
		{
			name:       "jwt_token",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`),
			message:    "JWT token detected",
			maxMatches: 10,
		},
		// Passwords in URLs
		{
			name:       "password_in_url",
			severity:   SeverityCritical,
			pattern:    regexp.MustCompile(`[a-zA-Z]+://[^:]+:[^@]+@`),
			message:    "Password in URL detected",
			maxMatches: 10,
		},
		// Authorization Bearer tokens
		{
			name:       "bearer_token",
			severity:   SeverityHigh,
			pattern:    regexp.MustCompile(`(?i)bearer\s+[a-z0-9_\-\.]+`),
			message:    "Bearer authorization token detected",
			maxMatches: 10,
		},
	}
}

// Scan analyzes content for security issues and returns findings
func (s *Scanner) Scan(content string) []Finding {
	if content == "" {
		return nil
	}

	var findings []Finding
	scannedPositions := make(map[int]bool) // Track scanned positions to avoid duplicates

	for _, rule := range s.rules {
		matches := rule.pattern.FindAllStringIndex(content, rule.maxMatches)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			pos := match[0]
			// Skip if we've already found something at this position
			if scannedPositions[pos] {
				continue
			}
			scannedPositions[pos] = true

			findings = append(findings, Finding{
				Rule:     rule.name,
				Severity: rule.severity,
				Message:  rule.message,
				Position: pos,
			})
		}
	}

	return findings
}

// ScanWithRedaction scans content and returns redacted version with findings
func (s *Scanner) ScanWithRedaction(content string) (redacted string, findings []Finding) {
	findings = s.Scan(content)
	if len(findings) == 0 {
		return content, nil
	}

	redacted = RedactPII(content)
	return redacted, findings
}

// HasCriticalFindings checks if any finding is critical severity
func (s *Scanner) HasCriticalFindings(content string) bool {
	findings := s.Scan(content)
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// redactionPatterns for PII redaction
var redactionPatterns = []*regexp.Regexp{
	// API Keys
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
	regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z-]+`),
	// Private Keys - multiline
	regexp.MustCompile(`(?s)-----BEGIN (RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----.*?-----END (RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`),
	// Database URLs with passwords
	regexp.MustCompile(`(?i)(mongodb|postgres|mysql)://[^:]+:[^@]+@`),
	// JWT Tokens
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`),
	// Passwords in URLs
	regexp.MustCompile(`([a-zA-Z]+://[^:]+):[^@]+(@)`),
}

// RedactPII removes sensitive information from content
func RedactPII(content string) string {
	if content == "" {
		return content
	}

	redacted := content
	for _, pattern := range redactionPatterns {
		redacted = pattern.ReplaceAllString(redacted, "[REDACTED]")
	}

	// Redact potential AWS secret keys (40 char base64 strings following "aws")
	redacted = regexp.MustCompile(`(?i)(aws[_\-]?(secret)?[_\-]?access[_\-]?key[_\-]?(id)?[\s]*[=:]+[\s]*["'])[a-zA-Z0-9/+]{40}["']`).
		ReplaceAllString(redacted, "${1}[REDACTED]\"")

	// Redact generic secrets
	redacted = regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret[_-]?key|token)[\s]*[=:]+[\s]*["'][a-z0-9]{16,}["']`).
		ReplaceAllString(redacted, "${1}=[REDACTED]")

	return redacted
}

// RedactWithMask replaces sensitive data with a mask character
func RedactWithMask(content string, mask rune) string {
	if content == "" {
		return content
	}

	findings := NewScanner().Scan(content)
	if len(findings) == 0 {
		return content
	}

	// Convert to runes for proper Unicode handling
	runes := []rune(content)
	redacted := make([]rune, len(runes))
	copy(redacted, runes)

	// Mark positions for redaction
	redactPositions := make(map[int]bool)
	for _, f := range findings {
		// Simple approach: redact 20 chars from position
		for i := f.Position; i < f.Position+20 && i < len(runes); i++ {
			redactPositions[i] = true
		}
	}

	// Apply redaction
	for pos := range redactPositions {
		if pos < len(redacted) {
			redacted[pos] = mask
		}
	}

	return string(redacted)
}

// ValidateContent checks if content is safe (no critical findings)
func ValidateContent(content string) (bool, []Finding) {
	scanner := NewScanner()
	findings := scanner.Scan(content)

	for _, f := range findings {
		if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
			return false, findings
		}
	}
	return true, findings
}

// SanitizeForLogging prepares content for logging by removing sensitive data
func SanitizeForLogging(content string) string {
	if len(content) > 10000 {
		// For large content, only scan first and last 5KB
		prefix := content[:5000]
		suffix := content[len(content)-5000:]
		return RedactPII(prefix) + "\n...[truncated]...\n" + RedactPII(suffix)
	}
	return RedactPII(content)
}

// IsSuspiciousContent checks for potentially malicious patterns
func IsSuspiciousContent(content string) bool {
	suspiciousPatterns := []*regexp.Regexp{
		// Shell injection attempts
		regexp.MustCompile(`[;&|\x60]\s*(rm|curl|wget|bash|sh|exec|eval|system)\s*`),
		// SQL injection patterns
		regexp.MustCompile(`(?i)(union\s+select|insert\s+into|delete\s+from|drop\s+table)`),
		// Path traversal
		regexp.MustCompile(`\.\./|\.\.\\|%2e%2e%2f|%252e%252e%252f`),
		// Null byte injection
		regexp.MustCompile(`\x00`),
		// Script tags
		regexp.MustCompile(`(?i)<script[^>]*>|javascript:`),
	}

	contentLower := strings.ToLower(content)
	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(contentLower) {
			return true
		}
	}
	return false
}
