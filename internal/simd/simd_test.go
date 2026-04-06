package simd

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "simple color code",
			input:    "\x1b[31mred text\x1b[0m",
			expected: "red text",
		},
		{
			name:     "multiple ANSI codes",
			input:    "\x1b[1m\x1b[31mbold red\x1b[0m normal",
			expected: "bold red normal",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "CSI sequence",
			input:    "before\x1b[2Jafter",
			expected: "beforeafter",
		},
		{
			name:     "OSC sequence",
			input:    "text\x1b]0;Title\x07more",
			expected: "textmore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHasANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"no ANSI", "plain text", false},
		{"has ANSI", "\x1b[31mred\x1b[0m", true},
		{"empty", "", false},
		{"ESC only", "\x1b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasANSI(tt.input)
			if result != tt.expected {
				t.Errorf("HasANSI(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIndexByteSet(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		set      []byte
		expected int
	}{
		{"found first", "hello", []byte{'e', 'o'}, 1},
		{"found last", "hello", []byte{'o'}, 4},
		{"not found", "hello", []byte{'x', 'y'}, -1},
		{"empty string", "", []byte{'a'}, -1},
		{"empty set", "hello", []byte{}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IndexByteSet(tt.s, tt.set)
			if result != tt.expected {
				t.Errorf("IndexByteSet(%q, %v) = %d, want %d", tt.s, tt.set, result, tt.expected)
			}
		})
	}
}

func TestCountByte(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		c        byte
		expected int
	}{
		{"count one", "hello", 'l', 2},
		{"count none", "hello", 'x', 0},
		{"count all", "aaa", 'a', 3},
		{"empty string", "", 'a', 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountByte(tt.s, tt.c)
			if result != tt.expected {
				t.Errorf("CountByte(%q, %c) = %d, want %d", tt.s, tt.c, result, tt.expected)
			}
		})
	}
}

func TestCountByteSet(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		set      []byte
		expected int
	}{
		{"count vowels", "hello world", []byte{'a', 'e', 'i', 'o', 'u'}, 3},
		{"count none", "bcdfg", []byte{'a', 'e', 'i', 'o', 'u'}, 0},
		{"empty string", "", []byte{'a'}, 0},
		{"empty set", "hello", []byte{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountByteSet(tt.s, tt.set)
			if result != tt.expected {
				t.Errorf("CountByteSet(%q, %v) = %d, want %d", tt.s, tt.set, result, tt.expected)
			}
		})
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		c        byte
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'_', true},
		{' ', false},
		{'-', false},
		{'!', false},
	}

	for _, tt := range tests {
		result := IsWordChar(tt.c)
		if result != tt.expected {
			t.Errorf("IsWordChar(%c) = %v, want %v", tt.c, result, tt.expected)
		}
	}
}

func TestIsWhitespace(t *testing.T) {
	tests := []struct {
		c        byte
		expected bool
	}{
		{' ', true},
		{'\t', true},
		{'\n', true},
		{'\r', true},
		{'a', false},
		{'0', false},
	}

	for _, tt := range tests {
		result := IsWhitespace(tt.c)
		if result != tt.expected {
			t.Errorf("IsWhitespace(%c) = %v, want %v", tt.c, result, tt.expected)
		}
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple", "hello world", []string{"hello", "world"}},
		{"multiple spaces", "hello  world", []string{"hello", "world"}},
		{"tabs", "hello\tworld", []string{"hello", "world"}},
		{"newlines", "hello\nworld", []string{"hello", "world"}},
		{"mixed", "hello \t\n world", []string{"hello", "world"}},
		{"empty", "", nil},
		{"single word", "hello", []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitWords(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("SplitWords(%q) returned %d words, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("SplitWords(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		w        string
		expected bool
	}{
		{"exact match", "hello", "hello", true},
		{"word in sentence", "hello world", "hello", true},
		{"substring not word", "helloworld", "hello", false},
		{"word at end", "say hello", "hello", true},
		{"not found", "hello world", "goodbye", false},
		{"empty word", "hello", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsWord(tt.s, tt.w)
			if result != tt.expected {
				t.Errorf("ContainsWord(%q, %q) = %v, want %v", tt.s, tt.w, result, tt.expected)
			}
		})
	}
}

func TestCountBrackets(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		opens  int
		closes int
	}{
		{"balanced", "{hello}", 1, 1},
		{"nested", "{{hello}}", 2, 2},
		{"unbalanced", "{{hello}", 2, 1},
		{"multiple types", "{[<hello>]}", 3, 3},
		{"none", "hello", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opens, closes := CountBrackets(tt.input, DefaultBracketPairs)
			if opens != tt.opens || closes != tt.closes {
				t.Errorf("CountBrackets(%q) = (%d, %d), want (%d, %d)",
					tt.input, opens, closes, tt.opens, tt.closes)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substrs  []string
		expected bool
	}{
		{"found first", "hello world", []string{"hello", "goodbye"}, true},
		{"found second", "hello world", []string{"goodbye", "world"}, true},
		{"not found", "hello world", []string{"foo", "bar"}, false},
		{"empty", "", []string{"hello"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAny(tt.s, tt.substrs)
			if result != tt.expected {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, result, tt.expected)
			}
		})
	}
}

// Benchmarks

func BenchmarkStripANSI(b *testing.B) {
	input := strings.Repeat("\x1b[31mcolored text\x1b[0m ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StripANSI(input)
	}
}

func BenchmarkHasANSI(b *testing.B) {
	input := strings.Repeat("plain text ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HasANSI(input)
	}
}

func BenchmarkCountByte(b *testing.B) {
	input := strings.Repeat("hello world ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountByte(input, 'l')
	}
}

func BenchmarkSplitWords(b *testing.B) {
	input := strings.Repeat("hello world test benchmark ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SplitWords(input)
	}
}
