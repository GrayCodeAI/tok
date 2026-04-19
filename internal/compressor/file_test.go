package compressor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompressFilePreview(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "CLAUDE.md")
	content := "# Rules\nPlease utilize `npm test` before merge.\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	result, compressed, err := CompressFilePreview(file, "ultra")
	if err != nil {
		t.Fatalf("CompressFilePreview error: %v", err)
	}
	if result.BackupPath != filepath.Join(dir, "CLAUDE.original.md") {
		t.Fatalf("unexpected backup path: %s", result.BackupPath)
	}
	if !strings.Contains(compressed, "`npm test`") {
		t.Fatalf("inline code not preserved: %q", compressed)
	}
}

func TestValidateFileRejectsBackupFiles(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "notes.original.md")
	if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	err := ValidateFile(file)
	if err == nil {
		t.Fatal("expected backup file validation error, got nil")
	}
}

func TestCompressFilePreviewWithForceAllowsExistingBackup(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "notes.md")
	backup := filepath.Join(dir, "notes.original.md")
	if err := os.WriteFile(file, []byte("Please utilize config file."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("old backup"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := CompressFilePreview(file, "full"); err == nil {
		t.Fatal("expected preview to fail when backup exists without force")
	}

	if _, _, err := CompressFilePreviewWithOptions(file, "full", true); err != nil {
		t.Fatalf("preview with force should pass: %v", err)
	}
}
