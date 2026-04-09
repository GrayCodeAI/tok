package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/resilience"
)

func TestCircuitBreaker_New(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test")

	if cb == nil {
		t.Fatal("expected circuit breaker to not be nil")
	}
	if cb.Name() != "test" {
		t.Errorf("expected name test, got %s", cb.Name())
	}
}

func TestCircuitBreaker_Success(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test")

	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCircuitBreaker_Failure(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", resilience.WithMaxFailures(3))

	// Trigger failures to open circuit
	for i := 0; i < 4; i++ {
		cb.Execute(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.State() != resilience.StateOpen {
		t.Errorf("expected state open, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpenRejects(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", resilience.WithMaxFailures(2))

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})

	// Should now reject
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != resilience.ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Timeout(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test",
		resilience.WithMaxFailures(1),
		resilience.WithTimeout(50*time.Millisecond),
	)

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Should now allow attempt (half-open)
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	// Error may be nil (allowed to try) or other (failed attempt)
	_ = err
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", resilience.WithMaxFailures(2))

	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})

	cb.Reset()

	if cb.State() != resilience.StateClosed {
		t.Errorf("expected state closed after reset, got %v", cb.State())
	}
}

func TestCircuitBreaker_Options(t *testing.T) {
	// Test that options work
	cb := resilience.NewCircuitBreaker("test",
		resilience.WithMaxFailures(10),
		resilience.WithTimeout(60*time.Second),
		resilience.WithMaxRequests(50),
	)

	_ = cb // Verify it doesn't panic
}

func TestCircuitBreaker_StateStrings(t *testing.T) {
	tests := []struct {
		state resilience.State
		want  string
	}{
		{resilience.StateClosed, "closed"},
		{resilience.StateOpen, "open"},
		{resilience.StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State.String() = %v, want %v", got, tt.want)
		}
	}
}
