package filtercmd

import (
	"encoding/json"
	"fmt"
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/filter"
)

// engramCmd manages EngramLearner rules and statistics.
var engramCmd = &cobra.Command{
	Use:   "engram",
	Short: "Manage EngramLearner compression rules",
	Long: `EngramLearner learns from compression patterns and generates evidence-based rules.

Examples:
  tok filter engram stats              # Show learning statistics
  tok filter engram rules              # List all learned rules
  tok filter engram analyze <file>     # Analyze a file for patterns
  tok filter engram reset              # Clear all learned rules`,
}

var (
	engramJSON bool
)

func init() {
	engramCmd.PersistentFlags().BoolVar(&engramJSON, "json", false, "Output as JSON")

	// Subcommands
	engramCmd.AddCommand(engramStatsCmd)
	engramCmd.AddCommand(engramRulesCmd)
	engramCmd.AddCommand(engramAnalyzeCmd)
	engramCmd.AddCommand(engramResetCmd)

	registry.Add(func() { registry.Register(engramCmd) })
}

// engramStatsCmd shows EngramLearner statistics.
var engramStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show EngramLearner statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		learner := filter.NewEngramLearner()
		stats := learner.GetStats()

		if engramJSON {
			data, _ := json.MarshalIndent(stats, "", "  ")
			out.Global().Println(string(data))
		} else {
			out.Global().Println("╔════════════════════════════════════════════════════╗")
			out.Global().Println("║           EngramLearner Statistics                 ║")
			out.Global().Println("╠════════════════════════════════════════════════════╣")
			for key, val := range stats {
				out.Global().Printf("║ %-20s: %-28v ║\n", key, val)
			}
			out.Global().Println("╚════════════════════════════════════════════════════╝")
		}

		return nil
	},
}

// engramRulesCmd lists learned rules.
var engramRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List learned Engram rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		learner := filter.NewEngramLearner()
		rules := learner.GetRules()

		if engramJSON {
			data, _ := json.MarshalIndent(rules, "", "  ")
			out.Global().Println(string(data))
		} else {
			if len(rules) == 0 {
				out.Global().Println("No rules learned yet.")
				out.Global().Println("Rules are generated automatically as the system processes content.")
				return nil
			}

			out.Global().Println("╔════════════════════════════════════════════════════╗")
			out.Global().Println("║              Learned Engram Rules                  ║")
			out.Global().Println("╠════════════════════════════════════════════════════╣")
			for _, rule := range rules {
				out.Global().Printf("║ ID:       %s\n", rule.ID)
				out.Global().Printf("║ Name:     %s\n", rule.Name)
				out.Global().Printf("║ Type:     %s\n", rule.Type)
				out.Global().Printf("║ Severity: %s\n", rule.Severity)
				out.Global().Printf("║ Pattern:  %s\n", truncateString(rule.Pattern, 40))
				out.Global().Printf("║ Confidence: %.2f\n", rule.Confidence)
				out.Global().Println("╠────────────────────────────────────────────────────╣")
			}
			out.Global().Printf("║ Total: %d rules\n", len(rules))
			out.Global().Println("╚════════════════════════════════════════════════════╝")
		}

		return nil
	},
}

// engramAnalyzeCmd analyzes a file for patterns.
var engramAnalyzeCmd = &cobra.Command{
	Use:   "analyze <file>",
	Short: "Analyze a file for error patterns",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		learner := filter.NewEngramLearner()
		learner.Apply(string(content), filter.ModeMinimal)

		rules := learner.GetRulesForContent(string(content))

		if engramJSON {
			data, _ := json.MarshalIndent(rules, "", "  ")
			out.Global().Println(string(data))
		} else {
			out.Global().Printf("Analyzed: %s\n", args[0])
			out.Global().Printf("Applicable rules: %d\n\n", len(rules))

			for _, rule := range rules {
				out.Global().Printf("• %s (%s): %s\n", rule.Name, rule.Severity, rule.Type)
			}
		}

		return nil
	},
}

// engramResetCmd clears all learned rules.
var engramResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all learned Engram rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		learner := filter.NewEngramLearner()
		// Save empty rules
		learner.SaveRules()

		out.Global().Println("All Engram rules have been reset.")
		return nil
	},
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
