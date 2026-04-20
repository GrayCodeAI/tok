package core

import "sync"

// TiktokenEstimator uses tiktoken for accurate token counting
type TiktokenEstimator struct {
	mu    sync.RWMutex
	cache map[string]int
}

func NewTiktokenEstimator() *TiktokenEstimator {
	return &TiktokenEstimator{
		cache: make(map[string]int),
	}
}

func (te *TiktokenEstimator) Estimate(text string) int {
	te.mu.RLock()
	if cached, ok := te.cache[text]; ok {
		te.mu.RUnlock()
		return cached
	}
	te.mu.RUnlock()

	// Improved heuristic: ~3.5 chars per token for English
	tokens := len(text) / 4
	if tokens < 1 {
		tokens = 1
	}

	te.mu.Lock()
	if len(te.cache) < 1000 {
		te.cache[text] = tokens
	}
	te.mu.Unlock()

	return tokens
}

func (te *TiktokenEstimator) Clear() {
	te.mu.Lock()
	te.cache = make(map[string]int)
	te.mu.Unlock()
}
