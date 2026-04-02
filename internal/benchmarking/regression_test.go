package benchmarking

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRegressionDetector(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	if detector == nil {
		t.Fatal("expected detector to be created")
	}

	if detector.history == nil {
		t.Error("expected history map to be initialized")
	}
}

func TestDefaultRegressionThresholds(t *testing.T) {
	thresholds := DefaultRegressionThresholds()

	if thresholds.LatencyRegression != 10.0 {
		t.Errorf("expected latency threshold 10.0, got %.2f", thresholds.LatencyRegression)
	}

	if thresholds.ThroughputDrop != 10.0 {
		t.Errorf("expected throughput threshold 10.0, got %.2f", thresholds.ThroughputDrop)
	}

	if thresholds.MemoryIncrease != 20.0 {
		t.Errorf("expected memory threshold 20.0, got %.2f", thresholds.MemoryIncrease)
	}
}

func TestRegressionDetectorDetect(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := []BenchmarkResult{
		{
			Name:         "test",
			Duration:     100 * time.Millisecond,
			Throughput:   1000,
			MemoryUsedMB: 10,
			Errors:       0,
			SuccessRate:  100,
		},
	}

	current := []BenchmarkResult{
		{
			Name:         "test",
			Duration:     120 * time.Millisecond, // 20% increase
			Throughput:   800,                    // 20% drop
			MemoryUsedMB: 15,                     // 50% increase
			Errors:       5,
			SuccessRate:  95,
		},
	}

	regressions := detector.DetectRegressions(baseline, current)

	if len(regressions) == 0 {
		t.Error("expected regressions to be detected")
	}
}

func TestRegressionDetectorNoRegression(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := []BenchmarkResult{
		{
			Name:     "test",
			Duration: 100 * time.Millisecond,
		},
	}

	current := []BenchmarkResult{
		{
			Name:     "test",
			Duration: 101 * time.Millisecond, // Only 1% increase
		},
	}

	regressions := detector.DetectRegressions(baseline, current)

	if len(regressions) != 0 {
		t.Errorf("expected no regressions, got %d", len(regressions))
	}
}

func TestRegressionDetectorNewBenchmark(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := []BenchmarkResult{}

	current := []BenchmarkResult{
		{
			Name: "new-benchmark",
		},
	}

	regressions := detector.DetectRegressions(baseline, current)

	if len(regressions) != 0 {
		t.Error("expected no regressions for new benchmark")
	}
}

func TestCheckLatencyRegression(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := BenchmarkResult{
		Name:     "test",
		Duration: 100 * time.Millisecond,
	}

	// Regression detected
	current := BenchmarkResult{
		Name:     "test",
		Duration: 120 * time.Millisecond, // 20% increase
	}

	regression := detector.checkLatency(baseline, current)

	if regression == nil {
		t.Fatal("expected regression to be detected")
	}

	if regression.Metric != "latency" {
		t.Errorf("expected metric 'latency', got %s", regression.Metric)
	}

	if regression.ChangePct < 10.0 {
		t.Errorf("expected change > 10%%, got %.2f%%", regression.ChangePct)
	}
}

func TestCheckThroughputRegression(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := BenchmarkResult{
		Name:       "test",
		Throughput: 1000,
	}

	current := BenchmarkResult{
		Name:       "test",
		Throughput: 800, // 20% drop
	}

	regression := detector.checkThroughput(baseline, current)

	if regression == nil {
		t.Fatal("expected regression to be detected")
	}

	if regression.ChangePct > -10.0 {
		t.Errorf("expected drop > 10%%, got %.2f%%", regression.ChangePct)
	}
}

func TestCalculateSeverity(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	tests := []struct {
		change    float64
		threshold float64
		expected  RegressionSeverity
	}{
		{50.0, 10.0, SeverityCritical}, // 5x threshold
		{30.0, 10.0, SeverityHigh},     // 3x threshold
		{20.0, 10.0, SeverityMedium},   // 2x threshold
		{10.0, 10.0, SeverityLow},      // 1x threshold
	}

	for _, tt := range tests {
		severity := detector.calculateSeverity(tt.change, tt.threshold)
		if severity != tt.expected {
			t.Errorf("change %.1f, threshold %.1f: expected %s, got %s",
				tt.change, tt.threshold, tt.expected, severity)
		}
	}
}

func TestGenerateReport(t *testing.T) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := []BenchmarkResult{
		{
			Name:         "test1",
			Duration:     100 * time.Millisecond,
			Throughput:   1000,
			MemoryUsedMB: 10,
		},
	}

	current := []BenchmarkResult{
		{
			Name:         "test1",
			Duration:     120 * time.Millisecond,
			Throughput:   800,
			MemoryUsedMB: 15,
		},
	}

	report := detector.GenerateReport(baseline, current)

	if report == nil {
		t.Fatal("expected report to be generated")
	}

	if len(report.Regressions) == 0 {
		t.Error("expected regressions in report")
	}

	if report.Summary.TotalRegressions == 0 {
		t.Error("expected non-zero total regressions")
	}
}

func TestFormatRegressionReport(t *testing.T) {
	report := &RegressionReport{
		GeneratedAt: time.Now(),
		Regressions: []Regression{
			{
				BenchmarkID: "test",
				Metric:      "latency",
				Severity:    SeverityHigh,
				Description: "Test regression",
				ChangePct:   20.0,
			},
		},
		Summary: RegressionSummary{
			TotalRegressions: 1,
			HighCount:        1,
			ByMetric:         map[string]int{"latency": 1},
		},
		Recommendations: []string{"Review changes"},
	}

	output := FormatRegressionReport(report)

	if output == "" {
		t.Error("expected non-empty output")
	}

	if !contains(output, "test") {
		t.Error("expected benchmark name in output")
	}

	if !contains(output, "latency") {
		t.Error("expected metric in output")
	}
}

func TestIsRegression(t *testing.T) {
	tests := []struct {
		baseline       float64
		current        float64
		threshold      float64
		higherIsBetter bool
		expected       bool
	}{
		{100, 120, 10, false, true},  // Latency increased > 10%
		{100, 105, 10, false, false}, // Latency increased < 10%
		{100, 80, 10, true, true},    // Throughput dropped > 10%
		{100, 95, 10, true, false},   // Throughput dropped < 10%
		{0, 100, 10, false, false},   // Baseline is 0
	}

	for _, tt := range tests {
		result := IsRegression(tt.baseline, tt.current, tt.threshold, tt.higherIsBetter)
		if result != tt.expected {
			t.Errorf("baseline %.1f, current %.1f, threshold %.1f, higherIsBetter %v: expected %v, got %v",
				tt.baseline, tt.current, tt.threshold, tt.higherIsBetter, tt.expected, result)
		}
	}
}

func TestCalculateRegressionScore(t *testing.T) {
	// No regressions = perfect score
	score := CalculateRegressionScore([]Regression{})
	if score != 100.0 {
		t.Errorf("expected score 100.0 for no regressions, got %.2f", score)
	}

	// With regressions
	regressions := []Regression{
		{Severity: SeverityLow, ChangePct: 10.0},
	}
	score = CalculateRegressionScore(regressions)
	if score >= 100.0 {
		t.Error("expected score < 100.0 with regressions")
	}

	// With critical regression
	regressions = []Regression{
		{Severity: SeverityCritical, ChangePct: 50.0},
	}
	score = CalculateRegressionScore(regressions)
	if score > 0 {
		t.Errorf("expected score near 0 with critical regression, got %.2f", score)
	}
}

func TestRegressionHistory(t *testing.T) {
	history := &RegressionHistory{}

	regressions := []Regression{
		{Severity: SeverityHigh, ChangePct: 20.0},
	}

	history.AddDetection("abc123", regressions)

	if len(history.Detections) != 1 {
		t.Errorf("expected 1 detection, got %d", len(history.Detections))
	}

	if history.Detections[0].CommitHash != "abc123" {
		t.Errorf("expected commit hash 'abc123', got %s", history.Detections[0].CommitHash)
	}

	if history.Detections[0].Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestRegressionHistoryTrend(t *testing.T) {
	history := &RegressionHistory{}

	// Initial detection with high score
	history.AddDetection("abc123", []Regression{})

	// Later detection with regressions (low score)
	history.AddDetection("def456", []Regression{
		{Severity: SeverityHigh, ChangePct: 30.0},
	})

	trend := history.Trend()
	if trend != "degrading" {
		t.Errorf("expected trend 'degrading', got %s", trend)
	}
}

func TestRegressionHistoryTrendInsufficientData(t *testing.T) {
	history := &RegressionHistory{}

	trend := history.Trend()
	if trend != "insufficient data" {
		t.Errorf("expected 'insufficient data', got %s", trend)
	}
}

func BenchmarkRegressionDetectorDetect(b *testing.B) {
	thresholds := DefaultRegressionThresholds()
	detector := NewRegressionDetector(thresholds)

	baseline := make([]BenchmarkResult, 100)
	current := make([]BenchmarkResult, 100)

	for i := 0; i < 100; i++ {
		baseline[i] = BenchmarkResult{
			Name:         fmt.Sprintf("bench-%d", i),
			Duration:     100 * time.Millisecond,
			Throughput:   1000,
			MemoryUsedMB: 10,
		}
		current[i] = BenchmarkResult{
			Name:         fmt.Sprintf("bench-%d", i),
			Duration:     120 * time.Millisecond,
			Throughput:   800,
			MemoryUsedMB: 15,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectRegressions(baseline, current)
	}
}
