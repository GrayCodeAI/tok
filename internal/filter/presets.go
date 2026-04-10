package filter

// Tier defines the depth of the compression pipeline.
// Higher tiers activate more layers for deeper compression.
type Tier string

const (
	// Tier 1: Surface — removes obvious noise, keeps everything intact
	TierSurface Tier = "surface" // 3 layers, 30-50% reduction

	// Tier 2: Trim — cuts dead weight, keeps structure
	TierTrim Tier = "trim" // 12 layers, 50-70% reduction

	// Tier 3: Extract — pulls out the essence
	TierExtract Tier = "extract" // 24 layers, 70-90% reduction

	// Tier 4: Core — practical high-impact 20-layer runtime profile
	TierCore Tier = "core" // 20 layers, quality-first compression

	// Tier C: Code — code-aware, preserves syntax structure
	TierCode Tier = "code" // 8 layers, 50-70% reduction

	// Tier L: Log — log-aware, deduplicates and groups
	TierLog Tier = "log" // 7 layers, 60-80% reduction

	// Tier T: Thread — conversation-aware, preserves context
	TierThread Tier = "thread" // 6 layers, 55-75% reduction

	// Tier A: Adaptive — routing + extractive-first for large/mixed outputs
	TierAdaptive Tier = "adaptive" // dynamic path, quality-first
)

// TierConfig returns a PipelineConfig for the given tier.
func TierConfig(tier Tier, baseMode Mode) PipelineConfig {
	cfg := PipelineConfig{
		Mode:            baseMode,
		SessionTracking: true,
	}

	switch tier {
	case TierSurface:
		cfg.EnableEntropy = true
		cfg.EnableGoalDriven = true
		cfg.EnableH2O = true
		cfg.EnableNumericalQuant = true

	case TierTrim:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableAST = true
		cfg.EnableContrastive = true
		cfg.NgramEnabled = true
		cfg.EnableEvaluator = true
		cfg.EnableH2O = true
		cfg.EnableAttentionSink = true
		cfg.EnableMetaToken = true
		cfg.EnableNumericalQuant = true
		cfg.EnableDynamicRatio = true

	case TierExtract:
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
		cfg.EnableMetaToken = true
		cfg.EnableSemanticChunk = true
		cfg.EnableLazyPruner = true
		cfg.EnableSemanticAnchor = true
		cfg.EnableAgentMemory = true
		cfg.EnableSymbolicCompress = true
		cfg.EnablePhraseGrouping = true
		cfg.EnableNumericalQuant = true
		cfg.EnableDynamicRatio = true
		cfg.EnableHypernym = true
		cfg.EnableSemanticCache = true
		cfg.EnableKVzip = true
		cfg.EnableDiffAdapt = true
		cfg.EnableEPiC = true
		cfg.EnableSSDP = true
		cfg.EnableAgentOCR = true
		cfg.EnableS2MAD = true

	case TierCore:
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
		cfg.EnableMetaToken = true
		cfg.EnableSemanticChunk = true
		cfg.EnableSketchStore = true
		cfg.EnableLazyPruner = true
		cfg.EnableSemanticAnchor = true
		cfg.EnableAgentMemory = true
		cfg.EnableQuestionAware = false
		cfg.EnableDensityAdaptive = false
		cfg.EnableSymbolicCompress = false
		cfg.EnablePhraseGrouping = false
		cfg.EnableNumericalQuant = false
		cfg.EnableDynamicRatio = false
		cfg.EnableHypernym = false
		cfg.EnableSemanticCache = false
		cfg.EnableScope = false
		cfg.EnableSmallKV = false
		cfg.EnableKVzip = false
		cfg.EnableSWEzze = false
		cfg.EnableMixedDim = false
		cfg.EnableBEAVER = false
		cfg.EnablePoC = false
		cfg.EnableTokenQuant = false
		cfg.EnableTokenRetention = false
		cfg.EnableACON = false
		cfg.EnablePlannedLayers = false

	case TierCode:
		cfg.EnableEntropy = true
		cfg.EnableAST = true
		cfg.EnableGoalDriven = true
		cfg.EnableH2O = true
		cfg.EnableMetaToken = true
		cfg.EnableSymbolicCompress = true
		cfg.EnableNumericalQuant = true
		cfg.EnableSWEzze = true

	case TierLog:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableEvaluator = true
		cfg.EnableH2O = true
		cfg.EnableAttribution = true
		cfg.EnableSketchStore = true
		cfg.EnableNumericalQuant = true

	case TierThread:
		cfg.EnableEntropy = true
		cfg.EnableCompaction = true
		cfg.EnableAttentionSink = true
		cfg.EnableLazyPruner = true
		cfg.EnableSemanticAnchor = true
		cfg.EnableAgentMemory = true

	case TierAdaptive:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableContrastive = true
		cfg.EnableEvaluator = true
		cfg.EnableAttribution = true
		cfg.EnableH2O = true
		cfg.EnableAttentionSink = true
		cfg.EnableQuestionAware = true
		cfg.EnableDynamicRatio = true
		cfg.EnablePolicyRouter = true
		cfg.EnableExtractivePrefilter = true
		cfg.EnableQualityGuardrail = true
		cfg.ExtractiveMaxLines = 400
		cfg.ExtractiveHeadLines = 80
		cfg.ExtractiveTailLines = 60
		cfg.ExtractiveSignalLines = 120
		cfg.EnableDiffAdapt = true
		cfg.EnableEPiC = true
		cfg.EnableSSDP = true
		cfg.EnableAgentOCR = true
		cfg.EnableS2MAD = true
	}

	return cfg
}

// ApplyTier compresses input using the specified tier.
func ApplyTier(input string, mode Mode, tier Tier) (string, int) {
	cfg := TierConfig(tier, mode)
	p := NewPipelineCoordinator(cfg)
	output, stats := p.Process(input)
	return output, stats.TotalSaved
}

// Backwards compatibility aliases
type Profile = Tier
type CompressionMode = Tier

const (
	ProfileFast     Tier = TierSurface
	ProfileBalanced Tier = TierTrim
	ProfileCode     Tier = TierCode
	ProfileLog      Tier = TierLog
	ProfileChat     Tier = TierThread
	ProfileMax      Tier = TierCore

	ModeSkim       Tier = TierSurface
	ModeRefine     Tier = TierTrim
	ModeDistill    Tier = TierExtract
	ModeAnnihilate Tier = TierCore
)

// ProfileConfig is an alias for TierConfig (backwards compat).
func ProfileConfig(profile Profile, baseMode Mode) PipelineConfig {
	return TierConfig(profile, baseMode)
}

// ApplyProfile is an alias for ApplyTier (backwards compat).
func ApplyProfile(input string, mode Mode, profile Profile) (string, int) {
	return ApplyTier(input, mode, Tier(profile))
}

// ModeConfig is an alias for TierConfig (backwards compat).
func ModeConfig(mode CompressionMode, baseMode Mode) PipelineConfig {
	return TierConfig(mode, baseMode)
}

// ApplyMode is an alias for ApplyTier (backwards compat).
func ApplyMode(input string, mode Mode, cm CompressionMode) (string, int) {
	return ApplyTier(input, mode, Tier(cm))
}

// PresetConfig for backwards compatibility.
type PipelinePreset = Tier

const (
	PresetFast     Tier = TierSurface
	PresetBalanced Tier = TierTrim
	PresetFull     Tier = TierCore
	PresetAuto     Tier = ""
)

func PresetConfig(preset PipelinePreset, baseMode Mode) PipelineConfig {
	if preset == PresetAuto {
		return PipelineConfig{Mode: baseMode, SessionTracking: true}
	}
	return TierConfig(preset, baseMode)
}

func QuickProcessPreset(input string, mode Mode, preset PipelinePreset) (string, int) {
	cfg := PresetConfig(preset, mode)
	p := NewPipelineCoordinator(cfg)
	output, stats := p.Process(input)
	return output, stats.TotalSaved
}
