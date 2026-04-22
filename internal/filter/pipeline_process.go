package filter

import (
	"github.com/GrayCodeAI/tok/internal/core"
)

// Process runs the six-layer compression pipeline.
//
// Layers run in order; each exits early if the token budget is already met.
// The quality guardrail runs after all six layers and may trigger a fallback
// to ModeMinimal if the output fails semantic validation.
func (p *PipelineCoordinator) Process(input string) (string, *PipelineStats) {
	if p == nil {
		return input, &PipelineStats{
			OriginalTokens: core.EstimateTokens(input),
			FinalTokens:    core.EstimateTokens(input),
			LayerStats:     make(map[string]LayerStat, 6),
		}
	}

	stats := &PipelineStats{
		OriginalTokens: core.EstimateTokens(input),
		LayerStats:     make(map[string]LayerStat, 6),
	}

	output := input

	type layerRun struct {
		name string
		fn   func(string, *PipelineStats) string
	}

	for _, l := range [6]layerRun{
		{"preprocess", p.runLayer1Preprocess},
		{"structural", p.runLayer2Structural},
		{"semantic", p.runLayer3Semantic},
		{"llm_specific", p.runLayer4LLMSpecific},
		{"content_type", p.runLayer5ContentType},
		{"budget_quality", p.runLayer6BudgetQuality},
	} {
		output = l.fn(output, stats)
		if p.shouldEarlyExit(stats) {
			return output, p.finalizeStats(stats, output)
		}
	}

	// KV-size post-compensation (applied after budget so it doesn't inflate token count)
	if p.smallKVCompensator != nil {
		output = p.smallKVCompensator.Compensate(input, output, p.config.Mode)
	}

	// Quality feedback loop — informs adaptive learning for future runs
	p.recordFeedback(input, output, stats)

	// Quality guardrail — falls back to ModeMinimal if output fails semantic validation
	if p.qualityGuardrail != nil {
		gr := p.qualityGuardrail.Validate(input, output)
		if !gr.Passed {
			safeOutput, safeStats := p.runGuardrailFallback(input)
			safeStats.AddLayerStatSafe(LayerGuardrailFallback, LayerStat{TokensSaved: 0})
			safeStats.AddLayerStatSafe("guardrail_reason_"+gr.Reason, LayerStat{TokensSaved: 0})
			return safeOutput, safeStats
		}
	}

	return output, p.finalizeStats(stats, output)
}

func (p *PipelineCoordinator) runGuardrailFallback(input string) (string, *PipelineStats) {
	// Build a minimal fallback config to avoid copying the full 100+ field struct
	fallbackCfg := &PipelineConfig{
		Mode:                ModeMinimal,
		QueryIntent:         p.config.QueryIntent,
		Budget:              p.config.Budget,
		LLMEnabled:          p.config.LLMEnabled,
		SessionTracking:     p.config.SessionTracking,
		NgramEnabled:        p.config.NgramEnabled,
		MultiFileEnabled:    p.config.MultiFileEnabled,
		EnableEntropy:       p.config.EnableEntropy,
		EnablePerplexity:    p.config.EnablePerplexity,
		EnableAST:           p.config.EnableAST,
		EnableGist:          p.config.EnableGist,
		EnableHierarchical:  p.config.EnableHierarchical,
		EnableCompaction:    p.config.EnableCompaction,
		EnableAttribution:   p.config.EnableAttribution,
		EnableH2O:           p.config.EnableH2O,
		EnableAttentionSink: p.config.EnableAttentionSink,
		EnableTOMLFilter:    p.config.EnableTOMLFilter,
		CacheEnabled:        p.config.CacheEnabled,
		CacheMaxSize:        p.config.CacheMaxSize,
	}
	fallback := NewPipelineCoordinator(fallbackCfg)
	return fallback.Process(input)
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
