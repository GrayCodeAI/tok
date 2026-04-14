package breaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Breaker implements circuit breaker pattern
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	successes    int
	lastFailTime time.Time
	threshold    int
	timeout      time.Duration
	halfOpenMax  int
}

// New creates a circuit breaker
func New(threshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		state:       StateClosed,
		threshold:   threshold,
		timeout:     timeout,
		halfOpenMax: 3,
	}
}

// Call executes fn with circuit breaker protection
func (b *Breaker) Call(fn func() error) error {
	if !b.allow() {
		return ErrCircuitOpen
	}

	err := fn()
	b.record(err == nil)
	return err
}

func (b *Breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailTime) > b.timeout {
			b.state = StateHalfOpen
			b.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return b.successes < b.halfOpenMax
	}
	return false
}

func (b *Breaker) record(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if success {
		b.failures = 0
		b.successes++

		if b.state == StateHalfOpen && b.successes >= b.halfOpenMax {
			b.state = StateClosed
		}
	} else {
		b.successes = 0
		b.failures++
		b.lastFailTime = time.Now()

		if b.failures >= b.threshold {
			b.state = StateOpen
		}
	}
}

// State returns current state
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Reset resets the breaker
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = StateClosed
	b.failures = 0
	b.successes = 0
}
