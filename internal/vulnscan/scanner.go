package vulnscan

import (
	"regexp"
	"strings"
)

type VulnType string

const (
	VulnSQLInjection  VulnType = "sql_injection"
	VulnSSRF          VulnType = "ssrf"
	VulnXSS           VulnType = "xss"
	VulnPathTraversal VulnType = "path_traversal"
	VulnCommandInject VulnType = "command_injection"
)

type VulnFinding struct {
	Type     VulnType `json:"type"`
	Match    string   `json:"match"`
	Severity int      `json:"severity"`
}

type VulnerabilityScanner struct {
	patterns map[VulnType]*regexp.Regexp
}

func NewVulnerabilityScanner() *VulnerabilityScanner {
	return &VulnerabilityScanner{
		patterns: map[VulnType]*regexp.Regexp{
			VulnSQLInjection:  regexp.MustCompile(`(?i)(union\s+select|or\s+1\s*=\s*1|drop\s+table|;\s*delete|;\s*update|'\s*or\s*')`),
			VulnSSRF:          regexp.MustCompile(`(?i)(file:///|http://127\.0\.0\.1|http://localhost|http://169\.254\.)`),
			VulnXSS:           regexp.MustCompile(`(?i)(<script|javascript:|onerror\s*=|onload\s*=)`),
			VulnPathTraversal: regexp.MustCompile(`(\.\.\/|\.\.\\|%2e%2e%2f|%252e%252e%252f)`),
			VulnCommandInject: regexp.MustCompile(`(?i)(;\s*rm\s|;\s*cat\s|;\s*wget\s|;\s*curl\s|\|\s*sh)`),
		},
	}
}

func (s *VulnerabilityScanner) Scan(input string) []VulnFinding {
	var findings []VulnFinding
	for vulnType, re := range s.patterns {
		if re.MatchString(input) {
			severity := 7
			if vulnType == VulnSQLInjection || vulnType == VulnCommandInject {
				severity = 10
			}
			findings = append(findings, VulnFinding{
				Type:     vulnType,
				Match:    re.FindString(input),
				Severity: severity,
			})
		}
	}
	return findings
}

func (s *VulnerabilityScanner) HasVulnerability(input string) bool {
	for _, re := range s.patterns {
		if re.MatchString(strings.ToLower(input)) {
			return true
		}
	}
	return false
}

func (s *VulnerabilityScanner) GetTypes() []VulnType {
	var types []VulnType
	for t := range s.patterns {
		types = append(types, t)
	}
	return types
}
