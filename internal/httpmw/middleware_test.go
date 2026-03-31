package httpmw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewDefault(t *testing.T) {
	rl := NewDefault()
	if rl.requests != DefaultRequests {
		t.Errorf("expected %d requests, got %d", DefaultRequests, rl.requests)
	}
	if rl.window != DefaultWindow {
		t.Errorf("expected %v window, got %v", DefaultWindow, rl.window)
	}
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(50, 30*time.Second)
	if rl.requests != 50 {
		t.Errorf("expected 50 requests, got %d", rl.requests)
	}
	if rl.window != 30*time.Second {
		t.Errorf("expected 30s window, got %v", rl.window)
	}
}

func TestRateLimiterMiddleware_AllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := rl.Middleware(handler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}
}

func TestRateLimiterMiddleware_RejectsOverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := rl.Middleware(handler)

	// Should allow 3 requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}

	// 4th request should be rejected
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestRateLimiterMiddleware_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := rl.Middleware(handler)

	// First IP uses its limit
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("expected status %d for IP1, got %d", http.StatusOK, rec1.Code)
	}

	// Second IP should still be allowed
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "127.0.0.2:12345"
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("expected status %d for IP2, got %d", http.StatusOK, rec2.Code)
	}
}

func TestRateLimiterMiddleware_RemoteAddrFallback(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := rl.Middleware(handler)

	// RemoteAddr without port (should not cause panic)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireContentType(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		expected    int
	}{
		{"GET no content type", http.MethodGet, "", http.StatusOK},
		{"POST with correct type", http.MethodPost, "application/json", http.StatusOK},
		{"POST with wrong type", http.MethodPost, "text/plain", http.StatusUnsupportedMediaType},
		{"PUT with correct type", http.MethodPut, "application/json", http.StatusOK},
		{"PUT with wrong type", http.MethodPut, "text/html", http.StatusUnsupportedMediaType},
		{"PATCH with correct type", http.MethodPatch, "application/json", http.StatusOK},
		{"PATCH with wrong type", http.MethodPatch, "text/plain", http.StatusUnsupportedMediaType},
		{"DELETE no content type", http.MethodDelete, "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := RequireContentType("application/json")
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			wrapped := mw(handler)

			req := httptest.NewRequest(tt.method, "/test", strings.NewReader("{}"))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.expected {
				t.Errorf("expected status %d, got %d", tt.expected, rec.Code)
			}
		})
	}
}

func TestRequireContentType_PrefixMatch(t *testing.T) {
	mw := RequireContentType("application/json")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(handler)

	// charset suffix should still match
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d for charset-prefixed type, got %d", http.StatusOK, rec.Code)
	}
}
