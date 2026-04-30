package filtercmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/core"
	"github.com/GrayCodeAI/tok/internal/filter"
)

// compressCmd implements Task #187: compression pipeline as a Unix pipe filter.
// Usage:
//
//	echo "some text" | tok compress
//	cat big-output.txt | tok compress --mode aggressive
//	tok compress --file input.txt --stats
var compressCmd = &cobra.Command{
	Use:   "compress [file]",
	Short: "Compress text via stdin or file (Unix pipe filter)",
	Long: `Read text from stdin (or a file) and write compressed output to stdout.
Designed for use in Unix pipelines:

  cat output.txt | tok compress | llm-tool
  tok compress --mode aggressive < big_file.txt`,
	RunE: runCompress,
}

var (
	compressMode   string
	compressStats  bool
	compressFile   string
	compressBudget int
)

func init() {
	compressCmd.Flags().StringVar(&compressMode, "mode", "minimal", "Compression mode: none|minimal|aggressive")
	compressCmd.Flags().BoolVar(&compressStats, "stats", false, "Print compression stats to stderr")
	compressCmd.Flags().StringVar(&compressFile, "file", "", "Input file (default: stdin)")
	compressCmd.Flags().IntVar(&compressBudget, "budget", 0, "Target token budget (0 = no limit)")
	registry.Add(func() { registry.Register(compressCmd) })
}

func runCompress(cmd *cobra.Command, args []string) error {
	// Determine input source: positional arg > --file > stdin
	inputFile := compressFile
	if len(args) > 0 {
		inputFile = args[0]
	}

	var raw []byte
	var err error
	if inputFile != "" {
		raw, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("compress: read %s: %w", inputFile, err)
		}
	} else {
		raw, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("compress: read stdin: %w", err)
		}
	}

	input := string(raw)
	if input == "" {
		return nil
	}

	mode := filter.ModeMinimal
	switch strings.ToLower(compressMode) {
	case "none":
		mode = filter.ModeNone
	case "aggressive":
		mode = filter.ModeAggressive
	}

	budget := compressBudget
	if budget == 0 {
		budget = shared.GetTokenBudget()
	}

	pipeline := filter.NewPipelineCoordinator(filter.PipelineConfig{
		Mode:                mode,
		QueryIntent:         shared.GetQueryIntent(),
		Budget:              budget,
		SessionTracking:     false,
		NgramEnabled:        true,
		EnableCompaction:    true,
		EnableAttribution:   true,
		EnableH2O:           true,
		EnableAttentionSink: true,
	})

	result, stats, err := pipeline.Process(input)
	if err != nil {
		return err
	}

	// Write compressed output to stdout
	out.Global().Print(result)

	// Optionally print stats to stderr
	if compressStats {
		origTokens := core.EstimateTokens(input)
		pct := stats.ReductionPercent
		out.Global().Errorf("[tok] %d → %d tokens (%.1f%% reduction, saved %s)\n",
			origTokens,
			stats.FinalTokens,
			pct,
			formatSaved(origTokens-stats.FinalTokens),
		)
	}

	return nil
}

// formatSaved returns a human-readable token count.
func formatSaved(n int) string {
	if n < 0 {
		return "0"
	}
	if n >= 1000 {
		return strconv.FormatFloat(float64(n)/1000, 'f', 1, 64) + "K"
	}
	return strconv.Itoa(n)
}
