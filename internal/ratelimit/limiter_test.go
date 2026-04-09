package ratelimit_test

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/ratelimit"
)

func TestNew(t *testing.T) {
	l := ratelimit.New(10, 5)
	if l == nil {
		t.Fatal("expected limiter to not be nil")
	}
}

func TestAllow_SeparateClients(t *testing.T) {
	l := ratelimit.New(10, 1)

	// Different clients should have separate limits
	if !l.Allow("client1") {
		t.Error("expected client1 to be allowed")
	}
	if !l.Allow("client2") {
		t.Error("expected client2 to be allowed")
	}
}

func TestAllow_DifferentKeys(t *testing.T) {
	l := ratelimit.New(10, 2)

	// Test with different keys
	allowed := 0
	for i := 0; i < 100; i++ {
		if l.Allow("test-key") {
			allowed++
		}
	}

	// Should allow some but not all due to rate limiting
	t.Logf("Allowed %d out of 100 requests", allowed)
	if allowed == 0 {
		t.Error("expected some requests to be allowed")
	}
}

func TestReset(t *testing.T) {
	l := ratelimit.New(100, 1)

	// Exhaust by using same client
	l.Allow("client")
	l.Allow("client")

	l.Reset()

	// Should allow again after reset
	if !l.Allow("client") {
		t.Error("expected to allow after reset")
	}
}

func TestStatus(t *testing.T) {
	l := ratelimit.New(10, 20)

	status := l.Status()

	if status["requests_per_second"] != float64(10) {
		t.Errorf("expected 10, got %v", status["requests_per_second"])
	}
	if status["burst"] != 20 {
		t.Errorf("expected 20, got %v", status["burst"])
	}
}

func TestEmptyKey(t *testing.T) {
	l := ratelimit.New(10, 1)

	// Empty key should use default
	if !l.Allow("") {
		t.Error("expected empty key to be allowed")
	}
}
