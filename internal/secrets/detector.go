package secrets

import (
	"regexp"
	"strings"
)

type SecretType string

const (
	SecretAPIKey     SecretType = "api_key"
	SecretToken      SecretType = "token"
	SecretPassword   SecretType = "password"
	SecretPrivateKey SecretType = "private_key"
	SecretJWT        SecretType = "jwt"
	SecretAWS        SecretType = "aws_key"
	SecretGitHub     SecretType = "github_token"
	SecretSlack      SecretType = "slack_token"
)

type SecretFinding struct {
	Type     SecretType `json:"type"`
	Match    string     `json:"match"`
	Line     int        `json:"line"`
	Severity int        `json:"severity"`
}

type SecretsDetector struct {
	patterns map[SecretType]*regexp.Regexp
}

func NewSecretsDetector() *SecretsDetector {
	return &SecretsDetector{
		patterns: map[SecretType]*regexp.Regexp{
			SecretAPIKey:     regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{20,})['"]?`),
			SecretToken:      regexp.MustCompile(`(?i)(token|bearer)\s*[:=]\s*['"]?([a-zA-Z0-9_\-\.]{20,})['"]?`),
			SecretPassword:   regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*['"]?([^\s'"]{8,})['"]?`),
			SecretPrivateKey: regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`),
			SecretJWT:        regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
			SecretAWS:        regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
			SecretGitHub:     regexp.MustCompile(`\b(ghp|gho|ghu|ghs|ghr)_[a-zA-Z0-9]{36}\b`),
			SecretSlack:      regexp.MustCompile(`\bxox[bpsa]-[a-zA-Z0-9-]+\b`),
		},
	}
}

func (d *SecretsDetector) Detect(input string) []SecretFinding {
	var findings []SecretFinding
	lines := strings.Split(input, "\n")
	for lineNum, line := range lines {
		for secretType, re := range d.patterns {
			if re.MatchString(line) {
				severity := 7
				if secretType == SecretPrivateKey || secretType == SecretAWS {
					severity = 10
				} else if secretType == SecretJWT {
					severity = 9
				}
				findings = append(findings, SecretFinding{
					Type:     secretType,
					Match:    re.FindString(line),
					Line:     lineNum + 1,
					Severity: severity,
				})
			}
		}
	}
	return findings
}

func (d *SecretsDetector) HasSecrets(input string) bool {
	for _, re := range d.patterns {
		if re.MatchString(input) {
			return true
		}
	}
	return false
}

func (d *SecretsDetector) Redact(input string) string {
	output := input
	for _, re := range d.patterns {
		output = re.ReplaceAllString(output, "[REDACTED_SECRET]")
	}
	return output
}
