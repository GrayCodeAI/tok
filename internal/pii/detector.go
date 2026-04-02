package pii

import (
	"regexp"
	"strings"
)

type PIIFinding struct {
	Type     string `json:"type"`
	Match    string `json:"match"`
	Position int    `json:"position"`
	Redacted string `json:"redacted"`
}

type PIIDetector struct {
	patterns map[string]*regexp.Regexp
}

func NewPIIDetector() *PIIDetector {
	return &PIIDetector{
		patterns: map[string]*regexp.Regexp{
			"email":       regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
			"phone":       regexp.MustCompile(`(\+?1?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})`),
			"ssn":         regexp.MustCompile(`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`),
			"credit_card": regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
			"ip_address":  regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			"url":         regexp.MustCompile(`https?://[^\s<>{}|\\^` + "`" + `\[\]]+`),
			"mac_address": regexp.MustCompile(`\b[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}\b`),
		},
	}
}

func (d *PIIDetector) Detect(input string) []PIIFinding {
	var findings []PIIFinding
	for piiType, re := range d.patterns {
		matches := re.FindAllStringIndex(input, -1)
		for _, m := range matches {
			match := input[m[0]:m[1]]
			findings = append(findings, PIIFinding{
				Type:     piiType,
				Match:    match,
				Position: m[0],
				Redacted: d.redact(piiType, match),
			})
		}
	}
	return findings
}

func (d *PIIDetector) Redact(input string) string {
	output := input
	for piiType, re := range d.patterns {
		output = re.ReplaceAllStringFunc(output, func(match string) string {
			return d.redact(piiType, match)
		})
	}
	return output
}

func (d *PIIDetector) redact(piiType, match string) string {
	switch piiType {
	case "email":
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			return "[REDACTED_EMAIL]@" + parts[1]
		}
		return "[REDACTED_EMAIL]"
	case "phone":
		return "[REDACTED_PHONE]"
	case "ssn":
		return "[REDACTED_SSN]"
	case "credit_card":
		last4 := match[len(match)-4:]
		return "[REDACTED_CC_****" + last4 + "]"
	case "ip_address":
		return "[REDACTED_IP]"
	case "url":
		return "[REDACTED_URL]"
	case "mac_address":
		return "[REDACTED_MAC]"
	default:
		return "[REDACTED]"
	}
}

func (d *PIIDetector) HasPII(input string) bool {
	for _, re := range d.patterns {
		if re.MatchString(input) {
			return true
		}
	}
	return false
}
