package compression

import (
	"fmt"
	"os"
	"path/filepath"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/compression"
)

var (
	brotliLevel      int
	brotliDecompress bool
	brotliOutput     string
)

func init() {
	registry.Add(func() {
		registry.Register(brotliCmd)
	})
	brotliCmd.Flags().IntVarP(&brotliLevel, "level", "l", 4, "Compression level (0-11)")
	brotliCmd.Flags().BoolVarP(&brotliDecompress, "decompress", "d", false, "Decompress mode")
	brotliCmd.Flags().StringVarP(&brotliOutput, "output", "o", "", "Output file")
}

var brotliCmd = &cobra.Command{
	Use:   "brotli [file]",
	Short: "Compress/decompress using Brotli algorithm",
	Long: `Compress or decompress files using Google's Brotli algorithm.

Brotli provides 2-4x better compression than gzip for text content
and up to 82x for repetitive content like logs.

Quality levels:
  0  = No compression (fastest)
  1-3 = Fast compression
  4-5 = Balanced (default)
  6-8 = Good compression
  9-11 = Maximum compression (slowest)

Examples:
  tok brotli file.txt                    # Compress file.txt to file.txt.br
  tok brotli file.txt -o output.br       # Compress to specific output
  tok brotli file.txt -l 11              # Use maximum compression
  tok brotli file.txt.br -d              # Decompress`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrotli,
}

func runBrotli(cmd *cobra.Command, args []string) error {
	// Validate level
	if brotliLevel < 0 || brotliLevel > 11 {
		return fmt.Errorf("compression level must be between 0 and 11")
	}

	// Read input
	var input []byte
	var err error

	if len(args) == 0 {
		// Read from stdin
		input, err = readStdin()
		if err != nil {
			return err
		}
	} else {
		// Read from file
		input, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Create compressor
	cfg := compression.BrotliConfig{
		Quality: brotliLevel,
		LGWin:   22,
	}
	compressor := compression.NewBrotliCompressorWithConfig(cfg)

	var output []byte
	var result *compression.CompressionResult

	if brotliDecompress {
		// Decompress
		output, err = compressor.Decompress(input)
		if err != nil {
			return fmt.Errorf("decompression failed: %w", err)
		}

		// Determine output filename
		if brotliOutput == "" && len(args) > 0 {
			brotliOutput = args[0]
			if len(brotliOutput) > 3 && brotliOutput[len(brotliOutput)-3:] == ".br" {
				brotliOutput = brotliOutput[:len(brotliOutput)-3]
			}
		}
	} else {
		// Compress
		result, err = compressor.CompressWithMetadata(input)
		if err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}
		output = result.Data

		// Determine output filename
		if brotliOutput == "" && len(args) > 0 {
			brotliOutput = args[0] + ".br"
		}
	}

	// Write output
	if brotliOutput == "" {
		// Write to stdout
		_, err = os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		// Write to file
		brotliOutput = filepath.Clean(brotliOutput)
		// #nosec G703 -- output path is an explicit CLI destination selected by the user.
		if err := os.WriteFile(brotliOutput, output, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	// Print stats
	if !brotliDecompress && result != nil {
		out.Global().Errorf("\n%s\n", color.New(color.Bold).Sprint("Compression Results"))
		out.Global().Errorf("  Algorithm:    %s\n", result.Algorithm)
		out.Global().Errorf("  Quality:      %d (%s)\n", brotliLevel, compression.GetQualityName(brotliLevel))
		out.Global().Errorf("  Original:     %s\n", formatBytes(result.OriginalSize))
		out.Global().Errorf("  Compressed:   %s\n", formatBytes(result.CompressedSize))
		out.Global().Errorf("  Ratio:        %.2f%%\n", result.Percentage())
		out.Global().Errorf("  Space Saved:  %s\n", formatBytes(result.SpaceSaved))
	}

	return nil
}

func readStdin() ([]byte, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if info.Mode()&os.ModeNamedPipe == 0 {
		return nil, fmt.Errorf("no input provided (use pipe or specify a file)")
	}

	return os.ReadFile("/dev/stdin")
}

func formatBytes(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
