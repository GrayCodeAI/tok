package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// GraphCoTFilter keeps high-centrality reasoning lines in long traces.
type GraphCoTFilter struct {
	targetRatio float64
}

// NewGraphCoTFilter creates a graph-CoT style filter.
func NewGraphCoTFilter() *GraphCoTFilter {
	return &GraphCoTFilter{targetRatio: 0.55}
}

// Name returns the filter name.
func (f *GraphCoTFilter) Name() string { return "38_graph_cot" }

// Apply scores reasoning lines and keeps high-centrality nodes.
func (f *GraphCoTFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 8 || !epicLooksLikeCoT(lines) {
		return input, 0
	}

	ratio := f.targetRatio
	if mode == ModeAggressive {
		ratio = 0.40
	}
	target := int(math.Ceil(float64(len(lines)) * ratio))
	if target < 2 {
		target = 2
	}

	termFreq := daTermFrequency(lines)
	type cand struct {
		idx   int
		score float64
	}
	cands := make([]cand, 0, len(lines))
	for i, line := range lines {
		cands = append(cands, cand{idx: i, score: graphLineScore(line, termFreq, len(lines))})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })

	keep := make(map[int]bool, target)
	for i := 0; i < len(cands) && len(keep) < target; i++ {
		keep[cands[i].idx] = true
	}
	// Anchors
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			keep[i] = true
			break
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			keep[i] = true
			break
		}
	}

	var out []string
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}

	if len(out) >= len(lines) {
		return input, 0
	}
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func graphLineScore(line string, termFreq map[string]int, nLines int) float64 {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return 0
	}
	score := 0.0
	if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
		score += 3.0
	}
	if epicIsCausalEdge(trimmed) || isReasoningLine(line) {
		score += 1.5
	}
	terms := ltTokenize(line)
	if len(terms) > 0 {
		rare := 0.0
		for _, t := range terms {
			f := termFreq[t]
			if f <= 0 {
				f = 1
			}
			rare += 1.0 / float64(f)
		}
		score += rare * (float64(nLines) / 10.0) / float64(len(terms))
	}
	return score
}
