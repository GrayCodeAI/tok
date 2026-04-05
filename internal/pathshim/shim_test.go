package pathshim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewPATHShim(t *testing.T) {
	dir := t.TempDir()
	shim, err := NewPATHShim(dir)
	if err != nil {
		t.Fatalf("NewPATHShim() error = %v", err)
	}
	if shim == nil {
		t.Fatal("NewPATHShim() returned nil")
	}
	if shim.GetPATHEntry() != dir {
		t.Errorf("GetPATHEntry() = %q, want %q", shim.GetPATHEntry(), dir)
	}
}

func TestNewPATHShim_DefaultDir(t *testing.T) {
	// Using empty string should default to /tmp/tokman-shims
	shim, err := NewPATHShim("")
	if err != nil {
		t.Fatalf("NewPATHShim('') error = %v", err)
	}
	if shim == nil {
		t.Fatal("NewPATHShim('') returned nil")
	}
}

func TestCreateShim(t *testing.T) {
	dir := t.TempDir()
	shim, err := NewPATHShim(dir)
	if err != nil {
		t.Fatalf("NewPATHShim() error = %v", err)
	}

	path, err := shim.CreateShim("git")
	if err != nil {
		t.Fatalf("CreateShim(git) error = %v", err)
	}
	expected := filepath.Join(dir, "git")
	if path != expected {
		t.Errorf("CreateShim(git) = %q, want %q", path, expected)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read shim: %v", err)
	}
	if !strings.Contains(string(content), "tokman git") {
		t.Errorf("shim content missing 'tokman git': %s", string(content))
	}
}

func TestCreateShim_InvalidName(t *testing.T) {
	dir := t.TempDir()
	shim, _ := NewPATHShim(dir)

	badNames := []string{
		"../malicious",
		"../../etc/passwd",
		"foo; rm -rf /",
		"foo`whoami`",
		"foo$(id)",
		"foo bar",
		"",
	}

	for _, name := range badNames {
		_, err := shim.CreateShim(name)
		if err == nil {
			t.Errorf("CreateShim(%q) should have returned error", name)
		}
	}
}

func TestCreateShim_Alphabetic(t *testing.T) {
	dir := t.TempDir()
	shim, _ := NewPATHShim(dir)

	goodNames := []string{"ls", "gcc-12", "node", "make", "go1.21", "cargo-clippy"}
	for _, name := range goodNames {
		_, err := shim.CreateShim(name)
		if err != nil {
			t.Errorf("CreateShim(%q) error = %v", name, err)
		}
	}
}

func TestInstallPrep(t *testing.T) {
	dir := t.TempDir()
	shim, _ := NewPATHShim(dir)

	original := os.Getenv("PATH")
	result := shim.InstallPrep()

	if result != dir+":"+original {
		t.Errorf("InstallPrep() = %q, want %q", result, dir+":"+original)
	}
}

func TestInstallPrep_Idempotent(t *testing.T) {
	dir := t.TempDir()
	shim, _ := NewPATHShim(dir)

	os.Setenv("PATH", dir+":/usr/bin")
	result := shim.InstallPrep()

	expect := dir + ":/usr/bin"
	if result != expect {
		t.Errorf("InstallPrep() duplicate guard = %q, want %q", result, expect)
	}

	// Restore
	os.Setenv("PATH", os.Getenv("PATH")) // keep the modified PATH for now
}

func TestPipeStripper_Strip(t *testing.T) {
	stripper := NewPipeStripper()

	tests := []struct {
		input, want string
	}{
		{"git status | head -20", "git status"},
		{"git log --oneline | grep foo", "git log --oneline"},
		{"cat file.txt", "cat file.txt"},
	}

	for _, tt := range tests {
		got := stripper.Strip(tt.input)
		if got != tt.want {
			t.Errorf("Strip(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPipeStripper_HasPipe(t *testing.T) {
	stripper := NewPipeStripper()

	if !stripper.HasPipe("git status | head") {
		t.Error("HasPipe should detect piped commands")
	}
	if stripper.HasPipe("git status") {
		t.Error("HasPipe should return false for non-piped commands")
	}
}

func TestNewPipeStripper_HasDefaults(t *testing.T) {
	stripper := NewPipeStripper()
	expected := []string{"head", "tail", "grep", "sort", "uniq", "wc", "cut", "awk", "sed", "tr", "xargs"}
	if len(stripper.commands) != len(expected) {
		t.Errorf("commands count = %d, want %d", len(stripper.commands), len(expected))
	}
}
