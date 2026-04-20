package configcmd

import (
	"fmt"
	"os"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check config files for errors",
	Long: `Validate tok configuration files for syntax errors,
invalid values, and deprecated options.`,
	RunE: runConfigValidate,
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	out.Global().Println("Validating tok configuration...")
	out.Global().Println()

	hasErrors := false

	configPaths := []string{
		shared.CfgFile,
	}
	if shared.CfgFile == "" {
		configPaths = []string{effectiveConfigPath()}
	}

	for _, path := range configPaths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			out.Global().Printf("  ⚠ %s: not found (using defaults)\n", path)
			continue
		}

		cfg, err := config.LoadFromFile(path)
		if err != nil {
			out.Global().Printf("  ✗ %s: %v\n", path, err)
			hasErrors = true
			continue
		}

		if cfg.Pipeline.MaxContextTokens < 0 {
			out.Global().Printf("  ✗ %s: max_context_tokens cannot be negative\n", path)
			hasErrors = true
		}
		if cfg.Pipeline.EntropyThreshold < 0 || cfg.Pipeline.EntropyThreshold > 1 {
			out.Global().Printf("  ✗ %s: entropy_threshold must be 0.0-1.0\n", path)
			hasErrors = true
		}
		if cfg.Pipeline.PerplexityThreshold < 0 || cfg.Pipeline.PerplexityThreshold > 1 {
			out.Global().Printf("  ✗ %s: perplexity_threshold must be 0.0-1.0\n", path)
			hasErrors = true
		}
		if cfg.Pipeline.H2OSinkSize < 0 {
			out.Global().Printf("  ✗ %s: h2o_sink_size cannot be negative\n", path)
			hasErrors = true
		}
		if cfg.Pipeline.CacheMaxSize < 0 {
			out.Global().Printf("  ✗ %s: cache_max_size cannot be negative\n", path)
			hasErrors = true
		}

		out.Global().Printf("  ✓ %s: valid\n", path)
	}

	defaults := config.Defaults()
	if defaults.Pipeline.MaxContextTokens > 0 {
		out.Global().Printf("  ✓ defaults: max_context=%d tokens\n", defaults.Pipeline.MaxContextTokens)
	}

	out.Global().Println()
	if hasErrors {
		return fmt.Errorf("configuration has errors")
	}
	out.Global().Println("All configuration checks passed!")
	return nil
}
