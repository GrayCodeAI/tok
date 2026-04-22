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

// safeLayer returns the filterLayer at index i if it exists, otherwise a zero value.
func (p *PipelineCoordinator) safeLayer(i int) (filterLayer, bool) {
	if p == nil || i < 0 || i >= len(p.layers) {
		return filterLayer{}, false
	}
	return p.layers[i], true
}

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
			stats.AddLayerStatSafe(LayerTOMLFilter, LayerStat{TokensSaved: saved})
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
		if l, ok := p.safeLayer(0); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.perplexityFilter != nil && p.config.EnablePerplexity && !p.shouldSkipPerplexity(output) {
		if l, ok := p.safeLayer(1); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.goalDrivenFilter != nil && p.config.EnableGoalDriven && !p.shouldSkipQueryDependent() {
		if l, ok := p.safeLayer(2); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.astPreserveFilter != nil && p.config.EnableAST {
		if l, ok := p.safeLayer(3); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.contrastiveFilter != nil && p.config.EnableContrastive && !p.shouldSkipQueryDependent() {
		if l, ok := p.safeLayer(4); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.ngramAbbreviator != nil && !p.shouldSkipNgram(output) {
		if l, ok := p.safeLayer(5); ok {
			output = p.processLayer(l, output, stats)
		}
	}

	return output
}

// runLayer3Semantic compresses at the meaning level: evaluator heads, gist extraction,
// hierarchical summarization, conversation compaction, attribution pruning,
// meta-token lossless encoding, and semantic chunking.
func (p *PipelineCoordinator) runLayer3Semantic(input string, stats *PipelineStats) string {
	output := input

	if p.evaluatorHeadsFilter != nil && p.config.EnableEvaluator {
		if l, ok := p.safeLayer(6); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.gistFilter != nil && p.config.EnableGist {
		if l, ok := p.safeLayer(7); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.hierarchicalSummaryFilter != nil && p.config.EnableHierarchical {
		if l, ok := p.safeLayer(8); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.compactionLayer != nil && !p.shouldSkipCompaction(output) {
		if l, ok := p.safeLayer(9); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.attributionFilter != nil {
		if l, ok := p.safeLayer(10); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.metaTokenFilter != nil && !p.shouldSkipMetaToken(output) {
		if l, ok := p.safeLayer(13); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.semanticChunkFilter != nil && !p.shouldSkipSemanticChunk(output) {
		if l, ok := p.safeLayer(14); ok {
			output = p.processLayer(l, output, stats)
		}
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
		if l, ok := p.safeLayer(11); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.attentionSinkFilter != nil && !p.shouldSkipAttentionSink(output) {
		if l, ok := p.safeLayer(12); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.sketchStoreFilter != nil && !p.shouldSkipBudgetDependent() {
		if l, ok := p.safeLayer(15); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.lazyPrunerFilter != nil && !p.shouldSkipBudgetDependent() {
		if l, ok := p.safeLayer(16); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.semanticAnchorFilter != nil {
		if l, ok := p.safeLayer(17); ok {
			output = p.processLayer(l, output, stats)
		}
	}

	return output
}

// runLayer5ContentType applies content-format aware passes: agent memory consolidation,
// edge-case handling (L21-25), reasoning trace compression (L26-30),
// and advanced research techniques (L31-45: diff, log, JSON, search, structural collapse).
func (p *PipelineCoordinator) runLayer5ContentType(input string, stats *PipelineStats) string {
	output := input

	if p.agentMemoryFilter != nil {
		if l, ok := p.safeLayer(18); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.edgeCaseFilter != nil {
		if l, ok := p.safeLayer(19); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.reasoningFilter != nil {
		if l, ok := p.safeLayer(20); ok {
			output = p.processLayer(l, output, stats)
			if p.shouldEarlyExit(stats) {
				return output
			}
		}
	}

	if p.advancedFilter != nil {
		if l, ok := p.safeLayer(21); ok {
			output = p.processLayer(l, output, stats)
		}
	}

	return output
}

// runLayer6BudgetQuality enforces token budget and session tracking.
// Quality guardrail and feedback are handled in Process() after all layers complete.
func (p *PipelineCoordinator) runLayer6BudgetQuality(input string, stats *PipelineStats) string {
	return p.processBudgetLayer(input, stats)
}
