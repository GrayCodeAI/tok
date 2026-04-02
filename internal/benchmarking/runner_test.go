package benchmarking

import (
	"context"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatal("expected runner to be created")
	}

	if runner.suites == nil {
		t.Error("expected suites map to be initialized")
	}
}

func TestNewSuite(t *testing.T) {
	suite := NewSuite("test-suite")
	if suite == nil {
		t.Fatal("expected suite to be created")
	}

	if suite.name != "test-suite" {
		t.Errorf("expected suite name 'test-suite', got %s", suite.name)
	}

	if suite.warmupRuns != 1 {
		t.Errorf("expected default warmupRuns 1, got %d", suite.warmupRuns)
	}

	if suite.iterations != 10 {
		t.Errorf("expected default iterations 10, got %d", suite.iterations)
	}
}

func TestSuiteConfiguration(t *testing.T) {
	suite := NewSuite("test").
		WithWarmup(5).
		WithIterations(100).
		WithDuration(60 * time.Second)

	if suite.warmupRuns != 5 {
		t.Errorf("expected warmupRuns 5, got %d", suite.warmupRuns)
	}

	if suite.iterations != 100 {
		t.Errorf("expected iterations 100, got %d", suite.iterations)
	}

	if suite.duration != 60*time.Second {
		t.Errorf("expected duration 60s, got %v", suite.duration)
	}
}

func TestMemorySnapshot(t *testing.T) {
	snapshot := MemorySnapshot()
	if snapshot == nil {
		t.Fatal("expected snapshot to be created")
	}

	// Memory values should be non-negative
	if snapshot.Alloc < 0 {
		t.Error("expected Alloc to be non-negative")
	}
}

func TestSuiteReportSummary(t *testing.T) {
	report := &SuiteReport{
		Name:      "test-report",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Second),
		Duration:  10 * time.Second,
		Results: []BenchmarkResult{
			{
				Name:       "test-1",
				Type:       TypeCompression,
				Duration:   1 * time.Second,
				TokensIn:   1000,
				TokensOut:  500,
				Throughput: 1000,
			},
			{
				Name:       "test-2",
				Type:       TypePipeline,
				Duration:   2 * time.Second,
				TokensIn:   2000,
				Throughput: 1000,
			},
		},
	}

	summary := report.Summary()

	if summary.TotalBenchmarks != 2 {
		t.Errorf("expected 2 benchmarks, got %d", summary.TotalBenchmarks)
	}

	// Summary sums TokensIn + TokensOut: 1000 + 500 + 2000 + 0 = 3500
	if summary.TotalTokens != 3500 {
		t.Errorf("expected 3500 total tokens, got %d", summary.TotalTokens)
	}

	if summary.SuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %.2f", summary.SuccessRate)
	}
}

func TestStandardBenchmarks(t *testing.T) {
	suite := StandardBenchmarks()
	if suite == nil {
		t.Fatal("expected standard benchmarks suite to be created")
	}

	if len(suite.benchmarks) == 0 {
		t.Error("expected benchmarks to be registered")
	}
}

func TestBenchmarkResultMetadata(t *testing.T) {
	result := BenchmarkResult{
		Name:      "test",
		Type:      TypeCompression,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"key": "value",
		},
	}

	if result.Metadata["key"] != "value" {
		t.Error("expected metadata to be accessible")
	}
}

func TestRunnerWithHooks(t *testing.T) {
	runner := NewRunner()

	hook := &testHook{
		beforeFunc: func(name string) {
			// Hook called
		},
	}

	runner.AddHook(hook)

	if len(runner.hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(runner.hooks))
	}
}

type testHook struct {
	beforeFunc func(name string)
	afterFunc  func(result *BenchmarkResult)
}

func (h *testHook) BeforeBenchmark(name string) {
	if h.beforeFunc != nil {
		h.beforeFunc(name)
	}
}

func (h *testHook) AfterBenchmark(result *BenchmarkResult) {
	if h.afterFunc != nil {
		h.afterFunc(result)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	if ctx.Err() != context.Canceled {
		t.Error("expected context to be canceled")
	}
}

func BenchmarkMemorySnapshot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MemorySnapshot()
	}
}

func BenchmarkGenerateTestData(b *testing.B) {
	size := 1024
	for i := 0; i < b.N; i++ {
		_ = generateTestData(size)
	}
}
