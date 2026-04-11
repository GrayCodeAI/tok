package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(10, 5) // 10 req/sec, burst of 5

	clientIP := "192.168.1.1"

	// Should allow initial burst
	for i := 0; i < 5; i++ {
		if !rl.Allow(clientIP) {
			t.Errorf("expected request %d to be allowed (burst)", i+1)
		}
	}

	// 6th request should be denied (burst exhausted)
	if rl.Allow(clientIP) {
		t.Error("expected 6th request to be denied")
	}

	// Different client should be allowed
	otherClient := "192.168.1.2"
	if !rl.Allow(otherClient) {
		t.Error("expected different client to be allowed")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	rl := NewRateLimiter(100, 1) // 100 req/sec, burst of 1

	clientIP := "192.168.1.1"

	// First request allowed
	if !rl.Allow(clientIP) {
		t.Error("expected first request to be allowed")
	}

	// Second request denied (burst = 1)
	if rl.Allow(clientIP) {
		t.Error("expected second request to be denied")
	}

	// Wait for token refill
	time.Sleep(20 * time.Millisecond)

	// Should be allowed now
	if !rl.Allow(clientIP) {
		t.Error("expected request after refill to be allowed")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	clientIP := "192.168.1.1"

	// Make a request
	rl.Allow(clientIP)

	// Verify client exists
	rl.mu.RLock()
	_, exists := rl.clients[clientIP]
	rl.mu.RUnlock()
	if !exists {
		t.Error("expected client to exist after request")
	}

	// Force cleanup (simulating old entries)
	rl.mu.Lock()
	for ip, cl := range rl.clients {
		cl.mu.Lock()
		cl.lastUpdate = time.Now().Add(-15 * time.Minute) // Make it look old
		cl.mu.Unlock()
		rl.clients[ip] = cl
	}
	rl.mu.Unlock()

	rl.cleanup()

	// Client should be removed (no recent activity)
	rl.mu.RLock()
	_, exists = rl.clients[clientIP]
	rl.mu.RUnlock()
	if exists {
		t.Error("expected client to be cleaned up after inactivity")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expected   string
	}{
		{
			name:       "X-Forwarded-For",
			remoteAddr: "192.168.1.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "10.0.0.1, 10.0.0.2"},
			expected:   "10.0.0.1",
		},
		{
			name:       "X-Real-Ip",
			remoteAddr: "192.168.1.1:1234",
			headers:    map[string]string{"X-Real-Ip": "10.0.0.5"},
			expected:   "10.0.0.5",
		},
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:1234",
			headers:    map[string]string{},
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			headers:    map[string]string{},
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getClientIP(req)
			if got != tt.expected {
				t.Errorf("getClientIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := &config.Config{}
	s := New(cfg, "test")
	s.rateLimiter = NewRateLimiter(1, 1) // Very strict: 1 req/sec, burst 1

	// Create handler that returns success
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := s.rateLimitMiddleware(handler)

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr1.Code)
	}

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rr2.Code)
	}

	// Health endpoint should bypass rate limiting
	req3 := httptest.NewRequest("GET", "/health", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	rr3 := httptest.NewRecorder()
	middleware.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Errorf("health endpoint should bypass rate limit, got status %d", rr3.Code)
	}
}

func TestRateLimitMiddleware_DifferentClients(t *testing.T) {
	cfg := &config.Config{}
	s := New(cfg, "test")
	s.rateLimiter = NewRateLimiter(1, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := s.rateLimitMiddleware(handler)

	// Client 1 makes request
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("client 1 first request should succeed, got %d", rr1.Code)
	}

	// Client 2 should also be allowed (different IP)
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("client 2 first request should succeed, got %d", rr2.Code)
	}
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	cfg := &config.Config{}
	s := New(cfg, "test")
	s.rateLimiter = NewRateLimiter(1, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := s.rateLimitMiddleware(handler)

	// Request that gets rate limited
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	// Exhaust burst
	middleware.ServeHTTP(httptest.NewRecorder(), req)

	// This should be rate limited
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Check Retry-After header
	retryAfter := rr.Header().Get("Retry-After")
	if retryAfter != "60" {
		t.Errorf("expected Retry-After header to be 60, got %s", retryAfter)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000, 1000)
	clientIP := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(clientIP)
	}
}

func BenchmarkRateLimiter_Parallel(b *testing.B) {
	rl := NewRateLimiter(10000, 1000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			clientIP := "192.168.1." + string(rune('0'+i%10))
			rl.Allow(clientIP)
			i++
		}
	})
}
