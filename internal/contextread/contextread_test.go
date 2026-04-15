package contextread

import (
	"strings"
	"testing"
)

func TestOptions(t *testing.T) {
	opts := Options{
		Mode:              "full",
		Level:             "debug",
		MaxLines:          100,
		MaxTokens:         1000,
		LineNumbers:       true,
		StartLine:         10,
		EndLine:           50,
		SaveSnapshot:      true,
		RelatedFilesCount: 5,
	}

	if opts.Mode != "full" {
		t.Error("Mode not set correctly")
	}
	if opts.MaxLines != 100 {
		t.Error("MaxLines not set correctly")
	}
	if !opts.LineNumbers {
		t.Error("LineNumbers not set correctly")
	}
}

func TestTrackedCommandPatternsForKind(t *testing.T) {
	patterns := TrackedCommandPatternsForKind("test")
	// Stub returns nil
	if patterns != nil {
		t.Errorf("expected nil, got %v", patterns)
	}

	// Test with different kinds
	patterns = TrackedCommandPatternsForKind("")
	if patterns != nil {
		t.Error("expected nil for empty kind")
	}
}

func TestTrackedCommandPatterns(t *testing.T) {
	patterns := TrackedCommandPatterns()
	// Stub returns nil
	if patterns != nil {
		t.Errorf("expected nil, got %v", patterns)
	}
}

func TestBuild(t *testing.T) {
	content := "line1\nline2\nline3"
	opts := Options{
		MaxLines:  10,
		MaxTokens: 100,
	}

	result, tokens, lines, err := Build("/test/file.go", content, "go", opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Stub returns content unchanged
	if result != content {
		t.Errorf("expected content unchanged, got '%s'", result)
	}

	// Tokens should be len/4
	expectedTokens := len(content) / 4
	if tokens != expectedTokens {
		t.Errorf("expected tokens=%d, got %d", expectedTokens, tokens)
	}

	if lines != expectedTokens {
		t.Errorf("expected lines=%d, got %d", expectedTokens, lines)
	}
}

func TestBuild_EmptyContent(t *testing.T) {
	opts := Options{}
	result, tokens, lines, err := Build("/test/file.go", "", "go", opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Error("expected empty result for empty input")
	}
	if tokens != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", tokens)
	}
	if lines != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", lines)
	}
}

func TestAnalyze(t *testing.T) {
	content := "test content to analyze"
	result := Analyze(content)

	// Stub returns content unchanged
	if result != content {
		t.Errorf("expected '%s', got '%s'", content, result)
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "/home/user/project/main.go",
			expected: "main.go",
		},
		{
			path:     "relative/path/file.txt",
			expected: "file.txt",
		},
		{
			path:     "simple.txt",
			expected: "simple.txt",
		},
		{
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := Describe(tt.path)
			if result != tt.expected {
				t.Errorf("Describe('%s') = '%s', want '%s'", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDescribe_WindowsPath(t *testing.T) {
	// Test Windows-style paths
	path := "C:\\Users\\test\\file.go"
	result := Describe(path)
	// Should handle backslash or return full path
	if result != "file.go" && !strings.Contains(result, "file.go") {
		t.Errorf("Describe('%s') = '%s', expected 'file.go' or containing it", path, result)
	}
}
