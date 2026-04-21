package filtercmd

import (
	"fmt"
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/filter"
)

// benchmarkCmd runs comprehensive compression benchmarks.
var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [file...]",
	Short: "Run compression benchmarks",
	Long: `Run comprehensive multi-dimensional benchmarks on input files or built-in test suite.

Examples:
  tok filter benchmark                    # Run built-in test suite
  tok filter benchmark *.go               # Benchmark specific files
  tok filter benchmark --json             # Output results as JSON
  tok filter benchmark --mode=aggressive  # Use aggressive compression`,
	RunE: runBenchmark,
}

var (
	benchmarkJSON  bool
	benchmarkMode  string
	benchmarkQuiet bool
)

func init() {
	benchmarkCmd.Flags().BoolVar(&benchmarkJSON, "json", false, "Output results as JSON")
	benchmarkCmd.Flags().StringVar(&benchmarkMode, "mode", "minimal", "Compression mode: minimal|aggressive")
	benchmarkCmd.Flags().BoolVarP(&benchmarkQuiet, "quiet", "q", false, "Only show summary")
	registry.Add(func() { registry.Register(benchmarkCmd) })
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	mode := filter.ModeMinimal
	if benchmarkMode == "aggressive" {
		mode = filter.ModeAggressive
	}

	cfg := filter.PipelineConfig{
		Mode:             mode,
		SessionTracking:  true,
		EnableEntropy:    true,
		EnablePerplexity: true,
	}

	bench := filter.NewCrunchBench()

	// Use built-in test inputs if no files provided
	var inputs []filter.TestInput
	if len(args) == 0 {
		inputs = filter.GetBuiltinTestInputs()
	} else {
		// Read provided files
		for _, path := range args {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", path, err)
			}
			inputs = append(inputs, filter.TestInput{
				Name:        path,
				Content:     string(content),
				ContentType: detectContentType(path),
				ExpectedMin: 10,
				ExpectedMax: 80,
			})
		}
	}

	for _, input := range inputs {
		bench.RegisterTestInput(input.Name, input.Content, input.ContentType, input.ExpectedMin, input.ExpectedMax)
	}

	report := bench.RunBenchmark(cfg)

	if benchmarkJSON {
		// JSON output
		out.Global().Println("{")
		out.Global().Printf("  \"timestamp\": \"%s\",\n", report.Timestamp.Format("2006-01-02T15:04:05Z"))
		out.Global().Printf("  \"total_tests\": %d,\n", report.TotalTests)
		out.Global().Printf("  \"passed\": %d,\n", report.Passed)
		out.Global().Printf("  \"failed\": %d,\n", report.Failed)
		out.Global().Printf("  \"avg_compression\": %.2f,\n", report.OverallStats.AvgCompression)
		out.Global().Printf("  \"avg_latency\": %.2f,\n", report.OverallStats.AvgLatency)
		out.Global().Printf("  \"avg_quality\": %.2f\n", report.OverallStats.AvgQuality)
		out.Global().Println("}")
	} else {
		// Formatted output
		out.Global().Print(bench.FormatReport(report))
	}

	return nil
}

func detectContentType(path string) string {
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext = path[i+1:]
			break
		}
	}

	switch ext {
	case "go", "py", "js", "ts", "rs", "java", "cpp", "c":
		return "code"
	case "json":
		return "json"
	case "log":
		return "log"
	case "md", "txt":
		return "text"
	case "diff":
		return "diff"
	default:
		return "unknown"
	}
}
