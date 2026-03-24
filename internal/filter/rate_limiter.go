package filter

import (
	"sync"
	"time"
)

// RateLimiter prevents abuse of expensive compression layers.
// Limits the number of compressions per time window to prevent
// resource exhaustion and ensure fair usage.
type RateLimiter struct {
	config    RateLimitConfig
	counters  map[string]*rateCounter
	mu        sync.RWMutex
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Enabled controls whether rate limiting is active
	Enabled bool

	// MaxCompressionsPerMinute limits total compressions
	MaxCompressionsPerMinute int

	// MaxCompressionsPerLayer limits per-layer compressions
	MaxCompressionsPerLayer int

	// WindowSize is the time window for rate limiting
	WindowSize time.Duration
}

// rateCounter tracks rate limiting state
type rateCounter struct {
	count    int
	windowStart time.Time
}

// DefaultRateLimitConfig returns default configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:                  true,
		MaxCompressionsPerMinute: 1000,
		MaxCompressionsPerLayer:  200,
		WindowSize:               time.Minute,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return NewRateLimiterWithConfig(DefaultRateLimitConfig())
}

// NewRateLimiterWithConfig creates a rate limiter with custom config
func NewRateLimiterWithConfig(cfg RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:   cfg,
		counters: make(map[string]*rateCounter),
	}
}

// Allow checks if a compression request is allowed
func (r *RateLimiter) Allow(key string) bool {
	if !r.config.Enabled {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	counter := r.counters[key]

	if counter == nil {
		r.counters[key] = &rateCounter{count: 1, windowStart: now}
		return true
	}

	// Check if window has expired
	if now.Sub(counter.windowStart) > r.config.WindowSize {
		counter.count = 1
		counter.windowStart = now
		return true
	}

	// Check limit
	if counter.count >= r.config.MaxCompressionsPerLayer {
		return false
	}

	counter.count++
	return true
}

// AllowGlobal checks if a global compression request is allowed
func (r *RateLimiter) AllowGlobal() bool {
	return r.Allow("global")
}

// Reset clears all rate limit counters
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters = make(map[string]*rateCounter)
}

// GetCount returns the current count for a key
func (r *RateLimiter) GetCount(key string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if counter, ok := r.counters[key]; ok {
		if time.Since(counter.windowStart) <= r.config.WindowSize {
			return counter.count
		}
	}
	return 0
}
