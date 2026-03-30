package system

import (
	"strings"
	"testing"
)

func TestFilterLSOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDirs int
		wantFiles int
	}{
		{
			name: "empty output",
			input: "",
			wantDirs: 0,
			wantFiles: 0,
		},
		{
			name: "basic ls output",
			input: `total 16
drwxr-xr-x   4 user  staff   128 Jan  1 12:00 src
-rw-r--r--   1 user  staff  1024 Jan  1 12:00 main.go
-rw-r--r--   1 user  staff   512 Jan  1 12:00 readme.md`,
			wantDirs: 1,
			wantFiles: 2,
		},
		{
			name: "with noise dirs filtered",
			input: `total 32
drwxr-xr-x   6 user  staff   192 Jan  1 12:00 .
drwxr-xr-x   3 user  staff    96 Jan  1 12:00 ..
drwxr-xr-x  10 user  staff   320 Jan  1 12:00 .git
drwxr-xr-x  20 user  staff   640 Jan  1 12:00 node_modules
drwxr-xr-x   4 user  staff   128 Jan  1 12:00 src
-rw-r--r--   1 user  staff  1024 Jan  1 12:00 main.go`,
			wantDirs: 1, // only src (. and .. are skipped, .git and node_modules are noise)
			wantFiles: 1,
		},
		{
			name: "file with spaces in name",
			input: `total 8
-rw-r--r--   1 user  staff  100 Jan  1 12:00 my file name.txt`,
			wantDirs: 0,
			wantFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := filterLSOutput(tt.input)
			
			// Count directories and files in output
			dirCount := strings.Count(output, "📁")
			fileCount := strings.Count(output, "📄")
			
			if dirCount != tt.wantDirs {
				t.Errorf("filterLSOutput() got %d dirs, want %d", dirCount, tt.wantDirs)
			}
			if fileCount != tt.wantFiles {
				t.Errorf("filterLSOutput() got %d files, want %d", fileCount, tt.wantFiles)
			}
		})
	}
}

func TestFilterLSOutputUltraCompact(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDirs int
		wantFiles int
	}{
		{
			name: "basic output",
			input: `total 16
drwxr-xr-x   4 user  staff   128 Jan  1 12:00 src
-rw-r--r--   1 user  staff  1024 Jan  1 12:00 main.go
-rw-r--r--   1 user  staff   512 Jan  1 12:00 readme.md`,
			wantDirs: 1,
			wantFiles: 2,
		},
		{
			name: "noise dirs filtered",
			input: `total 32
drwxr-xr-x  10 user  staff   320 Jan  1 12:00 .git
drwxr-xr-x  20 user  staff   640 Jan  1 12:00 node_modules
drwxr-xr-x   4 user  staff   128 Jan  1 12:00 src
-rw-r--r--   1 user  staff  1024 Jan  1 12:00 main.go`,
			wantDirs: 1,
			wantFiles: 1,
		},
		{
			name: "summary line included",
			input: `total 16
-rw-r--r--   1 user  staff  1024 Jan  1 12:00 main.go
-rw-r--r--   1 user  staff   512 Jan  1 12:00 test.go`,
			wantDirs: 0,
			wantFiles: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := filterLSOutputUltraCompact(tt.input)
			
			// Check that summary line is present
			if tt.wantDirs > 0 || tt.wantFiles > 0 {
				if !strings.Contains(output, "files") && !strings.Contains(output, "dirs") {
					t.Errorf("filterLSOutputUltraCompact() missing summary line")
				}
			}
			
			// Count entries (lines not starting with numbers for summary)
			lines := strings.Split(output, "\n")
			var entryCount int
			for _, line := range lines {
				if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "1") && !strings.HasPrefix(strings.TrimSpace(line), "2") {
					entryCount++
				}
			}
			
			expectedEntries := tt.wantDirs + tt.wantFiles
			// Account for summary line
			if expectedEntries > 0 && entryCount < expectedEntries {
				t.Errorf("filterLSOutputUltraCompact() got %d entries, want at least %d", entryCount, expectedEntries)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1024", 1024},
		{"0", 0},
		{"999999", 999999},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseSize(tt.input)
			if result != tt.expected {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadNoiseDirs(t *testing.T) {
	noiseDirs := loadNoiseDirs()
	
	// Should always contain common noise directories
	expectedNoise := []string{".git", "node_modules", "target", "vendor"}
	
	for _, dir := range expectedNoise {
		if !noiseDirs[dir] {
			t.Errorf("loadNoiseDirs() missing expected noise dir %q", dir)
		}
	}
}
