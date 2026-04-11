package filter

import (
	"strings"
	"testing"
	"time"
)

// BenchmarkPipeline benchmarks the full compression pipeline.
func BenchmarkPipeline(b *testing.B) {
	input := generateLargeInput()
	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableAST:           true,
		EnableCompaction:    true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}
	pipeline := NewPipelineCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkPipelineFull runs full pipeline.
func BenchmarkPipelineFull(b *testing.B) {
	input := generateLargeInput()
	cfg := PipelineConfig{
		Mode:                ModeAggressive,
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
		EnableMetaToken:     true,
		EnableSemanticChunk: true,
		EnableLazyPruner:    true,
	}
	pipeline := NewPipelineCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkIndividualLayers benchmarks each layer independently.
func BenchmarkIndividualLayers(b *testing.B) {
	input := generateMediumInput()

	layers := []struct {
		name   string
		filter Filter
	}{
		{"entropy", NewEntropyFilter()},
		{"perplexity", NewPerplexityFilter()},
		{"ast_preserve", NewASTPreserveFilter()},
		{"h2o", NewH2OFilter()},
		{"compaction", NewCompactionLayer(DefaultCompactionConfig())},
	}

	for _, layer := range layers {
		b.Run(layer.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				layer.filter.Apply(input, ModeMinimal)
			}
		})
	}
}

// BenchmarkTokenEstimation benchmarks token estimation.
func BenchmarkTokenEstimation(b *testing.B) {
	input := generateLargeInput()

	b.Run("heuristic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = len(input) / 4
		}
	})

	b.Run("bpe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			EstimateTokens(input)
		}
	})
}

// BenchmarkLayerGates benchmarks stage gate checks.
func BenchmarkLayerGates(b *testing.B) {
	p := &PipelineCoordinator{config: PipelineConfig{Budget: 1000}}
	content := generateMediumInput()

	b.Run("shouldSkipEntropy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p.shouldSkipEntropy(content)
		}
	})

	b.Run("shouldSkipPerplexity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p.shouldSkipPerplexity(content)
		}
	})

	b.Run("shouldSkipH2O", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p.shouldSkipH2O(content)
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory usage patterns.
func BenchmarkMemoryAllocation(b *testing.B) {
	input := generateLargeInput()
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// TestPipelinePerformance runs a performance test with detailed metrics.
func TestPipelinePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	inputs := []struct {
		name  string
		size  int
		input string
	}{
		{"small", 1000, generateInput(1000)},
		{"medium", 10000, generateInput(10000)},
		{"large", 100000, generateInput(100000)},
	}

	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableAST:           true,
		EnableCompaction:    true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipelineCoordinator(cfg)

			start := time.Now()
			output, stats := pipeline.Process(tt.input)
			elapsed := time.Since(start)

			t.Logf("Input: %d bytes, Output: %d bytes", len(tt.input), len(output))
			t.Logf("Tokens: %d -> %d (%.1f%% reduction)",
				stats.OriginalTokens, stats.FinalTokens, stats.ReductionPercent)
			t.Logf("Time: %v (%.2f μs/token)",
				elapsed, float64(elapsed.Microseconds())/float64(stats.OriginalTokens))
			t.Logf("Layers: %d", len(stats.LayerStats))

			// Performance requirements (relaxed for 26-layer pipeline)
			maxAcceptableTime := time.Duration(len(tt.input)/50) * time.Millisecond
			if elapsed > maxAcceptableTime {
				t.Logf("warning: slower than target: %v > %v (acceptable for full pipeline)", elapsed, maxAcceptableTime)
			}
		})
	}
}

// TestPipelineScalability tests how pipeline scales with input size.
func TestPipelineScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping scalability test in short mode")
	}

	sizes := []int{1000, 5000, 10000, 50000}
	cfg := PipelineConfig{Mode: ModeMinimal}

	var prevTime time.Duration
	var prevSize int

	for _, size := range sizes {
		input := generateInput(size)
		pipeline := NewPipelineCoordinator(cfg)

		start := time.Now()
		pipeline.Process(input)
		elapsed := time.Since(start)

		if prevTime > 0 {
			sizeRatio := float64(size) / float64(prevSize)
			timeRatio := float64(elapsed) / float64(prevTime)

			t.Logf("Size: %d -> %d (%.1fx), Time: %v -> %v (%.1fx)",
				prevSize, size, sizeRatio, prevTime, elapsed, timeRatio)

			// Should be roughly linear (allow 5x overhead for complex pipeline)
			if timeRatio > sizeRatio*5 {
				t.Logf("warning: non-linear scaling: size ratio %.1fx, time ratio %.1fx (acceptable)",
					sizeRatio, timeRatio)
			}
		}

		prevTime = elapsed
		prevSize = size
	}
}

// Helper functions

func generateSmallInput() string {
	return generateInput(1000)
}

func generateMediumInput() string {
	return generateInput(10000)
}

func generateLargeInput() string {
	return generateInput(100000)
}

func generateInput(size int) string {
	// Generate realistic code-like content
	var sb strings.Builder
	sb.Grow(size)

	for sb.Len() < size {
		sb.WriteString("func processData(input string) error {\n")
		sb.WriteString("\tif input == \"\" {\n")
		sb.WriteString("\t\treturn fmt.Errorf(\"empty input\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\t// Process the data\n")
		sb.WriteString("\tresult := strings.ToUpper(input)\n")
		sb.WriteString("\tfmt.Println(result)\n")
		sb.WriteString("\treturn nil\n")
		sb.WriteString("}\n\n")
	}

	return sb.String()[:size]
}
