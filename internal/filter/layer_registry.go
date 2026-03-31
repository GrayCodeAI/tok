package filter

// LayerDescriptor describes a compression layer.
type LayerDescriptor struct {
	Name       string
	Priority   int                                             // execution order (lower = earlier)
	Group      string                                          // "core", "semantic", "new", "phase2", "recovery"
	Enabled    func(cfg PipelineConfig) bool                   // whether layer is enabled
	ShouldSkip func(p *PipelineCoordinator, input string) bool // stage gate
	Factory    func(cfg PipelineConfig) Filter                 // creates the filter
}

// LayerRegistry holds registered layer descriptors.
type LayerRegistry struct {
	layers []LayerDescriptor
}

// NewLayerRegistry creates a new empty layer registry.
func NewLayerRegistry() *LayerRegistry {
	return &LayerRegistry{
		layers: make([]LayerDescriptor, 0),
	}
}

// Register adds a layer descriptor to the registry.
func (r *LayerRegistry) Register(desc LayerDescriptor) {
	r.layers = append(r.layers, desc)
}

// GetLayers returns all registered layers for a specific group, sorted by priority.
func (r *LayerRegistry) GetLayers(group string) []LayerDescriptor {
	var result []LayerDescriptor
	for _, l := range r.layers {
		if l.Group == group {
			result = append(result, l)
		}
	}
	// Sort by priority (simple insertion sort for small slices)
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j].Priority > key.Priority {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}
	return result
}

// All returns all registered layers sorted by priority.
func (r *LayerRegistry) All() []LayerDescriptor {
	result := make([]LayerDescriptor, len(r.layers))
	copy(result, r.layers)
	// Sort by priority
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j].Priority > key.Priority {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}
	return result
}

// Global registry for layer self-registration.
var globalRegistry = NewLayerRegistry()

// RegisterLayer adds a layer to the global registry.
func RegisterLayer(desc LayerDescriptor) {
	globalRegistry.Register(desc)
}

// GetRegistry returns the global layer registry.
func GetRegistry() *LayerRegistry {
	return globalRegistry
}

// ============================================================================
// Layer Registrations
// ============================================================================

func init() {
	// --- Core Layers (1-10) ---

	RegisterLayer(LayerDescriptor{
		Name:     "entropy",
		Priority: 1,
		Group:    "core",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableEntropy },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipEntropy(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			return NewEntropyFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "perplexity",
		Priority: 2,
		Group:    "core",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnablePerplexity },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipPerplexity(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			return NewPerplexityFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "goal_driven",
		Priority: 3,
		Group:    "core",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableGoalDriven && cfg.QueryIntent != "" },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipQueryDependent()
		},
		Factory: func(cfg PipelineConfig) Filter {
			return NewGoalDrivenFilter(cfg.QueryIntent)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "ast_preserve",
		Priority:   4,
		Group:      "core",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableAST },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewASTPreserveFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "contrastive",
		Priority: 5,
		Group:    "core",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableContrastive && cfg.QueryIntent != "" },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipQueryDependent()
		},
		Factory: func(cfg PipelineConfig) Filter {
			return NewContrastiveFilter(cfg.QueryIntent)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "ngram",
		Priority: 6,
		Group:    "core",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.NgramEnabled },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipNgram(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			return NewNgramAbbreviator()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "evaluator",
		Priority:   7,
		Group:      "core",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableEvaluator },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewEvaluatorHeadsFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "gist",
		Priority:   8,
		Group:      "core",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableGist },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewGistFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "hierarchical",
		Priority:   9,
		Group:      "core",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableHierarchical },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewHierarchicalSummaryFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "llm",
		Priority:   10,
		Group:      "core",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.LLMEnabled },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewLLMAwareFilter(LLMAwareConfig{
				Threshold:      2000,
				Enabled:        true,
				CacheEnabled:   true,
				PromptTemplate: cfg.PromptTemplate,
			})
		},
	})

	// --- Semantic Layers (11-20) ---

	RegisterLayer(LayerDescriptor{
		Name:     "compaction",
		Priority: 11,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableCompaction },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipCompaction(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			threshold := cfg.CompactionThreshold
			if threshold == 0 {
				threshold = 2000
			}
			preserveTurns := cfg.CompactionPreserveTurns
			if preserveTurns == 0 {
				preserveTurns = 5
			}
			maxTokens := cfg.CompactionMaxTokens
			if maxTokens == 0 {
				maxTokens = 500
			}
			return NewCompactionLayer(CompactionConfig{
				Enabled:             true,
				ThresholdTokens:     threshold,
				PreserveRecentTurns: preserveTurns,
				MaxSummaryTokens:    maxTokens,
				StateSnapshotFormat: cfg.CompactionStateSnapshot,
				AutoDetect:          cfg.CompactionAutoDetect,
				CacheEnabled:        true,
			})
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "attribution",
		Priority:   12,
		Group:      "semantic",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableAttribution },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			f := NewAttributionFilter()
			if cfg.AttributionThreshold > 0 {
				f.config.ImportanceThreshold = cfg.AttributionThreshold
			}
			return f
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "h2o",
		Priority: 13,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableH2O },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipH2O(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			f := NewH2OFilter()
			if cfg.H2OSinkSize > 0 {
				f.config.SinkSize = cfg.H2OSinkSize
			}
			if cfg.H2ORecentSize > 0 {
				f.config.RecentSize = cfg.H2ORecentSize
			}
			if cfg.H2OHeavyHitterSize > 0 {
				f.config.HeavyHitterSize = cfg.H2OHeavyHitterSize
			}
			return f
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "attention_sink",
		Priority: 14,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableAttentionSink },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipAttentionSink(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			f := NewAdaptiveAttentionSinkFilter(50)
			if cfg.AttentionSinkCount > 0 {
				f.config.SinkTokenCount = cfg.AttentionSinkCount
			}
			if cfg.AttentionRecentCount > 0 {
				f.config.RecentTokenCount = cfg.AttentionRecentCount
			}
			return f
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "meta_token",
		Priority: 15,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableMetaToken },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipMetaToken(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			metaCfg := DefaultMetaTokenConfig()
			if cfg.MetaTokenWindow > 0 {
				metaCfg.WindowSize = cfg.MetaTokenWindow
			}
			if cfg.MetaTokenMinSize > 0 {
				metaCfg.MinPattern = cfg.MetaTokenMinSize
			}
			return NewMetaTokenFilterWithConfig(metaCfg)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "semantic_chunk",
		Priority: 16,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableSemanticChunk },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipSemanticChunk(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			semanticCfg := DefaultSemanticChunkConfig()
			if cfg.SemanticChunkMinSize > 0 {
				semanticCfg.MinChunkSize = cfg.SemanticChunkMinSize
			}
			if cfg.SemanticChunkThreshold > 0 {
				semanticCfg.ImportanceThreshold = cfg.SemanticChunkThreshold
			}
			return NewSemanticChunkFilterWithConfig(semanticCfg)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "sketch_store",
		Priority: 17,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableSketchStore },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipBudgetDependent()
		},
		Factory: func(cfg PipelineConfig) Filter {
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
			return NewSketchStoreFilterWithConfig(sketchCfg)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "lazy_pruner",
		Priority: 18,
		Group:    "semantic",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableLazyPruner },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipBudgetDependent()
		},
		Factory: func(cfg PipelineConfig) Filter {
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
			return NewLazyPrunerFilterWithConfig(lazyCfg)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "semantic_anchor",
		Priority:   19,
		Group:      "semantic",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableSemanticAnchor },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			anchorCfg := DefaultSemanticAnchorConfig()
			if cfg.SemanticAnchorRatio > 0 {
				anchorCfg.AnchorRatio = cfg.SemanticAnchorRatio
			}
			if cfg.SemanticAnchorSpacing > 0 {
				anchorCfg.MinAnchorSpacing = cfg.SemanticAnchorSpacing
			}
			return NewSemanticAnchorFilterWithConfig(anchorCfg)
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "agent_memory",
		Priority:   20,
		Group:      "semantic",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableAgentMemory },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
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
			return NewAgentMemoryFilterWithConfig(agentCfg)
		},
	})

	// --- Recovery Layers (21-22) ---

	RegisterLayer(LayerDescriptor{
		Name:     "question_aware",
		Priority: 21,
		Group:    "recovery",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableQuestionAware && cfg.QueryIntent != "" },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipQueryDependent()
		},
		Factory: func(cfg PipelineConfig) Filter {
			f := NewQuestionAwareFilter(cfg.QueryIntent)
			if cfg.QuestionAwareThreshold > 0 {
				f.config.RelevanceThreshold = cfg.QuestionAwareThreshold
			}
			return f
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:     "density_adaptive",
		Priority: 22,
		Group:    "recovery",
		Enabled:  func(cfg PipelineConfig) bool { return cfg.EnableDensityAdaptive },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool {
			return p.shouldSkipSemanticChunk(input)
		},
		Factory: func(cfg PipelineConfig) Filter {
			f := NewDensityAdaptiveFilter()
			if cfg.DensityTargetRatio > 0 {
				f.config.TargetRatio = cfg.DensityTargetRatio
			}
			if cfg.DensityThreshold > 0 {
				f.config.DensityThreshold = cfg.DensityThreshold
			}
			return f
		},
	})

	// --- New Layers (23-26) ---

	RegisterLayer(LayerDescriptor{
		Name:       "symbolic_compress",
		Priority:   23,
		Group:      "new",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableSymbolicCompress },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewSymbolicCompressFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "phrase_grouping",
		Priority:   24,
		Group:      "new",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnablePhraseGrouping },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewPhraseGroupingFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "numerical_quant",
		Priority:   25,
		Group:      "new",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableNumericalQuant },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			numCfg := DefaultNumericalConfig()
			if cfg.DecimalPlaces > 0 {
				numCfg.DecimalPlaces = cfg.DecimalPlaces
			}
			f := NewNumericalQuantizer()
			f.config = numCfg
			return f
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "dynamic_ratio",
		Priority:   26,
		Group:      "new",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableDynamicRatio },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			dynCfg := DefaultDynamicRatioConfig()
			if cfg.DynamicRatioBase > 0 {
				dynCfg.BaseBudgetRatio = cfg.DynamicRatioBase
			}
			f := NewDynamicRatioFilter()
			f.config = dynCfg
			return f
		},
	})

	// --- Phase 2 Layers (27-30) ---

	RegisterLayer(LayerDescriptor{
		Name:       "hypernym",
		Priority:   27,
		Group:      "phase2",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableHypernym },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewHypernymCompressor()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "semantic_cache",
		Priority:   28,
		Group:      "phase2",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableSemanticCache },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewSemanticCacheFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "scope",
		Priority:   29,
		Group:      "phase2",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableScope },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewScopeFilter()
		},
	})

	RegisterLayer(LayerDescriptor{
		Name:       "kvzip",
		Priority:   30,
		Group:      "phase2",
		Enabled:    func(cfg PipelineConfig) bool { return cfg.EnableKVzip },
		ShouldSkip: func(p *PipelineCoordinator, input string) bool { return false },
		Factory: func(cfg PipelineConfig) Filter {
			return NewKVzipFilter()
		},
	})
}
