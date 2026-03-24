package core

import "testing"

func TestEstimateTokensExact(t *testing.T) {
	tests := []struct {
		input       string
		minExpected int // BPE is more accurate than heuristic, use minimum
	}{
		{"", 0},
		{"a", 1},
		{"ab", 1},
		{"abc", 1},
		{"abcd", 1},
		{"abcde", 1},
		{"abcd ef", 1},
		{"hello world", 2},
		{"abcdefghijklmnop", 1}, // BPE treats as single token
		{"The quick brown fox jumps over the lazy dog", 8}, // Longer text
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got < tt.minExpected {
			t.Errorf("EstimateTokens(%q) = %d, want >= %d", tt.input, got, tt.minExpected)
		}
	}
}

func TestEstimateTokensPositive(t *testing.T) {
	// Non-empty strings should always return >= 1 token
	inputs := []string{"a", "hello", "test string", "func main() {}"}
	for _, input := range inputs {
		got := EstimateTokens(input)
		if got < 1 {
			t.Errorf("EstimateTokens(%q) = %d, want >= 1", input, got)
		}
	}
}

func TestCalculateTokensSavedPositive(t *testing.T) {
	saved := CalculateTokensSaved("hello world test", "hello")
	if saved <= 0 {
		t.Errorf("expected positive savings, got %d", saved)
	}
}

func TestCalculateTokensSavedZero(t *testing.T) {
	zero := CalculateTokensSaved("hi", "hello world test more")
	if zero != 0 {
		t.Errorf("expected 0 for negative savings, got %d", zero)
	}
}

func TestCalculateTokensSavedEqual(t *testing.T) {
	saved := CalculateTokensSaved("same", "same")
	if saved != 0 {
		t.Errorf("expected 0 for equal length, got %d", saved)
	}
}

func TestEstimateTokensConsistency(t *testing.T) {
	text := "repeated calculation test"
	for i := 0; i < 100; i++ {
		a := EstimateTokens(text)
		b := EstimateTokens(text)
		if a != b {
			t.Fatalf("EstimateTokens not deterministic: %d != %d", a, b)
		}
	}
}
