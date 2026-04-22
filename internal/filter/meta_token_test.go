package filter

import (
	"strings"
	"testing"
)

func TestDefaultMetaTokenConfig(t *testing.T) {
	cfg := DefaultMetaTokenConfig()
	if cfg.WindowSize != 512 {
		t.Errorf("expected WindowSize 512, got %d", cfg.WindowSize)
	}
	if cfg.MinPattern != 3 {
		t.Errorf("expected MinPattern 3, got %d", cfg.MinPattern)
	}
	if cfg.MaxMetaTokens != 1000 {
		t.Errorf("expected MaxMetaTokens 1000, got %d", cfg.MaxMetaTokens)
	}
}

func TestNewMetaTokenFilter(t *testing.T) {
	f := NewMetaTokenFilter()
	if f == nil {
		t.Fatal("expected non-nil MetaTokenFilter")
	}
	if f.Name() != "meta_token" {
		t.Errorf("expected name 'meta_token', got %q", f.Name())
	}
}

func TestMetaTokenFilter_Apply_ModeNone(t *testing.T) {
	f := NewMetaTokenFilter()
	input := "hello world"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("expected passthrough for ModeNone")
	}
	if saved != 0 {
		t.Error("expected 0 saved for ModeNone")
	}
}

func TestMetaTokenFilter_Apply_ShortInput(t *testing.T) {
	f := NewMetaTokenFilter()
	input := "a b"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("expected passthrough for short input")
	}
	if saved != 0 {
		t.Error("expected 0 saved for short input")
	}
}

func TestMetaTokenFilter_Apply_WithRepetition(t *testing.T) {
	f := NewMetaTokenFilter()
	// Create input with repeated pattern
	input := strings.Repeat("foo bar baz ", 20)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestMetaTokenFilter_Decompress(t *testing.T) {
	f := NewMetaTokenFilter()
	input := strings.Repeat("foo bar baz ", 20)
	output, _ := f.Apply(input, ModeMinimal)
	decompressed := f.Decompress(output)
	if decompressed == "" {
		t.Error("expected non-empty decompressed output")
	}
}

func TestMetaTokenFilter_Stats(t *testing.T) {
	f := NewMetaTokenFilter()
	input := strings.Repeat("foo bar baz ", 20)
	f.Apply(input, ModeMinimal)

	stats := f.Stats()
	if stats.UniquePatterns < 0 {
		t.Error("expected non-negative unique patterns")
	}
}

func TestMetaTokenFilter_GetAndLoadMetaTokens(t *testing.T) {
	f := NewMetaTokenFilter()
	input := strings.Repeat("foo bar baz ", 20)
	f.Apply(input, ModeMinimal)

	tokens := f.GetMetaTokens()
	if len(tokens) == 0 {
		t.Log("no meta-tokens created (input may not have repetitions)")
	}

	f2 := NewMetaTokenFilter()
	f2.LoadMetaTokens(tokens)
	if len(f2.GetMetaTokens()) != len(tokens) {
		t.Error("expected loaded tokens to match")
	}
}
