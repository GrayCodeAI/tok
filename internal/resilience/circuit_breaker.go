// Package resilience provides resilience patterns for TokMan.
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ErrTooManyRequests is returned when too many requests are waiting
var ErrTooManyRequests = errors.New("too many requests waiting")

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker struct {
	mu          sync.RWMutex
	name        string
	state       State
	failures    int
	successes   int
	lastFailure time.Time

	// Configuration
	maxFailures    int
	successesReset int
	timeout        time.Duration
	maxRequests    int
	requestQueue   int

	// Callbacks
	onStateChange func(State, State)
}

// CircuitBreakerOption is a function that modifies the circuit breaker
type CircuitBreakerOption func(*CircuitBreaker)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, opts ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:         name,
		state:        StateClosed,
		maxFailures:  5,
		timeout:      30 * time.Second,
		maxRequests:  100,
		requestQueue: 10,
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// WithMaxFailures sets the maximum failures before opening the circuit
func WithMaxFailures(n int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.maxFailures = n
	}
}

// WithTimeout sets the timeout before attempting to close the circuit
func WithTimeout(d time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.timeout = d
	}
}

// WithMaxRequests sets the maximum concurrent requests
func WithMaxRequests(n int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.maxRequests = n
	}
}

// WithRequestQueue sets the maximum requests waiting for execution
func WithRequestQueue(n int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.requestQueue = n
	}
}

// WithStateChangeCallback sets the callback for state changes
func WithStateChangeCallback(fn func(State, State)) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.onStateChange = fn
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.canExecute(ctx); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.recordResult(err)

	return err
}

// ExecuteWithResult runs the given function with circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	if err := cb.canExecute(ctx); err != nil {
		return nil, err
	}

	// Execute the function
	result, err := fn()

	// Record the result
	cb.recordResult(err)

	return result, err
}

func (cb *CircuitBreaker) canExecute(ctx context.Context) error {
	cb.mu.RLock()
	state := cb.state
	timeout := cb.timeout
	lastFailure := cb.lastFailure
	cb.mu.RUnlock()

	// Check if circuit is open
	if state == StateOpen {
		// Check if timeout has passed
		if time.Since(lastFailure) > timeout {
			// Try to transition to half-open
			cb.mu.Lock()
			if cb.state == StateOpen && time.Since(cb.lastFailure) > timeout {
				cb.setState(StateHalfOpen)
			}
			cb.mu.Unlock()
		} else {
			return ErrCircuitOpen
		}
	}

	// Check request queue
	if cb.requestQueue > 0 {
		// Simplified - could add proper counting
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		// Success
		cb.successes++

		// Check if circuit can be closed
		if cb.state == StateHalfOpen && cb.successes >= cb.successesReset {
			cb.setState(StateClosed)
			cb.failures = 0
			cb.successes = 0
		}
	} else {
		// Failure
		cb.failures++
		cb.lastFailure = time.Now()

		// Check if circuit should be opened
		if cb.state == StateClosed && cb.failures >= cb.maxFailures {
			cb.setState(StateOpen)
		} else if cb.state == StateHalfOpen {
			cb.setState(StateOpen)
		}
	}
}

func (cb *CircuitBreaker) setState(newState State) {
	oldState := cb.state
	if oldState != newState {
		cb.state = newState
		if cb.onStateChange != nil {
			cb.onStateChange(oldState, newState)
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastFailure = time.Time{}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateOpen)
}

// ForceClosed forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClosed() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
}

// GetState returns the current state as a string
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
