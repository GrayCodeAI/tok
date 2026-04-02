package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	t.Run("SaveAndGetMetadata", func(t *testing.T) {
		meta := &CacheMetadata{
			Hash:      "abc123",
			FilePath:  "/test/file.go",
			Timestamp: time.Now(),
			Accessed:  time.Now(),
			HitCount:  5,
			Size:      1024,
		}

		if err := store.SaveMetadata(meta); err != nil {
			t.Fatalf("SaveMetadata() error = %v", err)
		}

		retrieved, err := store.GetMetadata("abc123")
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected metadata, got nil")
		}
		if retrieved.Hash != meta.Hash {
			t.Errorf("expected hash %q, got %q", meta.Hash, retrieved.Hash)
		}
		if retrieved.FilePath != meta.FilePath {
			t.Errorf("expected path %q, got %q", meta.FilePath, retrieved.FilePath)
		}
		if retrieved.HitCount != meta.HitCount {
			t.Errorf("expected hit count %d, got %d", meta.HitCount, retrieved.HitCount)
		}
	})

	t.Run("LoadAllMetadata", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			meta := &CacheMetadata{
				Hash:      fmt.Sprintf("hash%d", i),
				FilePath:  fmt.Sprintf("/test/file%d.go", i),
				Timestamp: time.Now(),
				Accessed:  time.Now(),
				HitCount:  i,
				Size:      int64(i * 100),
			}
			if err := store.SaveMetadata(meta); err != nil {
				t.Fatalf("SaveMetadata() error = %v", err)
			}
		}

		all, err := store.LoadAllMetadata()
		if err != nil {
			t.Fatalf("LoadAllMetadata() error = %v", err)
		}
		if len(all) < 5 {
			t.Errorf("expected at least 5 entries, got %d", len(all))
		}
	})

	t.Run("DeleteMetadata", func(t *testing.T) {
		meta := &CacheMetadata{
			Hash:      "todelete",
			FilePath:  "/test/delete.go",
			Timestamp: time.Now(),
			Accessed:  time.Now(),
			HitCount:  1,
			Size:      100,
		}
		if err := store.SaveMetadata(meta); err != nil {
			t.Fatalf("SaveMetadata() error = %v", err)
		}

		if err := store.DeleteMetadata("todelete"); err != nil {
			t.Fatalf("DeleteMetadata() error = %v", err)
		}

		retrieved, err := store.GetMetadata("todelete")
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}
		if retrieved != nil {
			t.Error("expected nil after deletion")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		count, size, _, _, err := store.GetStats()
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}
		if count < 0 {
			t.Error("expected non-negative count")
		}
		if size < 0 {
			t.Error("expected non-negative size")
		}
	})

	t.Run("StoreMetadata", func(t *testing.T) {
		if err := store.SetStoreMetadata("version", "1.0.0"); err != nil {
			t.Fatalf("SetStoreMetadata() error = %v", err)
		}

		value, err := store.GetStoreMetadata("version")
		if err != nil {
			t.Fatalf("GetStoreMetadata() error = %v", err)
		}
		if value != "1.0.0" {
			t.Errorf("expected version 1.0.0, got %q", value)
		}
	})

	t.Run("BackupAndRestore", func(t *testing.T) {
		meta := &CacheMetadata{
			Hash:      "backup123",
			FilePath:  "/test/backup.go",
			Timestamp: time.Now(),
			Accessed:  time.Now(),
			HitCount:  10,
			Size:      2048,
		}
		if err := store.SaveMetadata(meta); err != nil {
			t.Fatalf("SaveMetadata() error = %v", err)
		}

		backupPath := filepath.Join(dir, "backup.json")
		if err := store.Backup(backupPath); err != nil {
			t.Fatalf("Backup() error = %v", err)
		}

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Fatal("backup file not created")
		}

		newDBPath := filepath.Join(dir, "restore.db")
		newStore, err := NewStore(newDBPath)
		if err != nil {
			t.Fatalf("NewStore() error = %v", err)
		}
		defer newStore.Close()

		if err := newStore.Restore(backupPath); err != nil {
			t.Fatalf("Restore() error = %v", err)
		}

		restored, err := newStore.GetMetadata("backup123")
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}
		if restored == nil {
			t.Fatal("expected restored metadata, got nil")
		}
		if restored.Hash != "backup123" {
			t.Errorf("expected hash backup123, got %q", restored.Hash)
		}
	})
}

func TestDeleteOlderThan(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	oldMeta := &CacheMetadata{
		Hash:      "old",
		FilePath:  "/test/old.go",
		Timestamp: time.Now().Add(-48 * time.Hour),
		Accessed:  time.Now().Add(-48 * time.Hour),
		HitCount:  1,
		Size:      100,
	}
	if err := store.SaveMetadata(oldMeta); err != nil {
		t.Fatalf("SaveMetadata() error = %v", err)
	}

	newMeta := &CacheMetadata{
		Hash:      "new",
		FilePath:  "/test/new.go",
		Timestamp: time.Now(),
		Accessed:  time.Now(),
		HitCount:  1,
		Size:      100,
	}
	if err := store.SaveMetadata(newMeta); err != nil {
		t.Fatalf("SaveMetadata() error = %v", err)
	}

	deleted, err := store.DeleteOlderThan(24 * time.Hour)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	oldRetrieved, err := store.GetMetadata("old")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}
	if oldRetrieved != nil {
		t.Error("expected old entry to be deleted")
	}

	newRetrieved, err := store.GetMetadata("new")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}
	if newRetrieved == nil {
		t.Error("expected new entry to exist")
	}
}
