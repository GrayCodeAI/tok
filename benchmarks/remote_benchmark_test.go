package benchmarks

import (
	"context"
	"fmt"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/services/compression"
)

// BenchmarkLocalCompression measures in-process compression latency.
func BenchmarkLocalCompression(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:                filter.ModeMinimal,
		Budget:              4000,
		SessionTracking:     true,
		NgramEnabled:        true,
		EnableCompaction:    true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}

	svc := compression.NewService(cfg)
	ctx := context.Background()

	// Sample input: typical CLI output
	input := generateSampleOutput(1000) // 1KB of text

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Compress(ctx, &compression.CompressRequest{
			Input:  input,
			Mode:   filter.ModeMinimal,
			Budget: 4000,
		})
	}
}

// BenchmarkLocalCompressionAggressive measures aggressive mode latency.
func BenchmarkLocalCompressionAggressive(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:                filter.ModeAggressive,
		Budget:              2000,
		SessionTracking:     true,
		NgramEnabled:        true,
		EnableCompaction:    true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}

	svc := compression.NewService(cfg)
	ctx := context.Background()
	input := generateSampleOutput(5000) // 5KB of text

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Compress(ctx, &compression.CompressRequest{
			Input:  input,
			Mode:   filter.ModeAggressive,
			Budget: 2000,
		})
	}
}

// BenchmarkPipelineOverhead measures the overhead of creating new pipelines.
func BenchmarkPipelineOverhead(b *testing.B) {
	input := generateSampleOutput(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := filter.PipelineConfig{
			Mode:            filter.ModeMinimal,
			Budget:          4000,
			SessionTracking: true,
		}
		coordinator := filter.NewPipelineCoordinator(cfg)
		coordinator.Process(input)
	}
}

// BenchmarkLayerByLayer measures individual layer performance.
func BenchmarkLayerByLayer(b *testing.B) {
	input := generateSampleOutput(2000)

	// Test just entropy filtering
	b.Run("EntropyOnly", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:             filter.ModeMinimal,
			EnableEntropy:    true,
			EnableCompaction: false,
			EnableH2O:        false,
		}
		coordinator := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			coordinator.Process(input)
		}
	})

	// Test entropy + ngram
	b.Run("EntropyNgram", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:             filter.ModeMinimal,
			EnableEntropy:    true,
			NgramEnabled:     true,
			EnableCompaction: false,
			EnableH2O:        false,
		}
		coordinator := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			coordinator.Process(input)
		}
	})

	// Test full pipeline
	b.Run("FullPipeline", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:                filter.ModeMinimal,
			Budget:              4000,
			SessionTracking:     true,
			NgramEnabled:        true,
			EnableCompaction:    true,
			EnableH2O:           true,
			EnableAttentionSink: true,
		}
		coordinator := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			coordinator.Process(input)
		}
	})
}

// BenchmarkByInputSize measures latency across different input sizes.
func BenchmarkByInputSize(b *testing.B) {
	sizes := []int{500, 1000, 5000, 10000, 50000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Input%dBytes", size), func(b *testing.B) {
			cfg := filter.PipelineConfig{
				Mode:             filter.ModeMinimal,
				Budget:           4000,
				SessionTracking:  true,
				NgramEnabled:     true,
				EnableCompaction: true,
			}
			svc := compression.NewService(cfg)
			ctx := context.Background()
			input := generateSampleOutput(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = svc.Compress(ctx, &compression.CompressRequest{
					Input: input,
					Mode:  filter.ModeMinimal,
				})
			}
		})
	}
}

// generateSampleOutput creates realistic CLI output for benchmarking.
func generateSampleOutput(size int) string {
	// Simulate typical command output with repeated patterns
	base := `drwxr-xr-x   12 user  staff   384B Mar 28 10:00 .
drwxr-xr-x   45 user  staff  1.4K Mar 28 09:55 ..
-rw-r--r--    1 user  staff   123B Mar 28 10:00 README.md
-rw-r--r--    1 user  staff   4.5K Mar 28 10:00 main.go
-rw-r--r--    1 user  staff   2.3K Mar 28 10:00 config.yaml
DEBUG: Processing file main.go
INFO: Loaded configuration from config.yaml
WARN: Deprecated setting found in config
ERROR: Failed to connect to database
`

	result := ""
	for len(result) < size {
		result += base
	}
	return result[:size]
}
