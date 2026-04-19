package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// PlanBudgetFilter applies dynamic test-time budget allocation based on input difficulty.
type PlanBudgetFilter struct{}

// NewPlanBudgetFilter creates the dynamic budget controller filter.
func NewPlanBudgetFilter() *PlanBudgetFilter { return &PlanBudgetFilter{} }

// Name returns the filter name.
func (f *PlanBudgetFilter) Name() string { return "42_plan_budget" }

// Apply computes difficulty and keeps a matching budgeted subset of lines.
func (f *PlanBudgetFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 8 {
		return input, 0
	}

	difficulty := planBudgetDifficulty(lines)
	keepRatio := planBudgetKeepRatio(difficulty, mode)
	target := int(math.Ceil(float64(len(lines)) * keepRatio))
	if target < 3 {
		target = 3
	}

	type cand struct {
		idx   int
		score float64
	}
	cands := make([]cand, 0, len(lines))
	keep := make(map[int]bool, target)
	termFreq := daTermFrequency(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if i == 0 || i == len(lines)-1 || isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
			keep[i] = true
			continue
		}
		score := daLineScore(line, termFreq, len(lines))
		if isReasoningLine(line) || epicIsCausalEdge(line) {
			score += 0.8
		}
		if strings.Contains(strings.ToLower(line), "plan") || strings.Contains(strings.ToLower(line), "budget") {
			score += 0.6
		}
		cands = append(cands, cand{idx: i, score: score})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	for _, c := range cands {
		if len(keep) >= target {
			break
		}
		keep[c.idx] = true
	}

	out := make([]string, 0, len(keep)+1)
	out = append(out, "[plan-budget: difficulty="+planBudgetBucket(difficulty)+", keep="+itoa(int(keepRatio*100))+"%]")
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func planBudgetDifficulty(lines []string) float64 {
	score := 0.0
	for _, line := range lines {
		l := strings.ToLower(strings.TrimSpace(line))
		if l == "" {
			continue
		}
		if isErrorLine(line) || isWarningLine(line) {
			score += 1.4
		}
		if isCodeLine(line) {
			score += 1.0
		}
		if isReasoningLine(line) || epicIsCausalEdge(line) {
			score += 0.8
		}
		if strings.Contains(l, "stack") || strings.Contains(l, "trace") || strings.Contains(l, "migration") {
			score += 0.6
		}
	}
	norm := score / float64(len(lines))
	if norm > 1.0 {
		return 1.0
	}
	if norm < 0.0 {
		return 0.0
	}
	return norm
}

func planBudgetKeepRatio(difficulty float64, mode Mode) float64 {
	ratio := 0.35 + difficulty*0.30
	if mode == ModeAggressive {
		ratio -= 0.1
	}
	if ratio < 0.30 {
		ratio = 0.30
	}
	if ratio > 0.85 {
		ratio = 0.85
	}
	return ratio
}

func planBudgetBucket(difficulty float64) string {
	switch {
	case difficulty < 0.35:
		return "easy"
	case difficulty < 0.65:
		return "medium"
	default:
		return "hard"
	}
}
