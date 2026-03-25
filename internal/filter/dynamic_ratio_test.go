package filter

import (
	"testing"
)

func TestNewDynamicRatioFilter(t *testing.T) {
	f := NewDynamicRatioFilter()
	if f == nil {
		t.Fatal("NewDynamicRatioFilter returned nil")
	}
	if f.Name() != "dynamic_ratio" {
		t.Errorf("Name() = %q, want 'dynamic_ratio'", f.Name())
	}
}

func TestDefaultDynamicRatioConfig(t *testing.T) {
	cfg := DefaultDynamicRatioConfig()
	if !cfg.Enabled {
		t.Error("default config should be enabled")
	}
	if cfg.MinComplexity != 0.2 {
		t.Errorf("MinComplexity = %f, want 0.2", cfg.MinComplexity)
	}
	if cfg.MaxComplexity != 0.8 {
		t.Errorf("MaxComplexity = %f, want 0.8", cfg.MaxComplexity)
	}
	if cfg.BaseBudgetRatio != 1.0 {
		t.Errorf("BaseBudgetRatio = %f, want 1.0", cfg.BaseBudgetRatio)
	}
}

func TestDynamicRatioFilter_Apply_None(t *testing.T) {
	f := NewDynamicRatioFilter()
	input := "some content"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestDynamicRatioFilter_Apply_ShortInput(t *testing.T) {
	f := NewDynamicRatioFilter()
	input := "short"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("short input should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("short input should save 0, got %d", saved)
	}
}

func TestDynamicRatioFilter_Disabled(t *testing.T) {
	f := &DynamicRatioFilter{config: DynamicRatioConfig{Enabled: false}}
	input := "some content that is long enough to be processed"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("disabled filter should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("disabled filter should save 0, got %d", saved)
	}
}
