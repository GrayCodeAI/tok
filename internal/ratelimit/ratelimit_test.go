package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	l := NewLimiter(10, 5)

	// Should allow burst up to capacity
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("expected allow at burst %d", i)
		}
	}

	// Should deny after burst is exhausted
	if l.Allow() {
		t.Error("expected deny after burst exhausted")
	}
}

func TestLimiter_Wait(t *testing.T) {
	l := NewLimiter(1000, 1) // 1000/sec, burst 1

	ctx := context.Background()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}

	// With capacity 1 and rate 1000, second wait should be near-instant
	start := time.Now()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Error("wait took too long for high-rate limiter")
	}
}

func TestLimiter_WaitContextCancel(t *testing.T) {
	l := NewLimiter(0.001, 1) // Very slow rate
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Exhaust the single token
	_ = l.Allow()

	// Next wait should timeout
	if err := l.Wait(ctx); err == nil {
		t.Error("expected timeout error")
	}
}
