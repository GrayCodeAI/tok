package compressor

import (
	"strings"
	"testing"
)

func TestCompressLite(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove filler words",
			input:    "Please really just utilize the configuration",
			expected: "Please utilize the configuration",
		},
		{
			name:     "remove hedging",
			input:    "Perhaps maybe it could possibly work",
			expected: "it work",
		},
		{
			name:     "keep articles",
			input:    "The database and a table",
			expected: "The database and a table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressLite(tt.input)
			if result != tt.expected {
				t.Errorf("compressLite(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompressFull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove articles",
			input:    "The database connection is working",
			expected: "database connection is working",
		},
		{
			name:     "replace verbose phrases",
			input:    "In order to fix the bug",
			expected: "to fix bug",
		},
		{
			name:     "shorten words",
			input:    "Please utilize additional functionality",
			expected: "Please use more functionality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressFull(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("compressFull(%q) = %q, want containing %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompressUltra(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "abbreviate technical terms",
			input:    "database authentication configuration",
			expected: "DB auth config",
		},
		{
			name:     "remove connectors",
			input:    "The function and the implementation",
			expected: "fn impl",
		},
		{
			name:     "arrow causality",
			input:    "This causes the error",
			expected: "This → error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressUltra(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("compressUltra(%q) = %q, want containing %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompressWenyanFull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "database term",
			input:    "database configuration",
			expected: "庫 配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressWenyanFull(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("compressWenyanFull(%q) = %q, want containing %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mode     string
		expected string
	}{
		{
			name:     "lite mode",
			input:    "just test",
			mode:     "lite",
			expected: "test",
		},
		{
			name:     "full mode",
			input:    "the test",
			mode:     "full",
			expected: "test",
		},
		{
			name:     "ultra mode",
			input:    "database test",
			mode:     "ultra",
			expected: "DB test",
		},
		{
			name:     "wenyan mode",
			input:    "database configuration",
			mode:     "wenyan",
			expected: "庫 配置",
		},
		{
			name:     "invalid mode",
			input:    "test",
			mode:     "invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compress(tt.input, tt.mode)
			if tt.mode == "invalid" {
				if err == nil {
					t.Errorf("Compress(%q, %q) expected error, got nil", tt.input, tt.mode)
				}
				return
			}
			if err != nil {
				t.Errorf("Compress(%q, %q) error: %v", tt.input, tt.mode, err)
			}
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Compress(%q, %q) = %q, want containing %q", tt.input, tt.mode, result, tt.expected)
			}
		})
	}
}

func TestBackupPathFor(t *testing.T) {
	if got := backupPathFor("CLAUDE.md"); got != "CLAUDE.original.md" {
		t.Errorf("backupPathFor markdown = %q, want %q", got, "CLAUDE.original.md")
	}
	if got := backupPathFor("notes.txt"); got != "notes.txt.original" {
		t.Errorf("backupPathFor txt = %q, want %q", got, "notes.txt.original")
	}
}
