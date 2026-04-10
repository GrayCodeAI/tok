package filter

import (
	"strings"
	"testing"
)

func TestExtractivePrefilter_NoOpForSmallInput(t *testing.T) {
	f := NewExtractivePrefilter(ExtractivePrefilterConfig{MaxLines: 10})
	input := "line1\nline2\nline3"

	out, saved := f.Apply(input)
	if out != input {
		t.Fatalf("expected passthrough for small input")
	}
	if saved != 0 {
		t.Fatalf("expected zero saved tokens, got %d", saved)
	}
}

func TestExtractivePrefilter_KeepSignalLines(t *testing.T) {
	f := NewExtractivePrefilter(ExtractivePrefilterConfig{
		MaxLines:    5,
		HeadLines:   1,
		TailLines:   1,
		SignalLines: 3,
	})
	input := strings.Join([]string{
		"header",
		"info line",
		"WARNING: low disk",
		"more info",
		"ERROR: build failed",
		"tail",
	}, "\n")

	out, _ := f.Apply(input)
	if !strings.Contains(out, "WARNING: low disk") {
		t.Fatalf("expected warning line to be preserved")
	}
	if !strings.Contains(out, "ERROR: build failed") {
		t.Fatalf("expected error line to be preserved")
	}
}

func TestExtractivePrefilter_ReducesLargeInput(t *testing.T) {
	f := NewExtractivePrefilter(ExtractivePrefilterConfig{
		MaxLines:    20,
		HeadLines:   3,
		TailLines:   3,
		SignalLines: 2,
	})

	lines := make([]string, 0, 120)
	for i := 0; i < 120; i++ {
		lines = append(lines, "line content")
	}
	lines[60] = "panic: nil pointer dereference"
	input := strings.Join(lines, "\n")

	out, saved := f.Apply(input)
	if saved <= 0 {
		t.Fatalf("expected positive token reduction")
	}
	if !strings.Contains(out, "panic: nil pointer dereference") {
		t.Fatalf("expected panic line to be retained")
	}
	if !strings.Contains(out, "[... omitted by extractive prefilter ...]") {
		t.Fatalf("expected omission marker")
	}
}
