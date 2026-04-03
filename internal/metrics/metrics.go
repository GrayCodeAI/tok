// Package metrics provides Prometheus metrics for TokMan services.
package metrics

import (
	"math"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
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

// readMetricByName reads the sum of matching counters or gauges from the default registry.
func readMetricByName(name string, labelFilters map[string]string) float64 {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0
	}
	var value float64
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if !metricLabelsMatch(m, labelFilters) {
				continue
			}
			if c := m.GetCounter(); c != nil {
				value += c.GetValue()
				continue
			}
			if g := m.GetGauge(); g != nil {
				value += g.GetValue()
			}
		}
	}
	return value
}

func metricLabelsMatch(metric *dto.Metric, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	labels := metric.GetLabel()
	for name, want := range filters {
		found := false
		for _, label := range labels {
			if label.GetName() == name && label.GetValue() == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// readHistogramByName reads aggregated histogram sum, count, and cumulative buckets.
func readHistogramByName(name string, labelFilters map[string]string) (sum float64, count uint64, buckets []*dto.Bucket) {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0, 0, nil
	}
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		aggregated := make(map[float64]uint64)
		for _, m := range mf.GetMetric() {
			if !metricLabelsMatch(m, labelFilters) {
				continue
			}
			if h := m.GetHistogram(); h != nil {
				sum += h.GetSampleSum()
				count += h.GetSampleCount()
				for _, bucket := range h.GetBucket() {
					aggregated[bucket.GetUpperBound()] += bucket.GetCumulativeCount()
				}
			}
		}
		if len(aggregated) == 0 {
			return 0, 0, nil
		}
		bounds := make([]float64, 0, len(aggregated))
		for upperBound := range aggregated {
			bounds = append(bounds, upperBound)
		}
		sort.Float64s(bounds)
		buckets = make([]*dto.Bucket, 0, len(bounds))
		for _, upperBound := range bounds {
			upperBoundCopy := upperBound
			countCopy := aggregated[upperBound]
			buckets = append(buckets, &dto.Bucket{
				CumulativeCount: &countCopy,
				UpperBound:      &upperBoundCopy,
			})
		}
		return sum, count, buckets
	}
	return 0, 0, nil
}

// GetCompressionRequestCount returns total successful compression requests.
func GetCompressionRequestCount() int64 {
	return int64(readMetricByName("tokman_compression_requests_total", map[string]string{"status": "success"}))
}

// GetTokensSavedTotal returns cumulative tokens saved.
func GetTokensSavedTotal() float64 {
	return readMetricByName("tokman_tokens_saved_total", nil)
}

// GetTokensProcessedTotal returns cumulative tokens processed.
func GetTokensProcessedTotal() float64 {
	return readMetricByName("tokman_tokens_processed_total", nil)
}

// GetP99LatencyMs computes the approximate p99 compression duration by
// reading histogram buckets from the default registry.
func GetP99LatencyMs() float64 {
	const percentile = 0.99
	_, totalCount, buckets := readHistogramByName("tokman_compression_duration_ms", nil)
	if totalCount == 0 || len(buckets) == 0 {
		return 0
	}
	target := uint64(math.Ceil(float64(totalCount) * percentile))
	if target == 0 {
		target = 1
	}
	for _, bucket := range buckets {
		if bucket.GetCumulativeCount() >= target {
			return bucket.GetUpperBound()
		}
	}
	return buckets[len(buckets)-1].GetUpperBound()
}

// GetAllMetrics returns a snapshot of all counter and gauge values for HTTP export.
func GetAllMetrics() map[string]float64 {
	return map[string]float64{
		"tokens_saved_total":     GetTokensSavedTotal(),
		"tokens_processed_total": GetTokensProcessedTotal(),
		"compression_requests":   float64(GetCompressionRequestCount()),
		"cache_hits":             readMetricByName("tokman_cache_hits_total", nil),
		"cache_misses":           readMetricByName("tokman_cache_misses_total", nil),
		"cache_size":             readMetricByName("tokman_cache_size", nil),
		"p99_latency_ms":         GetP99LatencyMs(),
	}
}

// HasFilter returns true if any layer matching the prefix has been applied.
func HasFilter(prefix string) bool {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return false
	}
	for _, mf := range mfs {
		if mf.GetName() == "tokman_layers_applied_total" {
			for _, m := range mf.GetMetric() {
				for _, label := range m.GetLabel() {
					if label.GetName() == "layer" && strings.HasPrefix(label.GetValue(), prefix) {
						return true
					}
				}
			}
		}
	}
	return false
}
