package filtercmd

import (
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
	}{
		{"short", "hello", 80},
		{"long", "This is a long line that should be wrapped at a certain width", 20},
		{"empty", "", 20},
		{"single word", "word", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.input, tt.width)
			if got == "" && tt.input != "" {
				t.Error("wrapText should return non-empty for non-empty input")
			}
			// Verify the function uses the "║   " prefix for wrapped lines
			if strings.Contains(tt.input, " ") && tt.width < len(tt.input) {
				if !strings.Contains(got, "║   ") {
					t.Errorf("wrapText should use '║   ' prefix for wrapped content, got: %q", got)
				}
			}
		})
	}
}
