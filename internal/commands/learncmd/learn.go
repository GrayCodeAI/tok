// Package learncmd provides CLI commands for learning mode.
package learncmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/learn"
)

var (
	headerColor = color.New(color.FgCyan, color.Bold)
	statColor   = color.New(color.FgMagenta)
	warnColor   = color.New(color.FgYellow)
	okColor     = color.New(color.FgGreen)
	dimColor    = color.New(color.Faint)
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Automatic noise pattern discovery and filter generation",
	Long: `Learning mode monitors command outputs to discover repeated noise 
patterns and automatically generate filter suggestions.

By default, learning mode only shows its status. Use explicit flags 
to start learning, view discoveries, or apply filters.`,
	Example: `  tokman learn --status          # Show learning status
  tokman learn start             # Start collecting patterns
  tokman learn stop              # Stop collecting
  tokman learn show              # Show discovered patterns
  tokman learn apply             # Generate and apply filters
  tokman learn clear             # Clear all learned data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showStatus, _ := cmd.Flags().GetBool("status")
		if showStatus || len(args) == 0 {
			return showLearnStatus()
		}
		return cmd.Help()
	},
}

var learnStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start learning mode (collect patterns)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := learn.DefaultConfig()
		cfg.Enabled = true

		learner, err := learn.New(cfg)
		if err != nil {
			return fmt.Errorf("start learning: %w", err)
		}
		if learner != nil {
			learner.Start()
			learner.Close()
		}

		okColor.Println("✓ Learning mode started")
		fmt.Println("TokMan will now collect command output patterns.")
		fmt.Println("Use 'tokman learn show' to see discoveries.")
		fmt.Println("Use 'tokman learn stop' to stop collecting.")
		return nil
	},
}

var learnStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop learning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := learn.DefaultConfig()
		cfg.Enabled = true

		learner, err := learn.New(cfg)
		if err != nil {
			return fmt.Errorf("stop learning: %w", err)
		}
		if learner != nil {
			learner.Stop()
			learner.Close()
		}

		okColor.Println("✓ Learning mode stopped")
		return nil
	},
}

var learnShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show discovered patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("filter")

		cfg := learn.DefaultConfig()
		cfg.Enabled = true

		learner, err := learn.New(cfg)
		if err != nil {
			return fmt.Errorf("show patterns: %w", err)
		}
		if learner == nil {
			fmt.Println("Learning mode is not enabled.")
			return nil
		}
		defer learner.Close()

		patterns, err := learner.GetPatterns(status)
		if err != nil {
			return err
		}

		if len(patterns) == 0 {
			fmt.Println("No patterns discovered yet.")
			fmt.Println("Start learning with: tokman learn start")
			return nil
		}

		headerColor.Println("Discovered Patterns")
		fmt.Println(strings.Repeat("─", 80))
		fmt.Printf("%-4s  %-30s  %-10s  %5s  %6s  %s\n",
			"ID", "Pattern", "Command", "Freq", "Conf", "Category")
		fmt.Println(strings.Repeat("─", 80))

		for _, p := range patterns {
			patStr := p.Pattern
			if len(patStr) > 30 {
				patStr = patStr[:27] + "..."
			}

			confColor := dimColor
			if p.Confidence >= 0.8 {
				confColor = okColor
			} else if p.Confidence >= 0.5 {
				confColor = warnColor
			}

			fmt.Printf("%-4d  %-30s  %-10s  %5d  ", p.ID, patStr, p.Command, p.Frequency)
			confColor.Printf("%5.0f%%", p.Confidence*100)
			fmt.Printf("  %s\n", p.Category)
		}

		fmt.Println(strings.Repeat("─", 80))
		dimColor.Printf("%d patterns found\n", len(patterns))

		return nil
	},
}

var learnApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Generate filters from discovered patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfg := learn.DefaultConfig()
		cfg.Enabled = true

		learner, err := learn.New(cfg)
		if err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		if learner == nil {
			fmt.Println("Learning mode is not enabled.")
			return nil
		}
		defer learner.Close()

		suggestions, err := learner.GenerateFilters()
		if err != nil {
			return err
		}

		if len(suggestions) == 0 {
			fmt.Println("No filter suggestions yet.")
			fmt.Println("Collect more samples to discover patterns.")
			return nil
		}

		headerColor.Println("Filter Suggestions")
		fmt.Println(strings.Repeat("─", 60))

		for i, s := range suggestions {
			fmt.Printf("\n%d. %s (confidence: %.0f%%)\n", i+1, s.Command, s.Confidence*100)
			fmt.Printf("   Patterns: %d\n", len(s.Patterns))
			fmt.Println()
			fmt.Println("   Generated TOML:")
			fmt.Println("   " + strings.ReplaceAll(s.TOMLOutput, "\n", "\n   "))
		}

		if dryRun {
			warnColor.Println("\n[dry-run] No filters were applied.")
			fmt.Println("Remove --dry-run to apply these filters.")
		} else {
			fmt.Println()
			okColor.Printf("Generated %d filter suggestions.\n", len(suggestions))
			fmt.Println("Copy the TOML above to ~/.config/tokman/filters/ to apply.")
		}

		return nil
	},
}

var learnClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all learned data",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := learn.DefaultConfig()
		cfg.Enabled = true

		learner, err := learn.New(cfg)
		if err != nil {
			return fmt.Errorf("clear: %w", err)
		}
		if learner == nil {
			fmt.Println("Learning mode is not enabled.")
			return nil
		}
		defer learner.Close()

		if err := learner.Clear(); err != nil {
			return fmt.Errorf("clear failed: %w", err)
		}

		okColor.Println("✓ All learned data cleared")
		return nil
	},
}

func showLearnStatus() error {
	cfg := learn.DefaultConfig()
	cfg.Enabled = true

	learner, err := learn.New(cfg)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if learner == nil {
		fmt.Println("Learning mode is not enabled.")
		fmt.Println("Enable with: tokman learn start")
		return nil
	}
	defer learner.Close()

	stats, err := learner.GetStats()
	if err != nil {
		return err
	}

	headerColor.Println("Learning Mode Status")
	fmt.Println(strings.Repeat("─", 40))

	if stats.LearningActive {
		okColor.Println("Status: Active ✓")
	} else {
		warnColor.Println("Status: Inactive")
	}

	fmt.Printf("Samples collected:  %d\n", stats.SamplesCollected)
	fmt.Printf("Patterns found:     %d\n", stats.PatternsFound)
	statColor.Printf("  Pending:          %d\n", stats.PendingPatterns)
	fmt.Printf("  Approved:         %d\n", stats.ApprovedPatterns)
	fmt.Printf("  Rejected:         %d\n", stats.RejectedPatterns)
	fmt.Printf("Avg confidence:     %.0f%%\n", stats.AvgConfidence*100)
	fmt.Printf("Commands covered:   %d\n", stats.CommandsCovered)

	fmt.Println(strings.Repeat("─", 40))

	if stats.PendingPatterns > 0 {
		fmt.Printf("\n💡 %d patterns ready for review.\n", stats.PendingPatterns)
		fmt.Println("   Run: tokman learn show")
	}

	return nil
}

func init() {
	learnCmd.AddCommand(learnStartCmd)
	learnCmd.AddCommand(learnStopCmd)
	learnCmd.AddCommand(learnShowCmd)
	learnCmd.AddCommand(learnApplyCmd)
	learnCmd.AddCommand(learnClearCmd)

	// Flags
	learnCmd.Flags().Bool("status", true, "Show learning status")
	learnShowCmd.Flags().String("filter", "", "Filter by status: pending, approved, rejected")
	learnApplyCmd.Flags().Bool("dry-run", false, "Show what would be applied without changes")

	registry.Add(func() { registry.Register(learnCmd) })
}
