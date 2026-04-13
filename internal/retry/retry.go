package retry

import (
	"context"
	"fmt"
	"time"
)

// Config defines retry behavior
type Config struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

// Do executes fn with exponential backoff retry
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	wait := cfg.InitialWait
	
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			
			wait = time.Duration(float64(wait) * cfg.Multiplier)
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
		
		if err := fn(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	
	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// DoWithResult executes fn with retry and returns result
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	wait := cfg.InitialWait
	
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(wait):
			}
			
			wait = time.Duration(float64(wait) * cfg.Multiplier)
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
		
		var err error
		result, err = fn()
		if err != nil {
			lastErr = err
			continue
		}
		return result, nil
	}
	
	return result, fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}
