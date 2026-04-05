package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestInitialState(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		SuccessThreshold: 2,
	})

	if cb.State() != StateClosed {
		t.Errorf("expected closed state, got %v", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Errorf("expected allowed in closed state: %v", err)
	}
}

func TestOpensAfterFailures(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		SuccessThreshold: 2,
	})

	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Error("should still be closed after 1 failure")
	}

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected open state, got %v", cb.State())
	}

	if err := cb.Allow(); !errors.Is(err, ErrOpen) {
		t.Errorf("expected ErrOpen, got: %v", err)
	}
}

func TestRecoveryAfterTimeout(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 1,
		RecoveryTimeout:  50 * time.Millisecond,
		SuccessThreshold: 1,
	})

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatal("should be open")
	}

	// Wait for recovery timeout
	time.Sleep(100 * time.Millisecond)

	if err := cb.Allow(); err != nil {
		t.Errorf("expected allowed after timeout, got: %v", err)
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected halfopen, got %v", cb.State())
	}

	// Record success should close circuit
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected closed after success, got %v", cb.State())
	}
}

func TestSuccessResetsFailureCount(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 3,
		RecoveryTimeout:  time.Second,
		SuccessThreshold: 1,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	cb.RecordSuccess() // should reset failure counter
	cb.RecordFailure() // should be 1 again, not 3

	if cb.State() != StateClosed {
		t.Errorf("expected closed, got %v", cb.State())
	}
}

func TestConcurrentAccess(t *testing.T) {
	cb := New(DefaultConfig())
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cb.Allow()
			cb.RecordSuccess()
			cb.RecordFailure()
		}()
	}
	wg.Wait()
}
