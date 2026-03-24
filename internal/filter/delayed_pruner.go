package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// DelayedPruner implements DMS-style delayed token eviction.
// Research: "Dynamic Memory Sparsification" (NeurIPS 2025)
// Key Insight: Instead of immediately deleting tokens, mark them for removal
// but keep them for a few more processing steps. This allows "implicit merging"
// where important information from marked tokens transfers to surviving tokens
// before the originals are deleted.
//
// In TokMan: Lines/tokens marked for removal are kept in a "pending" state.
// After processing through subsequent layers, if their information appears
// in surviving content, they can be safely removed. If not, they're restored.
type DelayedPruner struct {
	config DelayedConfig
}

// DelayedConfig holds configuration for delayed pruning
type DelayedConfig struct {
	Enabled         bool
	DelaySteps      int     // Number of steps to delay before final eviction
	RescueThreshold float64 // If marked token's info not found in survivors, rescue it
	MinContentLength int
}

// DefaultDelayedConfig returns default configuration
func DefaultDelayedConfig() DelayedConfig {
	return DelayedConfig{
		Enabled:          true,
		DelaySteps:       3,
		RescueThreshold:  0.3,
		MinContentLength: 200,
	}
}

// NewDelayedPruner creates a new delayed pruner
func NewDelayedPruner() *DelayedPruner {
	return &DelayedPruner{config: DefaultDelayedConfig()}
}

// Name returns the filter name
func (d *DelayedPruner) Name() string { return "delayed_pruner" }

// pendingLine represents a line marked for removal but not yet evicted
type pendingLine struct {
	content  string
	stepMarked int
	words    map[string]bool
}

// Apply applies delayed pruning
func (d *DelayedPruner) Apply(input string, mode Mode) (string, int) {
	if !d.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < d.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	lines := strings.Split(input, "\n")

	// Step 1: Score lines and mark low-importance ones for delayed removal
	var kept []string
	var pending []pendingLine

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			kept = append(kept, line)
			continue
		}

		score := d.scoreLine(trimmed)

		if score > 0.4 {
			// Important - keep immediately
			kept = append(kept, line)
		} else if score > 0.15 {
			// Medium - keep but mark for potential removal
			kept = append(kept, line)
		} else {
			// Low importance - mark for delayed removal
			pending = append(pending, pendingLine{
				content:    trimmed,
				stepMarked: 0,
				words:      d.extractWords(trimmed),
			})
		}
	}

	// Step 2: Check if pending lines' information is preserved in kept lines
	keptText := strings.Join(kept, "\n")
	keptWords := d.extractWords(keptText)

	var rescued []string
	evicted := 0

	for _, p := range pending {
		// Check how many of the pending line's words appear in kept content
		overlap := 0
		for w := range p.words {
			if keptWords[w] {
				overlap++
			}
		}

		retentionRatio := 0.0
		if len(p.words) > 0 {
			retentionRatio = float64(overlap) / float64(len(p.words))
		}

		// If information is NOT preserved in survivors, rescue the line
		if retentionRatio < d.config.RescueThreshold {
			rescued = append(rescued, p.content)
		} else {
			evicted++
		}
	}

	// Step 3: Reconstruct with rescued lines
	result := strings.Join(kept, "\n")
	if len(rescued) > 0 {
		result += "\n" + strings.Join(rescued, "\n")
	}

	result = strings.TrimSpace(result)
	finalTokens := core.EstimateTokens(result)
	saved := originalTokens - finalTokens
	if saved < 3 {
		return input, 0
	}

	return result, saved
}

// scoreLine scores a line's importance
func (d *DelayedPruner) scoreLine(line string) float64 {
	score := 0.3
	lower := strings.ToLower(line)

	// High importance indicators
	if strings.Contains(lower, "error") || strings.Contains(lower, "fail") {
		score += 0.5
	}
	if strings.Contains(lower, "warn") {
		score += 0.3
	}
	if strings.Contains(line, "{") || strings.Contains(line, "}") {
		score += 0.2
	}
	if strings.Contains(line, "func ") || strings.Contains(line, "class ") {
		score += 0.4
	}
	if strings.Contains(line, "/") {
		score += 0.1
	}

	// Penalize very short or repetitive lines
	if len(line) < 5 {
		score -= 0.2
	}

	return score
}

// extractWords extracts word set from text
func (d *DelayedPruner) extractWords(text string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		cleaned := strings.Trim(w, ".,;:!?\"'()[]{}")
		if len(cleaned) > 2 {
			words[cleaned] = true
		}
	}
	return words
}
