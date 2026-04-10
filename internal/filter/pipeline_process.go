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
	p.runtimeQueryIntent = ""
	p.applyPolicyRouting(input)

	output := input

	// Pre-filters: TOML and TF-IDF
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

	// NEW layers: Symbolic, Phrase, Numerical, Dynamic
	output = p.processNewLayers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Phase 2 layers
	output = p.processPhase2Layers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Planned 30-49 experimental layers (disabled by default).
	output = p.processPlannedLayers(output, stats)
	if p.shouldEarlyExit(stats) {
		return output, p.finalizeStats(stats, output)
	}

	// Recovery layers
	output = p.processRecoveryLayers(output, stats)

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
	// Extractive pre-filter for very large outputs.
	if p.extractivePrefilter != nil && (p.layerGate == nil || p.layerGate.Allows("pre_extractive")) {
		filtered, saved := p.extractivePrefilter.Apply(output)
		if saved > 0 {
			stats.LayerStats["pre_extractive"] = LayerStat{TokensSaved: saved}
			output = filtered
			stats.TotalSaved += saved
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

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

	// TF-IDF Coarse Pre-filter
	if p.tfidfFilter != nil {
		output = p.processLayer(filterLayer{p.tfidfFilter, "pre_tfidf"}, output, stats)
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

	if p.llmFilter != nil {
		output = p.processLayer(p.layers[9], output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processSemanticLayers(output string, stats *PipelineStats) string {
	if p.compactionLayer != nil && !p.shouldSkipCompaction(output) {
		output = p.processLayer(p.layers[10], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.attributionFilter != nil {
		output = p.processLayer(p.layers[11], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.h2oFilter != nil && !p.shouldSkipH2O(output) {
		output = p.processLayer(p.layers[12], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.attentionSinkFilter != nil && !p.shouldSkipAttentionSink(output) {
		output = p.processLayer(p.layers[13], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.metaTokenFilter != nil && !p.shouldSkipMetaToken(output) {
		output = p.processLayer(p.layers[14], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticChunkFilter != nil && !p.shouldSkipSemanticChunk(output) {
		output = p.processLayer(p.layers[15], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.sketchStoreFilter != nil && !p.shouldSkipBudgetDependent() {
		output = p.processLayer(p.layers[16], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.lazyPrunerFilter != nil && !p.shouldSkipBudgetDependent() {
		output = p.processLayer(p.layers[17], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticAnchorFilter != nil {
		output = p.processLayer(p.layers[18], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.agentMemoryFilter != nil {
		output = p.processLayer(p.layers[19], output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processNewLayers(output string, stats *PipelineStats) string {
	if p.symbolicCompressFilter != nil {
		output = p.processLayer(filterLayer{p.symbolicCompressFilter, "20_symbolic_compress"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.phraseGroupingFilter != nil {
		output = p.processLayer(filterLayer{p.phraseGroupingFilter, "21_phrase_grouping"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.numericalQuantizer != nil {
		output = p.processLayer(filterLayer{p.numericalQuantizer, "22_numerical_quant"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.dynamicRatioFilter != nil {
		output = p.processLayer(filterLayer{p.dynamicRatioFilter, "23_dynamic_ratio"}, output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processPhase2Layers(output string, stats *PipelineStats) string {
	if p.hypernymCompressor != nil {
		output = p.processLayer(filterLayer{p.hypernymCompressor, "24_hypernym"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticCacheFilter != nil {
		output = p.processLayer(filterLayer{p.semanticCacheFilter, "25_semantic_cache"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.scopeFilter != nil {
		output = p.processLayer(filterLayer{p.scopeFilter, "26_scope"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.kvzipFilter != nil {
		output = p.processLayer(filterLayer{p.kvzipFilter, "27_kvzip"}, output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processRecoveryLayers(output string, stats *PipelineStats) string {
	if p.questionAwareFilter != nil && !p.shouldSkipQueryDependent() {
		output = p.processLayer(filterLayer{p.questionAwareFilter, "28_question_aware"}, output, stats)
	}

	if p.densityAdaptiveFilter != nil && !p.shouldSkipSemanticChunk(output) {
		output = p.processLayer(filterLayer{p.densityAdaptiveFilter, "29_density_adaptive"}, output, stats)
	}
	return output
}

func (p *PipelineCoordinator) processPlannedLayers(output string, stats *PipelineStats) string {
	if len(p.plannedLayers) == 0 {
		return output
	}
	for _, layer := range p.plannedLayers {
		output = p.processLayer(layer, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
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
