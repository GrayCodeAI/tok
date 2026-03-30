package benchmarks

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/tests/fixtures"
)

// Benchmark_CLIOutput_Fixtures tests compression on real-world CLI outputs
// These benchmarks measure token reduction on common development tool outputs

// Benchmark_GitStatus tests git status compression
func Benchmark_GitStatus(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeMinimal,
		EnableEntropy:  true,
		EnableH2O:      true,
		EnableCompaction: true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.GitStatusOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_GitLog tests git log compression
func Benchmark_GitLog(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeAggressive,
		EnableEntropy:  true,
		EnableH2O:      true,
		EnableCompaction: true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.GitLogOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_GitDiff tests git diff compression
func Benchmark_GitDiff(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeMinimal,
		EnableEntropy:  true,
		EnableAST:      true, // AST preservation for code
		EnableH2O:      true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.GitDiffOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_DockerPs tests docker ps compression
func Benchmark_DockerPs(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:          filter.ModeMinimal,
		EnableEntropy: true,
		EnableH2O:     true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.DockerPsOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_KubectlGetPods tests kubectl get pods compression
func Benchmark_KubectlGetPods(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:          filter.ModeMinimal,
		EnableEntropy: true,
		EnableH2O:     true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.KubectlGetPodsOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_NpmInstall tests npm install output compression
func Benchmark_NpmInstall(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeAggressive,
		EnableEntropy:  true,
		EnableH2O:      true,
		EnableCompaction: true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.NpmInstallOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_Pytest tests pytest output compression
func Benchmark_Pytest(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeMinimal,
		EnableEntropy:  true,
		EnableCompaction: true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.PytestOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_GoTest tests go test output compression
func Benchmark_GoTest(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:          filter.ModeMinimal,
		EnableEntropy: true,
		EnableH2O:     true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.GoTestOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		_ = output
	}
}

// Benchmark_LargeLogFile tests large log compression
func Benchmark_LargeLogFile(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeAggressive,
		EnableEntropy:  true,
		EnableH2O:      true,
		EnableCompaction: true,
		EnableTFIDF:    true, // TF-IDF for large inputs
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	input := fixtures.LargeLogFile

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, stats := pipeline.Process(input)
		b.ReportMetric(float64(stats.TotalSaved), "tokens_saved")
		b.ReportMetric(float64(stats.ReductionPercent), "reduction_pct")
		_ = output
	}
}

// Benchmark_AllFixtures runs compression on all fixture types
func Benchmark_AllFixtures(b *testing.B) {
	cfg := filter.PipelineConfig{
		Mode:           filter.ModeMinimal,
		EnableEntropy:  true,
		EnableH2O:      true,
		EnableCompaction: true,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	all := fixtures.AllFixtures()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for name, input := range all {
			output, stats := pipeline.Process(input)
			b.ReportMetric(float64(stats.TotalSaved), name+"_saved")
			_ = output
		}
	}
}
