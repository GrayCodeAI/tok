// Package metrics provides Prometheus metrics for TokMan services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TokMan metrics exposed via Prometheus.
var (
	// Compression metrics
	CompressionRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_compression_requests_total",
			Help: "Total number of compression requests",
		},
		[]string{"mode", "status"},
	)

	CompressionTokensOriginal = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tokman_compression_tokens_original",
			Help:    "Original token count before compression",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10), // 100 to 51200
		},
	)

	CompressionTokensCompressed = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tokman_compression_tokens_compressed",
			Help:    "Token count after compression",
			Buckets: prometheus.ExponentialBuckets(50, 2, 10),
		},
	)

	CompressionSavingsPercent = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tokman_compression_savings_percent",
			Help:    "Percentage of tokens saved",
			Buckets: prometheus.LinearBuckets(0, 10, 11), // 0% to 100%
		},
	)

	CompressionDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tokman_compression_duration_ms",
			Help:    "Compression processing time in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1ms to ~4s
		},
		[]string{"mode"},
	)

	LayersApplied = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_layers_applied_total",
			Help: "Number of times each layer was applied",
		},
		[]string{"layer"},
	)

	// Command execution metrics
	CommandsExecuted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_commands_executed_total",
			Help: "Total commands executed",
		},
		[]string{"command", "status"},
	)

	CommandDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tokman_command_duration_ms",
			Help:    "Command execution duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(10, 2, 12),
		},
		[]string{"command"},
	)

	// Service discovery metrics
	DiscoveryInstances = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tokman_discovery_instances",
			Help: "Number of discovered service instances",
		},
		[]string{"service_type", "healthy"},
	)

	DiscoveryHealthChecks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_discovery_health_checks_total",
			Help: "Total health checks performed",
		},
		[]string{"service_type", "status"},
	)

	// Load balancer metrics
	LoadBalancerSelections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_loadbalancer_selections_total",
			Help: "Number of instance selections by load balancer",
		},
		[]string{"service_type", "instance_id"},
	)

	// gRPC metrics
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokman_grpc_requests_total",
			Help: "Total gRPC requests",
		},
		[]string{"method", "status"},
	)

	GRPCDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tokman_grpc_duration_ms",
			Help:    "gRPC request duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"method"},
	)

	// Cache metrics
	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tokman_cache_hits_total",
			Help: "Total cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tokman_cache_misses_total",
			Help: "Total cache misses",
		},
	)

	CacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tokman_cache_size",
			Help: "Current cache size",
		},
	)

	// Token tracking metrics
	TokensSavedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tokman_tokens_saved_total",
			Help: "Total tokens saved across all operations",
		},
	)

	TokensProcessedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tokman_tokens_processed_total",
			Help: "Total tokens processed",
		},
	)
)

// RecordCompression records compression metrics.
func RecordCompression(mode string, originalTokens, compressedTokens int, durationMs float64) {
	savingsPercent := 0.0
	tokensSaved := 0

	if originalTokens > 0 {
		tokensSaved = originalTokens - compressedTokens
		if tokensSaved < 0 {
			tokensSaved = 0 // Ensure non-negative for metrics
		}
		savingsPercent = float64(tokensSaved) / float64(originalTokens) * 100
		if savingsPercent < 0 {
			savingsPercent = 0
		}
	}

	CompressionRequestsTotal.WithLabelValues(mode, "success").Inc()
	CompressionTokensOriginal.Observe(float64(originalTokens))
	CompressionTokensCompressed.Observe(float64(compressedTokens))
	CompressionSavingsPercent.Observe(savingsPercent)
	CompressionDurationMs.WithLabelValues(mode).Observe(durationMs)

	// Only add positive values to counter (Prometheus counters cannot decrease)
	if tokensSaved > 0 {
		TokensSavedTotal.Add(float64(tokensSaved))
	}
	TokensProcessedTotal.Add(float64(originalTokens))
}

// RecordCompressionError records a failed compression.
func RecordCompressionError(mode string) {
	CompressionRequestsTotal.WithLabelValues(mode, "error").Inc()
}

// RecordLayerApplied records that a layer was applied.
func RecordLayerApplied(layerName string) {
	LayersApplied.WithLabelValues(layerName).Inc()
}

// RecordCommand records command execution metrics.
func RecordCommand(command string, success bool, durationMs float64) {
	status := "success"
	if !success {
		status = "error"
	}
	CommandsExecuted.WithLabelValues(command, status).Inc()
	CommandDurationMs.WithLabelValues(command).Observe(durationMs)
}

// UpdateDiscoveryInstances updates discovery instance counts.
func UpdateDiscoveryInstances(serviceType string, healthy, unhealthy int) {
	DiscoveryInstances.WithLabelValues(serviceType, "true").Set(float64(healthy))
	DiscoveryInstances.WithLabelValues(serviceType, "false").Set(float64(unhealthy))
}

// RecordHealthCheck records a health check result.
func RecordHealthCheck(serviceType string, healthy bool) {
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}
	DiscoveryHealthChecks.WithLabelValues(serviceType, status).Inc()
}

// RecordLoadBalancerSelection records a load balancer selection.
func RecordLoadBalancerSelection(serviceType, instanceID string) {
	LoadBalancerSelections.WithLabelValues(serviceType, instanceID).Inc()
}

// RecordGRPCRequest records a gRPC request.
func RecordGRPCRequest(method string, success bool, durationMs float64) {
	status := "success"
	if !success {
		status = "error"
	}
	GRPCRequestsTotal.WithLabelValues(method, status).Inc()
	GRPCDurationMs.WithLabelValues(method).Observe(durationMs)
}

// RecordCacheHit records a cache hit.
func RecordCacheHit() {
	CacheHits.Inc()
}

// RecordCacheMiss records a cache miss.
func RecordCacheMiss() {
	CacheMisses.Inc()
}

// UpdateCacheSize updates the cache size gauge.
func UpdateCacheSize(size int) {
	CacheSize.Set(float64(size))
}
