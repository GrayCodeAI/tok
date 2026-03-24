package filter

// PipelinePreset defines a compression pipeline mode with specific layers enabled.
// T90: Provide fast/balanced/full presets for different use cases.
type PipelinePreset string

const (
	// PresetFast runs only layers 1, 3, 10 (entropy, goal-driven, budget).
	// ~3x faster than full, ~60% of the compression.
	PresetFast PipelinePreset = "fast"

	// PresetBalanced runs layers 1-6, 10, 14 (entropy through ngram + budget + attention sink).
	// ~1.5x faster than full, ~85% of the compression.
	PresetBalanced PipelinePreset = "balanced"

	// PresetFull runs all 29 layers for maximum compression.
	PresetFull PipelinePreset = "full"
)

// PresetConfig returns a PipelineConfig for the given preset.
func PresetConfig(preset PipelinePreset, baseMode Mode) PipelineConfig {
	cfg := PipelineConfig{
		Mode:            baseMode,
		SessionTracking: true,
	}

	switch preset {
	case PresetFast:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = false
		cfg.EnableGoalDriven = true
		cfg.EnableAST = false
		cfg.EnableContrastive = false
		cfg.NgramEnabled = false
		cfg.EnableEvaluator = false
		cfg.EnableGist = false
		cfg.EnableHierarchical = false
		cfg.EnableCompaction = false
		cfg.EnableAttribution = false
		cfg.EnableH2O = false
		cfg.EnableAttentionSink = false
		// NEW layers disabled in fast preset
		cfg.EnableTFIDF = false
		cfg.EnableReasoningTrace = false
		cfg.EnableSymbolicCompress = false
		cfg.EnablePhraseGrouping = false
		cfg.EnableNumericalQuant = false
		cfg.EnableDynamicRatio = false

	case PresetBalanced:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableAST = true
		cfg.EnableContrastive = true
		cfg.NgramEnabled = true
		cfg.EnableEvaluator = false
		cfg.EnableGist = false
		cfg.EnableHierarchical = false
		cfg.EnableCompaction = false
		cfg.EnableAttribution = false
		cfg.EnableH2O = false
		cfg.EnableAttentionSink = true
		// NEW layers: TF-IDF + numerical in balanced
		cfg.EnableTFIDF = true
		cfg.EnableReasoningTrace = false
		cfg.EnableSymbolicCompress = false
		cfg.EnablePhraseGrouping = false
		cfg.EnableNumericalQuant = true
		cfg.EnableDynamicRatio = false

	default: // PresetFull - all 29 layers
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableAST = true
		cfg.EnableContrastive = true
		cfg.NgramEnabled = true
		cfg.EnableEvaluator = true
		cfg.EnableGist = true
		cfg.EnableHierarchical = true
		cfg.EnableCompaction = true
		cfg.EnableAttribution = true
		cfg.EnableH2O = true
		cfg.EnableAttentionSink = true
		// NEW layers enabled in full preset
		cfg.EnableTFIDF = true
		cfg.EnableReasoningTrace = true
		cfg.EnableSymbolicCompress = true
		cfg.EnablePhraseGrouping = true
		cfg.EnableNumericalQuant = true
		cfg.EnableDynamicRatio = true
		// Phase 2 layers
		cfg.EnableHypernym = true
		cfg.EnableSemanticCache = true
		cfg.EnableScope = true
		cfg.EnableSmallKV = true
		cfg.EnableKVzip = true
	}

	return cfg
}

// QuickProcessPreset runs compression with a named preset.
func QuickProcessPreset(input string, mode Mode, preset PipelinePreset) (string, int) {
	cfg := PresetConfig(preset, mode)
	p := NewPipelineCoordinator(cfg)
	output, stats := p.Process(input)
	return output, stats.TotalSaved
}
