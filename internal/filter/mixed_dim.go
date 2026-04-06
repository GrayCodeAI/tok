package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "MixedDimKV: Beyond Token Eviction" — Miao et al., 2026
// https://arxiv.org/abs/2603.20616
// MixedDimFilter implements mixed-dimension token allocation — instead of
// evicting tokens entirely (0 or 100%), it reduces the "dimensionality" of
// less important tokens by abbreviating them.
type MixedDimFilter struct {
	baseRatio float64
}

// NewMixedDimFilter creates a new mixed-dimension allocation filter.
func NewMixedDimFilter() *MixedDimFilter {
	return &MixedDimFilter{baseRatio: 0.5}
}

// Apply reduces token dimensionality by abbreviating low-importance tokens.
func (f *MixedDimFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, line)
			continue
		}

		type wordScore struct {
			word  string
			score float64
			idx   int
		}
		scores := make([]wordScore, len(words))
		for i, w := range words {
			scores[i] = wordScore{word: w, score: f.wordImportance(w), idx: i}
		}

		sort.Slice(scores, func(a, b int) bool {
			return scores[a].score > scores[b].score
		})

		keepCount := int(math.Ceil(float64(len(scores)) * f.baseRatio))
		if mode == ModeAggressive {
			keepCount = int(math.Ceil(float64(len(scores)) * 0.3))
		}

		abbrev := make(map[int]bool)
		for i := keepCount; i < len(scores); i++ {
			abbrev[scores[i].idx] = true
		}

		var out []string
		for i, w := range words {
			if abbrev[i] {
				out = append(out, f.abbreviate(w))
			} else {
				out = append(out, w)
			}
		}
		result = append(result, strings.Join(out, " "))
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(original) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *MixedDimFilter) wordImportance(w string) float64 {
	lower := strings.ToLower(w)
	if strings.ContainsAny(w, "./\\_") {
		return 1.0
	}
	if numRe.MatchString(w) {
		return 0.9
	}
	for _, kw := range []string{"error", "fail", "func", "class", "type", "import",
		"return", "if", "else", "for", "struct", "interface"} {
		if lower == kw {
			return 0.8
		}
	}
	for _, kw := range []string{"the", "a", "an", "is", "are", "was", "were",
		"to", "of", "in", "on", "at", "by", "for", "with", "and", "or", "but"} {
		if lower == kw {
			return 0.1
		}
	}
	return 0.5
}

func (f *MixedDimFilter) abbreviate(w string) string {
	if len(w) <= 3 {
		return w
	}
	return string(w[0]) + string(w[1]) + ".." + string(w[len(w)-1])
}

// Name returns the layer name.
func (f *MixedDimFilter) Name() string { return "22_mixed_dim" }
