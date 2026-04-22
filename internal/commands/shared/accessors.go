package shared

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func (s *AppState) IsVerbose() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Verbose > 0
}

// IsUltraCompact returns true if ultra-compact mode is enabled.
func (s *AppState) IsUltraCompact() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UltraCompact
}

// GetQueryIntent returns the query intent from flag or environment.
func (s *AppState) GetQueryIntent() string {
	s.mu.RLock()
	intent := s.QueryIntent
	s.mu.RUnlock()
	if intent != "" {
		return intent
	}
	return os.Getenv("TOK_QUERY")
}

// IsLLMEnabled returns true if LLM compression is enabled.
func (s *AppState) IsLLMEnabled() bool {
	s.mu.RLock()
	enabled := s.LLMEnabled
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_LLM") == "true"
}

// GetTokenBudget returns the token budget from flag or environment.
func (s *AppState) GetTokenBudget() int {
	s.mu.RLock()
	budget := s.TokenBudget
	s.mu.RUnlock()
	if budget > 0 {
		return budget
	}
	envBudget := os.Getenv("TOK_BUDGET")
	if envBudget != "" {
		var b int
		if _, err := fmt.Sscanf(envBudget, "%d", &b); err == nil {
			return b
		}
		log.Printf("warning: invalid TOK_BUDGET value %q, defaulting to unlimited", envBudget)
	}
	return 0
}

// GetLayerProfile returns the compression profile from flag or environment.
func (s *AppState) GetLayerProfile() string {
	s.mu.RLock()
	profile := s.LayerProfile
	s.mu.RUnlock()
	if profile != "" {
		return profile
	}
	return os.Getenv("TOK_PROFILE")
}

// GetLayerPreset returns the layer preset from flag or environment.
func (s *AppState) GetLayerPreset() string {
	s.mu.RLock()
	preset := s.LayerPreset
	s.mu.RUnlock()
	if preset != "" {
		return preset
	}
	return os.Getenv("TOK_PRESET")
}

// IsQuietMode returns true if quiet mode is enabled.
func (s *AppState) IsQuietMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.QuietMode
}

// IsReversibleEnabled returns true if reversible mode is enabled.
func (s *AppState) IsReversibleEnabled() bool {
	s.mu.RLock()
	enabled := s.ReversibleEnabled
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_REVERSIBLE") == "true"
}

// IsRemoteMode returns true if remote mode is enabled.
func (s *AppState) IsRemoteMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RemoteMode || os.Getenv("TOK_REMOTE") == "true"
}

// GetCompressionAddr returns the compression service address.
func (s *AppState) GetCompressionAddr() string {
	s.mu.RLock()
	addr := s.CompressionAddr
	s.mu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOK_COMPRESSION_ADDR")
}

// GetAnalyticsAddr returns the analytics service address.
func (s *AppState) GetAnalyticsAddr() string {
	s.mu.RLock()
	addr := s.AnalyticsAddr
	s.mu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOK_ANALYTICS_ADDR")
}

// GetRemoteTimeout returns the remote operation timeout in seconds.
func (s *AppState) GetRemoteTimeout() int {
	s.mu.RLock()
	timeout := s.RemoteTimeout
	s.mu.RUnlock()
	if timeout > 0 {
		return timeout
	}
	return 30
}

// GetEnableLayers returns layers to explicitly enable.
func (s *AppState) GetEnableLayers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.EnableLayers
}

// GetDisableLayers returns layers to explicitly disable.
func (s *AppState) GetDisableLayers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.DisableLayers
}

// IsStreamMode returns true if streaming mode is enabled for large inputs.
func (s *AppState) IsStreamMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.StreamMode
}

// IsPolicyRouterEnabled returns true if policy router mode is enabled.
func (s *AppState) IsPolicyRouterEnabled() bool {
	s.mu.RLock()
	enabled := s.PolicyRouter
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_POLICY_ROUTER") == "true"
}

// IsExtractiveEnabled returns true if extractive prefilter is enabled.
func (s *AppState) IsExtractiveEnabled() bool {
	s.mu.RLock()
	enabled := s.Extractive
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_EXTRACTIVE_PREFILTER") == "true"
}

func envInt(name string, def int) int {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// GetExtractiveMax returns max lines for extractive prefilter.
func (s *AppState) GetExtractiveMax() int {
	s.mu.RLock()
	v := s.ExtractiveMax
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_MAX_LINES", 400)
}

// GetExtractiveHead returns preserved head lines for extractive prefilter.
func (s *AppState) GetExtractiveHead() int {
	s.mu.RLock()
	v := s.ExtractiveHead
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_HEAD_LINES", 80)
}

// GetExtractiveTail returns preserved tail lines for extractive prefilter.
func (s *AppState) GetExtractiveTail() int {
	s.mu.RLock()
	v := s.ExtractiveTail
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_TAIL_LINES", 60)
}

// GetExtractiveSignal returns signal line budget for extractive prefilter.
func (s *AppState) GetExtractiveSignal() int {
	s.mu.RLock()
	v := s.ExtractiveSignal
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_SIGNAL_LINES", 120)
}

// IsQualityGuardrailEnabled returns true if quality guardrail is enabled.
func (s *AppState) IsQualityGuardrailEnabled() bool {
	s.mu.RLock()
	enabled := s.QualityGuardrail
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_QUALITY_GUARDRAIL") == "true"
}

// IsDiffAdaptEnabled returns true if DiffAdapt layer is enabled.
func (s *AppState) IsDiffAdaptEnabled() bool {
	s.mu.RLock()
	enabled := s.DiffAdapt
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_DIFF_ADAPT") == "true"
}

// IsEPiCEnabled returns true if EPiC layer is enabled.
func (s *AppState) IsEPiCEnabled() bool {
	s.mu.RLock()
	enabled := s.EPiC
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_EPIC") == "true"
}

// IsSSDPEnabled returns true if SSDP layer is enabled.
func (s *AppState) IsSSDPEnabled() bool {
	s.mu.RLock()
	enabled := s.SSDP
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_SSDP") == "true"
}

// IsAgentOCREnabled returns true if AgentOCR layer is enabled.
func (s *AppState) IsAgentOCREnabled() bool {
	s.mu.RLock()
	enabled := s.AgentOCR
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_AGENT_OCR") == "true"
}

// IsS2MADEnabled returns true if S2-MAD layer is enabled.
func (s *AppState) IsS2MADEnabled() bool {
	s.mu.RLock()
	enabled := s.S2MAD
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_S2_MAD") == "true"
}

// IsACONEnabled returns true if ACON layer is enabled.
func (s *AppState) IsACONEnabled() bool {
	s.mu.RLock()
	enabled := s.ACON
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_ACON") == "true"
}

// IsResearchPackEnabled returns true if research pack is enabled.
func (s *AppState) IsResearchPackEnabled() bool {
	s.mu.RLock()
	enabled := s.ResearchPack
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_RESEARCH_PACK") == "true"
}

// IsLatentCollabEnabled returns true if latent collaboration layer is enabled.
func (s *AppState) IsLatentCollabEnabled() bool {
	s.mu.RLock()
	enabled := s.LatentCollab
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_LATENT_COLLAB") == "true"
}

// IsGraphCoTEnabled returns true if graph-CoT layer is enabled.
func (s *AppState) IsGraphCoTEnabled() bool {
	s.mu.RLock()
	enabled := s.GraphCoT
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_GRAPH_COT") == "true"
}

// IsRoleBudgetEnabled returns true if role-budget layer is enabled.
func (s *AppState) IsRoleBudgetEnabled() bool {
	s.mu.RLock()
	enabled := s.RoleBudget
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_ROLE_BUDGET") == "true"
}

// IsSWEAdaptiveEnabled returns true if SWE adaptive loop is enabled.
func (s *AppState) IsSWEAdaptiveEnabled() bool {
	s.mu.RLock()
	enabled := s.SWEAdaptive
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_SWE_ADAPTIVE") == "true"
}

// IsAgentOCRHistoryEnabled returns true if agent OCR history layer is enabled.
func (s *AppState) IsAgentOCRHistoryEnabled() bool {
	s.mu.RLock()
	enabled := s.AgentOCRHistory
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_AGENT_OCR_HISTORY") == "true"
}

// IsPlanBudgetEnabled returns true if plan-budget layer is enabled.
func (s *AppState) IsPlanBudgetEnabled() bool {
	s.mu.RLock()
	enabled := s.PlanBudget
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_PLAN_BUDGET") == "true"
}

// IsLightMemEnabled returns true if lightmem layer is enabled.
func (s *AppState) IsLightMemEnabled() bool {
	s.mu.RLock()
	enabled := s.LightMem
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_LIGHTMEM") == "true"
}

// IsPathShortenEnabled returns true if path-shorten layer is enabled.
func (s *AppState) IsPathShortenEnabled() bool {
	s.mu.RLock()
	enabled := s.PathShorten
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_PATH_SHORTEN") == "true"
}

// IsJSONSamplerEnabled returns true if json-sampler layer is enabled.
func (s *AppState) IsJSONSamplerEnabled() bool {
	s.mu.RLock()
	enabled := s.JSONSampler
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_JSON_SAMPLER") == "true"
}

// IsContextCrunchEnabled returns true if context-crunch layer is enabled.
func (s *AppState) IsContextCrunchEnabled() bool {
	s.mu.RLock()
	enabled := s.ContextCrunch
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_CONTEXT_CRUNCH") == "true"
}

// IsSearchCrunchEnabled returns true if search-crunch layer is enabled.
func (s *AppState) IsSearchCrunchEnabled() bool {
	s.mu.RLock()
	enabled := s.SearchCrunch
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_SEARCH_CRUNCH") == "true"
}

// IsStructCollapseEnabled returns true if structural-collapse layer is enabled.
func (s *AppState) IsStructCollapseEnabled() bool {
	s.mu.RLock()
	enabled := s.StructCollapse
	s.mu.RUnlock()
	return enabled || os.Getenv("TOK_STRUCTURAL_COLLAPSE") == "true"
}

// Global accessor functions for backward compatibility.
// These delegate to the global AppState instance and also sync package-level globals.
