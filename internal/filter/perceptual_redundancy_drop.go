package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "Perception Compressor: Training-Free Prompt Compression for Long Context"
// arXiv:2504.xxxxx — 2025
//
// PerceptionCompressFilter identifies "perceptually redundant" lines: those whose
// semantic content is already covered by their immediate neighbors.  Removing
// them does not change what an LLM would perceive as the meaning of the context.
//
// Proxy for perceptual redundancy (training-free):
//   - Compute term-overlap between line i and its window (i±windowSize)
//   - If overlap / own_terms ≥ threshold, the line is dominated by neighbors
//
// This catches verbose prose, repeated captions, duplicate log prefixes, and
// transitional boilerplate that carries no new information.
type PerceptionCompressFilter struct {
	windowSize int     // lines on each side to compare against
	threshold  float64 // min overlap fraction to consider a line redundant
	minLineLen int     // skip lines shorter than this
}

// NewPerceptionCompressFilter creates a new Perception Compressor filter.
func NewPerceptionCompressFilter() *PerceptionCompressFilter {
	return &PerceptionCompressFilter{
		windowSize: 3,
		threshold:  0.75,
		minLineLen: 15,
	}
}

// Name returns the filter name.
func (f *PerceptionCompressFilter) Name() string { return "25_perception_compress" }

// Apply removes perceptually redundant lines.
func (f *PerceptionCompressFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	thresh := f.threshold
	if mode == ModeAggressive {
		thresh = 0.60 // more aggressive overlap threshold
	}

	lines := strings.Split(input, "\n")
	if len(lines) < f.windowSize*2+2 {
		return input, 0
	}

	// Build term sets per line
	termSets := make([]map[string]bool, len(lines))
	for i, line := range lines {
		termSets[i] = pcTermSet(line)
	}

	keep := make([]bool, len(lines))
	// Always keep first and last lines
	keep[0] = true
	keep[len(lines)-1] = true

	for i := 1; i < len(lines)-1; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Never drop short lines, empty lines, or structural lines
		if len(trimmed) < f.minLineLen || trimmed == "" {
			keep[i] = true
			continue
		}
		if isErrorLine(line) || isWarningLine(line) || isHeadingLine(line) {
			keep[i] = true
			continue
		}

		own := termSets[i]
		if len(own) == 0 {
			keep[i] = true
			continue
		}

		// Build neighbor term set from window
		neighbors := make(map[string]bool)
		lo := i - f.windowSize
		if lo < 0 {
			lo = 0
		}
		hi := i + f.windowSize
		if hi >= len(lines) {
			hi = len(lines) - 1
		}
		for j := lo; j <= hi; j++ {
			if j == i {
				continue
			}
			for t := range termSets[j] {
				neighbors[t] = true
			}
		}

		// Compute overlap: fraction of own terms already in neighbors
		covered := 0
		for t := range own {
			if neighbors[t] {
				covered++
			}
		}
		overlap := float64(covered) / float64(len(own))
		keep[i] = overlap < thresh
	}

	var result []string
	for i, line := range lines {
		if keep[i] {
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

// pcTermSet builds a term set for a line (lowercase alphabetic tokens ≥ 3 chars).
func pcTermSet(line string) map[string]bool {
	set := make(map[string]bool)
	var word strings.Builder
	for _, ch := range strings.ToLower(line) {
		if (ch >= 'a' && ch <= 'z') || ch == '_' {
			word.WriteRune(ch)
		} else if word.Len() > 0 {
			if w := word.String(); len(w) >= 3 {
				set[w] = true
			}
			word.Reset()
		}
	}
	if word.Len() >= 3 {
		set[word.String()] = true
	}
	return set
}
