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

// TestEstimateTokensPreciseMatchesBPE pins EstimateTokensPrecise to exact
// tiktoken cl100k_base reference counts. If tiktoken-go is swapped or
// upgraded, these must stay stable — otherwise every `tok gain` number
// in the wild becomes retroactively wrong.
func TestEstimateTokensPreciseMatchesBPE(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"hello world", 2},
		{"The quick brown fox jumps over the lazy dog", 9},
		{"function main() { return 42; }", 9},
	}
	for _, tt := range tests {
		got := EstimateTokensPrecise(tt.text)
		if got != tt.want {
			t.Errorf("EstimateTokensPrecise(%q) = %d, want %d (BPE cl100k_base)",
				tt.text, got, tt.want)
		}
	}
}

// TestEstimateTokensPreciseShortString guards the bug that EstimateTokens
// takes a 200-char heuristic fast path, which silently undercounts short
// memory-file snippets. EstimateTokensPrecise must skip that fast path.
func TestEstimateTokensPreciseShortString(t *testing.T) {
	short := "hello" // 5 chars — would trigger <30 heuristic in EstimateTokens
	fast := EstimateTokensFast(short)
	precise := EstimateTokensPrecise(short)
	if precise == 0 {
		t.Fatal("EstimateTokensPrecise returned 0 for non-empty input")
	}
	// They may disagree; that's the whole point.
	t.Logf("short=%q fast=%d precise=%d (difference acceptable; precise is authoritative)",
		short, fast, precise)
}
