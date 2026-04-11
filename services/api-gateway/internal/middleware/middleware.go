package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Logging middleware logs all requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", getClientIP(r),
			"user_agent", r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recovery middleware recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, `{"error":"INTERNAL_ERROR","message":"Internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu         sync.RWMutex
	clients    map[string]*client
	rate       float64
	burst      int
}

type client struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rate:    rate,
		burst:   burst,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.RLock()
	c, exists := rl.clients[clientIP]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		c = &client{
			tokens:     float64(rl.burst) - 1,
			lastUpdate: time.Now(),
		}
		rl.clients[clientIP] = c
		rl.mu.Unlock()
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastUpdate).Seconds()
	c.tokens += elapsed * rl.rate
	if c.tokens > float64(rl.burst) {
		c.tokens = float64(rl.burst)
	}
	c.lastUpdate = now

	if c.tokens >= 1 {
		c.tokens--
		return true
	}
	return false
}

// cleanup removes stale clients periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, c := range rl.clients {
			c.mu.Lock()
			lastUpdate := c.lastUpdate
			c.mu.Unlock()
			if lastUpdate.Before(cutoff) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit middleware applies rate limiting
func RateLimit(next http.Handler) http.Handler {
	limiter := NewRateLimiter(100, 200) // 100 req/sec, burst 200

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/health/services" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)
		if !limiter.Allow(clientIP) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"RATE_LIMITED","message":"Rate limit exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORS middleware adds CORS headers
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
