package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Enabled bool
	Port    int
}

// Metrics holds all application metrics.
type Metrics struct {
	// Compression metrics
	TokensProcessed      metric.Int64Counter
	TokensSaved          metric.Int64Counter
	CompressionRatio     metric.Float64Histogram
	FilterActivations    metric.Int64Counter
	FilterLatency        metric.Float64Histogram

	// API metrics
	RequestsTotal        metric.Int64Counter
	RequestDuration      metric.Float64Histogram
	ErrorsTotal          metric.Int64Counter
	ActiveConnections    metric.Int64UpDownCounter

	// Team metrics
	TeamsActive          metric.Int64Gauge
	UsersActive          metric.Int64Gauge
	SessionsActive       metric.Int64UpDownCounter

	// Cost metrics
	EstimatedCostSaved   metric.Float64Counter
	TokenBudgetUsed      metric.Float64Gauge
	OverageTokens        metric.Int64Counter

	// Cache metrics
	CacheHits            metric.Int64Counter
	CacheMisses          metric.Int64Counter
	CacheSize            metric.Int64Gauge

	// Database metrics
	QueryLatency         metric.Float64Histogram
	ConnectionPoolSize   metric.Int64Gauge
}

// InitMetrics initializes Prometheus metrics.
func InitMetrics(ctx context.Context, config MetricsConfig, logger *slog.Logger) (*Metrics, error) {
	if !config.Enabled {
		logger.Info("metrics disabled")
		return &Metrics{}, nil
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("creating prometheus exporter: %w", err)
	}

	// Create metric provider
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(mp)

	// Create meter
	meter := mp.Meter("github.com/GrayCodeAI/tokman")

	// Create metrics
	metrics := &Metrics{}

	// Compression metrics
	metrics.TokensProcessed, err = meter.Int64Counter("tokman_tokens_processed_total",
		metric.WithDescription("Total tokens processed"))
	if err != nil {
		return nil, err
	}

	metrics.TokensSaved, err = meter.Int64Counter("tokman_tokens_saved_total",
		metric.WithDescription("Total tokens saved through compression"))
	if err != nil {
		return nil, err
	}

	metrics.CompressionRatio, err = meter.Float64Histogram("tokman_compression_ratio",
		metric.WithDescription("Compression ratio distribution"))
	if err != nil {
		return nil, err
	}

	metrics.FilterActivations, err = meter.Int64Counter("tokman_filter_activations_total",
		metric.WithDescription("Total filter activations by type"))
	if err != nil {
		return nil, err
	}

	metrics.FilterLatency, err = meter.Float64Histogram("tokman_filter_latency_ms",
		metric.WithDescription("Filter execution latency in milliseconds"))
	if err != nil {
		return nil, err
	}

	// API metrics
	metrics.RequestsTotal, err = meter.Int64Counter("tokman_requests_total",
		metric.WithDescription("Total API requests"))
	if err != nil {
		return nil, err
	}

	metrics.RequestDuration, err = meter.Float64Histogram("tokman_request_duration_ms",
		metric.WithDescription("Request duration in milliseconds"))
	if err != nil {
		return nil, err
	}

	metrics.ErrorsTotal, err = meter.Int64Counter("tokman_errors_total",
		metric.WithDescription("Total API errors"))
	if err != nil {
		return nil, err
	}

	metrics.ActiveConnections, err = meter.Int64UpDownCounter("tokman_active_connections",
		metric.WithDescription("Active concurrent connections"))
	if err != nil {
		return nil, err
	}

	// Team metrics
	metrics.TeamsActive, err = meter.Int64Gauge("tokman_teams_active",
		metric.WithDescription("Number of active teams"))
	if err != nil {
		return nil, err
	}

	metrics.UsersActive, err = meter.Int64Gauge("tokman_users_active",
		metric.WithDescription("Number of active users"))
	if err != nil {
		return nil, err
	}

	metrics.SessionsActive, err = meter.Int64UpDownCounter("tokman_sessions_active",
		metric.WithDescription("Active user sessions"))
	if err != nil {
		return nil, err
	}

	// Cost metrics
	metrics.EstimatedCostSaved, err = meter.Float64Counter("tokman_estimated_cost_saved_usd",
		metric.WithDescription("Estimated cost saved in USD"))
	if err != nil {
		return nil, err
	}

	metrics.TokenBudgetUsed, err = meter.Float64Gauge("tokman_token_budget_used_percent",
		metric.WithDescription("Token budget usage percentage"))
	if err != nil {
		return nil, err
	}

	metrics.OverageTokens, err = meter.Int64Counter("tokman_overage_tokens_total",
		metric.WithDescription("Total tokens over budget"))
	if err != nil {
		return nil, err
	}

	// Cache metrics
	metrics.CacheHits, err = meter.Int64Counter("tokman_cache_hits_total",
		metric.WithDescription("Total cache hits"))
	if err != nil {
		return nil, err
	}

	metrics.CacheMisses, err = meter.Int64Counter("tokman_cache_misses_total",
		metric.WithDescription("Total cache misses"))
	if err != nil {
		return nil, err
	}

	metrics.CacheSize, err = meter.Int64Gauge("tokman_cache_size_bytes",
		metric.WithDescription("Cache size in bytes"))
	if err != nil {
		return nil, err
	}

	// Database metrics
	metrics.QueryLatency, err = meter.Float64Histogram("tokman_query_latency_ms",
		metric.WithDescription("Database query latency in milliseconds"))
	if err != nil {
		return nil, err
	}

	metrics.ConnectionPoolSize, err = meter.Int64Gauge("tokman_db_connection_pool_size",
		metric.WithDescription("Database connection pool size"))
	if err != nil {
		return nil, err
	}

	logger.Info("metrics initialized",
		"port", config.Port,
	)

	return metrics, nil
}
