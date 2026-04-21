package filtercmd

// watch_cmd.go implements Task #162: --watch mode for continuous compression.
// Usage:
//
//	tok watch <file>
//	tok watch --mode aggressive --interval 2s file.txt
//	tok watch --outdir compressed/ file.txt

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/core"
	"github.com/GrayCodeAI/tok/internal/filter"
)

var watchCmd = &cobra.Command{
	Use:   "watch <file>",
	Short: "Watch a file and re-compress on changes",
	Long: `Poll a file for modification time changes and re-compress it each time.
Compressed output is written to stdout or --outdir.

Examples:
  tok watch input.txt
  tok watch --mode aggressive --interval 2s input.txt
  tok watch --outdir out/ input.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

var (
	watchMode     string
	watchInterval time.Duration
	watchOutDir   string
)

func init() {
	watchCmd.Flags().StringVar(&watchMode, "mode", "minimal", "Compression mode: none|minimal|aggressive")
	watchCmd.Flags().DurationVar(&watchInterval, "interval", time.Second, "Poll interval (e.g. 1s, 500ms)")
	watchCmd.Flags().StringVar(&watchOutDir, "outdir", "", "Output directory (default: stdout)")
	registry.Add(func() { registry.Register(watchCmd) })
}

func runWatch(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	mode := filter.ModeMinimal
	switch strings.ToLower(watchMode) {
	case "none":
		mode = filter.ModeNone
	case "aggressive":
		mode = filter.ModeAggressive
	}

	if watchOutDir != "" {
		if err := os.MkdirAll(watchOutDir, 0750); err != nil {
			return fmt.Errorf("watch: mkdir %s: %w", watchOutDir, err)
		}
	}

	// Get initial mtime so we don't compress on startup.
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("watch: stat %s: %w", filePath, err)
	}
	lastMod := info.ModTime()

	// Set up signal handling for graceful exit.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	var totalEvents int
	var totalOrigTokens, totalFinalTokens int

	out.Global().Errorf("[tok watch] Watching %s (interval: %s)\n", filePath, watchInterval)

	for {
		select {
		case <-sigCh:
			// Print summary and exit.
			out.Global().Errorf("\n[tok watch] Stopped. Summary:\n")
			out.Global().Errorf("  Events:       %d\n", totalEvents)
			out.Global().Errorf("  Total input:  %d tokens\n", totalOrigTokens)
			out.Global().Errorf("  Total output: %d tokens\n", totalFinalTokens)
			if totalOrigTokens > 0 {
				saved := totalOrigTokens - totalFinalTokens
				pct := float64(saved) / float64(totalOrigTokens) * 100
				out.Global().Errorf("  Saved:        %s tokens (%.1f%%)\n",
					formatSaved(saved), pct)
			}
			return nil

		case <-ticker.C:
			info, err := os.Stat(filePath)
			if err != nil {
				out.Global().Errorf("[tok watch] stat error: %v\n", err)
				continue
			}

			if !info.ModTime().After(lastMod) {
				continue
			}
			lastMod = info.ModTime()

			out.Global().Errorf("Recompressing %s...\n", filePath)

			raw, err := os.ReadFile(filePath)
			if err != nil {
				out.Global().Errorf("[tok watch] read error: %v\n", err)
				continue
			}

			input := string(raw)
			if input == "" {
				continue
			}

			pipeline := filter.NewPipelineCoordinator(filter.PipelineConfig{
				Mode:             mode,
				NgramEnabled:     true,
				EnableCompaction: true,
			})
			result, stats := pipeline.Process(input)

			origTokens := core.EstimateTokens(input)
			totalEvents++
			totalOrigTokens += origTokens
			totalFinalTokens += stats.FinalTokens

			if watchOutDir != "" {
				outPath := filepath.Join(watchOutDir, filepath.Base(filePath))
				// #nosec G703 -- destination is constrained to watchOutDir + basename(filePath).
				if err := os.WriteFile(outPath, []byte(result), 0600); err != nil {
					out.Global().Errorf("[tok watch] write error: %v\n", err)
				}
			} else {
				out.Global().Print(result)
			}
		}
	}
}
