// Package core provides core interfaces and utilities for tok.
//
// The core package defines fundamental abstractions used throughout the codebase,
// including the CommandRunner interface for shell command execution and token
// estimation utilities.
//
// # CommandRunner Interface
//
// The CommandRunner interface abstracts shell command execution, enabling
// dependency injection and testability:
//
//	type CommandRunner interface {
//	    Run(ctx context.Context, args []string) (output string, exitCode int, err error)
//	    LookPath(name string) (string, error)
//	}
//
// # Token Estimation
//
// EstimateTokens provides a fast heuristic for counting tokens in text:
//
//	tokens := core.EstimateTokens("some text to analyze")
package core
