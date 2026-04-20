package filter

import (
	"regexp"
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// ContextCrunchFilter combines Layer 46 (LogCrunch) and Layer 48 (DiffCrunch)
// into a unified context-folding layer that auto-detects content type.
//
// This merged layer:
// - Auto-detects if input is logs or diffs
// - Applies appropriate folding strategy
// - Handles both types with unified logic
type ContextCrunchFilter struct {
	logCrunch  *LogCrunchFilter
	diffCrunch *DiffCrunchFilter
}

var (
	ccDiffPattern      = regexp.MustCompile(`^(diff --git|@@|\+\+\+|---) `)
	ccLogPattern       = regexp.MustCompile(`(?i)\b(info|debug|warn|error|fatal)\b.*\d{4}-\d{2}-\d{2}`)
	ccTimestampPattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`)
)

// NewContextCrunchFilter creates a new context crunch filter.
// This replaces both NewLogCrunchFilter() and NewDiffCrunchFilter().
func NewContextCrunchFilter() *ContextCrunchFilter {
	return &ContextCrunchFilter{
		logCrunch:  NewLogCrunchFilter(),
		diffCrunch: NewDiffCrunchFilter(),
	}
}

// Name returns the filter name.
func (c *ContextCrunchFilter) Name() string { return "46_context_crunch" }

// ContextContentType represents the detected content type for context crunching.
type ContextContentType int

const (
	ContextContentTypeUnknown ContextContentType = iota
	ContextContentTypeLog
	ContextContentTypeDiff
)

// Apply auto-detects content type and applies appropriate folding.
func (c *ContextCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 20 {
		return input, 0
	}

	// Auto-detect content type
	contentType := c.detectContentType(lines)

	switch contentType {
	case ContextContentTypeDiff:
		return c.diffCrunch.Apply(input, mode)
	case ContextContentTypeLog:
		return c.logCrunch.Apply(input, mode)
	default:
		// Try both and use the one that saves more tokens
		logOutput, logSaved := c.logCrunch.Apply(input, mode)
		diffOutput, diffSaved := c.diffCrunch.Apply(input, mode)

		if logSaved > diffSaved {
			return logOutput, logSaved
		}
		return diffOutput, diffSaved
	}
}

// detectContentType analyzes input to determine if it's logs or diffs.
func (c *ContextCrunchFilter) detectContentType(lines []string) ContextContentType {
	diffIndicators := 0
	logIndicators := 0

	for _, line := range lines {
		if ccDiffPattern.MatchString(line) {
			diffIndicators++
		}
		if ccLogPattern.MatchString(line) || ccTimestampPattern.MatchString(line) {
			logIndicators++
		}
	}

	// Need at least 2 diff indicators to be confident
	if diffIndicators >= 2 {
		return ContextContentTypeDiff
	}
	// Need at least 3 log indicators
	if logIndicators >= 3 {
		return ContextContentTypeLog
	}

	return ContextContentTypeUnknown
}

// EstimateTokens provides token estimation for the filter.
func (c *ContextCrunchFilter) EstimateTokens(text string) int {
	return core.EstimateTokens(text)
}
