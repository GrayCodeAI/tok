package benchmarking

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNewParallelRunner(t *testing.T) {
	runner := NewParallelRunner(4)
	if runner == nil {
		t.Fatal("expected runner to be created")
	}

	if runner.maxWorkers != 4 {
		t.Errorf("expected 4 workers, got %d", runner.maxWorkers)
	}

	if runner.runner == nil {
		t.Error("expected internal runner")
	}
}

func TestNewParallelRunnerDefaultWorkers(t *testing.T) {
	runner := NewParallelRunner(0)

	expectedWorkers := runtime.NumCPU()
	if runner.maxWorkers != expectedWorkers {
		t.Errorf("expected %d workers (NumCPU), got %d", expectedWorkers, runner.maxWorkers)
	}
}

func TestParallelRunnerRegisterSuite(t *testing.T) {
	runner := NewParallelRunner(2)
	suite := NewSuite("test-suite")

	runner.RegisterSuite(suite)

	if len(runner.runner.suites) != 1 {
		t.Errorf("expected 1 suite, got %d", len(runner.runner.suites))
	}
}

func TestNewParallelSuite(t *testing.T) {
	ps := NewParallelSuite("parallel-suite", 4)
	if ps == nil {
		t.Fatal("expected parallel suite to be created")
	}

	if ps.name != "parallel-suite" {
		t.Errorf("expected name 'parallel-suite', got %s", ps.name)
	}

	if ps.Parallelism != 4 {
		t.Errorf("expected parallelism 4, got %d", ps.Parallelism)
	}

	if ps.Ordered {
		t.Error("expected ordered to be false by default")
	}
}

func TestParallelSuiteWithOrdered(t *testing.T) {
	ps := NewParallelSuite("test", 4).WithOrdered(true)

	if !ps.Ordered {
		t.Error("expected ordered to be true")
	}
}

func TestNewConcurrencyLimiter(t *testing.T) {
	limiter := NewConcurrencyLimiter(5)
	if limiter == nil {
		t.Fatal("expected limiter to be created")
	}

	if limiter.semaphore == nil {
		t.Error("expected semaphore to be initialized")
	}
}

func TestConcurrencyLimiterAcquireRelease(t *testing.T) {
	limiter := NewConcurrencyLimiter(2)

	// Should not block
	limiter.Acquire()
	limiter.Acquire()

	// Release should not block
	limiter.Release()
	limiter.Release()
}

func TestDefaultParallelOptions(t *testing.T) {
	opts := DefaultParallelOptions()

	if opts.MaxWorkers != runtime.NumCPU() {
		t.Errorf("expected MaxWorkers %d, got %d", runtime.NumCPU(), opts.MaxWorkers)
	}

	if opts.Ordered {
		t.Error("expected Ordered to be false")
	}

	if !opts.ContinueOnError {
		t.Error("expected ContinueOnError to be true")
	}
}

func TestNewParallelExecutor(t *testing.T) {
	opts := DefaultParallelOptions()
	executor := NewParallelExecutor(opts)

	if executor == nil {
		t.Fatal("expected executor to be created")
	}

	if executor.options.MaxWorkers != opts.MaxWorkers {
		t.Errorf("expected MaxWorkers %d, got %d", opts.MaxWorkers, executor.options.MaxWorkers)
	}
}

func TestNewBenchmarkPool(t *testing.T) {
	pool := NewBenchmarkPool(4)
	if pool == nil {
		t.Fatal("expected pool to be created")
	}

	if pool.workers != 4 {
		t.Errorf("expected 4 workers, got %d", pool.workers)
	}

	if pool.workQueue == nil {
		t.Error("expected work queue")
	}

	if pool.resultQueue == nil {
		t.Error("expected result queue")
	}
}

func TestBenchmarkPoolStartStop(t *testing.T) {
	pool := NewBenchmarkPool(2)

	pool.Start()

	// Give workers time to start
	time.Sleep(10 * time.Millisecond)

	pool.Stop()

	// Should be able to stop without panic
}

func TestBenchmarkPoolSubmitAndResults(t *testing.T) {
	pool := NewBenchmarkPool(2)
	pool.Start()
	defer pool.Stop()

	// Create a simple benchmark
	bm := &testBenchmark{name: "test"}

	submitted := pool.Submit(bm)
	if !submitted {
		t.Error("expected benchmark to be submitted")
	}

	// Wait for result
	select {
	case result := <-pool.Results():
		if result.Name != "test" {
			t.Errorf("expected name 'test', got %s", result.Name)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for result")
	}
}

func TestParallelBenchmarkRun(t *testing.T) {
	pb := &ParallelBenchmark{
		Name:        "parallel-test",
		Iterations:  100,
		Parallelism: 4,
		Fn: func(ctx context.Context) error {
			time.Sleep(time.Millisecond)
			return nil
		},
	}

	ctx := context.Background()
	result, err := pb.Run(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Name != "parallel-test" {
		t.Errorf("expected name 'parallel-test', got %s", result.Name)
	}

	if result.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", result.Errors)
	}

	if result.SuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %.2f", result.SuccessRate)
	}
}

func TestParallelBenchmarkRunWithErrors(t *testing.T) {
	errorCount := 0
	pb := &ParallelBenchmark{
		Name:        "parallel-errors",
		Iterations:  100,
		Parallelism: 4,
		Fn: func(ctx context.Context) error {
			if errorCount < 10 {
				errorCount++
				return context.DeadlineExceeded
			}
			return nil
		},
	}

	ctx := context.Background()
	result, err := pb.Run(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Errors != 10 {
		t.Errorf("expected 10 errors, got %d", result.Errors)
	}

	if result.SuccessRate != 90.0 {
		t.Errorf("expected 90%% success rate, got %.2f", result.SuccessRate)
	}
}

func TestParallelBenchmarkDefaultParallelism(t *testing.T) {
	pb := &ParallelBenchmark{
		Name:        "test",
		Iterations:  10,
		Parallelism: 0, // Should default to NumCPU
		Fn:          func(ctx context.Context) error { return nil },
	}

	ctx := context.Background()
	pb.Run(ctx)

	if pb.Parallelism != runtime.NumCPU() {
		t.Errorf("expected parallelism %d, got %d", runtime.NumCPU(), pb.Parallelism)
	}
}

func TestParallelExecutorExecute(t *testing.T) {
	opts := ParallelBenchmarkOptions{
		MaxWorkers:      2,
		Ordered:         false,
		ContinueOnError: true,
	}
	executor := NewParallelExecutor(opts)

	benchmarks := []Benchmark{
		&testBenchmark{name: "bench1"},
		&testBenchmark{name: "bench2"},
		&testBenchmark{name: "bench3"},
	}

	ctx := context.Background()
	results, err := executor.Execute(ctx, benchmarks)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestParallelExecutorExecuteWithTimeout(t *testing.T) {
	opts := ParallelBenchmarkOptions{
		MaxWorkers: 2,
		Timeout:    50 * time.Millisecond,
	}
	executor := NewParallelExecutor(opts)

	benchmarks := []Benchmark{
		&slowBenchmark{name: "slow1", delay: 100 * time.Millisecond},
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, benchmarks)

	if err == nil {
		t.Error("expected timeout error")
	}
}

// Test helpers
type testBenchmark struct {
	name string
}

func (tb *testBenchmark) Name() string        { return tb.name }
func (tb *testBenchmark) Type() BenchmarkType { return TypeCompression }
func (tb *testBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	return &BenchmarkResult{
		Name:      tb.name,
		Type:      TypeCompression,
		Timestamp: time.Now(),
	}, nil
}

type slowBenchmark struct {
	name  string
	delay time.Duration
}

func (sb *slowBenchmark) Name() string        { return sb.name }
func (sb *slowBenchmark) Type() BenchmarkType { return TypeCompression }
func (sb *slowBenchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	select {
	case <-time.After(sb.delay):
		return &BenchmarkResult{Name: sb.name}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func BenchmarkParallelBenchmark(b *testing.B) {
	pb := &ParallelBenchmark{
		Name:        "bench",
		Iterations:  b.N,
		Parallelism: 4,
		Fn:          func(ctx context.Context) error { return nil },
	}

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		pb.Run(ctx)
	}
}
