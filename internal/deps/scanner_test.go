package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		file string
		lang string
	}{
		{"go.mod", "go"},
		{"Cargo.toml", "rust"},
		{"package.json", "node"},
		{"requirements.txt", "python"},
		{"unknown.txt", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		got := detectLanguage(tt.file)
		if got != tt.lang {
			t.Errorf("detectLanguage(%q) = %q, want %q", tt.file, got, tt.lang)
		}
	}
}

func TestScanDependencies_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	summary := ScanDependencies(dir)
	if summary == nil {
		t.Fatal("ScanDependencies returned nil")
	}
}

func TestScanDependencies_WithGoMod(t *testing.T) {
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	os.WriteFile(goMod, []byte("module test\n\ngo 1.21\n\nrequire (\n\tgithub.com/example/lib v1.0.0\n\tgithub.com/example/util v2.0.0\n)"), 0644)

	summary := ScanDependencies(dir)
	if summary == nil {
		t.Fatal("ScanDependencies returned nil")
	}
	if summary.Language != "go" {
		t.Errorf("Language = %q, want go", summary.Language)
	}
}

func TestScanDependencies_WithPackageJson(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"dependencies": {"express": "^4.0", "lodash": "^4.0"},
		"devDependencies": {"jest": "^29.0"}
	}`), 0644)

	summary := ScanDependencies(dir)
	if summary == nil {
		t.Fatal("ScanDependencies returned nil")
	}
	if summary.Language != "node" {
		t.Errorf("Language = %q, want node", summary.Language)
	}
}

func TestScanDependencies_WithCargoToml(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[package]
name = "test"
version = "0.1.0"

[dependencies]
serde = "1.0"
`), 0644)

	summary := ScanDependencies(dir)
	if summary == nil {
		t.Fatal("ScanDependencies returned nil")
	}
	if summary.Language != "rust" {
		t.Errorf("Language = %q, want rust", summary.Language)
	}
}
