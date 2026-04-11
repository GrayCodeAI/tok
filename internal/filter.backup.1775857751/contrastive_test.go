package filter

import (
	"strings"
	"testing"
)

func TestNewContrastiveFilter(t *testing.T) {
	f := NewContrastiveFilter("error handling")
	if f == nil {
		t.Fatal("NewContrastiveFilter returned nil")
	}
	if f.Name() != "contrastive" {
		t.Errorf("Name() = %q, want 'contrastive'", f.Name())
	}
}

func TestContrastiveFilter_Apply_None(t *testing.T) {
	f := NewContrastiveFilter("query")
	input := "some content"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestContrastiveFilter_EmptyQuestion(t *testing.T) {
	f := NewContrastiveFilter("")
	input := "some content"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("empty question should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("empty question should save 0, got %d", saved)
	}
}

func TestContrastiveFilter_Minimal(t *testing.T) {
	f := NewContrastiveFilter("error handling")
	lines := make([]string, 50)
	for i := range lines {
		if i == 10 {
			lines[i] = "ERROR: failed to handle connection timeout"
		} else if i == 20 {
			lines[i] = "error handler caught exception in main loop"
		} else {
			lines[i] = "debug: processing request number " + string(rune('0'+i%10))
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
	// Should keep ERROR lines since they're relevant to "error handling"
	if !strings.Contains(output, "ERROR") {
		t.Error("should preserve error-related lines")
	}
}

func TestContrastiveFilter_Aggressive(t *testing.T) {
	f := NewContrastiveFilter("test query")
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "information line " + string(rune('0'+i%10))
	}
	input := strings.Join(lines, "\n")
	_, minimalSaved := f.Apply(input, ModeMinimal)
	_, aggressiveSaved := f.Apply(input, ModeAggressive)
	if aggressiveSaved < minimalSaved {
		t.Errorf("aggressive saved %d should be >= minimal saved %d", aggressiveSaved, minimalSaved)
	}
}

func TestContrastiveFilter_ShortInput(t *testing.T) {
	f := NewContrastiveFilter("query")
	input := "line1\nline2\nline3"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("short input (3 lines) should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("short input should save 0, got %d", saved)
	}
}

func TestContrastiveFilter_SetQuestion(t *testing.T) {
	f := NewContrastiveFilter("initial")
	f.SetQuestion("updated query")
	if f.question != "updated query" {
		t.Errorf("question = %q, want 'updated query'", f.question)
	}
}
