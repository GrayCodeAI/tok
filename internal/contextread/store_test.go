package contextread

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSaveLoadAndNormalize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "read-state.json")

	store := &Store{}
	store.Put(filepath.Join(dir, ".", "file.go"), "package main\n")
	if err := store.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	snap, ok := loaded.Get(filepath.Join(dir, "file.go"))
	if !ok {
		t.Fatal("expected snapshot to exist after load")
	}
	if snap.Fingerprint == "" {
		t.Fatal("expected fingerprint to be populated")
	}
}

func TestLoadMissingFileReturnsEmptyStore(t *testing.T) {
	loaded, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded == nil || len(loaded.Snapshots) != 0 {
		t.Fatalf("expected empty store, got %#v", loaded)
	}
}

func TestStorePrunesOldEntries(t *testing.T) {
	store := &Store{Snapshots: make(map[string]Snapshot)}
	for i := 0; i < maxSnapshots+10; i++ {
		name := filepath.Join(string(os.PathSeparator), "tmp", "tokman", fmt.Sprintf("file-%d", i))
		store.Put(name, "x")
	}
	if len(store.Snapshots) > maxSnapshots {
		t.Fatalf("expected at most %d snapshots, got %d", maxSnapshots, len(store.Snapshots))
	}
}

func TestStoreRenderCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "read-state.json")

	store := &Store{}
	store.PutRender("cache-key", "rendered output", 100, 40)
	if err := store.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	entry, ok := loaded.GetRender("cache-key")
	if !ok {
		t.Fatal("expected render cache entry")
	}
	if entry.Output != "rendered output" {
		t.Fatalf("unexpected output %q", entry.Output)
	}
}
