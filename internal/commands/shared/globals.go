package shared

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	globalsMu sync.RWMutex

	CfgFile              string
	Verbose              int
	DryRun               bool
	UltraCompact         bool
	SkipEnv              bool
	QueryIntent          string
	LLMEnabled           bool
	TokenBudget          int
	FallbackArgs         []string
	LayerPreset          string
	LayerProfile         string
	OutputFile           string
	QuietMode            bool
	JSONOutput           bool
	RemoteMode           bool
	CompressionAddr      string
	AnalyticsAddr        string
	RemoteTimeout        int
	CompactionEnabled    bool
	CompactionThreshold  int
	CompactionPreserve   int
	CompactionMaxTokens  int
	CompactionSnapshot   bool
	CompactionAutoDetect bool
	ReversibleEnabled    bool
	EnableLayers         []string
	DisableLayers        []string
	StreamMode           bool
	PolicyRouter         bool
	Extractive           bool
	ExtractiveMax        int
	ExtractiveHead       int
	ExtractiveTail       int
	ExtractiveSignal     int
	QualityGuardrail     bool
	DiffAdapt            bool
	EPiC                 bool
	SSDP                 bool
	AgentOCR             bool
	S2MAD                bool
	ACON                 bool
	ResearchPack         bool
	LatentCollab         bool
	GraphCoT             bool
	RoleBudget           bool
	SWEAdaptive          bool
	AgentOCRHistory      bool
	PlanBudget           bool
	LightMem             bool
	PathShorten          bool
	JSONSampler          bool
	ContextCrunch        bool
	SearchCrunch         bool
	StructCollapse       bool
)

// syncGlobals copies AppState fields to package-level globals.
func (s *AppState) syncGlobals() {
	// Read AppState under s.mu, then write globals under globalsMu.
	// Never hold both locks to avoid deadlock.
	s.mu.RLock()
	state := FlagConfig{
		Verbose:              s.Verbose,
		DryRun:               s.DryRun,
		UltraCompact:         s.UltraCompact,
		SkipEnv:              s.SkipEnv,
		QueryIntent:          s.QueryIntent,
		LLMEnabled:           s.LLMEnabled,
		TokenBudget:          s.TokenBudget,
		FallbackArgs:         s.FallbackArgs,
		LayerPreset:          s.LayerPreset,
		LayerProfile:         s.LayerProfile,
		OutputFile:           s.OutputFile,
		QuietMode:            s.QuietMode,
		JSONOutput:           s.JSONOutput,
		RemoteMode:           s.RemoteMode,
		CompressionAddr:      s.CompressionAddr,
		AnalyticsAddr:        s.AnalyticsAddr,
		RemoteTimeout:        s.RemoteTimeout,
		CompactionEnabled:    s.CompactionEnabled,
		CompactionThreshold:  s.CompactionThreshold,
		CompactionPreserve:   s.CompactionPreserve,
		CompactionMaxTokens:  s.CompactionMaxTokens,
		CompactionSnapshot:   s.CompactionSnapshot,
		CompactionAutoDetect: s.CompactionAutoDetect,
		ReversibleEnabled:    s.ReversibleEnabled,
		EnableLayers:         s.EnableLayers,
		DisableLayers:        s.DisableLayers,
		StreamMode:           s.StreamMode,
		PolicyRouter:         s.PolicyRouter,
		Extractive:           s.Extractive,
		ExtractiveMax:        s.ExtractiveMax,
		ExtractiveHead:       s.ExtractiveHead,
		ExtractiveTail:       s.ExtractiveTail,
		ExtractiveSignal:     s.ExtractiveSignal,
		QualityGuardrail:     s.QualityGuardrail,
		DiffAdapt:            s.DiffAdapt,
		EPiC:                 s.EPiC,
		SSDP:                 s.SSDP,
		AgentOCR:             s.AgentOCR,
		S2MAD:                s.S2MAD,
		ACON:                 s.ACON,
		ResearchPack:         s.ResearchPack,
		LatentCollab:         s.LatentCollab,
		GraphCoT:             s.GraphCoT,
		RoleBudget:           s.RoleBudget,
		SWEAdaptive:          s.SWEAdaptive,
		AgentOCRHistory:      s.AgentOCRHistory,
		PlanBudget:           s.PlanBudget,
		LightMem:             s.LightMem,
		PathShorten:          s.PathShorten,
		JSONSampler:          s.JSONSampler,
		ContextCrunch:        s.ContextCrunch,
		SearchCrunch:         s.SearchCrunch,
		StructCollapse:       s.StructCollapse,
	}
	s.mu.RUnlock()

	globalsMu.Lock()
	Verbose = state.Verbose
	DryRun = state.DryRun
	UltraCompact = state.UltraCompact
	SkipEnv = state.SkipEnv
	QueryIntent = state.QueryIntent
	LLMEnabled = state.LLMEnabled
	TokenBudget = state.TokenBudget
	FallbackArgs = state.FallbackArgs
	LayerPreset = state.LayerPreset
	LayerProfile = state.LayerProfile
	OutputFile = state.OutputFile
	QuietMode = state.QuietMode
	JSONOutput = state.JSONOutput
	RemoteMode = state.RemoteMode
	CompressionAddr = state.CompressionAddr
	AnalyticsAddr = state.AnalyticsAddr
	RemoteTimeout = state.RemoteTimeout
	CompactionEnabled = state.CompactionEnabled
	CompactionThreshold = state.CompactionThreshold
	CompactionPreserve = state.CompactionPreserve
	CompactionMaxTokens = state.CompactionMaxTokens
	CompactionSnapshot = state.CompactionSnapshot
	CompactionAutoDetect = state.CompactionAutoDetect
	ReversibleEnabled = state.ReversibleEnabled
	EnableLayers = state.EnableLayers
	DisableLayers = state.DisableLayers
	StreamMode = state.StreamMode
	PolicyRouter = state.PolicyRouter
	Extractive = state.Extractive
	ExtractiveMax = state.ExtractiveMax
	ExtractiveHead = state.ExtractiveHead
	ExtractiveTail = state.ExtractiveTail
	ExtractiveSignal = state.ExtractiveSignal
	QualityGuardrail = state.QualityGuardrail
	DiffAdapt = state.DiffAdapt
	EPiC = state.EPiC
	SSDP = state.SSDP
	AgentOCR = state.AgentOCR
	S2MAD = state.S2MAD
	ACON = state.ACON
	ResearchPack = state.ResearchPack
	LatentCollab = state.LatentCollab
	GraphCoT = state.GraphCoT
	RoleBudget = state.RoleBudget
	SWEAdaptive = state.SWEAdaptive
	AgentOCRHistory = state.AgentOCRHistory
	PlanBudget = state.PlanBudget
	LightMem = state.LightMem
	PathShorten = state.PathShorten
	JSONSampler = state.JSONSampler
	ContextCrunch = state.ContextCrunch
	SearchCrunch = state.SearchCrunch
	StructCollapse = state.StructCollapse
	globalsMu.Unlock()
}

// syncFromGlobals copies package-level globals into AppState.
func (s *AppState) syncFromGlobals() {
	// Read globals under globalsMu, then write AppState under s.mu.
	// Never hold both locks to avoid deadlock.
	globalsMu.RLock()
	globals := FlagConfig{
		Verbose:              Verbose,
		DryRun:               DryRun,
		UltraCompact:         UltraCompact,
		SkipEnv:              SkipEnv,
		QueryIntent:          QueryIntent,
		LLMEnabled:           LLMEnabled,
		TokenBudget:          TokenBudget,
		FallbackArgs:         FallbackArgs,
		LayerPreset:          LayerPreset,
		LayerProfile:         LayerProfile,
		OutputFile:           OutputFile,
		QuietMode:            QuietMode,
		JSONOutput:           JSONOutput,
		RemoteMode:           RemoteMode,
		CompressionAddr:      CompressionAddr,
		AnalyticsAddr:        AnalyticsAddr,
		RemoteTimeout:        RemoteTimeout,
		CompactionEnabled:    CompactionEnabled,
		CompactionThreshold:  CompactionThreshold,
		CompactionPreserve:   CompactionPreserve,
		CompactionMaxTokens:  CompactionMaxTokens,
		CompactionSnapshot:   CompactionSnapshot,
		CompactionAutoDetect: CompactionAutoDetect,
		ReversibleEnabled:    ReversibleEnabled,
		EnableLayers:         EnableLayers,
		DisableLayers:        DisableLayers,
		StreamMode:           StreamMode,
		PolicyRouter:         PolicyRouter,
		Extractive:           Extractive,
		ExtractiveMax:        ExtractiveMax,
		ExtractiveHead:       ExtractiveHead,
		ExtractiveTail:       ExtractiveTail,
		ExtractiveSignal:     ExtractiveSignal,
		QualityGuardrail:     QualityGuardrail,
		DiffAdapt:            DiffAdapt,
		EPiC:                 EPiC,
		SSDP:                 SSDP,
		AgentOCR:             AgentOCR,
		S2MAD:                S2MAD,
		ACON:                 ACON,
		ResearchPack:         ResearchPack,
		LatentCollab:         LatentCollab,
		GraphCoT:             GraphCoT,
		RoleBudget:           RoleBudget,
		SWEAdaptive:          SWEAdaptive,
		AgentOCRHistory:      AgentOCRHistory,
		PlanBudget:           PlanBudget,
		LightMem:             LightMem,
		PathShorten:          PathShorten,
		JSONSampler:          JSONSampler,
		ContextCrunch:        ContextCrunch,
		SearchCrunch:         SearchCrunch,
		StructCollapse:       StructCollapse,
	}
	globalsMu.RUnlock()

	s.Set(globals)
}

// IsVerbose returns true if verbose mode is enabled.
func IsVerbose() bool {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return Verbose > 0
}

// IsUltraCompact returns true if ultra-compact mode is enabled.
func IsUltraCompact() bool {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return UltraCompact
}

// GetQueryIntent returns the query intent from flag or environment.
func GetQueryIntent() string {
	globalsMu.RLock()
	intent := QueryIntent
	globalsMu.RUnlock()
	if intent != "" {
		return intent
	}
	return os.Getenv("TOK_QUERY")
}

// IsLLMEnabled returns true if LLM compression is enabled.
func IsLLMEnabled() bool {
	globalsMu.RLock()
	enabled := LLMEnabled
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_LLM") == "true"
}

// GetTokenBudget returns the token budget from flag or environment.
func GetTokenBudget() int {
	globalsMu.RLock()
	budget := TokenBudget
	globalsMu.RUnlock()
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

// GetLayerPreset returns the layer preset from flag or environment.
func GetLayerPreset() string {
	globalsMu.RLock()
	preset := LayerPreset
	globalsMu.RUnlock()
	if preset != "" {
		return preset
	}
	return os.Getenv("TOK_PRESET")
}

// GetLayerProfile returns the compression profile from flag or environment.
func GetLayerProfile() string {
	globalsMu.RLock()
	profile := LayerProfile
	globalsMu.RUnlock()
	if profile != "" {
		return profile
	}
	return os.Getenv("TOK_PROFILE")
}

// IsQuietMode returns true if quiet mode is enabled.
func IsQuietMode() bool {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return QuietMode
}

// IsReversibleEnabled returns true if reversible mode is enabled.
func IsReversibleEnabled() bool {
	globalsMu.RLock()
	enabled := ReversibleEnabled
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_REVERSIBLE") == "true"
}

// IsRemoteMode returns true if remote mode is enabled.
func IsRemoteMode() bool {
	globalsMu.RLock()
	enabled := RemoteMode
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_REMOTE") == "true"
}

// GetCompressionAddr returns the compression service address.
func GetCompressionAddr() string {
	globalsMu.RLock()
	addr := CompressionAddr
	globalsMu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOK_COMPRESSION_ADDR")
}

// GetAnalyticsAddr returns the analytics service address.
func GetAnalyticsAddr() string {
	globalsMu.RLock()
	addr := AnalyticsAddr
	globalsMu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOK_ANALYTICS_ADDR")
}

// GetRemoteTimeout returns the remote operation timeout in seconds.
func GetRemoteTimeout() int {
	globalsMu.RLock()
	timeout := RemoteTimeout
	globalsMu.RUnlock()
	if timeout > 0 {
		return timeout
	}
	return 30
}

// GetEnableLayers returns layers to explicitly enable.
func GetEnableLayers() []string {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return EnableLayers
}

// GetDisableLayers returns layers to explicitly disable.
func GetDisableLayers() []string {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return DisableLayers
}

// IsStreamMode returns true if streaming mode is enabled for large inputs.
func IsStreamMode() bool {
	globalsMu.RLock()
	defer globalsMu.RUnlock()
	return StreamMode
}

// IsPolicyRouterEnabled returns true if policy router is enabled.
func IsPolicyRouterEnabled() bool {
	globalsMu.RLock()
	enabled := PolicyRouter
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_POLICY_ROUTER") == "true"
}

// IsExtractiveEnabled returns true if extractive prefilter is enabled.
func IsExtractiveEnabled() bool {
	globalsMu.RLock()
	enabled := Extractive
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_EXTRACTIVE_PREFILTER") == "true"
}

// GetExtractiveMax returns max lines for extractive prefilter.
func GetExtractiveMax() int {
	globalsMu.RLock()
	v := ExtractiveMax
	globalsMu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_MAX_LINES", 400)
}

// GetExtractiveHead returns preserved head lines for extractive prefilter.
func GetExtractiveHead() int {
	globalsMu.RLock()
	v := ExtractiveHead
	globalsMu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_HEAD_LINES", 80)
}

// GetExtractiveTail returns preserved tail lines for extractive prefilter.
func GetExtractiveTail() int {
	globalsMu.RLock()
	v := ExtractiveTail
	globalsMu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_TAIL_LINES", 60)
}

// GetExtractiveSignal returns signal line budget for extractive prefilter.
func GetExtractiveSignal() int {
	globalsMu.RLock()
	v := ExtractiveSignal
	globalsMu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOK_EXTRACTIVE_SIGNAL_LINES", 120)
}

// IsQualityGuardrailEnabled returns true if quality guardrail is enabled.
func IsQualityGuardrailEnabled() bool {
	globalsMu.RLock()
	enabled := QualityGuardrail
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_QUALITY_GUARDRAIL") == "true"
}

// IsDiffAdaptEnabled returns true if DiffAdapt layer is enabled.
func IsDiffAdaptEnabled() bool {
	globalsMu.RLock()
	enabled := DiffAdapt
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_DIFF_ADAPT") == "true"
}

// IsEPiCEnabled returns true if EPiC layer is enabled.
func IsEPiCEnabled() bool {
	globalsMu.RLock()
	enabled := EPiC
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_EPIC") == "true"
}

// IsSSDPEnabled returns true if SSDP layer is enabled.
func IsSSDPEnabled() bool {
	globalsMu.RLock()
	enabled := SSDP
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_SSDP") == "true"
}

// IsAgentOCREnabled returns true if AgentOCR layer is enabled.
func IsAgentOCREnabled() bool {
	globalsMu.RLock()
	enabled := AgentOCR
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_AGENT_OCR") == "true"
}

// IsS2MADEnabled returns true if S2-MAD layer is enabled.
func IsS2MADEnabled() bool {
	globalsMu.RLock()
	enabled := S2MAD
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_S2_MAD") == "true"
}

// IsACONEnabled returns true if ACON layer is enabled.
func IsACONEnabled() bool {
	globalsMu.RLock()
	enabled := ACON
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_ACON") == "true"
}

// IsResearchPackEnabled returns true if research layer pack is enabled.
func IsResearchPackEnabled() bool {
	globalsMu.RLock()
	enabled := ResearchPack
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_RESEARCH_PACK") == "true"
}

// IsLatentCollabEnabled returns true if latent-collab layer is enabled.
func IsLatentCollabEnabled() bool {
	globalsMu.RLock()
	enabled := LatentCollab
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_LATENT_COLLAB") == "true"
}

// IsGraphCoTEnabled returns true if graph-cot layer is enabled.
func IsGraphCoTEnabled() bool {
	globalsMu.RLock()
	enabled := GraphCoT
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_GRAPH_COT") == "true"
}

// IsRoleBudgetEnabled returns true if role-budget layer is enabled.
func IsRoleBudgetEnabled() bool {
	globalsMu.RLock()
	enabled := RoleBudget
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_ROLE_BUDGET") == "true"
}

// IsSWEAdaptiveEnabled returns true if swe-adaptive-loop layer is enabled.
func IsSWEAdaptiveEnabled() bool {
	globalsMu.RLock()
	enabled := SWEAdaptive
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_SWE_ADAPTIVE") == "true"
}

// IsAgentOCRHistoryEnabled returns true if agent-ocr-history layer is enabled.
func IsAgentOCRHistoryEnabled() bool {
	globalsMu.RLock()
	enabled := AgentOCRHistory
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_AGENT_OCR_HISTORY") == "true"
}

// IsPlanBudgetEnabled returns true if plan-budget layer is enabled.
func IsPlanBudgetEnabled() bool {
	globalsMu.RLock()
	enabled := PlanBudget
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_PLAN_BUDGET") == "true"
}

// IsLightMemEnabled returns true if lightmem layer is enabled.
func IsLightMemEnabled() bool {
	globalsMu.RLock()
	enabled := LightMem
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_LIGHTMEM") == "true"
}

// IsPathShortenEnabled returns true if path-shorten layer is enabled.
func IsPathShortenEnabled() bool {
	globalsMu.RLock()
	enabled := PathShorten
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_PATH_SHORTEN") == "true"
}

// IsJSONSamplerEnabled returns true if json-sampler layer is enabled.
func IsJSONSamplerEnabled() bool {
	globalsMu.RLock()
	enabled := JSONSampler
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_JSON_SAMPLER") == "true"
}

// IsContextCrunchEnabled returns true if context-crunch layer is enabled.
func IsContextCrunchEnabled() bool {
	globalsMu.RLock()
	enabled := ContextCrunch
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_CONTEXT_CRUNCH") == "true"
}

// IsSearchCrunchEnabled returns true if search-crunch layer is enabled.
func IsSearchCrunchEnabled() bool {
	globalsMu.RLock()
	enabled := SearchCrunch
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_SEARCH_CRUNCH") == "true"
}

// IsStructCollapseEnabled returns true if structural-collapse layer is enabled.
func IsStructCollapseEnabled() bool {
	globalsMu.RLock()
	enabled := StructCollapse
	globalsMu.RUnlock()
	return enabled || os.Getenv("TOK_STRUCTURAL_COLLAPSE") == "true"
}
