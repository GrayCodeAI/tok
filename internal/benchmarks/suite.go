package benchmarks

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// BenchmarkResult represents the result of a benchmark run.
type BenchmarkResult struct {
	Name                string
	InputSize           int
	OutputSize          int
	TokensSaved         int
	CompressionRatio    float64
	ProcessingTimeMs    int64
	ThroughputMBPerSec  float64
	CostSavedUSD        float64
	ComparisonVsBaseline float64 // percentage
}

// BenchmarkSuite runs a series of benchmarks.
type BenchmarkSuite struct {
	name        string
	benchmarks  []Benchmark
	results     []*BenchmarkResult
	logger      *slog.Logger
	mu          sync.Mutex
}

// Benchmark represents a single benchmark test.
type Benchmark interface {
	Name() string
	Input() []byte
	Run() *BenchmarkResult
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(name string, logger *slog.Logger) *BenchmarkSuite {
	if logger == nil {
		logger = slog.Default()
	}

	return &BenchmarkSuite{
		name:       name,
		benchmarks: make([]Benchmark, 0),
		results:    make([]*BenchmarkResult, 0),
		logger:     logger,
	}
}

// AddBenchmark adds a benchmark to the suite.
func (bs *BenchmarkSuite) AddBenchmark(b Benchmark) {
	bs.benchmarks = append(bs.benchmarks, b)
}

// Run executes all benchmarks in the suite.
func (bs *BenchmarkSuite) Run() []*BenchmarkResult {
	bs.logger.Info("starting benchmark suite", slog.String("suite", bs.name))

	for _, bench := range bs.benchmarks {
		start := time.Now()

		result := bench.Run()
		if result != nil {
			result.ProcessingTimeMs = int64(time.Since(start).Milliseconds())
			bs.recordResult(result)

			bs.logger.Info("benchmark completed",
				slog.String("name", result.Name),
				slog.Int64("time_ms", result.ProcessingTimeMs),
				slog.Float64("ratio", result.CompressionRatio),
			)
		}
	}

	return bs.results
}

// recordResult records a benchmark result.
func (bs *BenchmarkSuite) recordResult(result *BenchmarkResult) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.results = append(bs.results, result)
}

// Summary returns a summary of all results.
func (bs *BenchmarkSuite) Summary() *BenchmarkSummary {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if len(bs.results) == 0 {
		return &BenchmarkSummary{}
	}

	summary := &BenchmarkSummary{
		TotalBenchmarks: len(bs.results),
		Results:         bs.results,
	}

	// Calculate aggregates
	totalTime := int64(0)
	totalCompression := 0.0
	totalCostSaved := 0.0
	minTime := int64(1<<63 - 1)
	maxTime := int64(0)
	minRatio := 1.0
	maxRatio := 0.0

	for _, result := range bs.results {
		totalTime += result.ProcessingTimeMs
		totalCompression += result.CompressionRatio
		totalCostSaved += result.CostSavedUSD

		if result.ProcessingTimeMs < minTime {
			minTime = result.ProcessingTimeMs
		}
		if result.ProcessingTimeMs > maxTime {
			maxTime = result.ProcessingTimeMs
		}

		if result.CompressionRatio < minRatio {
			minRatio = result.CompressionRatio
		}
		if result.CompressionRatio > maxRatio {
			maxRatio = result.CompressionRatio
		}
	}

	summary.AverageTimeMs = totalTime / int64(len(bs.results))
	summary.MinTimeMs = minTime
	summary.MaxTimeMs = maxTime
	summary.AverageCompressionRatio = totalCompression / float64(len(bs.results))
	summary.MinCompressionRatio = minRatio
	summary.MaxCompressionRatio = maxRatio
	summary.TotalCostSavedUSD = totalCostSaved

	return summary
}

// BenchmarkSummary contains aggregate benchmark statistics.
type BenchmarkSummary struct {
	TotalBenchmarks        int
	Results                []*BenchmarkResult
	AverageTimeMs          int64
	MinTimeMs              int64
	MaxTimeMs              int64
	AverageCompressionRatio float64
	MinCompressionRatio    float64
	MaxCompressionRatio    float64
	TotalCostSavedUSD      float64
}

// FormatResults returns a formatted string of benchmark results.
func (bs *BenchmarkSummary) FormatResults() string {
	output := fmt.Sprintf("=== Benchmark Results ===\n\n")
	output += fmt.Sprintf("Total Benchmarks: %d\n", bs.TotalBenchmarks)
	output += fmt.Sprintf("Average Time: %dms (min: %dms, max: %dms)\n", bs.AverageTimeMs, bs.MinTimeMs, bs.MaxTimeMs)
	output += fmt.Sprintf("Compression Ratio: %.2f%% (min: %.2f%%, max: %.2f%%)\n",
		bs.AverageCompressionRatio*100, bs.MinCompressionRatio*100, bs.MaxCompressionRatio*100)
	output += fmt.Sprintf("Total Cost Saved: $%.2f\n\n", bs.TotalCostSavedUSD)

	output += "Detailed Results:\n"
	output += "Name | Input | Output | Tokens Saved | Compression | Time (ms) | Cost Saved\n"
	output += "---|---|---|---|---|---|---\n"

	for _, result := range bs.Results {
		output += fmt.Sprintf("%s | %d B | %d B | %d | %.1f%% | %d | $%.2f\n",
			result.Name, result.InputSize, result.OutputSize,
			result.TokensSaved, result.CompressionRatio*100,
			result.ProcessingTimeMs, result.CostSavedUSD)
	}

	return output
}

// ComparisonResult compares TokMan against baseline tools.
type ComparisonResult struct {
	Tool             string
	CompressionRatio float64
	ProcessingTimeMs int64
	TokensSavedPerSec float64
	ScoreVsTokman    float64 // 1.0 = equal, >1.0 = better than TokMan
}

// Benchmark standard datasets

var (
	SmallCodeSample = []byte(`
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    fmt.Println(fibonacci(10))
}
`)

	LargeCodeSample = []byte(`
// Standard library imports
import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

// ComplexFunction demonstrates various programming patterns
type DataProcessor struct {
    mu      sync.Mutex
    cache   map[string]interface{}
    logger  *log.Logger
}

func (dp *DataProcessor) Process(data []byte) ([]byte, error) {
    dp.mu.Lock()
    defer dp.mu.Unlock()

    // Process data
    return data, nil
}
`)

	MLLogSample = []byte(`
Epoch 1/100
1000/1000 [==============================] - 12s 12ms/step - loss: 0.5234 - accuracy: 0.7823
Epoch 2/100
1000/1000 [==============================] - 11s 11ms/step - loss: 0.4156 - accuracy: 0.8234
Epoch 3/100
1000/1000 [==============================] - 11s 11ms/step - loss: 0.3421 - accuracy: 0.8567
    `)
)
