package litm

import "strings"

type LITMPosition struct {
	Content       string `json:"content"`
	OriginalSize  int    `json:"original_size"`
	OptimizedSize int    `json:"optimized_size"`
	Savings       int    `json:"savings"`
	Placement     string `json:"placement"`
}

type LITMOptimizer struct {
	maxTokens int
}

func NewLITMOptimizer(maxTokens int) *LITMOptimizer {
	if maxTokens == 0 {
		maxTokens = 4000
	}
	return &LITMOptimizer{maxTokens: maxTokens}
}

func (o *LITMOptimizer) Optimize(content string) *LITMPosition {
	lines := strings.Split(content, "\n")
	tokensPerLine := make([]int, len(lines))
	totalTokens := 0

	for i, line := range lines {
		tokensPerLine[i] = len(line) / 4
		totalTokens += tokensPerLine[i]
	}

	if totalTokens <= o.maxTokens {
		return &LITMPosition{
			Content:       content,
			OriginalSize:  totalTokens,
			OptimizedSize: totalTokens,
			Savings:       0,
			Placement:     "full",
		}
	}

	scored := make([]struct {
		index int
		score float64
	}, len(lines))

	for i, line := range lines {
		score := o.lineImportance(line)
		scored[i] = struct {
			index int
			score float64
		}{i, score}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	selected := make([]bool, len(lines))
	usedTokens := 0
	for _, s := range scored {
		if usedTokens+tokensPerLine[s.index] <= o.maxTokens {
			selected[s.index] = true
			usedTokens += tokensPerLine[s.index]
		}
	}

	var result []string
	for i, line := range lines {
		if selected[i] {
			result = append(result, line)
		}
	}

	optimized := strings.Join(result, "\n")
	return &LITMPosition{
		Content:       optimized,
		OriginalSize:  totalTokens,
		OptimizedSize: len(optimized) / 4,
		Savings:       totalTokens - len(optimized)/4,
		Placement:     "litm",
	}
}

func (o *LITMOptimizer) lineImportance(line string) float64 {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return 0
	}

	score := 1.0

	if strings.Contains(trimmed, "func ") || strings.Contains(trimmed, "def ") {
		score += 5
	}
	if strings.Contains(trimmed, "class ") || strings.Contains(trimmed, "type ") {
		score += 4
	}
	if strings.Contains(trimmed, "return ") || strings.Contains(trimmed, "yield ") {
		score += 3
	}
	if strings.Contains(trimmed, "if ") || strings.Contains(trimmed, "else") {
		score += 2
	}
	if strings.Contains(trimmed, "import ") || strings.Contains(trimmed, "from ") {
		score += 1
	}
	if strings.Contains(trimmed, "//") || strings.Contains(trimmed, "#") {
		score -= 1
	}
	if strings.Contains(trimmed, "log.") || strings.Contains(trimmed, "print(") {
		score -= 2
	}

	return score
}
