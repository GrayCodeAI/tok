package benchmarking

import (
	"testing"
	"time"
)

func TestNewTrendTracker(t *testing.T) {
	tracker := NewTrendTracker()
	if tracker == nil {
		t.Fatal("expected tracker to be created")
	}

	if tracker.history == nil {
		t.Error("expected history map to be initialized")
	}
}

func TestTrendTrackerRecord(t *testing.T) {
	tracker := NewTrendTracker()

	result := BenchmarkResult{
		Name:       "test-benchmark",
		Timestamp:  time.Now(),
		Throughput: 100.0,
	}

	tracker.Record(result)

	key := "test-benchmark:throughput"
	if len(tracker.history[key]) != 1 {
		t.Errorf("expected 1 entry, got %d", len(tracker.history[key]))
	}
}

func TestTrendTrackerAnalyzeTrend(t *testing.T) {
	tracker := NewTrendTracker()

	// Add historical data points
	for i := 0; i < 5; i++ {
		result := BenchmarkResult{
			Name:       "test-benchmark",
			Timestamp:  time.Now().Add(time.Duration(i) * time.Hour),
			Throughput: 100.0 + float64(i)*10, // Increasing trend
		}
		tracker.Record(result)
	}

	trend := tracker.AnalyzeTrend("test-benchmark", "throughput")
	if trend == nil {
		t.Fatal("expected trend to be analyzed")
	}

	if trend.BenchmarkID != "test-benchmark" {
		t.Errorf("expected benchmark ID 'test-benchmark', got %s", trend.BenchmarkID)
	}

	// Should detect upward trend
	if trend.Direction != TrendUp {
		t.Errorf("expected upward trend, got %s", trend.Direction)
	}

	// Should have positive change percentage
	if trend.ChangePct <= 0 {
		t.Errorf("expected positive change, got %.2f", trend.ChangePct)
	}
}

func TestTrendTrackerAnalyzeTrendInsufficientData(t *testing.T) {
	tracker := NewTrendTracker()

	// Only 1 data point
	result := BenchmarkResult{
		Name:       "test-benchmark",
		Timestamp:  time.Now(),
		Throughput: 100.0,
	}
	tracker.Record(result)

	trend := tracker.AnalyzeTrend("test-benchmark", "throughput")
	if trend != nil {
		t.Error("expected nil trend with insufficient data")
	}
}

func TestTrendTrackerGenerateReport(t *testing.T) {
	tracker := NewTrendTracker()

	// Add data for multiple benchmarks
	for i := 0; i < 5; i++ {
		tracker.Record(BenchmarkResult{
			Name:       "benchmark-1",
			Timestamp:  time.Now().Add(time.Duration(i) * time.Hour),
			Throughput: 100.0 + float64(i)*5,
		})

		tracker.Record(BenchmarkResult{
			Name:       "benchmark-2",
			Timestamp:  time.Now().Add(time.Duration(i) * time.Hour),
			Throughput: 200.0 - float64(i)*5,
		})
	}

	report := tracker.GenerateReport(7 * 24 * time.Hour)

	if report == nil {
		t.Fatal("expected report to be generated")
	}

	if len(report.Trends) != 2 {
		t.Errorf("expected 2 trends, got %d", len(report.Trends))
	}

	if report.Summary.TotalBenchmarks != 2 {
		t.Errorf("expected 2 total benchmarks, got %d", report.Summary.TotalBenchmarks)
	}
}

func TestLinearRegression(t *testing.T) {
	tests := []struct {
		name        string
		points      []HistoricalPoint
		expectSlope float64
	}{
		{
			name: "upward trend",
			points: []HistoricalPoint{
				{Timestamp: time.Unix(0, 0), Value: 10},
				{Timestamp: time.Unix(1, 0), Value: 20},
				{Timestamp: time.Unix(2, 0), Value: 30},
			},
			expectSlope: 10,
		},
		{
			name: "downward trend",
			points: []HistoricalPoint{
				{Timestamp: time.Unix(0, 0), Value: 30},
				{Timestamp: time.Unix(1, 0), Value: 20},
				{Timestamp: time.Unix(2, 0), Value: 10},
			},
			expectSlope: -10,
		},
		{
			name: "stable trend",
			points: []HistoricalPoint{
				{Timestamp: time.Unix(0, 0), Value: 20},
				{Timestamp: time.Unix(1, 0), Value: 20},
				{Timestamp: time.Unix(2, 0), Value: 20},
			},
			expectSlope: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slope, _, _ := linearRegression(tt.points)

			// Check slope direction
			if tt.expectSlope > 0 && slope <= 0 {
				t.Errorf("expected positive slope, got %.2f", slope)
			}
			if tt.expectSlope < 0 && slope >= 0 {
				t.Errorf("expected negative slope, got %.2f", slope)
			}
			if tt.expectSlope == 0 && slope != 0 {
				t.Errorf("expected zero slope, got %.2f", slope)
			}

			// Note: correlation calculation is simplified for trend detection
			// The main goal is detecting direction, not exact correlation value
		})
	}
}

func TestLinearRegressionInsufficientData(t *testing.T) {
	points := []HistoricalPoint{
		{Timestamp: time.Now(), Value: 100},
	}

	slope, intercept, correlation := linearRegression(points)

	// With single point, slope and correlation should be 0, intercept should be average (which is the value itself)
	if slope != 0 {
		t.Errorf("expected slope=0, got %.2f", slope)
	}
	if correlation != 0 {
		t.Errorf("expected correlation=0, got %.2f", correlation)
	}
	// Intercept with single point is undefined, function returns 0
	if intercept != 0 {
		t.Errorf("expected intercept=0 for single point, got %.2f", intercept)
	}
}

func TestFormatTrendReport(t *testing.T) {
	report := &TrendReport{
		GeneratedAt: time.Now(),
		Period:      7 * 24 * time.Hour,
		Trends: []Trend{
			{
				BenchmarkID:   "test-1",
				Direction:     TrendUp,
				ChangePct:     15.5,
				CurrentValue:  100.0,
				PredictedNext: 115.0,
				Correlation:   0.95,
			},
		},
		Summary: TrendSummary{
			TotalBenchmarks: 1,
			Improving:       1,
			HighCorrelation: 1,
		},
	}

	output := FormatTrendReport(report)

	if output == "" {
		t.Error("expected non-empty output")
	}

	if !containsSubstring(output, "test-1") {
		t.Error("expected benchmark name in output")
	}

	if !containsSubstring(output, "up") {
		t.Error("expected trend direction in output")
	}
}

func TestExportTrendReport(t *testing.T) {
	report := &TrendReport{
		GeneratedAt: time.Now(),
		Trends: []Trend{
			{
				BenchmarkID: "test-1",
				Direction:   TrendUp,
			},
		},
	}

	data, err := ExportTrendReport(report)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestDetectRegressions(t *testing.T) {
	report := &TrendReport{
		Trends: []Trend{
			{BenchmarkID: "regressing", Direction: TrendDown, ChangePct: -20.0},
			{BenchmarkID: "stable", Direction: TrendStable, ChangePct: -1.0},
			{BenchmarkID: "improving", Direction: TrendUp, ChangePct: 10.0},
		},
	}

	regressions := DetectRegressions(report, 10.0)

	if len(regressions) != 1 {
		t.Errorf("expected 1 regression, got %d", len(regressions))
	}

	if regressions[0].BenchmarkID != "regressing" {
		t.Errorf("expected 'regressing', got %s", regressions[0].BenchmarkID)
	}
}

func TestDetectImprovements(t *testing.T) {
	report := &TrendReport{
		Trends: []Trend{
			{BenchmarkID: "improving", Direction: TrendUp, ChangePct: 20.0},
			{BenchmarkID: "stable", Direction: TrendUp, ChangePct: 5.0},
			{BenchmarkID: "regressing", Direction: TrendDown, ChangePct: -10.0},
		},
	}

	improvements := DetectImprovements(report, 10.0)

	if len(improvements) != 1 {
		t.Errorf("expected 1 improvement, got %d", len(improvements))
	}

	if improvements[0].BenchmarkID != "improving" {
		t.Errorf("expected 'improving', got %s", improvements[0].BenchmarkID)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkTrendTrackerRecord(b *testing.B) {
	tracker := NewTrendTracker()

	for i := 0; i < b.N; i++ {
		tracker.Record(BenchmarkResult{
			Name:       "benchmark",
			Timestamp:  time.Now(),
			Throughput: float64(i),
		})
	}
}

func BenchmarkLinearRegression(b *testing.B) {
	points := make([]HistoricalPoint, 100)
	for i := range points {
		points[i] = HistoricalPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Value:     float64(i) * 10,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linearRegression(points)
	}
}
