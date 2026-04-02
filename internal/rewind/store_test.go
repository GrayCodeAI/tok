package rewind

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRewindStore(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	store := NewRewindStore(db)
	if err := store.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	markerID, err := store.Store("original content", "compressed")
	if err != nil {
		t.Fatalf("Store error: %v", err)
	}
	if markerID == "" {
		t.Error("Expected non-empty marker ID")
	}

	retrieved, err := store.Retrieve(markerID)
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	if retrieved != "original content" {
		t.Errorf("Expected 'original content', got %s", retrieved)
	}

	size, _ := store.Size()
	if size != 1 {
		t.Errorf("Expected size 1, got %d", size)
	}

	stats, _ := store.Stats()
	if stats["count"].(int) != 1 {
		t.Errorf("Expected count 1, got %v", stats["count"])
	}
}

func TestRewindStoreDedup(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()

	store := NewRewindStore(db)
	store.Init()

	m1, _ := store.Store("same content", "compressed")
	m2, _ := store.Store("same content", "compressed")

	if m1 != m2 {
		t.Error("Expected same marker ID for duplicate content")
	}

	size, _ := store.Size()
	if size != 1 {
		t.Errorf("Expected size 1 for deduped content, got %d", size)
	}
}
