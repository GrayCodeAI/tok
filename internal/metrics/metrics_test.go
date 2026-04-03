package metrics

import (
	"fmt"
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestRecordCompression(t *testing.T) {
	tests := []struct {
		name             string
		mode             string
		originalTokens   int
		compressedTokens int
		durationMs       float64
	}{
		{"minimal compression", "minimal", 1000, 400, 50.0},
		{"aggressive compression", "aggressive", 5000, 500, 150.0},
		{"no compression needed", "minimal", 100, 100, 5.0},
		{"edge case: compressed > original", "minimal", 100, 150, 10.0},
		{"zero tokens", "minimal", 0, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic for any input
			RecordCompression(tt.mode, tt.originalTokens, tt.compressedTokens, tt.durationMs)
		})
	}
}

func TestRecordCompressionError(t *testing.T) {
	// Should not panic
	RecordCompressionError("minimal")
	RecordCompressionError("aggressive")
}

func TestRecordLayerApplied(t *testing.T) {
	layers := []string{"entropy", "perplexity", "ngram", "h2o", "compaction"}

	for _, layer := range layers {
		RecordLayerApplied(layer)
	}
}

func TestRecordCommand(t *testing.T) {
	tests := []struct {
		command    string
		success    bool
		durationMs float64
	}{
		{"git status", true, 50.0},
		{"npm install", true, 5000.0},
		{"docker build", false, 100.0},
		{"go test", true, 250.0},
	}

	for _, tt := range tests {
		RecordCommand(tt.command, tt.success, tt.durationMs)
	}
}

func TestDiscoveryMetrics(t *testing.T) {
	// Test discovery instance updates
	UpdateDiscoveryInstances("compression", 3, 1)
	UpdateDiscoveryInstances("analytics", 5, 0)

	// Test health checks
	RecordHealthCheck("compression", true)
	RecordHealthCheck("compression", false)
	RecordHealthCheck("analytics", true)
}

func TestLoadBalancerMetrics(t *testing.T) {
	RecordLoadBalancerSelection("compression", "instance-1")
	RecordLoadBalancerSelection("compression", "instance-2")
	RecordLoadBalancerSelection("analytics", "instance-3")
}

func TestGRPCMetrics(t *testing.T) {
	RecordGRPCRequest("CompressionService.Compress", true, 25.5)
	RecordGRPCRequest("CompressionService.Compress", true, 30.2)
	RecordGRPCRequest("AnalyticsService.Record", false, 5.0)
}

func TestCacheMetrics(t *testing.T) {
	RecordCacheHit()
	RecordCacheHit()
	RecordCacheMiss()
	RecordCacheHit()
	UpdateCacheSize(1024)
}

func TestGetCompressionRequestCountSumsSuccessfulModes(t *testing.T) {
	before := GetCompressionRequestCount()

	RecordCompression("minimal", 100, 50, 5)
	RecordCompression("aggressive", 100, 25, 7)
	RecordCompressionError("minimal")

	got := GetCompressionRequestCount() - before
	if got != 2 {
		t.Fatalf("GetCompressionRequestCount delta = %d, want 2", got)
	}
}

func TestGetAllMetricsIncludesGaugeAndCounters(t *testing.T) {
	beforeHits := GetAllMetrics()["cache_hits"]

	RecordCacheHit()
	UpdateCacheSize(2048)

	got := GetAllMetrics()
	if got["cache_hits"] != beforeHits+1 {
		t.Fatalf("cache_hits = %v, want %v", got["cache_hits"], beforeHits+1)
	}
	if got["cache_size"] != 2048 {
		t.Fatalf("cache_size = %v, want 2048", got["cache_size"])
	}
}

func TestReadHistogramByNameAggregatesAcrossModes(t *testing.T) {
	modeA := fmt.Sprintf("hist-a-%s", t.Name())
	modeB := fmt.Sprintf("hist-b-%s", t.Name())
	modeC := fmt.Sprintf("hist-c-%s", t.Name())

	_, beforeCount, beforeBuckets := readHistogramByName("tokman_compression_duration_ms", nil)

	for i := 0; i < 50; i++ {
		RecordCompression(modeA, 100, 50, 4)
		RecordCompression(modeB, 100, 50, 8)
	}
	RecordCompression(modeC, 100, 50, 2048)

	_, afterCount, afterBuckets := readHistogramByName("tokman_compression_duration_ms", nil)
	if delta := afterCount - beforeCount; delta != 101 {
		t.Fatalf("sample count delta = %d, want 101", delta)
	}

	deltaAt8 := bucketDelta(beforeBuckets, afterBuckets, 8)
	if deltaAt8 != 100 {
		t.Fatalf("bucket delta at 8ms = %d, want 100", deltaAt8)
	}

	deltaAt2048 := bucketDelta(beforeBuckets, afterBuckets, 2048)
	if deltaAt2048 != 101 {
		t.Fatalf("bucket delta at 2048ms = %d, want 101", deltaAt2048)
	}
}

func bucketDelta(before, after []*dto.Bucket, upperBound float64) uint64 {
	return bucketCount(after, upperBound) - bucketCount(before, upperBound)
}

func bucketCount(buckets []*dto.Bucket, upperBound float64) uint64 {
	for _, bucket := range buckets {
		if bucket.GetUpperBound() == upperBound {
			return bucket.GetCumulativeCount()
		}
	}
	return 0
}
