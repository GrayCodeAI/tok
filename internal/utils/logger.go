package utils

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// loggerMu protects global logger state
var loggerMu sync.RWMutex

// Logger is the global logger instance.
var Logger *slog.Logger

// logFile stores the file handle for cleanup
var logFile *os.File

// LogLevel represents logging severity.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// InitLogger initializes the global logger.
func InitLogger(logPath string, level LogLevel) error {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return err
	}

	// Close any previously opened log file to prevent descriptor leak.
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	logFile = file

	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	Logger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slogLevel,
	}))

	return nil
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	loggerMu.RLock()
	logger := Logger
	loggerMu.RUnlock()

	if logger == nil {
		return
	}
	logger.Warn(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	loggerMu.RLock()
	logger := Logger
	loggerMu.RUnlock()

	if logger == nil {
		return
	}
	logger.Info(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	loggerMu.RLock()
	logger := Logger
	loggerMu.RUnlock()

	if logger == nil {
		return
	}
	logger.Debug(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	loggerMu.RLock()
	logger := Logger
	loggerMu.RUnlock()

	if logger == nil {
		return
	}
	logger.Error(msg, args...)
}

// CloseLogger closes the log file and resets the logger.
func CloseLogger() error {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		Logger = nil
		return err
	}
	return nil
}
