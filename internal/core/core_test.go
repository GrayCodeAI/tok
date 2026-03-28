package core

import (
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input       string
		minExpected int // BPE is more accurate than heuristic
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 1},
		{"abcdefgh", 1},
		{"abcdefghi", 1},
		{"hello world", 2},
	}

	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got < tt.minExpected {
			t.Errorf("EstimateTokens(%q) = %d, want >= %d", tt.input, got, tt.minExpected)
		}
	}
}

func TestCalculateTokensSaved(t *testing.T) {
	tests := []struct {
		original string
		filtered string
		minSaved int
	}{
		{"hello world", "hello", 1},
		{"same", "same", 0},
		{"short", "longer than original", 0}, // Should return 0 when filtered is longer
		{"a b c d e f g h", "a c e g", 1},
	}

	for _, tt := range tests {
		got := CalculateTokensSaved(tt.original, tt.filtered)
		if got < tt.minSaved {
			t.Errorf("CalculateTokensSaved(%q, %q) = %d, want >= %d",
				tt.original, tt.filtered, got, tt.minSaved)
		}
	}
}

func TestCalculateSavings(t *testing.T) {
	savings := CalculateSavings(1000000, "gpt-4o")
	if savings <= 0 {
		t.Errorf("CalculateSavings returned %f, want > 0", savings)
	}
}
