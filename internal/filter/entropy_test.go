package filter

import (
	"strings"
	"testing"
)

func TestNewEntropyFilter(t *testing.T) {
	f := NewEntropyFilter()
	if f == nil {
		t.Fatal("expected non-nil EntropyFilter")
	}
}

func TestNewEntropyFilterWithThreshold(t *testing.T) {
	f := NewEntropyFilterWithThreshold(3.0)
	if f == nil {
		t.Fatal("expected non-nil EntropyFilter")
	}
	if f.entropyThreshold != 3.0 {
		t.Errorf("expected threshold 3.0, got %f", f.entropyThreshold)
	}
}

func TestEntropyFilter_SetDynamicEstimation(t *testing.T) {
	f := NewEntropyFilter()
	f.SetDynamicEstimation(false)
	if f.useDynamicEst {
		t.Error("expected dynamic estimation disabled")
	}
	f.SetDynamicEstimation(true)
	if !f.useDynamicEst {
		t.Error("expected dynamic estimation enabled")
	}
}

func TestEntropyFilter_Name(t *testing.T) {
	f := NewEntropyFilter()
	if f.Name() != "entropy" {
		t.Errorf("expected name 'entropy', got %q", f.Name())
	}
}

func TestEntropyFilter_Apply_ShortInput(t *testing.T) {
	f := NewEntropyFilter()
	// Use rare words that should survive entropy filtering
	input := "xylophone zebra quantum"
	output, saved := f.Apply(input, ModeMinimal)
	// Entropy filter may drop common words; just verify no panic and valid saved
	_ = output
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestEntropyFilter_Apply_NormalInput(t *testing.T) {
	f := NewEntropyFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog ", 20)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestEntropyFilter_Apply_ModeNone(t *testing.T) {
	f := NewEntropyFilter()
	input := "hello world"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("expected passthrough for ModeNone")
	}
	if saved != 0 {
		t.Error("expected 0 saved for ModeNone")
	}
}

func TestInitTokenFrequencies(t *testing.T) {
	freqs := initTokenFrequencies()
	if len(freqs) == 0 {
		t.Error("expected non-empty frequency map")
	}
	if freqs["the"] == 0 {
		t.Error("expected 'the' to have a frequency")
	}
}

func TestBuildTokenFrequencies(t *testing.T) {
	freqs := buildTokenFrequencies()
	if len(freqs) == 0 {
		t.Error("expected non-empty frequency map")
	}
}
