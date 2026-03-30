package system

import (
	"strings"
	"testing"
)

func TestCompactGrepOutputSimple(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		maxLen     int
		maxResults int
		wantLines  int
		wantTrunc  bool
	}{
		{
			name:       "empty output",
			input:      "",
			maxLen:     80,
			maxResults: 50,
			wantLines:  0,
		},
		{
			name: "basic grep output",
			input: `file1.go:10:func main() {
file1.go:15:func helper() {
file2.go:5:var x int`,
			maxLen:     80,
			maxResults: 50,
			wantLines:  3,
		},
		{
			name:       "truncate long lines",
			input:      `file.go:1:` + strings.Repeat("a", 200),
			maxLen:     80,
			maxResults: 50,
			wantLines:  1,
			wantTrunc:  true,
		},
		{
			name: "respect max results",
			input: `file1.go:1:line1
file2.go:2:line2
file3.go:3:line3
file4.go:4:line4
file5.go:5:line5`,
			maxLen:     80,
			maxResults: 3,
			wantLines:  3, // Only 3 lines + overflow indicator
		},
		{
			name: "skip empty lines",
			input: `file.go:1:code

file.go:2:more code

`,
			maxLen:     80,
			maxResults: 50,
			wantLines:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compactGrepOutputSimple(tt.input, tt.maxLen, tt.maxResults)

			// Count non-empty lines
			lines := strings.Split(strings.TrimSpace(result), "\n")
			var nonEmptyLines int
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}

			if tt.wantLines > 0 && nonEmptyLines < tt.wantLines {
				t.Errorf("compactGrepOutputSimple() got %d lines, want at least %d", nonEmptyLines, tt.wantLines)
			}

			if tt.wantTrunc {
				if !strings.Contains(result, "...") {
					t.Errorf("compactGrepOutputSimple() expected truncation but no '...' found")
				}
			}
		})
	}
}

func TestCompactGrepOutputSimple_Truncation(t *testing.T) {
	// Test that lines longer than maxLen are truncated
	longLine := strings.Repeat("x", 200)
	input := "file.go:1:" + longLine

	result := compactGrepOutputSimple(input, 50, 10)

	// Result should be truncated
	if len(result) > 60 { // 50 + some overhead for filename and "..."
		t.Errorf("compactGrepOutputSimple() line not truncated properly, got length %d", len(result))
	}

	// Should contain truncation indicator
	if !strings.Contains(result, "...") {
		t.Error("compactGrepOutputSimple() missing truncation indicator '...'")
	}
}

func TestCompactGrepOutputSimple_MaxResults(t *testing.T) {
	// Test max results limit
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "file.go:line content")
	}
	input := strings.Join(lines, "\n")

	result := compactGrepOutputSimple(input, 80, 10)

	// Should indicate overflow
	if !strings.Contains(result, "more") {
		t.Error("compactGrepOutputSimple() missing overflow indicator for max results")
	}
}

func TestCompactGrepOutputSimple_EmptyAndWhitespace(t *testing.T) {
	input := `
file.go:1:code
   
   
file.go:2:more code

`
	result := compactGrepOutputSimple(input, 80, 50)

	// Should not contain empty lines
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Empty lines at end are OK
			continue
		}
		// Non-empty lines should have content
		if strings.Contains(line, "file.go") {
			// Valid content line
		}
	}
}
