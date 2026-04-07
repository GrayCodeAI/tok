package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Backup struct {
	ID        string
	Path      string
	Size      int64
	CreatedAt time.Time
	Checksum  string
}

type BackupManager struct {
	backupDir string
}

func NewBackupManager(backupDir string) *BackupManager {
	return &BackupManager{backupDir: backupDir}
}

func (m *BackupManager) CreateBackup(ctx context.Context, source string) (*Backup, error) {
	stat, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	backup := &Backup{
		ID:        generateID(),
		Path:      source,
		Size:      stat.Size(),
		CreatedAt: time.Now(),
	}

	return backup, nil
}

func (m *BackupManager) Restore(ctx context.Context, backup *Backup, dest string) error {
	return nil
}

func (m *BackupManager) List(ctx context.Context) ([]*Backup, error) {
	return []*Backup{}, nil
}

func (m *BackupManager) Delete(ctx context.Context, id string) error {
	return nil
}

func generateID() string {
	return fmt.Sprintf("backup-%d", time.Now().Unix())
}

func CreateTimestampedBackup(source, dest string) (string, error) {
	ext := filepath.Ext(source)
	name := source[:len(source)-len(ext)]
	timestamp := time.Now().Format("20060102-150405")
	destPath := fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	return destPath, nil
}

func CompressBackup(source, dest string) error {
	return nil
}

func EncryptBackup(source, dest string) error {
	return nil
}

func VerifyBackup(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
