package filtercmd

import (
	"fmt"
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/toml"
)

var filterTestInput string
var filterTestCommand string

var filterTestCmd = &cobra.Command{
	Use:   "filter-test <filter-name>",
	Short: "Test a TOML filter against sample input",
	Long: `Apply a TOML filter to sample input and show the result.

This command loads a TOML filter and applies it to the provided input text,
showing both the original and filtered output with token savings.

Examples:
  echo "git status output" | tok filter-test git --command "git status"
  tok filter-test cargo --input "$(cat cargo_output.txt)" --command "cargo build"`,
	Args: cobra.ExactArgs(1),
	RunE: runFilterTest,
}

func init() {
	filterTestCmd.Flags().StringVarP(&filterTestInput, "input", "i", "", "input text to test (or read from stdin)")
	filterTestCmd.Flags().StringVarP(&filterTestCommand, "command", "c", "", "command to match filter against (required)")
	registry.Add(func() { registry.Register(filterTestCmd) })
}

func runFilterTest(cmd *cobra.Command, args []string) error {
	_ = args[0] // filter name for future use

	if filterTestCommand == "" {
		return fmt.Errorf("--command flag is required to match filter")
	}

	// Load filters using the global loader
	loader := toml.GetLoader()
	registry, err := loader.LoadAll("")
	if err != nil {
		return fmt.Errorf("failed to load filters: %w", err)
	}

	// Find matching filter
	filename, filterKey, config := registry.FindMatchingFilter(filterTestCommand)
	if config == nil {
		// Try to find by filter name directly
		out.Global().Errorf("No filter matches command %q\n", filterTestCommand)
		out.Global().Errorf("Available filters: %d\n", registry.Count())
		return fmt.Errorf("filter not found for command")
	}

	out.Global().Printf("Filter: %s/%s\n", filename, filterKey)

	// Get input
	input := filterTestInput
	if input == "" {
		buf := make([]byte, 1024*1024)
		n, err := os.Stdin.Read(buf)
		if err != nil && n == 0 {
			return fmt.Errorf("no input provided (use --input or pipe to stdin)")
		}
		input = string(buf[:n])
	}

	// Apply filter
	filtered, tokensSaved := toml.ApplyTOMLFilter(input, config)

	// Show results
	originalTokens := len(input) / 4
	filteredTokens := len(filtered) / 4

	out.Global().Printf("\n=== Original (%d chars, ~%d tokens) ===\n", len(input), originalTokens)
	if len(input) > 500 {
		out.Global().Printf("%s...\n", input[:500])
	} else {
		out.Global().Println(input)
	}

	out.Global().Printf("\n=== Filtered (%d chars, ~%d tokens) ===\n", len(filtered), filteredTokens)
	if len(filtered) > 500 {
		out.Global().Printf("%s...\n", filtered[:500])
	} else {
		out.Global().Println(filtered)
	}

	// Calculate savings
	savingsPct := 0.0
	if originalTokens > 0 {
		savingsPct = float64(tokensSaved) / float64(originalTokens) * 100
	}

	out.Global().Printf("\n=== Summary ===\n")
	out.Global().Printf("Tokens saved: %d (%.1f%%)\n", tokensSaved, savingsPct)
	out.Global().Printf("Compression ratio: %.2fx\n", float64(originalTokens)/float64(max(filteredTokens, 1)))

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
