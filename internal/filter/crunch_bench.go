package filter

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lakshmanpatel/tok/internal/core"
)

// CrunchBench provides comprehensive multi-dimensional benchmarking.
type CrunchBench struct {
	testInputs []TestInput
	results    []BenchmarkResult
}

// TestInput represents a test case for benchmarking.
type TestInput struct {
	Name        string
	Content     string
	ContentType string
	ExpectedMin float64 // Minimum expected compression ratio
	ExpectedMax float64 // Maximum expected compression ratio
}

// BenchmarkResult holds results for a single test case.
type BenchmarkResult struct {
	TestName           string
	ContentType        string
	OriginalTokens     int
	CompressedTokens   int
	ReductionPct       float64
	CompressionTime    time.Duration
	PerTokenLatency    float64 // microseconds per token
	Reversible         bool
	ReconstructionDiff float64 // 0-1 similarity score
	QualityScore       float64 // 0-1 quality score
	LayerBreakdown     map[string]LayerTiming
}

// LayerTiming tracks per-layer performance.
type LayerTiming struct {
	TokensSaved int
	Duration    time.Duration
}

// BenchmarkReport aggregates all benchmark results.
type BenchmarkReport struct {
	Timestamp       time.Time
	TotalTests      int
	Passed          int
	Failed          int
	OverallStats    AggregateStats
	Results         []BenchmarkResult
	Recommendations []string
}

// AggregateStats provides summary statistics.
type AggregateStats struct {
	AvgCompression    float64
	MinCompression    float64
	MaxCompression    float64
	StdDevCompression float64
	AvgLatency        float64
	TotalTime         time.Duration
	AvgQuality        float64
}

// NewCrunchBench creates a new benchmark instance.
func NewCrunchBench() *CrunchBench {
	return &CrunchBench{
		testInputs: make([]TestInput, 0),
		results:    make([]BenchmarkResult, 0),
	}
}

// Name returns the filter name.
func (cb *CrunchBench) Name() string { return "crunch_bench" }

// Apply is a passthrough - this is a benchmark tool, not a compression layer.
func (cb *CrunchBench) Apply(input string, mode Mode) (string, int) {
	return input, 0
}

// RegisterTestInput adds a test case.
func (cb *CrunchBench) RegisterTestInput(name, content, contentType string, minExpected, maxExpected float64) {
	cb.testInputs = append(cb.testInputs, TestInput{
		Name:        name,
		Content:     content,
		ContentType: contentType,
		ExpectedMin: minExpected,
		ExpectedMax: maxExpected,
	})
}

// RunBenchmark executes benchmarks on all registered inputs.
func (cb *CrunchBench) RunBenchmark(cfg PipelineConfig) *BenchmarkReport {
	report := &BenchmarkReport{
		Timestamp:       time.Now(),
		TotalTests:      len(cb.testInputs),
		Results:         make([]BenchmarkResult, 0, len(cb.testInputs)),
		Recommendations: make([]string, 0),
	}

	pipeline := NewPipelineCoordinator(cfg)

	for _, input := range cb.testInputs {
		result := cb.benchmarkInput(input, pipeline)
		report.Results = append(report.Results, result)

		// Check if test passed
		if result.ReductionPct >= input.ExpectedMin && result.ReductionPct <= input.ExpectedMax {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	report.OverallStats = cb.calculateAggregateStats(report.Results)
	report.Recommendations = cb.generateRecommendations(report.Results, report.OverallStats)

	return report
}

// benchmarkInput benchmarks a single input.
func (cb *CrunchBench) benchmarkInput(input TestInput, pipeline *PipelineCoordinator) BenchmarkResult {
	start := time.Now()
	compressed, stats := pipeline.Process(input.Content)
	elapsed := time.Since(start)

	origTokens := core.EstimateTokens(input.Content)
	compTokens := core.EstimateTokens(compressed)

	reduction := 0.0
	if origTokens > 0 {
		reduction = float64(origTokens-compTokens) / float64(origTokens) * 100
	}

	// Calculate per-token latency
	perTokenLatency := 0.0
	if origTokens > 0 {
		perTokenLatency = float64(elapsed.Microseconds()) / float64(origTokens)
	}

	// Extract layer breakdown
	layerBreakdown := make(map[string]LayerTiming)
	for layerName, layerStat := range stats.LayerStats {
		layerBreakdown[layerName] = LayerTiming{
			TokensSaved: layerStat.TokensSaved,
			Duration:    time.Duration(layerStat.Duration),
		}
	}

	// Estimate quality
	qe := NewQualityEstimator()
	quality := qe.EstimateQuality(input.Content, compressed)

	// Check reversibility (simplified - actual would need full reconstruction)
	reversible := strings.Contains(compressed, "[reversible]") || strings.Contains(compressed, "[")

	return BenchmarkResult{
		TestName:           input.Name,
		ContentType:        input.ContentType,
		OriginalTokens:     origTokens,
		CompressedTokens:   compTokens,
		ReductionPct:       reduction,
		CompressionTime:    elapsed,
		PerTokenLatency:    perTokenLatency,
		Reversible:         reversible,
		ReconstructionDiff: quality, // Using quality as proxy for reconstruction
		QualityScore:       quality,
		LayerBreakdown:     layerBreakdown,
	}
}

// calculateAggregateStats computes summary statistics.
func (cb *CrunchBench) calculateAggregateStats(results []BenchmarkResult) AggregateStats {
	if len(results) == 0 {
		return AggregateStats{}
	}

	var sumCompression, sumLatency, sumQuality float64
	var minCompression, maxCompression float64 = 100, 0
	var totalTime time.Duration

	for _, r := range results {
		sumCompression += r.ReductionPct
		sumLatency += r.PerTokenLatency
		sumQuality += r.QualityScore
		totalTime += r.CompressionTime

		if r.ReductionPct < minCompression {
			minCompression = r.ReductionPct
		}
		if r.ReductionPct > maxCompression {
			maxCompression = r.ReductionPct
		}
	}

	count := float64(len(results))
	avgCompression := sumCompression / count
	avgLatency := sumLatency / count
	avgQuality := sumQuality / count

	// Calculate standard deviation
	var sumSquaredDiff float64
	for _, r := range results {
		diff := r.ReductionPct - avgCompression
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / count)

	return AggregateStats{
		AvgCompression:    avgCompression,
		MinCompression:    minCompression,
		MaxCompression:    maxCompression,
		StdDevCompression: stdDev,
		AvgLatency:        avgLatency,
		TotalTime:         totalTime,
		AvgQuality:        avgQuality,
	}
}

// generateRecommendations creates optimization suggestions.
func (cb *CrunchBench) generateRecommendations(results []BenchmarkResult, stats AggregateStats) []string {
	var recommendations []string

	// Check compression variability
	if stats.StdDevCompression > 20 {
		recommendations = append(recommendations,
			"High compression variability detected. Consider content-type-specific presets.")
	}

	// Check latency
	if stats.AvgLatency > 100 {
		recommendations = append(recommendations,
			"High latency (>100μs/token). Consider disabling slower layers or enabling parallel processing.")
	}

	// Check quality
	if stats.AvgQuality < 0.7 {
		recommendations = append(recommendations,
			"Low average quality score. Consider using more conservative compression settings.")
	}

	// Check per-content-type performance
	byType := make(map[string][]float64)
	for _, r := range results {
		byType[r.ContentType] = append(byType[r.ContentType], r.ReductionPct)
	}

	for contentType, ratios := range byType {
		if len(ratios) > 0 {
			avg := 0.0
			for _, r := range ratios {
				avg += r
			}
			avg /= float64(len(ratios))

			if avg < 10 {
				recommendations = append(recommendations,
					fmt.Sprintf("Low compression for %s (%.1f%%). Consider type-specific optimizations.", contentType, avg))
			}
		}
	}

	return recommendations
}

// FormatReport formats the benchmark report as a string.
func (cb *CrunchBench) FormatReport(report *BenchmarkReport) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║              CrunchBench - Compression Benchmark Report          ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("║ Tests: %d passed, %d failed, %d total\n", report.Passed, report.Failed, report.TotalTests))
	sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║ Aggregate Statistics:\n")
	sb.WriteString(fmt.Sprintf("║   Avg Compression: %.2f%% (std: %.2f%%)\n", report.OverallStats.AvgCompression, report.OverallStats.StdDevCompression))
	sb.WriteString(fmt.Sprintf("║   Range: %.2f%% - %.2f%%\n", report.OverallStats.MinCompression, report.OverallStats.MaxCompression))
	sb.WriteString(fmt.Sprintf("║   Avg Latency: %.2f μs/token\n", report.OverallStats.AvgLatency))
	sb.WriteString(fmt.Sprintf("║   Avg Quality: %.2f\n", report.OverallStats.AvgQuality))
	sb.WriteString(fmt.Sprintf("║   Total Time: %v\n", report.OverallStats.TotalTime))
	sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║ Per-Test Results:\n")

	for _, r := range report.Results {
		status := "✓"
		if r.QualityScore < 0.7 {
			status = "⚠"
		}
		sb.WriteString(fmt.Sprintf("║ %s %-20s (%s): %.1f%% in %v [Q:%.2f]\n",
			status, r.TestName, r.ContentType, r.ReductionPct, r.CompressionTime, r.QualityScore))
	}

	if len(report.Recommendations) > 0 {
		sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
		sb.WriteString("║ Recommendations:\n")
		for _, rec := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("║   • %s\n", rec))
		}
	}

	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// GetBuiltinTestInputs returns a set of standard test inputs.
func GetBuiltinTestInputs() []TestInput {
	return []TestInput{
		{
			Name:        "python_source",
			Content:     getPythonSourceSample(),
			ContentType: "code",
			ExpectedMin: 15,
			ExpectedMax: 50,
		},
		{
			Name:        "json_data",
			Content:     getJSONSample(),
			ContentType: "json",
			ExpectedMin: 30,
			ExpectedMax: 85,
		},
		{
			Name:        "build_logs",
			Content:     getBuildLogSample(),
			ContentType: "log",
			ExpectedMin: 10,
			ExpectedMax: 40,
		},
		{
			Name:        "agent_conversation",
			Content:     getConversationSample(),
			ContentType: "conversation",
			ExpectedMin: 20,
			ExpectedMax: 50,
		},
		{
			Name:        "git_diff",
			Content:     getGitDiffSample(),
			ContentType: "diff",
			ExpectedMin: 10,
			ExpectedMax: 30,
		},
		{
			Name:        "search_results",
			Content:     getSearchResultsSample(),
			ContentType: "search",
			ExpectedMin: 20,
			ExpectedMax: 60,
		},
	}
}

// Sample content generators

func getPythonSourceSample() string {
	return `import os
import sys
from typing import List, Dict, Optional

class DataProcessor:
    """Processes data from various sources."""
    
    def __init__(self, config: Dict[str, any]):
        self.config = config
        self.cache = {}
    
    def process(self, items: List[str]) -> List[Dict]:
        """Process a list of items."""
        results = []
        for item in items:
            result = self._transform(item)
            results.append(result)
        return results
    
    def _transform(self, item: str) -> Dict:
        """Transform a single item."""
        return {"original": item, "processed": item.upper()}

def main():
    processor = DataProcessor({"mode": "test"})
    data = ["hello", "world", "foo", "bar"]
    results = processor.process(data)
    print(results)

if __name__ == "__main__":
    main()
`
}

func getJSONSample() string {
	return `{
  "users": [
    {"id": 1, "name": "Alice", "email": "alice@example.com", "role": "admin"},
    {"id": 2, "name": "Bob", "email": "bob@example.com", "role": "user"},
    {"id": 3, "name": "Charlie", "email": "charlie@example.com", "role": "user"}
  ],
  "settings": {
    "theme": "dark",
    "notifications": true,
    "language": "en"
  }
}`
}

func getBuildLogSample() string {
	return `[INFO] Starting build process
[INFO] Compiling module A
[INFO] Compiling module B
[INFO] Compiling module C
[WARNING] Deprecated API usage in module B
[INFO] Running tests
[INFO] Test 1 passed
[INFO] Test 2 passed
[INFO] Test 3 passed
[ERROR] Test 4 failed - assertion error
[INFO] Retrying...
[INFO] Build completed successfully
`
}

func getConversationSample() string {
	return "User: How do I implement a cache in Go?\n\n" +
		"Assistant: Here's a simple cache implementation:\n\n" +
		"```go\n" +
		"type Cache struct {\n" +
		"    data map[string]interface{}\n" +
		"    mu   sync.RWMutex\n" +
		"}\n\n" +
		"func (c *Cache) Get(key string) (interface{}, bool) {\n" +
		"    c.mu.RLock()\n" +
		"    defer c.mu.RUnlock()\n" +
		"    val, ok := c.data[key]\n" +
		"    return val, ok\n" +
		"}\n" +
		"```\n\n" +
		"User: Thanks! How do I add TTL?\n\n" +
		"Assistant: You can add TTL like this:\n\n" +
		"```go\n" +
		"type CacheItem struct {\n" +
		"    value      interface{}\n" +
		"    expiration time.Time\n" +
		"}\n" +
		"```\n"
}

func getGitDiffSample() string {
	return `diff --git a/main.go b/main.go
index 1234..5678 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,7 @@ func main() {
     fmt.Println("Hello, World!")
-    oldFunc()
+    newFunc()
 }
 
 func helper() {
`
}

func getSearchResultsSample() string {
	return `1. How to use channels in Go - golang.org/doc/effective_go.html#channels
2. Go channels tutorial - gobyexample.com/channels
3. Buffered channels in Go - yourbasic.org/golang/buffered-channels/
4. Channel patterns - github.com/matryer/respond/blob/master/patterns.md
5. Advanced concurrency - go.dev/blog/pipelines
`
}

// Compile-time check
var _ Filter = (*CrunchBench)(nil)
