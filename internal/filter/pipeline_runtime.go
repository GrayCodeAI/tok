package filter

import "github.com/GrayCodeAI/tok/internal/config"

// PipelineRuntimeOptions carries per-request overrides for the runtime pipeline.
type PipelineRuntimeOptions struct {
	Mode        Mode
	QueryIntent string
	Budget      int
	LLMEnabled  bool
}

// ToFilterPipelineConfig converts user-facing config into the runtime pipeline config.
// Some fields are best-effort mappings because the public config and runtime pipeline
// have diverged over time; centralizing that mapping keeps behavior consistent.
func ToFilterPipelineConfig(c config.PipelineConfig, opts PipelineRuntimeOptions) PipelineConfig {
	cfg := PipelineConfig{
		Mode:                      opts.Mode,
		QueryIntent:               opts.QueryIntent,
		Budget:                    opts.Budget,
		LLMEnabled:                opts.LLMEnabled,
		SessionTracking:           true,
		NgramEnabled:              c.EnableNgram,
		EnableEntropy:             c.EnableEntropy,
		EnablePerplexity:          c.EnablePerplexity,
		EnableGoalDriven:          c.EnableGoalDriven,
		EnableAST:                 c.EnableAST,
		EnableContrastive:         c.EnableContrastive,
		EnableEvaluator:           c.EnableEvaluator,
		EnableGist:                c.EnableGist,
		EnableHierarchical:        c.EnableHierarchical,
		EnableCompaction:          c.EnableCompaction,
		CompactionThreshold:       c.CompactionThreshold,
		CompactionPreserveTurns:   c.CompactionPreserveTurns,
		CompactionMaxTokens:       c.CompactionMaxTokens,
		CompactionStateSnapshot:   c.CompactionStateSnapshot,
		CompactionAutoDetect:      c.CompactionAutoDetect,
		EnableAttribution:         c.EnableAttribution,
		AttributionThreshold:      c.AttributionThreshold,
		EnableH2O:                 c.EnableH2O,
		H2OSinkSize:               c.H2OSinkSize,
		H2ORecentSize:             c.H2ORecentSize,
		H2OHeavyHitterSize:        c.H2OHeavyHitterSize,
		EnableAttentionSink:       c.EnableAttentionSink,
		AttentionSinkCount:        c.AttentionSinkCount,
		AttentionRecentCount:      c.AttentionRecentCount,
		EnableMetaToken:           c.EnableMetaToken,
		MetaTokenWindow:           c.MetaTokenWindow,
		MetaTokenMinSize:          c.MetaTokenMinMatch,
		EnableSemanticChunk:       c.EnableSemanticChunk,
		SemanticChunkMinSize:      c.ChunkMinSize,
		SemanticChunkThreshold:    c.SemanticThreshold,
		EnableSketchStore:         c.EnableSketchStore,
		SketchBudgetRatio:         float64(c.SketchMemoryRatio) / 100.0,
		EnableLazyPruner:          c.EnableLazyPruner,
		LazyDecayRate:             c.LazyLayerDecay,
		EnableSemanticAnchor:      c.EnableSemanticAnchor,
		EnableAgentMemory:         c.EnableAgentMemory,
		AgentConsolidationMax:     c.AgentMemoryMaxNodes,
		EnablePolicyRouter:        c.EnablePolicyRouter,
		EnableExtractivePrefilter: c.EnableExtractiveFilter,
		ExtractiveMaxLines:        c.ExtractiveMaxLines,
		ExtractiveHeadLines:       c.ExtractiveHeadLines,
		ExtractiveTailLines:       c.ExtractiveTailLines,
		ExtractiveSignalLines:     c.ExtractiveSignalLines,
		EnableQualityGuardrail:    c.EnableQualityGuardrail,
		EnablePlannedLayers:       c.EnablePlannedLayers,
		EnableDiffAdapt:           c.EnableDiffAdapt,
		EnableEPiC:                c.EnableEPiC,
		EnableSSDP:                c.EnableSSDP,
		EnableAgentOCR:            c.EnableAgentOCR,
		EnableS2MAD:               c.EnableS2MAD,
		EnableACON:                c.EnableACON,
		EnableLatentCollab:        c.EnableLatentCollab,
		EnableGraphCoT:            c.EnableGraphCoT,
		EnableRoleBudget:          c.EnableRoleBudget,
		EnableSWEAdaptive:         c.EnableSWEAdaptive,
		EnableAgentOCRHist:        c.EnableAgentOCRHist,
		EnablePlanBudget:          c.EnablePlanBudget,
		EnableLightMem:            c.EnableLightMem,
		EnablePathShorten:         c.EnablePathShorten,
		EnableJSONSampler:         c.EnableJSONSampler,
		EnableContextCrunch:       c.EnableContextCrunch,
		EnableSearchCrunch:        c.EnableSearchCrunch,
		EnableStructColl:          c.EnableStructColl,
	}

	if c.EnableResearchPack {
		cfg.EnableDiffAdapt = true
		cfg.EnableEPiC = true
		cfg.EnableSSDP = true
		cfg.EnableAgentOCR = true
		cfg.EnableS2MAD = true
		cfg.EnableACON = true
		cfg.EnableLatentCollab = true
		cfg.EnableGraphCoT = true
		cfg.EnableRoleBudget = true
		cfg.EnableSWEAdaptive = true
		cfg.EnableAgentOCRHist = true
		cfg.EnablePlanBudget = true
		cfg.EnableLightMem = true
		cfg.EnablePathShorten = true
		cfg.EnableJSONSampler = true
		cfg.EnableContextCrunch = true
		cfg.EnableSearchCrunch = true
		cfg.EnableStructColl = true
	}

	if c.DefaultBudget > 0 && c.LazyBudgetRatio > 0 {
		cfg.LazyBaseBudget = int(float64(c.DefaultBudget) * c.LazyBudgetRatio)
	}
	if c.AnchorMinPreserve > 0 {
		cfg.SemanticAnchorSpacing = c.AnchorMinPreserve
	}

	return cfg
}
