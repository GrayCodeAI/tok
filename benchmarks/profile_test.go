package benchmarks

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// BenchmarkProfilePipeline benchmarks the full pipeline with memory tracking.
func BenchmarkProfilePipeline(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:             filter.ModeMinimal,
		Budget:           4000,
		SessionTracking:  true,
		NgramEnabled:     true,
		EnableCompaction: true,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)
	input := generateSampleOutput(10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkLayerProfiling profiles individual layers.
func BenchmarkLayerProfiling(b *testing.B) {
	input := generateSampleOutput(5000)

	b.Run("EntropyLayer", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:          filter.ModeMinimal,
			EnableEntropy: true,
		}
		pipeline := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Process(input)
		}
	})

	b.Run("NgramLayer", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:         filter.ModeMinimal,
			NgramEnabled: true,
		}
		pipeline := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Process(input)
		}
	})

	b.Run("CompactionLayer", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:             filter.ModeMinimal,
			EnableCompaction: true,
		}
		pipeline := filter.NewPipelineCoordinator(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Process(input)
		}
	})
}

// BenchmarkMemoryAllocation tracks memory allocations.
func BenchmarkMemoryAllocation(b *testing.B) {
	input := generateSampleOutput(5000)

	b.Run("NewPipeline", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = filter.NewPipelineCoordinator(filter.PipelineConfig{
				Mode: filter.ModeMinimal,
			})
		}
	})

	b.Run("Process", func(b *testing.B) {
		cfg := filter.PipelineConfig{
			Mode:            filter.ModeMinimal,
			Budget:          4000,
			SessionTracking: true,
		}
		pipeline := filter.NewPipelineCoordinator(cfg)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			pipeline.Process(input)
		}
	})
}
