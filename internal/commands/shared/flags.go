package shared

import (
	"fmt"
	"os"
	"sync"
)

// Global flags and their thread-safe accessors.
// This file contains only flag state - no external dependencies.

var (
	configMu sync.RWMutex

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

	// Custom layer configuration (Task 5)
	EnableLayers  []string
	DisableLayers []string
	StreamMode    bool

	// Version (set at build time)
	Version string = "dev"
)

// IsVerbose returns true if verbose mode is enabled.
func IsVerbose() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return Verbose > 0
}

// IsUltraCompact returns true if ultra-compact mode is enabled.
func IsUltraCompact() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return UltraCompact
}

// GetQueryIntent returns the query intent from flag or environment.
func GetQueryIntent() string {
	configMu.RLock()
	intent := QueryIntent
	configMu.RUnlock()
	if intent != "" {
		return intent
	}
	return os.Getenv("TOKMAN_QUERY")
}

// IsLLMEnabled returns true if LLM compression is enabled.
func IsLLMEnabled() bool {
	configMu.RLock()
	enabled := LLMEnabled
	configMu.RUnlock()
	return enabled || os.Getenv("TOKMAN_LLM") == "true"
}

// GetTokenBudget returns the token budget from flag or environment.
func GetTokenBudget() int {
	configMu.RLock()
	budget := TokenBudget
	configMu.RUnlock()
	if budget > 0 {
		return budget
	}
	envBudget := os.Getenv("TOKMAN_BUDGET")
	if envBudget != "" {
		var b int
		if _, err := fmt.Sscanf(envBudget, "%d", &b); err == nil {
			return b
		}
	}
	return 0
}

// GetLayerPreset returns the layer preset from flag or environment.
func GetLayerPreset() string {
	configMu.RLock()
	preset := LayerPreset
	configMu.RUnlock()
	if preset != "" {
		return preset
	}
	return os.Getenv("TOKMAN_PRESET")
}

// IsQuietMode returns true if quiet mode is enabled.
func IsQuietMode() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return QuietMode
}

// IsReversibleEnabled returns true if reversible mode is enabled.
func IsReversibleEnabled() bool {
	configMu.RLock()
	enabled := ReversibleEnabled
	configMu.RUnlock()
	return enabled || os.Getenv("TOKMAN_REVERSIBLE") == "true"
}

// IsRemoteMode returns true if remote mode is enabled.
func IsRemoteMode() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return RemoteMode || os.Getenv("TOKMAN_REMOTE") == "true"
}

// GetCompressionAddr returns the compression service address.
func GetCompressionAddr() string {
	configMu.RLock()
	addr := CompressionAddr
	configMu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOKMAN_COMPRESSION_ADDR")
}

// GetAnalyticsAddr returns the analytics service address.
func GetAnalyticsAddr() string {
	configMu.RLock()
	addr := AnalyticsAddr
	configMu.RUnlock()
	if addr != "" {
		return addr
	}
	return os.Getenv("TOKMAN_ANALYTICS_ADDR")
}

// GetRemoteTimeout returns the remote operation timeout in seconds.
func GetRemoteTimeout() int {
	configMu.RLock()
	timeout := RemoteTimeout
	configMu.RUnlock()
	if timeout > 0 {
		return timeout
	}
	return 30
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
}

// SetFlags sets all flag values atomically under a single lock.
func SetFlags(cfg FlagConfig) {
	configMu.Lock()
	Verbose = cfg.Verbose
	DryRun = cfg.DryRun
	UltraCompact = cfg.UltraCompact
	SkipEnv = cfg.SkipEnv
	QueryIntent = cfg.QueryIntent
	LLMEnabled = cfg.LLMEnabled
	TokenBudget = cfg.TokenBudget
	FallbackArgs = cfg.FallbackArgs
	LayerPreset = cfg.LayerPreset
	OutputFile = cfg.OutputFile
	QuietMode = cfg.QuietMode
	JSONOutput = cfg.JSONOutput
	RemoteMode = cfg.RemoteMode
	CompressionAddr = cfg.CompressionAddr
	AnalyticsAddr = cfg.AnalyticsAddr
	RemoteTimeout = cfg.RemoteTimeout
	CompactionEnabled = cfg.CompactionEnabled
	CompactionThreshold = cfg.CompactionThreshold
	CompactionPreserve = cfg.CompactionPreserve
	CompactionMaxTokens = cfg.CompactionMaxTokens
	CompactionSnapshot = cfg.CompactionSnapshot
	CompactionAutoDetect = cfg.CompactionAutoDetect
	ReversibleEnabled = cfg.ReversibleEnabled
	EnableLayers = cfg.EnableLayers
	DisableLayers = cfg.DisableLayers
	StreamMode = cfg.StreamMode
	configMu.Unlock()
}

// GetEnableLayers returns layers to explicitly enable.
func GetEnableLayers() []string {
	configMu.RLock()
	defer configMu.RUnlock()
	return EnableLayers
}

// GetDisableLayers returns layers to explicitly disable.
func GetDisableLayers() []string {
	configMu.RLock()
	defer configMu.RUnlock()
	return DisableLayers
}

// IsStreamMode returns true if streaming mode is enabled for large inputs.
func IsStreamMode() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return StreamMode
}
