package analysis

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/quality"
	"github.com/GrayCodeAI/tokman/internal/visual"
)

var (
	qualityShowDiff   bool
	qualityExportHTML string
	qualityCompareAll bool
)

var qualityCmd = &cobra.Command{
	Use:   "quality [file]",
	Short: "Analyze compression quality (competitive feature vs LLMLingua/AutoCompressor)",
	Long: `Evaluate the quality of compression beyond just token count.

This is a competitive feature that automatically measures:
- Semantic preservation
- Structure integrity
- Readability score
- Information density
- Keyword preservation

Provides actionable recommendations for improving compression quality.

Examples:
  # Analyze compression quality from stdin
  cat file.txt | tokman quality

  # Show visual diff
  tokman quality --diff < input.txt

  # Compare all compression modes
  tokman quality --compare-all file.txt

  # Export HTML diff
  tokman quality --html output.html < input.txt`,
	RunE: runQuality,
}

func init() {
	qualityCmd.Flags().BoolVar(&qualityShowDiff, "diff", false, "show visual before/after comparison")
	qualityCmd.Flags().StringVar(&qualityExportHTML, "html", "", "export diff as HTML file")
	qualityCmd.Flags().BoolVar(&qualityCompareAll, "compare-all", false, "compare all compression modes")
	
	registry.Add(func() { registry.Register(qualityCmd) })
}

func runQuality(cmd *cobra.Command, args []string) error {
	// Read input
	var input []byte
	var err error
	
	if len(args) > 0 {
		input, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	}
	
	original := string(input)
	originalTokens := core.EstimateTokens(original)
	
	if qualityCompareAll {
		return compareAllModes(original, originalTokens)
	}
	
	// Compress with default settings
	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)
	compressed, _ := pipeline.Process(original)
	compressedTokens := core.EstimateTokens(compressed)
	
	// Calculate quality score
	score := quality.ScoreCompression(original, compressed, originalTokens, compressedTokens)
	
	// Display results
	fmt.Println(score.Details)
	
	// Show visual diff if requested
	if qualityShowDiff {
		fmt.Println("\n" + visual.VisualDiff(original, compressed, originalTokens, compressedTokens))
	}
	
	// Export HTML if requested
	if qualityExportHTML != "" {
		html := visual.ExportDiffHTML(original, compressed, originalTokens, compressedTokens)
		if err := os.WriteFile(qualityExportHTML, []byte(html), 0644); err != nil {
			return fmt.Errorf("write HTML: %w", err)
		}
		fmt.Printf("\n✅ HTML diff exported to: %s\n", qualityExportHTML)
	}
	
	return nil
}

func compareAllModes(original string, originalTokens int) error {
	fmt.Println("🔍 Comparing All Compression Modes...\n")
	
	modes := map[string]filter.Mode{
		"Minimal":    filter.ModeMinimal,
		"Aggressive": filter.ModeAggressive,
	}
	
	results := make(map[string]struct {
		Compressed string
		Tokens     int
	})
	
	// Run each mode
	for name, mode := range modes {
		cfg := filter.PipelineConfig{Mode: mode}
		pipeline := filter.NewPipelineCoordinator(cfg)
		compressed, _ := pipeline.Process(original)
		tokens := core.EstimateTokens(compressed)
		
		results[name] = struct {
			Compressed string
			Tokens     int
		}{compressed, tokens}
		
		fmt.Printf("Processing %s mode... ", name)
		fmt.Printf("%s\n", visual.CompactDiff(originalTokens, tokens))
	}
	
	// Calculate quality scores
	scores := quality.CompareCompressionMethods(original, originalTokens, results)
	
	// Find best method
	bestMethod, bestScore := quality.RecommendBestMethod(scores)
	
	fmt.Println("\n📊 QUALITY COMPARISON:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	for method, score := range scores {
		marker := "  "
		if method == bestMethod {
			marker = "🏆"
		}
		fmt.Printf("%s %s: %.1f%% (%s)\n", marker, method, score.Overall, score.Grade)
		fmt.Printf("     Compression: %.1f%% | Keywords: %.1f%% | Readability: %.1f%%\n",
			score.CompressionRatio, score.KeywordsPreserved, score.ReadabilityScore)
	}
	
	fmt.Println("\n✅ RECOMMENDATION:")
	fmt.Printf("Use '%s' mode for best quality (%.1f%% overall score)\n", 
		bestMethod, bestScore.Overall)
	
	return nil
}
