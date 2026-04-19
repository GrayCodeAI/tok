package metrics_test

import (
	"testing"
	"time"

	"github.com/lakshmanpatel/tok/internal/metrics"
)

func TestRecordCommandProcessed(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.RecordCommandProcessed()

	snap := m.Snapshot()
	if snap.CommandsProcessed != 1 {
		t.Errorf("expected 1, got %d", snap.CommandsProcessed)
	}
}

func TestRecordCommandFailed(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.RecordCommandFailed()
	m.RecordCommandFailed()

	snap := m.Snapshot()
	if snap.CommandsFailed != 2 {
		t.Errorf("expected 2, got %d", snap.CommandsFailed)
	}
}

func TestCompressionRun(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.RecordCompressionRun()
	m.RecordCompressionDuration(50 * time.Millisecond)

	snap := m.Snapshot()
	if snap.CompressionRuns != 1 {
		t.Errorf("expected 1, got %d", snap.CompressionRuns)
	}
	if snap.AverageDurationMs == 0 {
		t.Error("expected duration to be recorded")
	}
}

func TestCacheHitMiss(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheMiss()

	snap := m.Snapshot()
	if snap.CacheHits != 2 {
		t.Errorf("expected 2 hits, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 1 {
		t.Errorf("expected 1 miss, got %d", snap.CacheMisses)
	}

	rate := snap.CacheHitRate
	if rate < 0.66 || rate > 0.67 {
		t.Errorf("expected rate ~0.66, got %f", rate)
	}
}

func TestActiveConnections(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.IncActiveConnections()
	m.IncActiveConnections()
	m.DecActiveConnections()

	snap := m.Snapshot()
	if snap.ActiveConnections != 1 {
		t.Errorf("expected 1, got %d", snap.ActiveConnections)
	}
}

func TestMemoryUsage(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.SetMemoryUsage(256)

	snap := m.Snapshot()
	if snap.MemoryUsageMB != 256 {
		t.Errorf("expected 256, got %d", snap.MemoryUsageMB)
	}
}

func TestCompressionRate(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	// Record successes and failures
	for i := 0; i < 100; i++ {
		m.RecordCommandProcessed()
	}
	for i := 0; i < 5; i++ {
		m.RecordCommandFailed()
	}

	snap := m.Snapshot()
	rate := snap.CompressionRate
	if rate < 0.95 || rate > 0.96 {
		t.Errorf("expected rate ~0.95, got %f", rate)
	}
}

func TestUptime(t *testing.T) {
	m := metrics.Get()

	snap := m.Snapshot()
	if snap.UptimeSeconds <= 0 {
		t.Error("expected uptime > 0")
	}
}

func TestSnapshot(t *testing.T) {
	m := metrics.Get()
	m.Reset()

	m.RecordCommandProcessed()
	m.RecordCompressionRun()
	m.RecordCacheHit()
	m.SetMemoryUsage(100)

	snap := m.Snapshot()

	// Verify all fields are accessible
	_ = snap.CommandsProcessed
	_ = snap.CommandsFailed
	_ = snap.CompressionRuns
	_ = snap.CompressionErrors
	_ = snap.CacheHits
	_ = snap.CacheMisses
	_ = snap.ActiveConnections
	_ = snap.MemoryUsageMB
	_ = snap.QueueSize
	_ = snap.UptimeSeconds
	_ = snap.CompressionRate
	_ = snap.CacheHitRate
	_ = snap.AverageDurationMs
}

func TestReset(t *testing.T) {
	m := metrics.Get()

	m.RecordCommandProcessed()
	m.RecordCompressionRun()
	m.Reset()

	snap := m.Snapshot()
	if snap.CommandsProcessed != 0 {
		t.Errorf("expected 0 after reset, got %d", snap.CommandsProcessed)
	}
}
