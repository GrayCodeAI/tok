package core

import (
	"fmt"
	"os"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
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
  tok agent               # Show current agent mode
  tok agent set fast      # Use fast compression for speed
  tok agent set balanced  # Balanced (default)
  tok agent set deep      # Maximum compression`,
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
		if env := os.Getenv("TOK_PRESET"); env != "" {
			current = env + " (via TOK_PRESET)"
		}

		out.Global().Println("Current Agent Mode:", color.New(color.Bold).Sprint(current))
		out.Global().Println()
		out.Global().Println("Available modes:")
		out.Global().Println("  fast    - Quick compression (50-60% savings)")
		out.Global().Println("  balanced- Smart selection (70-80% savings, default)")
		out.Global().Println("  deep    - Maximum compression (85-95% savings)")
		out.Global().Println("  ultra   - Experimental research layers (90%+ savings)")
		out.Global().Println()
		out.Global().Println("Use: tok agent set <mode>")
		return nil
	}

	// Subcommands
	switch args[0] {
	case "set", "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: tok agent set <fast|balanced|deep|ultra>")
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
		out.Global().Printf("Agent mode set to '%s'. Config saved to %s\n", mode, cfgPath)
		return nil

	case "reset":
		// Reset to default (balanced)
		cfg.Pipeline.Preset = ""
		cfgPath := config.ConfigPath()
		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		out.Global().Println("Agent mode reset to default (balanced).")
		return nil

	default:
		return fmt.Errorf("unknown subcommand %q — use 'set' or 'reset'", args[0])
	}
}

func init() {
	registry.Add(func() { registry.Register(agentCmd) })
}
