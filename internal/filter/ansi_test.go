package filter

import (
	"strings"
	"testing"
)

func TestNewANSIFilter(t *testing.T) {
	f := NewANSIFilter()
	if f == nil {
		t.Fatal("NewANSIFilter returned nil")
	}
	if f.Name() != "ansi" {
		t.Errorf("Name() = %q, want 'ansi'", f.Name())
	}
}

func TestANSIFilter_Apply_None(t *testing.T) {
	f := NewANSIFilter()
	input := "plain text"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestANSIFilter_Apply_WithANSI(t *testing.T) {
	f := NewANSIFilter()
	input := "\x1b[32mSUCCESS\x1b[0m: build completed\n\x1b[31mERROR\x1b[0m: test failed"
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved <= 0 {
		t.Error("should save tokens when stripping ANSI")
	}
	// Should strip ANSI codes
	if strings.Contains(output, "\x1b[") {
		t.Error("should strip ANSI escape codes")
	}
	if !strings.Contains(output, "SUCCESS") {
		t.Error("should keep text content")
	}
}

func TestANSIFilter_Apply_NoANSI(t *testing.T) {
	f := NewANSIFilter()
	input := "plain text without codes"
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	// No ANSI = no tokens saved
	if saved != 0 {
		t.Errorf("no ANSI should save 0, got %d", saved)
	}
}

func TestHasANSI(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"\x1b[32mtext\x1b[0m", true},
		{"plain text", false},
		{"\x1b[1mbold", true},
		{"", false},
	}
	for _, tt := range tests {
		got := HasANSI(tt.input)
		if got != tt.want {
			t.Errorf("HasANSI(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
