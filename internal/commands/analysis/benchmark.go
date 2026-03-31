package analysis

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/toml"
)

var (
	benchmarkFilter  string
	benchmarkInput   string
	benchmarkCommand string
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [command...]",
	Short: "Benchmark token compression for a command",
	Long: `Run a command through the compression pipeline and compare
savings across fast/balanced/full presets.

Example:
  tokman benchmark git status
  tokman benchmark docker ps
  tokman benchmark --filter cargo --command "cargo build"`,
	RunE: runBenchmark,
}

func init() {
	benchmarkCmd.Flags().StringVar(&benchmarkFilter, "filter", "", "benchmark a specific TOML filter by name")
	benchmarkCmd.Flags().StringVarP(&benchmarkInput, "input", "i", "", "input text to benchmark (or read from stdin)")
	benchmarkCmd.Flags().StringVar(&benchmarkCommand, "command", "", "command to match TOML filter against")
	registry.Add(func() { registry.Register(benchmarkCmd) })
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	// TOML filter benchmark mode
	if benchmarkFilter != "" || benchmarkCommand != "" {
		return runFilterBenchmark()
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: tokman benchmark <command> [args...]")
	}

	// Execute the command
	commandStr := strings.Join(args, " ")
	fmt.Printf("Running: %s\n\n", commandStr)

	exePath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("command not found: %s", args[0])
	}

	execCmd := exec.Command(exePath, args[1:]...)
	execCmd.Env = os.Environ()
	output, _ := execCmd.CombinedOutput()
	rawOutput := string(output)

	originalTokens := core.EstimateTokens(rawOutput)
	fmt.Printf("Original output: %d chars, ~%d tokens\n\n", len(rawOutput), originalTokens)

	// Benchmark each preset
	presets := []filter.PipelinePreset{
		filter.PresetFast,
		filter.PresetBalanced,
		filter.PresetFull,
	}

	fmt.Printf("%-12s %8s %8s %8s %10s\n", "Preset", "Tokens", "Saved", "Pct", "Duration")
	fmt.Printf("%-12s %8s %8s %8s %10s\n", "------", "------", "-----", "---", "--------")

	for _, preset := range presets {
		cfg := filter.PresetConfig(preset, filter.ModeMinimal)
		pipeline := filter.NewPipelineCoordinator(cfg)

		start := time.Now()
		result, stats := pipeline.Process(rawOutput)
		duration := time.Since(start)

		finalTokens := core.EstimateTokens(result)
		saved := originalTokens - finalTokens
		pct := float64(0)
		if originalTokens > 0 {
			pct = float64(saved) / float64(originalTokens) * 100
		}

		_ = stats // use stats for layer breakdown in verbose mode
		fmt.Printf("%-12s %8d %8d %7.1f%% %10s\n",
			preset, finalTokens, saved, pct, duration.Round(time.Microsecond))
	}

	return nil
}

func runFilterBenchmark() error {
	commandMatch := benchmarkCommand
	if commandMatch == "" {
		commandMatch = benchmarkFilter
	}

	// Load filters
	loader := toml.GetLoader()
	reg, err := loader.LoadAll("")
	if err != nil {
		return fmt.Errorf("failed to load filters: %w", err)
	}

	// Find matching filter
	filename, filterKey, config := reg.FindMatchingFilter(commandMatch)
	if config == nil {
		return fmt.Errorf("no filter matches command %q", commandMatch)
	}

	fmt.Printf("Filter: %s/%s\n", filename, filterKey)

	// Get input
	input := benchmarkInput
	if input == "" {
		buf := make([]byte, 1024*1024)
		n, err := os.Stdin.Read(buf)
		if err != nil && n == 0 {
			return fmt.Errorf("no input provided (use --input or pipe to stdin)")
		}
		input = string(buf[:n])
	}

	originalTokens := core.EstimateTokens(input)
	fmt.Printf("Input: %d chars, ~%d tokens\n\n", len(input), originalTokens)

	// Benchmark TOML filter only
	fmt.Printf("%-20s %8s %8s %8s %10s %10s\n",
		"Method", "Tokens", "Saved", "Pct", "Duration", "Ops/sec")
	fmt.Printf("%-20s %8s %8s %8s %10s %10s\n",
		"------", "------", "-----", "---", "--------", "--------")

	// TOML filter only
	engine := toml.NewTOMLFilterEngine(config)
	iterations := 1000
	start := time.Now()
	var filtered string
	var tokensSaved int
	for i := 0; i < iterations; i++ {
		filtered, tokensSaved = engine.Apply(input, filter.ModeMinimal)
	}
	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)
	opsPerSec := int64(float64(time.Second) / float64(avgDuration))

	filteredTokens := core.EstimateTokens(filtered)
	pct := float64(0)
	if originalTokens > 0 {
		pct = float64(tokensSaved) / float64(originalTokens) * 100
	}

	fmt.Printf("%-20s %8d %8d %7.1f%% %10s %10d\n",
		"TOML filter", filteredTokens, tokensSaved, pct,
		avgDuration.Round(time.Microsecond), opsPerSec)

	// Combined: TOML + pipeline
	fmt.Println("\n--- Combined (TOML + Pipeline) ---")
	presets := []filter.PipelinePreset{
		filter.PresetFast,
		filter.PresetBalanced,
		filter.PresetFull,
	}

	for _, preset := range presets {
		cfg := filter.PresetConfig(preset, filter.ModeMinimal)
		pipeline := filter.NewPipelineCoordinator(cfg)
		wrapper := toml.NewTOMLFilterWrapper("bench", config)
		pipeline.SetTOMLFilter(wrapper, "bench")

		start := time.Now()
		result, stats := pipeline.Process(input)
		duration := time.Since(start)

		finalTokens := core.EstimateTokens(result)
		saved := originalTokens - finalTokens
		pct := float64(0)
		if originalTokens > 0 {
			pct = float64(saved) / float64(originalTokens) * 100
		}

		fmt.Printf("%-20s %8d %8d %7.1f%% %10s\n",
			fmt.Sprintf("TOML+%s", preset), finalTokens, saved, pct,
			duration.Round(time.Microsecond))
		_ = stats
	}

	return nil
}
