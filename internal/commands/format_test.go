package commands

import (
	"testing"
)

func TestDetectFormatter(t *testing.T) {
	// Note: These tests depend on the current working directory
	// In a real test environment, you'd use temp directories
	result := detectFormatter()
	// Just verify it returns a valid formatter
	validFormatters := []string{"prettier", "black", "ruff", "biome", "gofmt"}
	found := false
	for _, f := range validFormatters {
		if result == f {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("detectFormatter() = %q, expected one of %v", result, validFormatters)
	}
}

func TestFilterBlackOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		contains string
	}{
		{
			name:     "all formatted",
			output:   "All done! ✨ 🍰 ✨\n5 files left unchanged.",
			contains: "All files formatted",
		},
		{
			name: "needs formatting",
			output: `would reformat: src/main.py
would reformat: tests/test_utils.py
Oh no! 💥 💔 💥
2 files would be reformatted, 3 files would be left unchanged.`,
			contains: "2 files need formatting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterBlackOutput(tt.output)
			if !containsStr(result, tt.contains) {
				t.Errorf("filterBlackOutput() = %q, should contain %q", result, tt.contains)
			}
		})
	}
}

func TestFilterRuffFormatOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		contains string
	}{
		{
			name:     "all formatted",
			output:   "5 files left unchanged",
			contains: "All files formatted",
		},
		{
			name: "needs formatting",
			output: `would reformat: src/main.py
would reformat: src/lib.py
2 files would be reformatted`,
			contains: "2 files need formatting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRuffFormatOutput(tt.output)
			if !containsStr(result, tt.contains) {
				t.Errorf("filterRuffFormatOutput() = %q, should contain %q", result, tt.contains)
			}
		})
	}
}

func TestFilterPrettierOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		contains string
	}{
		{
			name:     "all formatted",
			output:   "All matched files use Prettier code style!",
			contains: "All files formatted",
		},
		{
			name:     "files need formatting",
			output:   "src/main.ts\nsrc/lib.ts\nChecking formatting...\n[warn] src/main.ts\n[warn] src/lib.ts",
			contains: "need formatting",
		},
		{
			name:     "empty output",
			output:   "",
			contains: "All files formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPrettierOutput(tt.output)
			if !containsStr(result, tt.contains) {
				t.Errorf("filterPrettierOutput() = %q, should contain %q", result, tt.contains)
			}
		})
	}
}

func TestCompactPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/project/src/main.go", "src/main.go"},
		{"/home/user/project/lib/utils.py", "lib/utils.py"},
		{"/home/user/project/tests/test.go", "tests/test.go"},
		{"main.go", "main.go"},
		{"/main.go", "main.go"},
		{"C:\\Users\\project\\src\\main.go", "src/main.go"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := compactPath(tt.path)
			if result != tt.expected {
				t.Errorf("compactPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
