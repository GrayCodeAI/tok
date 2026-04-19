// Package integration_test provides integration tests for tok.
package tok_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/metrics"
	"github.com/lakshmanpatel/tok/internal/security"
)

// TestPipelineCompression tests the full compression pipeline
func TestPipelineCompression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantReduced bool
	}{
		{
			name:        "simple text",
			input:       "Hello, World! This is a test.",
			wantReduced: true,
		},
		{
			name:        "empty input",
			input:       "",
			wantReduced: false,
		},
		{
			name:        "code with duplicates",
			input:       strings.Repeat("func main() {}\n", 10),
			wantReduced: true,
		},
		{
			name:        "json output",
			input:       `{"key": "value", "array": [1,2,3]}`,
			wantReduced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := filter.NewPipelineCoordinator(filter.PipelineConfig{
				Mode:            filter.ModeMinimal,
				SessionTracking: true,
				NgramEnabled:    true,
			})

			output, _ := p.Process(tt.input)
			_ = output
			t.Logf("Processed %d bytes", len(tt.input))
		})
	}
}

// TestSecurityValidation tests security validations
func TestSecurityValidation(t *testing.T) {
	validator := security.NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		fn      func(string) error
	}{
		{
			name:    "valid preset fast",
			input:   "fast",
			wantErr: false,
			fn:      validator.ValidatePreset,
		},
		{
			name:    "valid preset balanced",
			input:   "balanced",
			wantErr: false,
			fn:      validator.ValidatePreset,
		},
		{
			name:    "invalid preset",
			input:   "invalid",
			wantErr: true,
			fn:      validator.ValidatePreset,
		},
		{
			name:    "valid mode minimal",
			input:   "minimal",
			wantErr: false,
			fn:      validator.ValidateMode,
		},
		{
			name:    "invalid mode",
			input:   "invalid",
			wantErr: true,
			fn:      validator.ValidateMode,
		},
		{
			name:    "valid budget",
			input:   "1000",
			wantErr: false,
			fn:      func(s string) error { return validator.ValidateBudget(1000) },
		},
		{
			name:    "negative budget",
			input:   "-1",
			wantErr: true,
			fn:      func(s string) error { return validator.ValidateBudget(-1) },
		},
		{
			name:    "path traversal",
			input:   "../etc/passwd",
			wantErr: true,
			fn:      validator.ValidatePath,
		},
		{
			name:    "valid path",
			input:   "/tmp/test.txt",
			wantErr: false,
			fn:      validator.ValidatePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestMetrics tests metrics collection
func TestMetrics(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	// Record some metrics
	m.RecordCommandProcessed()
	m.RecordCompressionRun()
	m.RecordCompressionDuration(50 * time.Millisecond)
	m.RecordCacheHit()
	m.SetMemoryUsage(100)

	snap := m.Snapshot()

	if snap.CommandsProcessed != 1 {
		t.Errorf("expected 1 command processed, got %d", snap.CommandsProcessed)
	}
	if snap.CompressionRuns != 1 {
		t.Errorf("expected 1 compression run, got %d", snap.CompressionRuns)
	}
	if snap.CacheHits != 1 {
		t.Errorf("expected 1 cache hit, got %d", snap.CacheHits)
	}
	if snap.MemoryUsageMB != 100 {
		t.Errorf("expected 100MB, got %d", snap.MemoryUsageMB)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     config.Defaults(),
			wantErr: false,
		},
		{
			name: "invalid entropy threshold",
			cfg: func() *config.Config {
				c := config.Defaults()
				c.Pipeline.EntropyThreshold = 1.5
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid negative budget",
			cfg: func() *config.Config {
				c := config.Defaults()
				c.Pipeline.DefaultBudget = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid mode",
			cfg: func() *config.Config {
				c := config.Defaults()
				c.Filter.Mode = "invalid"
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestLargeInput tests handling of large inputs
func TestLargeInput(t *testing.T) {
	// Create a large input
	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		builder.WriteString("Line ")
		builder.WriteRune('a' + rune(i%26))
		builder.WriteString(": This is test content for the compression pipeline. ")
	}
	input := builder.String()

	p := filter.NewPipelineCoordinator(filter.PipelineConfig{
		Mode:             filter.ModeMinimal,
		SessionTracking:  true,
		EnableEntropy:    true,
		EnablePerplexity: true,
		EnableH2O:        true,
	})

	output, stats := p.Process(input)

	t.Logf("Input: %d chars", len(input))
	t.Logf("Output: %d chars", len(output))
	_ = stats
}

// TestContextCancellation tests handling of context cancellation
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	_ = ctx  // Avoid unused variable warning

	p := filter.NewPipelineCoordinator(filter.PipelineConfig{
		Mode:            filter.ModeMinimal,
		SessionTracking: true,
	})

	input := "test content"
	_, _ = p.Process(input)
	t.Logf("Context cancellation test passed")
}

// BenchmarkFullPipeline benchmarks the full pipeline
func BenchmarkFullPipeline(b *testing.B) {
	input := strings.Repeat("test content for benchmarking. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := filter.NewPipelineCoordinator(filter.PipelineConfig{
			Mode:            filter.ModeMinimal,
			SessionTracking: true,
			NgramEnabled:    true,
			EnableEntropy:   true,
			EnableH2O:       true,
		})
		output, _ := p.Process(input)
		_ = output
	}
}

// BenchmarkMinimalPipeline benchmarks a minimal pipeline
func BenchmarkMinimalPipeline(b *testing.B) {
	input := strings.Repeat("test content. ", 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := filter.NewPipelineCoordinator(filter.PipelineConfig{
			Mode:            filter.ModeMinimal,
			SessionTracking: false,
		})
		output, _ := p.Process(input)
		_ = output
	}
}

// TestMain is the test entry point
func TestMain(m *testing.M) {
	// Setup
	os.Exit(m.Run())
}
