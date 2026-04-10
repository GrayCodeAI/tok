package shared

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// AppState encapsulates all CLI flag state in a single struct.
// This replaces the global variable pattern and enables:
// - Testability (pass different state to different tests)
// - Concurrency (multiple commands with different configs)
// - Dependency injection (pass state explicitly)
type AppState struct {
	mu sync.RWMutex

	// Global flags set by CLI
	CfgFile      string
	Verbose      int
	DryRun       bool
	UltraCompact bool
	SkipEnv      bool
	QueryIntent  string
	LLMEnabled   bool
	TokenBudget  int
	FallbackArgs []string
	LayerPreset  string
	LayerProfile string
	OutputFile   string
	QuietMode    bool
	JSONOutput   bool

	// Remote mode flags (Phase 4)
	RemoteMode      bool
	CompressionAddr string
	AnalyticsAddr   string
	RemoteTimeout   int // seconds

	// Compaction flags
	CompactionEnabled    bool
	CompactionThreshold  int
	CompactionPreserve   int
	CompactionMaxTokens  int
	CompactionSnapshot   bool
	CompactionAutoDetect bool

	// Reversible mode
	ReversibleEnabled bool

	// Custom layer configuration
	EnableLayers     []string
	DisableLayers    []string
	StreamMode       bool
	PolicyRouter     bool
	Extractive       bool
	ExtractiveMax    int
	ExtractiveHead   int
	ExtractiveTail   int
	ExtractiveSignal int
	QualityGuardrail bool
	DiffAdapt        bool
	EPiC             bool
	SSDP             bool
	AgentOCR         bool
	S2MAD            bool
	ACON             bool
	ResearchPack     bool
	LatentCollab     bool
	GraphCoT         bool
	RoleBudget       bool
	SWEAdaptive      bool
	AgentOCRHistory  bool
	PlanBudget       bool
	LightMem         bool
	PathShorten      bool
	JSONSampler      bool
	LogCrunch        bool
	SearchCrunch     bool
	DiffCrunch       bool
	StructCollapse   bool
}

// Version is set at build time.
var Version string = "dev"

// Global instance for backward compatibility.
// New code should use explicit AppState instances.
var globalState = &AppState{}

// Global returns the global AppState instance.
// Deprecated: Pass AppState explicitly where possible.
func Global() *AppState {
	return globalState
}

// SetFlags sets all flag values atomically on the global state.
func SetFlags(cfg FlagConfig) {
	globalState.Set(cfg)
	globalState.syncGlobals()
}

// FlagConfig holds all flag values for atomic setting.
type FlagConfig struct {
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
	LogCrunch            bool
	SearchCrunch         bool
	DiffCrunch           bool
	StructCollapse       bool
}

// Set sets all flag values atomically.
func (s *AppState) Set(cfg FlagConfig) {
	s.mu.Lock()
	s.Verbose = cfg.Verbose
	s.DryRun = cfg.DryRun
	s.UltraCompact = cfg.UltraCompact
	s.SkipEnv = cfg.SkipEnv
	s.QueryIntent = cfg.QueryIntent
	s.LLMEnabled = cfg.LLMEnabled
	s.TokenBudget = cfg.TokenBudget
	s.FallbackArgs = cfg.FallbackArgs
	s.LayerPreset = cfg.LayerPreset
	s.LayerProfile = cfg.LayerProfile
	s.OutputFile = cfg.OutputFile
	s.QuietMode = cfg.QuietMode
	s.JSONOutput = cfg.JSONOutput
	s.RemoteMode = cfg.RemoteMode
	s.CompressionAddr = cfg.CompressionAddr
	s.AnalyticsAddr = cfg.AnalyticsAddr
	s.RemoteTimeout = cfg.RemoteTimeout
	s.CompactionEnabled = cfg.CompactionEnabled
	s.CompactionThreshold = cfg.CompactionThreshold
	s.CompactionPreserve = cfg.CompactionPreserve
	s.CompactionMaxTokens = cfg.CompactionMaxTokens
	s.CompactionSnapshot = cfg.CompactionSnapshot
	s.CompactionAutoDetect = cfg.CompactionAutoDetect
	s.ReversibleEnabled = cfg.ReversibleEnabled
	s.EnableLayers = cfg.EnableLayers
	s.DisableLayers = cfg.DisableLayers
	s.StreamMode = cfg.StreamMode
	s.PolicyRouter = cfg.PolicyRouter
	s.Extractive = cfg.Extractive
	s.ExtractiveMax = cfg.ExtractiveMax
	s.ExtractiveHead = cfg.ExtractiveHead
	s.ExtractiveTail = cfg.ExtractiveTail
	s.ExtractiveSignal = cfg.ExtractiveSignal
	s.QualityGuardrail = cfg.QualityGuardrail
	s.DiffAdapt = cfg.DiffAdapt
	s.EPiC = cfg.EPiC
	s.SSDP = cfg.SSDP
	s.AgentOCR = cfg.AgentOCR
	s.S2MAD = cfg.S2MAD
	s.ACON = cfg.ACON
	s.ResearchPack = cfg.ResearchPack
	s.LatentCollab = cfg.LatentCollab
	s.GraphCoT = cfg.GraphCoT
	s.RoleBudget = cfg.RoleBudget
	s.SWEAdaptive = cfg.SWEAdaptive
	s.AgentOCRHistory = cfg.AgentOCRHistory
	s.PlanBudget = cfg.PlanBudget
	s.LightMem = cfg.LightMem
	s.PathShorten = cfg.PathShorten
	s.JSONSampler = cfg.JSONSampler
	s.LogCrunch = cfg.LogCrunch
	s.SearchCrunch = cfg.SearchCrunch
	s.DiffCrunch = cfg.DiffCrunch
	s.StructCollapse = cfg.StructCollapse
	s.mu.Unlock()
}

// IsVerbose returns true if verbose mode is enabled.
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
	return os.Getenv("TOKMAN_QUERY")
}

// IsLLMEnabled returns true if LLM compression is enabled.
func (s *AppState) IsLLMEnabled() bool {
	s.mu.RLock()
	enabled := s.LLMEnabled
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_LLM") == "true"
}

// GetTokenBudget returns the token budget from flag or environment.
func (s *AppState) GetTokenBudget() int {
	s.mu.RLock()
	budget := s.TokenBudget
	s.mu.RUnlock()
	if budget > 0 {
		return budget
	}
	envBudget := os.Getenv("TOKMAN_BUDGET")
	if envBudget != "" {
		var b int
		if _, err := fmt.Sscanf(envBudget, "%d", &b); err == nil {
			return b
		}
		log.Printf("warning: invalid TOKMAN_BUDGET value %q, defaulting to unlimited", envBudget)
	}
	return 0
}

// GetLayerPreset returns the layer preset from flag or environment.
func (s *AppState) GetLayerPreset() string {
	s.mu.RLock()
	preset := s.LayerPreset
	s.mu.RUnlock()
	if preset != "" {
		return preset
	}
	return os.Getenv("TOKMAN_PRESET")
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
	return enabled || os.Getenv("TOKMAN_REVERSIBLE") == "true"
}

// IsRemoteMode returns true if remote mode is enabled.
func (s *AppState) IsRemoteMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RemoteMode || os.Getenv("TOKMAN_REMOTE") == "true"
}

// GetCompressionAddr returns the compression service address.
func (s *AppState) GetCompressionAddr() string {
	s.mu.RLock()
	addr := s.CompressionAddr
	s.mu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOKMAN_COMPRESSION_ADDR")
}

// GetAnalyticsAddr returns the analytics service address.
func (s *AppState) GetAnalyticsAddr() string {
	s.mu.RLock()
	addr := s.AnalyticsAddr
	s.mu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOKMAN_ANALYTICS_ADDR")
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
	return enabled || os.Getenv("TOKMAN_POLICY_ROUTER") == "true"
}

// IsExtractiveEnabled returns true if extractive prefilter is enabled.
func (s *AppState) IsExtractiveEnabled() bool {
	s.mu.RLock()
	enabled := s.Extractive
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_EXTRACTIVE_PREFILTER") == "true"
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
	return envInt("TOKMAN_EXTRACTIVE_MAX_LINES", 400)
}

// GetExtractiveHead returns preserved head lines for extractive prefilter.
func (s *AppState) GetExtractiveHead() int {
	s.mu.RLock()
	v := s.ExtractiveHead
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOKMAN_EXTRACTIVE_HEAD_LINES", 80)
}

// GetExtractiveTail returns preserved tail lines for extractive prefilter.
func (s *AppState) GetExtractiveTail() int {
	s.mu.RLock()
	v := s.ExtractiveTail
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOKMAN_EXTRACTIVE_TAIL_LINES", 60)
}

// GetExtractiveSignal returns signal line budget for extractive prefilter.
func (s *AppState) GetExtractiveSignal() int {
	s.mu.RLock()
	v := s.ExtractiveSignal
	s.mu.RUnlock()
	if v > 0 {
		return v
	}
	return envInt("TOKMAN_EXTRACTIVE_SIGNAL_LINES", 120)
}

// IsQualityGuardrailEnabled returns true if quality guardrail is enabled.
func (s *AppState) IsQualityGuardrailEnabled() bool {
	s.mu.RLock()
	enabled := s.QualityGuardrail
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_QUALITY_GUARDRAIL") == "true"
}

// IsDiffAdaptEnabled returns true if DiffAdapt layer is enabled.
func (s *AppState) IsDiffAdaptEnabled() bool {
	s.mu.RLock()
	enabled := s.DiffAdapt
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_DIFF_ADAPT") == "true"
}

// IsEPiCEnabled returns true if EPiC layer is enabled.
func (s *AppState) IsEPiCEnabled() bool {
	s.mu.RLock()
	enabled := s.EPiC
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_EPIC") == "true"
}

// IsSSDPEnabled returns true if SSDP layer is enabled.
func (s *AppState) IsSSDPEnabled() bool {
	s.mu.RLock()
	enabled := s.SSDP
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_SSDP") == "true"
}

// IsAgentOCREnabled returns true if AgentOCR layer is enabled.
func (s *AppState) IsAgentOCREnabled() bool {
	s.mu.RLock()
	enabled := s.AgentOCR
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_AGENT_OCR") == "true"
}

// IsS2MADEnabled returns true if S2-MAD layer is enabled.
func (s *AppState) IsS2MADEnabled() bool {
	s.mu.RLock()
	enabled := s.S2MAD
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_S2_MAD") == "true"
}

// IsACONEnabled returns true if ACON layer is enabled.
func (s *AppState) IsACONEnabled() bool {
	s.mu.RLock()
	enabled := s.ACON
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_ACON") == "true"
}

// IsResearchPackEnabled returns true if research pack is enabled.
func (s *AppState) IsResearchPackEnabled() bool {
	s.mu.RLock()
	enabled := s.ResearchPack
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_RESEARCH_PACK") == "true"
}

// IsLatentCollabEnabled returns true if latent collaboration layer is enabled.
func (s *AppState) IsLatentCollabEnabled() bool {
	s.mu.RLock()
	enabled := s.LatentCollab
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_LATENT_COLLAB") == "true"
}

// IsGraphCoTEnabled returns true if graph-CoT layer is enabled.
func (s *AppState) IsGraphCoTEnabled() bool {
	s.mu.RLock()
	enabled := s.GraphCoT
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_GRAPH_COT") == "true"
}

// IsRoleBudgetEnabled returns true if role-budget layer is enabled.
func (s *AppState) IsRoleBudgetEnabled() bool {
	s.mu.RLock()
	enabled := s.RoleBudget
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_ROLE_BUDGET") == "true"
}

// IsSWEAdaptiveEnabled returns true if SWE adaptive loop is enabled.
func (s *AppState) IsSWEAdaptiveEnabled() bool {
	s.mu.RLock()
	enabled := s.SWEAdaptive
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_SWE_ADAPTIVE") == "true"
}

// IsAgentOCRHistoryEnabled returns true if agent OCR history layer is enabled.
func (s *AppState) IsAgentOCRHistoryEnabled() bool {
	s.mu.RLock()
	enabled := s.AgentOCRHistory
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_AGENT_OCR_HISTORY") == "true"
}

// IsPlanBudgetEnabled returns true if plan-budget layer is enabled.
func (s *AppState) IsPlanBudgetEnabled() bool {
	s.mu.RLock()
	enabled := s.PlanBudget
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_PLAN_BUDGET") == "true"
}

// IsLightMemEnabled returns true if lightmem layer is enabled.
func (s *AppState) IsLightMemEnabled() bool {
	s.mu.RLock()
	enabled := s.LightMem
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_LIGHTMEM") == "true"
}

// IsPathShortenEnabled returns true if path-shorten layer is enabled.
func (s *AppState) IsPathShortenEnabled() bool {
	s.mu.RLock()
	enabled := s.PathShorten
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_PATH_SHORTEN") == "true"
}

// IsJSONSamplerEnabled returns true if json-sampler layer is enabled.
func (s *AppState) IsJSONSamplerEnabled() bool {
	s.mu.RLock()
	enabled := s.JSONSampler
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_JSON_SAMPLER") == "true"
}

// IsLogCrunchEnabled returns true if log-crunch layer is enabled.
func (s *AppState) IsLogCrunchEnabled() bool {
	s.mu.RLock()
	enabled := s.LogCrunch
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_LOG_CRUNCH") == "true"
}

// IsSearchCrunchEnabled returns true if search-crunch layer is enabled.
func (s *AppState) IsSearchCrunchEnabled() bool {
	s.mu.RLock()
	enabled := s.SearchCrunch
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_SEARCH_CRUNCH") == "true"
}

// IsDiffCrunchEnabled returns true if diff-crunch layer is enabled.
func (s *AppState) IsDiffCrunchEnabled() bool {
	s.mu.RLock()
	enabled := s.DiffCrunch
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_DIFF_CRUNCH") == "true"
}

// IsStructCollapseEnabled returns true if structural-collapse layer is enabled.
func (s *AppState) IsStructCollapseEnabled() bool {
	s.mu.RLock()
	enabled := s.StructCollapse
	s.mu.RUnlock()
	return enabled || os.Getenv("TOKMAN_STRUCTURAL_COLLAPSE") == "true"
}

// Global accessor functions for backward compatibility.
// These delegate to the global AppState instance and also sync package-level globals.

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
	LogCrunch            bool
	SearchCrunch         bool
	DiffCrunch           bool
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
		LogCrunch:            s.LogCrunch,
		SearchCrunch:         s.SearchCrunch,
		DiffCrunch:           s.DiffCrunch,
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
	LogCrunch = state.LogCrunch
	SearchCrunch = state.SearchCrunch
	DiffCrunch = state.DiffCrunch
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
		LogCrunch:            LogCrunch,
		SearchCrunch:         SearchCrunch,
		DiffCrunch:           DiffCrunch,
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
	return os.Getenv("TOKMAN_PROFILE")
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

// IsLogCrunchEnabled returns true if log-crunch layer is enabled.
func IsLogCrunchEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsLogCrunchEnabled()
}

// IsSearchCrunchEnabled returns true if search-crunch layer is enabled.
func IsSearchCrunchEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsSearchCrunchEnabled()
}

// IsDiffCrunchEnabled returns true if diff-crunch layer is enabled.
func IsDiffCrunchEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsDiffCrunchEnabled()
}

// IsStructCollapseEnabled returns true if structural-collapse layer is enabled.
func IsStructCollapseEnabled() bool {
	globalState.syncFromGlobals()
	return globalState.IsStructCollapseEnabled()
}
