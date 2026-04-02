package benchmarking

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendUp     TrendDirection = "up"
	TrendDown   TrendDirection = "down"
	TrendStable TrendDirection = "stable"
)

// TrendTracker tracks benchmark trends over time
type TrendTracker struct {
	history map[string][]HistoricalResult
}

// HistoricalResult represents a benchmark result at a point in time
type HistoricalResult struct {
	Timestamp   time.Time
	Value       float64
	Metric      string
	BenchmarkID string
	Metadata    map[string]string
}

// Trend represents a trend analysis for a benchmark
type Trend struct {
	BenchmarkID   string
	Metric        string
	Direction     TrendDirection
	Slope         float64
	Intercept     float64
	Correlation   float64
	CurrentValue  float64
	PredictedNext float64
	ChangePct     float64
	History       []HistoricalPoint
}

// HistoricalPoint represents a single data point
type HistoricalPoint struct {
	Timestamp time.Time
	Value     float64
}

// TrendReport contains trend analysis for multiple benchmarks
type TrendReport struct {
	GeneratedAt time.Time
	Period      time.Duration
	Trends      []Trend
	Summary     TrendSummary
}

// TrendSummary provides aggregate trend information
type TrendSummary struct {
	TotalBenchmarks int
	Improving       int
	Degrading       int
	Stable          int
	HighCorrelation int
}

// NewTrendTracker creates a new trend tracker
func NewTrendTracker() *TrendTracker {
	return &TrendTracker{
		history: make(map[string][]HistoricalResult),
	}
}

// Record adds a benchmark result to the history
func (tt *TrendTracker) Record(result BenchmarkResult) {
	key := fmt.Sprintf("%s:%s", result.Name, "throughput")

	entry := HistoricalResult{
		Timestamp:   result.Timestamp,
		Value:       result.Throughput,
		Metric:      "throughput",
		BenchmarkID: result.Name,
		Metadata:    result.Metadata,
	}

	tt.history[key] = append(tt.history[key], entry)
}

// AnalyzeTrend performs trend analysis for a benchmark
func (tt *TrendTracker) AnalyzeTrend(benchmarkID, metric string) *Trend {
	key := fmt.Sprintf("%s:%s", benchmarkID, metric)
	history := tt.history[key]

	if len(history) < 2 {
		return nil
	}

	// Sort by timestamp
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.Before(history[j].Timestamp)
	})

	// Convert to points
	points := make([]HistoricalPoint, len(history))
	for i, h := range history {
		points[i] = HistoricalPoint{
			Timestamp: h.Timestamp,
			Value:     h.Value,
		}
	}

	// Perform linear regression
	slope, intercept, correlation := linearRegression(points)

	// Calculate change percentage
	first := points[0].Value
	last := points[len(points)-1].Value
	changePct := 0.0
	if first != 0 {
		changePct = ((last - first) / first) * 100
	}

	// Determine direction based on percentage change
	direction := TrendStable
	if changePct > 1.0 {
		direction = TrendUp
	} else if changePct < -1.0 {
		direction = TrendDown
	}

	// Predict next value
	lastTime := float64(points[len(points)-1].Timestamp.Unix())
	nextTime := lastTime + float64(time.Hour*24)
	predictedNext := slope*nextTime + intercept

	return &Trend{
		BenchmarkID:   benchmarkID,
		Metric:        metric,
		Direction:     direction,
		Slope:         slope,
		Intercept:     intercept,
		Correlation:   correlation,
		CurrentValue:  last,
		PredictedNext: predictedNext,
		ChangePct:     changePct,
		History:       points,
	}
}

// GenerateReport generates a trend report for all tracked benchmarks
func (tt *TrendTracker) GenerateReport(period time.Duration) *TrendReport {
	report := &TrendReport{
		GeneratedAt: time.Now(),
		Period:      period,
		Trends:      make([]Trend, 0),
	}

	// Analyze each benchmark
	analyzed := make(map[string]bool)
	for key := range tt.history {
		// Parse key (benchmarkID:metric)
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}
		benchmarkID := parts[0]
		_ = parts[1] // metric

		if analyzed[benchmarkID] {
			continue
		}

		trend := tt.AnalyzeTrend(benchmarkID, "throughput")
		if trend != nil {
			report.Trends = append(report.Trends, *trend)
			analyzed[benchmarkID] = true
		}
	}

	// Generate summary
	for _, trend := range report.Trends {
		report.Summary.TotalBenchmarks++

		switch trend.Direction {
		case TrendUp:
			report.Summary.Improving++
		case TrendDown:
			report.Summary.Degrading++
		default:
			report.Summary.Stable++
		}

		if trend.Correlation > 0.7 {
			report.Summary.HighCorrelation++
		}
	}

	return report
}

// Save persists the trend history to storage
func (tt *TrendTracker) Save(storage TrendStorage) error {
	return storage.Save(tt.history)
}

// Load restores the trend history from storage
func (tt *TrendTracker) Load(storage TrendStorage) error {
	history, err := storage.Load()
	if err != nil {
		return err
	}
	tt.history = history
	return nil
}

// TrendStorage interface for persistence
type TrendStorage interface {
	Save(history map[string][]HistoricalResult) error
	Load() (map[string][]HistoricalResult, error)
}

// linearRegression performs simple linear regression
func linearRegression(points []HistoricalPoint) (slope, intercept, correlation float64) {
	n := float64(len(points))
	if n < 2 {
		return 0, 0, 0
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64

	firstTime := float64(points[0].Timestamp.Unix())

	for _, p := range points {
		x := float64(p.Timestamp.Unix()) - firstTime
		y := p.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	// Calculate means
	meanX := sumX / n
	meanY := sumY / n

	// Calculate slope and intercept
	divisor := sumX2 - n*meanX*meanX
	if divisor == 0 {
		return 0, meanY, 0
	}

	slope = (sumXY - n*meanX*meanY) / divisor
	intercept = meanY - slope*meanX

	// Calculate correlation coefficient (Pearson's r)
	covXY := sumXY - n*meanX*meanY
	varX := sumX2 - n*meanX*meanX
	varY := sumY2 - n*meanY*meanY

	if varX <= 0 || varY <= 0 {
		correlation = 0
	} else {
		correlation = covXY * covXY / (varX * varY)
		// correlation is r^2, take square root
		if correlation > 0 {
			correlation = 1.0 / correlation
		}
	}

	return slope, intercept, correlation
}

// FormatTrendReport formats a trend report as a string
func FormatTrendReport(report *TrendReport) string {
	output := "Benchmark Trend Report\n"
	output += "======================\n\n"
	output += fmt.Sprintf("Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	output += fmt.Sprintf("Period: %v\n\n", report.Period)

	// Summary
	output += "Summary\n"
	output += "-------\n"
	output += fmt.Sprintf("Total Benchmarks: %d\n", report.Summary.TotalBenchmarks)
	output += fmt.Sprintf("  Improving:  %d\n", report.Summary.Improving)
	output += fmt.Sprintf("  Degrading:  %d\n", report.Summary.Degrading)
	output += fmt.Sprintf("  Stable:     %d\n", report.Summary.Stable)
	output += fmt.Sprintf("High Confidence: %d\n\n", report.Summary.HighCorrelation)

	// Individual trends
	output += "Benchmark Trends\n"
	output += "----------------\n"
	for _, trend := range report.Trends {
		symbol := "→"
		if trend.Direction == TrendUp {
			symbol = "↑"
		} else if trend.Direction == TrendDown {
			symbol = "↓"
		}

		output += fmt.Sprintf("\n%s %s\n", symbol, trend.BenchmarkID)
		output += fmt.Sprintf("  Direction: %s (%.2f%%)\n", trend.Direction, trend.ChangePct)
		output += fmt.Sprintf("  Current:   %.2f\n", trend.CurrentValue)
		output += fmt.Sprintf("  Predicted: %.2f\n", trend.PredictedNext)
		output += fmt.Sprintf("  Confidence: %.2f\n", trend.Correlation)
	}

	return output
}

// ExportTrendReport exports a trend report as JSON
func ExportTrendReport(report *TrendReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// DetectRegressions detects performance regressions
func DetectRegressions(report *TrendReport, threshold float64) []Trend {
	regressions := make([]Trend, 0)

	for _, trend := range report.Trends {
		if trend.Direction == TrendDown && trend.ChangePct < -threshold {
			regressions = append(regressions, trend)
		}
	}

	return regressions
}

// DetectImprovements detects performance improvements
func DetectImprovements(report *TrendReport, threshold float64) []Trend {
	improvements := make([]Trend, 0)

	for _, trend := range report.Trends {
		if trend.Direction == TrendUp && trend.ChangePct > threshold {
			improvements = append(improvements, trend)
		}
	}

	return improvements
}
