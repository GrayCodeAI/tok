package filter

// AutoTier represents an automatic tier for adaptive layer enablement.
// Tiers are enabled automatically based on content context.
type AutoTier int

const (
	// AutoTierPre: Pre-processing layers (always run first)
	// Layers: 0 (QuantumLock), 0.5 (Photon)
	AutoTierPre AutoTier = iota

	// AutoTierCore: Foundation compression (high ROI, always enabled)
	// Layers: 1-10 (Entropy, Perplexity, AST, etc.)
	AutoTierCore

	// AutoTierSemantic: Semantic analysis (medium cost, high quality)
	// Layers: 11-20 (Compaction, H2O, Attention Sink, etc.)
	AutoTierSemantic

	// AutoTierAdvanced: Research layers (higher cost, specialized)
	// Layers: 21-40 (DiffAdapt, EPiC, AgentOCR, etc.)
	AutoTierAdvanced

	// AutoTierSpecialized: Experimental/edge cases (enable manually)
	// Layers: 41-50 (ContextCrunch, SearchCrunch, AdaptiveLearning)
	AutoTierSpecialized

	// AutoTierCount: Total number of tiers
	AutoTierCount
)

// AutoTierConfig holds configuration for each tier.
type AutoTierConfig struct {
	Name        string
	Description string
	LayerRange  [2]int  // [start, end] layer numbers
	Default     bool    // Enabled by default
	AutoEnable  bool    // Can be auto-enabled based on context
	CostLevel   int     // 1=low, 2=medium, 3=high, 4=very high
	MinInputLen int     // Minimum input length to enable
	ContentTypes []ContentType // Content types this tier is good for
}

// AutoTierConfigs defines all tier configurations.
var AutoTierConfigs = map[AutoTier]AutoTierConfig{
	AutoTierPre: {
		Name:        "pre",
		Description: "Pre-processing (alignment, image compression)",
		LayerRange:  [2]int{0, 0},
		Default:     true,
		AutoEnable:  true,
		CostLevel:   1,
		MinInputLen: 0,
		ContentTypes: []ContentType{
			ContentTypeUnknown, ContentTypeCode, ContentTypeLogs,
			ContentTypeConversation, ContentTypeGitOutput,
			ContentTypeTestOutput, ContentTypeDockerOutput, ContentTypeMixed,
		},
	},
	AutoTierCore: {
		Name:        "core",
		Description: "Foundation compression (essential, high ROI)",
		LayerRange:  [2]int{1, 10},
		Default:     true,
		AutoEnable:  true,
		CostLevel:   1,
		MinInputLen: 100,
		ContentTypes: []ContentType{
			ContentTypeUnknown, ContentTypeCode, ContentTypeLogs,
			ContentTypeConversation, ContentTypeGitOutput,
			ContentTypeTestOutput, ContentTypeDockerOutput, ContentTypeMixed,
		},
	},
	AutoTierSemantic: {
		Name:        "semantic",
		Description: "Semantic analysis (medium cost, high quality)",
		LayerRange:  [2]int{11, 20},
		Default:     true,
		AutoEnable:  true,
		CostLevel:   2,
		MinInputLen: 500,
		ContentTypes: []ContentType{
			ContentTypeCode, ContentTypeConversation,
			ContentTypeGitOutput, ContentTypeMixed,
		},
	},
	AutoTierAdvanced: {
		Name:        "advanced",
		Description: "Research layers (specialized, higher cost)",
		LayerRange:  [2]int{21, 40},
		Default:     false,
		AutoEnable:  true,
		CostLevel:   3,
		MinInputLen: 1000,
		ContentTypes: []ContentType{
			ContentTypeCode, ContentTypeConversation, ContentTypeMixed,
		},
	},
	AutoTierSpecialized: {
		Name:        "specialized",
		Description: "Experimental/edge case layers (auto-enabled for large mixed content)",
		LayerRange:  [2]int{41, 50},
		Default:     false,
		AutoEnable:  true, // Auto-enable with strict conditions
		CostLevel:   4,
		MinInputLen: 5000, // Very large inputs only
		ContentTypes: []ContentType{
			ContentTypeMixed,        // Complex mixed content
			ContentTypeConversation, // Multi-turn conversations
		},
	},
}

// AutoTierRecommendation holds recommended tiers for a context.
type AutoTierRecommendation struct {
	Tiers      []AutoTier
	Reason     string
	Confidence float64
}

// RecommendTiers analyzes content and recommends which tiers to enable.
func RecommendTiers(contentType ContentType, inputLen int, queryIntent string) AutoTierRecommendation {
	rec := AutoTierRecommendation{
		Tiers:      []AutoTier{AutoTierPre, AutoTierCore}, // Always enable base tiers
		Reason:     "Base tiers for " + contentType.String(),
		Confidence: 0.9,
	}

	// Check each auto-enable tier
	for tier := AutoTierSemantic; tier < AutoTierCount; tier++ {
		cfg := AutoTierConfigs[tier]
		if !cfg.AutoEnable {
			continue
		}

		// Check minimum length
		if inputLen < cfg.MinInputLen {
			continue
		}

		// Check content type compatibility
		if !isContentTypeCompatible(contentType, cfg.ContentTypes) {
			continue
		}

		// Check query intent for advanced tiers
		if tier >= AutoTierAdvanced && !isIntentCompatible(queryIntent, tier) {
			continue
		}

		rec.Tiers = append(rec.Tiers, tier)
	}

	// Adjust confidence based on content type detection
	if contentType == ContentTypeUnknown {
		rec.Confidence = 0.6
		rec.Reason += " (low confidence - unknown content type)"
	}

	return rec
}

// isContentTypeCompatible checks if content type matches allowed types.
func isContentTypeCompatible(ct ContentType, allowed []ContentType) bool {
	for _, a := range allowed {
		if a == ct {
			return true
		}
	}
	return false
}

// isIntentCompatible checks if query intent justifies advanced tiers.
func isIntentCompatible(intent string, tier AutoTier) bool {
	switch tier {
	case AutoTierAdvanced:
		// Advanced tiers useful for debug, review, deep analysis
		return intent == "debug" || intent == "review" || intent == "analyze" || intent == ""
	case AutoTierSpecialized:
		// Specialized tiers for deep analysis or when no specific intent (aggressive default)
		return intent == "deep-analyze" || intent == "experiment" || intent == "analyze" || intent == ""
	}
	return true
}

// BuildConfigFromTiers creates a PipelineConfig with tiers enabled.
func BuildConfigFromTiers(tiers []AutoTier, baseConfig PipelineConfig) PipelineConfig {
	cfg := baseConfig

	for _, tier := range tiers {
		enableTierLayers(&cfg, tier)
	}

	return cfg
}

// enableTierLayers enables all layers in a tier.
func enableTierLayers(cfg *PipelineConfig, tier AutoTier) {
	switch tier {
	case AutoTierPre:
		cfg.EnableQuantumLock = true
		// Note: Photon is 0.5, handled separately if needed

	case AutoTierCore:
		cfg.EnableEntropy = true
		cfg.EnablePerplexity = true
		cfg.EnableGoalDriven = true
		cfg.EnableAST = true
		cfg.EnableContrastive = true
		cfg.NgramEnabled = true
		cfg.EnableEvaluator = true
		cfg.EnableGist = true
		cfg.EnableHierarchical = true
		// Layer 10: Budget is auto-enabled if Budget > 0

	case AutoTierSemantic:
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

	case AutoTierAdvanced:
		cfg.EnableMarginalInfoGain = true
		cfg.EnableNearDedup = true
		cfg.EnableCoTCompress = true
		cfg.EnableCodingAgentCtx = true
		cfg.EnablePerceptionCompress = true
		cfg.EnableLightThinker = true
		cfg.EnableThinkSwitcher = true
		cfg.EnableGMSA = true
		cfg.EnableCARL = true
		cfg.EnableSlimInfer = true
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

	case AutoTierSpecialized:
		cfg.EnableAgentOCRHist = true
		cfg.EnablePlanBudget = true
		cfg.EnableLightMem = true
		cfg.EnablePathShorten = true
		cfg.EnableJSONSampler = true
		cfg.EnableContextCrunch = true
		cfg.EnableSearchCrunch = true
		cfg.EnableStructColl = true
		cfg.EnableAdaptiveLearning = true
	}
}

// EstimateTierCost estimates processing cost for given tiers.
func EstimateTierCost(tiers []AutoTier, inputTokens int) int {
	cost := 0
	for _, tier := range tiers {
		cfg := AutoTierConfigs[tier]
		// Cost = base cost * input size factor
		tierCost := cfg.CostLevel * inputTokens / 100
		cost += tierCost
	}
	return cost
}
