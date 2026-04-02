package benchmarking

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// RegressionDetector detects performance regressions
type RegressionDetector struct {
	thresholds RegressionThresholds
	history    map[string][]HistoricalResult
}

// RegressionThresholds defines thresholds for regression detection
type RegressionThresholds struct {
	LatencyRegression float64 // Percentage increase threshold
	ThroughputDrop    float64 // Percentage drop threshold
	MemoryIncrease    float64 // Percentage increase threshold
	ErrorRateIncrease float64 // Absolute increase threshold
	SuccessRateDrop   float64 // Percentage drop threshold
}

// DefaultRegressionThresholds returns default thresholds
func DefaultRegressionThresholds() RegressionThresholds {
	return RegressionThresholds{
		LatencyRegression: 10.0, // 10% increase
		ThroughputDrop:    10.0, // 10% drop
		MemoryIncrease:    20.0, // 20% increase
		ErrorRateIncrease: 5.0,  // 5% absolute increase
		SuccessRateDrop:   5.0,  // 5% drop
	}
}

// Regression represents a detected regression
type Regression struct {
	BenchmarkID string
	Metric      string
	Baseline    float64
	Current     float64
	Change      float64
	ChangePct   float64
	Severity    RegressionSeverity
	Confidence  float64
	DetectedAt  time.Time
	Description string
}

// RegressionSeverity represents severity level
type RegressionSeverity string

const (
	SeverityCritical RegressionSeverity = "critical"
	SeverityHigh     RegressionSeverity = "high"
	SeverityMedium   RegressionSeverity = "medium"
	SeverityLow      RegressionSeverity = "low"
)

// RegressionReport contains all detected regressions
type RegressionReport struct {
	GeneratedAt     time.Time
	Regressions     []Regression
	Summary         RegressionSummary
	Recommendations []string
}

// RegressionSummary provides summary statistics
type RegressionSummary struct {
	TotalRegressions int
	CriticalCount    int
	HighCount        int
	MediumCount      int
	LowCount         int
	ByMetric         map[string]int
}

// NewRegressionDetector creates a new regression detector
func NewRegressionDetector(thresholds RegressionThresholds) *RegressionDetector {
	return &RegressionDetector{
		thresholds: thresholds,
		history:    make(map[string][]HistoricalResult),
	}
}

// DetectRegressions detects regressions between baseline and current results
func (rd *RegressionDetector) DetectRegressions(baseline, current []BenchmarkResult) []Regression {
	regressions := make([]Regression, 0)

	// Create lookup maps
	baselineMap := make(map[string]BenchmarkResult)
	for _, b := range baseline {
		baselineMap[b.Name] = b
	}

	currentMap := make(map[string]BenchmarkResult)
	for _, c := range current {
		currentMap[c.Name] = c
	}

	// Check each current result against baseline
	for name, currentResult := range currentMap {
		baselineResult, exists := baselineMap[name]
		if !exists {
			continue // New benchmark, no baseline
		}

		// Check various metrics
		if r := rd.checkLatency(baselineResult, currentResult); r != nil {
			regressions = append(regressions, *r)
		}

		if r := rd.checkThroughput(baselineResult, currentResult); r != nil {
			regressions = append(regressions, *r)
		}

		if r := rd.checkMemory(baselineResult, currentResult); r != nil {
			regressions = append(regressions, *r)
		}

		if r := rd.checkErrors(baselineResult, currentResult); r != nil {
			regressions = append(regressions, *r)
		}

		if r := rd.checkSuccessRate(baselineResult, currentResult); r != nil {
			regressions = append(regressions, *r)
		}
	}

	// Sort by severity
	sort.Slice(regressions, func(i, j int) bool {
		return severityWeight(regressions[i].Severity) > severityWeight(regressions[j].Severity)
	})

	return regressions
}

func (rd *RegressionDetector) checkLatency(baseline, current BenchmarkResult) *Regression {
	baselineLatency := float64(baseline.Duration)
	currentLatency := float64(current.Duration)

	if baselineLatency == 0 {
		return nil
	}

	changePct := ((currentLatency - baselineLatency) / baselineLatency) * 100

	if changePct > rd.thresholds.LatencyRegression {
		return &Regression{
			BenchmarkID: current.Name,
			Metric:      "latency",
			Baseline:    baselineLatency,
			Current:     currentLatency,
			Change:      currentLatency - baselineLatency,
			ChangePct:   changePct,
			Severity:    rd.calculateSeverity(changePct, rd.thresholds.LatencyRegression),
			DetectedAt:  time.Now(),
			Description: fmt.Sprintf("Latency increased by %.2f%%", changePct),
		}
	}

	return nil
}

func (rd *RegressionDetector) checkThroughput(baseline, current BenchmarkResult) *Regression {
	if baseline.Throughput == 0 {
		return nil
	}

	changePct := ((current.Throughput - baseline.Throughput) / baseline.Throughput) * 100

	if changePct < -rd.thresholds.ThroughputDrop {
		return &Regression{
			BenchmarkID: current.Name,
			Metric:      "throughput",
			Baseline:    baseline.Throughput,
			Current:     current.Throughput,
			Change:      current.Throughput - baseline.Throughput,
			ChangePct:   changePct,
			Severity:    rd.calculateSeverity(-changePct, rd.thresholds.ThroughputDrop),
			DetectedAt:  time.Now(),
			Description: fmt.Sprintf("Throughput dropped by %.2f%%", -changePct),
		}
	}

	return nil
}

func (rd *RegressionDetector) checkMemory(baseline, current BenchmarkResult) *Regression {
	if baseline.MemoryUsedMB == 0 {
		return nil
	}

	changePct := ((current.MemoryUsedMB - baseline.MemoryUsedMB) / baseline.MemoryUsedMB) * 100

	if changePct > rd.thresholds.MemoryIncrease {
		return &Regression{
			BenchmarkID: current.Name,
			Metric:      "memory",
			Baseline:    baseline.MemoryUsedMB,
			Current:     current.MemoryUsedMB,
			Change:      current.MemoryUsedMB - baseline.MemoryUsedMB,
			ChangePct:   changePct,
			Severity:    rd.calculateSeverity(changePct, rd.thresholds.MemoryIncrease),
			DetectedAt:  time.Now(),
			Description: fmt.Sprintf("Memory usage increased by %.2f%%", changePct),
		}
	}

	return nil
}

func (rd *RegressionDetector) checkErrors(baseline, current BenchmarkResult) *Regression {
	// Calculate error rates (errors per iteration)
	baselineErrorRate := float64(baseline.Errors)
	currentErrorRate := float64(current.Errors)

	absoluteIncrease := currentErrorRate - baselineErrorRate

	if absoluteIncrease > rd.thresholds.ErrorRateIncrease {
		return &Regression{
			BenchmarkID: current.Name,
			Metric:      "error_rate",
			Baseline:    baselineErrorRate,
			Current:     currentErrorRate,
			Change:      absoluteIncrease,
			ChangePct:   (absoluteIncrease / (baselineErrorRate + 1)) * 100, // Avoid division by zero
			Severity:    rd.calculateSeverity(absoluteIncrease, rd.thresholds.ErrorRateIncrease),
			DetectedAt:  time.Now(),
			Description: fmt.Sprintf("Error rate increased by %.0f errors", absoluteIncrease),
		}
	}

	return nil
}

func (rd *RegressionDetector) checkSuccessRate(baseline, current BenchmarkResult) *Regression {
	change := current.SuccessRate - baseline.SuccessRate

	if change < -rd.thresholds.SuccessRateDrop {
		return &Regression{
			BenchmarkID: current.Name,
			Metric:      "success_rate",
			Baseline:    baseline.SuccessRate,
			Current:     current.SuccessRate,
			Change:      change,
			ChangePct:   change,
			Severity:    rd.calculateSeverity(-change, rd.thresholds.SuccessRateDrop),
			DetectedAt:  time.Now(),
			Description: fmt.Sprintf("Success rate dropped by %.2f%%", -change),
		}
	}

	return nil
}

func (rd *RegressionDetector) calculateSeverity(changePct, threshold float64) RegressionSeverity {
	ratio := changePct / threshold

	switch {
	case ratio >= 5.0:
		return SeverityCritical
	case ratio >= 3.0:
		return SeverityHigh
	case ratio >= 1.5:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// GenerateReport generates a regression report
func (rd *RegressionDetector) GenerateReport(baseline, current []BenchmarkResult) *RegressionReport {
	regressions := rd.DetectRegressions(baseline, current)

	report := &RegressionReport{
		GeneratedAt:     time.Now(),
		Regressions:     regressions,
		Recommendations: rd.generateRecommendations(regressions),
	}

	// Generate summary
	report.Summary = RegressionSummary{
		TotalRegressions: len(regressions),
		ByMetric:         make(map[string]int),
	}

	for _, r := range regressions {
		switch r.Severity {
		case SeverityCritical:
			report.Summary.CriticalCount++
		case SeverityHigh:
			report.Summary.HighCount++
		case SeverityMedium:
			report.Summary.MediumCount++
		case SeverityLow:
			report.Summary.LowCount++
		}
		report.Summary.ByMetric[r.Metric]++
	}

	return report
}

func (rd *RegressionDetector) generateRecommendations(regressions []Regression) []string {
	recommendations := make([]string, 0)

	if len(regressions) == 0 {
		return recommendations
	}

	// Check for patterns
	latencyCount := 0
	throughputCount := 0
	memoryCount := 0
	criticalCount := 0

	for _, r := range regressions {
		switch r.Metric {
		case "latency":
			latencyCount++
		case "throughput":
			throughputCount++
		case "memory":
			memoryCount++
		}

		if r.Severity == SeverityCritical {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("URGENT: %d critical regressions detected. Consider reverting changes.", criticalCount))
	}

	if latencyCount > 0 {
		recommendations = append(recommendations,
			"Consider optimizing hot paths or reviewing recent algorithmic changes.")
	}

	if throughputCount > 0 {
		recommendations = append(recommendations,
			"Review concurrent operations and resource contention.")
	}

	if memoryCount > 0 {
		recommendations = append(recommendations,
			"Check for memory leaks or inefficient allocations.")
	}

	if len(regressions) > 5 {
		recommendations = append(recommendations,
			"Multiple regressions detected. Consider running git bisect to identify the problematic commit.")
	}

	return recommendations
}

// FormatReport formats a regression report as a string
func FormatRegressionReport(report *RegressionReport) string {
	output := "Performance Regression Report\n"
	output += "=============================\n\n"
	output += fmt.Sprintf("Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	output += fmt.Sprintf("Total Regressions: %d\n\n", report.Summary.TotalRegressions)

	// Summary
	output += "Summary by Severity:\n"
	output += fmt.Sprintf("  Critical: %d\n", report.Summary.CriticalCount)
	output += fmt.Sprintf("  High:     %d\n", report.Summary.HighCount)
	output += fmt.Sprintf("  Medium:   %d\n", report.Summary.MediumCount)
	output += fmt.Sprintf("  Low:      %d\n\n", report.Summary.LowCount)

	// By metric
	if len(report.Summary.ByMetric) > 0 {
		output += "Regressions by Metric:\n"
		for metric, count := range report.Summary.ByMetric {
			output += fmt.Sprintf("  %s: %d\n", metric, count)
		}
		output += "\n"
	}

	// Individual regressions
	if len(report.Regressions) > 0 {
		output += "Detected Regressions:\n"
		output += "--------------------\n"

		for _, r := range report.Regressions {
			symbol := "⚠"
			switch r.Severity {
			case SeverityCritical:
				symbol = "🚨"
			case SeverityHigh:
				symbol = "❌"
			case SeverityMedium:
				symbol = "⚠️"
			case SeverityLow:
				symbol = "ℹ️"
			}

			output += fmt.Sprintf("\n%s [%s] %s: %s\n", symbol, r.Severity, r.BenchmarkID, r.Metric)
			output += fmt.Sprintf("   %s\n", r.Description)
			output += fmt.Sprintf("   Baseline: %.2f, Current: %.2f (%.2f%%)\n",
				r.Baseline, r.Current, r.ChangePct)
		}
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		output += "\nRecommendations:\n"
		output += "-----------------\n"
		for i, rec := range report.Recommendations {
			output += fmt.Sprintf("%d. %s\n", i+1, rec)
		}
	}

	return output
}

// severityWeight returns a numeric weight for severity sorting
func severityWeight(s RegressionSeverity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// IsRegression checks if a single change constitutes a regression
func IsRegression(baseline, current float64, thresholdPct float64, higherIsBetter bool) bool {
	if baseline == 0 {
		return false
	}

	changePct := ((current - baseline) / baseline) * 100

	if higherIsBetter {
		return changePct < -thresholdPct
	}
	return changePct > thresholdPct
}

// CalculateRegressionScore calculates an overall regression score (0-100)
func CalculateRegressionScore(regressions []Regression) float64 {
	if len(regressions) == 0 {
		return 100.0
	}

	totalWeight := 0.0
	for _, r := range regressions {
		weight := float64(severityWeight(r.Severity)) * math.Abs(r.ChangePct)
		totalWeight += weight
	}

	// Score decreases as regressions increase
	score := 100.0 - (totalWeight / float64(len(regressions)))
	if score < 0 {
		score = 0
	}

	return score
}

// RegressionHistory tracks regression history over time
type RegressionHistory struct {
	Detections []RegressionDetection
}

// RegressionDetection represents a detection event
type RegressionDetection struct {
	Timestamp   time.Time
	CommitHash  string
	Regressions []Regression
	Score       float64
}

// AddDetection adds a detection to history
func (rh *RegressionHistory) AddDetection(commitHash string, regressions []Regression) {
	detection := RegressionDetection{
		Timestamp:   time.Now(),
		CommitHash:  commitHash,
		Regressions: regressions,
		Score:       CalculateRegressionScore(regressions),
	}
	rh.Detections = append(rh.Detections, detection)
}

// Trend shows regression trend over time
func (rh *RegressionHistory) Trend() string {
	if len(rh.Detections) < 2 {
		return "insufficient data"
	}

	first := rh.Detections[0].Score
	last := rh.Detections[len(rh.Detections)-1].Score

	if last > first {
		return "improving"
	} else if last < first {
		return "degrading"
	}
	return "stable"
}
