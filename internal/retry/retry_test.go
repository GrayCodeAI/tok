package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialWait != 100*time.Millisecond {
		t.Errorf("expected InitialWait=100ms, got %v", cfg.InitialWait)
	}
	if cfg.MaxWait != 5*time.Second {
		t.Errorf("expected MaxWait=5s, got %v", cfg.MaxWait)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", cfg.Multiplier)
	}
}

func TestDo_Success(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Do(context.Background(), cfg, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDo_RetryThenSuccess(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Do(context.Background(), cfg, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	callCount := 0
	expectedErr := errors.New("persistent error")
	fn := func() error {
		callCount++
		return expectedErr
	}

	err := Do(context.Background(), cfg, fn)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	cfg := Config{
		MaxAttempts: 5,
		InitialWait: 1 * time.Second,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return errors.New("error")
	}

	err := Do(ctx, cfg, fn)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDoWithResult_Success(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "success", nil
	}

	result, err := DoWithResult(context.Background(), cfg, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDoWithResult_RetryThenSuccess(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 2 {
			return 0, errors.New("temporary error")
		}
		return 42, nil
	}

	result, err := DoWithResult(context.Background(), cfg, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDoWithResult_MaxAttemptsExceeded(t *testing.T) {
	cfg := Config{
		MaxAttempts: 2,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	expectedErr := errors.New("persistent error")
	fn := func() (string, error) {
		return "", expectedErr
	}

	result, err := DoWithResult(context.Background(), cfg, fn)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}
}

func TestBackoffCalculation(t *testing.T) {
	cfg := Config{
		MaxAttempts: 5,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     50 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	start := time.Now()
	fn := func() error {
		attempts++
		if attempts < 4 {
			return errors.New("error")
		}
		return nil
	}

	Do(context.Background(), cfg, fn)
	elapsed := time.Since(start)

	// Expected waits: 0ms (first), 10ms, 20ms = ~30ms minimum
	if elapsed < 25*time.Millisecond {
		t.Errorf("backoff too fast: %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("backoff too slow: %v", elapsed)
	}
}

func TestBackoffMaxWait(t *testing.T) {
	cfg := Config{
		MaxAttempts: 10,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     25 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return errors.New("error")
	}

	start := time.Now()
	Do(context.Background(), cfg, fn)
	elapsed := time.Since(start)

	// With maxWait=25ms, waits should be: 0, 10, 20, 25, 25, 25, 25, 25, 25 = ~210ms max
	// But allow some tolerance
	if elapsed > 300*time.Millisecond {
		t.Errorf("max wait not respected, took %v", elapsed)
	}
}
