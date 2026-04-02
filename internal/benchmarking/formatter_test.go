package benchmarking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter(true)
	if formatter == nil {
		t.Fatal("expected formatter to be created")
	}
	if !formatter.Pretty {
		t.Error("expected pretty to be true")
	}
}

func TestJSONFormatterFormatReport(t *testing.T) {
	report := &SuiteReport{
		Name:      "test-report",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Second),
		Duration:  10 * time.Second,
		Results: []BenchmarkResult{
			{
				Name:      "test-1",
				Type:      TypeCompression,
				Duration:  1 * time.Second,
				TokensIn:  1000,
				TokensOut: 500,
			},
		},
	}

	formatter := NewJSONFormatter(true)
	data, err := formatter.FormatReport(report)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty output")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("expected valid JSON: %v", err)
	}

	if result["name"] != "test-report" {
		t.Errorf("expected name 'test-report', got %v", result["name"])
	}
}

func TestJSONFormatterFormatResult(t *testing.T) {
	result := &BenchmarkResult{
		Name:      "test",
		Type:      TypeCompression,
		Duration:  1 * time.Second,
		TokensIn:  1000,
		TokensOut: 500,
	}

	formatter := NewJSONFormatter(false)
	data, err := formatter.FormatResult(result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var decoded BenchmarkResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("expected valid JSON: %v", err)
	}

	if decoded.Name != "test" {
		t.Errorf("expected name 'test', got %s", decoded.Name)
	}
}

func TestNewCSVFormatter(t *testing.T) {
	formatter := NewCSVFormatter()
	if formatter == nil {
		t.Fatal("expected formatter to be created")
	}
}

func TestCSVFormatterFormatReport(t *testing.T) {
	report := &SuiteReport{
		Name:      "test-report",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Second),
		Duration:  10 * time.Second,
		Results: []BenchmarkResult{
			{
				Name:      "test-1",
				Type:      TypeCompression,
				Duration:  1 * time.Second,
				TokensIn:  1000,
				TokensOut: 500,
				Timestamp: time.Now(),
			},
		},
	}

	var buf bytes.Buffer
	formatter := NewCSVFormatter()
	err := formatter.FormatReport(&buf, report)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name,type,duration_ms") {
		t.Error("expected CSV header")
	}

	if !strings.Contains(output, "test-1") {
		t.Error("expected result name in output")
	}
}

func TestCSVFormatterFormatSummary(t *testing.T) {
	stats := &SummaryStats{
		TotalBenchmarks:  10,
		FailedBenchmarks: 1,
		TotalTokens:      5000,
		AvgThroughput:    100.5,
		SuccessRate:      90.0,
	}

	var buf bytes.Buffer
	formatter := NewCSVFormatter()
	err := formatter.FormatSummary(&buf, stats)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "metric,value") {
		t.Error("expected CSV header")
	}

	if !strings.Contains(output, "total_benchmarks") {
		t.Error("expected total_benchmarks in output")
	}
}

func TestNewTableFormatter(t *testing.T) {
	formatter := NewTableFormatter()
	if formatter == nil {
		t.Fatal("expected formatter to be created")
	}
}

func TestTableFormatterFormatReport(t *testing.T) {
	report := &SuiteReport{
		Name:      "test-report",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Second),
		Duration:  10 * time.Second,
		Results: []BenchmarkResult{
			{
				Name:      "test-1",
				Type:      TypeCompression,
				Duration:  1 * time.Second,
				TokensIn:  1000,
				TokensOut: 500,
			},
		},
	}

	formatter := NewTableFormatter()
	output := formatter.FormatReport(report)

	if !strings.Contains(output, "Benchmark Suite: test-report") {
		t.Error("expected suite name in output")
	}

	if !strings.Contains(output, "test-1") {
		t.Error("expected result name in output")
	}

	if !strings.Contains(output, "compression") {
		t.Error("expected type in output")
	}
}

func TestBenchmarkResultToJSON(t *testing.T) {
	result := BenchmarkResult{
		Name:      "test",
		Type:      TypeCompression,
		Duration:  1 * time.Second,
		TokensIn:  1000,
		TokensOut: 500,
		Metadata:  map[string]string{"key": "value"},
	}

	jsonResult := result.ToJSON()

	if jsonResult.Name != "test" {
		t.Errorf("expected name 'test', got %s", jsonResult.Name)
	}

	if jsonResult.TokensIn != 1000 {
		t.Errorf("expected 1000 tokens in, got %d", jsonResult.TokensIn)
	}

	if jsonResult.Metadata["key"] != "value" {
		t.Error("expected metadata to be preserved")
	}
}

func TestExportReport(t *testing.T) {
	report := &SuiteReport{
		Name:      "test",
		StartTime: time.Now(),
		Results: []BenchmarkResult{
			{
				Name:      "test-1",
				Type:      TypeCompression,
				Duration:  1 * time.Second,
				TokensIn:  1000,
				TokensOut: 500,
			},
		},
	}

	// Test JSON format
	var jsonBuf bytes.Buffer
	err := ExportReport(report, "json", &jsonBuf)
	if err != nil {
		t.Errorf("unexpected error for json: %v", err)
	}

	if jsonBuf.Len() == 0 {
		t.Error("expected JSON output")
	}

	// Test CSV format
	var csvBuf bytes.Buffer
	err = ExportReport(report, "csv", &csvBuf)
	if err != nil {
		t.Errorf("unexpected error for csv: %v", err)
	}

	if csvBuf.Len() == 0 {
		t.Error("expected CSV output")
	}

	// Test table format
	var tableBuf bytes.Buffer
	err = ExportReport(report, "table", &tableBuf)
	if err != nil {
		t.Errorf("unexpected error for table: %v", err)
	}

	if tableBuf.Len() == 0 {
		t.Error("expected table output")
	}

	// Test unsupported format
	var failBuf bytes.Buffer
	err = ExportReport(report, "unsupported", &failBuf)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exactly", 7, "exactly"},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q",
				tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestRepeatString(t *testing.T) {
	result := repeatString("-", 5)
	if result != "-----" {
		t.Errorf("expected '-----', got %q", result)
	}
}

func BenchmarkJSONFormatterFormatReport(b *testing.B) {
	report := &SuiteReport{
		Name:    "bench",
		Results: make([]BenchmarkResult, 100),
	}
	for i := range report.Results {
		report.Results[i] = BenchmarkResult{
			Name:     fmt.Sprintf("test-%d", i),
			Type:     TypeCompression,
			Duration: time.Second,
		}
	}

	formatter := NewJSONFormatter(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.FormatReport(report)
	}
}
