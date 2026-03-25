package filter

import (
	"math"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// CIBFilter implements Conditional Information Bottleneck compression for CoT reasoning.
// Research: "Reasoning as Compression: Unifying Budget Forcing via CIB" (arXiv 2603.08462, Mar 2026)
// Key Innovation: Treat CoT reasoning as lossy compression under Information Bottleneck principle.
// Prunes cognitive bloat while preserving fluency and logic using semantic surprisal.
// Results: Aggressive compression with minimal accuracy drop on reasoning tasks.
//
// Unlike simple truncation, CIB identifies which reasoning steps are essential
// vs redundant based on surprisal (how "surprising" a token is given context).
type CIBFilter struct {
	config CIBConfig
}

// CIBConfig holds configuration for CIB compression
type CIBConfig struct {
	Enabled             bool
	SurprisalThreshold  float64 // Tokens above this surprisal are kept
	MaxReasoningTokens  int     // Maximum tokens for reasoning traces
	PreserveConclusions bool
	MinContentLength    int
}

// DefaultCIBConfig returns default configuration
func DefaultCIBConfig() CIBConfig {
	return CIBConfig{
		Enabled:             true,
		SurprisalThreshold:  0.3,
		MaxReasoningTokens:  500,
		PreserveConclusions: true,
		MinContentLength:    200,
	}
}

// NewCIBFilter creates a new CIB filter
func NewCIBFilter() *CIBFilter {
	return &CIBFilter{config: DefaultCIBConfig()}
}

// Name returns the filter name
func (c *CIBFilter) Name() string { return "cib" }

// Apply applies CIB-based reasoning compression
func (c *CIBFilter) Apply(input string, mode Mode) (string, int) {
	if !c.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < c.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Detect if this is a reasoning trace
	if !c.isReasoningTrace(input) {
		return input, 0
	}

	// Split into reasoning steps
	steps := c.splitReasoningSteps(input)

	// Compute surprisal for each step
	surprisals := c.computeSurprisal(steps)

	// Select essential steps
	output := c.selectEssentialSteps(steps, surprisals, mode)

	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 5 {
		return input, 0
	}

	return output, saved
}

// isReasoningTrace detects CoT/reasoning content
func (c *CIBFilter) isReasoningTrace(input string) bool {
	lower := strings.ToLower(input)
	indicators := []string{
		"step 1", "step 2", "let me", "first,", "second,",
		"therefore", "thus", "conclusion:", "reasoning:",
		"thinking about", "approach:", "solution:",
	}

	count := 0
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			count++
		}
	}
	return count >= 2
}

// splitReasoningSteps splits into logical reasoning steps
func (c *CIBFilter) splitReasoningSteps(input string) []string {
	// Split by paragraphs
	paragraphs := strings.Split(input, "\n\n")
	var steps []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			steps = append(steps, trimmed)
		}
	}
	return steps
}

// computeSurprisal computes surprisal score for each step.
// High surprisal = novel/important information. Low surprisal = filler/repetition.
func (c *CIBFilter) computeSurprisal(steps []string) []float64 {
	surprisals := make([]float64, len(steps))
	wordFreq := make(map[string]int)

	// Build frequency from all steps
	for _, step := range steps {
		for _, w := range strings.Fields(strings.ToLower(step)) {
			wordFreq[w]++
		}
	}

	totalWords := 0
	for _, count := range wordFreq {
		totalWords += count
	}

	for i, step := range steps {
		words := strings.Fields(strings.ToLower(step))
		if len(words) == 0 {
			surprisals[i] = 0
			continue
		}

		// Compute average surprisal (-log P(word))
		totalSurprisal := 0.0
		for _, w := range words {
			freq := float64(wordFreq[w]) / float64(totalWords)
			if freq > 0 {
				totalSurprisal += -math.Log2(freq)
			}
		}
		surprisals[i] = totalSurprisal / float64(len(words))
	}

	// Normalize to 0-1
	maxS := 0.0
	for _, s := range surprisals {
		if s > maxS {
			maxS = s
		}
	}
	if maxS > 0 {
		for i := range surprisals {
			surprisals[i] /= maxS
		}
	}

	return surprisals
}

// selectEssentialSteps keeps only essential reasoning steps
func (c *CIBFilter) selectEssentialSteps(steps []string, surprisals []float64, mode Mode) string {
	threshold := c.config.SurprisalThreshold
	if mode == ModeAggressive {
		threshold += 0.2
	}

	var result strings.Builder

	for i, step := range steps {
		// Always keep first and last steps
		if i == 0 || i == len(steps)-1 {
			if result.Len() > 0 {
				result.WriteString("\n\n")
			}
			result.WriteString(step)
			continue
		}

		// Keep high-surprisal steps
		if surprisals[i] >= threshold {
			if result.Len() > 0 {
				result.WriteString("\n\n")
			}
			result.WriteString(step)
		}
	}

	output := result.String()
	if output == "" && len(steps) > 0 {
		return steps[0]
	}
	return output
}
