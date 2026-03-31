package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// shouldEarlyExit returns true if budget is already met (T81).
func (p *PipelineCoordinator) shouldEarlyExit(stats *PipelineStats) bool {
	if p.config.Budget <= 0 {
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
	return p.config.QueryIntent == ""
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
	stats.TotalSaved = stats.OriginalTokens - stats.FinalTokens
	if stats.OriginalTokens > 0 {
		stats.ReductionPercent = float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
	}
	return stats
}

// processLayer runs a single filter layer and records its stats.
func (p *PipelineCoordinator) processLayer(layer filterLayer, input string, stats *PipelineStats) string {
	output, saved := layer.filter.Apply(input, p.config.Mode)
	if p.config.SessionTracking {
		stats.LayerStats[layer.name] = LayerStat{TokensSaved: saved, Duration: 0}
	} else {
		stats.LayerStats[layer.name] = LayerStat{TokensSaved: saved}
	}
	stats.runningSaved += saved
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
