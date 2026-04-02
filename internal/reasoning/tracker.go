package reasoning

import (
	"sync"
)

type ReasoningToken struct {
	Model  string `json:"model"`
	Tokens int64  `json:"tokens"`
	Reason string `json:"reason"`
}

type ReasoningTracker struct {
	tokens map[string]int64
	mu     sync.RWMutex
}

func NewReasoningTracker() *ReasoningTracker {
	return &ReasoningTracker{
		tokens: make(map[string]int64),
	}
}

func (t *ReasoningTracker) Record(model string, tokens int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokens[model] += tokens
}

func (t *ReasoningTracker) GetTotal(model string) int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tokens[model]
}

func (t *ReasoningTracker) GetAll() map[string]int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make(map[string]int64)
	for k, v := range t.tokens {
		result[k] = v
	}
	return result
}
