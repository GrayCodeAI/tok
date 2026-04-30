package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewQueryCache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	qc, err := NewQueryCache(dbPath)
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	// Verify database was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		args       []string
		workingDir string
		fileHashes map[string]string
		wantDiff   bool
	}{
		{
			name:       "same inputs same key",
			command:    "git",
			args:       []string{"status"},
			workingDir: "/home/user/project",
			fileHashes: map[string]string{"file.go": "abc123"},
			wantDiff:   false,
		},
		{
			name:       "different command different key",
			command:    "npm",
			args:       []string{"test"},
			workingDir: "/home/user/project",
			fileHashes: map[string]string{},
			wantDiff:   true,
		},
	}

	key1 := GenerateKey("git", []string{"status"}, "/home/user/project", map[string]string{"file.go": "abc123"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key2 := GenerateKey(tt.command, tt.args, tt.workingDir, tt.fileHashes)
			if tt.wantDiff && key1 == key2 {
				t.Error("expected different keys, got same")
			}
			if !tt.wantDiff && key1 != key2 {
				t.Error("expected same keys, got different")
			}
		})
	}
}

func TestQueryCache_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	qc, err := NewQueryCache(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	// Set a cache entry
	key := "test-key-123"
	command := "git"
	args := []string{"status"}
	workingDir := "/home/user/project"
	fileHashes := map[string]string{"go.mod": "abc123"}
	filteredOutput := "## main\n M file.go"
	originalTokens := 100
	filteredTokens := 50

	err = qc.Set(key, command, args, workingDir, fileHashes, filteredOutput, originalTokens, filteredTokens)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the entry
	entry, found := qc.Get(key)
	if !found {
		t.Fatal("expected to find entry, got not found")
	}

	if entry.Command != command {
		t.Errorf("expected command %q, got %q", command, entry.Command)
	}

	if entry.FilteredOutput != filteredOutput {
		t.Errorf("expected output %q, got %q", filteredOutput, entry.FilteredOutput)
	}

	if entry.OriginalTokens != originalTokens {
		t.Errorf("expected original tokens %d, got %d", originalTokens, entry.OriginalTokens)
	}

	if entry.FilteredTokens != filteredTokens {
		t.Errorf("expected filtered tokens %d, got %d", filteredTokens, entry.FilteredTokens)
	}

	// Check hit count was incremented
	if entry.HitCount < 1 {
		t.Errorf("expected hit_count >= 1, got %d", entry.HitCount)
	}
}

func TestQueryCache_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	qc, err := NewQueryCache(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	_, found := qc.Get("non-existent-key")
	if found {
		t.Error("expected not found for non-existent key")
	}

	// Check miss counter
	hits, misses := qc.GetRuntimeStats()
	if hits != 0 {
		t.Errorf("expected 0 hits, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}

func TestQueryCache_InvalidateByCommand(t *testing.T) {
	tmpDir := t.TempDir()
	qc, err := NewQueryCache(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	// Add entries for different commands
	qc.Set("key1", "git", []string{"status"}, "/tmp", nil, "output1", 100, 50)
	qc.Set("key2", "npm", []string{"test"}, "/tmp", nil, "output2", 200, 100)
	qc.Set("key3", "git", []string{"log"}, "/tmp", nil, "output3", 150, 75)

	// Invalidate git entries
	err = qc.InvalidateByCommand("git")
	if err != nil {
		t.Fatalf("InvalidateByCommand failed: %v", err)
	}

	// git entries should be gone
	_, found := qc.Get("key1")
	if found {
		t.Error("expected key1 to be invalidated")
	}

	_, found = qc.Get("key3")
	if found {
		t.Error("expected key3 to be invalidated")
	}

	// npm entry should remain
	_, found = qc.Get("key2")
	if !found {
		t.Error("expected key2 to still exist")
	}
}

func TestQueryCache_Stats(t *testing.T) {
	tmpDir := t.TempDir()
	qc, err := NewQueryCache(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	// Add some entries
	qc.Set("key1", "git", []string{"status"}, "/tmp", nil, "output1", 100, 50)
	qc.Set("key2", "npm", []string{"test"}, "/tmp", nil, "output2", 200, 100)

	stats, err := qc.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.TotalEntries != 2 {
		t.Errorf("expected 2 entries, got %d", stats.TotalEntries)
	}

	// Total saved should be (100-50) + (200-100) = 150
	if stats.TotalSaved != 150 {
		t.Errorf("expected 150 total saved, got %d", stats.TotalSaved)
	}
}

func TestQueryCache_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	qc, err := NewQueryCache(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewQueryCache failed: %v", err)
	}
	defer qc.Close()

	// Add entry
	qc.Set("key1", "git", []string{"status"}, "/tmp", nil, "output1", 100, 50)

	// Cleanup with very long max age should remove nothing
	err = qc.Cleanup(24 * time.Hour)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Entry should still exist
	_, found := qc.Get("key1")
	if !found {
		t.Error("expected key1 to still exist after no-op cleanup")
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create temp dir with .git
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.Mkdir(gitDir, 0755)

	if !IsGitRepo(tmpDir) {
		t.Error("expected IsGitRepo to return true for directory with .git")
	}

	// Non-git directory
	nonGitDir := t.TempDir()
	if IsGitRepo(nonGitDir) {
		t.Error("expected IsGitRepo to return false for directory without .git")
	}
}
