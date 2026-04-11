package filter

import (
	"strings"
	"testing"
)

func TestNewEntropyFilter(t *testing.T) {
	f := NewEntropyFilter()
	if f == nil {
		t.Fatal("NewEntropyFilter returned nil")
	}
	if f.Name() != "entropy" {
		t.Errorf("Name() = %q, want 'entropy'", f.Name())
	}
}

func TestNewEntropyFilterWithThreshold(t *testing.T) {
	f := NewEntropyFilterWithThreshold(5.0)
	if f.entropyThreshold != 5.0 {
		t.Errorf("threshold = %f, want 5.0", f.entropyThreshold)
	}
}

func TestEntropyFilter_Apply_None(t *testing.T) {
	f := NewEntropyFilter()
	input := "the quick brown fox jumps over the lazy dog"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestEntropyFilter_Apply_Minimal(t *testing.T) {
	f := NewEntropyFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog\n", 50)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestEntropyFilter_Apply_Aggressive(t *testing.T) {
	f := NewEntropyFilter()
	input := strings.Repeat("the a an is are was were in on at to for of\n", 100)
	output, saved := f.Apply(input, ModeAggressive)
	if output == "" {
		t.Error("output should not be empty")
	}
	// Aggressive should save more than minimal
	_, minimalSaved := f.Apply(input, ModeMinimal)
	if saved < minimalSaved {
		t.Errorf("aggressive saved %d should be >= minimal saved %d", saved, minimalSaved)
	}
}

func TestEntropyFilter_ShortInput(t *testing.T) {
	f := NewEntropyFilter()
	input := "hello world test"
	_, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestEntropyFilter_SetThreshold(t *testing.T) {
	f := NewEntropyFilter()
	f.SetThreshold(10.0)
	if f.entropyThreshold != 10.0 {
		t.Errorf("threshold = %f, want 10.0", f.entropyThreshold)
	}
}

func TestEntropyFilter_DynamicEstimation(t *testing.T) {
	f := NewEntropyFilter()
	f.SetDynamicEstimation(true)
	if !f.useDynamicEst {
		t.Error("dynamic estimation should be enabled")
	}
	f.SetDynamicEstimation(false)
	if f.useDynamicEst {
		t.Error("dynamic estimation should be disabled")
	}
}

func TestEntropyFilter_CodeTokens(t *testing.T) {
	f := NewEntropyFilter()
	codeTokens := []string{"func", "return", "import", "package", "def", "class", "const", "let"}
	for _, token := range codeTokens {
		if _, exists := f.frequencies[token]; !exists {
			t.Errorf("token %q should be in frequency table", token)
		}
	}
}

func TestEntropyFilter_MixedContent(t *testing.T) {
	f := NewEntropyFilter()
	input := `package main

import "fmt"

func main() {
	// This is a comment with common words like the and a
	fmt.Println("hello world")
	x := 42
	if x > 0 {
		return
	}
}
`
	output, _ := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	// Should preserve code structure
	if !strings.Contains(output, "func main") {
		t.Error("should preserve function signature")
	}
	if !strings.Contains(output, "import") {
		t.Error("should preserve import")
	}
}
