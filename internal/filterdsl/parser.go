package filterdsl

import (
	"fmt"
	"regexp"
	"strings"
)

type RuleType string

const (
	RuleSkip      RuleType = "skip"
	RuleKeep      RuleType = "keep"
	RuleReplace   RuleType = "replace"
	RuleDedup     RuleType = "dedup"
	RuleSection   RuleType = "section"
	RuleAggregate RuleType = "aggregate"
	RuleChunk     RuleType = "chunk"
	RuleJSONPath  RuleType = "jsonpath"
	RuleTemplate  RuleType = "template"
)

type Rule struct {
	Type        RuleType
	Pattern     string
	Replacement string
	compiled    *regexp.Regexp
}

type SectionRule struct {
	Name  string
	Start string
	End   string
	Rules []Rule
}

type AggregateType string

const (
	AggCount AggregateType = "count"
	AggSum   AggregateType = "sum"
	AggAvg   AggregateType = "avg"
	AggFirst AggregateType = "first"
	AggLast  AggregateType = "last"
)

type AggregateRule struct {
	Type  AggregateType
	Field string
	Label string
}

type ChunkRule struct {
	Delimiter string
	MaxChunks int
}

type FilterDefinition struct {
	Name          string
	Match         string
	SkipRules     []Rule
	KeepRules     []Rule
	ReplaceRules  []Rule
	Dedup         bool
	Sections      []SectionRule
	Aggregates    []AggregateRule
	Chunks        []ChunkRule
	JSONPath      string
	Template      string
	Variables     map[string]string
	Conditions    []string
	compiledMatch *regexp.Regexp
}

type DSLParser struct{}

func NewDSLParser() *DSLParser {
	return &DSLParser{}
}

func (p *DSLParser) Parse(tomlContent string) (*FilterDefinition, error) {
	fd := &FilterDefinition{
		Variables: make(map[string]string),
	}

	lines := strings.Split(tomlContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")

		switch key {
		case "match":
			fd.Match = value
			re, err := regexp.Compile(value)
			if err == nil {
				fd.compiledMatch = re
			}
		case "name":
			fd.Name = value
		case "dedup":
			fd.Dedup = value == "true"
		case "jsonpath":
			fd.JSONPath = value
		case "template":
			fd.Template = value
		}

		if strings.HasPrefix(key, "skip.") {
			fd.SkipRules = append(fd.SkipRules, Rule{
				Type:    RuleSkip,
				Pattern: value,
			})
		}
		if strings.HasPrefix(key, "keep.") {
			fd.KeepRules = append(fd.KeepRules, Rule{
				Type:    RuleKeep,
				Pattern: value,
			})
		}
		if strings.HasPrefix(key, "replace.") {
			parts := strings.SplitN(value, "->", 2)
			rule := Rule{Type: RuleReplace}
			if len(parts) == 2 {
				rule.Pattern = strings.TrimSpace(parts[0])
				rule.Replacement = strings.TrimSpace(parts[1])
			} else {
				rule.Pattern = value
			}
			fd.ReplaceRules = append(fd.ReplaceRules, rule)
		}
		if strings.HasPrefix(key, "var.") {
			fd.Variables[key[4:]] = value
		}
	}

	for i := range fd.SkipRules {
		re, err := regexp.Compile(fd.SkipRules[i].Pattern)
		if err == nil {
			fd.SkipRules[i].compiled = re
		}
	}
	for i := range fd.KeepRules {
		re, err := regexp.Compile(fd.KeepRules[i].Pattern)
		if err == nil {
			fd.KeepRules[i].compiled = re
		}
	}
	for i := range fd.ReplaceRules {
		re, err := regexp.Compile(fd.ReplaceRules[i].Pattern)
		if err == nil {
			fd.ReplaceRules[i].compiled = re
		}
	}

	return fd, nil
}

func (fd *FilterDefinition) Matches(command string) bool {
	if fd.compiledMatch == nil {
		return false
	}
	return fd.compiledMatch.MatchString(command)
}

func (fd *FilterDefinition) Apply(input string) string {
	output := input

	for _, rule := range fd.SkipRules {
		if rule.compiled != nil {
			output = rule.compiled.ReplaceAllString(output, "")
		}
	}

	for _, rule := range fd.ReplaceRules {
		if rule.compiled != nil {
			output = rule.compiled.ReplaceAllString(output, rule.Replacement)
		}
	}

	if fd.Dedup {
		output = dedupLines(output)
	}

	if fd.Template != "" {
		output = applyTemplate(output, fd.Template, fd.Variables)
	}

	return output
}

func dedupLines(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		normalized := strings.TrimSpace(strings.ToLower(line))
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func applyTemplate(input, template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	result = strings.ReplaceAll(result, "{{input}}", input)
	return result
}

type DSLValidator struct{}

func NewDSLValidator() *DSLValidator {
	return &DSLValidator{}
}

func (v *DSLValidator) Validate(fd *FilterDefinition) []string {
	var errors []string
	if fd.Match == "" {
		errors = append(errors, "match pattern is required")
	}
	if fd.Name == "" {
		errors = append(errors, "name is required")
	}
	if _, err := regexp.Compile(fd.Match); err != nil {
		errors = append(errors, fmt.Sprintf("invalid match pattern: %v", err))
	}
	for i, rule := range fd.SkipRules {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			errors = append(errors, fmt.Sprintf("invalid skip.%d pattern: %v", i, err))
		}
	}
	for i, rule := range fd.ReplaceRules {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			errors = append(errors, fmt.Sprintf("invalid replace.%d pattern: %v", i, err))
		}
	}
	return errors
}
