package rewind

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "rewind-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	cfg := Config{
		DatabasePath: filepath.Join(tmpDir, "rewind_test.db"),
		MaxSize:      10 * 1024 * 1024,
		TTL:          24 * time.Hour,
		Enabled:      true,
	}

	store, err := New(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{"empty", "", 16},
		{"short", "hello", 16},
		{"long", "this is a much longer string with more content", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GenerateHash(tt.input)
			if len(hash) != tt.wantLen {
				t.Errorf("hash length = %d, want %d", len(hash), tt.wantLen)
			}
		})
	}

	// Same input should produce same hash
	h1 := GenerateHash("test")
	h2 := GenerateHash("test")
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %s vs %s", h1, h2)
	}

	// Different input should produce different hash
	h3 := GenerateHash("different")
	if h1 == h3 {
		t.Error("different inputs produced same hash")
	}
}

func TestStoreSaveAndRetrieve(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	entry := Entry{
		Command:        "git",
		Args:           "status",
		OriginalOutput: "On branch main\nChanges not staged:\n  modified: file.go\n",
		FilteredOutput: "M file.go",
		OriginalTokens: 20,
		FilteredTokens: 3,
	}

	// Save
	hash, err := store.Save(entry)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if hash == "" {
		t.Fatal("save returned empty hash")
	}

	// Retrieve
	retrieved, err := store.Retrieve(hash)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if retrieved.Command != entry.Command {
		t.Errorf("command = %q, want %q", retrieved.Command, entry.Command)
	}
	if retrieved.OriginalOutput != entry.OriginalOutput {
		t.Errorf("original output mismatch")
	}
	if retrieved.FilteredOutput != entry.FilteredOutput {
		t.Errorf("filtered output mismatch")
	}
	if retrieved.TokensSaved != 17 {
		t.Errorf("tokens saved = %d, want 17", retrieved.TokensSaved)
	}
	if retrieved.CompressionPct < 80 {
		t.Errorf("compression = %.1f%%, expected > 80%%", retrieved.CompressionPct)
	}
}

func TestStoreRetrieveNotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	_, err := store.Retrieve("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}

func TestStoreList(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Save multiple entries
	for i := 0; i < 5; i++ {
		entry := Entry{
			Command:        "git",
			Args:           "status",
			OriginalOutput: "output " + string(rune('A'+i)),
			FilteredOutput: "filtered",
			OriginalTokens: 100,
			FilteredTokens: 10,
		}
		if _, err := store.Save(entry); err != nil {
			t.Fatalf("save entry %d: %v", i, err)
		}
	}

	// List all
	entries, err := store.List(10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("list returned %d entries, want 5", len(entries))
	}

	// List limited
	entries, err = store.List(3)
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("list limited returned %d entries, want 3", len(entries))
	}
}

func TestStoreDelete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	entry := Entry{
		Command:        "git",
		Args:           "diff",
		OriginalOutput: "some diff output",
		FilteredOutput: "filtered",
		OriginalTokens: 50,
		FilteredTokens: 5,
	}

	hash, err := store.Save(entry)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Delete
	if err := store.Delete(hash); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Verify deleted
	_, err = store.Retrieve(hash)
	if err == nil {
		t.Fatal("expected error after delete")
	}

	// Delete non-existent
	err = store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}

func TestStoreGetStats(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Empty stats
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("total entries = %d, want 0", stats.TotalEntries)
	}

	// Add entries
	for i := 0; i < 3; i++ {
		entry := Entry{
			Command:        "test",
			Args:           "cmd",
			OriginalOutput: "original output " + string(rune('A'+i)),
			FilteredOutput: "filtered",
			OriginalTokens: 100,
			FilteredTokens: 20,
		}
		store.Save(entry)
	}

	stats, err = store.GetStats()
	if err != nil {
		t.Fatalf("get stats after save: %v", err)
	}
	if stats.TotalEntries != 3 {
		t.Errorf("total entries = %d, want 3", stats.TotalEntries)
	}
	if stats.TotalOriginal != 300 {
		t.Errorf("total original = %d, want 300", stats.TotalOriginal)
	}
	if stats.TotalSaved != 240 {
		t.Errorf("total saved = %d, want 240", stats.TotalSaved)
	}
}

func TestStoreSearch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	entries := []Entry{
		{Command: "git", Args: "status", OriginalOutput: "a", FilteredOutput: "a", OriginalTokens: 10, FilteredTokens: 2},
		{Command: "git", Args: "diff", OriginalOutput: "b", FilteredOutput: "b", OriginalTokens: 10, FilteredTokens: 2},
		{Command: "go", Args: "test", OriginalOutput: "c", FilteredOutput: "c", OriginalTokens: 10, FilteredTokens: 2},
		{Command: "docker", Args: "ps", OriginalOutput: "d", FilteredOutput: "d", OriginalTokens: 10, FilteredTokens: 2},
	}

	for _, e := range entries {
		store.Save(e)
	}

	// Search for git commands
	results, err := store.Search("git")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("search 'git' returned %d results, want 2", len(results))
	}

	// Search for go commands
	results, err = store.Search("go")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("search 'go' returned %d results, want 1", len(results))
	}
}

func TestStorePrune(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	entry := Entry{
		Command:        "test",
		Args:           "cmd",
		OriginalOutput: "output",
		FilteredOutput: "filtered",
		OriginalTokens: 10,
		FilteredTokens: 2,
	}

	store.Save(entry)

	// With normal TTL (24h), nothing should be pruned
	pruned, err := store.Prune()
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned != 0 {
		t.Errorf("pruned = %d, want 0 (entry is recent)", pruned)
	}

	// Set TTL to 1ns so everything is "expired" relative to now
	store.ttl = time.Nanosecond

	// Wait a moment so the entry is definitely older than 1ns
	time.Sleep(50 * time.Millisecond)

	pruned, err = store.Prune()
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned != 1 {
		t.Errorf("pruned = %d, want 1", pruned)
	}

	// Verify empty
	entries, _ := store.List(10)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after prune, got %d", len(entries))
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxSize != 100*1024*1024 {
		t.Errorf("max size = %d, want 100MB", cfg.MaxSize)
	}
	if cfg.TTL != 7*24*time.Hour {
		t.Errorf("TTL = %v, want 7 days", cfg.TTL)
	}
	if !cfg.Enabled {
		t.Error("expected enabled by default")
	}
	if cfg.DatabasePath == "" {
		t.Error("expected non-empty database path")
	}
}

func TestNewDisabled(t *testing.T) {
	cfg := Config{Enabled: false}
	store, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store != nil {
		t.Error("expected nil store when disabled")
	}
}

func TestConcurrentAccess(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Concurrent saves
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			entry := Entry{
				Command:        "test",
				Args:           "concurrent",
				OriginalOutput: "output " + string(rune('A'+i)),
				FilteredOutput: "filtered",
				OriginalTokens: 100,
				FilteredTokens: 10,
			}
			store.Save(entry)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries
	entries, err := store.List(20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 10 {
		t.Errorf("expected 10 entries, got %d", len(entries))
	}
}
