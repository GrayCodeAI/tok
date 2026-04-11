// Package shared provides shared state and configuration for TokMan CLI commands.
//
// The shared package centralizes CLI flag state management through the AppState
// struct, replacing global variables with a testable, concurrent design.
//
// # State Management
//
// AppState encapsulates all CLI flag state with proper mutex protection:
//
//	state := shared.Global()
//	if state.IsVerbose() { ... }
//
// # Flag Configuration
//
// Use SetFlags to atomically update all flag values:
//
//	shared.SetFlags(shared.FlagConfig{
//	    Verbose: 2,
//	    DryRun: false,
//	    TokenBudget: 2000,
//	})
//
// # Backward Compatibility
//
// Package-level accessor functions (e.g., shared.IsVerbose()) are provided
// for backward compatibility with existing code. New code should pass
// AppState explicitly where possible.
package shared
