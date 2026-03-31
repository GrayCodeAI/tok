package filter

import (
	"github.com/GrayCodeAI/tokman/internal/cache"
)

// NewPipelineCoordinator creates a new pipeline coordinator with all configured layers.
func NewPipelineCoordinator(cfg PipelineConfig) *PipelineCoordinator {
	p := &PipelineCoordinator{
		config:       cfg,
		resultCache:  cache.GetGlobalCache(),
		cacheEnabled: true,
	}

	// Set defaults - all layers enabled by default when using zero-config.
	allDisabled := !cfg.EnableEntropy && !cfg.EnablePerplexity && !cfg.EnableGoalDriven &&
		!cfg.EnableAST && !cfg.EnableContrastive && !cfg.EnableEvaluator &&
		!cfg.EnableGist && !cfg.EnableHierarchical
	hasExplicitSettings := cfg.Budget > 0 || cfg.QueryIntent != "" || cfg.LLMEnabled ||
		cfg.NgramEnabled || cfg.MultiFileEnabled || cfg.SessionTracking ||
		cfg.EnableCompaction || cfg.EnableAttribution || cfg.EnableH2O || cfg.EnableAttentionSink
	if allDisabled && !hasExplicitSettings {
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableAST = true
		cfg.EnableContrastive = true
		cfg.EnableEvaluator = true
		cfg.EnableGist = true
		cfg.EnableHierarchical = true
	}

	// Core filters (Layers 1-9)
	p.initCoreFilters(cfg)

	// Neural layer (optional)
	p.initNeuralLayer(cfg)

	// Semantic filters (Layers 11-20)
	p.initSemanticFilters(cfg)

	// Adaptive filters (T12, T17)
	p.initAdaptiveFilters(cfg)

	// NEW filters (TF-IDF, Symbolic, Phrase, Numerical, Dynamic)
	p.initNewFilters(cfg)

	// Feedback mechanism
	p.feedback = NewInterLayerFeedback()
	p.qualityEstimator = NewQualityEstimator()

	// Phase 2 filters
	p.initPhase2Filters(cfg)

	// Build layer execution order
	p.buildLayers()

	return p
}

func (p *PipelineCoordinator) initCoreFilters(cfg PipelineConfig) {
	p.entropyFilter = NewEntropyFilter()
	p.perplexityFilter = NewPerplexityFilter()

	if cfg.QueryIntent != "" {
		p.goalDrivenFilter = NewGoalDrivenFilter(cfg.QueryIntent)
	}

	p.astPreserveFilter = NewASTPreserveFilter()

	if cfg.QueryIntent != "" {
		p.contrastiveFilter = NewContrastiveFilter(cfg.QueryIntent)
	}

	if cfg.NgramEnabled {
		p.ngramAbbreviator = NewNgramAbbreviator()
	}

	p.evaluatorHeadsFilter = NewEvaluatorHeadsFilter()
	p.gistFilter = NewGistFilter()
	p.hierarchicalSummaryFilter = NewHierarchicalSummaryFilter()

	if cfg.Budget > 0 {
		p.budgetEnforcer = NewBudgetEnforcer(cfg.Budget)
	}
	if cfg.SessionTracking {
		p.sessionTracker = NewSessionTracker()
	}
}

func (p *PipelineCoordinator) initNeuralLayer(cfg PipelineConfig) {
	if cfg.LLMEnabled {
		p.llmFilter = NewLLMAwareFilter(LLMAwareConfig{
			Threshold:      2000,
			Enabled:        true,
			CacheEnabled:   true,
			PromptTemplate: cfg.PromptTemplate,
		})
	}
}

func (p *PipelineCoordinator) initSemanticFilters(cfg PipelineConfig) {
	if cfg.EnableCompaction {
		compactionCfg := CompactionConfig{
			Enabled:             true,
			ThresholdTokens:     cfg.CompactionThreshold,
			PreserveRecentTurns: cfg.CompactionPreserveTurns,
			MaxSummaryTokens:    cfg.CompactionMaxTokens,
			StateSnapshotFormat: cfg.CompactionStateSnapshot,
			AutoDetect:          cfg.CompactionAutoDetect,
			CacheEnabled:        true,
		}
		if compactionCfg.ThresholdTokens == 0 {
			compactionCfg.ThresholdTokens = 2000
		}
		if compactionCfg.PreserveRecentTurns == 0 {
			compactionCfg.PreserveRecentTurns = 5
		}
		if compactionCfg.MaxSummaryTokens == 0 {
			compactionCfg.MaxSummaryTokens = 500
		}
		p.compactionLayer = NewCompactionLayer(compactionCfg)
	}

	if cfg.EnableAttribution {
		p.attributionFilter = NewAttributionFilter()
		if cfg.AttributionThreshold > 0 {
			p.attributionFilter.config.ImportanceThreshold = cfg.AttributionThreshold
		}
	}

	if cfg.EnableH2O {
		p.h2oFilter = NewH2OFilter()
		if cfg.H2OSinkSize > 0 {
			p.h2oFilter.config.SinkSize = cfg.H2OSinkSize
		}
		if cfg.H2ORecentSize > 0 {
			p.h2oFilter.config.RecentSize = cfg.H2ORecentSize
		}
		if cfg.H2OHeavyHitterSize > 0 {
			p.h2oFilter.config.HeavyHitterSize = cfg.H2OHeavyHitterSize
		}
	}

	if cfg.EnableAttentionSink {
		p.attentionSinkFilter = NewAdaptiveAttentionSinkFilter(50)
		if cfg.AttentionSinkCount > 0 {
			p.attentionSinkFilter.config.SinkTokenCount = cfg.AttentionSinkCount
		}
		if cfg.AttentionRecentCount > 0 {
			p.attentionSinkFilter.config.RecentTokenCount = cfg.AttentionRecentCount
		}
	}

	if cfg.EnableMetaToken {
		metaCfg := DefaultMetaTokenConfig()
		if cfg.MetaTokenWindow > 0 {
			metaCfg.WindowSize = cfg.MetaTokenWindow
		}
		if cfg.MetaTokenMinSize > 0 {
			metaCfg.MinPattern = cfg.MetaTokenMinSize
		}
		p.metaTokenFilter = NewMetaTokenFilterWithConfig(metaCfg)
	}

	if cfg.EnableSemanticChunk {
		semanticCfg := DefaultSemanticChunkConfig()
		if cfg.SemanticChunkMinSize > 0 {
			semanticCfg.MinChunkSize = cfg.SemanticChunkMinSize
		}
		if cfg.SemanticChunkThreshold > 0 {
			semanticCfg.ImportanceThreshold = cfg.SemanticChunkThreshold
		}
		p.semanticChunkFilter = NewSemanticChunkFilterWithConfig(semanticCfg)
	}

	if cfg.EnableSketchStore {
		sketchCfg := DefaultSketchStoreConfig()
		if cfg.SketchBudgetRatio > 0 {
			sketchCfg.BudgetRatio = cfg.SketchBudgetRatio
		}
		if cfg.SketchMaxSize > 0 {
			sketchCfg.MaxSketchSize = cfg.SketchMaxSize
		}
		if cfg.SketchHeavyHitter > 0 {
			sketchCfg.HeavyHitterRatio = cfg.SketchHeavyHitter
		}
		p.sketchStoreFilter = NewSketchStoreFilterWithConfig(sketchCfg)
	}

	if cfg.EnableLazyPruner {
		lazyCfg := DefaultLazyPrunerConfig()
		if cfg.LazyBaseBudget > 0 {
			lazyCfg.BaseBudget = cfg.LazyBaseBudget
		}
		if cfg.LazyDecayRate > 0 {
			lazyCfg.DecayRate = cfg.LazyDecayRate
		}
		if cfg.LazyRevivalBudget > 0 {
			lazyCfg.RevivalBudget = cfg.LazyRevivalBudget
		}
		p.lazyPrunerFilter = NewLazyPrunerFilterWithConfig(lazyCfg)
	}

	if cfg.EnableSemanticAnchor {
		anchorCfg := DefaultSemanticAnchorConfig()
		if cfg.SemanticAnchorRatio > 0 {
			anchorCfg.AnchorRatio = cfg.SemanticAnchorRatio
		}
		if cfg.SemanticAnchorSpacing > 0 {
			anchorCfg.MinAnchorSpacing = cfg.SemanticAnchorSpacing
		}
		p.semanticAnchorFilter = NewSemanticAnchorFilterWithConfig(anchorCfg)
	}

	if cfg.EnableAgentMemory {
		agentCfg := DefaultAgentMemoryConfig()
		if cfg.AgentKnowledgeRetention > 0 {
			agentCfg.KnowledgeRetentionRatio = cfg.AgentKnowledgeRetention
		}
		if cfg.AgentHistoryPrune > 0 {
			agentCfg.HistoryPruneRatio = cfg.AgentHistoryPrune
		}
		if cfg.AgentConsolidationMax > 0 {
			agentCfg.KnowledgeMaxSize = cfg.AgentConsolidationMax
		}
		p.agentMemoryFilter = NewAgentMemoryFilterWithConfig(agentCfg)
	}
}

func (p *PipelineCoordinator) initAdaptiveFilters(cfg PipelineConfig) {
	if cfg.EnableQuestionAware && cfg.QueryIntent != "" {
		p.questionAwareFilter = NewQuestionAwareFilter(cfg.QueryIntent)
		if cfg.QuestionAwareThreshold > 0 {
			p.questionAwareFilter.config.RelevanceThreshold = cfg.QuestionAwareThreshold
		}
	}

	if cfg.EnableDensityAdaptive {
		p.densityAdaptiveFilter = NewDensityAdaptiveFilter()
		if cfg.DensityTargetRatio > 0 {
			p.densityAdaptiveFilter.config.TargetRatio = cfg.DensityTargetRatio
		}
		if cfg.DensityThreshold > 0 {
			p.densityAdaptiveFilter.config.DensityThreshold = cfg.DensityThreshold
		}
	}
}

func (p *PipelineCoordinator) initNewFilters(cfg PipelineConfig) {
	if cfg.EnableTFIDF {
		tfidfCfg := DefaultTFIDFConfig()
		if cfg.TFIDFThreshold > 0 {
			tfidfCfg.Threshold = cfg.TFIDFThreshold
		}
		p.tfidfFilter = NewTFIDFFilterWithConfig(tfidfCfg)
	}

	if cfg.EnableSymbolicCompress {
		p.symbolicCompressFilter = NewSymbolicCompressFilter()
	}

	if cfg.EnablePhraseGrouping {
		p.phraseGroupingFilter = NewPhraseGroupingFilter()
	}

	if cfg.EnableNumericalQuant {
		numCfg := DefaultNumericalConfig()
		if cfg.DecimalPlaces > 0 {
			numCfg.DecimalPlaces = cfg.DecimalPlaces
		}
		p.numericalQuantizer = NewNumericalQuantizer()
		p.numericalQuantizer.config = numCfg
	}

	if cfg.EnableDynamicRatio {
		dynCfg := DefaultDynamicRatioConfig()
		if cfg.DynamicRatioBase > 0 {
			dynCfg.BaseBudgetRatio = cfg.DynamicRatioBase
		}
		p.dynamicRatioFilter = NewDynamicRatioFilter()
		p.dynamicRatioFilter.config = dynCfg
	}
}

func (p *PipelineCoordinator) initPhase2Filters(cfg PipelineConfig) {
	if cfg.EnableHypernym {
		p.hypernymCompressor = NewHypernymCompressor()
	}

	if cfg.EnableSemanticCache {
		p.semanticCacheFilter = NewSemanticCacheFilter()
	}

	if cfg.EnableScope {
		p.scopeFilter = NewScopeFilter()
	}

	if cfg.EnableSmallKV {
		p.smallKVCompensator = NewSmallKVCompensator()
	}

	if cfg.EnableKVzip {
		p.kvzipFilter = NewKVzipFilter()
	}
}

func (p *PipelineCoordinator) buildLayers() {
	p.layers = []filterLayer{
		{p.entropyFilter, "1_entropy"},
		{p.perplexityFilter, "2_perplexity"},
		{p.goalDrivenFilter, "3_goal_driven"},
		{p.astPreserveFilter, "4_ast_preserve"},
		{p.contrastiveFilter, "5_contrastive"},
		{p.ngramAbbreviator, "6_ngram"},
		{p.evaluatorHeadsFilter, "7_evaluator"},
		{p.gistFilter, "8_gist"},
		{p.hierarchicalSummaryFilter, "9_hierarchical"},
		{p.llmFilter, "neural"},
		{p.compactionLayer, "11_compaction"},
		{p.attributionFilter, "12_attribution"},
		{p.h2oFilter, "13_h2o"},
		{p.attentionSinkFilter, "14_attention_sink"},
		{p.metaTokenFilter, "15_meta_token"},
		{p.semanticChunkFilter, "16_semantic_chunk"},
		{p.sketchStoreFilter, "17_sketch_store"},
		{p.lazyPrunerFilter, "18_lazy_pruner"},
		{p.semanticAnchorFilter, "19_semantic_anchor"},
		{p.agentMemoryFilter, "20_agent_memory"},
		{p.questionAwareFilter, "21_question_aware"},
		{p.densityAdaptiveFilter, "22_density_adaptive"},
	}
}
