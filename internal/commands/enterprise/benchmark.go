package enterprise

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/GrayCodeAI/tokman/internal/benchmarking"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/spf13/cobra"
)

func init() {
	benchmarkCmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Run performance benchmarks",
		Long:  "Run comprehensive performance benchmarks for TokMan pipelines and filters",
	}

	var format string
	var output string

	runCmd := &cobra.Command{
		Use:   "run [suite]",
		Short: "Run a benchmark suite",
		RunE: func(cmd *cobra.Command, args []string) error {
			suiteName := "standard"
			if len(args) > 0 {
				suiteName = args[0]
			}

			runner := benchmarking.NewRunner()

			// Register standard suite
			if suiteName == "standard" {
				runner.RegisterSuite(benchmarking.StandardBenchmarks())
			}

			ctx := context.Background()
			report, err := runner.RunSuite(ctx, suiteName)
			if err != nil {
				return err
			}

			// Determine output destination
			var out *os.File
			if output == "" || output == "-" {
				out = os.Stdout
			} else {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				out = f
			}

			// Export in requested format
			if err := benchmarking.ExportReport(report, format, out); err != nil {
				return fmt.Errorf("failed to export report: %w", err)
			}

			return nil
		},
	}

	runCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (json, csv, table)")
	runCmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")

	benchmarkCmd.AddCommand(runCmd)

	benchmarkCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available benchmark suites",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available benchmark suites:")
			fmt.Println("  standard      - Standard performance benchmarks")
			fmt.Println("  compression   - Compression pipeline benchmarks")
			fmt.Println("  memory        - Memory usage benchmarks")
			fmt.Println("  concurrency   - Concurrent operation benchmarks")
		},
	})

	// Add comparison command
	compareCmd := &cobra.Command{
		Use:   "compare [baseline] [current]",
		Short: "Compare two benchmark results",
		Long:  "Compare benchmark results to detect performance regressions or improvements",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baselineFile := args[0]
			currentFile := args[1]

			// Read baseline report
			baselineData, err := os.ReadFile(baselineFile)
			if err != nil {
				return fmt.Errorf("failed to read baseline: %w", err)
			}

			// Read current report
			currentData, err := os.ReadFile(currentFile)
			if err != nil {
				return fmt.Errorf("failed to read current: %w", err)
			}

			// Compare reports
			comparison, err := benchmarking.CompareReports(baselineData, currentData)
			if err != nil {
				return fmt.Errorf("failed to compare: %w", err)
			}

			fmt.Println(comparison)
			return nil
		},
	}

	benchmarkCmd.AddCommand(compareCmd)

	// Add chart command
	chartCmd := &cobra.Command{
		Use:   "chart [report-file]",
		Short: "Generate charts from benchmark report",
		Long:  "Generate ASCII charts from benchmark results for visualization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportFile := args[0]

			// Read report
			data, err := os.ReadFile(reportFile)
			if err != nil {
				return fmt.Errorf("failed to read report: %w", err)
			}

			// Parse report
			var report benchmarking.SuiteReport
			if err := json.Unmarshal(data, &report); err != nil {
				return fmt.Errorf("failed to parse report: %w", err)
			}

			// Generate charts
			charts := benchmarking.GenerateReportCharts(&report)
			fmt.Println(charts)

			return nil
		},
	}

	benchmarkCmd.AddCommand(chartCmd)

	registry.Add(func() { registry.Register(benchmarkCmd) })
}
