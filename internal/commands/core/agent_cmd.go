package core

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage compression agent mode (preset)",
	Long: `View or set the compression agent preset.

Agent modes control which compression layers are active:
  fast     - Quick compression (layers 1-6), ~50-60% reduction, sub-ms
  balanced - Smart selection (layers 1-15), ~70-80% reduction (default)
  deep     - Maximum compression (all layers), ~85-95% reduction
  ultra    - Experimental (layers 1-20 + research), ~90%+ reduction

Examples:
  tokman agent               # Show current agent mode
  tokman agent set fast      # Use fast compression for speed
  tokman agent set balanced  # Balanced (default)
  tokman agent set deep      # Maximum compression`,
	RunE: runAgent,
}

func runAgent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// No args = show current
	if len(args) == 0 {
		current := "balanced"
		if cfg.Pipeline.Preset != "" {
			current = cfg.Pipeline.Preset
		}
		if env := os.Getenv("TOKMAN_PRESET"); env != "" {
			current = env + " (via TOKMAN_PRESET)"
		}

		fmt.Println("Current Agent Mode:", color.New(color.Bold).Sprint(current))
		fmt.Println()
		fmt.Println("Available modes:")
		fmt.Println("  fast    - Quick compression (50-60% savings)")
		fmt.Println("  balanced- Smart selection (70-80% savings, default)")
		fmt.Println("  deep    - Maximum compression (85-95% savings)")
		fmt.Println("  ultra   - Experimental research layers (90%+ savings)")
		fmt.Println()
		fmt.Println("Use: tokman agent set <mode>")
		return nil
	}

	// Subcommands
	switch args[0] {
	case "set", "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: tokman agent set <fast|balanced|deep|ultra>")
		}
		mode := args[1]
		// Validate mode
		switch mode {
		case "fast", "balanced", "deep", "ultra":
		default:
			return fmt.Errorf("invalid mode '%s'. Choose: fast, balanced, deep, ultra", mode)
		}

		// Update config
		cfg.Pipeline.Preset = mode
		cfgPath := config.ConfigPath()
		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Agent mode set to '%s'. Config saved to %s\n", mode, cfgPath)
		return nil

	case "reset":
		// Reset to default (balanced)
		cfg.Pipeline.Preset = ""
		cfgPath := config.ConfigPath()
		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Agent mode reset to default (balanced).")
		return nil

	default:
		return fmt.Errorf("unknown subcommand '%s'. Use 'set' or 'reset'.", args[0])
	}
}

func init() {
	registry.Add(func() { registry.Register(agentCmd) })
}
