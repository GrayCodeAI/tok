package filter

import (
	"strings"
	"testing"
)

func TestNewEvaluatorHeadsFilter(t *testing.T) {
	f := NewEvaluatorHeadsFilter()
	if f == nil {
		t.Fatal("NewEvaluatorHeadsFilter returned nil")
	}
	if f.Name() != "evaluator_heads" {
		t.Errorf("Name() = %q, want 'evaluator_heads'", f.Name())
	}
}

func TestEvaluatorHeadsFilter_Apply_None(t *testing.T) {
	f := NewEvaluatorHeadsFilter()
	input := "some content"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestEvaluatorHeadsFilter_Apply_Minimal(t *testing.T) {
	f := NewEvaluatorHeadsFilter()
	lines := make([]string, 50)
	for i := range lines {
		if i%10 == 0 {
			lines[i] = "ERROR: something failed"
		} else {
			lines[i] = "debug: verbose output line " + string(rune('0'+i%10))
		}
	}
	input := strings.Join(lines, "\n")
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved <= 0 {
		t.Error("should save tokens on 50 lines")
	}
}

func TestEvaluatorHeadsFilter_PreservesErrors(t *testing.T) {
	f := NewEvaluatorHeadsFilter()
	input := strings.Repeat("debug: info\n", 30) +
		"ERROR: critical failure\n" +
		strings.Repeat("debug: info\n", 30)
	output, _ := f.Apply(input, ModeMinimal)
	if !strings.Contains(output, "ERROR") {
		t.Error("should preserve ERROR lines")
	}
}

func TestEvaluatorHeadsFilter_PreservesWarnings(t *testing.T) {
	f := NewEvaluatorHeadsFilter()
	input := strings.Repeat("trace: detail\n", 30) +
		"WARNING: deprecated function used\n" +
		strings.Repeat("trace: detail\n", 30)
	output, _ := f.Apply(input, ModeMinimal)
	if !strings.Contains(output, "WARNING") {
		t.Error("should preserve WARNING lines")
	}
}
