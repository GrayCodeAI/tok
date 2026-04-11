package filter

import (
	"strings"
	"testing"
)

func TestSimHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"medium", "the quick brown fox jumps over the lazy dog"},
		{"long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := SimHash(tt.input)
			// Hash should be deterministic
			hash2 := SimHash(tt.input)
			if hash != hash2 {
				t.Errorf("SimHash not deterministic for %q: %d != %d", tt.input, hash, hash2)
			}
		})
	}
}

func TestSimHash_DifferentInputs(t *testing.T) {
	h1 := SimHash("hello world")
	h2 := SimHash("goodbye world")
	// Different inputs should (likely) produce different hashes
	if h1 == h2 {
		t.Log("Warning: different inputs produced same hash (possible but unlikely)")
	}
}

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		a, b uint64
		want int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{1, 0, 1},
		{0xFFFFFFFFFFFFFFFF, 0, 64},
		{0, 0xFFFFFFFFFFFFFFFF, 64},
		{0b1010, 0b0101, 4},
		{0b1111, 0b1111, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := HammingDistance(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("HammingDistance(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsNearDuplicate(t *testing.T) {
	tests := []struct {
		a, b      string
		threshold int
		want      bool
	}{
		{"hello world", "hello world", 3, true},
		{"hello world", "hello wordl", 10, true},
		{"completely different text that has nothing in common with the other", "another completely unrelated string of text here", 3, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := IsNearDuplicate(tt.a, tt.b, tt.threshold)
			if got != tt.want {
				t.Errorf("IsNearDuplicate(%q, %q, %d) = %v, want %v", tt.a, tt.b, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestCrossMessageDedup(t *testing.T) {
	d := NewCrossMessageDedup()
	if d == nil {
		t.Fatal("NewCrossMessageDedup returned nil")
	}

	// First message should not be duplicate
	isDup, result := d.DedupMessage("hello world")
	if isDup {
		t.Error("first message should not be duplicate")
	}
	if result != "hello world" {
		t.Errorf("first message result = %q, want 'hello world'", result)
	}

	// Same message should be duplicate
	isDup, result = d.DedupMessage("hello world")
	if !isDup {
		t.Error("same message should be duplicate")
	}

	// Different message should not be duplicate
	isDup, result = d.DedupMessage("different content here with enough text to be unique")
	if isDup {
		t.Error("different message should not be duplicate")
	}

	if d.Count() < 1 {
		t.Errorf("Count() = %d, want >= 1", d.Count())
	}
}

func TestCrossMessageDedup_Clear(t *testing.T) {
	d := NewCrossMessageDedup()
	d.DedupMessage("message one with enough content")
	d.DedupMessage("message two with enough content")

	d.Clear()
	if d.Count() != 0 {
		t.Errorf("Count() after Clear = %d, want 0", d.Count())
	}
}

func TestGenerateDiff(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline2\nline4"

	diff := generateDiff(old, new)
	if diff == "" {
		t.Error("diff should not be empty for different inputs")
	}
	if !strings.Contains(diff, "[diff]") {
		t.Error("diff should contain '[diff]' marker")
	}
}

func TestGenerateDiff_Identical(t *testing.T) {
	content := "line1\nline2\nline3"
	diff := generateDiff(content, content)
	if diff != "" {
		t.Errorf("identical content should produce empty diff, got %q", diff)
	}
}

func BenchmarkSimHash(b *testing.B) {
	input := "The quick brown fox jumps over the lazy dog. This is a longer piece of text for benchmarking purposes."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SimHash(input)
	}
}

func BenchmarkHammingDistance(bm *testing.B) {
	a, bv := uint64(0x1234567890ABCDEF), uint64(0xFEDCBA0987654321)
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		HammingDistance(a, bv)
	}
}
