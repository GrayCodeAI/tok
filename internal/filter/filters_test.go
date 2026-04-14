package filter

import (
	"strings"
	"testing"
)

// Test data
const smallInput = "Hello world"

var (
	mediumInput = strings.Repeat("test line\n", 100)
	largeInput  = strings.Repeat("test line with more content\n", 1000)
)

// TestEntropyFilter tests entropy-based filtering
func TestEntropyFilter(t *testing.T) {
	f := NewEntropyFilter()

	tests := []struct {
		name  string
		input string
		mode  Mode
	}{
		{"empty", "", ModeMinimal},
		{"small", smallInput, ModeMinimal},
		{"medium", mediumInput, ModeAggressive},
		{"large", largeInput, ModeAggressive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, saved := f.Apply(tt.input, tt.mode)
			if len(output) > len(tt.input) {
				t.Errorf("output larger than input")
			}
			if saved < 0 {
				t.Errorf("negative tokens saved: %d", saved)
			}
		})
	}
}

// TestPerplexityFilter tests perplexity-based pruning
func TestPerplexityFilter(t *testing.T) {
	f := NewPerplexityFilter()

	input := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
	output, saved := f.Apply(input, ModeAggressive)

	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}
}

// TestASTPreserveFilter tests AST-aware compression
func TestASTPreserveFilter(t *testing.T) {
	f := NewASTPreserveFilter()

	code := `func main() {
		fmt.Println("hello")
		x := 42
		return x
	}`

	output, saved := f.Apply(code, ModeMinimal)
	if !strings.Contains(output, "func") {
		t.Error("should preserve function declaration")
	}
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}
}

// TestBudgetEnforcer tests budget enforcement
func TestBudgetEnforcer(t *testing.T) {
	f := NewBudgetEnforcer(100)

	// Test with large input that should trigger budget enforcement
	input := strings.Repeat("word ", 200)
	output, saved := f.Apply(input, ModeMinimal)

	// Output should not be empty
	if output == "" {
		t.Error("output should not be empty")
	}

	// Tokens saved should not be negative
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}

	// With a large input and small budget, we should have saved something
	// or at least gotten a valid output
	if len(output) > len(input) {
		t.Error("output should not be larger than input")
	}
}

// TestH2OFilter tests heavy-hitter oracle
func TestH2OFilter(t *testing.T) {
	f := NewH2OFilter()

	input := strings.Repeat("important ", 10) + strings.Repeat("noise ", 50)
	output, saved := f.Apply(input, ModeAggressive)

	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}
}

// TestAttentionSinkFilter tests attention sink stability
func TestAttentionSinkFilter(t *testing.T) {
	f := NewAttentionSinkFilter()

	input := "First line\nSecond line\nThird line\nLast line\n"
	output, _ := f.Apply(input, ModeMinimal)

	if !strings.Contains(output, "First") || !strings.Contains(output, "Last") {
		t.Error("should preserve first and last lines")
	}
}

// TestMetaTokenFilter tests meta-token compression
func TestMetaTokenFilter(t *testing.T) {
	f := NewMetaTokenFilterWithConfig(DefaultMetaTokenConfig())

	input := "error error error warning warning info"
	output, saved := f.Apply(input, ModeAggressive)

	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}
}

// TestSemanticChunkFilter tests semantic chunking
func TestSemanticChunkFilter(t *testing.T) {
	f := NewSemanticChunkFilterWithConfig(DefaultSemanticChunkConfig())

	input := "Paragraph 1.\n\nParagraph 2.\n\nParagraph 3.\n"
	output, _ := f.Apply(input, ModeMinimal)

	if output == "" {
		t.Error("output should not be empty")
	}
}

// TestLazyPrunerFilter tests lazy pruning
func TestLazyPrunerFilter(t *testing.T) {
	f := NewLazyPrunerFilterWithConfig(DefaultLazyPrunerConfig())

	input := strings.Repeat("token ", 200)
	output, saved := f.Apply(input, ModeAggressive)

	if len(output) >= len(input) {
		t.Error("should compress input")
	}
	if saved < 0 {
		t.Errorf("negative tokens saved: %d", saved)
	}
}

// TestSemanticAnchorFilter tests semantic anchoring
func TestSemanticAnchorFilter(t *testing.T) {
	f := NewSemanticAnchorFilterWithConfig(DefaultSemanticAnchorConfig())

	input := "Important: this is critical\nNormal text\nAnother important point"
	output, _ := f.Apply(input, ModeMinimal)

	if !strings.Contains(output, "Important") {
		t.Error("should preserve anchor points")
	}
}

// TestAgentMemoryFilter tests agent memory compression
func TestAgentMemoryFilter(t *testing.T) {
	f := NewAgentMemoryFilterWithConfig(DefaultAgentMemoryConfig())

	input := "Action: read file\nResult: success\nAction: write file\nResult: success"
	output, _ := f.Apply(input, ModeMinimal)

	if output == "" {
		t.Error("output should not be empty")
	}
}

// TestFilterChaining tests multiple filters in sequence
func TestFilterChaining(t *testing.T) {
	input := largeInput

	// Apply filters in sequence
	f1 := NewEntropyFilter()
	output1, _ := f1.Apply(input, ModeMinimal)

	f2 := NewPerplexityFilter()
	output2, _ := f2.Apply(output1, ModeMinimal)

	f3 := NewBudgetEnforcer(100)
	output3, _ := f3.Apply(output2, ModeMinimal)

	if len(output3) > len(input) {
		t.Error("chained filters should reduce size")
	}
}

// TestFilterNilSafety tests nil input handling
func TestFilterNilSafety(t *testing.T) {
	filters := []Filter{
		NewEntropyFilter(),
		NewPerplexityFilter(),
		NewASTPreserveFilter(),
		NewBudgetEnforcer(100),
	}

	for _, f := range filters {
		output, _ := f.Apply("", ModeMinimal)
		if output != "" {
			t.Errorf("%s: empty input should return empty output", f.Name())
		}
	}
}

// BenchmarkEntropyFilter benchmarks entropy filtering
func BenchmarkEntropyFilter(b *testing.B) {
	f := NewEntropyFilter()
	input := mediumInput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeMinimal)
	}
}

// BenchmarkPerplexityFilter benchmarks perplexity filtering
func BenchmarkPerplexityFilter(b *testing.B) {
	f := NewPerplexityFilter()
	input := mediumInput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeAggressive)
	}
}

// BenchmarkH2OFilter benchmarks H2O filtering
func BenchmarkH2OFilter(b *testing.B) {
	f := NewH2OFilter()
	input := mediumInput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeAggressive)
	}
}
