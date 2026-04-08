package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/tests/integration/helpers"
)

// setupTestArchiveManager creates a new archive manager with a temporary database
func setupTestArchiveManager(t *testing.T) (*archive.ArchiveManager, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := archive.ArchiveConfig{
		MaxSize:      10 * 1024 * 1024,
		Expiration:   24 * time.Hour,
		Enabled:      true,
		DatabasePath: filepath.Join(tmpDir, "archive.db"),
	}

	manager, err := archive.NewArchiveManager(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create archive manager: %v", err)
	}

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		manager.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to initialize manager: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.RemoveAll(tmpDir)
	}

	return manager, cleanup
}

// TestArchiveBasicFlow tests basic archive and retrieve flow
func TestArchiveBasicFlow(t *testing.T) {
	manager, cleanup := setupTestArchiveManager(t)
	defer cleanup()

	ctx := context.Background()

	// Test content
	content := []byte("This is test content for archiving")
	entry := &archive.ArchiveEntry{
		OriginalContent: content,
		FilteredContent: content,
		OriginalSize:    int64(len(content)),
		CompressedSize:  int64(len(content)),
		Command:         "test-command",
		Category:        archive.CategoryCommand,
		Tags:            []string{"test"},
	}

	// Archive content
	hash, err := manager.Archive(ctx, entry)
	if err != nil {
		t.Fatalf("failed to archive content: %v", err)
	}

	if hash == "" {
		t.Error("archive returned empty hash")
	}

	// Retrieve content
	retrieved, err := manager.Retrieve(ctx, hash)
	if err != nil {
		t.Fatalf("failed to retrieve content: %v", err)
	}

	// Verify content matches
	retrievedContent := retrieved.OriginalContent
	if string(retrievedContent) != string(content) {
		t.Errorf("retrieved content mismatch: got %q, want %q", retrievedContent, content)
	}

	t.Logf("Archive flow: archived %d bytes, retrieved %d bytes, hash: %s",
		len(content), len(retrievedContent), hash[:16])
}

// TestArchiveCompression tests compression and decompression
func TestArchiveCompression(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := archive.ArchiveConfig{
		MaxSize:           10 * 1024 * 1024,
		Expiration:        24 * time.Hour,
		Enabled:           true,
		EnableCompression: true,
		DatabasePath:      filepath.Join(tmpDir, "archive.db"),
	}

	manager, err := archive.NewArchiveManager(cfg)
	if err != nil {
		t.Fatalf("failed to create archive manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}

	// Large content that should compress well
	content := []byte(helpers.GetLargeContent())
	originalSize := int64(len(content))

	entry := &archive.ArchiveEntry{
		OriginalContent: content,
		FilteredContent: content,
		OriginalSize:    originalSize,
		Category:        archive.CategoryCommand,
		Tags:            []string{"large"},
	}

	// Archive with compression
	hash, err := manager.Archive(ctx, entry)
	if err != nil {
		t.Fatalf("failed to archive content: %v", err)
	}

	// Retrieve and verify
	retrieved, err := manager.Retrieve(ctx, hash)
	if err != nil {
		t.Fatalf("failed to retrieve content: %v", err)
	}

	retrievedContent := retrieved.OriginalContent
	if string(retrievedContent) != string(content) {
		t.Error("retrieved content mismatch after compression")
	}

	t.Logf("Compression: original=%d, archived with compression, hash=%s",
		originalSize, hash[:16])
}

// TestArchiveIntegrity tests integrity verification
func TestArchiveIntegrity(t *testing.T) {
	manager, cleanup := setupTestArchiveManager(t)
	defer cleanup()

	ctx := context.Background()

	content := []byte("Test content for integrity verification")
	entry := &archive.ArchiveEntry{
		OriginalContent: content,
		FilteredContent: content,
		OriginalSize:    int64(len(content)),
		Category:        archive.CategoryCommand,
		Tags:            []string{"integrity-test"},
	}

	// Archive
	hash, err := manager.Archive(ctx, entry)
	if err != nil {
		t.Fatalf("failed to archive: %v", err)
	}

	// Verify integrity
	valid, err := manager.Verify(ctx, hash)
	if err != nil {
		t.Fatalf("failed to verify: %v", err)
	}

	if !valid {
		t.Error("integrity verification failed for valid content")
	}

	t.Logf("Integrity: verified hash %s", hash[:16])
}

// TestArchiveList tests listing archives
func TestArchiveList(t *testing.T) {
	manager, cleanup := setupTestArchiveManager(t)
	defer cleanup()

	ctx := context.Background()

	// Create archives with different tags
	contents := []struct {
		content []byte
		tags    []string
	}{
		{[]byte("Go code"), []string{"go", "code"}},
		{[]byte("Python code"), []string{"python", "code"}},
		{[]byte("Log output"), []string{"logs"}},
	}

	for _, c := range contents {
		entry := &archive.ArchiveEntry{
			OriginalContent: c.content,
			FilteredContent: c.content,
			OriginalSize:    int64(len(c.content)),
			Category:        archive.CategoryCommand,
			Tags:            c.tags,
		}
		_, err := manager.Archive(ctx, entry)
		if err != nil {
			t.Fatalf("failed to archive: %v", err)
		}
	}

	// List all archives
	opts := archive.ArchiveListOptions{}
	result, err := manager.List(ctx, opts)
	if err != nil {
		t.Fatalf("failed to list archives: %v", err)
	}

	if len(result.Entries) != len(contents) {
		t.Errorf("expected %d entries, got %d", len(contents), len(result.Entries))
	}

	t.Logf("List: total=%d archives found", len(result.Entries))
}

// TestArchiveStats tests statistics generation
func TestArchiveStats(t *testing.T) {
	manager, cleanup := setupTestArchiveManager(t)
	defer cleanup()

	ctx := context.Background()

	// Create some archives
	for i := 0; i < 5; i++ {
		content := []byte("Test content " + string(rune('0'+i)))
		entry := &archive.ArchiveEntry{
			OriginalContent: content,
			FilteredContent: content,
			OriginalSize:    int64(len(content)),
			Category:        archive.CategoryCommand,
			Tags:            []string{"test"},
		}
		_, err := manager.Archive(ctx, entry)
		if err != nil {
			t.Fatalf("failed to archive: %v", err)
		}
	}

	// Get stats
	stats, err := manager.Stats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalArchives != 5 {
		t.Errorf("expected 5 total archives, got %d", stats.TotalArchives)
	}

	t.Logf("Stats: archives=%d, original=%d bytes", stats.TotalArchives, stats.TotalOriginalSize)
}

// TestArchiveDelete tests archive deletion
func TestArchiveDelete(t *testing.T) {
	manager, cleanup := setupTestArchiveManager(t)
	defer cleanup()

	ctx := context.Background()

	// Create archive
	content := []byte("Content to delete")
	entry := &archive.ArchiveEntry{
		OriginalContent: content,
		FilteredContent: content,
		OriginalSize:    int64(len(content)),
		Category:        archive.CategoryCommand,
		Tags:            []string{"delete-test"},
	}

	hash, err := manager.Archive(ctx, entry)
	if err != nil {
		t.Fatalf("failed to archive: %v", err)
	}

	// Verify it exists
	_, err = manager.Retrieve(ctx, hash)
	if err != nil {
		t.Fatalf("failed to retrieve before delete: %v", err)
	}

	// Delete
	if err := manager.Delete(ctx, hash); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify it's gone
	_, err = manager.Retrieve(ctx, hash)
	if err == nil {
		t.Error("expected error after deletion, but got none")
	}

	t.Logf("Delete: archived and deleted hash %s", hash[:16])
}

// TestArchiveCleanup tests cleanup of expired entries
func TestArchiveCleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager with very short TTL
	cfg := archive.ArchiveConfig{
		MaxSize:      10 * 1024 * 1024,
		Expiration:   1 * time.Millisecond, // Very short for testing
		Enabled:      true,
		DatabasePath: filepath.Join(tmpDir, "archive.db"),
	}

	manager, err := archive.NewArchiveManager(cfg)
	if err != nil {
		t.Fatalf("failed to create archive manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}

	// Create archive
	content := []byte("Expiring content")
	expiresAt := time.Now().Add(1 * time.Millisecond)
	entry := &archive.ArchiveEntry{
		OriginalContent: content,
		FilteredContent: content,
		OriginalSize:    int64(len(content)),
		Category:        archive.CategoryCommand,
		Tags:            []string{"expire-test"},
		ExpiresAt:       &expiresAt,
	}

	hash, err := manager.Archive(ctx, entry)
	if err != nil {
		t.Fatalf("failed to archive: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Run cleanup
	deleted, err := manager.CleanupExpired(ctx)
	if err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}

	if deleted == 0 {
		t.Error("expected at least one expired entry to be cleaned up")
	}

	// Verify it's gone
	_, err = manager.Retrieve(ctx, hash)
	if err == nil {
		t.Error("expected archived content to be cleaned up")
	}

	t.Logf("Cleanup: %d expired entries cleaned up", deleted)
}
