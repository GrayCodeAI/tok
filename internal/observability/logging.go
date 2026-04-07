package observability

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// LogConfig holds logging configuration.
type LogConfig struct {
	Level     string // "debug", "info", "warn", "error"
	Format    string // "json", "text"
	Output    string // "stdout", "stderr", "file"
	FilePath  string // if Output == "file"
	AddSource bool
}

// InitLogger initializes the global logger.
func InitLogger(config LogConfig) *slog.Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Select output
	var writer io.Writer
	switch config.Output {
	case "stderr":
		writer = os.Stderr
	case "file":
		f, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			writer = os.Stdout
		} else {
			writer = f
		}
	default:
		writer = os.Stdout
	}

	// Create handler
	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     level,
			AddSource: config.AddSource,
		})
	default:
		handler = tint.NewHandler(writer, &tint.Options{
			Level:     level,
			AddSource: config.AddSource,
			NoColor:   config.Output == "file",
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// LogSpanInfo logs structured information about a span.
func LogSpanInfo(logger *slog.Logger, spanName string, attrs map[string]interface{}) {
	var groupArgs []any
	for k, v := range attrs {
		groupArgs = append(groupArgs, slog.Any(k, v))
	}
	logger.Debug("span executed",
		slog.String("span", spanName),
		slog.Group("attributes", groupArgs...),
	)
}

// LogError logs an error with context.
func LogError(logger *slog.Logger, err error, msg string, attrs ...any) {
	logger.Error(msg, append([]any{slog.String("error", err.Error())}, attrs...)...)
}

// LogMetric logs a metric value.
func LogMetric(logger *slog.Logger, metricName string, value interface{}, attrs ...any) {
	logger.Debug("metric recorded",
		append([]any{
			slog.String("metric", metricName),
			slog.Any("value", value),
		}, attrs...)...,
	)
}

// convertToSlogAttrs converts a map to slog attributes.
func convertToSlogAttrs(m map[string]interface{}) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(m))
	for k, v := range m {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}
