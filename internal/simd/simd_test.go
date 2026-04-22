package simd

import (
	"strings"
	"testing"
)

func TestFastHasANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"plain text", "hello world", false},
		{"ansi escape", "\x1b[31mred\x1b[0m", true},
		{"long plain", strings.Repeat("a", 256), false},
		{"long ansi", strings.Repeat("a", 200) + "\x1b[0m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FastHasANSI(tt.input)
			if got != tt.want {
				t.Errorf("FastHasANSI(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mred\x1b[0m text"
	want := "red text"
	got := StripANSI(input)
	if got != want {
		t.Errorf("StripANSI(%q) = %q, want %q", input, got, want)
	}

	// No ANSI should be no-op
	plain := "plain text"
	if StripANSI(plain) != plain {
		t.Error("StripANSI on plain text should return input unchanged")
	}
}
