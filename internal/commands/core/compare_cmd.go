package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/core"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare TokMan filtered vs unfiltered output",
	Long: `Run the same command both through TokMan and directly, then compare the results.

This benchmark shows exactly how many tokens and cost you're saving by using TokMan.

Examples:
  tokman benchmark compare "git status"              # Compare git status
  tokman benchmark compare "git log --oneline -20"  # Compare git log
  tokman benchmark compare "docker ps"               # Compare docker output
  tokman benchmark compare --runs 3                 # Run 3 times for avg`,
	Args: cobra.ExactArgs(1),
	RunE: runCompare,
}

var (
	compareRuns   int
	compareJSON   bool
	compareDryRun bool
)

func init() {
	registry.Add(func() { registry.Register(compareCmd) })

	compareCmd.Flags().IntVarP(&compareRuns, "runs", "r", 1, "Number of runs to average")
	compareCmd.Flags().BoolVar(&compareJSON, "json", false, "Output as JSON")
	compareCmd.Flags().BoolVar(&compareDryRun, "dry-run", false, "Show what would be compared")
}

func runCompare(cmd *cobra.Command, args []string) error {
	command := args[0]

	if compareDryRun {
		fmt.Println("Would compare:")
		fmt.Printf("  1. tokman %s\n", command)
		fmt.Printf("  2. %s (raw)\n", command)
		return nil
	}

	fmt.Printf("Comparing: %s\n", command)
	fmt.Printf("Runs: %d\n", compareRuns)
	fmt.Println()

	var totalRawTokens, totalFilteredTokens int
	var totalRawCost, totalFilteredCost float64

	for i := 0; i < compareRuns; i++ {
		if compareRuns > 1 {
			fmt.Printf("Run %d/%d...\n", i+1, compareRuns)
		}

		rawOut, rawTokens, rawCost := runVanilla(command)
		tokmanOut, filteredTokens, filteredCost := runWithTokman(command)

		totalRawTokens += rawTokens
		totalFilteredTokens += filteredTokens
		totalRawCost += rawCost
		totalFilteredCost += filteredCost

		if compareRuns == 1 {
			fmt.Printf("  Raw:     %d tokens (~$%.4f)\n", rawTokens, rawCost)
			fmt.Printf("  TokMan:  %d tokens (~$%.4f)\n", filteredTokens, filteredCost)
			saved := rawTokens - filteredTokens
			savingsPct := float64(saved) / float64(rawTokens) * 100
			fmt.Printf("  Saved:   %d tokens (%.1f%%)\n", saved, savingsPct)
			fmt.Println()
			showSample(rawOut, tokmanOut)
		}
	}

	if compareRuns > 1 {
		avgRawTokens := totalRawTokens / compareRuns
		avgFilteredTokens := totalFilteredTokens / compareRuns
		avgRawCost := totalRawCost / float64(compareRuns)
		avgFilteredCost := totalFilteredCost / float64(compareRuns)

		fmt.Println("Averages:")
		fmt.Printf("  Raw:     %d tokens (~$%.4f)\n", avgRawTokens, avgRawCost)
		fmt.Printf("  TokMan:  %d tokens (~$%.4f)\n", avgFilteredTokens, avgFilteredCost)
		saved := avgRawTokens - avgFilteredTokens
		savingsPct := float64(saved) / float64(avgRawTokens) * 100
		fmt.Printf("  Saved:   %d tokens (%.1f%%)\n", saved, savingsPct)
	}

	if compareJSON {
		printCompareJSON(totalRawTokens/compareRuns, totalFilteredTokens/compareRuns, totalRawCost/float64(compareRuns), totalFilteredCost/float64(compareRuns))
	}

	return nil
}

func runVanilla(cmd string) (output string, tokens int, cost float64) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", 0, 0
	}

	exe := parts[0]
	args := parts[1:]

	c := exec.Command(exe, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		output = fmt.Sprintf("error: %v\n%s", err, out)
	} else {
		output = string(out)
	}

	tokens = core.EstimateTokens(output)
	cost = tokenToCost(tokens)
	return
}

func runWithTokman(cmd string) (output string, tokens int, cost float64) {
	c := exec.Command("tokman", strings.Fields(cmd)...)
	out, err := c.CombinedOutput()
	if err != nil {
		output = fmt.Sprintf("error: %v\n%s", err, out)
	} else {
		output = string(out)
	}

	tokens = core.EstimateTokens(output)
	cost = tokenToCost(tokens)
	return
}

func tokenToCost(tokens int) float64 {
	costPer1M := 3.0 // Claude Sonnet per 1M input tokens
	return float64(tokens) / 1_000_000 * costPer1M
}

func showSample(raw, filtered string) {
	rawLines := strings.Split(raw, "\n")
	filteredLines := strings.Split(filtered, "\n")

	fmt.Println("Sample output (first 5 lines):")
	fmt.Println("Raw:")
	for i, line := range rawLines {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s\n", line)
	}
	fmt.Println("TokMan:")
	for i, line := range filteredLines {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s\n", line)
	}
}

func printCompareJSON(rawTokens, filteredTokens int, rawCost, filteredCost float64) {
	savedTokens := rawTokens - filteredTokens
	savedCost := rawCost - filteredCost
	savingsPct := float64(savedTokens) / float64(rawTokens) * 100

	fmt.Printf(`{
  "command": "comparison",
  "raw_tokens": %d,
  "filtered_tokens": %d,
  "saved_tokens": %d,
  "savings_percent": %.1f,
  "raw_cost_usd": %.4f,
  "filtered_cost_usd": %.4f,
  "saved_cost_usd": %.4f
}
`, rawTokens, filteredTokens, savedTokens, savingsPct, rawCost, filteredCost, savedCost)
}
