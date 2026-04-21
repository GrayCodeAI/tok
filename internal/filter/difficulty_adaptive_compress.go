package filter

import (
	"math"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "DiffAdapt: Difficulty-Adaptive Token Compression for LLM Inference"
// ICLR 2026
//
// DiffAdaptFilter is a meta-controller that measures input "difficulty" —
// the structural complexity of the text — and scales the compression ratio
// applied downstream accordingly.
//
// Difficulty is estimated via three signals:
//  1. Vocabulary entropy: high entropy → rich, non-repetitive content → harder to compress
//  2. Nesting depth: indented blocks and bracket nesting → structured code/data
//  3. Average line length: longer lines tend to carry more information density
//
// The filter then prunes lines whose per-line information score falls below a
// difficulty-scaled threshold. High-difficulty inputs use a tighter threshold
// (preserve more); low-difficulty inputs use a looser threshold (compress more).
//
// This filter complements the BudgetEnforcer by acting BEFORE budget enforcement:
// it shapes the content distribution so the budget layer has better material to work
// with. Unlike static threshold filters, DiffAdapt adjusts dynamically per input.
type DiffAdaptFilter struct {
	baseThreshold float64 // baseline per-line score threshold (difficulty=0.5)
	minThreshold  float64 // floor when input is very easy (low complexity)
	maxThreshold  float64 // ceiling when input is very hard (high complexity)
}

// NewDiffAdaptFilter creates a new difficulty-adaptive compression filter.
func NewDiffAdaptFilter() *DiffAdaptFilter {
	return &DiffAdaptFilter{
		baseThreshold: 0.30,
		minThreshold:  0.15,
		maxThreshold:  0.55,
	}
}

// Name returns the filter name.
func (f *DiffAdaptFilter) Name() string { return "31_difft_adapt" }

// Apply measures input difficulty and prunes low-scoring lines adaptively.
func (f *DiffAdaptFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 8 {
		return input, 0
	}

	difficulty := f.measureDifficulty(lines)

	// Scale threshold inversely with difficulty:
	// difficulty=0.0 → maxThreshold (easy content → compress aggressively)
	// difficulty=1.0 → minThreshold (hard content → compress conservatively)
	threshold := f.maxThreshold - difficulty*(f.maxThreshold-f.minThreshold)
	if mode == ModeAggressive {
		threshold *= 0.75 // push thresholds down → more lines dropped
	}

	// Score each line and drop below-threshold lines that aren't anchors
	termFreq := daTermFrequency(lines)
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}
		if isErrorLine(line) || isWarningLine(line) || isHeadingLine(line) {
			result = append(result, line)
			continue
		}
		score := daLineScore(line, termFreq, len(lines))
		if score >= threshold {
			result = append(result, line)
		}
	}

	if len(result) == len(lines) {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// measureDifficulty returns a 0.0–1.0 score for input complexity.
func (f *DiffAdaptFilter) measureDifficulty(lines []string) float64 {
	if len(lines) == 0 {
		return 0.5
	}

	// Signal 1: vocabulary entropy
	termCounts := make(map[string]int)
	totalTerms := 0
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			termCounts[t]++
			totalTerms++
		}
	}
	entropy := 0.0
	if totalTerms > 0 {
		for _, count := range termCounts {
			p := float64(count) / float64(totalTerms)
			if p > 0 {
				entropy -= p * math.Log2(p)
			}
		}
	}
	// Normalise: most CLI output sits between 3-10 bits. Map to 0-1.
	entropyScore := math.Min(entropy/10.0, 1.0)

	// Signal 2: nesting depth (indentation level)
	totalDepth := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		spaces := 0
		for _, ch := range line {
			if ch == ' ' {
				spaces++
			} else if ch == '\t' {
				spaces += 4
			} else {
				break
			}
		}
		totalDepth += spaces / 4
	}
	avgDepth := float64(totalDepth) / float64(len(lines))
	nestScore := math.Min(avgDepth/6.0, 1.0)

	// Signal 3: average line length
	totalLen := 0
	for _, line := range lines {
		totalLen += len(line)
	}
	avgLen := float64(totalLen) / float64(len(lines))
	lenScore := math.Min(avgLen/120.0, 1.0)

	// Weighted combination
	return 0.5*entropyScore + 0.3*nestScore + 0.2*lenScore
}

// daTermFrequency builds a term frequency map across all lines.
func daTermFrequency(lines []string) map[string]int {
	freq := make(map[string]int)
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			freq[t]++
		}
	}
	return freq
}

// daLineScore scores a line by its average inverse term frequency (rare terms = high score).
func daLineScore(line string, termFreq map[string]int, nLines int) float64 {
	terms := ltTokenize(line)
	if len(terms) == 0 {
		return 0
	}
	score := 0.0
	for _, t := range terms {
		freq := termFreq[t]
		if freq == 0 {
			freq = 1
		}
		// ITF: lines with rare terms score higher
		score += 1.0 / float64(freq)
	}
	return score / float64(len(terms)) * float64(nLines) / 10.0
}
