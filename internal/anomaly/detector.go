// Package anomaly provides cost anomaly detection
package anomaly

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Detector detects cost anomalies
type Detector struct {
	threshold   float64
	window      time.Duration
	sensitivity float64
}

// NewDetector creates a new anomaly detector
func NewDetector(threshold float64, window time.Duration) *Detector {
	return &Detector{
		threshold:   threshold,
		window:      window,
		sensitivity: 2.0,
	}
}

// DataPoint represents a cost data point
type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Type        string
	Description string
	Severity    string
	DetectedAt  time.Time
	Value       float64
	Expected    float64
	Deviation   float64
}

// Detect detects anomalies in cost data
func (d *Detector) Detect(data []DataPoint) []Anomaly {
	if len(data) < 3 {
		return nil
	}

	anomalies := make([]Anomaly, 0)

	for i := 2; i < len(data); i++ {
		window := data[max(0, i-5):i]
		mean, stdDev := calculateStats(window)

		current := data[i].Value
		deviation := 0.0
		if stdDev > 0 {
			deviation = (current - mean) / stdDev
		}

		if math.Abs(deviation) > d.sensitivity {
			severity := "medium"
			if math.Abs(deviation) > 3.0 {
				severity = "critical"
			} else if math.Abs(deviation) > 2.5 {
				severity = "high"
			}

			anomalyType := "spike"
			if current < mean {
				anomalyType = "drop"
			}

			anomalies = append(anomalies, Anomaly{
				Type:        anomalyType,
				Description: fmt.Sprintf("Cost %s detected: %.2f (expected ~%.2f)", anomalyType, current, mean),
				Severity:    severity,
				DetectedAt:  data[i].Timestamp,
				Value:       current,
				Expected:    mean,
				Deviation:   deviation,
			})
		}
	}

	return anomalies
}

// DetectSuddenChange detects sudden changes in cost
func (d *Detector) DetectSuddenChange(data []DataPoint) []Anomaly {
	anomalies := make([]Anomaly, 0)

	for i := 1; i < len(data); i++ {
		prev := data[i-1].Value
		curr := data[i].Value

		if prev == 0 {
			continue
		}

		changePct := ((curr - prev) / prev) * 100

		if math.Abs(changePct) > d.threshold {
			severity := "medium"
			if math.Abs(changePct) > 200 {
				severity = "critical"
			} else if math.Abs(changePct) > 100 {
				severity = "high"
			}

			anomalyType := "spike"
			if changePct < 0 {
				anomalyType = "drop"
			}

			anomalies = append(anomalies, Anomaly{
				Type:        anomalyType,
				Description: fmt.Sprintf("Sudden cost %s: %.1f%% change", anomalyType, changePct),
				Severity:    severity,
				DetectedAt:  data[i].Timestamp,
				Value:       curr,
				Expected:    prev,
				Deviation:   changePct,
			})
		}
	}

	return anomalies
}

// DetectTrendChange detects changes in cost trend
func (d *Detector) DetectTrendChange(data []DataPoint) []Anomaly {
	if len(data) < 10 {
		return nil
	}

	anomalies := make([]Anomaly, 0)

	firstHalf := data[:len(data)/2]
	secondHalf := data[len(data)/2:]

	firstTrend := calculateTrend(firstHalf)
	secondTrend := calculateTrend(secondHalf)

	trendChange := secondTrend - firstTrend

	if math.Abs(trendChange) > d.threshold/100 {
		direction := "increasing"
		if trendChange < 0 {
			direction = "decreasing"
		}

		anomalies = append(anomalies, Anomaly{
			Type:        "trend_change",
			Description: fmt.Sprintf("Cost trend %s: %.2f%%/day", direction, trendChange*100),
			Severity:    "high",
			DetectedAt:  data[len(data)/2].Timestamp,
			Value:       secondTrend,
			Expected:    firstTrend,
			Deviation:   trendChange,
		})
	}

	return anomalies
}

func calculateStats(data []DataPoint) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	var sum float64
	for _, dp := range data {
		sum += dp.Value
	}
	mean := sum / float64(len(data))

	var variance float64
	for _, dp := range data {
		variance += (dp.Value - mean) * (dp.Value - mean)
	}
	stdDev := math.Sqrt(variance / float64(len(data)))

	return mean, stdDev
}

func calculateTrend(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	n := float64(len(data))
	var sumX, sumY, sumXY, sumX2 float64

	firstTime := data[0].Timestamp.Unix()
	for _, dp := range data {
		x := float64(dp.Timestamp.Unix() - firstTime)
		y := dp.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	divisor := n*sumX2 - sumX*sumX
	if divisor == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / divisor
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AnomalyReport represents a report of detected anomalies
type AnomalyReport struct {
	TotalAnomalies int
	BySeverity     map[string]int
	ByType         map[string]int
	Anomalies      []Anomaly
	Summary        string
}

// GenerateReport generates an anomaly report
func GenerateReport(anomalies []Anomaly) *AnomalyReport {
	report := &AnomalyReport{
		TotalAnomalies: len(anomalies),
		BySeverity:     make(map[string]int),
		ByType:         make(map[string]int),
		Anomalies:      anomalies,
	}

	for _, a := range anomalies {
		report.BySeverity[a.Severity]++
		report.ByType[a.Type]++
	}

	sort.Slice(report.Anomalies, func(i, j int) bool {
		return severityOrder(report.Anomalies[i].Severity) > severityOrder(report.Anomalies[j].Severity)
	})

	report.Summary = fmt.Sprintf("Detected %d anomalies: %d critical, %d high, %d medium",
		report.TotalAnomalies,
		report.BySeverity["critical"],
		report.BySeverity["high"],
		report.BySeverity["medium"],
	)

	return report
}

func severityOrder(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "high":
		return 2
	case "medium":
		return 1
	default:
		return 0
	}
}
