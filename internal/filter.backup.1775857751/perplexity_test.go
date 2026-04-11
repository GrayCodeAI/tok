package filter

import (
	"strings"
	"testing"
)

func TestNewPerplexityFilter(t *testing.T) {
	f := NewPerplexityFilter()
	if f == nil {
		t.Fatal("NewPerplexityFilter returned nil")
	}
	if f.Name() != "perplexity" {
		t.Errorf("Name() = %q, want 'perplexity'", f.Name())
	}
}

func TestPerplexityFilter_Apply_None(t *testing.T) {
	f := NewPerplexityFilter()
	input := "the quick brown fox jumps over the lazy dog"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestPerplexityFilter_Apply_Minimal(t *testing.T) {
	f := NewPerplexityFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog\n", 50)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestPerplexityFilter_Apply_Aggressive(t *testing.T) {
	f := NewPerplexityFilter()
	input := strings.Repeat("the a an is are was were in on at to for of\n", 100)
	output, saved := f.Apply(input, ModeAggressive)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestPerplexityFilter_ShortLine(t *testing.T) {
	f := NewPerplexityFilter()
	input := "short"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("short input should not be modified")
	}
	if saved != 0 {
		t.Errorf("short input should save 0, got %d", saved)
	}
}

func TestPerplexityFilter_SetTargetRatio(t *testing.T) {
	f := NewPerplexityFilter()
	f.SetTargetRatio(0.5)
	if f.targetRatio != 0.5 {
		t.Errorf("targetRatio = %f, want 0.5", f.targetRatio)
	}
}

func TestPerplexityFilter_SetIterations(t *testing.T) {
	f := NewPerplexityFilter()
	f.SetIterations(5)
	if f.iterationSteps != 5 {
		t.Errorf("iterationSteps = %d, want 5", f.iterationSteps)
	}
}

func TestIsCodeToken(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"foo.bar", true},
		{"my_func", true},
		{"$var", true},
		{"@decorator", true},
		{"CamelCase", true},
		{"hello", false},
		{"123", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := isCodeToken(tt.word)
			if got != tt.want {
				t.Errorf("isCodeToken(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"123", true},
		{"3.14", true},
		{"-42", true},
		{"+100", true},
		{"abc", false},
		{"12a", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := isNumeric(tt.word)
			if got != tt.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func BenchmarkPerplexityFilter_Apply(b *testing.B) {
	f := NewPerplexityFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog\n", 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeMinimal)
	}
}
