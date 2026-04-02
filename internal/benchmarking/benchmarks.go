package benchmarking

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// CompressionBenchmark benchmarks the compression pipeline
type CompressionBenchmark struct {
	name       string
	inputSize  int
	iterations int
	pipeline   *filter.PipelineCoordinator
}

// NewCompressionBenchmark creates a compression benchmark
func NewCompressionBenchmark(name string, inputSize int) *CompressionBenchmark {
	return &CompressionBenchmark{
		name:       name,
		inputSize:  inputSize,
		iterations: 100,
	}
}

func (b *CompressionBenchmark) Name() string {
	return b.name
}

func (b *CompressionBenchmark) Type() BenchmarkType {
	return TypeCompression
}

func (b *CompressionBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	input := generateTestData(b.inputSize)
	tokensIn := len(input) / 4

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	var totalTokensOut int

	for i := 0; i < b.iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Simulate compression
		output, stats := b.pipeline.Process(input)
		totalTokensOut += len(output) / 4
		_ = stats
	}

	duration := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	tokensOut := totalTokensOut / b.iterations
	compressionRatio := float64(tokensIn) / float64(tokensOut)

	return &BenchmarkResult{
		Name:         b.name,
		Type:         TypeCompression,
		Duration:     duration / time.Duration(b.iterations),
		TokensIn:     tokensIn,
		TokensOut:    tokensOut,
		Throughput:   float64(tokensIn*b.iterations) / duration.Seconds(),
		MemoryUsedMB: float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024,
		Allocations:  memAfter.TotalAlloc - memBefore.TotalAlloc,
		SuccessRate:  100.0,
		Timestamp:    time.Now(),
		Metadata: map[string]string{
			"input_size":        fmt.Sprintf("%d", b.inputSize),
			"iterations":        fmt.Sprintf("%d", b.iterations),
			"compression_ratio": fmt.Sprintf("%.2f", compressionRatio),
		},
	}, nil
}

// PipelineBenchmark benchmarks the full pipeline with different configurations
type PipelineBenchmark struct {
	name       string
	mode       filter.Mode
	layers     []string
	input      string
	iterations int
}

// NewPipelineBenchmark creates a pipeline benchmark
func NewPipelineBenchmark(name string, mode filter.Mode, input string) *PipelineBenchmark {
	return &PipelineBenchmark{
		name:       name,
		mode:       mode,
		input:      input,
		iterations: 50,
	}
}

func (b *PipelineBenchmark) Name() string {
	return b.name
}

func (b *PipelineBenchmark) Type() BenchmarkType {
	return TypePipeline
}

func (b *PipelineBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	tokensIn := len(b.input) / 4

	cfg := filter.PipelineConfig{
		Mode: b.mode,
	}

	latencies := make([]time.Duration, 0, b.iterations)
	startTotal := time.Now()

	for i := 0; i < b.iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		iterStart := time.Now()
		pipeline := filter.NewPipelineCoordinator(cfg)
		_, _ = pipeline.Process(b.input)
		latencies = append(latencies, time.Since(iterStart))
	}

	duration := time.Since(startTotal)

	// Calculate percentiles
	sortDurations(latencies)
	p50 := latencies[len(latencies)/2]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	return &BenchmarkResult{
		Name:        b.name,
		Type:        TypePipeline,
		Duration:    duration / time.Duration(b.iterations),
		TokensIn:    tokensIn,
		Throughput:  float64(tokensIn*b.iterations) / duration.Seconds(),
		LatencyP50:  p50,
		LatencyP95:  p95,
		LatencyP99:  p99,
		SuccessRate: 100.0,
		Timestamp:   time.Now(),
		Metadata: map[string]string{
			"mode":       string(b.mode),
			"iterations": fmt.Sprintf("%d", b.iterations),
		},
	}, nil
}

// MemoryBenchmark measures memory usage patterns
type MemoryBenchmark struct {
	name       string
	allocSize  int
	iterations int
}

// NewMemoryBenchmark creates a memory benchmark
func NewMemoryBenchmark(name string, allocSize int) *MemoryBenchmark {
	return &MemoryBenchmark{
		name:       name,
		allocSize:  allocSize,
		iterations: 1000,
	}
}

func (b *MemoryBenchmark) Name() string {
	return b.name
}

func (b *MemoryBenchmark) Type() BenchmarkType {
	return TypeMemory
}

func (b *MemoryBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	var allocations uint64

	for i := 0; i < b.iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Simulate memory allocation pattern
		data := make([]byte, b.allocSize)
		_ = data
		allocations += uint64(b.allocSize)
	}

	duration := time.Since(start)

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	return &BenchmarkResult{
		Name:         b.name,
		Type:         TypeMemory,
		Duration:     duration,
		MemoryUsedMB: float64(memAfter.HeapInuse-memBefore.HeapInuse) / 1024 / 1024,
		Allocations:  allocations,
		SuccessRate:  100.0,
		Timestamp:    time.Now(),
		Metadata: map[string]string{
			"alloc_size":  fmt.Sprintf("%d", b.allocSize),
			"iterations":  fmt.Sprintf("%d", b.iterations),
			"total_alloc": fmt.Sprintf("%d", allocations),
		},
	}, nil
}

// ConcurrencyBenchmark tests concurrent operations
type ConcurrencyBenchmark struct {
	name    string
	workers int
	tasks   int
	workFn  func() error
}

// NewConcurrencyBenchmark creates a concurrency benchmark
func NewConcurrencyBenchmark(name string, workers, tasks int) *ConcurrencyBenchmark {
	return &ConcurrencyBenchmark{
		name:    name,
		workers: workers,
		tasks:   tasks,
	}
}

func (b *ConcurrencyBenchmark) Name() string {
	return b.name
}

func (b *ConcurrencyBenchmark) Type() BenchmarkType {
	return TypeConcurrency
}

func (b *ConcurrencyBenchmark) WithWork(fn func() error) *ConcurrencyBenchmark {
	b.workFn = fn
	return b
}

func (b *ConcurrencyBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	start := time.Now()

	taskCh := make(chan int, b.tasks)
	errCh := make(chan error, b.tasks)

	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < b.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range taskCh {
				if b.workFn != nil {
					if err := b.workFn(); err != nil {
						errCh <- err
					}
				}
			}
		}()
	}

	// Distribute tasks
	for i := 0; i < b.tasks; i++ {
		select {
		case <-ctx.Done():
			close(taskCh)
			return nil, ctx.Err()
		case taskCh <- i:
		}
	}
	close(taskCh)

	wg.Wait()
	close(errCh)

	duration := time.Since(start)

	errors := 0
	for range errCh {
		errors++
	}

	successRate := float64(b.tasks-errors) / float64(b.tasks) * 100

	return &BenchmarkResult{
		Name:        b.name,
		Type:        TypeConcurrency,
		Duration:    duration,
		Throughput:  float64(b.tasks) / duration.Seconds(),
		Errors:      errors,
		SuccessRate: successRate,
		Timestamp:   time.Now(),
		Metadata: map[string]string{
			"workers": fmt.Sprintf("%d", b.workers),
			"tasks":   fmt.Sprintf("%d", b.tasks),
		},
	}, nil
}

// EndToEndBenchmark tests the complete system
type EndToEndBenchmark struct {
	name     string
	scenario string
	steps    []func() error
}

// NewEndToEndBenchmark creates an E2E benchmark
func NewEndToEndBenchmark(name, scenario string) *EndToEndBenchmark {
	return &EndToEndBenchmark{
		name:     name,
		scenario: scenario,
		steps:    make([]func() error, 0),
	}
}

func (b *EndToEndBenchmark) Name() string {
	return b.name
}

func (b *EndToEndBenchmark) Type() BenchmarkType {
	return TypeEndToEnd
}

func (b *EndToEndBenchmark) AddStep(fn func() error) *EndToEndBenchmark {
	b.steps = append(b.steps, fn)
	return b
}

func (b *EndToEndBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	start := time.Now()
	errors := 0

	for _, step := range b.steps {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := step(); err != nil {
			errors++
		}
	}

	duration := time.Since(start)
	successRate := float64(len(b.steps)-errors) / float64(len(b.steps)) * 100

	return &BenchmarkResult{
		Name:        b.name,
		Type:        TypeEndToEnd,
		Duration:    duration,
		Errors:      errors,
		SuccessRate: successRate,
		Timestamp:   time.Now(),
		Metadata: map[string]string{
			"scenario": b.scenario,
			"steps":    fmt.Sprintf("%d", len(b.steps)),
		},
	}, nil
}

// Helper functions
func generateTestData(size int) string {
	var buf bytes.Buffer
	buf.Grow(size)

	for buf.Len() < size {
		buf.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
	}

	return buf.String()[:size]
}

func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

// StandardBenchmarks returns a suite of standard benchmarks
func StandardBenchmarks() *Suite {
	suite := NewSuite("standard").
		WithWarmup(2).
		WithIterations(100).
		WithDuration(60 * time.Second)

	// Compression benchmarks
	suite.AddBenchmark(NewCompressionBenchmark("small_compression", 1024))
	suite.AddBenchmark(NewCompressionBenchmark("medium_compression", 10240))
	suite.AddBenchmark(NewCompressionBenchmark("large_compression", 102400))

	// Pipeline benchmarks
	suite.AddBenchmark(NewPipelineBenchmark("pipeline_minimal", filter.ModeMinimal, generateTestData(5000)))
	suite.AddBenchmark(NewPipelineBenchmark("pipeline_aggressive", filter.ModeAggressive, generateTestData(5000)))

	// Memory benchmarks
	suite.AddBenchmark(NewMemoryBenchmark("memory_small", 1024))
	suite.AddBenchmark(NewMemoryBenchmark("memory_large", 102400))

	// Concurrency benchmarks
	suite.AddBenchmark(NewConcurrencyBenchmark("concurrent_10x100", 10, 100))
	suite.AddBenchmark(NewConcurrencyBenchmark("concurrent_50x500", 50, 500))

	return suite
}
