// Package security provides PII detection and redaction.
package security

import (
	"regexp"
	"sort"
	"strings"
)

// PIIPattern represents a detected PII pattern.
type PIIPattern struct {
	Type     string `json:"type"`
	Pattern  string `json:"pattern"`
	Position int    `json:"position"`
	Length   int    `json:"length"`
	Value    string `json:"value,omitempty"` // Redacted in output
}

// PIIDetector detects and redacts PII.
type PIIDetector struct {
	patterns map[string]*regexp.Regexp
	enabled  map[string]bool
}

// NewPIIDetector creates a new PII detector with default patterns.
func NewPIIDetector() *PIIDetector {
	d := &PIIDetector{
		patterns: make(map[string]*regexp.Regexp),
		enabled:  make(map[string]bool),
	}
	d.initPatterns()
	return d
}

// Enable enables a pattern type.
func (d *PIIDetector) Enable(patternType string) {
	d.enabled[patternType] = true
}

// Disable disables a pattern type.
func (d *PIIDetector) Disable(patternType string) {
	d.enabled[patternType] = false
}

// IsEnabled checks if a pattern type is enabled.
func (d *PIIDetector) IsEnabled(patternType string) bool {
	return d.enabled[patternType]
}

func (d *PIIDetector) initPatterns() {
	// Email addresses
	d.patterns["email"] = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	d.enabled["email"] = true

	// Phone numbers (US format)
	d.patterns["phone_us"] = regexp.MustCompile(`\b(\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`)
	d.enabled["phone_us"] = true

	// SSN
	d.patterns["ssn"] = regexp.MustCompile(`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`)
	d.enabled["ssn"] = true

	// Credit cards
	d.patterns["credit_card"] = regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3[0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`)
	d.enabled["credit_card"] = true

	// IP addresses
	d.patterns["ip_address"] = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	d.enabled["ip_address"] = true

	// API keys (common patterns)
	d.patterns["api_key"] = regexp.MustCompile(`\b(?:api[_-]?key|apikey|key)["\s]*[:=]["\s]*[a-zA-Z0-9_\-]{16,}\b`)
	d.enabled["api_key"] = true

	// Passwords in URLs or config
	d.patterns["password"] = regexp.MustCompile(`(?i)(?:password|passwd|pwd)["\s]*[:=]["\s]*[^\s"]{4,}`)
	d.enabled["password"] = true

	// AWS keys
	d.patterns["aws_key"] = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	d.enabled["aws_key"] = true

	// GitHub tokens
	d.patterns["github_token"] = regexp.MustCompile(`\bghp_[a-zA-Z0-9]{36}\b`)
	d.enabled["github_token"] = true

	// Slack tokens
	d.patterns["slack_token"] = regexp.MustCompile(`\bxox[baprs]-[a-zA-Z0-9]{10,48}\b`)
	d.enabled["slack_token"] = true

	// Private keys
	d.patterns["private_key"] = regexp.MustCompile(`-----BEGIN (?:RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`)
	d.enabled["private_key"] = true
}

// Detect finds all PII in content.
func (d *PIIDetector) Detect(content string) []PIIPattern {
	var findings []PIIPattern

	for patternType, re := range d.patterns {
		if !d.enabled[patternType] {
			continue
		}

		matches := re.FindAllStringIndex(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				findings = append(findings, PIIPattern{
					Type:     patternType,
					Pattern:  re.String(),
					Position: match[0],
					Length:   match[1] - match[0],
					Value:    content[match[0]:match[1]],
				})
			}
		}
	}

	return findings
}

// Redact replaces PII with placeholders.
func (d *PIIDetector) Redact(content string) string {
	findings := d.Detect(content)
	if len(findings) == 0 {
		return content
	}

	// Sort by position (descending) to replace from end to start
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Position > findings[j].Position
	})

	result := content
	for _, finding := range findings {
		placeholder := "[" + strings.ToUpper(finding.Type) + "]"
		result = result[:finding.Position] + placeholder + result[finding.Position+finding.Length:]
	}

	return result
}

// RedactWithOptions redacts with custom options.
func (d *PIIDetector) RedactWithOptions(content string, opts RedactOptions) string {
	findings := d.Detect(content)
	if len(findings) == 0 {
		return content
	}

	// Filter by type if specified
	if len(opts.Types) > 0 {
		var filtered []PIIPattern
		typeSet := make(map[string]bool)
		for _, t := range opts.Types {
			typeSet[t] = true
		}
		for _, f := range findings {
			if typeSet[f.Type] {
				filtered = append(filtered, f)
			}
		}
		findings = filtered
	}

	if len(findings) == 0 {
		return content
	}

	// Sort by position (descending)
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Position > findings[j].Position
	})

	result := content
	for _, finding := range findings {
		var placeholder string
		if opts.Mask {
			// Partial masking: show first 2 and last 2 chars
			value := finding.Value
			if len(value) > 8 {
				placeholder = value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
			} else {
				placeholder = strings.Repeat("*", len(value))
			}
		} else {
			placeholder = "[" + strings.ToUpper(finding.Type) + "]"
		}

		if opts.Replacement != "" {
			placeholder = opts.Replacement
		}

		result = result[:finding.Position] + placeholder + result[finding.Position+finding.Length:]
	}

	return result
}

// RedactOptions provides redaction options.
type RedactOptions struct {
	Types       []string // Specific types to redact
	Mask        bool     // Partial masking instead of full replacement
	Replacement string   // Custom replacement string
}

// HasPII checks if content contains PII.
func (d *PIIDetector) HasPII(content string) bool {
	return len(d.Detect(content)) > 0
}

// PIIScanResult contains PII scan results.
type PIIScanResult struct {
	HasPII   bool         `json:"has_pii"`
	Findings []PIIPattern `json:"findings"`
	Redacted string       `json:"redacted,omitempty"`
}

// Scan scans content and returns detailed results.
func (d *PIIDetector) Scan(content string) PIIScanResult {
	findings := d.Detect(content)
	return PIIScanResult{
		HasPII:   len(findings) > 0,
		Findings: findings,
		Redacted: d.Redact(content),
	}
}

// AddPattern adds a custom pattern.
func (d *PIIDetector) AddPattern(name, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	d.patterns[name] = re
	d.enabled[name] = true
	return nil
}

// RemovePattern removes a pattern.
func (d *PIIDetector) RemovePattern(name string) {
	delete(d.patterns, name)
	delete(d.enabled, name)
}

// GetPatternNames returns all pattern names.
func (d *PIIDetector) GetPatternNames() []string {
	names := make([]string, 0, len(d.patterns))
	for name := range d.patterns {
		names = append(names, name)
	}
	return names
}
