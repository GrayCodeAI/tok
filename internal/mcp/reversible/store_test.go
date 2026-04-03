package reversible

import (
	"path/filepath"
	"testing"
)

func TestNewSQLiteStoreRequiresPath(t *testing.T) {
	_, err := NewSQLiteStore(Config{})
	if err == nil {
		t.Fatal("expected error for empty store path")
	}
}

func TestSQLiteStoreSaveDefaults(t *testing.T) {
	dir := t.TempDir()
	store, err := NewSQLiteStore(Config{
		StorePath:        filepath.Join(dir, "reversible.db"),
		MaxEntrySize:     1024 * 1024,
		DefaultAlgorithm: "zstd",
		AutoVacuum:       false,
	})
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	hash, err := store.Save(&Entry{Original: "hello world"})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash")
	}

	entry, err := store.Retrieve(hash)
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}
	if entry.Original != "hello world" {
		t.Fatalf("retrieved original = %q", entry.Original)
	}
	if entry.CompressionAlg == "" {
		t.Fatal("expected compression algorithm default to be set")
	}
}
