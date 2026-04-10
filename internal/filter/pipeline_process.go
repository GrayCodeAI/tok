package filter

import "github.com/GrayCodeAI/tokman/internal/core"

// Process runs the full compression pipeline with early-exit support.
// Stage gates skip layers when not applicable (zero cost).
// Skip remaining layers if budget already met.
func (p *PipelineCoordinator) Process(input string) (string, *PipelineStats) {
	stats := &PipelineStats{
		OriginalTokens: core.EstimateTokens(input),
		LayerStats:     make(map[string]LayerStat),
	}

	output := input

	// Pre-filters: TOML
	output = p.processPreFilters(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Core layers (1-9) + Neural
	output = p.processCoreLayers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Semantic layers (11-20)
	output = p.processSemanticLayers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Research layers (21-25)
	output = p.processResearchLayers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Budget enforcement
	output = p.processBudgetLayer(output, stats)

	// Post-compensation
	if p.smallKVCompensator != nil {
		output = p.smallKVCompensator.Compensate(input, output, p.config.Mode)
	}

	// Quality feedback
	p.recordFeedback(input, output, stats)

	if p.qualityGuardrail != nil {
		gr := p.qualityGuardrail.Validate(input, output)
		if !gr.Passed {
			safeOutput, safeStats := p.runGuardrailFallback(input)
			safeStats.LayerStats["guardrail_fallback"] = LayerStat{TokensSaved: 0}
			safeStats.LayerStats["guardrail_reason_"+gr.Reason] = LayerStat{TokensSaved: 0}
			return safeOutput, safeStats
		}
	}

	return output, p.finalizeStats(stats, output)
}

func (p *PipelineCoordinator) runGuardrailFallback(input string) (string, *PipelineStats) {
	fallbackCfg := p.config
	fallbackCfg.Mode = ModeMinimal
	fallbackCfg.EnableExtractivePrefilter = false
	fallbackCfg.EnableQualityGuardrail = false
	fallback := NewPipelineCoordinator(fallbackCfg)
	return fallback.Process(input)
}

func (p *PipelineCoordinator) processPreFilters(output string, stats *PipelineStats) string {
	// TOML Filter
	if p.tomlFilterWrapper != nil && p.config.EnableTOMLFilter {
		filtered, saved := p.tomlFilterWrapper.Apply(output, ModeMinimal)
		if saved > 0 {
			stats.LayerStats["0_toml_filter"] = LayerStat{TokensSaved: saved}
			output = filtered
			stats.TotalSaved += saved
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	return output
}

func (p *PipelineCoordinator) processCoreLayers(output string, stats *PipelineStats) string {
	if p.entropyFilter != nil && p.config.EnableEntropy && !p.shouldSkipEntropy(output) {
		output = p.processLayer(p.layers[0], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.perplexityFilter != nil && p.config.EnablePerplexity && !p.shouldSkipPerplexity(output) {
		output = p.processLayer(p.layers[1], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.goalDrivenFilter != nil && p.config.EnableGoalDriven && !p.shouldSkipQueryDependent() {
		output = p.processLayer(p.layers[2], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.astPreserveFilter != nil && p.config.EnableAST {
		output = p.processLayer(p.layers[3], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.contrastiveFilter != nil && p.config.EnableContrastive && !p.shouldSkipQueryDependent() {
		output = p.processLayer(p.layers[4], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.ngramAbbreviator != nil && !p.shouldSkipNgram(output) {
		output = p.processLayer(p.layers[5], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.evaluatorHeadsFilter != nil && p.config.EnableEvaluator {
		output = p.processLayer(p.layers[6], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.gistFilter != nil && p.config.EnableGist {
		output = p.processLayer(p.layers[7], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.hierarchicalSummaryFilter != nil && p.config.EnableHierarchical {
		output = p.processLayer(p.layers[8], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	return output
}

func (p *PipelineCoordinator) processSemanticLayers(output string, stats *PipelineStats) string {
	if p.compactionLayer != nil && !p.shouldSkipCompaction(output) {
		output = p.processLayer(p.layers[9], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.attributionFilter != nil {
		output = p.processLayer(p.layers[10], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.h2oFilter != nil && !p.shouldSkipH2O(output) {
		output = p.processLayer(p.layers[11], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.attentionSinkFilter != nil && !p.shouldSkipAttentionSink(output) {
		output = p.processLayer(p.layers[12], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.metaTokenFilter != nil && !p.shouldSkipMetaToken(output) {
		output = p.processLayer(p.layers[13], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticChunkFilter != nil && !p.shouldSkipSemanticChunk(output) {
		output = p.processLayer(p.layers[14], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.sketchStoreFilter != nil && !p.shouldSkipBudgetDependent() {
		output = p.processLayer(p.layers[15], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.lazyPrunerFilter != nil && !p.shouldSkipBudgetDependent() {
		output = p.processLayer(p.layers[16], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticAnchorFilter != nil {
		output = p.processLayer(p.layers[17], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.agentMemoryFilter != nil {
		output = p.processLayer(p.layers[18], output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processResearchLayers(output string, stats *PipelineStats) string {
	if p.marginalInfoGainFilter != nil {
		output = p.processLayer(p.layers[19], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.nearDedupFilter != nil {
		output = p.processLayer(p.layers[20], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.cotCompressFilter != nil {
		output = p.processLayer(p.layers[21], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.codingAgentCtxFilter != nil {
		output = p.processLayer(p.layers[22], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.perceptionCompressFilter != nil {
		output = p.processLayer(p.layers[23], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.lightThinkerFilter != nil {
		output = p.processLayer(p.layers[24], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.thinkSwitcherFilter != nil {
		output = p.processLayer(p.layers[25], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.gmsaFilter != nil {
		output = p.processLayer(p.layers[26], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.carlFilter != nil {
		output = p.processLayer(p.layers[27], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.slimInferFilter != nil {
		output = p.processLayer(p.layers[28], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.diffAdaptFilter != nil {
		output = p.processLayer(p.layers[29], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.epicFilter != nil {
		output = p.processLayer(p.layers[30], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.ssdpFilter != nil {
		output = p.processLayer(p.layers[31], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.agentOCRFilter != nil {
		output = p.processLayer(p.layers[32], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.s2madFilter != nil {
		output = p.processLayer(p.layers[33], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.aconFilter != nil {
		output = p.processLayer(p.layers[34], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.latentCollabFilter != nil {
		output = p.processLayer(p.layers[35], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.graphCoTFilter != nil {
		output = p.processLayer(p.layers[36], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}
	if p.roleBudgetFilter != nil {
		output = p.processLayer(p.layers[37], output, stats)
	}
	return output
}

func (p *PipelineCoordinator) recordFeedback(input, output string, stats *PipelineStats) {
	if p.feedback != nil && p.qualityEstimator != nil {
		quality := p.qualityEstimator.EstimateQuality(input, output)
		p.feedback.RecordSignal(FeedbackSignal{
			LayerName:           "pipeline",
			QualityScore:        quality,
			CompressionRatio:    stats.ReductionPercent / 100.0,
			SuggestedAdjustment: (quality - 0.8) * 0.5,
		})
	}
}
