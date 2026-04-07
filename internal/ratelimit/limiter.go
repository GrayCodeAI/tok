package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.RWMutex
	requests map[string]*clientLimit
	maxReqs  int
	window   time.Duration
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

func NewLimiter(maxReqs int, window time.Duration) *Limiter {
	return &Limiter{
		requests: make(map[string]*clientLimit),
		maxReqs:  maxReqs,
		window:   window,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cl, exists := l.requests[key]

	if !exists || now.After(cl.resetTime) {
		l.requests[key] = &clientLimit{
			count:     1,
			resetTime: now.Add(l.window),
		}
		return true
	}

	if cl.count >= l.maxReqs {
		return false
	}

	cl.count++
	return true
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.requests, key)
}

func (l *Limiter) GetRemaining(key string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cl, exists := l.requests[key]
	if !exists {
		return l.maxReqs
	}

	remaining := l.maxReqs - cl.count
	if remaining < 0 {
		return 0
	}
	return remaining
}
