package stress

import (
	"context"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	config := DefaultConfig()
	runner := NewRunner(config)

	if runner == nil {
		t.Fatal("expected runner to be created")
	}

	if runner.config.TargetRPS != 100 {
		t.Errorf("expected default TargetRPS 100, got %d", runner.config.TargetRPS)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Duration != 5*time.Minute {
		t.Errorf("expected default Duration 5m, got %v", config.Duration)
	}

	if config.MaxConcurrency != 1000 {
		t.Errorf("expected default MaxConcurrency 1000, got %d", config.MaxConcurrency)
	}
}

func TestRegisterScenario(t *testing.T) {
	config := DefaultConfig()
	runner := NewRunner(config)

	scenario := &Scenario{
		Name:        "test-scenario",
		Type:        TypeLoad,
		Description: "Test scenario",
		Fn: func(ctx context.Context) error {
			return nil
		},
	}

	runner.RegisterScenario(scenario)

	if len(runner.scenarios) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(runner.scenarios))
	}
}

func TestStandardScenarios(t *testing.T) {
	scenarios := StandardScenarios()

	if len(scenarios) == 0 {
		t.Error("expected standard scenarios to be available")
	}

	for _, s := range scenarios {
		if s.Name == "" {
			t.Error("expected scenario to have a name")
		}
		if s.Fn == nil {
			t.Error("expected scenario to have a function")
		}
	}
}

func TestResultGenerateReport(t *testing.T) {
	result := &Result{
		Scenario:      "test",
		Type:          TypeLoad,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(60 * time.Second),
		Duration:      60 * time.Second,
		TotalRequests: 1000,
		SuccessCount:  950,
		ErrorCount:    50,
		LatencyP50:    100 * time.Millisecond,
		LatencyP95:    200 * time.Millisecond,
		LatencyP99:    500 * time.Millisecond,
		MinLatency:    50 * time.Millisecond,
		MaxLatency:    1000 * time.Millisecond,
		ThroughputRPS: 16.67,
	}

	report := result.GenerateReport()

	if report == "" {
		t.Error("expected report to be generated")
	}

	// Report should contain key metrics
	if !contains(report, "test") {
		t.Error("expected report to contain scenario name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPhaseMetricsRecord(t *testing.T) {
	metrics := &phaseMetrics{}

	metrics.record(100*time.Millisecond, nil, false)

	if metrics.totalRequests != 1 {
		t.Errorf("expected 1 request, got %d", metrics.totalRequests)
	}

	if metrics.successCount != 1 {
		t.Errorf("expected 1 success, got %d", metrics.successCount)
	}

	metrics.record(200*time.Millisecond, nil, true)

	if metrics.timeoutCount != 1 {
		t.Errorf("expected 1 timeout, got %d", metrics.timeoutCount)
	}
}

func TestPhaseMetricsCalculatePercentile(t *testing.T) {
	metrics := &phaseMetrics{}

	// Add latencies
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, l := range latencies {
		metrics.latencies = append(metrics.latencies, l)
	}

	p50 := metrics.calculatePercentile(0.5)
	if p50 != 50*time.Millisecond && p50 != 60*time.Millisecond {
		t.Errorf("expected P50 around 50-60ms, got %v", p50)
	}

	p95 := metrics.calculatePercentile(0.95)
	// P95 should be around 90-100ms (index 8 or 9)
	if p95 != 90*time.Millisecond && p95 != 100*time.Millisecond {
		t.Errorf("expected P95 around 90-100ms, got %v", p95)
	}
}

func BenchmarkPhaseMetricsRecord(b *testing.B) {
	metrics := &phaseMetrics{}
	latency := 100 * time.Millisecond

	for i := 0; i < b.N; i++ {
		metrics.record(latency, nil, false)
	}
}
