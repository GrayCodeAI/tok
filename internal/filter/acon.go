package filter

import (
	"math"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "ACON: Optimizing Context Compression for Long-Context LLMs" — ICLR 2026
// ACONFilter implements adaptive context optimization — dynamically adjusts
// compression based on content complexity and context length.
type ACONFilter struct {
	targetRatio float64
}

// NewACONFilter creates a new ACON-style context compression filter.
func NewACONFilter() *ACONFilter {
	return &ACONFilter{targetRatio: 0.6}
}

// Apply applies adaptive context compression.
func (f *ACONFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	tokens := core.EstimateTokens(input)

	complexity := f.complexityScore(input)

	targetRatio := f.targetRatio
	if complexity > 0.7 {
		targetRatio = 0.8
	} else if complexity < 0.3 {
		targetRatio = 0.4
	}

	if mode == ModeAggressive {
		targetRatio *= 0.7
	}

	keepCount := int(math.Ceil(float64(tokens) * targetRatio))

	lines := strings.Split(input, "\n")
	type lineInfo struct {
		line  string
		score float64
		idx   int
	}

	scored := make([]lineInfo, len(lines))
	for i, line := range lines {
		scored[i] = lineInfo{line: line, score: f.lineScore(line), idx: i}
	}

	for i := 1; i < len(scored); i++ {
		for j := i; j > 0 && scored[j].score > scored[j-1].score; j-- {
			scored[j], scored[j-1] = scored[j-1], scored[j]
		}
	}

	keepLines := keepCount / 10
	if keepLines < 1 {
		keepLines = 1
	}
	if keepLines > len(scored) {
		keepLines = len(scored)
	}

	kept := make(map[int]string)
	for i := 0; i < keepLines; i++ {
		kept[scored[i].idx] = scored[i].line
	}

	var result []string
	for i := 0; i < len(lines); i++ {
		if l, ok := kept[i]; ok {
			result = append(result, l)
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

func (f *ACONFilter) complexityScore(input string) float64 {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return 0
	}

	uniqueWords := make(map[string]bool)
	totalWords := 0
	for _, line := range lines {
		words := strings.Fields(line)
		for _, w := range words {
			uniqueWords[strings.ToLower(w)] = true
			totalWords++
		}
	}

	if totalWords == 0 {
		return 0
	}

	return float64(len(uniqueWords)) / float64(totalWords)
}

func (f *ACONFilter) lineScore(line string) float64 {
	score := 0.0
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return 0
	}

	if strings.ContainsAny(trimmed, "{}[]()") {
		score += 2.0
	}

	lower := strings.ToLower(trimmed)
	for _, kw := range []string{"error", "fail", "panic", "func", "class", "type",
		"import", "return", "struct", "interface"} {
		if strings.Contains(lower, kw) {
			score += 1.5
		}
	}

	for _, c := range trimmed {
		if c >= '0' && c <= '9' {
			score += 0.5
			break
		}
	}

	if strings.Contains(trimmed, "/") {
		score += 1.0
	}

	score += float64(len(trimmed)) / 100.0

	return score
}

// Name returns the layer name.
func (f *ACONFilter) Name() string { return "29_acon" }
