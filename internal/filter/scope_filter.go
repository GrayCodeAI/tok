package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// ScopeFilter implements SCOPE-style separate prefill/decode optimization.
// Research Source: "SCOPE: Optimizing Key-Value Cache Compression in Long-context
// Generation" (ACL 2025)
// Key Innovation: Separate optimization strategies for initial context (prefill)
// vs ongoing conversation (decode). Prefill preserves more; decode compresses more.
// Results: 35% KV cache with near-full performance.
//
// This detects whether content is initial context or ongoing conversation and
// applies appropriate compression strategy.
type ScopeFilter struct {
	config ScopeConfig
}

// ScopeConfig holds configuration for SCOPE optimization
type ScopeConfig struct {
	// Enabled controls whether the filter is active
	Enabled bool

	// PrefillBudgetRatio is the fraction of budget for prefill content (higher = preserve more)
	PrefillBudgetRatio float64

	// DecodeBudgetRatio is the fraction of budget for decode content (lower = compress more)
	DecodeBudgetRatio float64

	// ConversationTurns threshold to switch from prefill to decode mode
	ConversationTurns int

	// MinContentLength minimum chars to apply
	MinContentLength int
}

// ScopeMode represents the detected content mode
type ScopeMode int

const (
	// ScopePrefill is initial context (code, files, documentation)
	ScopePrefill ScopeMode = iota
	// ScopeDecode is ongoing conversation (chat turns, tool outputs)
	ScopeDecode
)

// DefaultScopeConfig returns default configuration
func DefaultScopeConfig() ScopeConfig {
	return ScopeConfig{
		Enabled:            true,
		PrefillBudgetRatio: 1.5, // Preserve 50% more for initial context
		DecodeBudgetRatio:  0.6, // Compress 40% more for ongoing conversation
		ConversationTurns:  3,
		MinContentLength:   200,
	}
}

// NewScopeFilter creates a new SCOPE filter
func NewScopeFilter() *ScopeFilter {
	return &ScopeFilter{
		config: DefaultScopeConfig(),
	}
}

// Name returns the filter name
func (f *ScopeFilter) Name() string {
	return "scope"
}

// Apply applies SCOPE-optimized compression
func (f *ScopeFilter) Apply(input string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	// Detect content mode
	scopeMode := f.detectMode(input)

	// Apply mode-specific compression
	switch scopeMode {
	case ScopePrefill:
		return f.compressPrefill(input, mode)
	case ScopeDecode:
		return f.compressDecode(input, mode)
	default:
		return input, 0
	}
}

// detectMode determines if content is prefill (initial) or decode (ongoing)
func (f *ScopeFilter) detectMode(input string) ScopeMode {
	// Count conversation markers
	conversationMarkers := []string{
		"User:", "Assistant:", "Human:", "AI:", "System:",
		">>> ", "## ", "### ",
	}

	markerCount := 0
	for _, marker := range conversationMarkers {
		markerCount += strings.Count(input, marker)
	}

	// Check for code/file indicators
	codeIndicators := []string{
		"func ", "function ", "def ", "class ", "struct ",
		"import ", "package ", "module ",
		"```", "type ", "interface ",
	}

	codeCount := 0
	for _, indicator := range codeIndicators {
		if strings.Contains(input, indicator) {
			codeCount++
		}
	}

	// Decision: more code indicators = prefill, more conversation markers = decode
	if codeCount > markerCount {
		return ScopePrefill
	}
	if markerCount >= f.config.ConversationTurns {
		return ScopeDecode
	}

	return ScopePrefill
}

// compressPrefill applies gentle compression for initial context
func (f *ScopeFilter) compressPrefill(input string, mode Mode) (string, int) {
	originalTokens := core.EstimateTokens(input)

	// Prefill: preserve structure, only remove obvious noise
	lines := strings.Split(input, "\n")
	var result strings.Builder
	removed := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines in aggressive mode
		if trimmed == "" && mode == ModeAggressive {
			removed++
			continue
		}

		// Skip comment-only lines in aggressive mode
		if mode == ModeAggressive && isCommentLine(trimmed) {
			removed++
			continue
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	output := strings.TrimSpace(result.String())
	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens

	return output, saved
}

// compressDecode applies aggressive compression for ongoing conversation
func (f *ScopeFilter) compressDecode(input string, mode Mode) (string, int) {
	originalTokens := core.EstimateTokens(input)

	// Decode: apply sliding window - keep recent, compress older
	lines := strings.Split(input, "\n")

	// Keep last 30% of lines, compress first 70%
	keepRecent := int(float64(len(lines)) * 0.3)
	if keepRecent < 3 {
		keepRecent = 3
	}

	var result strings.Builder

	// Compress older lines (first 70%)
	olderLines := lines[:len(lines)-keepRecent]
	compressed := f.compressOlderLines(olderLines, mode)
	if compressed != "" {
		result.WriteString(compressed)
		result.WriteString("\n")
		result.WriteString("[... earlier context compressed ...]\n\n")
	}

	// Keep recent lines as-is
	for _, line := range lines[len(lines)-keepRecent:] {
		result.WriteString(line)
		result.WriteString("\n")
	}

	output := strings.TrimSpace(result.String())
	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens

	return output, saved
}

// compressOlderLines compresses older conversation lines
func (f *ScopeFilter) compressOlderLines(lines []string, mode Mode) string {
	if len(lines) == 0 {
		return ""
	}

	// Keep only lines with important content
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Keep lines with errors, warnings, file paths, numbers
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "warn") ||
			strings.Contains(lower, "fail") ||
			strings.Contains(trimmed, "/") ||
			strings.Contains(trimmed, ":") ||
			isNumber(trimmed) {
			kept = append(kept, trimmed)
		}
	}

	return strings.Join(kept, "\n")
}

// isCommentLine checks if a line is a code comment
func isCommentLine(line string) bool {
	commentPrefixes := []string{"//", "#", "/*", "*", "<!--", "-->"}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
