// Package metrics provides comprehensive metrics for TokMan.
package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds all metrics for TokMan
type Metrics struct {
	mu sync.RWMutex

	// Counters
	commandsProcessed atomic.Int64
	commandsFailed    atomic.Int64
	compressionRuns   atomic.Int64
	compressionErrors atomic.Int64
	cacheHits         atomic.Int64
	cacheMisses       atomic.Int64

	// Histograms (simplified implementation)
	compressionDurations []time.Duration
	durationIndex        atomic.Int64

	// Current state
	activeConnections atomic.Int64
	memoryUsageMB     atomic.Int64
	queueSize         atomic.Int64

	// Timestamps
	startTime time.Time
	lastReset time.Time
}

// Global metrics instance
var global = &Metrics{
	startTime: time.Now(),
	lastReset: time.Now(),
}

// Get returns the global metrics instance
func Get() *Metrics {
	return global
}

// RecordCommandProcessed records a processed command
func (m *Metrics) RecordCommandProcessed() {
	m.commandsProcessed.Add(1)
}

// RecordCommandFailed records a failed command
func (m *Metrics) RecordCommandFailed() {
	m.commandsFailed.Add(1)
}

// RecordCompressionRun records a compression run
func (m *Metrics) RecordCompressionRun() {
	m.compressionRuns.Add(1)
}

// RecordCompressionError records a compression error
func (m *Metrics) RecordCompressionError() {
	m.compressionErrors.Add(1)
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.cacheHits.Add(1)
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.cacheMisses.Add(1)
}

// RecordCompressionDuration records compression duration
func (m *Metrics) RecordCompressionDuration(d time.Duration) {
	idx := m.durationIndex.Add(1) - 1
	slot := int(idx % 1000)

	m.mu.Lock()
	if len(m.compressionDurations) == 0 {
		m.compressionDurations = make([]time.Duration, 1000)
	}
	m.compressionDurations[slot] = d
	m.mu.Unlock()
}

// IncActiveConnections increments active connections
func (m *Metrics) IncActiveConnections() {
	m.activeConnections.Add(1)
}

// DecActiveConnections decrements active connections
func (m *Metrics) DecActiveConnections() {
	m.activeConnections.Add(-1)
}

// SetMemoryUsage sets memory usage in MB
func (m *Metrics) SetMemoryUsage(mb int64) {
	m.memoryUsageMB.Store(mb)
}

// SetQueueSize sets the current queue size
func (m *Metrics) SetQueueSize(size int64) {
	m.queueSize.Store(size)
}

// Snapshot returns a snapshot of current metrics
func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		CommandsProcessed: m.commandsProcessed.Load(),
		CommandsFailed:    m.commandsFailed.Load(),
		CompressionRuns:   m.compressionRuns.Load(),
		CompressionErrors: m.compressionErrors.Load(),
		CacheHits:         m.cacheHits.Load(),
		CacheMisses:       m.cacheMisses.Load(),
		ActiveConnections: m.activeConnections.Load(),
		MemoryUsageMB:     m.memoryUsageMB.Load(),
		QueueSize:         m.queueSize.Load(),
		UptimeSeconds:     time.Since(m.startTime).Seconds(),
		CompressionRate:   m.calculateCompressionRate(),
		CacheHitRate:      m.calculateCacheHitRate(),
		AverageDurationMs: m.calculateAverageDuration(),
	}
}

// Snapshot represents a point-in-time metrics snapshot
type Snapshot struct {
	CommandsProcessed int64   `json:"commands_processed"`
	CommandsFailed    int64   `json:"commands_failed"`
	CompressionRuns   int64   `json:"compression_runs"`
	CompressionErrors int64   `json:"compression_errors"`
	CacheHits         int64   `json:"cache_hits"`
	CacheMisses       int64   `json:"cache_misses"`
	ActiveConnections int64   `json:"active_connections"`
	MemoryUsageMB     int64   `json:"memory_usage_mb"`
	QueueSize         int64   `json:"queue_size"`
	UptimeSeconds     float64 `json:"uptime_seconds"`
	CompressionRate   float64 `json:"compression_rate"`
	CacheHitRate      float64 `json:"cache_hit_rate"`
	AverageDurationMs float64 `json:"average_duration_ms"`
}

func (m *Metrics) calculateCompressionRate() float64 {
	total := m.commandsProcessed.Load() + m.commandsFailed.Load()
	if total == 0 {
		return 0
	}
	return float64(m.commandsProcessed.Load()) / float64(total)
}

func (m *Metrics) calculateCacheHitRate() float64 {
	total := m.cacheHits.Load() + m.cacheMisses.Load()
	if total == 0 {
		return 0
	}
	return float64(m.cacheHits.Load()) / float64(total)
}

func (m *Metrics) calculateAverageDuration() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.compressionDurations) == 0 {
		return 0
	}

	var total time.Duration
	count := 0
	for _, d := range m.compressionDurations {
		if d > 0 {
			total += d
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return float64(total.Milliseconds()) / float64(count)
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.commandsProcessed.Store(0)
	m.commandsFailed.Store(0)
	m.compressionRuns.Store(0)
	m.compressionErrors.Store(0)
	m.cacheHits.Store(0)
	m.cacheMisses.Store(0)
	m.activeConnections.Store(0)
	m.memoryUsageMB.Store(0)
	m.queueSize.Store(0)
	m.lastReset = time.Now()
}

// GetMetrics returns a context-aware metrics getter
func GetMetrics(ctx context.Context) *Metrics {
	return global
}

// Record functions with context

// RecordCommandProcessedWithContext records a processed command with context
func RecordCommandProcessedWithContext(ctx context.Context) {
	GetMetrics(ctx).RecordCommandProcessed()
}

// RecordCompressionWithContext records compression with context
func RecordCompressionWithContext(ctx context.Context, duration time.Duration, inputTokens, outputTokens int64) {
	m := GetMetrics(ctx)
	m.RecordCompressionRun()
	m.RecordCompressionDuration(duration)
}

// RecordErrorWithContext records an error with context
func RecordErrorWithContext(ctx context.Context) {
	GetMetrics(ctx).RecordCompressionError()
}
