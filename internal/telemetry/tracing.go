// Package telemetry provides tracing infrastructure for TokMan.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides distributed tracing for TokMan
type Tracer struct {
	serviceName    string
	serviceVersion string
}

// Config holds telemetry configuration
type Config struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
}

// NewConfig creates a default telemetry configuration
func NewConfig() *Config {
	return &Config{
		Enabled:        false,
		ServiceName:    "tokman",
		ServiceVersion: "dev",
	}
}

// NewTracer creates a new telemetry tracer
func NewTracer(ctx context.Context, cfg *Config) *Tracer {
	// Set text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		serviceName:    cfg.ServiceName,
		serviceVersion: cfg.ServiceVersion,
	}
}

// StartSpan starts a new tracing span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer(t.serviceName)
	return tracer.Start(ctx, name)
}

// EndSpan ends a tracing span
func (t *Tracer) EndSpan(span trace.Span, err error) {
	if span == nil {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}

// AddSpanAttributes adds attributes to a span
func (t *Tracer) AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}

	span.SetAttributes(attrs...)
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	return nil
}

// GetTraceID returns the trace ID from a span
func GetTraceID(span trace.Span) string {
	if span == nil {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
