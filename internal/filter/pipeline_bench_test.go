package filter

import (
	"sync"
	"testing"
)

// BenchmarkPipeline_ProcessSmall benchmarks processing small input
func BenchmarkPipeline_ProcessSmall(b *testing.B) {
	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableGoalDriven:    true,
		EnableAST:           true,
		EnableContrastive:   true,
		EnableEvaluator:     true,
		EnableGist:          true,
		EnableHierarchical:  true,
		EnableCompaction:    true,
		EnableAttribution:   true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}
	p := NewPipelineCoordinator(cfg)
	input := "This is a small test input for the compression pipeline."

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Process(input)
	}
}

// BenchmarkPipeline_ProcessMedium benchmarks processing medium input
func BenchmarkPipeline_ProcessMedium(b *testing.B) {
	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableGoalDriven:    true,
		EnableAST:           true,
		EnableContrastive:   true,
		EnableEvaluator:     true,
		EnableGist:          true,
		EnableHierarchical:  true,
		EnableCompaction:    true,
		EnableAttribution:   true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}
	p := NewPipelineCoordinator(cfg)
	// ~1KB input
	input := ""
	for i := 0; i < 50; i++ {
		input += "This is line " + string(rune('0'+i%10)) + " of the test input for the compression pipeline. "
		input += "It contains various words and patterns that should be processed.\n"
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Process(input)
	}
}

// BenchmarkPipeline_ProcessWithBudget benchmarks processing with budget constraint
func BenchmarkPipeline_ProcessWithBudget(b *testing.B) {
	cfg := PipelineConfig{
		Mode:                ModeAggressive,
		Budget:              100,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableGoalDriven:    true,
		EnableAST:           true,
		EnableContrastive:   true,
		EnableEvaluator:     true,
		EnableGist:          true,
		EnableHierarchical:  true,
		EnableCompaction:    true,
		EnableAttribution:   true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}
	p := NewPipelineCoordinator(cfg)
	input := ""
	for i := 0; i < 50; i++ {
		input += "This is line " + string(rune('0'+i%10)) + " of the test input for the compression pipeline.\n"
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Process(input)
	}
}

// BenchmarkEstimateTokens benchmarks token estimation
func BenchmarkEstimateTokens_Small(b *testing.B) {
	input := "Small input text"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EstimateTokens(input)
	}
}

func BenchmarkEstimateTokens_Medium(b *testing.B) {
	input := ""
	for i := 0; i < 100; i++ {
		input += "This is a medium sized input that should be processed quickly. "
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EstimateTokens(input)
	}
}

func BenchmarkEstimateTokens_Large(b *testing.B) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += "This is a larger input that should still be processed efficiently by the estimation algorithm. "
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EstimateTokens(input)
	}
}

// Benchmark individual layers
func BenchmarkLayer_Entropy(b *testing.B) {
	filter := NewEntropyFilter()
	input := "This is test input with some entropy and various patterns"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		filter.Apply(input, ModeMinimal)
	}
}

func BenchmarkLayer_Perplexity(b *testing.B) {
	filter := NewPerplexityFilter()
	input := "Test input for perplexity calculation with multiple words"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		filter.Apply(input, ModeMinimal)
	}
}

// Benchmark parallel processing.
// NOTE: PipelineCoordinator.Process is not goroutine-safe. Each goroutine
// gets its own coordinator via sync.Pool to measure true throughput.
func BenchmarkPipeline_ProcessParallel(b *testing.B) {
	cfg := PipelineConfig{
		Mode:             ModeMinimal,
		EnableEntropy:    true,
		EnablePerplexity: true,
		EnableAST:        true,
	}
	input := "Test input for parallel processing benchmark"

	pool := sync.Pool{
		New: func() any {
			return NewPipelineCoordinator(cfg)
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		p := pool.Get().(*PipelineCoordinator)
		defer pool.Put(p)
		for pb.Next() {
			_, _, _ = p.Process(input)
		}
	})
}
