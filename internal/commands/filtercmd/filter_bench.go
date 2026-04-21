package filtercmd

import (
	"fmt"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/toml"
)

var benchmarkIterations int
var benchmarkInput string

var filterBenchCmd = &cobra.Command{
	Use:   "filter-bench [command]",
	Short: "Benchmark TOML filter performance",
	Long: `Run performance benchmarks on TOML filters.

Benchmarks measure:
- Filter matching time
- Filter application time
- Token savings ratio
- Throughput (chars/sec)

Examples:
  tok filter-bench "git status"      # Benchmark git status filter
  tok filter-bench "cargo build"     # Benchmark cargo build filter
  tok filter-bench -n 1000 "git log" # Run 1000 iterations`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFilterBench,
}

func init() {
	filterBenchCmd.Flags().IntVarP(&benchmarkIterations, "iterations", "n", 100, "number of benchmark iterations")
	filterBenchCmd.Flags().StringVarP(&benchmarkInput, "input", "i", "", "custom input text for benchmarking")
	registry.Add(func() { registry.Register(filterBenchCmd) })
}

func runFilterBench(cmd *cobra.Command, args []string) error {
	loader := toml.GetLoader()
	reg, err := loader.LoadAll("")
	if err != nil {
		return fmt.Errorf("failed to load filters: %w", err)
	}

	if reg.Count() == 0 {
		return fmt.Errorf("no filters loaded")
	}

	input := benchmarkInput
	if input == "" {
		input = generateBenchmarkInput()
	}

	out.Global().Printf("Benchmarking filters with %d iterations\n", benchmarkIterations)
	out.Global().Printf("Input size: %d chars (~%d tokens)\n\n", len(input), len(input)/4)

	if len(args) > 0 {
		command := args[0]
		return benchmarkCommand(reg, command, input, benchmarkIterations)
	}

	// Benchmark common commands
	commands := []string{
		"git status",
		"git log",
		"git diff",
		"docker ps",
		"docker images",
		"cargo build",
		"npm install",
		"pip install",
	}

	out.Global().Println("| Command              | Match (ns) | Apply (ns) | Savings % | Throughput (MB/s) |")
	out.Global().Println("|----------------------|------------|------------|-----------|-------------------|")

	for _, c := range commands {
		benchmarkCommandRow(reg, c, input, benchmarkIterations)
	}

	return nil
}

func benchmarkCommand(reg *toml.FilterRegistry, command, input string, iterations int) error {
	out.Global().Printf("Benchmarking filter for: %s\n\n", command)

	_, filterName, config := reg.FindMatchingFilter(command)
	if config == nil {
		return fmt.Errorf("no filter matches command %q", command)
	}

	out.Global().Printf("Filter: %s\n", filterName)

	var totalMatch, totalApply time.Duration
	var totalSaved int

	for i := 0; i < iterations; i++ {
		matchStart := time.Now()
		_, _, _ = reg.FindMatchingFilter(command)
		totalMatch += time.Since(matchStart)

		applyStart := time.Now()
		filtered, saved := toml.ApplyTOMLFilter(input, config)
		totalApply += time.Since(applyStart)
		totalSaved += saved
		_ = filtered
	}

	avgMatch := totalMatch / time.Duration(iterations)
	avgApply := totalApply / time.Duration(iterations)
	avgSaved := totalSaved / iterations
	savingsPct := float64(avgSaved) / float64(len(input)) * 100
	throughput := float64(len(input)*iterations) / totalApply.Seconds() / 1024 / 1024

	out.Global().Printf("\nResults (%d iterations):\n", iterations)
	out.Global().Printf("  Match time:    %v\n", avgMatch)
	out.Global().Printf("  Apply time:    %v\n", avgApply)
	out.Global().Printf("  Token savings: %.1f%%\n", savingsPct)
	out.Global().Printf("  Throughput:    %.2f MB/s\n", throughput)

	return nil
}

func benchmarkCommandRow(reg *toml.FilterRegistry, command, input string, iterations int) {
	_, _, config := reg.FindMatchingFilter(command)
	if config == nil {
		out.Global().Printf("| %-20s | %10s | %10s | %9s | %17s |\n",
			shared.Truncate(command, 20), "-", "-", "no match", "-")
		return
	}

	var totalMatch, totalApply time.Duration
	var totalSaved int

	for i := 0; i < iterations; i++ {
		matchStart := time.Now()
		_, _, _ = reg.FindMatchingFilter(command)
		totalMatch += time.Since(matchStart)

		applyStart := time.Now()
		filtered, saved := toml.ApplyTOMLFilter(input, config)
		totalApply += time.Since(applyStart)
		totalSaved += saved
		_ = filtered
	}

	avgMatch := totalMatch / time.Duration(iterations)
	avgApply := totalApply / time.Duration(iterations)
	avgSaved := totalSaved / iterations
	savingsPct := float64(avgSaved) / float64(len(input)) * 100
	throughput := float64(len(input)*iterations) / totalApply.Seconds() / 1024 / 1024

	out.Global().Printf("| %-20s | %10d | %10d | %8.1f%% | %17.2f |\n",
		shared.Truncate(command, 20), avgMatch.Nanoseconds(), avgApply.Nanoseconds(), savingsPct, throughput)
}

func generateBenchmarkInput() string {
	return `On branch main
Changes to be committed:
  (use "git restore --file <file>..." to unstage)
	new file:   src/main.go
	modified:   README.md
	deleted:    old_file.txt

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
	modified:   internal/config/config.go
	modified:   internal/filter/pipeline.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	new_feature.go
	docs/guide.md

no changes added to commit (use "git add" and/or "git commit -a")
`
}
