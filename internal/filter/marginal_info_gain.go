package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "COMI: Coarse-to-fine Context Compression via Marginal Information Gain"
// 2026 — scores each line by how much NEW information it contributes to the
// already-retained set, rather than scoring lines in isolation.
//
// Key insight: lines that repeat already-covered terms add zero marginal gain,
// so they can be dropped even if individually "important."
//
// Algorithm:
//  1. Build global TF map; identify discriminative terms (frequent but not ubiquitous)
//  2. Assign each line a term-set covering its discriminative terms
//  3. Greedy selection: rank lines by marginal_gain / token_cost, apply structural bonus
//  4. Fill token budget top-down; anchor first and last non-empty lines
type MarginalInfoGainFilter struct {
	targetRatio float64
	minTermFreq int
	stopWords   map[string]bool
}

// NewMarginalInfoGainFilter creates a new COMI-style marginal information gain filter.
func NewMarginalInfoGainFilter() *MarginalInfoGainFilter {
	return &MarginalInfoGainFilter{
		targetRatio: 0.55,
		minTermFreq: 2,
		stopWords:   migStopWords(),
	}
}

// Name returns the filter name.
func (f *MarginalInfoGainFilter) Name() string { return "21_marginal_info_gain" }

// Apply selects lines that maximize marginal information gain within a token budget.
func (f *MarginalInfoGainFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 5 {
		return input, 0
	}

	ratio := f.targetRatio
	if mode == ModeAggressive {
		ratio *= 0.7
	}

	budget := int(math.Ceil(float64(core.EstimateTokens(input)) * ratio))

	globalFreq := f.buildTermFreq(lines)
	lineTerms := f.buildLineTermSets(lines, globalFreq)
	kept := f.greedySelect(lines, lineTerms, budget)

	var result []string
	for i, line := range lines {
		if kept[i] {
			result = append(result, line)
		}
	}
	if len(result) == 0 {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *MarginalInfoGainFilter) buildTermFreq(lines []string) map[string]int {
	freq := make(map[string]int)
	for _, line := range lines {
		for _, t := range f.tokenizeLine(line) {
			freq[t]++
		}
	}
	return freq
}

func (f *MarginalInfoGainFilter) buildLineTermSets(lines []string, freq map[string]int) []map[string]bool {
	n := len(lines)
	sets := make([]map[string]bool, n)
	for i, line := range lines {
		set := make(map[string]bool)
		for _, t := range f.tokenizeLine(line) {
			tf := freq[t]
			// Keep discriminative terms: appear ≥ minTermFreq times but in fewer than half the lines
			if (tf >= f.minTermFreq && tf < n/2) || tf == 1 {
				set[t] = true
			}
		}
		sets[i] = set
	}
	return sets
}

type migCandidate struct {
	idx   int
	score float64 // marginal_gain * structural_bonus / token_cost
}

func (f *MarginalInfoGainFilter) greedySelect(lines []string, lineTerms []map[string]bool, budget int) map[int]bool {
	kept := make(map[int]bool)
	covered := make(map[string]bool)
	used := 0

	// Anchor: first non-empty line
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			kept[i] = true
			used += core.EstimateTokens(line)
			for t := range lineTerms[i] {
				covered[t] = true
			}
			break
		}
	}
	// Anchor: last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" && !kept[i] {
			kept[i] = true
			used += core.EstimateTokens(lines[i])
			for t := range lineTerms[i] {
				covered[t] = true
			}
			break
		}
	}

	// Score remaining candidates
	candidates := make([]migCandidate, 0, len(lines))
	for i, line := range lines {
		if kept[i] || strings.TrimSpace(line) == "" {
			continue
		}
		toks := core.EstimateTokens(line)
		if toks == 0 {
			toks = 1
		}
		gain := marginalGain(lineTerms[i], covered)
		bonus := structuralBonus(line)
		candidates = append(candidates, migCandidate{idx: i, score: (gain * bonus) / float64(toks)})
	}

	sort.Slice(candidates, func(a, b int) bool {
		return candidates[a].score > candidates[b].score
	})

	for _, c := range candidates {
		if used >= budget {
			break
		}
		toks := core.EstimateTokens(lines[c.idx])
		kept[c.idx] = true
		used += toks
		for t := range lineTerms[c.idx] {
			covered[t] = true
		}
	}
	return kept
}

func marginalGain(terms map[string]bool, covered map[string]bool) float64 {
	gain := 0.0
	for t := range terms {
		if !covered[t] {
			gain++
		}
	}
	return gain + 0.1 // small floor so zero-gain lines can still win via structural bonus
}

func structuralBonus(line string) float64 {
	if isErrorLine(line) || isWarningLine(line) {
		return 3.0
	}
	if isHeadingLine(line) {
		return 1.8
	}
	if isCodeLine(line) {
		return 1.2
	}
	return 1.0
}

func (f *MarginalInfoGainFilter) tokenizeLine(line string) []string {
	var terms []string
	var word strings.Builder
	for _, ch := range strings.ToLower(line) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			word.WriteRune(ch)
		} else if word.Len() > 0 {
			w := word.String()
			if len(w) > 2 && !f.stopWords[w] {
				terms = append(terms, w)
			}
			word.Reset()
		}
	}
	if word.Len() > 2 {
		if w := word.String(); !f.stopWords[w] {
			terms = append(terms, w)
		}
	}
	return terms
}

func migStopWords() map[string]bool {
	sw := map[string]bool{}
	for _, w := range []string{
		"the", "and", "for", "that", "this", "with", "from", "are", "was",
		"were", "has", "have", "had", "not", "but", "its", "can", "will",
		"all", "any", "one", "more", "also", "when", "then", "than", "too",
		"use", "used", "using", "new", "get", "set", "add", "let", "var",
	} {
		sw[w] = true
	}
	return sw
}
