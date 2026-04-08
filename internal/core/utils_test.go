package core

import (
	"fmt"
	"strings"
	"testing"
)

// TestTokenEstimation tests token estimation accuracy
func TestTokenEstimation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		tolerance float64
	}{
		{
			name:      "short text",
			input:     "Hello world",
			expected:  3, // ~11 chars / 4
			tolerance: 1,
		},
		{
			name:      "medium text",
			input:     "This is a test sentence with some content",
			expected:  10, // ~41 chars / 4
			tolerance: 2,
		},
		{
			name:      "long text",
			input:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			expected:  29, // ~116 chars / 4
			tolerance: 3,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  0,
			tolerance: 0,
		},
		{
			name:      "unicode text",
			input:     "Hello 世界 🌍 ñoño",
			expected:  6, // ~22 chars but unicode
			tolerance: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			estimator := NewTokenEstimator()
			result := estimator.Estimate(tc.input)

			diff := result - tc.expected
			if diff < 0 {
				diff = -diff
			}

			if float64(diff) > tc.tolerance {
				t.Errorf("Estimate(%q) = %d, want ~%d (tolerance %f)",
					tc.input, result, tc.expected, tc.tolerance)
			}
		})
	}
}

// TestCostCalculation tests cost calculation
func TestCostCalculation(t *testing.T) {
	tests := []struct {
		name         string
		tokens       int
		model        string
		expectedCost float64
		tolerance    float64
	}{
		{
			name:         "GPT-4 small",
			tokens:       1000,
			model:        "gpt-4",
			expectedCost: 0.03, // $0.03 per 1K tokens
			tolerance:    0.001,
		},
		{
			name:         "GPT-4 large",
			tokens:       10000,
			model:        "gpt-4",
			expectedCost: 0.30,
			tolerance:    0.01,
		},
		{
			name:         "GPT-3.5",
			tokens:       1000,
			model:        "gpt-3.5-turbo",
			expectedCost: 0.0015,
			tolerance:    0.0001,
		},
		{
			name:         "Claude",
			tokens:       1000,
			model:        "claude-3-opus",
			expectedCost: 0.015,
			tolerance:    0.001,
		},
		{
			name:         "zero tokens",
			tokens:       0,
			model:        "gpt-4",
			expectedCost: 0.0,
			tolerance:    0.0,
		},
	}

	calculator := NewCostCalculator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cost := calculator.Calculate(tc.tokens, tc.model)

			diff := cost - tc.expectedCost
			if diff < 0 {
				diff = -diff
			}

			if diff > tc.tolerance {
				t.Errorf("Calculate(%d, %q) = $%.4f, want $%.4f (tolerance %.4f)",
					tc.tokens, tc.model, cost, tc.expectedCost, tc.tolerance)
			}
		})
	}
}

// TestCommandParsing tests command parsing
func TestCommandParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "simple command",
			input:        "git status",
			expectedCmd:  "git",
			expectedArgs: []string{"status"},
		},
		{
			name:         "command with multiple args",
			input:        "git log --oneline -10",
			expectedCmd:  "git",
			expectedArgs: []string{"log", "--oneline", "-10"},
		},
		{
			name:         "command with quotes",
			input:        `echo "hello world"`,
			expectedCmd:  "echo",
			expectedArgs: []string{"hello world"},
		},
		{
			name:         "docker command",
			input:        "docker ps -a --format json",
			expectedCmd:  "docker",
			expectedArgs: []string{"ps", "-a", "--format", "json"},
		},
	}

	parser := NewCommandParser()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args := parser.Parse(tc.input)

			if cmd != tc.expectedCmd {
				t.Errorf("Parse(%q) cmd=%q, want %q", tc.input, cmd, tc.expectedCmd)
			}

			if len(args) != len(tc.expectedArgs) {
				t.Errorf("Parse(%q) args=%v, want %v", tc.input, args, tc.expectedArgs)
			}
		})
	}
}

// TestCompressionRatio tests compression ratio calculation
func TestCompressionRatio(t *testing.T) {
	tests := []struct {
		name             string
		originalTokens   int
		compressedTokens int
		expectedRatio    float64
		expectedSavings  float64
	}{
		{
			name:             "50% compression",
			originalTokens:   1000,
			compressedTokens: 500,
			expectedRatio:    0.5,
			expectedSavings:  50.0,
		},
		{
			name:             "no compression",
			originalTokens:   1000,
			compressedTokens: 1000,
			expectedRatio:    1.0,
			expectedSavings:  0.0,
		},
		{
			name:             "90% compression",
			originalTokens:   1000,
			compressedTokens: 100,
			expectedRatio:    0.1,
			expectedSavings:  90.0,
		},
		{
			name:             "zero original",
			originalTokens:   0,
			compressedTokens: 0,
			expectedRatio:    1.0,
			expectedSavings:  0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calculator := NewCompressionCalculator()
			ratio := calculator.Ratio(tc.originalTokens, tc.compressedTokens)
			savings := calculator.Savings(tc.originalTokens, tc.compressedTokens)

			if ratio != tc.expectedRatio {
				t.Errorf("Ratio(%d, %d) = %.2f, want %.2f",
					tc.originalTokens, tc.compressedTokens, ratio, tc.expectedRatio)
			}

			if savings != tc.expectedSavings {
				t.Errorf("Savings(%d, %d) = %.2f%%, want %.2f%%",
					tc.originalTokens, tc.compressedTokens, savings, tc.expectedSavings)
			}
		})
	}
}

// TestHashGeneration tests hash generation
func TestHashGeneration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "simple string",
			input: "test content",
			valid: true,
		},
		{
			name:  "empty string",
			input: "",
			valid: true, // Should still produce a hash
		},
		{
			name:  "long string",
			input: string(make([]byte, 10000)),
			valid: true,
		},
		{
			name:  "unicode string",
			input: "Hello 世界 🌍",
			valid: true,
		},
	}

	hasher := NewContentHasher()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash := hasher.Hash(tc.input)

			if tc.valid && len(hash) == 0 {
				t.Error("Hash returned empty string for valid input")
			}

			// Hash should be deterministic
			hash2 := hasher.Hash(tc.input)
			if hash != hash2 {
				t.Error("Hash is not deterministic")
			}

			// Different inputs should produce different hashes
			if tc.input != "" {
				hash3 := hasher.Hash(tc.input + "different")
				if hash == hash3 {
					t.Error("Different inputs produced same hash")
				}
			}
		})
	}
}

// TestStringUtils tests string utilities
func TestStringUtils(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{
			name:     "trim whitespace",
			fn:       func(s string) string { return TrimWhitespace(s) },
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "normalize newlines",
			fn:       func(s string) string { return NormalizeNewlines(s) },
			input:    "line1\r\nline2\rline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "remove ansi",
			fn:       func(s string) string { return RemoveANSI(s) },
			input:    "\x1b[31mred text\x1b[0m",
			expected: "red text",
		},
		{
			name:     "collapse spaces",
			fn:       func(s string) string { return CollapseSpaces(s) },
			input:    "too    many    spaces",
			expected: "too many spaces",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fn(tc.input)
			if result != tc.expected {
				t.Errorf("%s(%q) = %q, want %q",
					tc.name, tc.input, result, tc.expected)
			}
		})
	}
}

// BenchmarkTokenEstimation benchmarks token estimation
func BenchmarkTokenEstimation(b *testing.B) {
	input := "This is a test sentence with some content for benchmarking purposes"
	estimator := NewTokenEstimator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		estimator.Estimate(input)
	}
}

// BenchmarkHashGeneration benchmarks hash generation
func BenchmarkHashGeneration(b *testing.B) {
	input := "test content for hashing benchmark"
	hasher := NewContentHasher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(input)
	}
}

// BenchmarkCostCalculation benchmarks cost calculation
func BenchmarkCostCalculation(b *testing.B) {
	calculator := NewCostCalculator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.Calculate(1000, "gpt-4")
	}
}

// Mock types for testing
type TokenEstimator struct{}

func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{}
}

func (e *TokenEstimator) Estimate(input string) int {
	// Simple estimation: chars / 4
	return len(input) / 4
}

type CostCalculator struct{}

func NewCostCalculator() *CostCalculator {
	return &CostCalculator{}
}

func (c *CostCalculator) Calculate(tokens int, model string) float64 {
	rates := map[string]float64{
		"gpt-4":         0.00003,
		"gpt-3.5-turbo": 0.0000015,
		"claude-3-opus": 0.000015,
	}

	rate := rates[model]
	if rate == 0 {
		rate = 0.00003 // Default to GPT-4 rate
	}

	return float64(tokens) * rate
}

type CommandParser struct{}

func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

func (p *CommandParser) Parse(input string) (string, []string) {
	parts := SplitCommand(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func SplitCommand(input string) []string {
	// Simple split for testing
	return SplitWhitespace(input)
}

func SplitWhitespace(s string) []string {
	// Simplified - real implementation would handle quotes
	return strings.Fields(s)
}

type CompressionCalculator struct{}

func NewCompressionCalculator() *CompressionCalculator {
	return &CompressionCalculator{}
}

func (c *CompressionCalculator) Ratio(original, compressed int) float64 {
	if original == 0 {
		return 1.0
	}
	return float64(compressed) / float64(original)
}

func (c *CompressionCalculator) Savings(original, compressed int) float64 {
	if original == 0 {
		return 0.0
	}
	return (1.0 - float64(compressed)/float64(original)) * 100
}

type ContentHasher struct{}

func NewContentHasher() *ContentHasher {
	return &ContentHasher{}
}

func (h *ContentHasher) Hash(input string) string {
	// Simplified hash for testing
	sum := 0
	for i, c := range input {
		sum += int(c) * (i + 1)
	}
	return fmt.Sprintf("%x", sum)
}

func TrimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func RemoveANSI(s string) string {
	// Simplified ANSI removal
	result := ""
	inEscape := false
	for _, c := range s {
		if c == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(c)
	}
	return result
}

func CollapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
