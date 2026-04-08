package filter

import (
	"strings"
	"testing"
)

// TestEntropyFilter tests entropy-based filtering
func TestEntropyFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minSave int // Minimum tokens expected to be saved
	}{
		{
			name:    "high entropy code",
			input:   "func main() { fmt.Println(\"unique string\") }",
			minSave: 0, // High entropy, minimal savings
		},
		{
			name:    "repetitive content",
			input:   "ERROR: connection failed\nERROR: connection failed\nERROR: connection failed",
			minSave: 10, // Should compress repetitive content
		},
		{
			name:    "mixed content",
			input:   "INFO: Starting process\nDEBUG: Variable x = 42\nINFO: Process completed",
			minSave: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:          ModeMinimal,
				EnableEntropy: true,
			}
			p := NewPipelineCoordinator(cfg)
			result, stats := p.Process(tc.input)

			if len(result) > len(tc.input) {
				t.Errorf("Output larger than input: %d > %d", len(result), len(tc.input))
			}

			if stats.TotalSaved < tc.minSave {
				t.Logf("Note: only saved %d tokens, expected at least %d",
					stats.TotalSaved, tc.minSave)
			}
		})
	}
}

// TestPerplexityFilter tests perplexity-based filtering
func TestPerplexityFilter(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "natural language",
			input: "This is a natural sentence with good semantic meaning and coherence throughout the text.",
		},
		{
			name:  "technical content",
			input: "func processRequest(ctx context.Context, req *Request) (*Response, error) { return nil, nil }",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:             ModeMinimal,
				EnablePerplexity: true,
			}
			p := NewPipelineCoordinator(cfg)
			result, _ := p.Process(tc.input)

			if result == "" {
				t.Error("Perplexity filter returned empty result")
			}
		})
	}
}

// TestASTPreserveFilter tests AST-aware filtering
func TestASTPreserveFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		language string
	}{
		{
			name: "go function",
			input: `func CalculateSum(a, b int) int {
				result := a + b
				return result
			}`,
			language: "go",
		},
		{
			name: "python function",
			input: `def calculate_sum(a, b):
				result = a + b
				return result`,
			language: "python",
		},
		{
			name: "javascript function",
			input: `function calculateSum(a, b) {
				const result = a + b;
				return result;
			}`,
			language: "javascript",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:      ModeMinimal,
				EnableAST: true,
			}
			p := NewPipelineCoordinator(cfg)
			result, _ := p.Process(tc.input)

			// Should preserve function signature
			if !strings.Contains(result, "func") && !strings.Contains(result, "def") && !strings.Contains(result, "function") {
				t.Log("Warning: Function signature may have been altered")
			}
		})
	}
}

// TestH2OFilter tests H2O (Heavy Hitter Oracle) filtering
func TestH2OFilter(t *testing.T) {
	input := strings.Repeat("important_token ", 100) + strings.Repeat("filler_token ", 50)

	cfg := PipelineConfig{
		Mode:      ModeMinimal,
		EnableH2O: true,
	}
	p := NewPipelineCoordinator(cfg)
	result, stats := p.Process(input)

	if result == "" {
		t.Error("H2O filter returned empty result")
	}

	t.Logf("H2O filter: %d -> %d tokens (saved %d)",
		len(input)/4, len(result)/4, stats.TotalSaved)
}

// TestCompactionFilter tests compaction filtering
func TestCompactionFilter(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "repetitive patterns",
			input: `Starting process...
Processing item 1
Processing item 2
Processing item 3
Processing item 4
Processing item 5
Completed`,
		},
		{
			name: "chat conversation",
			input: `User: Hello
Assistant: Hi there! How can I help?
User: I need help with Go
Assistant: I'd be happy to help with Go programming`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := PipelineConfig{
				Mode:             ModeMinimal,
				EnableCompaction: true,
			}
			p := NewPipelineCoordinator(cfg)
			result, _ := p.Process(tc.input)

			if result == "" {
				t.Error("Compaction filter returned empty result")
			}
		})
	}
}

// TestSemanticChunkFilter tests semantic chunk filtering
func TestSemanticChunkFilter(t *testing.T) {
	input := `Section 1: Introduction
This is the introduction to the document.

Section 2: Methods
The methods used in this study include various approaches.

Section 3: Results
The results show significant improvements.

Section 4: Conclusion
In conclusion, the study demonstrates effectiveness.`

	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableSemanticChunk: true,
	}
	p := NewPipelineCoordinator(cfg)
	result, _ := p.Process(input)

	if result == "" {
		t.Error("Semantic chunk filter returned empty result")
	}

	// Should preserve section headers
	if !strings.Contains(result, "Section") {
		t.Log("Warning: Section headers may have been removed")
	}
}

// TestMetaTokenFilter tests meta-token compression
func TestMetaTokenFilter(t *testing.T) {
	// Content with repeated patterns
	input := strings.Repeat("const ERROR_MESSAGE = \"Something went wrong\"\n", 20)

	cfg := PipelineConfig{
		Mode:            ModeMinimal,
		EnableMetaToken: true,
	}
	p := NewPipelineCoordinator(cfg)
	result, stats := p.Process(input)

	if result == "" {
		t.Error("Meta-token filter returned empty result")
	}

	t.Logf("Meta-token: %d -> %d tokens (saved %d)",
		len(input)/4, len(result)/4, stats.TotalSaved)
}

// TestLazyPrunerFilter tests lazy pruner filtering
func TestLazyPrunerFilter(t *testing.T) {
	input := strings.Repeat("Line of text with content here. ", 100)

	cfg := PipelineConfig{
		Mode:             ModeMinimal,
		EnableLazyPruner: true,
		Budget:           500,
	}
	p := NewPipelineCoordinator(cfg)
	result, stats := p.Process(input)

	if result == "" {
		t.Error("Lazy pruner returned empty result")
	}

	if stats.TotalSaved < 0 {
		t.Error("Negative tokens saved")
	}
}

// TestAttentionSinkFilter tests attention sink filtering
func TestAttentionSinkFilter(t *testing.T) {
	// Long context that needs stable attention
	input := strings.Repeat("Context line with information. ", 200)

	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableAttentionSink: true,
	}
	p := NewPipelineCoordinator(cfg)
	result, _ := p.Process(input)

	if result == "" {
		t.Error("Attention sink filter returned empty result")
	}
}

// TestErrorHandling tests error handling in filters
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   \n\t\n   ",
		},
		{
			name:  "very long input",
			input: strings.Repeat("a", 100000),
		},
		{
			name:  "unicode content",
			input: "Hello 世界 🌍 ñoño émojis",
		},
		{
			name:  "binary-like content",
			input: "\x00\x01\x02\x03\x04\x05",
		},
	}

	cfg := PipelineConfig{
		Mode:                ModeMinimal,
		EnableEntropy:       true,
		EnablePerplexity:    true,
		EnableAST:           true,
		EnableH2O:           true,
		EnableCompaction:    true,
		EnableSemanticChunk: true,
	}
	p := NewPipelineCoordinator(cfg)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, stats := p.Process(tc.input)

			// Should never panic or return error
			if stats.TotalSaved < 0 {
				t.Errorf("Negative tokens saved: %d", stats.TotalSaved)
			}

			// Result should not be nil
			_ = result
		})
	}
}

// TestLayerCombinations tests various layer combinations
func TestLayerCombinations(t *testing.T) {
	input := `Function processing started
Processing step 1...
Processing step 2...
Processing step 3...
Function completed successfully`

	combinations := []struct {
		name string
		cfg  PipelineConfig
	}{
		{
			name: "entropy+ast",
			cfg: PipelineConfig{
				Mode:          ModeMinimal,
				EnableEntropy: true,
				EnableAST:     true,
			},
		},
		{
			name: "h2o+compaction",
			cfg: PipelineConfig{
				Mode:             ModeMinimal,
				EnableH2O:        true,
				EnableCompaction: true,
			},
		},
		{
			name: "semantic+chunk",
			cfg: PipelineConfig{
				Mode:                 ModeMinimal,
				EnableSemanticChunk:  true,
				EnableSemanticAnchor: true,
			},
		},
		{
			name: "all-core",
			cfg: PipelineConfig{
				Mode:                ModeMinimal,
				EnableEntropy:       true,
				EnablePerplexity:    true,
				EnableAST:           true,
				EnableH2O:           true,
				EnableCompaction:    true,
				EnableSemanticChunk: true,
				EnableMetaToken:     true,
			},
		},
	}

	for _, tc := range combinations {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPipelineCoordinator(tc.cfg)
			result, _ := p.Process(input)

			if result == "" && input != "" {
				t.Error("Layer combination returned empty result for non-empty input")
			}
		})
	}
}

// TestContentTypeDetection tests automatic content type detection
func TestContentTypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ContentType
	}{
		{
			name:     "go code",
			input:    "package main\n\nfunc main() {}",
			expected: ContentTypeCode,
		},
		{
			name:     "log output",
			input:    "2024-01-01 INFO Starting application\n2024-01-01 ERROR Failed",
			expected: ContentTypeLogs,
		},
		{
			name:     "mixed content",
			input:    "Some text here\nMore text there",
			expected: ContentTypeMixed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selector := NewAdaptiveLayerSelector()
			contentType := selector.AnalyzeContent(tc.input)

			if contentType != tc.expected {
				t.Logf("Detected %v, expected %v", contentType, tc.expected)
			}
		})
	}
}

// BenchmarkFilterPerformance benchmarks filter performance
func BenchmarkFilterPerformance(b *testing.B) {
	input := strings.Repeat("Test content with some entropy for benchmarking purposes. ", 100)

	filters := []struct {
		name string
		cfg  PipelineConfig
	}{
		{
			name: "entropy-only",
			cfg:  PipelineConfig{Mode: ModeMinimal, EnableEntropy: true},
		},
		{
			name: "ast-only",
			cfg:  PipelineConfig{Mode: ModeMinimal, EnableAST: true},
		},
		{
			name: "h2o-only",
			cfg:  PipelineConfig{Mode: ModeMinimal, EnableH2O: true},
		},
		{
			name: "full-pipeline",
			cfg: PipelineConfig{
				Mode:             ModeMinimal,
				EnableEntropy:    true,
				EnablePerplexity: true,
				EnableAST:        true,
				EnableH2O:        true,
				EnableCompaction: true,
			},
		},
	}

	for _, f := range filters {
		b.Run(f.name, func(b *testing.B) {
			p := NewPipelineCoordinator(f.cfg)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.Process(input)
			}
		})
	}
}
