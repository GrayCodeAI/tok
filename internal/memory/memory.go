package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MemoryStore provides cross-session memory for AI agents.
// Inspired by lean-ctx's CCP and claw-compactor's Engram system.
type MemoryStore struct {
	mu        sync.RWMutex
	tasks     []MemoryItem
	findings  []MemoryItem
	decisions []MemoryItem
	facts     []MemoryItem
	storePath string
}

// MemoryItem represents a single memory entry.
type MemoryItem struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Priority  int       `json:"priority"`
}

// NewMemoryStore creates a new memory store.
func NewMemoryStore(storePath string) *MemoryStore {
	ms := &MemoryStore{
		storePath: storePath,
	}
	ms.load()
	return ms
}

// AddTask adds a task to memory.
func (ms *MemoryStore) AddTask(content string, tags ...string) string {
	item := MemoryItem{
		ID:        generateID(),
		Content:   content,
		Category:  "task",
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Priority:  1,
	}
	ms.mu.Lock()
	ms.tasks = append(ms.tasks, item)
	ms.mu.Unlock()
	ms.save()
	return item.ID
}

// AddFinding adds a finding to memory.
func (ms *MemoryStore) AddFinding(content string, tags ...string) string {
	item := MemoryItem{
		ID:        generateID(),
		Content:   content,
		Category:  "finding",
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Priority:  2,
	}
	ms.mu.Lock()
	ms.findings = append(ms.findings, item)
	ms.mu.Unlock()
	ms.save()
	return item.ID
}

// AddDecision adds a decision to memory.
func (ms *MemoryStore) AddDecision(content string, tags ...string) string {
	item := MemoryItem{
		ID:        generateID(),
		Content:   content,
		Category:  "decision",
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Priority:  3,
	}
	ms.mu.Lock()
	ms.decisions = append(ms.decisions, item)
	ms.mu.Unlock()
	ms.save()
	return item.ID
}

// AddFact adds a fact to memory.
func (ms *MemoryStore) AddFact(content string, tags ...string) string {
	item := MemoryItem{
		ID:        generateID(),
		Content:   content,
		Category:  "fact",
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Priority:  1,
	}
	ms.mu.Lock()
	ms.facts = append(ms.facts, item)
	ms.mu.Unlock()
	ms.save()
	return item.ID
}

// Query searches memory by category and tags.
func (ms *MemoryStore) Query(category string, tags ...string) []MemoryItem {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []MemoryItem
	var items []MemoryItem

	switch category {
	case "task":
		items = ms.tasks
	case "finding":
		items = ms.findings
	case "decision":
		items = ms.decisions
	case "fact":
		items = ms.facts
	default:
		items = append(append(append(ms.tasks, ms.findings...), ms.decisions...), ms.facts...)
	}

	for _, item := range items {
		if len(tags) == 0 {
			results = append(results, item)
			continue
		}
		for _, tag := range tags {
			for _, itemTag := range item.Tags {
				if strings.Contains(strings.ToLower(itemTag), strings.ToLower(tag)) {
					results = append(results, item)
					break
				}
			}
		}
	}

	return results
}

// Clear clears all memory.
func (ms *MemoryStore) Clear() {
	ms.mu.Lock()
	ms.tasks = nil
	ms.findings = nil
	ms.decisions = nil
	ms.facts = nil
	ms.mu.Unlock()
	ms.save()
}

// Stats returns memory statistics.
func (ms *MemoryStore) Stats() map[string]int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return map[string]int{
		"tasks":     len(ms.tasks),
		"findings":  len(ms.findings),
		"decisions": len(ms.decisions),
		"facts":     len(ms.facts),
		"total":     len(ms.tasks) + len(ms.findings) + len(ms.decisions) + len(ms.facts),
	}
}

func (ms *MemoryStore) load() {
	if ms.storePath == "" {
		return
	}
	data, err := os.ReadFile(ms.storePath)
	if err != nil {
		return
	}
	var store struct {
		Tasks     []MemoryItem `json:"tasks"`
		Findings  []MemoryItem `json:"findings"`
		Decisions []MemoryItem `json:"decisions"`
		Facts     []MemoryItem `json:"facts"`
	}
	if err := json.Unmarshal(data, &store); err != nil {
		return
	}
	ms.tasks = store.Tasks
	ms.findings = store.Findings
	ms.decisions = store.Decisions
	ms.facts = store.Facts
}

func (ms *MemoryStore) save() {
	if ms.storePath == "" {
		return
	}
	dir := filepath.Dir(ms.storePath)
	os.MkdirAll(dir, 0755)

	store := struct {
		Tasks     []MemoryItem `json:"tasks"`
		Findings  []MemoryItem `json:"findings"`
		Decisions []MemoryItem `json:"decisions"`
		Facts     []MemoryItem `json:"facts"`
	}{
		Tasks:     ms.tasks,
		Findings:  ms.findings,
		Decisions: ms.decisions,
		Facts:     ms.facts,
	}
	data, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(ms.storePath, data, 0600)
}

var idCounter int

func generateID() string {
	idCounter++
	return "mem_" + itoa(idCounter)
}

func itoa(n int) string {
	return string(rune('0' + n%10))
}
