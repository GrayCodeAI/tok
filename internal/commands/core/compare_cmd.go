package core

import (
	"fmt"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/core"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare tok filtered vs unfiltered output",
	Long: `Run the same command both through tok and directly, then compare the results.

This benchmark shows exactly how many tokens and cost you're saving by using tok.

Examples:
  tok benchmark compare "git status"              # Compare git status
  tok benchmark compare "git log --oneline -20"  # Compare git log
  tok benchmark compare "docker ps"               # Compare docker output
  tok benchmark compare --runs 3                 # Run 3 times for avg`,
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
		out.Global().Println("Would compare:")
		out.Global().Printf("  1. tok %s\n", command)
		out.Global().Printf("  2. %s (raw)\n", command)
		return nil
	}

	out.Global().Printf("Comparing: %s\n", command)
	out.Global().Printf("Runs: %d\n", compareRuns)
	out.Global().Println()

	var totalRawTokens, totalFilteredTokens int
	var totalRawCost, totalFilteredCost float64

	for i := 0; i < compareRuns; i++ {
		if compareRuns > 1 {
			out.Global().Printf("Run %d/%d...\n", i+1, compareRuns)
		}

		rawOut, rawTokens, rawCost := runVanilla(command)
		tokOut, filteredTokens, filteredCost := runWithTok(command)

		totalRawTokens += rawTokens
		totalFilteredTokens += filteredTokens
		totalRawCost += rawCost
		totalFilteredCost += filteredCost

		if compareRuns == 1 {
			out.Global().Printf("  Raw:     %d tokens (~$%.4f)\n", rawTokens, rawCost)
			out.Global().Printf("  tok:  %d tokens (~$%.4f)\n", filteredTokens, filteredCost)
			saved := rawTokens - filteredTokens
			savingsPct := float64(saved) / float64(rawTokens) * 100
			out.Global().Printf("  Saved:   %d tokens (%.1f%%)\n", saved, savingsPct)
			out.Global().Println()
			showSample(rawOut, tokOut)
		}
	}

	if compareRuns > 1 {
		avgRawTokens := totalRawTokens / compareRuns
		avgFilteredTokens := totalFilteredTokens / compareRuns
		avgRawCost := totalRawCost / float64(compareRuns)
		avgFilteredCost := totalFilteredCost / float64(compareRuns)

		out.Global().Println("Averages:")
		out.Global().Printf("  Raw:     %d tokens (~$%.4f)\n", avgRawTokens, avgRawCost)
		out.Global().Printf("  tok:  %d tokens (~$%.4f)\n", avgFilteredTokens, avgFilteredCost)
		saved := avgRawTokens - avgFilteredTokens
		savingsPct := float64(saved) / float64(avgRawTokens) * 100
		out.Global().Printf("  Saved:   %d tokens (%.1f%%)\n", saved, savingsPct)
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

func runWithTok(cmd string) (output string, tokens int, cost float64) {
	c := exec.Command("tok", strings.Fields(cmd)...)
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

	out.Global().Println("Sample output (first 5 lines):")
	out.Global().Println("Raw:")
	for i, line := range rawLines {
		if i >= 5 {
			break
		}
		out.Global().Printf("  %s\n", line)
	}
	out.Global().Println("tok:")
	for i, line := range filteredLines {
		if i >= 5 {
			break
		}
		out.Global().Printf("  %s\n", line)
	}
}

func printCompareJSON(rawTokens, filteredTokens int, rawCost, filteredCost float64) {
	savedTokens := rawTokens - filteredTokens
	savedCost := rawCost - filteredCost
	savingsPct := float64(savedTokens) / float64(rawTokens) * 100

	out.Global().Printf(`{
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
