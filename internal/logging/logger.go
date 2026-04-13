package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog with structured logging
type Logger struct {
	*slog.Logger
}

// New creates a structured logger
func New(level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	return &Logger{slog.New(handler)}
}

// WithContext adds context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return &Logger{l.With("trace_id", traceID)}
	}
	return l
}

// WithError adds error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.With("error", err)}
}

// WithFields adds multiple fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{l.With(args...)}
}

// Command logs command execution
func (l *Logger) Command(cmd string, args []string, duration int64) {
	l.Info("command executed",
		"command", cmd,
		"args", args,
		"duration_ms", duration,
	)
}

// Filter logs filter application
func (l *Logger) Filter(name string, input, output, saved int) {
	l.Debug("filter applied",
		"filter", name,
		"input_tokens", input,
		"output_tokens", output,
		"saved_tokens", saved,
	)
}

// RateLimit logs rate limit events
func (l *Logger) RateLimit(clientIP string, allowed bool) {
	if allowed {
		l.Debug("rate limit check", "client_ip", clientIP, "allowed", true)
	} else {
		l.Warn("rate limit exceeded", "client_ip", clientIP)
	}
}

// Validation logs validation events
func (l *Logger) Validation(field string, value any, valid bool, reason string) {
	if valid {
		l.Debug("validation passed", "field", field, "value", value)
	} else {
		l.Warn("validation failed", "field", field, "value", value, "reason", reason)
	}
}

// Database logs database operations
func (l *Logger) Database(operation string, duration int64, err error) {
	if err != nil {
		l.Error("database operation failed",
			"operation", operation,
			"duration_ms", duration,
			"error", err,
		)
	} else {
		l.Debug("database operation",
			"operation", operation,
			"duration_ms", duration,
		)
	}
}

// Global logger instance
var global *Logger

// Init initializes global logger
func Init(level slog.Level) {
	global = New(level)
}

// Global returns global logger
func Global() *Logger {
	if global == nil {
		global = New(slog.LevelInfo)
	}
	return global
}

// Helper functions for global logger
func Info(msg string, args ...any)  { Global().Info(msg, args...) }
func Debug(msg string, args ...any) { Global().Debug(msg, args...) }
func Warn(msg string, args ...any)  { Global().Warn(msg, args...) }
func Error(msg string, args ...any) { Global().Error(msg, args...) }
