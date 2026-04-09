// Package errors provides domain-specific errors for TokMan.
// All errors should be defined here for consistency and easier error handling.
package errors

import (
	"errors"
	"fmt"
)

// Domain errors for TokMan
var (
	// Configuration errors
	ErrConfigInvalid     = errors.New("configuration is invalid")
	ErrConfigNotFound    = errors.New("configuration file not found")
	ErrConfigParseFailed = errors.New("failed to parse configuration")
	ErrConfigValidation  = errors.New("configuration validation failed")

	// Command execution errors
	ErrCommandNotFound  = errors.New("command not found")
	ErrCommandExecution = errors.New("command execution failed")
	ErrCommandTimeout   = errors.New("command execution timed out")
	ErrCommandInvalid   = errors.New("invalid command")
	ErrShellMetaChars   = errors.New("command contains shell meta-characters")

	// Compression/Filter errors
	ErrCompressionFailed = errors.New("compression failed")
	ErrFilterNotFound    = errors.New("filter not found")
	ErrFilterExecution   = errors.New("filter execution failed")
	ErrBudgetExceeded    = errors.New("token budget exceeded")
	ErrInvalidMode       = errors.New("invalid compression mode")
	ErrInvalidPreset     = errors.New("invalid pipeline preset")

	// Database/Storage errors
	ErrDatabaseOpen      = errors.New("failed to open database")
	ErrDatabaseQuery     = errors.New("database query failed")
	ErrDatabaseMigration = errors.New("database migration failed")
	ErrRecordNotFound    = errors.New("record not found")

	// Pipeline errors
	ErrPipelineInit      = errors.New("pipeline initialization failed")
	ErrPipelineExecution = errors.New("pipeline execution failed")
	ErrLayerNotFound     = errors.New("pipeline layer not found")
	ErrLayerDisabled     = errors.New("pipeline layer is disabled")

	// LLM/AI errors
	ErrLLMNotAvailable    = errors.New("LLM service not available")
	ErrLLMRequestFailed   = errors.New("LLM request failed")
	ErrLLMTimeout         = errors.New("LLM request timed out")
	ErrLLMInvalidResponse = errors.New("invalid LLM response")

	// Hook/Integration errors
	ErrHookInstallFailed = errors.New("failed to install hooks")
	ErrHookIntegrity     = errors.New("hook integrity check failed")
	ErrAgentNotFound     = errors.New("agent not found")
	ErrAgentInitFailed   = errors.New("agent initialization failed")

	// General errors
	ErrInvalidInput   = errors.New("invalid input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrNotImplemented = errors.New("not implemented")
	ErrInternal       = errors.New("internal error")
)

// ErrorWithContext wraps an error with additional context.
type ErrorWithContext struct {
	Err     error
	Context string
	Op      string
}

func (e *ErrorWithContext) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Context, e.Err)
}

func (e *ErrorWithContext) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with context.
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return &ErrorWithContext{Err: err, Context: context}
}

// Wrapf wraps an error with formatted context.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &ErrorWithContext{
		Err:     err,
		Context: fmt.Sprintf(format, args...),
	}
}

// New creates a new error with formatted message.
func New(message string) error {
	return errors.New(message)
}

// Errorf creates a new formatted error.
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Join returns an error that wraps the given errors.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// IsConfigError returns true if the error is a configuration error.
func IsConfigError(err error) bool {
	return Is(err, ErrConfigInvalid) ||
		Is(err, ErrConfigNotFound) ||
		Is(err, ErrConfigParseFailed) ||
		Is(err, ErrConfigValidation)
}

// IsCommandError returns true if the error is a command execution error.
func IsCommandError(err error) bool {
	return Is(err, ErrCommandNotFound) ||
		Is(err, ErrCommandExecution) ||
		Is(err, ErrCommandTimeout) ||
		Is(err, ErrCommandInvalid)
}

// IsCompressionError returns true if the error is a compression error.
func IsCompressionError(err error) bool {
	return Is(err, ErrCompressionFailed) ||
		Is(err, ErrFilterNotFound) ||
		Is(err, ErrBudgetExceeded) ||
		Is(err, ErrInvalidMode)
}

// IsRetryable returns true if the error is transient and can be retried.
func IsRetryable(err error) bool {
	// Add logic to determine if an error is retryable
	if Is(err, ErrCommandTimeout) ||
		Is(err, ErrLLMTimeout) ||
		Is(err, ErrLLMRequestFailed) {
		return true
	}
	return false
}

// ExitCode returns the appropriate exit code for an error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if Is(err, ErrCommandNotFound) {
		return 127
	}
	if Is(err, ErrCommandInvalid) || Is(err, ErrShellMetaChars) {
		return 126
	}
	if IsConfigError(err) {
		return 78 // EX_CONFIG
	}

	return 1
}
