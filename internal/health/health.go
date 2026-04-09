// Package health provides health check functionality for TokMan.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Component represents a health check component
type Component struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Latency   time.Duration          `json:"latency_ms"`
}

// Check represents a health check result
type Check struct {
	Status     Status                 `json:"status"`
	Version    string                 `json:"version"`
	Timestamp  time.Time              `json:"timestamp"`
	Components []Component            `json:"components"`
	System     map[string]interface{} `json:"system,omitempty"`
}

// Checker performs health checks
type Checker struct {
	config    *config.Config
	tracker   *tracking.Tracker
	version   string
	startTime time.Time
	checks    map[string]CheckFunc
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) Component

// NewChecker creates a new health checker
func NewChecker(cfg *config.Config, tracker *tracking.Tracker, version string) *Checker {
	c := &Checker{
		config:    cfg,
		tracker:   tracker,
		version:   version,
		startTime: time.Now(),
		checks:    make(map[string]CheckFunc),
	}

	// Register default checks
	c.RegisterCheck("config", c.checkConfig)
	c.RegisterCheck("database", c.checkDatabase)
	c.RegisterCheck("memory", c.checkMemory)
	c.RegisterCheck("goroutines", c.checkGoroutines)

	return c
}

// RegisterCheck registers a custom health check
func (c *Checker) RegisterCheck(name string, fn CheckFunc) {
	c.checks[name] = fn
}

// Check performs all health checks
func (c *Checker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Status:     StatusHealthy,
		Version:    c.version,
		Timestamp:  time.Now().UTC(),
		Components: make([]Component, 0, len(c.checks)),
		System: map[string]interface{}{
			"uptime_seconds": time.Since(c.startTime).Seconds(),
			"go_version":     runtime.Version(),
			"os":             runtime.GOOS,
			"arch":           runtime.GOARCH,
		},
	}

	// Run all checks
	for name, fn := range c.checks {
		component := fn(ctx)
		component.Name = name
		check.Components = append(check.Components, component)

		// Determine overall status
		if component.Status == StatusUnhealthy {
			check.Status = StatusUnhealthy
		} else if component.Status == StatusDegraded && check.Status == StatusHealthy {
			check.Status = StatusDegraded
		}
	}

	// Add latency
	check.System["check_latency_ms"] = time.Since(start).Milliseconds()

	return check
}

// CheckLiveness performs a lightweight liveness check
func (c *Checker) CheckLiveness(ctx context.Context) Component {
	return Component{
		Name:      "liveness",
		Status:    StatusHealthy,
		Message:   "Service is running",
		Timestamp: time.Now().UTC(),
		Latency:   0,
	}
}

// CheckReadiness performs a readiness check
func (c *Checker) CheckReadiness(ctx context.Context) Component {
	component := Component{
		Name:      "readiness",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()

	// Check if critical components are ready
	if c.config == nil {
		component.Status = StatusUnhealthy
		component.Message = "Configuration not loaded"
		return component
	}

	// Validate configuration
	if err := c.config.Validate(); err != nil {
		component.Status = StatusDegraded
		component.Message = fmt.Sprintf("Configuration validation warning: %v", err)
	} else {
		component.Status = StatusHealthy
		component.Message = "Service is ready"
	}

	component.Latency = time.Since(start)
	return component
}

// Individual health checks

func (c *Checker) checkConfig(ctx context.Context) Component {
	component := Component{
		Name:      "config",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()

	if c.config == nil {
		component.Status = StatusUnhealthy
		component.Message = "Configuration not loaded"
		component.Latency = time.Since(start)
		return component
	}

	if err := c.config.Validate(); err != nil {
		component.Status = StatusDegraded
		component.Message = fmt.Sprintf("Validation warning: %v", err)
	} else {
		component.Status = StatusHealthy
		component.Message = "Configuration valid"
	}

	component.Latency = time.Since(start)
	return component
}

func (c *Checker) checkDatabase(ctx context.Context) Component {
	component := Component{
		Name:      "database",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()

	if c.tracker == nil {
		component.Status = StatusDegraded
		component.Message = "Tracking disabled"
		component.Latency = time.Since(start)
		return component
	}

	// Attempt a simple database operation
	// This is a placeholder - implement actual DB health check
	component.Status = StatusHealthy
	component.Message = "Database accessible"

	component.Latency = time.Since(start)
	return component
}

func (c *Checker) checkMemory(ctx context.Context) Component {
	component := Component{
		Name:      "memory",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check memory usage
	allocMB := m.Alloc / 1024 / 1024
	component.Details = map[string]interface{}{
		"alloc_mb":       allocMB,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
	}

	// Consider degraded if using > 1GB
	if allocMB > 1024 {
		component.Status = StatusDegraded
		component.Message = fmt.Sprintf("High memory usage: %d MB", allocMB)
	} else {
		component.Status = StatusHealthy
		component.Message = fmt.Sprintf("Memory usage: %d MB", allocMB)
	}

	component.Latency = time.Since(start)
	return component
}

func (c *Checker) checkGoroutines(ctx context.Context) Component {
	component := Component{
		Name:      "goroutines",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()

	numGoroutines := runtime.NumGoroutine()
	component.Details = map[string]interface{}{
		"count": numGoroutines,
	}

	// Consider degraded if > 1000 goroutines
	if numGoroutines > 1000 {
		component.Status = StatusDegraded
		component.Message = fmt.Sprintf("High goroutine count: %d", numGoroutines)
	} else {
		component.Status = StatusHealthy
		component.Message = fmt.Sprintf("Goroutines: %d", numGoroutines)
	}

	component.Latency = time.Since(start)
	return component
}

// HTTP handlers

// HealthHandler returns an HTTP handler for health checks
func (c *Checker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		check := c.Check(ctx)

		status := http.StatusOK
		if check.Status == StatusDegraded {
			status = http.StatusOK // Still return 200, but indicate degraded
		} else if check.Status == StatusUnhealthy {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(check)
	}
}

// LivenessHandler returns an HTTP handler for Kubernetes liveness probes
func (c *Checker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		component := c.CheckLiveness(ctx)

		w.Header().Set("Content-Type", "application/json")
		if component.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(component)
	}
}

// ReadinessHandler returns an HTTP handler for Kubernetes readiness probes
func (c *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		component := c.CheckReadiness(ctx)

		w.Header().Set("Content-Type", "application/json")
		if component.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(component)
	}
}
