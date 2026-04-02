// Package benchmarking provides comprehensive performance benchmarking for TokMan
package benchmarking

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// BenchmarkType defines the type of benchmark to run
type BenchmarkType string

const (
	TypeCompression BenchmarkType = "compression"
	TypePipeline    BenchmarkType = "pipeline"
	TypeMemory      BenchmarkType = "memory"
	TypeConcurrency BenchmarkType = "concurrency"
	TypeEndToEnd    BenchmarkType = "e2e"
)

// BenchmarkResult holds the results of a single benchmark run
type BenchmarkResult struct {
	Name         string
	Type         BenchmarkType
	Duration     time.Duration
	TokensIn     int
	TokensOut    int
	Throughput   float64 // tokens per second
	MemoryUsedMB float64
	Allocations  uint64
	LatencyP50   time.Duration
	LatencyP95   time.Duration
	LatencyP99   time.Duration
	Errors       int
	SuccessRate  float64
	Timestamp    time.Time
	Metadata     map[string]string
}

// Suite represents a collection of benchmarks
type Suite struct {
	name       string
	benchmarks []Benchmark
	warmupRuns int
	duration   time.Duration
	iterations int
}

// Benchmark interface for pluggable benchmarks
type Benchmark interface {
	Name() string
	Type() BenchmarkType
	Run(ctx context.Context) (*BenchmarkResult, error)
}

// Runner executes benchmarks and collects results
type Runner struct {
	suites  map[string]*Suite
	results []BenchmarkResult
	mu      sync.RWMutex
	hooks   []Hook
}

// Hook allows extensibility in the benchmark lifecycle
type Hook interface {
	BeforeBenchmark(name string)
	AfterBenchmark(result *BenchmarkResult)
}

// NewRunner creates a new benchmark runner
func NewRunner() *Runner {
	return &Runner{
		suites:  make(map[string]*Suite),
		results: make([]BenchmarkResult, 0),
	}
}

// RegisterSuite adds a benchmark suite
func (r *Runner) RegisterSuite(suite *Suite) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.suites[suite.name] = suite
}

// AddHook registers a lifecycle hook
func (r *Runner) AddHook(hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = append(r.hooks, hook)
}

// RunSuite executes a benchmark suite
func (r *Runner) RunSuite(ctx context.Context, suiteName string) (*SuiteReport, error) {
	r.mu.RLock()
	suite, exists := r.suites[suiteName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("suite %s not found", suiteName)
	}

	report := &SuiteReport{
		Name:      suiteName,
		StartTime: time.Now(),
		Results:   make([]BenchmarkResult, 0, len(suite.benchmarks)),
	}

	// Warmup runs
	if suite.warmupRuns > 0 {
		for i := 0; i < suite.warmupRuns && i < len(suite.benchmarks); i++ {
			_, _ = suite.benchmarks[i].Run(ctx)
		}
	}

	// Actual benchmark runs
	for _, bm := range suite.benchmarks {
		r.runBenchmarkWithHooks(ctx, bm, report)
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	return report, nil
}

func (r *Runner) runBenchmarkWithHooks(ctx context.Context, bm Benchmark, report *SuiteReport) {
	// Call before hooks
	r.mu.RLock()
	hooks := make([]Hook, len(r.hooks))
	copy(hooks, r.hooks)
	r.mu.RUnlock()

	for _, hook := range hooks {
		hook.BeforeBenchmark(bm.Name())
	}

	// Run benchmark
	result, err := bm.Run(ctx)
	if err != nil {
		result = &BenchmarkResult{
			Name:      bm.Name(),
			Type:      bm.Type(),
			Errors:    1,
			Timestamp: time.Now(),
			Metadata:  map[string]string{"error": err.Error()},
		}
	}

	// Call after hooks
	for _, hook := range hooks {
		hook.AfterBenchmark(result)
	}

	r.mu.Lock()
	r.results = append(r.results, *result)
	report.Results = append(report.Results, *result)
	r.mu.Unlock()
}

// SuiteReport contains the results of running a suite
type SuiteReport struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Results   []BenchmarkResult
}

// Summary provides aggregate statistics
func (sr *SuiteReport) Summary() *SummaryStats {
	if len(sr.Results) == 0 {
		return &SummaryStats{}
	}

	stats := &SummaryStats{
		TotalBenchmarks: len(sr.Results),
		ByType:          make(map[BenchmarkType][]BenchmarkResult),
	}

	var totalTokens, totalErrors int
	var totalDuration time.Duration
	var throughputs []float64

	for _, r := range sr.Results {
		totalTokens += r.TokensIn + r.TokensOut
		totalErrors += r.Errors
		totalDuration += r.Duration
		throughputs = append(throughputs, r.Throughput)

		stats.ByType[r.Type] = append(stats.ByType[r.Type], r)
		if r.Errors > 0 {
			stats.FailedBenchmarks++
		}
	}

	stats.TotalTokens = totalTokens
	stats.TotalErrors = totalErrors
	stats.AvgThroughput = average(throughputs)
	stats.SuccessRate = float64(len(sr.Results)-stats.FailedBenchmarks) / float64(len(sr.Results)) * 100

	// Calculate percentiles
	stats.P50Latency = percentile(sr.Results, 0.5)
	stats.P95Latency = percentile(sr.Results, 0.95)
	stats.P99Latency = percentile(sr.Results, 0.99)

	return stats
}

// SummaryStats provides aggregate statistics across benchmarks
type SummaryStats struct {
	TotalBenchmarks  int
	FailedBenchmarks int
	TotalTokens      int
	TotalErrors      int
	AvgThroughput    float64
	SuccessRate      float64
	P50Latency       time.Duration
	P95Latency       time.Duration
	P99Latency       time.Duration
	ByType           map[BenchmarkType][]BenchmarkResult
}

// NewSuite creates a new benchmark suite
func NewSuite(name string) *Suite {
	return &Suite{
		name:       name,
		benchmarks: make([]Benchmark, 0),
		warmupRuns: 1,
		iterations: 10,
		duration:   30 * time.Second,
	}
}

// WithWarmup sets warmup runs
func (s *Suite) WithWarmup(n int) *Suite {
	s.warmupRuns = n
	return s
}

// WithIterations sets iterations per benchmark
func (s *Suite) WithIterations(n int) *Suite {
	s.iterations = n
	return s
}

// WithDuration sets benchmark duration
func (s *Suite) WithDuration(d time.Duration) *Suite {
	s.duration = d
	return s
}

// AddBenchmark adds a benchmark to the suite
func (s *Suite) AddBenchmark(bm Benchmark) *Suite {
	s.benchmarks = append(s.benchmarks, bm)
	return s
}

// MemorySnapshot captures memory statistics
func MemorySnapshot() *MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		NumGC:         m.NumGC,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		StackInuse:    m.StackInuse,
		MSpanInuse:    m.MSpanInuse,
		MCacheInuse:   m.MCacheInuse,
		GCCPUFraction: m.GCCPUFraction,
	}
}

// MemStats holds memory statistics
type MemStats struct {
	Alloc         uint64
	TotalAlloc    uint64
	Sys           uint64
	NumGC         uint32
	HeapAlloc     uint64
	HeapSys       uint64
	HeapInuse     uint64
	StackInuse    uint64
	MSpanInuse    uint64
	MCacheInuse   uint64
	GCCPUFraction float64
}

// Diff returns the difference between two snapshots
func (m *MemStats) Diff(other *MemStats) *MemStats {
	return &MemStats{
		Alloc:      m.Alloc - other.Alloc,
		TotalAlloc: m.TotalAlloc - other.TotalAlloc,
		Sys:        m.Sys - other.Sys,
		HeapAlloc:  m.HeapAlloc - other.HeapAlloc,
		StackInuse: m.StackInuse - other.StackInuse,
	}
}

// Helper functions
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func percentile(results []BenchmarkResult, p float64) time.Duration {
	if len(results) == 0 {
		return 0
	}

	durations := make([]time.Duration, len(results))
	for i, r := range results {
		durations[i] = r.Duration
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	index := int(float64(len(durations)-1) * p)
	return durations[index]
}

// GetRunner returns the global benchmark runner instance
var globalRunner = NewRunner()

func GetRunner() *Runner {
	return globalRunner
}

// ResetRunner resets the global runner (useful for testing)
func ResetRunner() {
	globalRunner = NewRunner()
}
