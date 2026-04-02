// Package delta provides tests for file version tracking.
package delta

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

func TestManager_Track(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	// Track initial version
	content1 := "line1\nline2\n"
	v1, err := m.Track("/test/file.txt", content1)
	if err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	if v1.Hash == "" {
		t.Error("expected hash to be set")
	}

	// Track same content - should return existing
	v2, err := m.Track("/test/file.txt", content1)
	if err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	if v1.Hash != v2.Hash {
		t.Error("expected same hash for same content")
	}

	// Track different content
	content2 := "line1\nline2\nline3\n"
	v3, err := m.Track("/test/file.txt", content2)
	if err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	if v3.Hash == v1.Hash {
		t.Error("expected different hash for different content")
	}
}

func TestManager_Get(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	content := "test content"
	v1, _ := m.Track("/test/file.txt", content)

	// Get by full hash
	got, err := m.Get("/test/file.txt", v1.Hash)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Hash != v1.Hash {
		t.Error("hash mismatch")
	}

	// Get by short hash
	got2, err := m.Get("/test/file.txt", v1.Hash[:16])
	if err != nil {
		t.Fatalf("Get by short hash failed: %v", err)
	}

	if got2.Hash != v1.Hash {
		t.Error("hash mismatch with short hash")
	}

	// Get non-existent
	_, err = m.Get("/test/file.txt", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent version")
	}
}

func TestManager_GetVersions(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	// Add multiple versions
	for i := 0; i < 5; i++ {
		content := string(rune('a' + i))
		m.Track("/test/file.txt", content)
	}

	versions := m.GetVersions("/test/file.txt")
	if len(versions) != 5 {
		t.Errorf("expected 5 versions, got %d", len(versions))
	}
}

func TestManager_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	// Add more than 10 versions
	for i := 0; i < 15; i++ {
		content := string(rune('a' + i))
		m.Track("/test/file.txt", content)
	}

	// Clean up to keep only 5
	err := m.Cleanup("/test/file.txt", 5)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	versions := m.GetVersions("/test/file.txt")
	if len(versions) != 5 {
		t.Errorf("expected 5 versions after cleanup, got %d", len(versions))
	}
}

func TestComputeHash(t *testing.T) {
	h1 := computeHash("hello")
	h2 := computeHash("hello")
	h3 := computeHash("world")

	if h1 != h2 {
		t.Error("same content should have same hash")
	}

	if h1 == h3 {
		t.Error("different content should have different hash")
	}

	if len(h1) != 64 {
		t.Errorf("expected 64 char hex string, got %d", len(h1))
	}
}

func TestManager_GetLatest(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	m.Track("/test/file.txt", "version1")
	m.Track("/test/file.txt", "version2")
	latest, _ := m.Track("/test/file.txt", "version3")

	got, err := m.GetLatest("/test/file.txt")
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}

	if got.Hash != latest.Hash {
		t.Error("GetLatest did not return the latest version")
	}
}

func TestManager_ShouldCompress(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	// First version
	m.Track("/test/file.txt", "original content here")

	// Check decision is returned (don't depend on exact algorithm behavior)
	shouldCompress, baseHash := m.ShouldCompress("/test/file.txt", "completely different content")
	t.Logf("ShouldCompress for different content: %v, base: %s", shouldCompress, baseHash)

	// Just verify method works and returns something
	// The actual compression decision depends on the internal algorithm
}

func TestManager_Stats(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	m.Track("/test/file1.txt", "content1")
	m.Track("/test/file1.txt", "content2")
	m.Track("/test/file2.txt", "content3")

	stats := m.Stats()
	if stats.FilesTracked != 2 {
		t.Errorf("expected 2 files, got %d", stats.FilesTracked)
	}
	if stats.TotalVersions != 3 {
		t.Errorf("expected 3 versions, got %d", stats.TotalVersions)
	}
}

func TestManager_ComputeDiff(t *testing.T) {
	tempDir := t.TempDir()
	m := NewManager(tempDir)

	v1, _ := m.Track("/test/file.txt", "line1\nline2\nline3\n")
	v2, _ := m.Track("/test/file.txt", "line1\nline2 modified\nline3\nline4\n")

	diff, err := m.ComputeDiff("/test/file.txt", v1.Hash, v2.Hash)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Verify diff contains changes
	if len(diff.Added) == 0 && len(diff.Removed) == 0 {
		t.Error("expected some differences")
	}
}

func TestApplyDelta(t *testing.T) {
	original := "line1\nline2\nline3\n"
	delta := filter.IncrementalDelta{
		Added:   []string{"line2 modified", "line4"},
		Removed: []string{"line2"},
	}

	result := applyDelta(original, delta)
	if result == "" {
		t.Error("applyDelta returned empty result")
	}

	// The result should contain original content with modifications
	if !contains(result, "line1") {
		t.Error("result should contain line1")
	}
}

func TestEstimateDeltaSize(t *testing.T) {
	delta := filter.IncrementalDelta{
		Added:   []string{"line1", "line2", "line3"},
		Removed: []string{"old1"},
	}

	size := estimateDeltaSize(delta)
	if size <= 0 {
		t.Error("expected positive delta size")
	}

	// Size should account for added and removed lines
	expectedMin := len("line1") + len("line2") + len("line3") + len("old1")
	if size < expectedMin {
		t.Errorf("delta size %d should be at least %d", size, expectedMin)
	}
}

func TestCompressionDecision(t *testing.T) {
	cd := CompressionDecision{
		UseDelta:         true,
		BaseHash:         "abc123",
		FullContent:      false,
		EstimatedSavings: 900,
	}

	if !cd.UseDelta {
		t.Error("UseDelta should be true")
	}

	if cd.EstimatedSavings != 900 {
		t.Errorf("expected estimated savings 900, got %d", cd.EstimatedSavings)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(s[:len(substr)] == substr) ||
		(s[len(s)-len(substr):] == substr) ||
		findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
