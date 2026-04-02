package tampstages

import (
	"regexp"
	"strings"
)

type TampStage struct{}

func NewTampStage() *TampStage {
	return &TampStage{}
}

func (s *TampStage) Prune(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" || trimmed == "..." {
			continue
		}
		if strings.Contains(trimmed, "resolved") || strings.Contains(trimmed, "integrity") ||
			strings.Contains(trimmed, "shasum") || strings.Contains(trimmed, "tarball") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func (s *TampStage) StripComments(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func (s *TampStage) TextPress(content string) string {
	if len(content) < 100 {
		return content
	}
	lines := strings.Split(content, "\n")
	var important []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 20 && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "//") {
			important = append(important, trimmed)
		}
	}
	if len(important) > 10 {
		important = important[:10]
	}
	return strings.Join(important, "\n")
}

func (s *TampStage) CacheSafe(content string) string {
	return content
}

func (s *TampStage) MaxBodyCheck(content string, maxSize int) bool {
	return len(content) <= maxSize
}

func (s *TampStage) LogRequest(content string) string {
	lines := strings.Split(content, "\n")
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

type CommandMatcher struct{}

func NewCommandMatcher() *CommandMatcher {
	return &CommandMatcher{}
}

func (m *CommandMatcher) Match(command, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(command)
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

type BranchEngine struct {
	onSuccess map[string]string
	onFailure map[string]string
}

func NewBranchEngine() *BranchEngine {
	return &BranchEngine{
		onSuccess: make(map[string]string),
		onFailure: make(map[string]string),
	}
}

func (e *BranchEngine) AddSuccess(command, handler string) {
	e.onSuccess[command] = handler
}

func (e *BranchEngine) AddFailure(command, handler string) {
	e.onFailure[command] = handler
}

func (e *BranchEngine) Execute(command string, success bool, output string) string {
	if success {
		if handler, ok := e.onSuccess[command]; ok {
			return handler
		}
	} else {
		if handler, ok := e.onFailure[command]; ok {
			return handler
		}
	}
	return output
}

func (e *BranchEngine) MatchOutput(output, substring string) bool {
	return strings.Contains(output, substring)
}
