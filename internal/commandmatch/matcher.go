package commandmatch

import (
	"log"
	"regexp"
	"strings"
)

type CommandRule struct {
	Name     string `json:"name"`
	Basename string `json:"basename"`
	Pattern  string `json:"pattern"`
	compiled *regexp.Regexp
}

type CommandMatcher struct {
	rules []CommandRule
}

func NewCommandMatcher() *CommandMatcher {
	return &CommandMatcher{
		rules: []CommandRule{
			{Name: "git_status", Basename: "git", Pattern: `^git\s+status`},
			{Name: "git_log", Basename: "git", Pattern: `^git\s+log`},
			{Name: "git_diff", Basename: "git", Pattern: `^git\s+diff`},
			{Name: "docker_ps", Basename: "docker", Pattern: `^docker\s+ps`},
			{Name: "kubectl_pods", Basename: "kubectl", Pattern: `^kubectl\s+get\s+pods`},
			{Name: "npm_list", Basename: "npm", Pattern: `^npm\s+list`},
			{Name: "cargo_test", Basename: "cargo", Pattern: `^cargo\s+test`},
			{Name: "go_test", Basename: "go", Pattern: `^go\s+test`},
		},
	}
}

func (m *CommandMatcher) Match(command string) *CommandRule {
	trimmed := strings.TrimSpace(command)
	basename := strings.Fields(trimmed)
	if len(basename) == 0 {
		return nil
	}

	for i := range m.rules {
		rule := &m.rules[i]
		if rule.Basename != "" && basename[0] == rule.Basename {
			if rule.compiled == nil {
				if compiled, err := regexp.Compile(rule.Pattern); err != nil {
					log.Printf("failed to compile regex %q for %s: %v", rule.Pattern, rule.Basename, err)
				} else {
					rule.compiled = compiled
				}
			}
			if rule.compiled != nil && rule.compiled.MatchString(trimmed) {
				return rule
			}
		}
	}
	return nil
}

func (m *CommandMatcher) ExtractBasename(command string) string {
	fields := strings.Fields(command)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

func (m *CommandMatcher) StripFlags(command string) string {
	fields := strings.Fields(command)
	var result []string
	for _, f := range fields {
		if !strings.HasPrefix(f, "-") {
			result = append(result, f)
		}
	}
	return strings.Join(result, " ")
}

type BranchResult struct {
	Command string `json:"command"`
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type BranchingEngine struct {
	onSuccess map[string]string
	onFailure map[string]string
}

func NewBranchingEngine() *BranchingEngine {
	return &BranchingEngine{
		onSuccess: make(map[string]string),
		onFailure: make(map[string]string),
	}
}

func (e *BranchingEngine) AddSuccess(command, handler string) {
	e.onSuccess[command] = handler
}

func (e *BranchingEngine) AddFailure(command, handler string) {
	e.onFailure[command] = handler
}

func (e *BranchingEngine) Execute(result *BranchResult) string {
	if result.Success {
		if handler, ok := e.onSuccess[result.Command]; ok {
			return handler
		}
	} else {
		if handler, ok := e.onFailure[result.Command]; ok {
			return handler
		}
	}
	return result.Output
}

type MatchOutput struct {
	FullOutput string `json:"full_output"`
	Substring  string `json:"substring"`
	MatchCount int    `json:"match_count"`
	Matched    bool   `json:"matched"`
}

func MatchWholeOutput(output, substring string) *MatchOutput {
	count := strings.Count(output, substring)
	return &MatchOutput{
		FullOutput: output,
		Substring:  substring,
		MatchCount: count,
		Matched:    count > 0,
	}
}
