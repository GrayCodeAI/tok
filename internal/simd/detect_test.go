package simd

import (
	"testing"
)

func TestFastCountBytes(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		target byte
		want   int
	}{
		{"empty", "", 'a', 0},
		{"single match", "a", 'a', 1},
		{"no match", "hello", 'z', 0},
		{"multiple matches", "banana", 'a', 3},
		{"all match", "aaaa", 'a', 4},
		{"long string", "the quick brown fox jumps over the lazy dog", ' ', 8},
		{"16+ bytes", "aaaaaaaaaaaaaaaaaaaa", 'a', 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FastCountBytes(tt.data, tt.target)
			if got != tt.want {
				t.Errorf("FastCountBytes(%q, %q) = %d, want %d", tt.data, tt.target, got, tt.want)
			}
		})
	}
}

func TestFastLower(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"already lower", "hello", "hello"},
		{"all upper", "HELLO", "hello"},
		{"mixed", "HeLLo WoRLd", "hello world"},
		{"no letters", "12345", "12345"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FastLower(tt.input)
			if got != tt.want {
				t.Errorf("FastLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFastEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"both empty", "", "", true},
		{"equal short", "abc", "abc", true},
		{"different", "abc", "abd", false},
		{"different length", "abc", "abcd", false},
		{"equal long", "the quick brown fox jumps over the lazy dog", "the quick brown fox jumps over the lazy dog", true},
		{"diff at end", "hello world!", "hello world.", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FastEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("FastEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFastContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty substr", "hello", "", true},
		{"found", "hello world", "world", true},
		{"not found", "hello world", "foo", false},
		{"substr longer", "hi", "hello", false},
		{"exact match", "abc", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FastContains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("FastContains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substrs []string
		want    bool
	}{
		{"found first", "hello world", []string{"foo", "world"}, true},
		{"found last", "hello world", []string{"foo", "bar", "hello"}, true},
		{"none found", "hello world", []string{"foo", "bar"}, false},
		{"empty list", "hello", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsAny(tt.s, tt.substrs)
			if got != tt.want {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, got, tt.want)
			}
		})
	}
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		name string
		s    string
		word string
		want bool
	}{
		{"exact word", "hello world", "hello", true},
		{"substring not word", "hello world", "hell", false},
		{"middle word", "the quick brown fox", "brown", true},
		{"empty word", "hello", "", false},
		{"case insensitive", "Hello WORLD", "world", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsWord(tt.s, tt.word)
			if got != tt.want {
				t.Errorf("ContainsWord(%q, %q) = %v, want %v", tt.s, tt.word, got, tt.want)
			}
		})
	}
}

func TestIsWordChar(t *testing.T) {
	if !IsWordChar('a') || !IsWordChar('Z') || !IsWordChar('0') || !IsWordChar('_') {
		t.Error("expected a, Z, 0, _ to be word chars")
	}
	if IsWordChar('-') || IsWordChar(' ') || IsWordChar('.') {
		t.Error("expected -, space, . to not be word chars")
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single", "hello", []string{"hello"}},
		{"multiple", "hello world", []string{"hello", "world"}},
		{"punctuation", "hello, world!", []string{"hello", "world"}},
		{"apostrophe", "don't stop", []string{"don't", "stop"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitWords(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("SplitWords(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitWords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestHasANSI_Alias(t *testing.T) {
	if HasANSI("\x1b[31m") != FastHasANSI("\x1b[31m") {
		t.Error("HasANSI should be alias for FastHasANSI")
	}
}

func TestProcess_Alias(t *testing.T) {
	input := "\x1b[31mred\x1b[0m"
	if Process(input) != StripANSI(input) {
		t.Error("Process should be alias for StripANSI")
	}
}
