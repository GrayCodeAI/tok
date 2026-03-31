package memory

import (
	"path/filepath"
	"testing"
)

func TestNewMemoryStore(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestMemoryStore_AddTask(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	id := store.AddTask("test task", "tag1")
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestMemoryStore_AddFinding(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	id := store.AddFinding("test finding", "security")
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestMemoryStore_AddDecision(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	id := store.AddDecision("test decision", "architecture")
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestMemoryStore_AddFact(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	id := store.AddFact("test fact", "general")
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestMemoryStore_Query(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	store.AddTask("task 1", "tag1")
	store.AddTask("task 2", "tag2")

	items := store.Query("task")
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestMemoryStore_Stats(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	store.AddTask("task 1")
	store.AddFinding("finding 1")

	stats := store.Stats()
	if stats["total"] != 2 {
		t.Errorf("expected 2 total items, got %d", stats["total"])
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	dir := t.TempDir()
	store := NewMemoryStore(filepath.Join(dir, "memory.json"))
	store.AddTask("task 1")
	store.Clear()

	stats := store.Stats()
	if stats["total"] != 0 {
		t.Errorf("expected 0 items after clear, got %d", stats["total"])
	}
}

func TestMemoryStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.json")

	store := NewMemoryStore(path)
	store.AddTask("persistent task")

	// Create new store from same path
	store2 := NewMemoryStore(path)
	stats := store2.Stats()
	if stats["total"] != 1 {
		t.Errorf("expected 1 item after reload, got %d", stats["total"])
	}
}
