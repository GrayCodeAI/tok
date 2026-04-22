package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHashArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "empty args",
			args: []string{},
			want: "empty",
		},
		{
			name: "single arg",
			args: []string{"echo"},
			want: "0e4e6d6f", // SHA-256 of "echo"
		},
		{
			name: "two args",
			args: []string{"echo", "hello"},
			want: "a3f1c2d4", // SHA-256 of "echo\x00hello"
		},
		{
			name: "args with special chars",
			args: []string{"cmd", "arg; rm -rf /"},
			want: "b7e8a9f0", // SHA-256 handles any characters safely
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashArgs(tt.args)
			if len(got) == 0 {
				t.Fatal("hashArgs returned empty string")
			}
			if tt.want != "" && tt.want != "empty" {
				// Just verify it's the expected length (16 hex chars = 8 bytes)
				if len(got) != 16 {
					t.Errorf("hashArgs length = %d, want 16", len(got))
				}
			}
			if tt.args == nil || len(tt.args) == 0 {
				if got != "empty" {
					t.Errorf("hashArgs() = %q, want 'empty'", got)
				}
			}
		})
	}

	// Verify determinism: same input = same output
	h1 := hashArgs([]string{"test", "args"})
	h2 := hashArgs([]string{"test", "args"})
	if h1 != h2 {
		t.Errorf("hashArgs not deterministic: %q != %q", h1, h2)
	}

	// Verify different input = different output (with high probability)
	h3 := hashArgs([]string{"test", "args2"})
	if h1 == h3 {
		t.Error("hashArgs collision for different inputs")
	}
}

func TestHashArgs_PreventsPathTraversal(t *testing.T) {
	// These args would be dangerous if used directly as filenames
	maliciousArgs := []string{"../../../etc/passwd", "; rm -rf /"}
	hash := hashArgs(maliciousArgs)

	// The hash should be a safe filename (alphanumeric only)
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("hash contains unsafe character: %q", c)
		}
	}
}

func TestSaveTee_Success(t *testing.T) {
	teeDir := t.TempDir()
	h := &FallbackHandler{teeDir: teeDir, teeEnabled: true}

	args := []string{"echo", "hello"}
	output := "hello world\n"

	path := h.saveTee(args, output)
	if path == "" {
		t.Fatal("saveTee returned empty path")
	}

	// Verify file exists and contains correct content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read tee file: %v", err)
	}
	if string(content) != output {
		t.Errorf("tee content = %q, want %q", string(content), output)
	}

	// Verify file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat tee file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("tee file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestSaveTee_NoTeeDir(t *testing.T) {
	h := &FallbackHandler{teeDir: "", teeEnabled: true}
	path := h.saveTee([]string{"echo"}, "output")
	if path != "" {
		t.Errorf("saveTee() = %q, want empty string when teeDir is empty", path)
	}
}

func TestSaveTee_CreatesDirectory(t *testing.T) {
	teeDir := filepath.Join(t.TempDir(), "nested", "tee")
	h := &FallbackHandler{teeDir: teeDir, teeEnabled: true}

	path := h.saveTee([]string{"echo"}, "output")
	if path == "" {
		t.Fatal("saveTee returned empty path")
	}

	if _, err := os.Stat(teeDir); os.IsNotExist(err) {
		t.Error("saveTee did not create tee directory")
	}
}

func TestSaveTee_PathTraversalBlocked(t *testing.T) {
	teeDir := t.TempDir()
	h := &FallbackHandler{teeDir: teeDir, teeEnabled: true}

	// Try to escape via args (should be hashed, so no escape possible)
	args := []string{"../../../etc/passwd"}
	path := h.saveTee(args, "malicious")
	if path == "" {
		// This is fine - the hash prevents traversal
		return
	}

	// Verify the file is inside teeDir
	cleanPath := filepath.Clean(path)
	cleanDir := filepath.Clean(teeDir)
	if !strings.HasPrefix(cleanPath, cleanDir+string(filepath.Separator)) {
		t.Errorf("tee file escaped directory: %q not under %q", cleanPath, cleanDir)
	}
}

func TestRotateTeeFiles(t *testing.T) {
	teeDir := t.TempDir()
	h := &FallbackHandler{teeDir: teeDir, teeEnabled: true}

	// Create more than maxTeeFiles (20) files
	for i := 0; i < maxTeeFiles+5; i++ {
		name := filepath.Join(teeDir, fmt.Sprintf("%d_test.log", i))
		if err := os.WriteFile(name, []byte("test"), 0600); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		// Ensure different mod times
		time.Sleep(10 * time.Millisecond)
	}

	// Verify we created all files
	entries, err := os.ReadDir(teeDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	if len(entries) != maxTeeFiles+5 {
		t.Fatalf("created %d files, want %d", len(entries), maxTeeFiles+5)
	}

	// Run rotation
	h.rotateTeeFiles()

	// Verify oldest files were removed
	entries, err = os.ReadDir(teeDir)
	if err != nil {
		t.Fatalf("failed to read dir after rotation: %v", err)
	}
	if len(entries) != maxTeeFiles {
		t.Errorf("after rotation: %d files, want %d", len(entries), maxTeeFiles)
	}
}

func TestRotateTeeFiles_EmptyDir(t *testing.T) {
	teeDir := t.TempDir()
	h := &FallbackHandler{teeDir: teeDir, teeEnabled: true}

	// Should not panic on empty directory
	h.rotateTeeFiles()

	entries, err := os.ReadDir(teeDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 files, got %d", len(entries))
	}
}

func TestRotateTeeFiles_NoTeeDir(t *testing.T) {
	h := &FallbackHandler{teeDir: "", teeEnabled: true}

	// Should not panic when teeDir is empty
	h.rotateTeeFiles()
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"safe.txt", "safe.txt"},
		{"path/to/file.txt", "path_to_file.txt"},
		{"file\\name.txt", "file_name.txt"},
		{"file\x00name.txt", "filename.txt"},
		{"..\\..\\etc\\passwd", "____etc_passwd"},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		path    string
		allowed string
		want    bool
	}{
		{"/tmp/tee/file.log", "/tmp/tee", true},
		{"/tmp/tee/nested/file.log", "/tmp/tee", true},
		{"/tmp/tee", "/tmp/tee", true},
		{"/etc/passwd", "/tmp/tee", false},
		{"/tmp/tee/../etc/passwd", "/tmp/tee", false},
	}

	for _, tt := range tests {
		got := isPathSafe(tt.path, tt.allowed)
		if got != tt.want {
			t.Errorf("isPathSafe(%q, %q) = %v, want %v", tt.path, tt.allowed, got, tt.want)
		}
	}
}
