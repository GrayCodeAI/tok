package filter

import "github.com/GrayCodeAI/tokman/internal/cache"

// Pipeline defines the interface for compression pipelines.
// This allows mock testing and future pipeline implementations.
type Pipeline interface {
	Process(input string) (string, *PipelineStats)
}

// filterLayer pairs a compression filter with its stats key.
type filterLayer struct {
	filter Filter
	name   string
}

// PipelineCoordinator orchestrates the practical 20-layer compression pipeline.
// Research-based: Combines the best techniques from 120+ research papers worldwide
// to achieve maximum token reduction for CLI/Agent output.
type PipelineCoordinator struct {
	config PipelineConfig

	layers []filterLayer

	runtimeQueryIntent string
	layerRegistry      *LayerRegistry
	layerGate          *LayerGate

	// Layer 1: Entropy Filtering
	entropyFilter *EntropyFilter

	// Layer 2: Perplexity Pruning
	perplexityFilter *PerplexityFilter

	// Layer 3: Goal-Driven Selection
	goalDrivenFilter *GoalDrivenFilter

	// Layer 4: AST Preservation
	astPreserveFilter *ASTPreserveFilter

	// Layer 5: Contrastive Ranking
	contrastiveFilter *ContrastiveFilter

	// Layer 6: N-gram Abbreviation
	ngramAbbreviator *NgramAbbreviator

	// Layer 7: Evaluator Heads
	evaluatorHeadsFilter *EvaluatorHeadsFilter

	// Layer 8: Gist Compression
	gistFilter *GistFilter

	// Layer 9: Hierarchical Summary
	hierarchicalSummaryFilter *HierarchicalSummaryFilter

	// Layer 10: Budget Enforcement
	budgetEnforcer *BudgetEnforcer
	sessionTracker *SessionTracker

	// Layer 11: Compaction Layer (Semantic compression)
	compactionLayer *CompactionLayer

	// Layer 12: Attribution Filter (ProCut-style pruning)
	attributionFilter *AttributionFilter

	// Layer 13: H2O Filter (Heavy-Hitter Oracle)
	h2oFilter *H2OFilter

	// Layer 14: Attention Sink Filter (StreamingLLM-style)
	attentionSinkFilter *AttentionSinkFilter

	// Layer 15: Meta-Token Lossless Compression (arXiv:2506.00307)
	metaTokenFilter *MetaTokenFilter

	// Layer 16: Semantic Chunk Filter (ChunkKV style)
	semanticChunkFilter *SemanticChunkFilter

	// Layer 17: Sketch-based Reversible Store (KVReviver style)
	sketchStoreFilter *SketchStoreFilter

	// Layer 18: Budget-aware Dynamic Pruning (LazyLLM style)
	lazyPrunerFilter *LazyPrunerFilter

	// Layer 19: Semantic-Anchor Compression (SAC style)
	semanticAnchorFilter *SemanticAnchorFilter

	// Layer 20: Agent Memory Mode (Focus-inspired)
	agentMemoryFilter *AgentMemoryFilter

	// NEW: Inter-Layer Feedback Mechanism
	feedback *InterLayerFeedback

	// NEW: Quality Estimator for feedback
	qualityEstimator *QualityEstimator

	// TOML Filter Integration (declarative filters)
	tomlFilterWrapper Filter
	tomlFilterName    string

	// Optional guardrail
	qualityGuardrail *QualityGuardrail

	// Layers 21-25: 2026 Research filters
	marginalInfoGainFilter   *MarginalInfoGainFilter
	nearDedupFilter          *NearDedupFilter
	cotCompressFilter        *CoTCompressFilter
	codingAgentCtxFilter     *CodingAgentContextFilter
	perceptionCompressFilter *PerceptionCompressFilter

	// Layers 26-30: 2025/2026 reasoning + agent filters
	lightThinkerFilter  *LightThinkerFilter
	thinkSwitcherFilter *ThinkSwitcherFilter
	gmsaFilter          *GMSAFilter
	carlFilter          *CARLFilter
	slimInferFilter     *SlimInferFilter

	// Phase 2: SmallKV Model Compensation (2025)
	smallKVCompensator *SmallKVCompensator

	// Phase 2: Pipeline result cache for repeated inputs
	resultCache    *cache.FingerprintCache
	cacheEnabled   bool
	cacheHitCount  int64
	cacheMissCount int64
}

// CoreLayersConfig groups Layer 1-9 shared settings.
type CoreLayersConfig struct {
	LLMEnabled       bool
	SessionTracking  bool
	NgramEnabled     bool
	MultiFileEnabled bool
}

// CompactionLayerConfig groups Layer 11 settings.
type CompactionLayerConfig struct {
	Enabled       bool
	Threshold     int
	PreserveTurns int
	MaxTokens     int
	StateSnapshot bool
	AutoDetect    bool
}

// AttributionLayerConfig groups Layer 12 settings.
type AttributionLayerConfig struct {
	Enabled   bool
	Threshold float64
}

// H2OLayerConfig groups Layer 13 settings.
type H2OLayerConfig struct {
	Enabled         bool
	SinkSize        int
	RecentSize      int
	HeavyHitterSize int
}

// AttentionSinkLayerConfig groups Layer 14 settings.
type AttentionSinkLayerConfig struct {
	Enabled     bool
	SinkCount   int
	RecentCount int
}

// MetaTokenLayerConfig groups Layer 15 settings.
type MetaTokenLayerConfig struct {
	Enabled bool
	Window  int
	MinSize int
}

// SemanticChunkLayerConfig groups Layer 16 settings.
type SemanticChunkLayerConfig struct {
	Enabled   bool
	Method    string
	MinSize   int
	Threshold float64
}

// SketchStoreLayerConfig groups Layer 17 settings.
type SketchStoreLayerConfig struct {
	Enabled     bool
	BudgetRatio float64
	MaxSize     int
	HeavyHitter float64
}

// LazyPrunerLayerConfig groups Layer 18 settings.
type LazyPrunerLayerConfig struct {
	Enabled       bool
	BaseBudget    int
	DecayRate     float64
	RevivalBudget int
}

// SemanticAnchorLayerConfig groups Layer 19 settings.
type SemanticAnchorLayerConfig struct {
	Enabled bool
	Ratio   float64
	Spacing int
}

// AgentMemoryLayerConfig groups Layer 20 settings.
type AgentMemoryLayerConfig struct {
	Enabled            bool
	KnowledgeRetention float64
	HistoryPrune       float64
	ConsolidationMax   int
}

// QuestionAwareLayerConfig groups T12 settings.
type QuestionAwareLayerConfig struct {
	Enabled   bool
	Threshold float64
}

// DensityAdaptiveLayerConfig groups T17 settings.
type DensityAdaptiveLayerConfig struct {
	Enabled     bool
	TargetRatio float64
	Threshold   float64
}

// TFIDFLayerConfig groups TF-IDF filter settings.
type TFIDFLayerConfig struct {
	Enabled   bool
	Threshold float64
}

// NumericalQuantLayerConfig groups numerical quantization settings.
type NumericalQuantLayerConfig struct {
	Enabled       bool
	DecimalPlaces int
}

// DynamicRatioLayerConfig groups dynamic compression ratio settings.
type DynamicRatioLayerConfig struct {
	Enabled bool
	Base    float64
}

// LayerConfig groups per-layer config structs.
type LayerConfig struct {
	Core              CoreLayersConfig
	Compaction        CompactionLayerConfig
	Attribution       AttributionLayerConfig
	H2O               H2OLayerConfig
	AttentionSink     AttentionSinkLayerConfig
	MetaToken         MetaTokenLayerConfig
	SemanticChunk     SemanticChunkLayerConfig
	SketchStore       SketchStoreLayerConfig
	LazyPruner        LazyPrunerLayerConfig
	SemanticAnchor    SemanticAnchorLayerConfig
	AgentMemory       AgentMemoryLayerConfig
	QuestionAware     QuestionAwareLayerConfig
	DensityAdaptive   DensityAdaptiveLayerConfig
	TFIDF             TFIDFLayerConfig
	NumericalQuant    NumericalQuantLayerConfig
	DynamicRatio      DynamicRatioLayerConfig
	SymbolicCompress  bool
	PhraseGrouping    bool
	Hypernym          bool
	SemanticCache     bool
	Scope             bool
	SmallKV           bool
	KVzip             bool
	SWEzze            bool
	MixedDim          bool
	BEAVER            bool
	PoC               bool
	TokenQuant        bool
	TokenRetention    bool
	ACON              bool
	TOMLFilter        bool
	TOMLFilterCommand string
	CacheEnabled      bool
}

// PipelineConfigWithNestedLayers is a helper type for the new nested config structure.
// Use this gradually: migrate from flat fields to nested Layers config over time.
type PipelineConfigWithNestedLayers struct {
	// Core fields
	Mode                       Mode
	QueryIntent                string
	Budget                     int
	LLMEnabled                 bool
	SessionTracking            bool
	NgramEnabled               bool
	MultiFileEnabled           bool
	PromptTemplate             string
	EnableTOMLFilter           bool
	TOMLFilterCommand          string
	EnablePolicyRouter         bool
	EnableExtractivePrefilter  bool
	ExtractiveMaxLines         int
	ExtractiveHeadLines        int
	ExtractiveTailLines        int
	ExtractiveSignalLines      int
	EnableQualityGuardrail     bool
	LayerGateMode              string
	LayerGateAllowExperimental []string
	EnablePlannedLayers        bool

	// Layer sub-configs (preferred)
	Layers LayerConfig

	// Legacy flat fields for backward compatibility with existing callers.
	// Gradually migrate code to use cfg.Layers.* and remove these fields.

	// Core layer enable flags (Layers 1-9)
	EnableEntropy      bool
	EnablePerplexity   bool
	EnableGoalDriven   bool
	EnableAST          bool
	EnableContrastive  bool
	EnableEvaluator    bool
	EnableGist         bool
	EnableHierarchical bool

	// Layer 11: Compaction
	EnableCompaction        bool
	CompactionThreshold     int
	CompactionPreserveTurns int
	CompactionMaxTokens     int
	CompactionStateSnapshot bool
	CompactionAutoDetect    bool

	// Layer 12: Attribution
	EnableAttribution    bool
	AttributionThreshold float64

	// Layer 13: H2O
	EnableH2O          bool
	H2OSinkSize        int
	H2ORecentSize      int
	H2OHeavyHitterSize int

	// Layer 14: Attention Sink
	EnableAttentionSink  bool
	AttentionSinkCount   int
	AttentionRecentCount int

	// Layer 15: Meta-Token
	EnableMetaToken  bool
	MetaTokenWindow  int
	MetaTokenMinSize int

	// Layer 16: Semantic Chunk
	EnableSemanticChunk    bool
	SemanticChunkMethod    string
	SemanticChunkMinSize   int
	SemanticChunkThreshold float64

	// Layer 17: Semantic Cache
	EnableSketchStore bool
	SketchBudgetRatio float64
	SketchMaxSize     int
	SketchHeavyHitter float64

	// Layer 18: Lazy Pruner
	EnableLazyPruner  bool
	LazyBaseBudget    int
	LazyDecayRate     float64
	LazyRevivalBudget int

	// Layer 19: Semantic Anchor
	EnableSemanticAnchor  bool
	SemanticAnchorRatio   float64
	SemanticAnchorSpacing int

	// Layer 20: Agent Memory
	EnableAgentMemory       bool
	AgentKnowledgeRetention float64
	AgentHistoryPrune       float64
	AgentConsolidationMax   int

	// Adaptive layers
	EnableQuestionAware    bool
	QuestionAwareThreshold float64
	EnableDensityAdaptive  bool
	DensityTargetRatio     float64
	DensityThreshold       float64

	// TF-IDF
	EnableTFIDF    bool
	TFIDFThreshold float64

	// Reasoning trace
	EnableReasoningTrace bool
	MaxReflectionLoops   int

	// Phase 1: NEW filters
	EnableSymbolicCompress bool
	EnablePhraseGrouping   bool
	EnableNumericalQuant   bool
	DecimalPlaces          int
	EnableDynamicRatio     bool
	DynamicRatioBase       float64

	// Phase 2: Advanced filters
	EnableHypernym      bool
	EnableSemanticCache bool
	EnableScope         bool
	EnableSmallKV       bool
	EnableKVzip         bool

	// 2026 Research layers
	EnableSWEzze         bool
	EnableMixedDim       bool
	EnableBEAVER         bool
	EnablePoC            bool
	EnableTokenQuant     bool
	EnableTokenRetention bool
	EnableACON           bool

	// Layers 21-25: new 2025/2026 research filters
	EnableMarginalInfoGain   bool
	EnableNearDedup          bool
	EnableCoTCompress        bool
	EnableCodingAgentCtx     bool
	EnablePerceptionCompress bool

	// Layers 26-30: reasoning + agent filters
	EnableLightThinker  bool
	EnableThinkSwitcher bool
	EnableGMSA          bool
	EnableCARL          bool
	EnableSlimInfer     bool

	// Cache
	CacheEnabled bool
	CacheMaxSize int
}

// PipelineConfig is an alias for the full config type with backward-compatible flat fields.
// New code should use PipelineConfigWithNestedLayers to take advantage of nested structure.
type PipelineConfig = PipelineConfigWithNestedLayers

// PipelineStats holds statistics from the compression pipeline
type PipelineStats struct {
	OriginalTokens   int
	FinalTokens      int
	TotalSaved       int
	ReductionPercent float64
	LayerStats       map[string]LayerStat
	runningSaved     int
	CacheHit         bool
}

// LayerStat holds statistics for a single layer
type LayerStat struct {
	TokensSaved int
	Duration    int64
}
