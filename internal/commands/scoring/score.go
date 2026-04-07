package scoring

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/scoring"
)

var (
	scoreQuery    string
	scoreTopN     int
	scoreMinTier  string
	scoreKeywords []string
)

func init() {
	registry.Add(func() {
		registry.Register(scoreCmd)
	})
}

var scoreCmd = &cobra.Command{
	Use:   "score [file]",
	Short: "Score content using semantic signals",
	Long: `Analyze and score content using semantic signal scoring.

Scores content based on:
- Position (beginning/end weighted higher)
- Keywords (important terms)
- Frequency (rare terms weighted higher)
- Query relevance
- Semantic similarity

Output is sorted by relevance score.`,
	Example: `  tokman score file.txt
  tokman score file.txt --query="error handling"
  tokman score file.txt --top=20
  tokman score file.txt --tier=important`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScore,
}

func init() {
	scoreCmd.Flags().StringVar(&scoreQuery, "query", "", "Query for relevance scoring")
	scoreCmd.Flags().IntVar(&scoreTopN, "top", 50, "Show top N lines")
	scoreCmd.Flags().StringVar(&scoreMinTier, "tier", "", "Minimum tier (critical, important, nice_to_have)")
	scoreCmd.Flags().StringSliceVar(&scoreKeywords, "keywords", []string{}, "Important keywords")
}

func runScore(cmd *cobra.Command, args []string) error {
	// Read content
	var content []byte
	var err error

	if len(args) == 0 {
		content, err = os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("no input provided")
		}
	} else {
		content, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Create scoring engine
	engine := scoring.NewScoringEngine()

	// Add custom keywords
	for _, kw := range scoreKeywords {
		engine.AddKeyword(kw, 0.5)
	}

	// Score content
	opts := scoring.ScoringOptions{
		Query: scoreQuery,
	}

	result := engine.ScoreContent(string(content), opts)

	// Filter by tier if specified
	var lines []*scoring.ScoredLine
	if scoreMinTier != "" {
		tier := scoring.SignalTier(scoreMinTier)
		lines = result.FilterByTier(tier)
	} else {
		lines = result.Lines
	}

	// Get top N
	if len(lines) > scoreTopN {
		lines = lines[:scoreTopN]
	}

	// Output results
	fmt.Printf("\n%s\n\n", color.New(color.Bold).Sprint("Scoring Results"))
	fmt.Printf("Total lines: %d | Avg score: %.2f | Max: %.2f\n\n",
		result.TotalLines, result.AvgScore, result.MaxScore)

	// Tier distribution
	fmt.Printf("Tier distribution: ")
	for tier, count := range result.TierCounts {
		fmt.Printf("%s:%d ", tier, count)
	}
	fmt.Println()

	// Top lines
	fmt.Printf("%-6s %-10s %-8s %s\n", "LINE", "TIER", "SCORE", "CONTENT")
	fmt.Println(string(make([]byte, 80)))

	for _, line := range lines {
		tierColor := getTierColor(line.Tier)
		fmt.Printf("%-6d %-10s %-8.2f %s\n",
			line.LineNumber,
			tierColor.Sprintf("%s", line.Tier),
			line.Score,
			truncate(line.Content, 50))
	}

	fmt.Println()
	return nil
}

func getTierColor(tier scoring.SignalTier) *color.Color {
	switch tier {
	case scoring.TierCritical:
		return color.New(color.FgRed, color.Bold)
	case scoring.TierImportant:
		return color.New(color.FgYellow)
	case scoring.TierNiceToHave:
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgWhite)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
