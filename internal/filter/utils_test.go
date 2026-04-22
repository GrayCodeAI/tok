package filter

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  int // expected number of tokens
	}{
		{"hello world", 2},
		{"one, two, three", 3},
		{"func main() { return }", 3}, // splits on punctuation: func, main, return
		{"", 0},
		{"   ", 0},
		{"a-b-c", 3},
	}

	for _, tt := range tests {
		got := tokenize(tt.input)
		if len(got) != tt.want {
			t.Errorf("tokenize(%q) = %v (%d tokens), want %d tokens", tt.input, got, len(got), tt.want)
		}
	}
}

func TestTokenize_NoEmptyStrings(t *testing.T) {
	got := tokenize("  a   b  c  ")
	for i, w := range got {
		if w == "" {
			t.Errorf("token %d is empty string", i)
		}
	}
}
