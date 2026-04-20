package logging

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	logger := New(slog.LevelDebug)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if logger.Logger == nil {
		t.Error("expected non-nil internal logger")
	}
}

func TestLogger_WithContext(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Test without trace ID
	ctx := context.Background()
	newLogger := logger.WithContext(ctx)
	if newLogger == nil {
		t.Error("expected non-nil logger")
	}

	// Test with trace ID (typed key per SA1029)
	type ctxKey string
	ctx = context.WithValue(context.Background(), ctxKey("trace_id"), "abc-123")
	newLogger = logger.WithContext(ctx)
	if newLogger == nil {
		t.Error("expected non-nil logger with trace_id")
	}
}

func TestLogger_WithError(t *testing.T) {
	logger := New(slog.LevelDebug)
	testErr := errors.New("test error")

	newLogger := logger.WithError(testErr)
	if newLogger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger := New(slog.LevelDebug)
	fields := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)
	if newLogger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestLogger_Command(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Should not panic
	logger.Command("git", []string{"status"}, 100)
	logger.Command("docker", []string{"ps", "-a"}, 250)
	logger.Command("ls", []string{}, 10)
}

func TestLogger_Filter(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Should not panic
	logger.Filter("entropy", 1000, 800, 200)
	logger.Filter("h2o", 5000, 3000, 2000)
}

func TestLogger_RateLimit(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Should not panic
	logger.RateLimit("192.168.1.1", true)
	logger.RateLimit("10.0.0.1", false)
}

func TestLogger_Validation(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Should not panic - valid
	logger.Validation("email", "test@example.com", true, "")

	// Should not panic - invalid
	logger.Validation("email", "invalid", false, "missing @ symbol")
}

func TestLogger_Database(t *testing.T) {
	logger := New(slog.LevelDebug)

	// Success case
	logger.Database("SELECT", 50, nil)

	// Error case
	testErr := errors.New("connection timeout")
	logger.Database("INSERT", 100, testErr)
}

func TestInit(t *testing.T) {
	// Reset global
	global = nil

	Init(slog.LevelWarn)

	if global == nil {
		t.Error("expected global to be initialized")
	}
}

func TestGlobal(t *testing.T) {
	// Reset global
	global = nil

	g := Global()
	if g == nil {
		t.Fatal("expected non-nil global logger")
	}

	// Second call should return same instance
	g2 := Global()
	if g != g2 {
		t.Error("expected same global instance")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Ensure global is initialized
	global = nil
	Init(slog.LevelDebug)

	// These should not panic
	Info("info message", "key", "value")
	Debug("debug message")
	Warn("warning message", "count", 5)
	Error("error message", "err", "something failed")
}

func TestLogger_Levels(t *testing.T) {
	// Test different log levels
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		logger := New(level)
		if logger == nil {
			t.Errorf("failed to create logger with level %v", level)
		}
	}
}
