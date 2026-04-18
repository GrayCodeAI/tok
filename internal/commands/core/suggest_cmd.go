package core

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Show optimization tips and nudges",
	Long: `Display context-aware tips for improving token efficiency.

Shows suggestions based on your usage patterns and current settings.

Examples:
  tokman suggest               # Show random tip
  tokman suggest --category    # Show by category
  tokman suggest --action      # Apply suggestion`,
	RunE: runSuggest,
}

var suggestCategory string

func init() {
	registry.Add(func() { registry.Register(suggestCmd) })

	suggestCmd.Flags().StringVar(&suggestCategory, "category", "", "Filter by category (pipeline, config, usage)")
	suggestCmd.Flags().StringVar(&suggestCategory, "c", "", "Filter by category (short)")
}

func runSuggest(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	tips := getTips(cfg)

	if suggestCategory != "" {
		var filtered []tip
		for _, t := range tips {
			if t.category == suggestCategory {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) > 0 {
			tips = filtered
		}
	}

	if len(tips) == 0 {
		fmt.Println("No tips found for this category.")
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	t := tips[rand.Intn(len(tips))]

	fmt.Println(color.CyanString("💡 Tip:"))
	fmt.Println()
	bold := color.New(color.Bold)
	bold.Println(t.title)
	fmt.Println(t.description)
	fmt.Println()
	fmt.Printf("Category: %s\n", t.category)
	if t.action != "" {
		fmt.Printf("Action: %s\n", color.GreenString(t.action))
	}

	return nil
}

type tip struct {
	title       string
	description string
	category    string
	action      string
}

func getTips(cfg *config.Config) []tip {
	tips := []tip{
		{
			title:       "Use presets for faster filtering",
			description: "The --preset flag lets you quickly switch between compression levels: fast (50-60%), balanced (70-80%), or full (85-95%).",
			category:    "pipeline",
			action:      "tokman agent set fast",
		},
		{
			title:       "Enable budget enforcement",
			description: "Set a token budget to automatically stop filtering once reached. Use --budget 2000 to limit output to 2000 tokens.",
			category:    "pipeline",
			action:      "tokman config set pipeline.default_budget 2000",
		},
		{
			title:       "Use query intent for better filtering",
			description: "Specify --query debug, --query review, or --query deploy to enable goal-driven filtering that preserves relevant context.",
			category:    "pipeline",
			action:      "tokman config set pipeline.enable_goal_driven true",
		},
		{
			title:       "Enable attribution for commits",
			description: "Track AI contributions with Co-Authored-By on git commits. Use tokman attribution enable.",
			category:    "usage",
			action:      "tokman attribution enable",
		},
		{
			title:       "Use MCP for IDE integration",
			description: "Connect TokMan to Claude Code, Cursor, or other IDEs via MCP for seamless token optimization.",
			category:    "config",
			action:      "tokman mcp",
		},
	}

	if cfg.Pipeline.Preset == "" || cfg.Pipeline.Preset == "balanced" {
		tips = append(tips, tip{
			title:       "Try ultra-compact mode",
			description: "Use --ultra-compact for maximum compression with ASCII-only output, ideal for log files and large outputs.",
			category:    "pipeline",
			action:      "tokman config set pipeline.enable_compaction true",
		})
	}

	if !cfg.Pipeline.EnableCompaction {
		tips = append(tips, tip{
			title:       "Enable semantic compaction",
			description: "Compaction summarizes long conversations while preserving key information. It can reduce context by 30-50%.",
			category:    "pipeline",
			action:      "tokman config set pipeline.enable_compaction true",
		})
	}

	return tips
}
