package tee

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- sanitizeSlug tests ---

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cargo_test", "cargo_test"},
		{"cargo test", "cargo_test"},
		{"cargo-test", "cargo-test"},
		{"go/test/./pkg", "go_test___pkg"},
		{"git status --short", "git_status_--short"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeSlug(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeSlug_Truncate(t *testing.T) {
	// Test truncation at 40 chars
	long := strings.Repeat("a", 50)
	result := sanitizeSlug(long)
	if len(result) != 40 {
		t.Errorf("sanitizeSlug truncation: got len %d, want 40", len(result))
	}
}

// --- shouldTee tests ---

func TestShouldTee_Disabled(t *testing.T) {
	result, _, _ := shouldTee(false, TeeModeFailures, 1000, 1, "/tmp/tee")
	if result {
		t.Error("shouldTee should return false when disabled")
	}
}

func TestShouldTee_NeverMode(t *testing.T) {
	result, _, _ := shouldTee(true, TeeModeNever, 1000, 1, "/tmp/tee")
	if result {
		t.Error("shouldTee should return false in 'never' mode")
	}
}

func TestShouldTee_FailuresMode_Success(t *testing.T) {
	// In failures mode, exit code 0 should not tee
	result, _, _ := shouldTee(true, TeeModeFailures, 1000, 0, "/tmp/tee")
	if result {
		t.Error("shouldTee should return false on success in 'failures' mode")
	}
}

func TestShouldTee_FailuresMode_Failure(t *testing.T) {
	// In failures mode, non-zero exit code should tee
	result, _, _ := shouldTee(true, TeeModeFailures, 1000, 1, "/tmp/tee")
	if !result {
		t.Error("shouldTee should return true on failure in 'failures' mode")
	}
}

func TestShouldTee_AlwaysMode(t *testing.T) {
	// In always mode, should tee regardless of exit code
	result1, _, _ := shouldTee(true, TeeModeAlways, 1000, 0, "/tmp/tee")
	result2, _, _ := shouldTee(true, TeeModeAlways, 1000, 1, "/tmp/tee")
	if !result1 || !result2 {
		t.Error("shouldTee should return true in 'always' mode")
	}
}

func TestShouldTee_SmallOutput(t *testing.T) {
	// Below MinTeeSize should not tee
	result, _, _ := shouldTee(true, TeeModeAlways, 100, 1, "/tmp/tee")
	if result {
		t.Error("shouldTee should return false for small output")
	}
}

// --- writeTeeFile tests ---

func TestWriteTeeFile_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	content := strings.Repeat("error: test failed\n", 50)

	path, err := writeTeeFile(content, "cargo_test", tmpDir, DefaultMaxFileSize, 20)
	if err != nil {
		t.Fatalf("writeTeeFile failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("writeTeeFile should create the file")
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if !strings.Contains(string(written), "error: test failed") {
		t.Error("Written content should contain the original content")
	}
}

func TestWriteTeeFile_Truncation(t *testing.T) {
	tmpDir := t.TempDir()
	bigOutput := strings.Repeat("x", 2000)

	// Set max_file_size to 1000 bytes
	path, err := writeTeeFile(bigOutput, "test", tmpDir, 1000, 20)
	if err != nil {
		t.Fatalf("writeTeeFile failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "--- truncated at 1000 bytes ---") {
		t.Error("Content should contain truncation notice")
	}

	if len(content) >= 2000 {
		t.Error("Content should be truncated")
	}
}

func TestWriteTeeFile_TruncationUTF8(t *testing.T) {
	tmpDir := t.TempDir()
	// Japanese chars are 3 bytes each in UTF-8
	japanese := strings.Repeat("\u6F22", 333) // 999 bytes of 3-byte chars

	// Truncate at 998 — falls in the middle of the 333rd character
	path, err := writeTeeFile(japanese, "test_utf8", tmpDir, 998, 20)
	if err != nil {
		t.Fatalf("writeTeeFile failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "--- truncated at 998 bytes ---") {
		t.Error("Content should contain truncation notice")
	}

	// Should contain 332 full characters (996 bytes), not panic
	expectedPrefix := strings.Repeat("\u6F22", 332)
	if !strings.HasPrefix(string(content), expectedPrefix) {
		t.Error("Content should preserve valid UTF-8 boundary")
	}
}

func TestWriteTeeFile_TruncationEmoji(t *testing.T) {
	tmpDir := t.TempDir()
	// Emoji are 4 bytes each in UTF-8
	emojis := strings.Repeat("\U0001F600", 100) // 400 bytes

	// Truncate at 201 — falls mid-emoji (4-byte boundary is at 200, 204)
	path, err := writeTeeFile(emojis, "test_emoji", tmpDir, 201, 20)
	if err != nil {
		t.Fatalf("writeTeeFile failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "--- truncated at 201 bytes ---") {
		t.Error("Content should contain truncation notice")
	}

	// The emoji portion should be exactly 200 bytes (50 emojis),
	// rounded down from 201 to the nearest char boundary
	target := strings.Repeat("\U0001F600", 50)
	if !strings.HasPrefix(string(content), target) {
		t.Error("Content should preserve valid UTF-8 boundary for emoji")
	}
}

// --- cleanupOldFiles tests ---

func TestCleanupOldFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 25 .log files with different timestamps in names
	for i := 0; i < 25; i++ {
		filename := fmt.Sprintf("%010d_test.log", 1000000+i)
		os.WriteFile(filepath.Join(tmpDir, filename), []byte("content"), 0644)
	}

	err := cleanupOldFiles(tmpDir, 20)
	if err != nil {
		t.Fatalf("cleanupOldFiles failed: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) != 20 {
		t.Errorf("Expected 20 files, got %d", len(entries))
	}

	// Oldest 5 should be removed
	for i := 0; i < 5; i++ {
		filename := fmt.Sprintf("%010d_test.log", 1000000+i)
		if _, err := os.Stat(filepath.Join(tmpDir, filename)); !os.IsNotExist(err) {
			t.Errorf("Old file %s should be removed", filename)
		}
	}

	// Newest 20 should remain
	for i := 5; i < 25; i++ {
		filename := fmt.Sprintf("%010d_test.log", 1000000+i)
		if _, err := os.Stat(filepath.Join(tmpDir, filename)); os.IsNotExist(err) {
			t.Errorf("New file %s should remain", filename)
		}
	}
}

// --- formatHint tests ---

func TestFormatHint(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     filepath.Join(home, ".local", "share", "tokman", "tee", "test.log"),
			expected: "[full output: ~/.local/share/tokman/tee/test.log]",
		},
		{
			path:     "/tmp/rtk/tee/123_cargo_test.log",
			expected: "[full output: /tmp/rtk/tee/123_cargo_test.log]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := formatHint(tt.path)
			if result != tt.expected {
				t.Errorf("formatHint(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// --- TeeConfig tests ---

func TestDefaultTeeConfig(t *testing.T) {
	cfg := DefaultTeeConfig()

	if !cfg.Enabled {
		t.Error("DefaultTeeConfig should have Enabled = true")
	}
	if cfg.Mode != TeeModeFailures {
		t.Errorf("DefaultTeeConfig should have Mode = %q, got %q", TeeModeFailures, cfg.Mode)
	}
	if cfg.MaxFiles != DefaultMaxFiles {
		t.Errorf("DefaultTeeConfig should have MaxFiles = %d, got %d", DefaultMaxFiles, cfg.MaxFiles)
	}
	if cfg.MaxFileSize != DefaultMaxFileSize {
		t.Errorf("DefaultTeeConfig should have MaxFileSize = %d, got %d", DefaultMaxFileSize, cfg.MaxFileSize)
	}
}

// --- Integration tests ---

func TestTeeRaw_EnvDisable(t *testing.T) {
	os.Setenv("TOKMAN_TEE", "0")
	defer os.Unsetenv("TOKMAN_TEE")

	content := strings.Repeat("error output\n", 100)
	path := TeeRaw(content, "test_cmd", 1)

	if path != "" {
		t.Error("TeeRaw should return empty when TOKMAN_TEE=0")
	}
}

func TestTeeRaw_SmallOutput(t *testing.T) {
	// Small output should not be teed
	content := "short error"
	path := TeeRaw(content, "test_cmd", 1)

	if path != "" {
		t.Error("TeeRaw should return empty for small output")
	}
}

func TestTeeRaw_SuccessInFailuresMode(t *testing.T) {
	// In default mode (failures), success should not tee
	content := strings.Repeat("success output\n", 100)
	path := TeeRaw(content, "test_cmd", 0)

	if path != "" {
		t.Error("TeeRaw should not tee on success in 'failures' mode")
	}
}

func TestTeeAndHint(t *testing.T) {
	// Set up temp directory for tee files
	tmpDir := t.TempDir()
	os.Setenv("TOKMAN_TEE_DIR", tmpDir)
	defer os.Unsetenv("TOKMAN_TEE_DIR")

	// Set mode to always so we can test with exit code 0
	os.Setenv("TOKMAN_TEE_MODE", "always")
	defer os.Unsetenv("TOKMAN_TEE_MODE")

	content := strings.Repeat("test output\n", 100)
	hint := TeeAndHint(content, "test_cmd", 0)

	if hint == "" {
		t.Fatal("TeeAndHint should return a hint")
	}

	if !strings.HasPrefix(hint, "[full output: ") {
		t.Errorf("Hint should start with '[full output: ', got %q", hint)
	}

	if !strings.HasSuffix(hint, "]") {
		t.Errorf("Hint should end with ']', got %q", hint)
	}
}

func TestForceTeeHint(t *testing.T) {
	// Set up temp directory for tee files
	tmpDir := t.TempDir()
	os.Setenv("TOKMAN_TEE_DIR", tmpDir)
	defer os.Unsetenv("TOKMAN_TEE_DIR")

	content := strings.Repeat("force tee output\n", 100)
	hint := ForceTeeHint(content, "test_cmd")

	if hint == "" {
		t.Fatal("ForceTeeHint should return a hint")
	}

	if !strings.HasPrefix(hint, "[full output: ") {
		t.Errorf("Hint should start with '[full output: ', got %q", hint)
	}
}

func TestForceTeeHint_SmallOutput(t *testing.T) {
	// Small output should be skipped even in force mode
	content := "short"
	hint := ForceTeeHint(content, "test_cmd")

	if hint != "" {
		t.Error("ForceTeeHint should return empty for small output")
	}
}

func TestForceTeeHint_EnvDisable(t *testing.T) {
	os.Setenv("TOKMAN_TEE", "0")
	defer os.Unsetenv("TOKMAN_TEE")

	content := strings.Repeat("force tee output\n", 100)
	hint := ForceTeeHint(content, "test_cmd")

	if hint != "" {
		t.Error("ForceTeeHint should respect TOKMAN_TEE=0")
	}
}

// --- ListTeeFiles and CleanupTeeFiles tests ---

func TestListTeeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TOKMAN_TEE_DIR", tmpDir)
	defer os.Unsetenv("TOKMAN_TEE_DIR")

	// Create some test files
	for i := 0; i < 3; i++ {
		filename := fmt.Sprintf("%d_test.log", time.Now().Unix()+int64(i))
		os.WriteFile(filepath.Join(tmpDir, filename), []byte("content"), 0644)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	files, err := ListTeeFiles()
	if err != nil {
		t.Fatalf("ListTeeFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestCleanupTeeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TOKMAN_TEE_DIR", tmpDir)
	defer os.Unsetenv("TOKMAN_TEE_DIR")

	// Create files with old and new timestamps
	oldTime := time.Now().Add(-7 * 24 * time.Hour)
	newTime := time.Now()

	// Create old file
	oldFile := filepath.Join(tmpDir, "1000000_old.log")
	os.WriteFile(oldFile, []byte("old"), 0644)
	os.Chtimes(oldFile, oldTime, oldTime)

	// Create new file
	newFile := filepath.Join(tmpDir, fmt.Sprintf("%d_new.log", newTime.Unix()))
	os.WriteFile(newFile, []byte("new"), 0644)

	removed, err := CleanupTeeFiles(24 * time.Hour)
	if err != nil {
		t.Fatalf("CleanupTeeFiles failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 file removed, got %d", removed)
	}

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should be removed")
	}

	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should remain")
	}
}

// --- Benchmarks ---

func BenchmarkSanitizeSlug(b *testing.B) {
	input := "cargo test --package my-app --lib"
	for i := 0; i < b.N; i++ {
		sanitizeSlug(input)
	}
}

func BenchmarkWriteTeeFile(b *testing.B) {
	tmpDir := b.TempDir()
	content := strings.Repeat("benchmark test output line\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeTeeFile(content, "benchmark_test", tmpDir, DefaultMaxFileSize, 100)
	}
}
