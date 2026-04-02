package modelcalllimits

import (
	"sync"
	"time"
)

type ModelCallLimiter struct {
	limits map[string]int
	counts map[string]int
	mu     sync.RWMutex
}

func NewModelCallLimiter() *ModelCallLimiter {
	return &ModelCallLimiter{
		limits: make(map[string]int),
		counts: make(map[string]int),
	}
}

func (l *ModelCallLimiter) SetLimit(model string, limit int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limits[model] = limit
}

func (l *ModelCallLimiter) RecordCall(model string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limit, hasLimit := l.limits[model]
	if !hasLimit {
		l.counts[model]++
		return true
	}

	if l.counts[model] >= limit {
		return false
	}

	l.counts[model]++
	return true
}

func (l *ModelCallLimiter) CheckLimit(model string) (bool, int, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	limit, hasLimit := l.limits[model]
	if !hasLimit {
		return true, 0, 0
	}

	count := l.counts[model]
	return count < limit, count, limit
}

func (l *ModelCallLimiter) Reset(model string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[model] = 0
}

func (l *ModelCallLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts = make(map[string]int)
}

func (l *ModelCallLimiter) GetUsage() map[string]int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make(map[string]int)
	for k, v := range l.counts {
		result[k] = v
	}
	return result
}

type RateLimitTracker struct {
	events map[string][]time.Time
	mu     sync.RWMutex
}

func NewRateLimitTracker() *RateLimitTracker {
	return &RateLimitTracker{
		events: make(map[string][]time.Time),
	}
}

func (t *RateLimitTracker) Record(model string, statusCode int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if statusCode == 429 {
		t.events[model] = append(t.events[model], time.Now())
		if len(t.events[model]) > 100 {
			t.events[model] = t.events[model][1:]
		}
	}
}

func (t *RateLimitTracker) GetCount(model string, window time.Duration) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := t.events[model]
	cutoff := time.Now().Add(-window)
	count := 0
	for _, ts := range events {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}

func (t *RateLimitTracker) GetAll() map[string]int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make(map[string]int)
	for model, events := range t.events {
		result[model] = len(events)
	}
	return result
}
