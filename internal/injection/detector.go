package injection

import (
	"regexp"
	"strings"
)

type InjectionType string

const (
	InjectionDirect    InjectionType = "direct"
	InjectionIndirect  InjectionType = "indirect"
	InjectionJailbreak InjectionType = "jailbreak"
)

type InjectionFinding struct {
	Type     InjectionType `json:"type"`
	Match    string        `json:"match"`
	Severity int           `json:"severity"`
}

type InjectionDetector struct {
	patterns map[InjectionType][]*regexp.Regexp
}

func NewInjectionDetector() *InjectionDetector {
	return &InjectionDetector{
		patterns: map[InjectionType][]*regexp.Regexp{
			InjectionDirect: {
				regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions`),
				regexp.MustCompile(`(?i)disregard\s+(all\s+)?prior\s+instructions`),
				regexp.MustCompile(`(?i)you\s+are\s+now\s+a`),
				regexp.MustCompile(`(?i)forget\s+(all\s+)?previous\s+prompts`),
				regexp.MustCompile(`(?i)ignore\s+(all\s+)?rules`),
				regexp.MustCompile(`(?i)override\s+(your|the)\s+(rules|instructions|guidelines)`),
				regexp.MustCompile(`(?i)system\s+prompt\s*[:=]`),
				regexp.MustCompile(`(?i)prompt\s+injection`),
				regexp.MustCompile(`(?i)act\s+as\s+(an?|the)\s+new`),
			},
			InjectionIndirect: {
				regexp.MustCompile(`(?i)ignore\s+above`),
				regexp.MustCompile(`(?i)ignore\s+everything\s+above`),
				regexp.MustCompile(`(?i)stop\s+being`),
				regexp.MustCompile(`(?i)pretend\s+(to\s+be|you\s+are)`),
				regexp.MustCompile(`(?i)role\s*play`),
				regexp.MustCompile(`(?i)simulation\s+mode`),
				regexp.MustCompile(`(?i)developer\s+mode`),
				regexp.MustCompile(`(?i)dan\s+mode`),
			},
			InjectionJailbreak: {
				regexp.MustCompile(`(?i)do\s+anything\s+now`),
				regexp.MustCompile(`(?i)jailbreak`),
				regexp.MustCompile(`(?i)uncensored`),
				regexp.MustCompile(`(?i)without\s+restrictions`),
				regexp.MustCompile(`(?i)no\s+content\s+policy`),
				regexp.MustCompile(`(?i)disable\s+safety`),
				regexp.MustCompile(`(?i)bypass\s+(all\s+)?filters`),
			},
		},
	}
}

func (d *InjectionDetector) Detect(input string) []InjectionFinding {
	var findings []InjectionFinding
	lower := strings.ToLower(input)
	for injType, patterns := range d.patterns {
		for _, re := range patterns {
			if re.MatchString(lower) {
				severity := 5
				if injType == InjectionDirect {
					severity = 9
				} else if injType == InjectionJailbreak {
					severity = 10
				}
				findings = append(findings, InjectionFinding{
					Type:     injType,
					Match:    re.FindString(lower),
					Severity: severity,
				})
			}
		}
	}
	return findings
}

func (d *InjectionDetector) HasInjection(input string) bool {
	for _, patterns := range d.patterns {
		for _, re := range patterns {
			if re.MatchString(strings.ToLower(input)) {
				return true
			}
		}
	}
	return false
}
