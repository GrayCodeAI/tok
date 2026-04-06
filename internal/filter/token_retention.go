package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "Cache What Lasts: Token Retention for Memory-Bounded KV Cache" — Bui et al., Yale/JPMorgan, 2026
// TokenRetentionFilter identifies tokens that should be retained in memory
// based on their lasting importance across the context window.
type TokenRetentionFilter struct {
	retentionWindow int
}

// NewTokenRetentionFilter creates a new token retention filter.
func NewTokenRetentionFilter() *TokenRetentionFilter {
	return &TokenRetentionFilter{retentionWindow: 10}
}

// Apply retains tokens with lasting importance, prunes transient ones.
func (f *TokenRetentionFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	lines := strings.Split(input, "\n")

	tokenFreq := make(map[string]int)
	tokenLines := make(map[string][]int)

	for i, line := range lines {
		words := strings.Fields(line)
		for _, w := range words {
			clean := strings.ToLower(strings.Trim(w, ".,;:()[]{}\"'"))
			if len(clean) > 2 {
				tokenFreq[clean]++
				tokenLines[clean] = append(tokenLines[clean], i)
			}
		}
	}

	retainedLines := make(map[int]bool)
	for token, freq := range tokenFreq {
		if freq >= 2 {
			for _, lineIdx := range tokenLines[token] {
				retainedLines[lineIdx] = true
			}
		}
	}

	for i := 0; i < f.retentionWindow && i < len(lines); i++ {
		retainedLines[i] = true
	}
	for i := len(lines) - f.retentionWindow; i < len(lines); i++ {
		if i >= 0 {
			retainedLines[i] = true
		}
	}

	var result []string
	for i, line := range lines {
		if retainedLines[i] {
			result = append(result, line)
		}
	}

	if len(result) == 0 {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(original) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// Name returns the layer name.
func (f *TokenRetentionFilter) Name() string { return "26_token_retention" }
