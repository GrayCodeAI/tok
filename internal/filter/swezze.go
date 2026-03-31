package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "SWEzze: Code Distillation for Issue Resolution" — Wang et al., PKU/UCL, 2026
// https://arxiv.org/abs/2603.28119
// SWEzzeFilter implements code distillation — extracts only "patch ingredients"
// (file paths, error types, function signatures) and discards surrounding context.
// Achieves 6x compression while improving issue resolution rates by 5-9%.
type SWEzzeFilter struct {
	minScore float64
}

// NewSWEzzeFilter creates a new SWEzze-style code distillation filter.
func NewSWEzzeFilter() *SWEzzeFilter {
	return &SWEzzeFilter{minScore: 0.3}
}

// Apply extracts minimal sufficient subsequence for code understanding.
func (f *SWEzzeFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	lines := strings.Split(input, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		score := f.scoreLine(line, inCodeBlock)
		if score >= f.minScore {
			result = append(result, line)
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
		}
	}

	if len(result) == 0 {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(original) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *SWEzzeFilter) scoreLine(line string, inCodeBlock bool) float64 {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return 0
	}

	score := 0.0

	if strings.HasPrefix(trimmed, "diff --git") || strings.HasPrefix(trimmed, "---") ||
		strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "@@") {
		return 1.0
	}

	if pathRe.MatchString(trimmed) {
		score += 0.8
	}

	lower := strings.ToLower(trimmed)
	for _, kw := range []string{"error", "fail", "panic", "exception", "fatal"} {
		if strings.Contains(lower, kw) {
			score += 0.9
		}
	}

	for _, kw := range []string{"func ", "class ", "def ", "fn ", "pub fn ", "impl ",
		"struct ", "interface ", "type ", "const ", "var ", "let "} {
		if strings.Contains(trimmed, kw) {
			score += 0.7
		}
	}

	if strings.HasPrefix(trimmed, "import") || strings.HasPrefix(trimmed, "from ") ||
		strings.HasPrefix(trimmed, "require") || strings.HasPrefix(trimmed, "use ") {
		score += 0.5
	}

	if inCodeBlock {
		score += 0.2
	}

	if numRe.MatchString(trimmed) {
		score += 0.1
	}

	return score
}

// Name returns the layer name.
func (f *SWEzzeFilter) Name() string { return "23_swezze" }

var (
	pathRe = regexp.MustCompile(`[\w./\\-]+\.(go|rs|py|ts|js|java|cpp|c|rb|toml|yaml|yml|json|xml|html|css|sql|sh|md|txt)`)
	numRe  = regexp.MustCompile(`\d+`)
)
