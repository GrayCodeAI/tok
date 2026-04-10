package filter

import (
	"strings"
	"testing"
)

func TestCrunchBench_New(t *testing.T) {
	cb := NewCrunchBench()
	if cb == nil {
		t.Fatal("expected non-nil CrunchBench")
	}
	if cb.testInputs == nil {
		t.Error("expected testInputs to be initialized")
	}
	if cb.results == nil {
		t.Error("expected results to be initialized")
	}
}

func TestCrunchBench_Name(t *testing.T) {
	cb := NewCrunchBench()
	if cb.Name() != "crunch_bench" {
		t.Errorf("expected name 'crunch_bench', got '%s'", cb.Name())
	}
}

func TestCrunchBench_Apply(t *testing.T) {
	cb := NewCrunchBench()
	input := "test content"
	output, saved := cb.Apply(input, ModeMinimal)

	// Apply is a passthrough
	if output != input {
		t.Error("expected passthrough")
	}
	if saved != 0 {
		t.Error("expected 0 savings")
	}
}

func TestCrunchBench_RegisterTestInput(t *testing.T) {
	cb := NewCrunchBench()
	cb.RegisterTestInput("test", "content", "code", 10, 50)

	if len(cb.testInputs) != 1 {
		t.Errorf("expected 1 test input, got %d", len(cb.testInputs))
	}

	input := cb.testInputs[0]
	if input.Name != "test" {
		t.Error("expected name to match")
	}
	if input.Content != "content" {
		t.Error("expected content to match")
	}
	if input.ContentType != "code" {
		t.Error("expected content type to match")
	}
	if input.ExpectedMin != 10 {
		t.Error("expected min to match")
	}
	if input.ExpectedMax != 50 {
		t.Error("expected max to match")
	}
}

func TestCrunchBench_RunBenchmark(t *testing.T) {
	cb := NewCrunchBench()

	// Register a simple test input
	cb.RegisterTestInput("simple", "This is test content for benchmarking.", "text", 0, 50)

	cfg := PipelineConfig{
		Mode:            ModeMinimal,
		EnableEntropy:   true,
		EnablePerplexity: true,
	}

	report := cb.RunBenchmark(cfg)

	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if report.TotalTests != 1 {
		t.Errorf("expected 1 test, got %d", report.TotalTests)
	}

	if len(report.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(report.Results))
	}
}

func TestCrunchBench_BenchmarkResult(t *testing.T) {
	cb := NewCrunchBench()

	// Register test with known content
	content := strings.Repeat("word ", 100)
	cb.RegisterTestInput("test", content, "text", 0, 80)

	cfg := PipelineConfig{
		Mode:           ModeMinimal,
		EnableEntropy:  true,
	}

	report := cb.RunBenchmark(cfg)

	if len(report.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	result := report.Results[0]

	// Check basic fields
	if result.TestName != "test" {
		t.Error("expected test name to match")
	}
	if result.ContentType != "text" {
		t.Error("expected content type to match")
	}
	if result.OriginalTokens == 0 {
		t.Error("expected non-zero original tokens")
	}
}

func TestCrunchBench_CalculateAggregateStats(t *testing.T) {
	cb := NewCrunchBench()

	results := []BenchmarkResult{
		{ReductionPct: 20.0, QualityScore: 0.8, PerTokenLatency: 10.0},
		{ReductionPct: 30.0, QualityScore: 0.9, PerTokenLatency: 15.0},
		{ReductionPct: 40.0, QualityScore: 0.7, PerTokenLatency: 12.0},
	}

	stats := cb.calculateAggregateStats(results)

	// Check average
	if stats.AvgCompression != 30.0 {
		t.Errorf("expected avg compression 30.0, got %.2f", stats.AvgCompression)
	}

	// Check min/max
	if stats.MinCompression != 20.0 {
		t.Errorf("expected min compression 20.0, got %.2f", stats.MinCompression)
	}
	if stats.MaxCompression != 40.0 {
		t.Errorf("expected max compression 40.0, got %.2f", stats.MaxCompression)
	}

	// Check standard deviation is calculated
	if stats.StdDevCompression == 0 {
		t.Error("expected non-zero standard deviation")
	}
}

func TestCrunchBench_GenerateRecommendations(t *testing.T) {
	cb := NewCrunchBench()

	results := []BenchmarkResult{
		{ContentType: "text", ReductionPct: 5.0, QualityScore: 0.6},
	}

	stats := AggregateStats{
		AvgCompression:   5.0,
		StdDevCompression: 25.0,
		AvgLatency:       150.0,
		AvgQuality:       0.6,
	}

	recommendations := cb.generateRecommendations(results, stats)

	// Should have recommendations for low compression
	if len(recommendations) == 0 {
		t.Error("expected recommendations for low performance")
	}
}

func TestCrunchBench_FormatReport(t *testing.T) {
	cb := NewCrunchBench()

	report := &BenchmarkReport{
		TotalTests: 2,
		Passed:     2,
		Failed:     0,
		OverallStats: AggregateStats{
			AvgCompression: 30.0,
			AvgLatency:     10.0,
			AvgQuality:     0.8,
		},
		Results: []BenchmarkResult{
			{TestName: "test1", ReductionPct: 25.0, QualityScore: 0.8},
			{TestName: "test2", ReductionPct: 35.0, QualityScore: 0.9},
		},
	}

	formatted := cb.FormatReport(report)

	if formatted == "" {
		t.Error("expected non-empty formatted report")
	}

	// Should contain key information
	if !strings.Contains(formatted, "test1") {
		t.Error("expected report to contain test name")
	}
	if !strings.Contains(formatted, "30.0") {
		t.Error("expected report to contain avg compression")
	}
}

func TestCrunchBench_GetBuiltinTestInputs(t *testing.T) {
	inputs := GetBuiltinTestInputs()

	if len(inputs) == 0 {
		t.Error("expected built-in test inputs")
	}

	// Should have various content types
	types := make(map[string]bool)
	for _, input := range inputs {
		types[input.ContentType] = true
	}

	expectedTypes := []string{"code", "json", "log", "conversation", "diff", "search"}
	for _, et := range expectedTypes {
		if !types[et] {
			t.Errorf("expected built-in input with type '%s'", et)
		}
	}
}

func TestCrunchBench_SampleContent(t *testing.T) {
	// Test Python sample
	python := getPythonSourceSample()
	if python == "" {
		t.Error("expected non-empty Python sample")
	}
	if !strings.Contains(python, "func") && !strings.Contains(python, "def") {
		t.Error("expected Python sample to contain function definitions")
	}

	// Test JSON sample
	json := getJSONSample()
	if json == "" {
		t.Error("expected non-empty JSON sample")
	}
	if !strings.Contains(json, "{") {
		t.Error("expected JSON sample to contain braces")
	}

	// Test log sample
	log := getBuildLogSample()
	if log == "" {
		t.Error("expected non-empty log sample")
	}
	if !strings.Contains(log, "[INFO]") {
		t.Error("expected log sample to contain log levels")
	}

	// Test diff sample
	diff := getGitDiffSample()
	if diff == "" {
		t.Error("expected non-empty diff sample")
	}
	if !strings.Contains(diff, "diff --git") {
		t.Error("expected diff sample to contain git diff markers")
	}
}

func TestCrunchBench_EmptyResults(t *testing.T) {
	cb := NewCrunchBench()

	stats := cb.calculateAggregateStats([]BenchmarkResult{})

	// Should handle empty results gracefully
	if stats.AvgCompression != 0 {
		t.Error("expected zero avg for empty results")
	}
}

func TestCrunchBench_LayerBreakdown(t *testing.T) {
	result := BenchmarkResult{
		LayerBreakdown: map[string]LayerTiming{
			"entropy": {TokensSaved: 100, Duration: 1000000},
			"h2o":     {TokensSaved: 50, Duration: 500000},
		},
	}

	if len(result.LayerBreakdown) != 2 {
		t.Error("expected 2 layers in breakdown")
	}

	if result.LayerBreakdown["entropy"].TokensSaved != 100 {
		t.Error("expected entropy tokens saved to match")
	}
}
