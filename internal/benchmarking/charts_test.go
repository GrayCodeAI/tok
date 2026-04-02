package benchmarking

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewChartGenerator(t *testing.T) {
	cg := NewChartGenerator()
	if cg == nil {
		t.Fatal("expected chart generator to be created")
	}

	if cg.defaultWidth != 80 {
		t.Errorf("expected default width 80, got %d", cg.defaultWidth)
	}
}

func TestGenerateBarChart(t *testing.T) {
	cg := NewChartGenerator()

	labels := []string{"test1", "test2", "test3"}
	values := []float64{100, 200, 150}

	chart := cg.GenerateBarChart("Test Chart", labels, values)

	if chart == "" {
		t.Error("expected non-empty chart")
	}

	if !strings.Contains(chart, "Test Chart") {
		t.Error("expected chart title")
	}

	if !strings.Contains(chart, "test1") {
		t.Error("expected label in chart")
	}

	if !strings.Contains(chart, "100") {
		t.Error("expected value in chart")
	}
}

func TestGenerateBarChartMismatchedLengths(t *testing.T) {
	cg := NewChartGenerator()

	labels := []string{"test1", "test2"}
	values := []float64{100}

	chart := cg.GenerateBarChart("Test", labels, values)

	if !strings.Contains(chart, "Error") {
		t.Error("expected error message")
	}
}

func TestGenerateLineChart(t *testing.T) {
	cg := NewChartGenerator()

	labels := []string{"A", "B", "C", "D"}
	series := []DataSeries{
		{
			Name:   "Series 1",
			Values: []float64{10, 20, 15, 25},
		},
	}

	chart := cg.GenerateLineChart("Line Chart", labels, series)

	if chart == "" {
		t.Error("expected non-empty chart")
	}

	if !strings.Contains(chart, "Line Chart") {
		t.Error("expected chart title")
	}

	if !strings.Contains(chart, "Legend") {
		t.Error("expected legend")
	}
}

func TestGenerateLineChartEmptySeries(t *testing.T) {
	cg := NewChartGenerator()

	chart := cg.GenerateLineChart("Test", []string{}, []DataSeries{})

	if !strings.Contains(chart, "Error") {
		t.Error("expected error for empty series")
	}
}

func TestGenerateHistogram(t *testing.T) {
	cg := NewChartGenerator()

	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	chart := cg.GenerateHistogram("Histogram", values, 5)

	if chart == "" {
		t.Error("expected non-empty chart")
	}

	if !strings.Contains(chart, "Histogram") {
		t.Error("expected chart title")
	}

	if !strings.Contains(chart, "Total") {
		t.Error("expected total count")
	}
}

func TestGenerateHistogramEmptyValues(t *testing.T) {
	cg := NewChartGenerator()

	chart := cg.GenerateHistogram("Test", []float64{}, 5)

	if !strings.Contains(chart, "Error") {
		t.Error("expected error for empty values")
	}
}

func TestGenerateComparisonChart(t *testing.T) {
	cg := NewChartGenerator()

	baseline := map[string]float64{
		"metric1": 100,
		"metric2": 200,
	}

	current := map[string]float64{
		"metric1": 120,
		"metric2": 180,
	}

	chart := cg.GenerateComparisonChart("Comparison", baseline, current)

	if chart == "" {
		t.Error("expected non-empty chart")
	}

	if !strings.Contains(chart, "Comparison") {
		t.Error("expected chart title")
	}

	if !strings.Contains(chart, "metric1") {
		t.Error("expected metrics in chart")
	}

	if !strings.Contains(chart, "Baseline") {
		t.Error("expected baseline column")
	}
}

func TestGenerateReportCharts(t *testing.T) {
	report := &SuiteReport{
		Name: "Test Report",
		Results: []BenchmarkResult{
			{Name: "bench1", Throughput: 100, Duration: 1000000000},
			{Name: "bench2", Throughput: 200, Duration: 2000000000},
			{Name: "bench3", Throughput: 150, Duration: 1500000000},
		},
	}

	charts := GenerateReportCharts(report)

	if charts == "" {
		t.Error("expected non-empty charts")
	}

	if !strings.Contains(charts, "Throughput") {
		t.Error("expected throughput chart")
	}
}

func TestGenerateTrendChart(t *testing.T) {
	trend := &Trend{
		BenchmarkID: "test-bench",
		Direction:   TrendUp,
		ChangePct:   15.5,
		History: []HistoricalPoint{
			{Value: 100},
			{Value: 110},
			{Value: 115},
		},
	}

	chart := GenerateTrendChart(trend)

	if chart == "" {
		t.Error("expected non-empty chart")
	}

	if !strings.Contains(chart, "test-bench") {
		t.Error("expected benchmark name")
	}
}

func TestGenerateTrendChartEmptyHistory(t *testing.T) {
	trend := &Trend{
		BenchmarkID: "test-bench",
		History:     []HistoricalPoint{},
	}

	chart := GenerateTrendChart(trend)

	if !strings.Contains(chart, "No trend data") {
		t.Error("expected no data message")
	}
}

func TestChartDataSeries(t *testing.T) {
	series := DataSeries{
		Name:   "Test Series",
		Color:  "blue",
		Values: []float64{1, 2, 3, 4, 5},
	}

	if series.Name != "Test Series" {
		t.Errorf("expected name 'Test Series', got %s", series.Name)
	}

	if len(series.Values) != 5 {
		t.Errorf("expected 5 values, got %d", len(series.Values))
	}
}

func BenchmarkGenerateBarChart(b *testing.B) {
	cg := NewChartGenerator()
	labels := make([]string, 20)
	values := make([]float64, 20)

	for i := 0; i < 20; i++ {
		labels[i] = fmt.Sprintf("bench-%d", i)
		values[i] = float64(i * 10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GenerateBarChart("Benchmark", labels, values)
	}
}

func BenchmarkGenerateHistogram(b *testing.B) {
	cg := NewChartGenerator()
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GenerateHistogram("Histogram", values, 10)
	}
}
