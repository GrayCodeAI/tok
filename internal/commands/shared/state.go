package shared

import (
	"sync"

	"github.com/GrayCodeAI/tok/internal/version"
)

// Version is the application version (re-exported from version package).
var Version = version.Version

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
	ContextCrunch    bool // Enable ContextCrunch (merged LogCrunch + DiffCrunch)
	SearchCrunch     bool
	StructCollapse   bool
	AdaptiveLearning bool // Enable AdaptiveLearning (merged EngramLearner + TieredSummary)
}

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
