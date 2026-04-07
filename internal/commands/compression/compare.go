package compression

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/compression"
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
	Example: `  tokman compression-compare file.txt
  cat file.txt | tokman compression-compare`,
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

	fmt.Println("Running compression comparison...")
	fmt.Printf("Input size: %s\n\n", formatBytes(len(data)))

	result, err := compression.CompareAlgorithms(data)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	fmt.Print(result.PrintComparison())
	return nil
}
