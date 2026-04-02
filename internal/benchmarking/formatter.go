package benchmarking

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// JSONFormatter formats benchmark results as JSON
type JSONFormatter struct {
	Pretty bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{Pretty: pretty}
}

// FormatReport formats a suite report as JSON
func (f *JSONFormatter) FormatReport(report *SuiteReport) ([]byte, error) {
	data := struct {
		Name      string            `json:"name"`
		StartTime time.Time         `json:"start_time"`
		EndTime   time.Time         `json:"end_time"`
		Duration  string            `json:"duration"`
		Results   []BenchmarkResult `json:"results"`
		Summary   *SummaryStats     `json:"summary"`
	}{
		Name:      report.Name,
		StartTime: report.StartTime,
		EndTime:   report.EndTime,
		Duration:  report.Duration.String(),
		Results:   report.Results,
		Summary:   report.Summary(),
	}

	if f.Pretty {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

// FormatResult formats a single benchmark result as JSON
func (f *JSONFormatter) FormatResult(result *BenchmarkResult) ([]byte, error) {
	if f.Pretty {
		return json.MarshalIndent(result, "", "  ")
	}
	return json.Marshal(result)
}

// CSVFormatter formats benchmark results as CSV
type CSVFormatter struct{}

// NewCSVFormatter creates a new CSV formatter
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

// FormatReport formats a suite report as CSV
func (f *CSVFormatter) FormatReport(w io.Writer, report *SuiteReport) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"name", "type", "duration_ms", "tokens_in", "tokens_out",
		"throughput", "memory_mb", "allocations", "latency_p50_ms",
		"latency_p95_ms", "latency_p99_ms", "errors", "success_rate",
		"timestamp", "metadata",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write results
	for _, r := range report.Results {
		record := []string{
			r.Name,
			string(r.Type),
			strconv.FormatInt(r.Duration.Milliseconds(), 10),
			strconv.Itoa(r.TokensIn),
			strconv.Itoa(r.TokensOut),
			strconv.FormatFloat(r.Throughput, 'f', 2, 64),
			strconv.FormatFloat(r.MemoryUsedMB, 'f', 2, 64),
			strconv.FormatUint(r.Allocations, 10),
			strconv.FormatInt(r.LatencyP50.Milliseconds(), 10),
			strconv.FormatInt(r.LatencyP95.Milliseconds(), 10),
			strconv.FormatInt(r.LatencyP99.Milliseconds(), 10),
			strconv.Itoa(r.Errors),
			strconv.FormatFloat(r.SuccessRate, 'f', 2, 64),
			r.Timestamp.Format(time.RFC3339),
			encodeMetadata(r.Metadata),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// FormatSummary formats summary statistics as CSV
func (f *CSVFormatter) FormatSummary(w io.Writer, stats *SummaryStats) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"metric", "value",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write metrics
	metrics := [][]string{
		{"total_benchmarks", strconv.Itoa(stats.TotalBenchmarks)},
		{"failed_benchmarks", strconv.Itoa(stats.FailedBenchmarks)},
		{"total_tokens", strconv.Itoa(stats.TotalTokens)},
		{"total_errors", strconv.Itoa(stats.TotalErrors)},
		{"avg_throughput", strconv.FormatFloat(stats.AvgThroughput, 'f', 2, 64)},
		{"success_rate", strconv.FormatFloat(stats.SuccessRate, 'f', 2, 64)},
		{"p50_latency_ms", strconv.FormatInt(stats.P50Latency.Milliseconds(), 10)},
		{"p95_latency_ms", strconv.FormatInt(stats.P95Latency.Milliseconds(), 10)},
		{"p99_latency_ms", strconv.FormatInt(stats.P99Latency.Milliseconds(), 10)},
	}

	for _, metric := range metrics {
		if err := writer.Write(metric); err != nil {
			return err
		}
	}

	return nil
}

func encodeMetadata(metadata map[string]string) string {
	if metadata == nil {
		return ""
	}
	data, _ := json.Marshal(metadata)
	return string(data)
}

// TableFormatter formats benchmark results as a table
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// FormatReport formats a suite report as a table
func (f *TableFormatter) FormatReport(report *SuiteReport) string {
	output := fmt.Sprintf("Benchmark Suite: %s\n", report.Name)
	output += fmt.Sprintf("Duration: %v\n\n", report.Duration)

	// Format results table
	output += fmt.Sprintf("%-30s %-15s %-12s %-12s %-12s %-10s\n",
		"NAME", "TYPE", "DURATION", "TOKENS_IN", "TOKENS_OUT", "THROUGHPUT")
	output += repeatString("-", 95) + "\n"

	for _, r := range report.Results {
		output += fmt.Sprintf("%-30s %-15s %-12v %-12d %-12d %-10.2f\n",
			truncateString(r.Name, 30),
			r.Type,
			r.Duration,
			r.TokensIn,
			r.TokensOut,
			r.Throughput,
		)
	}

	// Add summary
	output += "\n" + repeatString("-", 95) + "\n"
	summary := report.Summary()
	output += fmt.Sprintf("Total: %d | Failed: %d | Success Rate: %.2f%% | Avg Throughput: %.2f\n",
		summary.TotalBenchmarks,
		summary.FailedBenchmarks,
		summary.SuccessRate,
		summary.AvgThroughput,
	)

	return output
}

// BenchmarkResultJSON is a JSON-serializable version of BenchmarkResult
type BenchmarkResultJSON struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Duration     string            `json:"duration"`
	TokensIn     int               `json:"tokens_in"`
	TokensOut    int               `json:"tokens_out"`
	Throughput   float64           `json:"throughput"`
	MemoryUsedMB float64           `json:"memory_used_mb"`
	Allocations  uint64            `json:"allocations"`
	LatencyP50   string            `json:"latency_p50"`
	LatencyP95   string            `json:"latency_p95"`
	LatencyP99   string            `json:"latency_p99"`
	Errors       int               `json:"errors"`
	SuccessRate  float64           `json:"success_rate"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ToJSON converts BenchmarkResult to JSON format
func (r *BenchmarkResult) ToJSON() BenchmarkResultJSON {
	return BenchmarkResultJSON{
		Name:         r.Name,
		Type:         string(r.Type),
		Duration:     r.Duration.String(),
		TokensIn:     r.TokensIn,
		TokensOut:    r.TokensOut,
		Throughput:   r.Throughput,
		MemoryUsedMB: r.MemoryUsedMB,
		Allocations:  r.Allocations,
		LatencyP50:   r.LatencyP50.String(),
		LatencyP95:   r.LatencyP95.String(),
		LatencyP99:   r.LatencyP99.String(),
		Errors:       r.Errors,
		SuccessRate:  r.SuccessRate,
		Timestamp:    r.Timestamp,
		Metadata:     r.Metadata,
	}
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Comparison represents a comparison between two benchmark reports
type Comparison struct {
	Baseline  *SuiteReport
	Current   *SuiteReport
	Changes   map[string]MetricChange
	Improved  []string
	Regressed []string
	Unchanged []string
}

// MetricChange represents a change in a metric
type MetricChange struct {
	Name      string
	Baseline  float64
	Current   float64
	Change    float64
	ChangePct float64
	Direction string // "improved", "regressed", "unchanged"
}

// CompareReports compares two benchmark reports
func CompareReports(baselineData, currentData []byte) (string, error) {
	// Parse baseline report
	var baseline SuiteReport
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return "", fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Parse current report
	var current SuiteReport
	if err := json.Unmarshal(currentData, &current); err != nil {
		return "", fmt.Errorf("failed to parse current: %w", err)
	}

	comparison := &Comparison{
		Baseline:  &baseline,
		Current:   &current,
		Changes:   make(map[string]MetricChange),
		Improved:  make([]string, 0),
		Regressed: make([]string, 0),
		Unchanged: make([]string, 0),
	}

	// Compare metrics
	baselineSummary := baseline.Summary()
	currentSummary := current.Summary()

	// Compare throughput
	compareMetric(comparison, "throughput", baselineSummary.AvgThroughput, currentSummary.AvgThroughput, true)

	// Compare success rate
	compareMetric(comparison, "success_rate", baselineSummary.SuccessRate, currentSummary.SuccessRate, true)

	// Compare P50 latency
	compareMetric(comparison, "p50_latency", float64(baselineSummary.P50Latency), float64(currentSummary.P50Latency), false)

	// Compare P95 latency
	compareMetric(comparison, "p95_latency", float64(baselineSummary.P95Latency), float64(currentSummary.P95Latency), false)

	// Compare P99 latency
	compareMetric(comparison, "p99_latency", float64(baselineSummary.P99Latency), float64(currentSummary.P99Latency), false)

	return formatComparison(comparison), nil
}

func compareMetric(c *Comparison, name string, baseline, current float64, higherIsBetter bool) {
	var change, changePct float64
	if baseline != 0 {
		change = current - baseline
		changePct = (change / baseline) * 100
	}

	direction := "unchanged"
	if change > 0.01 {
		if higherIsBetter {
			direction = "improved"
		} else {
			direction = "regressed"
		}
	} else if change < -0.01 {
		if higherIsBetter {
			direction = "regressed"
		} else {
			direction = "improved"
		}
	}

	c.Changes[name] = MetricChange{
		Name:      name,
		Baseline:  baseline,
		Current:   current,
		Change:    change,
		ChangePct: changePct,
		Direction: direction,
	}

	switch direction {
	case "improved":
		c.Improved = append(c.Improved, name)
	case "regressed":
		c.Regressed = append(c.Regressed, name)
	default:
		c.Unchanged = append(c.Unchanged, name)
	}
}

func formatComparison(c *Comparison) string {
	output := "Benchmark Comparison\n"
	output += "====================\n\n"

	output += fmt.Sprintf("Baseline: %s\n", c.Baseline.Name)
	output += fmt.Sprintf("Current:  %s\n\n", c.Current.Name)

	// Summary
	output += fmt.Sprintf("Improved:  %d\n", len(c.Improved))
	output += fmt.Sprintf("Regressed: %d\n", len(c.Regressed))
	output += fmt.Sprintf("Unchanged: %d\n\n", len(c.Unchanged))

	// Changes
	output += "Metric Changes:\n"
	output += fmt.Sprintf("%-20s %12s %12s %12s %12s\n", "Metric", "Baseline", "Current", "Change", "Change %")
	output += repeatString("-", 70) + "\n"

	for _, change := range c.Changes {
		symbol := "="
		if change.Direction == "improved" {
			symbol = "▲"
		} else if change.Direction == "regressed" {
			symbol = "▼"
		}

		output += fmt.Sprintf("%-20s %12.2f %12.2f %12.2f %11.2f%% %s\n",
			change.Name,
			change.Baseline,
			change.Current,
			change.Change,
			change.ChangePct,
			symbol,
		)
	}

	return output
}

// ExportReport exports a report in the specified format
func ExportReport(report *SuiteReport, format string, w io.Writer) error {
	switch format {
	case "json":
		formatter := NewJSONFormatter(true)
		data, err := formatter.FormatReport(report)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err

	case "csv":
		formatter := NewCSVFormatter()
		return formatter.FormatReport(w, report)

	case "table":
		formatter := NewTableFormatter()
		_, err := w.Write([]byte(formatter.FormatReport(report)))
		return err

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
