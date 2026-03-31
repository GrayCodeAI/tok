package core

import (
	"sync"
	"time"
)

type TurnRecord struct {
	Timestamp    time.Time
	Command      string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Model        string
	Provider     string
}

type CostTracker struct {
	mu    sync.RWMutex
	turns []TurnRecord
}

func NewCostTracker() *CostTracker {
	return &CostTracker{turns: make([]TurnRecord, 0, 1000)}
}

func (t *CostTracker) RecordTurn(cmd string, inputTokens, outputTokens int, cost float64, model, provider string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.turns = append(t.turns, TurnRecord{
		Timestamp:    time.Now(),
		Command:      cmd,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Cost:         cost,
		Model:        model,
		Provider:     provider,
	})
	if len(t.turns) > 10000 {
		t.turns = t.turns[len(t.turns)-5000:]
	}
}

func (t *CostTracker) GetTurns(limit int) []TurnRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()
	n := limit
	if n > len(t.turns) {
		n = len(t.turns)
	}
	result := make([]TurnRecord, n)
	copy(result, t.turns[len(t.turns)-n:])
	return result
}

func (t *CostTracker) GetTotalCost() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var total float64
	for _, tr := range t.turns {
		total += tr.Cost
	}
	return total
}

func (t *CostTracker) GetCostByProvider() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	byProvider := make(map[string]float64)
	for _, tr := range t.turns {
		byProvider[tr.Provider] += tr.Cost
	}
	return byProvider
}

func (t *CostTracker) GetCostByModel() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	byModel := make(map[string]float64)
	for _, tr := range t.turns {
		byModel[tr.Model] += tr.Cost
	}
	return byModel
}
