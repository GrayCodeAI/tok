package compression

import (
	"fmt"
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/compression"
)

func init() {
	registry.Add(func() {
		registry.Register(compareCmd)
	})
}

var compareCmd = &cobra.Command{
	Use:   "compression-compare [file]",
	Short: "Compare compression algorithms",
	Long: `Compare different compression algorithms (Brotli, Gzip) at various levels.

Shows compression ratio, speed, and space saved for each algorithm.`,
	Example: `  tok compression-compare file.txt
  cat file.txt | tok compression-compare`,
	RunE: runCompare,
}

func runCompare(cmd *cobra.Command, args []string) error {
	var data []byte
	var err error

	if len(args) == 0 {
		data, err = readStdin()
		if err != nil {
			return err
		}
	} else {
		data, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	out.Global().Println("Running compression comparison...")
	out.Global().Printf("Input size: %s\n\n", formatBytes(len(data)))

	result, err := compression.CompareAlgorithms(data)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	out.Global().Print(result.PrintComparison())
	return nil
}
