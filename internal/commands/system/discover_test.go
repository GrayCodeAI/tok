package system

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a very long string", 10, "this is..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"hello world", 5, "he..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestTruncateEmptyString(t *testing.T) {
	got := truncate("", 10)
	if got != "" {
		t.Errorf("truncate(\"\", 10) = %q, want \"\"", got)
	}
}

func TestTruncateExactLength(t *testing.T) {
	got := truncate("12345", 5)
	if got != "12345" {
		t.Errorf("truncate(\"12345\", 5) = %q, want \"12345\"", got)
	}
}

func TestTruncateUnicode(t *testing.T) {
	// Test with unicode characters
	got := truncate("hello 世界 this is long", 10)
	if len(got) > 10 {
		t.Errorf("truncate() result too long: %d chars", len(got))
	}
}
