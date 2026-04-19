package filtercmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/filter"
)

// summarizeCmd generates tiered summaries (L0/L1/L2).
var summarizeCmd = &cobra.Command{
	Use:   "summarize <file>",
	Short: "Generate tiered summaries (L0/L1/L2)",
	Long: `Generate progressive summaries at different detail levels:

L0 (Surface): Keywords, entities, topics
L1 (Structural): Sections, outline, hierarchy  
L2 (Deep): Semantic summary with key points

Examples:
  tok filter summarize document.txt           # Auto-select tier
  tok filter summarize document.txt --tier=l2 # Force L2 deep summary
  tok filter summarize document.txt --json    # JSON output`,
	Args: cobra.ExactArgs(1),
	RunE: runSummarize,
}

var (
	summarizeTier string
	summarizeJSON bool
)

func init() {
	summarizeCmd.Flags().StringVar(&summarizeTier, "tier", "auto", "Summary tier: l0|l1|l2|auto")
	summarizeCmd.Flags().BoolVar(&summarizeJSON, "json", false, "Output as JSON")
	registry.Add(func() { registry.Register(summarizeCmd) })
}

func runSummarize(cmd *cobra.Command, args []string) error {
	content, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	tsf := filter.NewTieredSummaryFilter()
	tiers := tsf.GenerateTiers(string(content))

	if summarizeJSON {
		output := map[string]interface{}{
			"file": args[0],
			"l0":   tiers.L0,
			"l1":   tiers.L1,
			"l2":   tiers.L2,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Determine which tier to show
	tier := summarizeTier
	if tier == "auto" {
		// Select based on content length
		if len(content) > 10000 {
			tier = "l2"
		} else if len(content) > 2000 {
			tier = "l1"
		} else {
			tier = "l0"
		}
	}

	fmt.Printf("╔════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Tiered Summary: %-34s ║\n", args[0])
	fmt.Printf("╚════════════════════════════════════════════════════╝\n\n")

	switch tier {
	case "l0", "L0":
		printL0Summary(tiers.L0)
	case "l1", "L1":
		printL1Summary(tiers.L1)
	case "l2", "L2":
		printL2Summary(tiers.L2)
	default:
		// Show all tiers
		fmt.Println("=== L0: Surface Summary ===")
		printL0Summary(tiers.L0)
		fmt.Println("\n=== L1: Structural Summary ===")
		printL1Summary(tiers.L1)
		fmt.Println("\n=== L2: Deep Summary ===")
		printL2Summary(tiers.L2)
	}

	return nil
}

func printL0Summary(l0 *filter.L0Summary) {
	if l0 == nil {
		fmt.Println("No L0 summary available.")
		return
	}

	if len(l0.Topics) > 0 {
		fmt.Printf("Topics:   %v\n", l0.Topics)
	}
	if len(l0.Keywords) > 0 {
		fmt.Printf("Keywords: %v\n", l0.Keywords)
	}
	if len(l0.Entities) > 0 {
		fmt.Printf("Entities: %v\n", l0.Entities)
	}
	fmt.Printf("Tokens:   %d\n", l0.TokenCount)
}

func printL1Summary(l1 *filter.L1Summary) {
	if l1 == nil {
		fmt.Println("No L1 summary available.")
		return
	}

	if l1.Title != "" {
		fmt.Printf("Title: %s\n", l1.Title)
	}

	if l1.Outline != "" {
		fmt.Println("\nOutline:")
		fmt.Println(l1.Outline)
	}

	if len(l1.Sections) > 0 {
		fmt.Println("\nSections:")
		for _, sec := range l1.Sections {
			indent := ""
			for i := 0; i < sec.Level-1; i++ {
				indent += "  "
			}
			fmt.Printf("%s- %s\n", indent, sec.Heading)
			if sec.Summary != "" {
				fmt.Printf("%s  %s\n", indent, truncateString(sec.Summary, 60))
			}
		}
	}
	fmt.Printf("\nTokens: %d\n", l1.TokenCount)
}

func printL2Summary(l2 *filter.L2Summary) {
	if l2 == nil {
		fmt.Println("No L2 summary available.")
		return
	}

	if l2.Summary != "" {
		fmt.Println("Summary:")
		fmt.Println(l2.Summary)
	}

	if len(l2.KeyPoints) > 0 {
		fmt.Println("\nKey Points:")
		for i, point := range l2.KeyPoints {
			fmt.Printf("%d. %s\n", i+1, point)
		}
	}

	if len(l2.Implications) > 0 {
		fmt.Println("\nImplications:")
		for _, imp := range l2.Implications {
			fmt.Printf("• %s\n", imp)
		}
	}
	fmt.Printf("\nTokens: %d\n", l2.TokenCount)
}
