package pattern

import (
	"fmt"
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/pattern"
)

var (
	patternMinConfidence float64
	patternMaxResults    int
)

func init() {
	registry.Add(func() {
		registry.Register(patternCmd)
	})

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered patterns",
		RunE:  runPatternList,
	}
	listCmd.Flags().Float64Var(&patternMinConfidence, "min-confidence", 0.7, "Minimum confidence threshold")
	listCmd.Flags().IntVar(&patternMaxResults, "limit", 50, "Maximum results")
	patternCmd.AddCommand(listCmd)

	discoverCmd := &cobra.Command{
		Use:   "discover [file]",
		Short: "Analyze file for patterns",
		RunE:  runPatternDiscover,
	}
	patternCmd.AddCommand(discoverCmd)

	showCmd := &cobra.Command{
		Use:   "show <pattern-id>",
		Short: "Show pattern details",
		Args:  cobra.ExactArgs(1),
		RunE:  runPatternShow,
	}
	patternCmd.AddCommand(showCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <pattern-id>",
		Short: "Delete a pattern",
		Args:  cobra.ExactArgs(1),
		RunE:  runPatternDelete,
	}
	patternCmd.AddCommand(deleteCmd)
}

var patternCmd = &cobra.Command{
	Use:   "pattern",
	Short: "Pattern discovery and management",
	Long: `Discover and manage patterns in content.

Pattern discovery automatically identifies:
- Common log patterns
- Error patterns
- File paths
- Timestamps
- Hashes
- Stack traces

Discovered patterns can be used to create filters.`,
}

func runPatternList(cmd *cobra.Command, args []string) error {
	engine, err := pattern.NewPatternDiscoveryEngine()
	if err != nil {
		return fmt.Errorf("failed to create pattern engine: %w", err)
	}
	defer engine.Close()

	patterns := engine.GetPatterns(patternMinConfidence)

	if len(patterns) == 0 {
		out.Global().Println("\nNo patterns discovered yet.")
		out.Global().Println("Run 'tok pattern discover <file>' to analyze content.")
		return nil
	}

	if len(patterns) > patternMaxResults {
		patterns = patterns[:patternMaxResults]
	}

	out.Global().Printf("\n%s (%d patterns found)\n\n",
		color.New(color.Bold).Sprint("Discovered Patterns"),
		len(patterns))

	out.Global().Printf("%-16s %-20s %-12s %-10s %s\n",
		"ID", "TYPE", "CONFIDENCE", "FREQ", "PATTERN")
	out.Global().Println(string(make([]byte, 90)))

	for _, p := range patterns {
		patternStr := p.Pattern
		if len(patternStr) > 40 {
			patternStr = patternStr[:37] + "..."
		}

		out.Global().Printf("%-16s %-20s %-12.2f %-10d %s\n",
			p.ID[:16],
			p.Type,
			p.Confidence,
			p.Frequency,
			patternStr)
	}

	out.Global().Println()
	return nil
}

func runPatternDiscover(cmd *cobra.Command, args []string) error {
	var content []byte
	var source string
	var err error

	if len(args) == 0 {
		content, err = os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("no input provided")
		}
		source = "stdin"
	} else {
		content, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		source = args[0]
	}

	engine, err := pattern.NewPatternDiscoveryEngine()
	if err != nil {
		return fmt.Errorf("failed to create pattern engine: %w", err)
	}
	defer engine.Close()

	// Start engine
	engine.Start()

	// Submit sample
	engine.SubmitSample(string(content), source)

	// Stop and consolidate
	engine.Stop()

	out.Global().Printf("\n%s Analyzed %s\n\n", color.GreenString("✓"), source)
	out.Global().Println("Patterns have been submitted for analysis.")
	out.Global().Println("Run 'tok pattern list' to see discovered patterns.")

	return nil
}

func runPatternShow(cmd *cobra.Command, args []string) error {
	patternID := args[0]

	engine, err := pattern.NewPatternDiscoveryEngine()
	if err != nil {
		return fmt.Errorf("failed to create pattern engine: %w", err)
	}
	defer engine.Close()

	p, found := engine.GetPatternByID(patternID)
	if !found {
		return fmt.Errorf("pattern not found: %s", patternID)
	}

	out.Global().Printf("\n%s Pattern Details\n\n", color.New(color.Bold).Sprint("→"))
	out.Global().Printf("  ID:          %s\n", p.ID)
	out.Global().Printf("  Type:        %s\n", p.Type)
	out.Global().Printf("  Pattern:     %s\n", p.Pattern)
	out.Global().Printf("  Regex:       %s\n", p.Regex)
	out.Global().Printf("  Confidence:  %.2f\n", p.Confidence)
	out.Global().Printf("  Frequency:   %d\n", p.Frequency)
	out.Global().Printf("  First Seen:  %s\n", p.FirstSeen.Format("2006-01-02 15:04:05"))
	out.Global().Printf("  Last Seen:   %s\n", p.LastSeen.Format("2006-01-02 15:04:05"))
	out.Global().Printf("  Status:      %s\n", p.Status)
	out.Global().Printf("  Sources:     %v\n", p.SourceFiles)
	out.Global().Printf("\n  Generated Filter:\n    %s\n\n", p.GenerateFilter())

	return nil
}

func runPatternDelete(cmd *cobra.Command, args []string) error {
	patternID := args[0]

	engine, err := pattern.NewPatternDiscoveryEngine()
	if err != nil {
		return fmt.Errorf("failed to create pattern engine: %w", err)
	}
	defer engine.Close()

	if err := engine.DeletePattern(patternID); err != nil {
		return fmt.Errorf("failed to delete pattern: %w", err)
	}

	out.Global().Printf("\n%s Deleted pattern %s\n\n", color.GreenString("✓"), patternID)
	return nil
}
