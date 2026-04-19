package shared

import (
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
	globalState.syncFromGlobals()
	return globalState.IsVerbose()
}

// IsUltraCompact returns true if ultra-compact mode is enabled.
func IsUltraCompact() bool {
	globalState.syncFromGlobals()
	return globalState.IsUltraCompact()
}

// GetQueryIntent returns the query intent from flag or environment.
func GetQueryIntent() string {
	globalState.syncFromGlobals()
	return globalState.GetQueryIntent()
}

// IsLLMEnabled returns true if LLM compression is enabled.
func IsLLMEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsLLMEnabled()
}

// GetTokenBudget returns the token budget from flag or environment.
func GetTokenBudget() int {
	globalState.syncFromGlobals()
	return globalState.GetTokenBudget()
}

// GetLayerPreset returns the layer preset from flag or environment.
func GetLayerPreset() string {
	globalState.syncFromGlobals()
	return globalState.GetLayerPreset()
}

// GetLayerProfile returns the compression profile from flag or environment.
func GetLayerProfile() string {
	globalState.syncFromGlobals()
	profile := globalState.LayerProfile
	if profile != "" {
		return profile
	}
	return os.Getenv("TOK_PROFILE")
}

// IsQuietMode returns true if quiet mode is enabled.
func IsQuietMode() bool {
	globalState.syncFromGlobals()
	return globalState.IsQuietMode()
}

// IsReversibleEnabled returns true if reversible mode is enabled.
func IsReversibleEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsReversibleEnabled()
}

// IsRemoteMode returns true if remote mode is enabled.
func IsRemoteMode() bool {
	globalState.syncFromGlobals()
	return globalState.IsRemoteMode()
}

// GetCompressionAddr returns the compression service address.
func GetCompressionAddr() string {
	globalState.syncFromGlobals()
	return globalState.GetCompressionAddr()
}

// GetAnalyticsAddr returns the analytics service address.
func GetAnalyticsAddr() string {
	globalState.syncFromGlobals()
	return globalState.GetAnalyticsAddr()
}

// GetRemoteTimeout returns the remote operation timeout in seconds.
func GetRemoteTimeout() int {
	globalState.syncFromGlobals()
	return globalState.GetRemoteTimeout()
}

// GetEnableLayers returns layers to explicitly enable.
func GetEnableLayers() []string {
	globalState.syncFromGlobals()
	return globalState.GetEnableLayers()
}

// GetDisableLayers returns layers to explicitly disable.
func GetDisableLayers() []string {
	globalState.syncFromGlobals()
	return globalState.GetDisableLayers()
}

// IsStreamMode returns true if streaming mode is enabled for large inputs.
func IsStreamMode() bool {
	globalState.syncFromGlobals()
	return globalState.IsStreamMode()
}

// IsPolicyRouterEnabled returns true if policy router is enabled.
func IsPolicyRouterEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsPolicyRouterEnabled()
}

// IsExtractiveEnabled returns true if extractive prefilter is enabled.
func IsExtractiveEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsExtractiveEnabled()
}

// GetExtractiveMax returns max lines for extractive prefilter.
func GetExtractiveMax() int {
	globalState.syncFromGlobals()
	return globalState.GetExtractiveMax()
}

// GetExtractiveHead returns preserved head lines for extractive prefilter.
func GetExtractiveHead() int {
	globalState.syncFromGlobals()
	return globalState.GetExtractiveHead()
}

// GetExtractiveTail returns preserved tail lines for extractive prefilter.
func GetExtractiveTail() int {
	globalState.syncFromGlobals()
	return globalState.GetExtractiveTail()
}

// GetExtractiveSignal returns signal line budget for extractive prefilter.
func GetExtractiveSignal() int {
	globalState.syncFromGlobals()
	return globalState.GetExtractiveSignal()
}

// IsQualityGuardrailEnabled returns true if quality guardrail is enabled.
func IsQualityGuardrailEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsQualityGuardrailEnabled()
}

// IsDiffAdaptEnabled returns true if DiffAdapt layer is enabled.
func IsDiffAdaptEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsDiffAdaptEnabled()
}

// IsEPiCEnabled returns true if EPiC layer is enabled.
func IsEPiCEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsEPiCEnabled()
}

// IsSSDPEnabled returns true if SSDP layer is enabled.
func IsSSDPEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsSSDPEnabled()
}

// IsAgentOCREnabled returns true if AgentOCR layer is enabled.
func IsAgentOCREnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsAgentOCREnabled()
}

// IsS2MADEnabled returns true if S2-MAD layer is enabled.
func IsS2MADEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsS2MADEnabled()
}

// IsACONEnabled returns true if ACON layer is enabled.
func IsACONEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsACONEnabled()
}

// IsResearchPackEnabled returns true if research layer pack is enabled.
func IsResearchPackEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsResearchPackEnabled()
}

// IsLatentCollabEnabled returns true if latent-collab layer is enabled.
func IsLatentCollabEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsLatentCollabEnabled()
}

// IsGraphCoTEnabled returns true if graph-cot layer is enabled.
func IsGraphCoTEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsGraphCoTEnabled()
}

// IsRoleBudgetEnabled returns true if role-budget layer is enabled.
func IsRoleBudgetEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsRoleBudgetEnabled()
}

// IsSWEAdaptiveEnabled returns true if swe-adaptive-loop layer is enabled.
func IsSWEAdaptiveEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsSWEAdaptiveEnabled()
}

// IsAgentOCRHistoryEnabled returns true if agent-ocr-history layer is enabled.
func IsAgentOCRHistoryEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsAgentOCRHistoryEnabled()
}

// IsPlanBudgetEnabled returns true if plan-budget layer is enabled.
func IsPlanBudgetEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsPlanBudgetEnabled()
}

// IsLightMemEnabled returns true if lightmem layer is enabled.
func IsLightMemEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsLightMemEnabled()
}

// IsPathShortenEnabled returns true if path-shorten layer is enabled.
func IsPathShortenEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsPathShortenEnabled()
}

// IsJSONSamplerEnabled returns true if json-sampler layer is enabled.
func IsJSONSamplerEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsJSONSamplerEnabled()
}

// IsContextCrunchEnabled returns true if context-crunch layer is enabled.
func IsContextCrunchEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsContextCrunchEnabled()
}

// IsSearchCrunchEnabled returns true if search-crunch layer is enabled.
func IsSearchCrunchEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsSearchCrunchEnabled()
}

// IsStructCollapseEnabled returns true if structural-collapse layer is enabled.
func IsStructCollapseEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsStructCollapseEnabled()
}
