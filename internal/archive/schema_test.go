package archive

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_archive.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestSchemaManager_Initialize(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	// Initialize schema
	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify tables were created
	tables := []string{
		"archives",
		"archive_tags",
		"archive_access_log",
		"schema_version",
		"archive_stats",
		"archive_quotas",
		"cleanup_jobs",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("Table %s not created", table)
		}
	}
}

func TestSchemaManager_GetVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	// Before initialization, version query should fail (no table)
	// This is expected behavior - schema not initialized yet
	_, err := sm.GetVersion(ctx)
	if err == nil {
		t.Error("GetVersion() should error before initialization")
	}

	// After initialization, version should be 2 (current schema version)
	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	version, err := sm.GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}
	if version != 2 {
		t.Errorf("GetVersion() = %d, want 2", version)
	}
}

func TestSchemaManager_Verify(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	// Before initialization, verify should fail
	if err := sm.Verify(ctx); err == nil {
		t.Error("Verify() should fail before initialization")
	}

	// After initialization, verify should pass
	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if err := sm.Verify(ctx); err != nil {
		t.Errorf("Verify() error = %v", err)
	}
}

func TestSchemaManager_Stats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	// Initialize schema
	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Get stats
	stats, err := sm.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}

	// Verify initial stats
	if stats.TotalArchives != 0 {
		t.Errorf("TotalArchives = %d, want 0", stats.TotalArchives)
	}

	if stats.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want 2", stats.SchemaVersion)
	}
}

func TestDBStats_CompressionRatio(t *testing.T) {
	tests := []struct {
		name           string
		originalSize   int64
		compressedSize int64
		expectedRatio  float64
	}{
		{
			name:           "no compression",
			originalSize:   1000,
			compressedSize: 1000,
			expectedRatio:  1.0,
		},
		{
			name:           "50% compression",
			originalSize:   1000,
			compressedSize: 500,
			expectedRatio:  0.5,
		},
		{
			name:           "zero original size",
			originalSize:   0,
			compressedSize: 0,
			expectedRatio:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &DBStats{
				TotalOriginalSize:   tt.originalSize,
				TotalCompressedSize: tt.compressedSize,
			}

			ratio := stats.CompressionRatio()
			if ratio != tt.expectedRatio {
				t.Errorf("CompressionRatio() = %v, want %v", ratio, tt.expectedRatio)
			}
		})
	}
}

func TestDBStats_SpaceSaved(t *testing.T) {
	tests := []struct {
		name           string
		originalSize   int64
		compressedSize int64
		expectedSaved  int64
	}{
		{
			name:           "no savings",
			originalSize:   1000,
			compressedSize: 1000,
			expectedSaved:  0,
		},
		{
			name:           "50% savings",
			originalSize:   1000,
			compressedSize: 500,
			expectedSaved:  500,
		},
		{
			name:           "compression larger (edge case)",
			originalSize:   500,
			compressedSize: 1000,
			expectedSaved:  -500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &DBStats{
				TotalOriginalSize:   tt.originalSize,
				TotalCompressedSize: tt.compressedSize,
			}

			saved := stats.SpaceSaved()
			if saved != tt.expectedSaved {
				t.Errorf("SpaceSaved() = %v, want %v", saved, tt.expectedSaved)
			}
		})
	}
}

func TestSchemaManager_Reset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	// Initialize and verify
	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Reset
	if err := sm.Reset(ctx); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Verify tables are gone
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='archives'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check archives table: %v", err)
	}
	if count != 0 {
		t.Error("Reset() should delete archives table")
	}
}

func TestIndexes(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify required indexes exist
	indexes := []string{
		"idx_archives_hash",
		"idx_archives_category",
		"idx_archives_agent",
		"idx_archives_created_at",
		"idx_archive_tags_tag",
		"idx_archive_tags_archive_id",
	}

	for _, index := range indexes {
		var count int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check index %s: %v", index, err)
			continue
		}
		if count != 1 {
			t.Errorf("Index %s not created", index)
		}
	}
}

func TestForeignKeys(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	sm := NewSchemaManager(db)

	if err := sm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Insert a test archive
	result, err := db.ExecContext(ctx, `
		INSERT INTO archives (hash, original_content, filtered_content, original_size, compressed_size)
		VALUES ('test_hash', 'original', 'filtered', 100, 50)
	`)
	if err != nil {
		t.Fatalf("Failed to insert archive: %v", err)
	}

	archiveID, _ := result.LastInsertId()

	// Insert a tag for the archive
	_, err = db.ExecContext(ctx, `
		INSERT INTO archive_tags (archive_id, tag)
		VALUES (?, 'test_tag')
	`, archiveID)
	if err != nil {
		t.Fatalf("Failed to insert tag: %v", err)
	}

	// Delete the archive (should cascade delete the tag)
	_, err = db.ExecContext(ctx, "DELETE FROM archives WHERE id = ?", archiveID)
	if err != nil {
		t.Fatalf("Failed to delete archive: %v", err)
	}

	// Verify tag was deleted
	var tagCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM archive_tags WHERE archive_id = ?", archiveID).Scan(&tagCount)
	if err != nil {
		t.Fatalf("Failed to count tags: %v", err)
	}
	if tagCount != 0 {
		t.Error("Tag should be deleted via cascade")
	}
}
