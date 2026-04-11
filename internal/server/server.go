// Package server provides HTTP API server for TokMan.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/health"
	"github.com/GrayCodeAI/tokman/internal/metrics"
	"github.com/GrayCodeAI/tokman/internal/security"
)

// RateLimiter implements token bucket rate limiting per IP
type RateLimiter struct {
	mu              sync.RWMutex
	clients         map[string]*clientLimiter
	limit           float64       // tokens per second
	burst           int           // bucket capacity
	cleanupInterval time.Duration // read-only after creation
}

type clientLimiter struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients:         make(map[string]*clientLimiter),
		limit:           limit,
		burst:           burst,
		cleanupInterval: 5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// Allow checks if request from clientIP is allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.RLock()
	cl, exists := rl.clients[clientIP]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		cl = &clientLimiter{
			tokens:     float64(rl.burst) - 1,
			lastUpdate: time.Now(),
		}
		rl.clients[clientIP] = cl
		rl.mu.Unlock()
		return true
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Add tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(cl.lastUpdate).Seconds()
	cl.tokens += elapsed * rl.limit
	if cl.tokens > float64(rl.burst) {
		cl.tokens = float64(rl.burst)
	}
	cl.lastUpdate = now

	// Check if request can be allowed
	if cl.tokens >= 1 {
		cl.tokens--
		return true
	}
	return false
}

// cleanupLoop removes stale entries periodically
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes clients that haven't made requests recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for ip, cl := range rl.clients {
		cl.mu.Lock()
		lastUpdate := cl.lastUpdate
		cl.mu.Unlock()

		if lastUpdate.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-Ip header
	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Server represents the HTTP API server
type Server struct {
	cfg          *config.Config
	server       *http.Server
	metrics      *metrics.Metrics
	health       *health.Checker
	validate     *security.Validator
	version      string
	rateLimiter  *RateLimiter
}

// New creates a new server instance
func New(cfg *config.Config, version string) *Server {
	s := &Server{
		cfg:          cfg,
		version:      version,
		metrics:      metrics.Get(),
		validate:     security.NewValidator(),
		rateLimiter:  NewRateLimiter(10, 100), // 10 req/sec, burst of 100
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", s.handleHealth())
	mux.HandleFunc("/health/live", s.handleLiveness())
	mux.HandleFunc("/health/ready", s.handleReadiness())

	// API v1
	mux.HandleFunc("/api/v1/compress", s.handleCompress())
	mux.HandleFunc("/api/v1/config", s.handleGetConfig())
	mux.HandleFunc("/api/v1/metrics", s.handleGetMetrics())
	mux.HandleFunc("/api/v1/filters", s.handleListFilters())
	mux.HandleFunc("/api/v1/stats", s.handleGetStats())
	mux.HandleFunc("/api/v1/openapi.json", s.handleOpenAPI())

	// Middleware chain: logging -> rate limiting -> recovery
	handler := s.loggingMiddleware(s.rateLimitMiddleware(s.recoveryMiddleware(mux)))

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	fmt.Printf("Server started on %s\n", addr)
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Run runs the server until shutdown
func (s *Server) Run(ctx context.Context, addr string) error {
	if err := s.Start(addr); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sigChan:
		fmt.Println("\nShutting down server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.Stop(shutdownCtx)
}

// Middleware

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.writeError(w, http.StatusInternalServerError, "internal", fmt.Sprintf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/health/live" || r.URL.Path == "/health/ready" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)
		if !s.rateLimiter.Allow(clientIP) {
			w.Header().Set("Retry-After", "60")
			s.writeError(w, http.StatusTooManyRequests, "ERR_RATE_LIMIT", "rate limit exceeded, please retry later")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handlers

func (s *Server) handleHealth() http.HandlerFunc {
	checker := health.NewChecker(s.cfg, nil, s.version)
	return func(w http.ResponseWriter, r *http.Request) {
		check := checker.Check(r.Context())
		s.writeJSON(w, http.StatusOK, check)
	}
}

func (s *Server) handleLiveness() http.HandlerFunc {
	checker := health.NewChecker(s.cfg, nil, s.version)
	return checker.LivenessHandler()
}

func (s *Server) handleReadiness() http.HandlerFunc {
	checker := health.NewChecker(s.cfg, nil, s.version)
	return checker.ReadinessHandler()
}

type compressRequest struct {
	Text   string `json:"text"`
	Mode   string `json:"mode"`
	Budget int    `json:"budget"`
	Preset string `json:"preset"`
	Query  string `json:"query"`
}

type compressResponse struct {
	Original   string        `json:"original"`
	Compressed string        `json:"compressed"`
	Stats      pipelineStats `json:"stats"`
}

type pipelineStats struct {
	OriginalTokens   int     `json:"original_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	TokensSaved      int     `json:"tokens_saved"`
	CompressionRatio float64 `json:"compression_ratio"`
	ProcessingTimeMs int64   `json:"processing_time_ms"`
}

func (s *Server) handleCompress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var req compressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "ERR_INVALID_INPUT", err.Error())
			return
		}

		if req.Text == "" {
			s.writeError(w, http.StatusBadRequest, "ERR_INVALID_INPUT", "text is required")
			return
		}

		if req.Budget > 0 {
			if err := s.validate.ValidateBudget(req.Budget); err != nil {
				s.writeError(w, http.StatusBadRequest, "ERR_INVALID_INPUT", err.Error())
				return
			}
		}

		mode := filter.ModeMinimal
		if req.Mode == "aggressive" {
			mode = filter.ModeAggressive
		}

		p := filter.NewPipelineCoordinator(filter.PipelineConfig{
			Mode:            mode,
			QueryIntent:     req.Query,
			Budget:          req.Budget,
			SessionTracking: true,
		})

		output, stats := p.Process(req.Text)

		s.metrics.RecordCompressionRun()
		s.metrics.RecordCompressionDuration(time.Since(start))

		ratio := 0.0
		if stats.OriginalTokens > 0 {
			ratio = float64(stats.FinalTokens) / float64(stats.OriginalTokens)
		}

		resp := compressResponse{
			Original:   req.Text,
			Compressed: output,
			Stats: pipelineStats{
				OriginalTokens:   stats.OriginalTokens,
				OutputTokens:     stats.FinalTokens,
				TokensSaved:      stats.TotalSaved,
				CompressionRatio: ratio,
				ProcessingTimeMs: time.Since(start).Milliseconds(),
			},
		}

		s.writeJSON(w, http.StatusOK, resp)
	}
}

func (s *Server) handleGetConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.cfg)
	}
}

func (s *Server) handleUpdateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newCfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			s.writeError(w, http.StatusBadRequest, "ERR_INVALID_CONFIG", err.Error())
			return
		}

		if err := newCfg.Validate(); err != nil {
			s.writeError(w, http.StatusBadRequest, "ERR_INVALID_CONFIG", err.Error())
			return
		}

		s.cfg = &newCfg
		s.writeJSON(w, http.StatusOK, map[string]string{"message": "Configuration updated"})
	}
}

func (s *Server) handleGetMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.metrics.Snapshot())
	}
}

type filterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func (s *Server) handleListFilters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := []filterInfo{
			{"entropy", "Entropy-based token pruning", s.cfg.Pipeline.EnableEntropy},
			{"perplexity", "LLMLingua-style perplexity pruning", s.cfg.Pipeline.EnablePerplexity},
			{"h2o", "Heavy-Hitter Oracle compression", s.cfg.Pipeline.EnableH2O},
			{"compaction", "Semantic compaction", s.cfg.Pipeline.EnableCompaction},
			{"attribution", "Token attribution filtering", s.cfg.Pipeline.EnableAttribution},
		}
		s.writeJSON(w, http.StatusOK, map[string][]filterInfo{"filters": filters})
	}
}

func (s *Server) handleGetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"total_commands":     s.metrics.Snapshot().CommandsProcessed,
			"total_tokens_saved": s.metrics.Snapshot().CommandsProcessed * 100,
			"uptime_seconds":     s.metrics.Snapshot().UptimeSeconds,
		}
		s.writeJSON(w, http.StatusOK, stats)
	}
}

func (s *Server) handleOpenAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		openapi := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]string{
				"title":   "TokMan API",
				"version": s.version,
			},
		}
		s.writeJSON(w, http.StatusOK, openapi)
	}
}

// Helpers

func (s *Server) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, code int, errCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errCode,
		"message": message,
	})
}
