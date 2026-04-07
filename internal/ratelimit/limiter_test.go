package ratelimit

import (
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(100, time.Minute)
	if limiter == nil {
		t.Error("Expected non-nil limiter")
	}
}

func TestLimiterAllow(t *testing.T) {
	limiter := NewLimiter(3, time.Minute)

	if !limiter.Allow(nil, "client1") {
		t.Error("Expected first request to be allowed")
	}
	if !limiter.Allow(nil, "client1") {
		t.Error("Expected second request to be allowed")
	}
	if !limiter.Allow(nil, "client1") {
		t.Error("Expected third request to be allowed")
	}
	if limiter.Allow(nil, "client1") {
		t.Error("Expected fourth request to be denied")
	}
}

func TestLimiterGetRemaining(t *testing.T) {
	limiter := NewLimiter(10, time.Minute)

	limiter.Allow(nil, "client1")
	limiter.Allow(nil, "client1")

	remaining := limiter.GetRemaining("client1")
	if remaining != 8 {
		t.Errorf("Expected 8 remaining, got %d", remaining)
	}
}

func TestLimiterReset(t *testing.T) {
	limiter := NewLimiter(5, time.Minute)

	limiter.Allow(nil, "client1")
	limiter.Reset("client1")

	remaining := limiter.GetRemaining("client1")
	if remaining != 5 {
		t.Errorf("Expected 5 remaining after reset, got %d", remaining)
	}
}
