package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements token bucket rate limiting
type Limiter struct {
	mu       sync.Mutex
	rate     float64
	capacity int
	tokens   float64
	lastTime time.Time
}

// NewLimiter creates a rate limiter (rate per second, burst capacity)
func NewLimiter(rate float64, capacity int) *Limiter {
	return &Limiter{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		lastTime: time.Now(),
	}
}

// Allow checks if request is allowed
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.capacity) {
		l.tokens = float64(l.capacity)
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until request can proceed
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Global limiter: 100 commands/sec, burst of 200
var globalLimiter = NewLimiter(100, 200)

// CheckGlobal checks global rate limit
func CheckGlobal() bool {
	return globalLimiter.Allow()
}

// WaitGlobal waits for global rate limit
func WaitGlobal(ctx context.Context) error {
	return globalLimiter.Wait(ctx)
}
