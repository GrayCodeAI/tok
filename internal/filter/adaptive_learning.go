package filter

// AdaptiveLearningFilter combines Layer 50 (EngramLearner) and Layer 51 (TieredSummary)
// into a single post-processing layer that learns from patterns and generates
// progressive summaries.
//
// This merged layer:
// 1. First applies EngramLearner to detect and learn error patterns
// 2. Then applies TieredSummary for L0/L1/L2 progressive summarization
//
// Research basis:
// - EngramLearner: Error pattern learning with 14 classifiers
// - TieredSummary: Progressive summarization (surface → structural → deep)
type AdaptiveLearningFilter struct {
	engram *EngramLearner
	tiered *TieredSummaryFilter
	enabled bool
}

// NewAdaptiveLearningFilter creates a new adaptive learning filter.
// This replaces both NewEngramLearner() and NewTieredSummaryFilter().
func NewAdaptiveLearningFilter() *AdaptiveLearningFilter {
	return &AdaptiveLearningFilter{
		engram:  NewEngramLearner(),
		tiered:  NewTieredSummaryFilter(),
		enabled: true,
	}
}

// Name returns the filter name.
func (a *AdaptiveLearningFilter) Name() string { return "adaptive_learning" }

// Apply runs both learning and summarization in sequence.
func (a *AdaptiveLearningFilter) Apply(input string, mode Mode) (string, int) {
	if !a.enabled || mode == ModeNone {
		return input, 0
	}

	// Phase 1: Pattern learning (from EngramLearner)
	// Run engram learner (it analyzes but doesn't modify)
	_, engramSaved := a.engram.Apply(input, mode)

	// Phase 2: Progressive summarization (from TieredSummary)
	// This generates L0/L1/L2 tiered summaries
	output, tieredSaved := a.tiered.Apply(input, mode)

	// Calculate total savings
	totalSaved := engramSaved + tieredSaved

	return output, totalSaved
}

// GetLearningStats returns statistics from the learning component.
func (a *AdaptiveLearningFilter) GetLearningStats() map[string]interface{} {
	return a.engram.GetStats()
}

// GenerateTiers generates tiered summaries from input.
func (a *AdaptiveLearningFilter) GenerateTiers(input string) *TieredResult {
	return a.tiered.GenerateTiers(input)
}

// SetEnabled enables or disables the filter.
func (a *AdaptiveLearningFilter) SetEnabled(enabled bool) {
	a.enabled = enabled
}

// IsEnabled returns whether the filter is enabled.
func (a *AdaptiveLearningFilter) IsEnabled() bool {
	return a.enabled
}
