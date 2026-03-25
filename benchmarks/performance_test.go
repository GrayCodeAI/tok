package benchmarks

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// BenchmarkTokmanPerformance measures Tokman's 22-layer pipeline performance

// defaultConfig returns a default pipeline config for benchmarking
func defaultConfig(mode filter.Mode) filter.PipelineConfig {
	return filter.PipelineConfig{
		Mode:                mode,
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
}

// ============================================================================
// FEATURE COMPARISON BENCHMARKS
// ============================================================================

// Benchmark_TokenReduction_Core measures basic token reduction
func Benchmark_TokenReduction_Core(b *testing.B) {
	cfg := defaultConfig(filter.ModeAggressive)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("This is a sample log line with some content that repeats. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, _ := pipeline.Process(input)
		_ = output
	}
}

// Benchmark_Compression_LargeInput tests compression on large inputs
// Target: 95-99% compression on logs via 22-layer pipeline
func Benchmark_Compression_LargeInput(b *testing.B) {
	cfg := defaultConfig(filter.ModeAggressive)
	pipeline := filter.NewPipelineCoordinator(cfg)

	// Simulate large log output (100KB)
	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		builder.WriteString("2024-01-15 10:30:45 INFO [main] Processing request from user_id=12345\n")
		builder.WriteString("2024-01-15 10:30:46 DEBUG [worker] Executing query took 45ms\n")
		builder.WriteString("2024-01-15 10:30:47 INFO [api] Response sent successfully\n")
	}
	input := builder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_H2OFilter_AttentionSinks measures H2O attention sink handling
func Benchmark_H2OFilter_AttentionSinks(b *testing.B) {
	h2o := filter.NewH2OFilter()
	input := strings.Repeat("token line content\n", 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, _ := h2o.Apply(input, filter.ModeAggressive)
		_ = output
	}
}

// ============================================================================
// ADVANCED LAYER BENCHMARKS
// ============================================================================

// Benchmark_MetaToken_LosslessCompression tests Layer 15
// Meta-Token lossless compression (27% token reduction)
func Benchmark_MetaToken_LosslessCompression(b *testing.B) {
	meta := filter.NewMetaTokenFilter()

	// Input with repeated patterns (common in code/logs)
	input := `func main() {
		fmt.Println("Hello, World!")
		fmt.Println("Hello, World!")
		fmt.Println("Hello, World!")
		fmt.Println("Hello, World!")
		fmt.Println("Hello, World!")
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, saved := meta.Apply(input, filter.ModeMinimal)
		b.ReportMetric(float64(saved), "tokens_saved")
		_ = output
	}
}

// Benchmark_LazyPruner_DynamicBudget tests Layer 18
// LazyLLM-style budget-aware pruning (2.34x speedup)
func Benchmark_LazyPruner_DynamicBudget(b *testing.B) {
	lp := filter.NewLazyPrunerFilter()
	input := strings.Repeat("Content line with varying importance scores. ", 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, saved := lp.Apply(input, filter.ModeAggressive)
		b.ReportMetric(float64(saved), "tokens_saved")
		_ = output
	}
}

// Benchmark_SketchStore_Reversibility tests Layer 17
// Sketch-based recovery (90% memory reduction)
func Benchmark_SketchStore_Reversibility(b *testing.B) {
	ss := filter.NewSketchStoreFilter()
	input := strings.Repeat("Revivable content that can be reconstructed. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, saved := ss.Apply(input, filter.ModeAggressive)
		b.ReportMetric(float64(saved), "tokens_saved")
		_ = output
	}
}

// Benchmark_SemanticAnchor_ContextPreservation tests Layer 19
// Attention gradient anchor detection
func Benchmark_SemanticAnchor_ContextPreservation(b *testing.B) {
	anchor := filter.NewSemanticAnchorFilter()
	input := `Introduction to the topic.
Some filler content here.
More filler content.
KEY POINT: This is important.
Additional filler.
Another KEY INSIGHT: Critical finding.
End of document.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, _ := anchor.Apply(input, filter.ModeAggressive)
		_ = output
	}
}

// Benchmark_AgentMemory_KnowledgeExtraction tests Layer 20
// Knowledge graph extraction for agents
func Benchmark_AgentMemory_KnowledgeExtraction(b *testing.B) {
	am := filter.NewAgentMemoryFilter()
	input := `I found the issue in the code.
The solution is to fix the regex pattern.
I discovered that the bug was in the parser.
This is some regular content.
I learned that caching improves performance.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, _ := am.Apply(input, filter.ModeAggressive)
		_ = output
	}
}

// ============================================================================
// THROUGHPUT BENCHMARKS
// ============================================================================

// Benchmark_Throughput_SmallInput tests small input processing speed
func Benchmark_Throughput_SmallInput(b *testing.B) {
	cfg := defaultConfig(filter.ModeMinimal)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := "This is a small test input."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_Throughput_MediumInput tests medium input processing speed
func Benchmark_Throughput_MediumInput(b *testing.B) {
	cfg := defaultConfig(filter.ModeMinimal)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("Medium length content line. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_Throughput_LargeInput tests large input processing speed
func Benchmark_Throughput_LargeInput(b *testing.B) {
	cfg := defaultConfig(filter.ModeAggressive)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("Large content line with more text. ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// ============================================================================
// MODE COMPARISON BENCHMARKS
// ============================================================================

// Benchmark_Mode_None tests passthrough mode
func Benchmark_Mode_None(b *testing.B) {
	cfg := defaultConfig(filter.ModeNone)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("test content\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_Mode_Minimal tests minimal compression mode
func Benchmark_Mode_Minimal(b *testing.B) {
	cfg := defaultConfig(filter.ModeMinimal)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("test content\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_Mode_Aggressive tests aggressive compression mode
func Benchmark_Mode_Aggressive(b *testing.B) {
	cfg := defaultConfig(filter.ModeAggressive)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("test content\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// ============================================================================
// CONTENT TYPE BENCHMARKS
// ============================================================================

// Benchmark_ContentType_Code tests code compression
func Benchmark_ContentType_Code(b *testing.B) {
	cfg := defaultConfig(filter.ModeMinimal)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("Hello, World!")
	fmt.Println("Hello, World!")
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_ContentType_Logs tests log compression
func Benchmark_ContentType_Logs(b *testing.B) {
	cfg := defaultConfig(filter.ModeAggressive)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := strings.Repeat("2024-01-15 10:30:45 INFO [main] Processing request\n", 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// Benchmark_ContentType_Conversation tests conversation compression
func Benchmark_ContentType_Conversation(b *testing.B) {
	cfg := defaultConfig(filter.ModeMinimal)
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := `User: Hello, how are you?
Assistant: I'm doing well, thank you for asking!
User: Can you help me with something?
Assistant: Of course! What do you need help with?
User: I need to implement a compression algorithm.
Assistant: I'd be happy to help with that. Let me explain...`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}
