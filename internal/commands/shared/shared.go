package shared

// Package shared provides global state, utilities, and the fallback handler
// for CLI command processing.
//
// # Global State Design
//
// This package maintains global flag and configuration state used across all
// CLI commands. The state is accessed via thread-safe getter/setter functions
// that internally use a sync.RWMutex (see flags.go).
//
// Global state includes:
//   - Command flags (verbose, dry-run, budget, preset, mode, query)
//   - Configuration references (loaded from Viper/TOML)
//   - Runtime context (pipeline coordinator, tracker, session state)
//
// # Thread Safety
//
// Flag accessors (IsVerbose, GetBudget, etc.) are thread-safe via RWMutex.
// Configuration loading should only happen during CLI initialization (main goroutine).
//
// # Deprecation Path
//
// The global state pattern will eventually be replaced with dependency injection
// via Cobra command context (cmd.Context()). Until then, use the thread-safe
// accessor functions instead of direct package variable access:
//
//	GOOD:  if shared.IsVerbose() { ... }
//	BAD:   if shared.Verbose { ... }
//
// # Package Structure
//
//   - flags.go: Global flags and thread-safe accessors
//   - config.go: Configuration loading and caching
//   - executor.go: Command execution and recording
//   - fallback.go: TOML-based fallback handler
//   - utils.go: Utility functions (truncation, sanitization, etc.)
//
// # Dependency Graph
//
//   - flags.go:  NO external tokman dependencies (safe to import anywhere)
//   - utils.go:  NO external tokman dependencies
//   - config.go: depends on internal/config only
//   - executor.go: depends on core, tracking, tee, config
//   - fallback.go: depends on filter, config, toml (largest dependency footprint)

// Backward compatibility: SetConfig alias for SetFlags
//
// Deprecated: Use SetFlags instead. This alias exists only for callers
// that haven't migrated to the new naming convention.
func SetConfig(cfg FlagConfig) {
	SetFlags(cfg)
}
