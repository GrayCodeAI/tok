package shared

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
