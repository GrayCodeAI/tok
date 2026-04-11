package filter

import (
	"strings"
	"testing"
)

func TestCacheabilityScore(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		minScore int
		maxScore int
	}{
		{"empty", "", 100, 100},
		{"stable system prompt", "You are a helpful assistant.\nFollow these rules:\n1. Be concise\n2. Be accurate", 70, 100},
		{"dynamic log output", "timestamp: 2024-01-01\nfile_path: /tmp/test.log\nline: 42\nerror: something failed", 0, 40},
		{"mixed content", "You are a coding assistant.\nHere is the output:\ntimestamp: 2024-01-01\nresult: ok", 10, 70},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CacheabilityScore(tt.content)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("score %d outside expected range [%d, %d]", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestClassifyContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantStatic bool
	}{
		{"system prompt", "You are a helpful assistant. Follow these instructions.", true},
		{"log output", "timestamp: 2024-01-01\nerror: file not found\nline: 42", false},
		{"instructions", "Instructions:\n1. Read the file\n2. Fix the bug\n3. Run tests", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isStatic, _ := ClassifyContent(tt.content)
			if isStatic != tt.wantStatic {
				t.Errorf("expected static=%v, got %v", tt.wantStatic, isStatic)
			}
		})
	}
}

func TestAlignPrefix(t *testing.T) {
	aligner := NewKVCacheAligner(DefaultKVCacheConfig())

	content := strings.Repeat("You are a helpful assistant.\n", 15) +
		strings.Repeat("timestamp: 2024-01-01\n", 15)

	prefix, suffix, cacheKey := aligner.AlignPrefix(content)
	if prefix == "" {
		t.Error("expected non-empty prefix")
	}
	if suffix == "" {
		t.Error("expected non-empty suffix")
	}
	if cacheKey == "" {
		t.Error("expected non-empty cache key")
	}
}

func TestCacheAwareCompress(t *testing.T) {
	cfg := PipelineConfig{
		Mode:              ModeMinimal,
		SessionTracking:   true,
		EnableCompaction:  true,
		EnableAttribution: true,
	}
	compressor := NewPipelineCoordinator(cfg)
	aligner := NewKVCacheAligner(DefaultKVCacheConfig())

	content := strings.Repeat("You are a helpful assistant.\n", 10) +
		strings.Repeat("This is repeated content that should be compressed.\n", 20)

	result, saved := aligner.CacheAwareCompress(content, compressor)
	if result == "" {
		t.Error("expected non-empty result")
	}
	if saved < 0 {
		t.Errorf("expected non-negative saved tokens, got %d", saved)
	}
}

func TestEstimateCacheHitRate(t *testing.T) {
	tests := []struct {
		content string
		minRate float64
		maxRate float64
	}{
		{"You are a helpful assistant.", 0.5, 1.0},
		{"timestamp: 2024-01-01\nfile_path: /tmp/test\nline: 42\nerror: something", 0.0, 0.5},
	}

	for _, tt := range tests {
		rate := EstimateCacheHitRate(tt.content)
		if rate < tt.minRate || rate > tt.maxRate {
			t.Errorf("rate %.2f outside expected range [%.2f, %.2f]", rate, tt.minRate, tt.maxRate)
		}
	}
}

func TestIsStableLine(t *testing.T) {
	tests := []struct {
		line       string
		wantStable bool
	}{
		{"You are a helpful assistant.", true},
		{"Follow these instructions:", true},
		{"timestamp: 2024-01-01T00:00:00Z", false},
		{"file_path: /tmp/test.log", false},
		{"line: 42", false},
		{"error: something failed", false},
		{"", true},
		{"   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isStableLine(tt.line)
			if got != tt.wantStable {
				t.Errorf("isStableLine(%q) = %v, want %v", tt.line, got, tt.wantStable)
			}
		})
	}
}
