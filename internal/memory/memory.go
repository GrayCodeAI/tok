package memory

import "time"

type Memory struct{}

func New() *Memory {
	return &Memory{}
}

type MemoryItem struct {
	Category  string
	Content   string
	Tags      []string
	CreatedAt time.Time
}

type MemoryStore struct{}

func NewMemoryStore(path string) *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) Get(key string) (string, bool) {
	return "", false
}

func (m *MemoryStore) Set(key, value string) error {
	return nil
}

func (m *MemoryStore) AddTask(content string, tags ...string) string {
	return "task-1"
}

func (m *MemoryStore) AddFinding(content string, tags ...string) string {
	return "finding-1"
}

func (m *MemoryStore) AddDecision(content string, tags ...string) string {
	return "decision-1"
}

func (m *MemoryStore) AddFact(content string, tags ...string) string {
	return "fact-1"
}

func (m *MemoryStore) Query(category string, tags ...string) []MemoryItem {
	return nil
}

func (m *MemoryStore) Stats() map[string]int {
	return map[string]int{}
}

func (m *MemoryStore) Clear() error {
	return nil
}
