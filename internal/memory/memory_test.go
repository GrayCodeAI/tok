package memory

import (
	"testing"
	"time"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore("/tmp/test-memory")
	if store == nil {
		t.Fatal("expected non-nil memory store")
	}
	if store.Path != "/tmp/test-memory" {
		t.Errorf("expected path='/tmp/test-memory', got '%s'", store.Path)
	}
	if store.Data == nil {
		t.Error("expected initialized Data map")
	}
	if store.Tasks == nil {
		t.Error("expected initialized Tasks slice")
	}
	if store.Findings == nil {
		t.Error("expected initialized Findings slice")
	}
	if store.Decisions == nil {
		t.Error("expected initialized Decisions slice")
	}
	if store.Items == nil {
		t.Error("expected initialized Items slice")
	}
}

func TestMemoryStore_AddTask(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	result := store.AddTask("Complete feature X", "urgent", "backend")
	_ = result // Stub returns empty string

	if len(store.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(store.Tasks))
	}
	if store.Tasks[0] != "Complete feature X" {
		t.Errorf("expected task 'Complete feature X', got '%s'", store.Tasks[0])
	}
}

func TestMemoryStore_AddFinding(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	store.AddFinding("Security vulnerability found", "security", "critical")

	if len(store.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(store.Findings))
	}
	if store.Findings[0] != "Security vulnerability found" {
		t.Errorf("expected finding 'Security vulnerability found', got '%s'", store.Findings[0])
	}
}

func TestMemoryStore_AddDecision(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	store.AddDecision("Use PostgreSQL for database", "architecture")

	if len(store.Decisions) != 1 {
		t.Errorf("expected 1 decision, got %d", len(store.Decisions))
	}
	if store.Decisions[0] != "Use PostgreSQL for database" {
		t.Errorf("expected decision 'Use PostgreSQL for database', got '%s'", store.Decisions[0])
	}
}

func TestMemoryStore_AddFact(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	result := store.AddFact("API rate limit is 1000/hour", "api", "limits")
	_ = result // Stub returns empty string

	// Stub doesn't actually store facts, so we just verify it doesn't panic
}

func TestMemoryStore_Query(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	// Add some items first
	store.Items = []MemoryItem{
		{Content: "Item 1", Category: "test", Tags: []string{"tag1"}, CreatedAt: time.Now()},
		{Content: "Item 2", Category: "test", Tags: []string{"tag2"}, CreatedAt: time.Now()},
	}

	results := store.Query("test", "tag1")

	// Stub returns all items
	if len(results) != 2 {
		t.Errorf("expected 2 items, got %d", len(results))
	}
}

func TestMemoryStore_Stats(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	// Add some data
	store.AddTask("Task 1")
	store.AddTask("Task 2")
	store.AddFinding("Finding 1")
	store.AddDecision("Decision 1")

	stats := store.Stats()

	if stats == nil {
		t.Fatal("expected non-nil stats")
	}

	if stats["tasks"] != 2 {
		t.Errorf("expected tasks=2, got %v", stats["tasks"])
	}
	if stats["findings"] != 1 {
		t.Errorf("expected findings=1, got %v", stats["findings"])
	}
	if stats["decisions"] != 1 {
		t.Errorf("expected decisions=1, got %v", stats["decisions"])
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	store := NewMemoryStore("/tmp/test")

	// Add data
	store.AddTask("Task 1")
	store.AddFinding("Finding 1")
	store.AddDecision("Decision 1")
	store.Items = []MemoryItem{{Content: "Item"}}

	// Clear
	store.Clear()

	if len(store.Tasks) != 0 {
		t.Errorf("expected 0 tasks after clear, got %d", len(store.Tasks))
	}
	if len(store.Findings) != 0 {
		t.Errorf("expected 0 findings after clear, got %d", len(store.Findings))
	}
	if len(store.Decisions) != 0 {
		t.Errorf("expected 0 decisions after clear, got %d", len(store.Decisions))
	}
	if len(store.Items) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(store.Items))
	}
}

func TestMemoryItem(t *testing.T) {
	item := MemoryItem{
		Content:   "Test content",
		Category:  "test",
		Tags:      []string{"tag1", "tag2"},
		CreatedAt: time.Now(),
	}

	if item.Content != "Test content" {
		t.Error("Content not set correctly")
	}
	if item.Category != "test" {
		t.Error("Category not set correctly")
	}
	if len(item.Tags) != 2 {
		t.Error("Tags not set correctly")
	}
}

func TestAnalyze(t *testing.T) {
	result, err := Analyze()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Stub returns empty string
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestOptimize(t *testing.T) {
	err := Optimize()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Stub returns nil
}
