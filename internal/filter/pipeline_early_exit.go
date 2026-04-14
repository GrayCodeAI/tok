package filter

import (
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// shouldEarlyExit returns true if budget is already met (T81).
// Uses aggressive checking when budget is tight (< TightBudgetThreshold tokens),
// otherwise checks every 3 layers to reduce overhead.
func (p *PipelineCoordinator) shouldEarlyExit(stats *PipelineStats) bool {
	if p.config.Budget <= 0 {
		return false
	}
	// Use aggressive checking when budget is tight
	if p.config.Budget < TightBudgetThreshold {
		currentTokens := stats.OriginalTokens - stats.computeTotalSaved()
		return currentTokens <= p.config.Budget
	}
	// Check every 3 layers to reduce overhead (optimization)
	if len(stats.LayerStats)%3 != 0 {
		return false
	}
	currentTokens := stats.OriginalTokens - stats.computeTotalSaved()
	return currentTokens <= p.config.Budget
}

// shouldSkipEntropy checks if entropy filtering would help.
func (p *PipelineCoordinator) shouldSkipEntropy(content string) bool {
	if len(content) < 50 {
		return true
	}
	limit := len(content)
	if limit > 500 {
		limit = 500
	}
	var seen [32]byte
	uniqueCount := 0
	for i := 0; i < limit; i++ {
		b := content[i]
		idx := b >> 3
		bit := byte(1) << (b & 7)
		if seen[idx]&bit == 0 {
			seen[idx] |= bit
			uniqueCount++
			if uniqueCount > 30 {
				return false
			}
		}
	}
	return true
}

// shouldSkipPerplexity checks if perplexity pruning would help.
func (p *PipelineCoordinator) shouldSkipPerplexity(content string) bool {
	nlCount := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			nlCount++
			if nlCount >= 5 {
				return false
			}
		}
	}
	return true
}

// shouldSkipQueryDependent checks if query-dependent layers apply.
func (p *PipelineCoordinator) shouldSkipQueryDependent() bool {
	return p.effectiveQueryIntent() == ""
}

func (p *PipelineCoordinator) effectiveQueryIntent() string {
	if p.runtimeQueryIntent != "" {
		return p.runtimeQueryIntent
	}
	return p.config.QueryIntent
}

// shouldSkipNgram checks if N-gram abbreviation would help.
func (p *PipelineCoordinator) shouldSkipNgram(content string) bool {
	if len(content) < 200 {
		return true
	}
	wordCount := 0
	inWord := false
	for i := 0; i < len(content); i++ {
		isSpace := content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r'
		if isSpace {
			inWord = false
		} else if !inWord {
			inWord = true
			wordCount++
			if wordCount >= 20 {
				return false
			}
		}
	}
	return true
}

// shouldSkipCompaction checks if compaction would help.
func (p *PipelineCoordinator) shouldSkipCompaction(content string) bool {
	conversationMarkers := []string{"User:", "Assistant:", "AI:", "Human:", "\n\n", ">>>"}
	for _, marker := range conversationMarkers {
		if strings.Contains(content, marker) {
			return false
		}
	}
	return true
}

// shouldSkipH2O checks if H2O heavy-hitter filtering would help.
func (p *PipelineCoordinator) shouldSkipH2O(content string) bool {
	tokens := EstimateTokens(content)
	return tokens < 50
}

// shouldSkipAttentionSink checks if attention sink filtering would help.
func (p *PipelineCoordinator) shouldSkipAttentionSink(content string) bool {
	lines := strings.Count(content, "\n")
	return lines < 3
}

// shouldSkipMetaToken checks if meta-token compression would help.
func (p *PipelineCoordinator) shouldSkipMetaToken(content string) bool {
	return len(content) < 500
}

// shouldSkipSemanticChunk checks if semantic chunking would help.
func (p *PipelineCoordinator) shouldSkipSemanticChunk(content string) bool {
	return len(content) < 300
}

// shouldSkipBudgetDependent checks if budget-dependent layers apply.
func (p *PipelineCoordinator) shouldSkipBudgetDependent() bool {
	return p.config.Budget <= 0
}

// computeTotalSaved returns total tokens saved across all layers.
func (s *PipelineStats) computeTotalSaved() int {
	return s.runningSaved
}

// finalizeStats computes final pipeline statistics.
func (p *PipelineCoordinator) finalizeStats(stats *PipelineStats, output string) *PipelineStats {
	stats.FinalTokens = core.EstimateTokens(output)

	// Safely calculate TotalSaved with overflow protection
	if stats.OriginalTokens >= stats.FinalTokens {
		stats.TotalSaved = stats.OriginalTokens - stats.FinalTokens
	} else {
		// If filtering somehow increased tokens, report 0 savings
		stats.TotalSaved = 0
	}

	// Safely calculate ReductionPercent with bounds checking
	if stats.OriginalTokens > 0 {
		stats.ReductionPercent = float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
		// Clamp to valid range [0, 100]
		if stats.ReductionPercent < 0 {
			stats.ReductionPercent = 0
		} else if stats.ReductionPercent > 100 {
			stats.ReductionPercent = 100
		}
	}
	return stats
}

// processLayer runs a single filter layer and records its stats.
// Uses layer cache when available to avoid redundant processing.
func (p *PipelineCoordinator) processLayer(layer filterLayer, input string, stats *PipelineStats) string {
	// Defensive nil checks
	if p == nil || stats == nil {
		return input
	}
	if layer.filter == nil {
		return input
	}

	if p.layerGate != nil && !p.layerGate.Allows(layer.name) {
		return input
	}

	// Check layer cache first (Phase 2 optimization)
	if p.layerCache != nil {
		if cached, hit := p.layerCache.Get(layer.name, input, p.config.Mode); hit {
			stats.LayerStats[layer.name] = LayerStat{
				TokensSaved: cached.TokensSaved,
				Duration:    0, // Cache hit = no processing time
			}
			stats.runningSaved += cached.TokensSaved
			return cached.Output
		}
	}

	start := time.Now()
	output, saved := layer.filter.Apply(input, p.config.Mode)
	dur := time.Since(start).Nanoseconds()

	stats.AddLayerStatSafe(layer.name, LayerStat{TokensSaved: saved, Duration: dur})
	// stats.runningSaved updated via AddLayerStatSafe

	// Cache the result for future use
	if p.layerCache != nil {
		p.layerCache.Put(layer.name, input, p.config.Mode, output, saved)
	}

	return output
}

// processBudgetLayer handles Layer 10: Budget Enforcement.
func (p *PipelineCoordinator) processBudgetLayer(input string, stats *PipelineStats) string {
	output := input
	totalSaved := 0

	if p.sessionTracker != nil {
		filtered, saved := p.sessionTracker.Apply(output, p.config.Mode)
		output = filtered
		totalSaved += saved
		stats.LayerStats["10_session"] = LayerStat{TokensSaved: saved}
	}

	if p.budgetEnforcer != nil {
		filtered, saved := p.budgetEnforcer.Apply(output, p.config.Mode)
		output = filtered
		totalSaved += saved
		stats.LayerStats["10_budget"] = LayerStat{TokensSaved: saved}
	}

	stats.LayerStats["10_total"] = LayerStat{TokensSaved: totalSaved}
	stats.runningSaved += totalSaved
	return output
}
