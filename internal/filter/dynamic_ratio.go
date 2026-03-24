package filter

import (
	"math"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// DynamicRatioFilter implements PruneSID-style dynamic compression ratio.
// Research Source: "Prune Redundancy, Preserve Essence" (Mar 2026)
// Key Innovation: Content complexity analysis to auto-adjust compression ratio,
// enabling more aggressive compression on redundant content while preserving
// simple/important content.
//
// This is a meta-layer that adjusts the effective compression budget based on
// the information density of the content. High-density content gets more tokens;
// low-density (redundant) content gets compressed more aggressively.
type DynamicRatioFilter struct {
	config DynamicRatioConfig
}

// DynamicRatioConfig holds configuration for dynamic ratio adjustment
type DynamicRatioConfig struct {
	// Enabled controls whether the filter is active
	Enabled bool

	// MinComplexity is the minimum complexity score (0-1)
	MinComplexity float64

	// MaxComplexity is the maximum complexity score (0-1)
	MaxComplexity float64

	// BaseBudgetRatio is the default budget ratio (1.0 = no change)
	BaseBudgetRatio float64

	// HighComplexityBoost multiplies budget for high-complexity content
	HighComplexityBoost float64

	// LowComplexityPenalty multiplies budget for low-complexity content
	LowComplexityPenalty float64

	// MinContentLength is minimum chars to analyze
	MinContentLength int
}

// DefaultDynamicRatioConfig returns default configuration
func DefaultDynamicRatioConfig() DynamicRatioConfig {
	return DynamicRatioConfig{
		Enabled:              true,
		MinComplexity:        0.2,
		MaxComplexity:        0.8,
		BaseBudgetRatio:      1.0,
		HighComplexityBoost:  1.5,
		LowComplexityPenalty: 0.5,
		MinContentLength:     100,
	}
}

// NewDynamicRatioFilter creates a new dynamic ratio filter
func NewDynamicRatioFilter() *DynamicRatioFilter {
	return &DynamicRatioFilter{
		config: DefaultDynamicRatioConfig(),
	}
}

// Name returns the filter name
func (f *DynamicRatioFilter) Name() string {
	return "dynamic_ratio"
}

// Apply applies dynamic compression ratio based on content complexity
func (f *DynamicRatioFilter) Apply(input string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Analyze content complexity
	complexity := f.analyzeComplexity(input)

	// Compute dynamic budget multiplier
	multiplier := f.computeMultiplier(complexity, mode)

	// Apply ratio-based compression
	output := f.applyRatioCompression(input, multiplier, mode)

	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 3 {
		return input, 0
	}

	return output, saved
}

// ContentComplexity holds analysis results
type ContentComplexity struct {
	EntropyDensity   float64 // Shannon entropy per character
	VocabularyRatio  float64 // Unique words / total words
	StructureDensity float64 // Structural elements ratio
	RedundancyRatio  float64 // Estimated redundancy (0 = no redundancy)
	OverallScore     float64 // Combined complexity score (0-1)
}

// analyzeComplexity analyzes the content's information density
func (f *DynamicRatioFilter) analyzeComplexity(input string) ContentComplexity {
	var c ContentComplexity

	// 1. Shannon entropy density
	c.EntropyDensity = f.shannonEntropyDensity(input)

	// 2. Vocabulary ratio
	c.VocabularyRatio = f.vocabularyRatio(input)

	// 3. Structure density
	c.StructureDensity = f.structureDensity(input)

	// 4. Redundancy ratio
	c.RedundancyRatio = f.redundancyRatio(input)

	// Combined score (higher = more complex = preserve more)
	c.OverallScore = (c.EntropyDensity*0.3 + c.VocabularyRatio*0.3 +
		c.StructureDensity*0.2 + (1-c.RedundancyRatio)*0.2)

	return c
}

// shannonEntropyDensity computes entropy per character
func (f *DynamicRatioFilter) shannonEntropyDensity(input string) float64 {
	if len(input) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, r := range input {
		freq[r]++
	}

	entropy := 0.0
	total := float64(len(input))
	for _, count := range freq {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize to [0, 1] (max entropy for ASCII ~7 bits)
	return math.Min(entropy/7.0, 1.0)
}

// vocabularyRatio computes unique words / total words
func (f *DynamicRatioFilter) vocabularyRatio(input string) float64 {
	words := strings.Fields(strings.ToLower(input))
	if len(words) == 0 {
		return 0
	}

	unique := make(map[string]bool)
	for _, w := range words {
		unique[w] = true
	}

	return float64(len(unique)) / float64(len(words))
}

// structureDensity computes the ratio of structural elements
func (f *DynamicRatioFilter) structureDensity(input string) float64 {
	lines := strings.Split(input, "\n")
	structuralCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Structural indicators
		if strings.Contains(trimmed, "{") || strings.Contains(trimmed, "}") ||
			strings.Contains(trimmed, "(") || strings.Contains(trimmed, ")") ||
			strings.Contains(trimmed, "[") || strings.Contains(trimmed, "]") ||
			strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "package ") ||
			strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "for ") ||
			strings.HasPrefix(trimmed, "while ") || strings.HasPrefix(trimmed, "return ") {
			structuralCount++
		}
	}

	totalLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			totalLines++
		}
	}

	if totalLines == 0 {
		return 0
	}

	return float64(structuralCount) / float64(totalLines)
}

// redundancyRatio estimates content redundancy
func (f *DynamicRatioFilter) redundancyRatio(input string) float64 {
	lines := strings.Split(input, "\n")
	if len(lines) < 3 {
		return 0
	}

	// Count duplicate lines
	lineCounts := make(map[string]int)
	totalNonEmpty := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lineCounts[trimmed]++
			totalNonEmpty++
		}
	}

	if totalNonEmpty == 0 {
		return 0
	}

	// Compute redundancy as fraction of non-unique lines
	duplicateCount := 0
	for _, count := range lineCounts {
		if count > 1 {
			duplicateCount += count - 1
		}
	}

	return float64(duplicateCount) / float64(totalNonEmpty)
}

// computeMultiplier computes budget multiplier from complexity
func (f *DynamicRatioFilter) computeMultiplier(c ContentComplexity, mode Mode) float64 {
	multiplier := f.config.BaseBudgetRatio

	if c.OverallScore > f.config.MaxComplexity {
		// High complexity → preserve more
		multiplier *= f.config.HighComplexityBoost
	} else if c.OverallScore < f.config.MinComplexity {
		// Low complexity → compress more aggressively
		multiplier *= f.config.LowComplexityPenalty
	}

	// Mode adjustments
	if mode == ModeAggressive {
		multiplier *= 0.7
	}

	return multiplier
}

// applyRatioCompression applies ratio-based line-level compression
func (f *DynamicRatioFilter) applyRatioCompression(input string, multiplier float64, mode Mode) string {
	if multiplier >= 1.0 {
		return input // No compression needed
	}

	lines := strings.Split(input, "\n")
	targetLines := int(math.Ceil(float64(len(lines)) * multiplier))

	if targetLines >= len(lines) {
		return input
	}

	// Score each line and keep the top targetLines (uses existing scoredLine from budget.go)
	var scored []scoredLine
	for i, line := range lines {
		scored = append(scored, scoredLine{
			line:  line,
			score: f.scoreLine(line),
			index: i,
		})
	}

	// Simple: keep top lines by score
	kept := f.selectTopLines(scored, targetLines)
	return strings.Join(kept, "\n")
}

// scoreLine scores a line's importance (higher = more important)
func (f *DynamicRatioFilter) scoreLine(line string) float64 {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return 0.1
	}

	score := 0.5

	// Structural lines are important
	if strings.Contains(trimmed, "{") || strings.Contains(trimmed, "}") {
		score += 0.3
	}
	if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "class ") {
		score += 0.4
	}
	if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "package ") {
		score += 0.1
	}

	// Error/warning lines are important
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "error") || strings.Contains(lower, "fail") {
		score += 0.5
	}
	if strings.Contains(lower, "warn") {
		score += 0.3
	}

	// Short lines are less informative
	if len(trimmed) < 10 {
		score -= 0.2
	}

	return math.Max(0, math.Min(1, score))
}

// selectTopLines selects the top N lines by score, preserving order
func (f *DynamicRatioFilter) selectTopLines(scored []scoredLine, n int) []string {
	if n >= len(scored) {
		var result []string
		for _, s := range scored {
			result = append(result, s.line)
		}
		return result
	}

	// Mark which indices to keep
	keep := make([]bool, len(scored))

	// Simple selection: find top n by score
	for i := 0; i < n; i++ {
		bestIdx := -1
		bestScore := -1.0
		for j := 0; j < len(scored); j++ {
			if !keep[j] && scored[j].score > bestScore {
				bestScore = scored[j].score
				bestIdx = j
			}
		}
		if bestIdx >= 0 {
			keep[bestIdx] = true
		}
	}

	// Reconstruct in original order
	var result []string
	for i, s := range scored {
		if keep[i] {
			result = append(result, s.line)
		}
	}

	return result
}
