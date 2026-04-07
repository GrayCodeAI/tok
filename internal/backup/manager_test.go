package backup

import (
	"os"
	"testing"
)

func TestNewBackupManager(t *testing.T) {
	mgr := NewBackupManager("/tmp/backups")
	if mgr == nil {
		t.Error("Expected non-nil manager")
	}
}

func TestBackupManagerCreateBackup(t *testing.T) {
	mgr := NewBackupManager("/tmp")

	tmpFile := "/tmp/test_backup.txt"
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Skip("Skipping - temp file creation failed")
	}
	defer os.Remove(tmpFile)

	backup, err := mgr.CreateBackup(nil, tmpFile)
	if err != nil {
		t.Errorf("CreateBackup failed: %v", err)
	}
	if backup == nil {
		t.Error("Expected non-nil backup")
	}
}

func TestCreateTimestampedBackup(t *testing.T) {
	dest, err := CreateTimestampedBackup("/tmp/file.txt", "/tmp")
	if err != nil {
		t.Errorf("CreateTimestampedBackup failed: %v", err)
	}
	t.Logf("Destination: %s", dest)
}

func TestVerifyBackup(t *testing.T) {
	result := VerifyBackup("/nonexistent")
	if result {
		t.Error("Expected false for nonexistent file")
	}
}
