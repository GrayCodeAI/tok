package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds tracing configuration.
type TracingConfig struct {
	Enabled        bool
	ExporterURL    string // e.g., "localhost:4317"
	ServiceName    string
	ServiceVersion string
	Environment    string
	SampleRate     float64
}

// InitTracing initializes OpenTelemetry tracing.
func InitTracing(ctx context.Context, config TracingConfig, logger *slog.Logger) (func(context.Context) error, error) {
	if !config.Enabled {
		logger.Info("tracing disabled")
		return func(context.Context) error { return nil }, nil
	}

	// Create OTLP exporter
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(config.ExporterURL),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SampleRate))),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	logger.Info("tracing initialized",
		"service", config.ServiceName,
		"exporter", config.ExporterURL,
		"sample_rate", config.SampleRate,
	)

	return tp.Shutdown, nil
}

// GetTracer returns a named tracer.
func GetTracer(name string) interface{} {
	return otel.Tracer(name)
}

// SpanAttributes contains common span attributes.
type SpanAttributes struct {
	TeamID           string
	UserID           string
	RequestID        string
	CommandID        string
	FilterName       string
	InputTokens      int
	OutputTokens     int
	CompressionRatio float64
	Duration         int64 // milliseconds
	Status           string
	Error            error
}

// AddAttributesToSpan adds attributes to a span.
func AddAttributesToSpan(ctx context.Context, attrs SpanAttributes) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	var spanAttrs []attribute.KeyValue
	if attrs.TeamID != "" {
		spanAttrs = append(spanAttrs, attribute.Int64("http.request.body_size", int64(attrs.InputTokens)))
	}
	if attrs.UserID != "" {
		spanAttrs = append(spanAttrs, attribute.String("user.id", attrs.UserID))
	}
	if attrs.RequestID != "" {
		spanAttrs = append(spanAttrs, attribute.String("http.request.id", attrs.RequestID))
	}

	if attrs.Status != "" {
		spanAttrs = append(spanAttrs, attribute.Int("http.status_code", 0))
	}

	if attrs.Error != nil {
		span.RecordError(attrs.Error)
		spanAttrs = append(spanAttrs, attribute.String("error", attrs.Error.Error()))
	}

	if len(spanAttrs) > 0 {
		span.SetAttributes(spanAttrs...)
	}
}
