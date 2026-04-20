package filter

// Six-layer pipeline orchestration.
//
// All 50+ filter implementations are unchanged; this file only reorganizes
// how they are sequenced. Each layer has a single clear responsibility:
//
//   Layer 1 Preprocess   — noise removal before any analysis (TOML, routing, dedup)
//   Layer 2 Structural   — statistical signal/noise decisions (entropy, perplexity, AST)
//   Layer 3 Semantic     — meaning-level compression (gist, compaction, attribution)
//   Layer 4 LLM-Specific — KV-cache aware techniques (H2O, attention sink, anchoring)
//   Layer 5 ContentType  — format-specific passes (diff, log, JSON, agent memory)
//   Layer 6 Budget       — hard enforcement (budget, session tracking)

// runLayer1Preprocess applies TOML filters, adaptive routing, extractive prefilter,
// and adaptive learning (session-learned patterns applied earliest for best ROI).
func (p *PipelineCoordinator) runLayer1Preprocess(input string, stats *PipelineStats) string {
	output := p.applyAdaptiveRouting(input, stats)
	if p.shouldEarlyExit(stats) {
		return output
	}

	if p.tomlFilterWrapper != nil && p.config.EnableTOMLFilter {
		filtered, saved := p.tomlFilterWrapper.Apply(output, ModeMinimal)
		if saved > 0 {
			stats.LayerStats["0_toml_filter"] = LayerStat{TokensSaved: saved}
			stats.runningSaved += saved
			output = filtered
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.adaptiveLearning != nil && p.config.EnableAdaptiveLearning && len(output) > 1000 {
		output = p.processLayer(filterLayer{p.adaptiveLearning, "1_adaptive_learning"}, output, stats)
	}

	return output
}

// runLayer2Structural removes statistically low-value content: entropy pruning,
// perplexity scoring, AST preservation, n-gram abbreviation.
func (p *PipelineCoordinator) runLayer2Structural(input string, stats *PipelineStats) string {
	output := input

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
	}

	return output
}

// runLayer3Semantic compresses at the meaning level: evaluator heads, gist extraction,
// hierarchical summarization, conversation compaction, attribution pruning,
// meta-token lossless encoding, and semantic chunking.
func (p *PipelineCoordinator) runLayer3Semantic(input string, stats *PipelineStats) string {
	output := input

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

	if p.metaTokenFilter != nil && !p.shouldSkipMetaToken(output) {
		output = p.processLayer(p.layers[13], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.semanticChunkFilter != nil && !p.shouldSkipSemanticChunk(output) {
		output = p.processLayer(p.layers[14], output, stats)
	}

	return output
}

// runLayer4LLMSpecific applies KV-cache aware techniques: QuantumLock alignment,
// H2O heavy-hitter eviction, attention sink preservation, sketch store,
// lazy pruning, and semantic anchor compression.
func (p *PipelineCoordinator) runLayer4LLMSpecific(input string, stats *PipelineStats) string {
	output := input

	if p.quantumLockFilter != nil && p.config.EnableQuantumLock {
		output = p.processLayer(filterLayer{p.quantumLockFilter, "4_quantum_lock"}, output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.photonFilter != nil && p.config.EnablePhoton {
		output = p.processLayer(filterLayer{p.photonFilter, "4_photon"}, output, stats)
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
	}

	return output
}

// runLayer5ContentType applies content-format aware passes: agent memory consolidation,
// edge-case handling (L21-25), reasoning trace compression (L26-30),
// and advanced research techniques (L31-45: diff, log, JSON, search, structural collapse).
func (p *PipelineCoordinator) runLayer5ContentType(input string, stats *PipelineStats) string {
	output := input

	if p.agentMemoryFilter != nil {
		output = p.processLayer(p.layers[18], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.edgeCaseFilter != nil {
		output = p.processLayer(p.layers[19], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.reasoningFilter != nil {
		output = p.processLayer(p.layers[20], output, stats)
		if p.shouldEarlyExit(stats) {
			return output
		}
	}

	if p.advancedFilter != nil {
		output = p.processLayer(p.layers[21], output, stats)
	}

	return output
}

// runLayer6BudgetQuality enforces token budget and session tracking.
// Quality guardrail and feedback are handled in Process() after all layers complete.
func (p *PipelineCoordinator) runLayer6BudgetQuality(input string, stats *PipelineStats) string {
	return p.processBudgetLayer(input, stats)
}
