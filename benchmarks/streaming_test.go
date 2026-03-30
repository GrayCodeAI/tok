package benchmarks

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// Benchmark_Streaming_LargeInput tests streaming mode on very large inputs (>500K tokens)
func Benchmark_Streaming_LargeInput(b *testing.B) {
	// Generate 600K token input (2.4M characters)
	var builder strings.Builder
	for i := 0; i < 60000; i++ {
		builder.WriteString("2026-03-30 10:30:45 INFO  [main] Processing request user_id=12345 latency=45ms\n")
	}
	input := builder.String()

	cfg := filter.LayerConfigs{
		EnableEntropy:    true,
		EnableTFIDF:      true,
		EnableH2O:        true,
		EnableCompaction: true,
	}
	processor := filter.NewStreamingProcessor(filter.ModeAggressive, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := processor.ProcessStream(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		b.ReportMetric(float64(stats.ReductionPercent), "reduction_pct")
		_ = output
	}
}

// Benchmark_Streaming_WithOverlap tests streaming with overlap handling
func Benchmark_Streaming_WithOverlap(b *testing.B) {
	// Generate content with context that needs overlap
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString("=== Section " + string(rune('A'+i%26)) + " ===\n")
		for j := 0; j < 1000; j++ {
			builder.WriteString("Log line with important context data\n")
		}
	}
	input := builder.String()

	cfg := filter.LayerConfigs{
		EnableEntropy:    true,
		EnableH2O:        true,
		EnableCompaction: true,
	}
	processor := filter.NewStreamingProcessor(filter.ModeMinimal, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := processor.ProcessStream(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_Streaming_SmallInput tests that streaming falls back to standard for small inputs
func Benchmark_Streaming_SmallInput(b *testing.B) {
	input := strings.Repeat("Small test content\n", 100) // ~500 tokens

	cfg := filter.LayerConfigs{
		EnableEntropy: true,
		EnableH2O:     true,
	}
	processor := filter.NewStreamingProcessor(filter.ModeMinimal, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, _ := processor.ProcessStream(input)
		_ = output
	}
}
