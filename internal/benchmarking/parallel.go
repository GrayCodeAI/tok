package benchmarking

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ParallelRunner executes benchmarks in parallel
type ParallelRunner struct {
	runner     *Runner
	maxWorkers int
	sem        chan struct{}
}

// NewParallelRunner creates a new parallel benchmark runner
func NewParallelRunner(maxWorkers int) *ParallelRunner {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	return &ParallelRunner{
		runner:     NewRunner(),
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
	}
}

// RegisterSuite adds a benchmark suite
func (pr *ParallelRunner) RegisterSuite(suite *Suite) {
	pr.runner.RegisterSuite(suite)
}

// AddHook registers a lifecycle hook
func (pr *ParallelRunner) AddHook(hook Hook) {
	pr.runner.AddHook(hook)
}

// RunSuite executes a benchmark suite in parallel
func (pr *ParallelRunner) RunSuite(ctx context.Context, suiteName string) (*SuiteReport, error) {
	pr.runner.mu.RLock()
	suite, exists := pr.runner.suites[suiteName]
	pr.runner.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("suite %s not found", suiteName)
	}

	report := &SuiteReport{
		Name:      suiteName,
		StartTime: time.Now(),
		Results:   make([]BenchmarkResult, 0, len(suite.benchmarks)),
	}

	// Warmup runs (sequential for consistency)
	if suite.warmupRuns > 0 {
		for i := 0; i < suite.warmupRuns && i < len(suite.benchmarks); i++ {
			_, _ = suite.benchmarks[i].Run(ctx)
		}
	}

	// Run benchmarks in parallel
	results := pr.runParallel(ctx, suite.benchmarks, suite)
	report.Results = results

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	return report, nil
}

func (pr *ParallelRunner) runParallel(ctx context.Context, benchmarks []Benchmark, suite *Suite) []BenchmarkResult {
	results := make([]BenchmarkResult, len(benchmarks))
	var wg sync.WaitGroup
	resultMu := sync.Mutex{}

	for i, bm := range benchmarks {
		wg.Add(1)
		go func(index int, benchmark Benchmark) {
			defer wg.Done()

			// Acquire semaphore
			pr.sem <- struct{}{}
			defer func() { <-pr.sem }()

			// Check context
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Run benchmark with hooks
			result := pr.runBenchmarkWithHooks(ctx, benchmark)

			resultMu.Lock()
			results[index] = *result
			resultMu.Unlock()
		}(i, bm)
	}

	wg.Wait()
	return results
}

func (pr *ParallelRunner) runBenchmarkWithHooks(ctx context.Context, bm Benchmark) *BenchmarkResult {
	// Call before hooks
	pr.runner.mu.RLock()
	hooks := make([]Hook, len(pr.runner.hooks))
	copy(hooks, pr.runner.hooks)
	pr.runner.mu.RUnlock()

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

	pr.runner.mu.Lock()
	pr.runner.results = append(pr.runner.results, *result)
	pr.runner.mu.Unlock()

	return result
}

// ParallelSuite extends Suite with parallel execution options
type ParallelSuite struct {
	*Suite
	Parallelism int
	Ordered     bool
}

// NewParallelSuite creates a parallel benchmark suite
func NewParallelSuite(name string, parallelism int) *ParallelSuite {
	return &ParallelSuite{
		Suite:       NewSuite(name),
		Parallelism: parallelism,
		Ordered:     false,
	}
}

// WithOrdered sets ordered execution (preserves result order)
func (ps *ParallelSuite) WithOrdered(ordered bool) *ParallelSuite {
	ps.Ordered = ordered
	return ps
}

// ParallelBenchmarkResult holds results from parallel execution
type ParallelBenchmarkResult struct {
	BenchmarkResult
	WorkerID  int
	StartTime time.Time
	EndTime   time.Time
}

// ConcurrencyLimiter limits concurrent benchmark execution
type ConcurrencyLimiter struct {
	semaphore chan struct{}
}

// NewConcurrencyLimiter creates a new concurrency limiter
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Acquire acquires a slot
func (cl *ConcurrencyLimiter) Acquire() {
	cl.semaphore <- struct{}{}
}

// Release releases a slot
func (cl *ConcurrencyLimiter) Release() {
	<-cl.semaphore
}

// ParallelBenchmarkOptions configures parallel execution
type ParallelBenchmarkOptions struct {
	MaxWorkers      int
	Ordered         bool
	ContinueOnError bool
	Timeout         time.Duration
}

// DefaultParallelOptions returns default parallel options
func DefaultParallelOptions() ParallelBenchmarkOptions {
	return ParallelBenchmarkOptions{
		MaxWorkers:      runtime.NumCPU(),
		Ordered:         false,
		ContinueOnError: true,
		Timeout:         0,
	}
}

// ParallelExecutor manages parallel benchmark execution
type ParallelExecutor struct {
	options ParallelBenchmarkOptions
	hooks   []Hook
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(options ParallelBenchmarkOptions) *ParallelExecutor {
	return &ParallelExecutor{
		options: options,
		hooks:   make([]Hook, 0),
	}
}

// AddHook adds a lifecycle hook
func (pe *ParallelExecutor) AddHook(hook Hook) {
	pe.hooks = append(pe.hooks, hook)
}

// Execute runs benchmarks in parallel
func (pe *ParallelExecutor) Execute(ctx context.Context, benchmarks []Benchmark) ([]BenchmarkResult, error) {
	if pe.options.MaxWorkers <= 0 {
		pe.options.MaxWorkers = runtime.NumCPU()
	}

	// Create worker pool
	sem := make(chan struct{}, pe.options.MaxWorkers)
	results := make([]BenchmarkResult, len(benchmarks))
	var wg sync.WaitGroup
	errChan := make(chan error, len(benchmarks))

	// Apply timeout if specified
	if pe.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pe.options.Timeout)
		defer cancel()
	}

	for i, bm := range benchmarks {
		wg.Add(1)
		go func(index int, benchmark Benchmark) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				errChan <- fmt.Errorf("benchmark %s cancelled", benchmark.Name())
				return
			default:
			}

			result := pe.executeBenchmark(ctx, benchmark)
			results[index] = *result

			if result.Errors > 0 && !pe.options.ContinueOnError {
				errChan <- fmt.Errorf("benchmark %s failed", benchmark.Name())
			}
		}(i, bm)
	}

	// Wait for completion
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("parallel execution had %d errors", len(errs))
	}

	return results, nil
}

func (pe *ParallelExecutor) executeBenchmark(ctx context.Context, bm Benchmark) *BenchmarkResult {
	// Call before hooks
	for _, hook := range pe.hooks {
		hook.BeforeBenchmark(bm.Name())
	}

	// Execute
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
	for _, hook := range pe.hooks {
		hook.AfterBenchmark(result)
	}

	return result
}

// BenchmarkPool manages a pool of benchmark workers
type BenchmarkPool struct {
	workers     int
	workQueue   chan Benchmark
	resultQueue chan BenchmarkResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBenchmarkPool creates a new benchmark pool
func NewBenchmarkPool(workers int) *BenchmarkPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &BenchmarkPool{
		workers:     workers,
		workQueue:   make(chan Benchmark, workers*2),
		resultQueue: make(chan BenchmarkResult, workers*2),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker pool
func (bp *BenchmarkPool) Start() {
	for i := 0; i < bp.workers; i++ {
		bp.wg.Add(1)
		go bp.worker(i)
	}
}

// Stop stops the worker pool
func (bp *BenchmarkPool) Stop() {
	bp.cancel()
	bp.wg.Wait()
	close(bp.resultQueue)
}

// Submit submits a benchmark to the pool
func (bp *BenchmarkPool) Submit(bm Benchmark) bool {
	select {
	case bp.workQueue <- bm:
		return true
	case <-bp.ctx.Done():
		return false
	}
}

// Results returns the result channel
func (bp *BenchmarkPool) Results() <-chan BenchmarkResult {
	return bp.resultQueue
}

func (bp *BenchmarkPool) worker(id int) {
	defer bp.wg.Done()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case bm, ok := <-bp.workQueue:
			if !ok {
				return
			}

			result, err := bm.Run(bp.ctx)
			if err != nil {
				result = &BenchmarkResult{
					Name:      bm.Name(),
					Type:      bm.Type(),
					Errors:    1,
					Timestamp: time.Now(),
					Metadata:  map[string]string{"error": err.Error(), "worker": fmt.Sprintf("%d", id)},
				}
			}

			select {
			case bp.resultQueue <- *result:
			case <-bp.ctx.Done():
				return
			}
		}
	}
}

// ParallelBenchmark runs a single benchmark with parallel iterations
type ParallelBenchmark struct {
	Name        string
	Iterations  int
	Parallelism int
	Fn          func(ctx context.Context) error
}

// Run executes parallel iterations
func (pb *ParallelBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	if pb.Parallelism <= 0 {
		pb.Parallelism = runtime.NumCPU()
	}

	start := time.Now()
	var totalErrors atomic.Int64
	var wg sync.WaitGroup

	// Distribute iterations across workers
	iterationsPerWorker := pb.Iterations / pb.Parallelism
	remaining := pb.Iterations % pb.Parallelism

	for i := 0; i < pb.Parallelism; i++ {
		wg.Add(1)
		count := iterationsPerWorker
		if i < remaining {
			count++
		}

		go func(workerID, iterCount int) {
			defer wg.Done()

			for j := 0; j < iterCount; j++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := pb.Fn(ctx); err != nil {
					totalErrors.Add(1)
				}
			}
		}(i, count)
	}

	wg.Wait()
	duration := time.Since(start)

	errors := int(totalErrors.Load())
	return &BenchmarkResult{
		Name:        pb.Name,
		Duration:    duration,
		Errors:      errors,
		SuccessRate: float64(pb.Iterations-errors) / float64(pb.Iterations) * 100,
		Timestamp:   time.Now(),
		Metadata: map[string]string{
			"parallelism": fmt.Sprintf("%d", pb.Parallelism),
			"iterations":  fmt.Sprintf("%d", pb.Iterations),
		},
	}, nil
}
