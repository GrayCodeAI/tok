package filter

import (
	"strings"
	"testing"
)

func TestNewNgramAbbreviator(t *testing.T) {
	a := NewNgramAbbreviator()
	if a == nil {
		t.Fatal("NewNgramAbbreviator returned nil")
	}
	if a.Name() != "ngram" {
		t.Errorf("Name() = %q, want 'ngram'", a.Name())
	}
}

func TestNgramAbbreviator_Apply_None(t *testing.T) {
	a := NewNgramAbbreviator()
	input := "the quick brown fox"
	output, saved := a.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestNgramAbbreviator_Apply_ShortInput(t *testing.T) {
	a := NewNgramAbbreviator()
	input := "short"
	output, saved := a.Apply(input, ModeMinimal)
	if output != input {
		t.Error("short input should not be modified")
	}
	if saved != 0 {
		t.Errorf("short input should save 0, got %d", saved)
	}
}

func TestNgramAbbreviator_Apply_Minimal(t *testing.T) {
	a := NewNgramAbbreviator()
	input := strings.Repeat("function return const var import export\n", 30)
	output, saved := a.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestNgramAbbreviator_Apply_Aggressive(t *testing.T) {
	a := NewNgramAbbreviator()
	input := strings.Repeat("Successfully completed the function return value\n", 30)
	output, saved := a.Apply(input, ModeAggressive)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestNgramAbbreviator_DetectCodeContext(t *testing.T) {
	a := NewNgramAbbreviator()

	tests := []struct {
		input string
		want  bool
	}{
		{"func main() { return 42 }", true},
		{"def hello(): pass", true},
		{"class Foo { }", true},
		{"import os", true},
		{"package main", true},
		{"const x = 1", true},
		{"just some plain text without code", false},
		{"hello world this is a normal sentence", false},
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(30, len(tt.input))], func(t *testing.T) {
			got := a.detectCodeContext(tt.input)
			if got != tt.want {
				t.Errorf("detectCodeContext(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNgramAbbreviator_GetAbbreviationLegend(t *testing.T) {
	a := NewNgramAbbreviator()
	legend := a.GetAbbreviationLegend()
	if !strings.HasPrefix(legend, "Abbreviations: ") {
		t.Errorf("legend should start with 'Abbreviations: ', got %q", legend)
	}
}

func TestNgramAbbreviator_ReplaceWord(t *testing.T) {
	a := NewNgramAbbreviator()

	tests := []struct {
		input   string
		pattern string
		repl    string
		want    string
	}{
		{"function test", "function", "fn", "fn test"},
		{"functionality test", "function", "fn", "functionality test"}, // should not match inside word
		{"the function", "function", "fn", "the fn"},
		{"FUNCTION test", "function", "fn", "fn test"}, // case-insensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := a.replaceWord(tt.input, tt.pattern, tt.repl)
			if got != tt.want {
				t.Errorf("replaceWord(%q, %q, %q) = %q, want %q", tt.input, tt.pattern, tt.repl, got, tt.want)
			}
		})
	}
}

func TestToLowerByte(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'a', 'a'},
		{'z', 'z'},
		{'0', '0'},
		{'_', '_'},
	}

	for _, tt := range tests {
		got := toLowerByte(tt.input)
		if got != tt.want {
			t.Errorf("toLowerByte(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func BenchmarkNgramAbbreviator_Apply(b *testing.B) {
	a := NewNgramAbbreviator()
	input := strings.Repeat("function return const var import export package module class\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Apply(input, ModeMinimal)
	}
}
