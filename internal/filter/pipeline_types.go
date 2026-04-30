package filter

import (
	"sync"

	"github.com/GrayCodeAI/tok/internal/cache"
)

// Pipeline defines the interface for compression pipelines.
// This allows mock testing and future pipeline implementations.
type Pipeline interface {
	Process(input string) (string, *PipelineStats, error)
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
	resultCache        *cache.FingerprintCache
	cacheEnabled       bool
	layerCache         *LayerCache
	qualityGuardrail   *QualityGuardrail

	// Progress tracking for status line
	processedLayers int

	// Layer 0: QuantumLock (KV-cache alignment)
	quantumLockFilter *QuantumLockFilter

	// Layer 0.5: Photon (image compression)
	photonFilter *PhotonFilter

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

	// Unified research layers
	edgeCaseFilter  *EdgeCaseFilter
	reasoningFilter *ReasoningFilter
	advancedFilter  *AdvancedFilter

	// Post-processing / compatibility
	smallKVCompensator *SmallKVCompensator
	tomlFilterWrapper  Filter

	// Inter-Layer Feedback Mechanism
	feedback         *InterLayerFeedback
	qualityEstimator *QualityEstimator
	adaptiveLearning *AdaptiveLearningFilter
	crunchBench      *CrunchBench
}

// reportProgress emits a progress event if a callback is registered.
func (p *PipelineCoordinator) reportProgress(layer string, originalTokens, currentTokens int) {
	cb := GetProgressCallback()
	if cb != nil {
		// Compute progress percentage based on layers processed
		total := len(p.layers)
		if total > 0 {
			progress := float64(p.processedLayers) / float64(total) * 100
			cb(layer, originalTokens, currentTokens, progress)
		} else {
			cb(layer, originalTokens, currentTokens, 0)
		}
	}
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
// NOTE: Currently unused - reserved for future implementation.
type QuestionAwareLayerConfig struct {
	Enabled   bool
	Threshold float64
}

// DensityAdaptiveLayerConfig groups T17 settings.
// NOTE: Currently unused - reserved for future implementation.
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
// NOTE: Currently unused - reserved for future implementation.
type NumericalQuantLayerConfig struct {
	Enabled       bool
	DecimalPlaces int
}

// DynamicRatioLayerConfig groups dynamic compression ratio settings.
// NOTE: Currently unused - reserved for future implementation.
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

	// New: Claw Compactor features
	EnableAdaptiveLearning bool // Enable adaptive learning (merged EngramLearner + TieredSummary)
	EnableCrunchBench      bool // Enable comprehensive benchmarking
}

// PipelineConfigWithNestedLayers is a helper type for the new nested config structure.
// Use this gradually: migrate from flat fields to nested Layers config over time.
type PipelineConfigWithNestedLayers struct {
	// Core fields
	Mode                      Mode
	QueryIntent               string
	Budget                    int
	LLMEnabled                bool
	SessionTracking           bool
	NgramEnabled              bool
	MultiFileEnabled          bool
	PromptTemplate            string
	EnableTOMLFilter          bool
	TOMLFilterCommand         string
	EnablePolicyRouter        bool
	EnableExtractivePrefilter bool
	ExtractiveMaxLines        int
	ExtractiveHeadLines       int
	ExtractiveTailLines       int
	ExtractiveSignalLines     int
	EnableQualityGuardrail    bool
	LayerGateMode             string

	// New: Claw Compactor features
	EnableAdaptiveLearning     bool // Enable adaptive learning (merged EngramLearner + TieredSummary)
	EnableCrunchBench          bool // Enable comprehensive benchmarking
	LayerGateAllowExperimental []string
	EnablePlannedLayers        bool

	// Layer 0: QuantumLock (KV-cache alignment)
	EnableQuantumLock bool

	// Layer 0.5: Photon (image compression)
	EnablePhoton bool

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

	// Layers 31-45: adaptive reasoning + trajectory filters
	EnableDiffAdapt     bool
	EnableEPiC          bool
	EnableSSDP          bool
	EnableAgentOCR      bool
	EnableS2MAD         bool
	EnableLatentCollab  bool
	EnableGraphCoT      bool
	EnableRoleBudget    bool
	EnableSWEAdaptive   bool
	EnableAgentOCRHist  bool
	EnablePlanBudget    bool
	EnableLightMem      bool
	EnablePathShorten   bool
	EnableJSONSampler   bool
	EnableContextCrunch bool // Merged LogCrunch + DiffCrunch
	EnableSearchCrunch  bool
	EnableStructColl    bool

	// Unified experimental layers (L14-L16)
	EnableEdgeCase  bool // L14: merges L21-L25
	EnableReasoning bool // L15: merges L26-L30
	EnableAdvanced  bool // L16: merges L31-L45

	// Cache
	CacheEnabled bool
	CacheMaxSize int

	// Tier-based configuration (new)
	EnableTiers  bool       // Enable tier-based automatic layer selection
	EnabledTiers []AutoTier // Explicit list of tiers to enable (if empty, auto-select)
}

// PipelineConfig is an alias for the full config type with backward-compatible flat fields.
// New code should use PipelineConfigWithNestedLayers to take advantage of nested structure.
type PipelineConfig = PipelineConfigWithNestedLayers

// AllCoreLayersDisabled returns true when none of the core layers (1-9) are enabled.
// This is used by NewPipelineCoordinator to decide whether to apply zero-config defaults.
func (cfg *PipelineConfig) AllCoreLayersDisabled() bool {
	return !cfg.EnableEntropy && !cfg.EnablePerplexity && !cfg.EnableGoalDriven &&
		!cfg.EnableAST && !cfg.EnableContrastive && !cfg.EnableEvaluator &&
		!cfg.EnableGist && !cfg.EnableHierarchical
}

// HasExplicitSettings returns true when the user has provided any non-default
// configuration that should prevent zero-config defaults from being applied.
func (cfg *PipelineConfig) HasExplicitSettings() bool {
	return cfg.Budget > 0 || cfg.QueryIntent != "" || cfg.LLMEnabled ||
		cfg.NgramEnabled || cfg.MultiFileEnabled || cfg.SessionTracking ||
		cfg.EnableCompaction || cfg.EnableAttribution || cfg.EnableH2O || cfg.EnableAttentionSink ||
		cfg.EnableAdaptiveLearning || cfg.EnableContextCrunch
}

// PipelineStats holds statistics from the compression pipeline
type PipelineStats struct {
	OriginalTokens   int
	FinalTokens      int
	TotalSaved       int
	ReductionPercent float64
	LayerStats       map[string]LayerStat
	runningSaved     int
	CacheHit         bool

	// Thread-safety fields
	mu sync.RWMutex
}

// LayerStat holds statistics for a single layer
type LayerStat struct {
	TokensSaved int
	Duration    int64
}

// UpdateConfig allows pooled coordinators to be reconfigured before reuse.
func (p *PipelineCoordinator) UpdateConfig(fn func(*PipelineConfig)) {
	fn(&p.config)
}

// RunningSavedSafe returns the running saved count safely.
func (s *PipelineStats) RunningSavedSafe() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runningSaved
}

// AddLayerStatSafe adds a layer stat in a thread-safe manner.
func (s *PipelineStats) AddLayerStatSafe(name string, stat LayerStat) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.LayerStats == nil {
		s.LayerStats = make(map[string]LayerStat)
	}
	s.LayerStats[name] = stat
	s.runningSaved += stat.TokensSaved
}

// LayerBitset provides a compact bitset representation of layer enable flags.
// This is an optional optimization for callers that need to pass config across
// wire boundaries or store many configs in memory.
//
// Usage:
//
//	bits := cfg.ToLayerBitset()
//	// ... pass bits over the wire ...
//	restored := bits.ToConfig(cfg.Mode)
//
// Bit positions (0-indexed):
//
//	0  - Entropy
//	1  - Perplexity
//	2  - GoalDriven
//	3  - AST
//	4  - Contrastive
//	5  - Evaluator
//	6  - Gist
//	7  - Hierarchical
//	8  - Compaction
//	9  - Attribution
//	10 - H2O
//	11 - AttentionSink
//	12 - MetaToken
//	13 - SemanticChunk
//	14 - SketchStore
//	15 - LazyPruner
//	16 - SemanticAnchor
//	17 - AgentMemory
//	18 - EdgeCase
//	19 - Reasoning
//	20 - Advanced
//	21 - QuantumLock
//	22 - Photon
//	23 - AdaptiveLearning
//	24 - CrunchBench
//	25..63 reserved
type LayerBitset uint64

// ToLayerBitset packs the core layer enable flags into a compact uint64.
func (cfg *PipelineConfig) ToLayerBitset() LayerBitset {
	var b LayerBitset
	if cfg.EnableEntropy {
		b |= 1 << 0
	}
	if cfg.EnablePerplexity {
		b |= 1 << 1
	}
	if cfg.EnableGoalDriven {
		b |= 1 << 2
	}
	if cfg.EnableAST {
		b |= 1 << 3
	}
	if cfg.EnableContrastive {
		b |= 1 << 4
	}
	if cfg.EnableEvaluator {
		b |= 1 << 5
	}
	if cfg.EnableGist {
		b |= 1 << 6
	}
	if cfg.EnableHierarchical {
		b |= 1 << 7
	}
	if cfg.EnableCompaction {
		b |= 1 << 8
	}
	if cfg.EnableAttribution {
		b |= 1 << 9
	}
	if cfg.EnableH2O {
		b |= 1 << 10
	}
	if cfg.EnableAttentionSink {
		b |= 1 << 11
	}
	if cfg.EnableMetaToken {
		b |= 1 << 12
	}
	if cfg.EnableSemanticChunk {
		b |= 1 << 13
	}
	if cfg.EnableSketchStore {
		b |= 1 << 14
	}
	if cfg.EnableLazyPruner {
		b |= 1 << 15
	}
	if cfg.EnableSemanticAnchor {
		b |= 1 << 16
	}
	if cfg.EnableAgentMemory {
		b |= 1 << 17
	}
	if cfg.EnableEdgeCase {
		b |= 1 << 18
	}
	if cfg.EnableReasoning {
		b |= 1 << 19
	}
	if cfg.EnableAdvanced {
		b |= 1 << 20
	}
	if cfg.EnableQuantumLock {
		b |= 1 << 21
	}
	if cfg.EnablePhoton {
		b |= 1 << 22
	}
	if cfg.EnableAdaptiveLearning {
		b |= 1 << 23
	}
	if cfg.EnableCrunchBench {
		b |= 1 << 24
	}
	return b
}

// ToConfig restores layer enable flags from a bitset into a PipelineConfig.
// Callers should set Mode, Budget, QueryIntent, etc. separately.
func (b LayerBitset) ToConfig() PipelineConfig {
	return PipelineConfig{
		EnableEntropy:          b&(1<<0) != 0,
		EnablePerplexity:       b&(1<<1) != 0,
		EnableGoalDriven:       b&(1<<2) != 0,
		EnableAST:              b&(1<<3) != 0,
		EnableContrastive:      b&(1<<4) != 0,
		EnableEvaluator:        b&(1<<5) != 0,
		EnableGist:             b&(1<<6) != 0,
		EnableHierarchical:     b&(1<<7) != 0,
		EnableCompaction:       b&(1<<8) != 0,
		EnableAttribution:      b&(1<<9) != 0,
		EnableH2O:              b&(1<<10) != 0,
		EnableAttentionSink:    b&(1<<11) != 0,
		EnableMetaToken:        b&(1<<12) != 0,
		EnableSemanticChunk:    b&(1<<13) != 0,
		EnableSketchStore:      b&(1<<14) != 0,
		EnableLazyPruner:       b&(1<<15) != 0,
		EnableSemanticAnchor:   b&(1<<16) != 0,
		EnableAgentMemory:      b&(1<<17) != 0,
		EnableEdgeCase:         b&(1<<18) != 0,
		EnableReasoning:        b&(1<<19) != 0,
		EnableAdvanced:         b&(1<<20) != 0,
		EnableQuantumLock:      b&(1<<21) != 0,
		EnablePhoton:           b&(1<<22) != 0,
		EnableAdaptiveLearning: b&(1<<23) != 0,
		EnableCrunchBench:      b&(1<<24) != 0,
	}
}
