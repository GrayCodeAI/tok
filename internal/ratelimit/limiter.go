package ratelimit

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	mu sync.RWMutex

	clients map[string]*clientState

	RequestsPerSecond float64
	Burst             int
	maxClients        int
}

type clientState struct {
	tokens    float64
	maxTokens float64
	lastAdd   time.Time
}

func New(requestsPerSecond float64, burst int) *Limiter {
	return &Limiter{
		RequestsPerSecond: requestsPerSecond,
		Burst:             burst,
		maxClients:        10000,
		clients:           make(map[string]*clientState),
	}
}

func (l *Limiter) Allow(key string) bool {
	if key == "" {
		key = "default"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	client, exists := l.clients[key]
	if !exists {
		if len(l.clients) >= l.maxClients {
			l.cleanup(now)
		}
		l.clients[key] = &clientState{
			tokens:    float64(l.Burst),
			maxTokens: float64(l.Burst),
			lastAdd:   now,
		}
		return true
	}

	elapsed := now.Sub(client.lastAdd).Seconds()
	client.tokens += elapsed * l.RequestsPerSecond
	if client.tokens > client.maxTokens {
		client.tokens = client.maxTokens
	}
	client.lastAdd = now

	if client.tokens >= 1 {
		client.tokens--
		return true
	}

	return false
}

func (l *Limiter) cleanup(now time.Time) {
	maxAge := 5 * time.Minute
	for k, v := range l.clients {
		if now.Sub(v.lastAdd) > maxAge {
			delete(l.clients, k)
		}
	}
}

func (l *Limiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(r.RemoteAddr) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "ERR_RATE_LIMITED",
				"message": "Rate limit exceeded",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clients = make(map[string]*clientState)
}

func (l *Limiter) Status() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return map[string]interface{}{
		"requests_per_second": l.RequestsPerSecond,
		"burst":               l.Burst,
		"active_clients":      len(l.clients),
	}
}
