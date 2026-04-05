package utils

import "testing"

func TestMin(t *testing.T) {
	if Min(1, 2) != 1 {
		t.Error("Min(1,2) should be 1")
	}
	if Min(5, 3) != 3 {
		t.Error("Min(5,3) should be 3")
	}
	if Min(0, 0) != 0 {
		t.Error("Min(0,0) should be 0")
	}
}

func TestMax(t *testing.T) {
	if Max(1, 2) != 2 {
		t.Error("Max(1,2) should be 2")
	}
	if Max(5, 3) != 5 {
		t.Error("Max(5,3) should be 5")
	}
	if Max(0, 0) != 0 {
		t.Error("Max(0,0) should be 0")
	}
}

func TestAbs(t *testing.T) {
	if Abs(-5) != 5 {
		t.Error("Abs(-5) should be 5")
	}
	if Abs(5) != 5 {
		t.Error("Abs(5) should be 5")
	}
	if Abs(0) != 0 {
		t.Error("Abs(0) should be 0")
	}
}

func TestClamp(t *testing.T) {
	if Clamp(5, 0, 10) != 5 {
		t.Error("Clamp(5, 0, 10) should be 5")
	}
	if Clamp(-1, 0, 10) != 0 {
		t.Error("Clamp(-1, 0, 10) should be 0")
	}
	if Clamp(20, 0, 10) != 10 {
		t.Error("Clamp(20, 0, 10) should be 10")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{1099511627776, "1.0T"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{500, "500ms"},
		{1000, "1.0s"},
		{1500, "1.5s"},
		{60000, "1m"},
		{90000, "1m 30s"},
		{3600000, "1h"},
		{5400000, "1h 30m"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.input)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{500, "500"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{1000000000, "1.0B"},
	}
	for _, tt := range tests {
		got := FormatTokens(tt.input)
		if got != tt.want {
			t.Errorf("FormatTokens(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatTokens64(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0"},
		{1000, "1.0K"},
		{1000000, "1.0M"},
		{1000000000, "1.0B"},
	}
	for _, tt := range tests {
		got := FormatTokens64(tt.input)
		if got != tt.want {
			t.Errorf("FormatTokens64(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		wantLen int
	}{
		{"/usr/local/bin", 20, 14},                             // fits
		{"/usr/local/bin/some/very/long/path/file.go", 20, 20}, // truncates
	}
	for _, tt := range tests {
		got := ShortenPath(tt.input, tt.maxLen)
		if len(got) != tt.wantLen {
			t.Errorf("ShortenPath(%q, %d) len = %d, want %d (got %q)",
				tt.input, tt.maxLen, len(got), tt.wantLen, got)
		}
	}
}

func TestGetModelFamily(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"claude-3-opus", "claude"},
		{"claude-3.5-sonnet", "claude"},
		{"gpt-4-turbo", "gpt"},
		{"gpt-3.5-turbo", "gpt"},
		{"o1-preview", "gpt"},
		{"gemini-pro", "gemini"},
		{"llama-3-70b", "llama"},
		{"qwen-2-72b", "qwen"},
		{"deepseek-coder", "deepseek"},
		{"mistral-large", "mistral"},
		{"unknown-model", "other"},
		{"", ""},
	}
	for _, tt := range tests {
		got := GetModelFamily(tt.input)
		if got != tt.want {
			t.Errorf("GetModelFamily(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
