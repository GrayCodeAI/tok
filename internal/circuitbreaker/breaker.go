// Package circuitbreaker implements the circuit breaker pattern to prevent
// cascading failures when external services (LLM, gRPC, SQLite) degrade.
//
// States: Closed (normal) → Open (failing) → HalfOpen (testing) → Closed
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit breaker is open and requests are rejected.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current state of a circuit breaker.
type State int

const (
	StateClosed   State = iota // Normal operation, requests pass through
	StateOpen                  // Service failing, requests rejected immediately
	StateHalfOpen              // Testing if service has recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "halfopen"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker parameters.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int
	// RecoveryTimeout is how long the circuit stays open before transitioning to HalfOpen.
	RecoveryTimeout time.Duration
	// SuccessThreshold is the number of consecutive successes in HalfOpen before closing.
	SuccessThreshold int
}

// DefaultConfig returns a production-ready configuration.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		SuccessThreshold: 3,
	}
}

// Breaker implements the circuit breaker pattern with thread-safe state transitions.
type Breaker struct {
	mu          sync.RWMutex
	state       State
	cfg         Config
	failures    int
	successes   int
	lastFailure time.Time
	allowed     chan struct{} // semaphore for HalfOpen testing
}

// New creates a circuit breaker with the given configuration.
func New(cfg Config) *Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.RecoveryTimeout <= 0 {
		cfg.RecoveryTimeout = 30 * time.Second
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 3
	}
	return &Breaker{
		cfg:     cfg,
		state:   StateClosed,
		allowed: make(chan struct{}, 1),
	}
}

// Allow checks if a request should be allowed through.
// Returns ErrOpen if the circuit is open and recovery timeout hasn't elapsed.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if recovery timeout has elapsed
		if time.Since(b.lastFailure) > b.cfg.RecoveryTimeout {
			b.state = StateHalfOpen
			b.successes = 0
			return nil
		}
		return ErrOpen

	case StateHalfOpen:
		// Allow one request through at a time for testing
		select {
		case b.allowed <- struct{}{}:
			return nil
		default:
			return ErrOpen
		}
	default:
		return ErrOpen
	}
}

// RecordSuccess records a successful operation.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateHalfOpen:
		b.successes++
		if b.successes >= b.cfg.SuccessThreshold {
			b.state = StateClosed
			b.failures = 0
			b.successes = 0
		}
	case StateClosed:
		b.failures = 0 // Reset failure counter on success
	}
}

// RecordFailure records a failed operation.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		b.failures++
		if b.failures >= b.cfg.FailureThreshold {
			b.state = StateOpen
		}
	case StateHalfOpen:
		// Any failure in HalfOpen immediately reopens the circuit
		b.state = StateOpen
		b.successes = 0
	}
}

// State returns the current circuit breaker state.
func (b *Breaker) State() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Reset manually resets the circuit breaker to closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = StateClosed
	b.failures = 0
	b.successes = 0
}
