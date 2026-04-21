package shared

import (
	"context"
	"os"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/filter"
)

// CLIContext provides dependency injection for all CLI operations.
// It replaces global state access with explicit context passing.
type CLIContext struct {
	// Core configuration
	Config *config.Config

	// Feature flags (consolidated from 44+ global functions)
	Flags FeatureFlags

	// Services
	FilterMode filter.Mode

	// I/O (testable)
	Stdout *os.File
	Stderr *os.File
	Stdin  *os.File
}

// FeatureFlags consolidates all CLI feature flags into one struct.
// This replaces the 44+ global accessor functions in globals.go
type FeatureFlags struct {
	// Output control
	Verbose      int // 0=warn, 1=info, 2=debug, 3=trace
	QuietMode    bool
	UltraCompact bool
	JSONOutput   bool

	// Core functionality
	DryRun      bool
	LLMEnabled  bool
	TokenBudget int
	QueryIntent string

	// Layer configuration
	LayerPreset   string
	LayerProfile  string
	EnableLayers  []string
	DisableLayers []string

	// Processing modes
	StreamMode        bool
	ReversibleEnabled bool
	RemoteMode        bool

	// Remote settings
	CompressionAddr string
	AnalyticsAddr   string
	RemoteTimeout   int

	// Compaction
	CompactionEnabled    bool
	CompactionThreshold  int
	CompactionPreserve   int
	CompactionMaxTokens  int
	CompactionSnapshot   bool
	CompactionAutoDetect bool

	// Extractive prefilter
	Extractive       bool
	ExtractiveMax    int
	ExtractiveHead   int
	ExtractiveTail   int
	ExtractiveSignal int

	// Quality and routing
	PolicyRouter     bool
	QualityGuardrail bool

	// Research layers (consolidated)
	ResearchPack bool
	DiffAdapt    bool
	EPiC         bool
	SSDP         bool
	ACON         bool
	LatentCollab bool
	GraphCoT     bool
	RoleBudget   bool
	SWEAdaptive  bool
	PlanBudget   bool
	LightMem     bool

	// Agent layers (consolidated)
	AgentOCR        bool
	AgentOCRHistory bool
	S2MAD           bool

	// Utility layers (consolidated)
	PathShorten      bool
	JSONSampler      bool
	ContextCrunch    bool
	SearchCrunch     bool
	StructCollapse   bool
	AdaptiveLearning bool
}

// NewCLIContext creates a new CLI context from AppState.
// This is the migration path from global state to DI.
func NewCLIContext(state *AppState) *CLIContext {
	state.mu.RLock()
	defer state.mu.RUnlock()

	return &CLIContext{
		Flags: FeatureFlags{
			Verbose:              state.Verbose,
			QuietMode:            state.QuietMode,
			UltraCompact:         state.UltraCompact,
			JSONOutput:           state.JSONOutput,
			DryRun:               state.DryRun,
			LLMEnabled:           state.LLMEnabled,
			TokenBudget:          state.TokenBudget,
			QueryIntent:          state.QueryIntent,
			LayerPreset:          state.LayerPreset,
			LayerProfile:         state.LayerProfile,
			EnableLayers:         state.EnableLayers,
			DisableLayers:        state.DisableLayers,
			StreamMode:           state.StreamMode,
			ReversibleEnabled:    state.ReversibleEnabled,
			RemoteMode:           state.RemoteMode,
			CompressionAddr:      state.CompressionAddr,
			AnalyticsAddr:        state.AnalyticsAddr,
			RemoteTimeout:        state.RemoteTimeout,
			CompactionEnabled:    state.CompactionEnabled,
			CompactionThreshold:  state.CompactionThreshold,
			CompactionPreserve:   state.CompactionPreserve,
			CompactionMaxTokens:  state.CompactionMaxTokens,
			CompactionSnapshot:   state.CompactionSnapshot,
			CompactionAutoDetect: state.CompactionAutoDetect,
			Extractive:           state.Extractive,
			ExtractiveMax:        state.ExtractiveMax,
			ExtractiveHead:       state.ExtractiveHead,
			ExtractiveTail:       state.ExtractiveTail,
			ExtractiveSignal:     state.ExtractiveSignal,
			PolicyRouter:         state.PolicyRouter,
			QualityGuardrail:     state.QualityGuardrail,
			ResearchPack:         state.ResearchPack,
			DiffAdapt:            state.DiffAdapt,
			EPiC:                 state.EPiC,
			SSDP:                 state.SSDP,
			ACON:                 state.ACON,
			LatentCollab:         state.LatentCollab,
			GraphCoT:             state.GraphCoT,
			RoleBudget:           state.RoleBudget,
			SWEAdaptive:          state.SWEAdaptive,
			PlanBudget:           state.PlanBudget,
			LightMem:             state.LightMem,
			AgentOCR:             state.AgentOCR,
			AgentOCRHistory:      state.AgentOCRHistory,
			S2MAD:                state.S2MAD,
			PathShorten:          state.PathShorten,
			JSONSampler:          state.JSONSampler,
			ContextCrunch:        state.ContextCrunch,
			SearchCrunch:         state.SearchCrunch,
			StructCollapse:       state.StructCollapse,
			AdaptiveLearning:     state.AdaptiveLearning,
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}

// ContextKey is the key for storing CLIContext in context.Context
type ContextKey struct{}

// WithContext stores the CLIContext in a context.Context.
func WithContext(ctx context.Context, cliCtx *CLIContext) context.Context {
	return context.WithValue(ctx, ContextKey{}, cliCtx)
}

// FromContext retrieves the CLIContext from a context.Context.
// Returns nil if not found (callers should handle gracefully).
func FromContext(ctx context.Context) *CLIContext {
	if ctx == nil {
		return nil
	}
	if cliCtx, ok := ctx.Value(ContextKey{}).(*CLIContext); ok {
		return cliCtx
	}
	return nil
}

// MustFromContext retrieves the CLIContext or panics.
// Use only in commands where context is guaranteed to be set.
func MustFromContext(ctx context.Context) *CLIContext {
	cliCtx := FromContext(ctx)
	if cliCtx == nil {
		panic("CLIContext not found in context")
	}
	return cliCtx
}

// Convenience methods for common flag checks
func (c *CLIContext) IsVerbose() bool            { return c.Flags.Verbose > 0 }
func (c *CLIContext) IsDebug() bool              { return c.Flags.Verbose >= 2 }
func (c *CLIContext) IsUltraCompact() bool       { return c.Flags.UltraCompact }
func (c *CLIContext) IsQuietMode() bool          { return c.Flags.QuietMode }
func (c *CLIContext) IsLLMEnabled() bool         { return c.Flags.LLMEnabled }
func (c *CLIContext) IsStreamMode() bool         { return c.Flags.StreamMode }
func (c *CLIContext) IsRemoteMode() bool         { return c.Flags.RemoteMode }
func (c *CLIContext) IsReversibleEnabled() bool  { return c.Flags.ReversibleEnabled }
func (c *CLIContext) GetTokenBudget() int        { return c.Flags.TokenBudget }
func (c *CLIContext) GetLayerPreset() string     { return c.Flags.LayerPreset }
func (c *CLIContext) GetQueryIntent() string     { return c.Flags.QueryIntent }
func (c *CLIContext) GetCompressionAddr() string { return c.Flags.CompressionAddr }
func (c *CLIContext) GetAnalyticsAddr() string   { return c.Flags.AnalyticsAddr }
func (c *CLIContext) GetRemoteTimeout() int      { return c.Flags.RemoteTimeout }

// GetFilterMode returns the appropriate filter mode based on flags.
func (c *CLIContext) GetFilterMode() filter.Mode {
	if c.Flags.UltraCompact {
		return filter.ModeAggressive
	}
	return filter.ModeMinimal
}
