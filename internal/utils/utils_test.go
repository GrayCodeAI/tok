package utils

import "testing"

func TestMin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		if got := Min(tt.a, tt.b); got != tt.want {
			t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
	}
	for _, tt := range tests {
		if got := Max(tt.a, tt.b); got != tt.want {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestClamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x, min, max, want int
	}{
		{5, 0, 10, 5},
		{15, 0, 10, 10},
		{-5, 0, 10, 0},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, tt := range tests {
		if got := Clamp(tt.x, tt.min, tt.max); got != tt.want {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.x, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestAbs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x, want int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}
	for _, tt := range tests {
		if got := Abs(tt.x); got != tt.want {
			t.Errorf("Abs(%d) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func TestShortenPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path     string
		maxLen   int
		contains string
	}{
		{"/short/path", 20, "/short/path"},
		{"/very/long/path/that/should/be/truncated/file.go", 20, "file.go"},
	}
	for _, tt := range tests {
		got := ShortenPath(tt.path, tt.maxLen)
		if len(got) > tt.maxLen+1 {
			t.Errorf("ShortenPath(%q, %d) = %q (len %d), max len exceeded", tt.path, tt.maxLen, got, len(got))
		}
		if !containsStr(got, tt.contains) {
			t.Errorf("ShortenPath(%q, %d) = %q, want to contain %q", tt.path, tt.maxLen, got, tt.contains)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{1099511627776, "1.0T"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0ms"},
		{500, "500ms"},
		{1000, "1.0s"},
		{60000, "1m 0s"},
		{3600000, "1h 0m"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.ms)
		if got == "" {
			t.Errorf("FormatDuration(%d) returned empty", tt.ms)
		}
	}
}

func TestFormatTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1.0K"},
		{1500000, "1.5M"},
		{2000000000, "2.0B"},
	}
	for _, tt := range tests {
		got := FormatTokens(tt.n)
		if got != tt.want {
			t.Errorf("FormatTokens(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestGetModelFamily(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, want string
	}{
		{"claude-sonnet-4", "claude"},
		{"gpt-4o", "gpt"},
		{"gpt-3.5-turbo", "gpt"},
		{"o1-mini", "gpt"},
		{"o3-mini", "gpt"},
		{"gemini-pro", "gemini"},
		{"llama-3-8b", "llama"},
		{"qwen-2.5-72b", "qwen"},
		{"deepseek-coder-v2", "deepseek"},
		{"mistral-large", "mistral"},
		{"unknown-model", "other"},
		{"", ""},
	}
	for _, tt := range tests {
		got := GetModelFamily(tt.name)
		if got != tt.want {
			t.Errorf("GetModelFamily(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatTokens64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000000, "1.0M"},
		{1000000000, "1.0B"},
	}
	for _, tt := range tests {
		got := FormatTokens64(tt.n)
		if got != tt.want {
			t.Errorf("FormatTokens64(%d) = %q, want %q", tt.n, got, tt.want)
		}
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
