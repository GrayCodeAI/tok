package filter

import (
	"strings"
	"testing"
)

func TestNewPerplexityFilter(t *testing.T) {
	f := NewPerplexityFilter()
	if f == nil {
		t.Fatal("expected non-nil PerplexityFilter")
	}
	if f.targetRatio != 0.3 {
		t.Errorf("expected targetRatio 0.3, got %f", f.targetRatio)
	}
}

func TestPerplexityFilter_Name(t *testing.T) {
	f := NewPerplexityFilter()
	if f.Name() != "perplexity" {
		t.Errorf("expected name 'perplexity', got %q", f.Name())
	}
}

func TestPerplexityFilter_Apply_ModeNone(t *testing.T) {
	f := NewPerplexityFilter()
	input := "hello world"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("expected passthrough for ModeNone")
	}
	if saved != 0 {
		t.Error("expected 0 saved for ModeNone")
	}
}

func TestPerplexityFilter_Apply_NormalInput(t *testing.T) {
	f := NewPerplexityFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog ", 20)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestPerplexityFilter_pruneLine(t *testing.T) {
	f := NewPerplexityFilter()
	line := "this is a test line with many words for perplexity pruning"
	result := f.pruneLine(line, ModeMinimal)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestPerplexityFilter_pruneLine_Short(t *testing.T) {
	f := NewPerplexityFilter()
	line := "hi"
	result := f.pruneLine(line, ModeMinimal)
	if result != line {
		t.Error("expected passthrough for short line")
	}
}
