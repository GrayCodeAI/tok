// Package benchmarking provides chart generation capabilities
package benchmarking

import (
	"fmt"
	"math"
	"strings"
)

// ChartType represents the type of chart
type ChartType string

const (
	ChartBar       ChartType = "bar"
	ChartLine      ChartType = "line"
	ChartScatter   ChartType = "scatter"
	ChartHeatmap   ChartType = "heatmap"
	ChartBox       ChartType = "box"
	ChartHistogram ChartType = "histogram"
)

// Chart represents a generated chart
type Chart struct {
	Type    ChartType
	Title   string
	Width   int
	Height  int
	Data    ChartData
	Options ChartOptions
}

// ChartData represents chart data
type ChartData struct {
	Labels []string
	Series []DataSeries
}

// DataSeries represents a data series
type DataSeries struct {
	Name   string
	Color  string
	Values []float64
}

// ChartOptions represents chart options
type ChartOptions struct {
	XAxis  AxisOptions
	YAxis  AxisOptions
	Legend bool
	Grid   bool
}

// AxisOptions represents axis options
type AxisOptions struct {
	Label string
	Min   *float64
	Max   *float64
}

// ChartGenerator generates charts
type ChartGenerator struct {
	defaultWidth  int
	defaultHeight int
}

// NewChartGenerator creates a new chart generator
func NewChartGenerator() *ChartGenerator {
	return &ChartGenerator{
		defaultWidth:  80,
		defaultHeight: 20,
	}
}

// GenerateBarChart generates an ASCII bar chart
func (cg *ChartGenerator) GenerateBarChart(title string, labels []string, values []float64) string {
	if len(labels) != len(values) {
		return "Error: labels and values must have same length"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("\n%s\n", title))
	output.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	// Find max value for scaling
	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		maxVal = 1
	}

	// Find max label length
	maxLabelLen := 0
	for _, label := range labels {
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
	}

	// Generate bars
	barWidth := 50
	for i, label := range labels {
		value := values[i]
		barLen := int((value / maxVal) * float64(barWidth))
		if barLen < 1 && value > 0 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen)
		padding := strings.Repeat(" ", maxLabelLen-len(label))

		output.WriteString(fmt.Sprintf("%s%s │%s %.2f\n", padding, label, bar, value))
	}

	return output.String()
}

// GenerateLineChart generates an ASCII line chart
func (cg *ChartGenerator) GenerateLineChart(title string, labels []string, series []DataSeries) string {
	if len(series) == 0 {
		return "Error: no data series provided"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("\n%s\n", title))
	output.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	// Find min/max values across all series
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	maxPoints := 0
	for _, s := range series {
		if len(s.Values) > maxPoints {
			maxPoints = len(s.Values)
		}
		for _, v := range s.Values {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}

	if minVal == maxVal {
		maxVal = minVal + 1
	}

	// Chart dimensions
	chartHeight := 15
	chartWidth := maxPoints
	if chartWidth < 10 {
		chartWidth = 10
	}

	// Generate chart grid
	grid := make([][]rune, chartHeight)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot each series
	symbols := []rune{'●', '○', '■', '□', '▲', '△'}
	for si, s := range series {
		symbol := symbols[si%len(symbols)]
		for i, v := range s.Values {
			if i >= chartWidth {
				break
			}
			// Normalize value to chart height (inverted, 0 at bottom)
			normalized := (v - minVal) / (maxVal - minVal)
			y := chartHeight - 1 - int(normalized*float64(chartHeight-1))
			if y >= 0 && y < chartHeight {
				grid[y][i] = symbol
			}
		}
	}

	// Render grid
	for y := 0; y < chartHeight; y++ {
		// Y-axis label
		value := maxVal - (float64(y)/float64(chartHeight-1))*(maxVal-minVal)
		output.WriteString(fmt.Sprintf("%6.1f │", value))

		// Row
		for x := 0; x < chartWidth && x < len(grid[y]); x++ {
			output.WriteString(string(grid[y][x]) + " ")
		}
		output.WriteString("\n")
	}

	// X-axis
	output.WriteString("       └" + strings.Repeat("─", chartWidth*2) + "\n")

	// Legend
	output.WriteString("\nLegend:\n")
	for i, s := range series {
		symbol := symbols[i%len(symbols)]
		output.WriteString(fmt.Sprintf("  %c %s\n", symbol, s.Name))
	}

	return output.String()
}

// GenerateHistogram generates an ASCII histogram
func (cg *ChartGenerator) GenerateHistogram(title string, values []float64, bins int) string {
	if len(values) == 0 {
		return "Error: no values provided"
	}

	if bins <= 0 {
		bins = 10
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("\n%s\n", title))
	output.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	// Find min/max
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Create bins
	binWidth := (maxVal - minVal) / float64(bins)
	if binWidth == 0 {
		binWidth = 1
	}

	binCounts := make([]int, bins)
	for _, v := range values {
		binIndex := int((v - minVal) / binWidth)
		if binIndex >= bins {
			binIndex = bins - 1
		}
		if binIndex < 0 {
			binIndex = 0
		}
		binCounts[binIndex]++
	}

	// Find max count for scaling
	maxCount := 0
	for _, c := range binCounts {
		if c > maxCount {
			maxCount = c
		}
	}

	if maxCount == 0 {
		maxCount = 1
	}

	// Generate histogram
	barWidth := 40
	for i := 0; i < bins; i++ {
		binStart := minVal + float64(i)*binWidth
		binEnd := binStart + binWidth
		count := binCounts[i]

		barLen := int((float64(count) / float64(maxCount)) * float64(barWidth))
		if barLen < 1 && count > 0 {
			barLen = 1
		}
		if count == 0 {
			barLen = 0
		}

		bar := strings.Repeat("█", barLen)
		output.WriteString(fmt.Sprintf("[%6.1f - %6.1f] │%s %d\n", binStart, binEnd, bar, count))
	}

	output.WriteString(fmt.Sprintf("\nTotal: %d values\n", len(values)))

	return output.String()
}

// GenerateComparisonChart generates a comparison chart
func (cg *ChartGenerator) GenerateComparisonChart(title string, baseline, current map[string]float64) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("\n%s\n", title))
	output.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	// Find all keys
	keys := make([]string, 0)
	for k := range baseline {
		keys = append(keys, k)
	}
	for k := range current {
		if _, exists := baseline[k]; !exists {
			keys = append(keys, k)
		}
	}

	// Find max key length
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	// Generate comparison
	output.WriteString(fmt.Sprintf("%-*s │ %10s │ %10s │ %10s\n", maxKeyLen, "Metric", "Baseline", "Current", "Change%"))
	output.WriteString(strings.Repeat("-", maxKeyLen+45) + "\n")

	for _, key := range keys {
		b := baseline[key]
		c := current[key]

		var changePct float64
		if b != 0 {
			changePct = ((c - b) / b) * 100
		}

		symbol := "="
		if changePct > 1 {
			symbol = "▲"
		} else if changePct < -1 {
			symbol = "▼"
		}

		output.WriteString(fmt.Sprintf("%-*s │ %10.2f │ %10.2f │ %9.1f%% %s\n",
			maxKeyLen, key, b, c, changePct, symbol))
	}

	return output.String()
}

// GenerateReportCharts generates charts for a benchmark report
func GenerateReportCharts(report *SuiteReport) string {
	cg := NewChartGenerator()
	var output strings.Builder

	// Throughput chart
	throughputData := make(map[string]float64)
	for _, r := range report.Results {
		throughputData[r.Name] = r.Throughput
	}

	if len(throughputData) > 0 {
		labels := make([]string, 0, len(throughputData))
		values := make([]float64, 0, len(throughputData))
		for name, val := range throughputData {
			labels = append(labels, name)
			values = append(values, val)
		}

		output.WriteString(cg.GenerateBarChart("Throughput by Benchmark", labels, values))
	}

	// Latency distribution histogram
	if len(report.Results) > 0 {
		latencies := make([]float64, len(report.Results))
		for i, r := range report.Results {
			latencies[i] = float64(r.Duration.Milliseconds())
		}

		output.WriteString(cg.GenerateHistogram("Latency Distribution", latencies, 10))
	}

	return output.String()
}

// GenerateTrendChart generates a trend chart from historical data
func GenerateTrendChart(trend *Trend) string {
	cg := NewChartGenerator()

	if len(trend.History) == 0 {
		return "No trend data available"
	}

	// Extract labels and values
	labels := make([]string, len(trend.History))
	values := make([]float64, len(trend.History))

	for i, point := range trend.History {
		labels[i] = point.Timestamp.Format("01/02")
		values[i] = point.Value
	}

	series := []DataSeries{
		{
			Name:   trend.BenchmarkID,
			Values: values,
		},
	}

	return cg.GenerateLineChart(
		fmt.Sprintf("Trend: %s (%s %.1f%%)", trend.BenchmarkID, trend.Direction, trend.ChangePct),
		labels,
		series,
	)
}
