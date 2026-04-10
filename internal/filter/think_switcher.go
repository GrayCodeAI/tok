package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Papers:
//
//	"ThinkSwitcher: When to Think Hard, When to Think Fast" — EMNLP 2025
//	"Thinkless: LLM Learns When to Think" — NeurIPS 2025 (VainF/Thinkless)
//
// ThinkSwitcherFilter is a meta-routing filter: it measures the "reasoning
// density" of an output (fraction of lines that look like deliberate
// reasoning) and routes to the appropriate compression level.
//
// Three paths:
//
//	fast   — reasoning density < fastThreshold: no reasoning detected,
//	          pass through unchanged (avoids wasted CPU on direct answers)
//	light  — fastThreshold ≤ density < heavyThreshold: some reasoning,
//	          compress to 50% of reasoning lines keeping key sentences
//	heavy  — density ≥ heavyThreshold: heavy reasoning present,
//	          collapse to a one-line summary per reasoning block
//
// Key insight from ThinkSwitcher: the majority of LLM outputs need no CoT
// compression at all. Running compression unconditionally wastes resources
// and can degrade quality by removing relevant content from direct answers.
type ThinkSwitcherFilter struct {
	fastThreshold  float64 // below this → fast path
	heavyThreshold float64 // at or above this → heavy compression
}

// NewThinkSwitcherFilter creates a new ThinkSwitcher routing filter.
func NewThinkSwitcherFilter() *ThinkSwitcherFilter {
	return &ThinkSwitcherFilter{
		fastThreshold:  0.12,
		heavyThreshold: 0.35,
	}
}

// Name returns the filter name.
func (f *ThinkSwitcherFilter) Name() string { return "27_think_switcher" }

// Apply routes compression based on detected reasoning density.
func (f *ThinkSwitcherFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 6 {
		return input, 0
	}

	density := f.reasoningDensity(lines)

	if density < f.fastThreshold {
		// Fast path: no meaningful reasoning present
		return input, 0
	}

	if mode == ModeAggressive || density >= f.heavyThreshold {
		return f.heavyCompress(input, lines)
	}

	return f.lightCompress(input, lines)
}

// reasoningDensity returns the fraction of lines that look like deliberate reasoning.
func (f *ThinkSwitcherFilter) reasoningDensity(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}
	count := 0
	for _, line := range lines {
		if isReasoningLine(line) {
			count++
		}
	}
	return float64(count) / float64(len(lines))
}

// isReasoningLine returns true for lines that look like deliberate reasoning steps.
func isReasoningLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return false
	}
	for _, prefix := range []string{
		"step ", "first,", "second,", "third,", "fourth,", "fifth,",
		"finally,", "lastly,", "next,", "now,", "then,",
		"let me ", "i need to ", "i should ", "i will ", "i can ",
		"to do this,", "therefore,", "thus,", "hence,", "so,",
		"consider", "analyze", "check", "verify", "note that",
		"wait,", "actually,", "hmm,", "on second thought",
		"the reason", "because", "since", "given that",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	// Numbered list items ("1. ", "2. ", etc.)
	if len(lower) > 2 && lower[0] >= '1' && lower[0] <= '9' && (lower[1] == '.' || lower[1] == ')') {
		return true
	}
	return false
}

// lightCompress retains ~50% of reasoning lines, keeping the most informative ones.
func (f *ThinkSwitcherFilter) lightCompress(input string, lines []string) (string, int) {
	type tsScoredLine struct {
		idx   int
		score float64
	}

	// Score reasoning lines by term density
	termFreq := tsTermFreq(lines)
	var reasoningScored []tsScoredLine
	for i, line := range lines {
		if !isReasoningLine(line) {
			continue
		}
		score := tsLineScore(line, termFreq)
		reasoningScored = append(reasoningScored, tsScoredLine{idx: i, score: score})
	}

	// Keep top 50% of reasoning lines
	keep := len(reasoningScored) / 2
	if keep < 1 {
		keep = 1
	}

	// Sort by score descending, mark top-keep as retained
	for i := 1; i < len(reasoningScored); i++ {
		for j := i; j > 0 && reasoningScored[j].score > reasoningScored[j-1].score; j-- {
			reasoningScored[j], reasoningScored[j-1] = reasoningScored[j-1], reasoningScored[j]
		}
	}
	retained := make(map[int]bool)
	for _, s := range reasoningScored[:keep] {
		retained[s.idx] = true
	}

	var result []string
	for i, line := range lines {
		if isReasoningLine(line) && !retained[i] {
			continue
		}
		result = append(result, line)
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// heavyCompress collapses each contiguous reasoning block to one summary line.
func (f *ThinkSwitcherFilter) heavyCompress(input string, lines []string) (string, int) {
	type block struct{ start, end int }
	var blocks []block
	inBlock := false
	start := 0

	for i, line := range lines {
		if isReasoningLine(line) {
			if !inBlock {
				inBlock = true
				start = i
			}
		} else {
			if inBlock {
				if i-start >= 3 {
					blocks = append(blocks, block{start, i - 1})
				}
				inBlock = false
			}
		}
	}
	if inBlock && len(lines)-start >= 3 {
		blocks = append(blocks, block{start, len(lines) - 1})
	}

	if len(blocks) == 0 {
		return f.lightCompress(input, lines)
	}

	suppress := make(map[int]bool)
	annotation := make(map[int]string)
	termFreq := tsTermFreq(lines)

	for _, b := range blocks {
		body := lines[b.start : b.end+1]
		toks := core.EstimateTokens(strings.Join(body, "\n"))
		// Pick best representative line
		best := body[0]
		bestScore := -1.0
		for _, line := range body {
			if s := tsLineScore(line, termFreq); s > bestScore {
				bestScore = s
				best = line
			}
		}
		annotation[b.start] = best + " [reasoning: ~" + itoa(toks) + " tok compressed]"
		for i := b.start + 1; i <= b.end; i++ {
			suppress[i] = true
		}
		suppress[b.start] = true // replaced by annotation
	}

	var result []string
	for i, line := range lines {
		if suppress[i] {
			if ann, ok := annotation[i]; ok {
				result = append(result, ann)
			}
			continue
		}
		result = append(result, line)
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// -- helpers --

func tsTermFreq(lines []string) map[string]int {
	freq := make(map[string]int)
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			freq[t]++
		}
	}
	return freq
}

func tsLineScore(line string, freq map[string]int) float64 {
	terms := ltTokenize(line)
	if len(terms) == 0 {
		return 0
	}
	score := 0.0
	for _, t := range terms {
		if f := freq[t]; f > 0 {
			score += 1.0 / float64(f)
		}
	}
	return score / float64(len(terms))
}
