package core

import (
	"math/rand"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/config"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Show optimization tips and nudges",
	Long: `Display context-aware tips for improving token efficiency.

Shows suggestions based on your usage patterns and current settings.

Examples:
  tok suggest               # Show random tip
  tok suggest --category    # Show by category
  tok suggest --action      # Apply suggestion`,
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
		out.Global().Println("No tips found for this category.")
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	t := tips[r.Intn(len(tips))]

	out.Global().Println(color.CyanString("💡 Tip:"))
	out.Global().Println()
	bold := color.New(color.Bold)
	bold.Println(t.title)
	out.Global().Println(t.description)
	out.Global().Println()
	out.Global().Printf("Category: %s\n", t.category)
	if t.action != "" {
		out.Global().Printf("Action: %s\n", color.GreenString(t.action))
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
			action:      "tok agent set fast",
		},
		{
			title:       "Enable budget enforcement",
			description: "Set a token budget to automatically stop filtering once reached. Use --budget 2000 to limit output to 2000 tokens.",
			category:    "pipeline",
			action:      "tok config set pipeline.default_budget 2000",
		},
		{
			title:       "Use query intent for better filtering",
			description: "Specify --query debug, --query review, or --query deploy to enable goal-driven filtering that preserves relevant context.",
			category:    "pipeline",
			action:      "tok config set pipeline.enable_goal_driven true",
		},
		{
			title:       "Enable attribution for commits",
			description: "Track AI contributions with Co-Authored-By on git commits. Use tok attribution enable.",
			category:    "usage",
			action:      "tok attribution enable",
		},
		{
			title:       "Use MCP for IDE integration",
			description: "Connect tok to Claude Code, Cursor, or other IDEs via MCP for seamless token optimization.",
			category:    "config",
			action:      "tok mcp",
		},
	}

	if cfg.Pipeline.Preset == "" || cfg.Pipeline.Preset == "balanced" {
		tips = append(tips, tip{
			title:       "Try ultra-compact mode",
			description: "Use --ultra-compact for maximum compression with ASCII-only output, ideal for log files and large outputs.",
			category:    "pipeline",
			action:      "tok config set pipeline.enable_compaction true",
		})
	}

	if !cfg.Pipeline.EnableCompaction {
		tips = append(tips, tip{
			title:       "Enable semantic compaction",
			description: "Compaction summarizes long conversations while preserving key information. It can reduce context by 30-50%.",
			category:    "pipeline",
			action:      "tok config set pipeline.enable_compaction true",
		})
	}

	return tips
}
