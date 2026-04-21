package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// SWEAdaptiveLoopFilter adds a lightweight self-adaptive prune loop inspired by
// SWE-Pruner style iterative scoring and retention feedback.
type SWEAdaptiveLoopFilter struct {
	passesMinimal    int
	passesAggressive int
	baseKeepRatio    float64
}

// NewSWEAdaptiveLoopFilter creates the adaptive loop filter.
func NewSWEAdaptiveLoopFilter() *SWEAdaptiveLoopFilter {
	return &SWEAdaptiveLoopFilter{
		passesMinimal:    2,
		passesAggressive: 3,
		baseKeepRatio:    0.80,
	}
}

// Name returns the filter name.
func (f *SWEAdaptiveLoopFilter) Name() string { return "40_swe_adaptive_loop" }

// Apply runs a small iterative pruning loop with progressively tighter budgets.
func (f *SWEAdaptiveLoopFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 10 {
		return input, 0
	}

	passes := f.passesMinimal
	if mode == ModeAggressive {
		passes = f.passesAggressive
	}

	current := lines
	for pass := 0; pass < passes; pass++ {
		ratio := f.baseKeepRatio - float64(pass)*0.12
		if mode == ModeAggressive {
			ratio -= 0.08
		}
		if ratio < 0.35 {
			ratio = 0.35
		}
		next := swePrunePass(current, ratio)
		if len(next) == len(current) {
			break
		}
		current = next
	}

	output := strings.Join(current, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func swePrunePass(lines []string, keepRatio float64) []string {
	if len(lines) == 0 {
		return lines
	}
	target := int(math.Ceil(float64(len(lines)) * keepRatio))
	if target < 4 {
		target = 4
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
			score += 1.0
		}
		if _, ok := detectRoleHeader(line); ok {
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

	out := make([]string, 0, len(keep))
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return lines
	}
	return out
}
