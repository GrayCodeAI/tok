package guardrails

import (
	"regexp"
	"strings"
)

type GuardrailType string

const (
	GuardrailRegex  GuardrailType = "regex"
	GuardrailPII    GuardrailType = "pii"
	GuardrailInject GuardrailType = "injection"
	GuardrailCustom GuardrailType = "custom"
)

type GuardrailRule struct {
	ID       string        `json:"id"`
	Type     GuardrailType `json:"type"`
	Pattern  string        `json:"pattern"`
	Action   string        `json:"action"`
	compiled *regexp.Regexp
}

type GuardrailEngine struct {
	requestRules  []GuardrailRule
	responseRules []GuardrailRule
}

func NewGuardrailEngine() *GuardrailEngine {
	return &GuardrailEngine{}
}

func (e *GuardrailEngine) AddRequestRule(rule GuardrailRule) {
	re, err := regexp.Compile(rule.Pattern)
	if err == nil {
		rule.compiled = re
	}
	e.requestRules = append(e.requestRules, rule)
}

func (e *GuardrailEngine) AddResponseRule(rule GuardrailRule) {
	re, err := regexp.Compile(rule.Pattern)
	if err == nil {
		rule.compiled = re
	}
	e.responseRules = append(e.responseRules, rule)
}

func (e *GuardrailEngine) CheckRequest(input string) []GuardrailRule {
	return e.checkRules(input, e.requestRules)
}

func (e *GuardrailEngine) CheckResponse(input string) []GuardrailRule {
	return e.checkRules(input, e.responseRules)
}

func (e *GuardrailEngine) checkRules(input string, rules []GuardrailRule) []GuardrailRule {
	var triggered []GuardrailRule
	for _, rule := range rules {
		if rule.compiled != nil && rule.compiled.MatchString(input) {
			triggered = append(triggered, rule)
		}
		if rule.Type == GuardrailPII && containsPII(input) {
			triggered = append(triggered, rule)
		}
	}
	return triggered
}

func containsPII(input string) bool {
	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	return emailRe.MatchString(input)
}

func (e *GuardrailEngine) RedactTriggered(input string, triggered []GuardrailRule) string {
	output := input
	for _, rule := range triggered {
		if rule.compiled != nil {
			output = rule.compiled.ReplaceAllString(output, "[BLOCKED:"+rule.ID+"]")
		}
	}
	return output
}

func (e *GuardrailEngine) ListRequestRules() []GuardrailRule {
	return e.requestRules
}

func (e *GuardrailEngine) ListResponseRules() []GuardrailRule {
	return e.responseRules
}

func (e *GuardrailEngine) StripRedundantLines(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
