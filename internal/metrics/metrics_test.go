package metrics

import (
	"testing"
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
