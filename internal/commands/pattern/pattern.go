package pattern

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/pattern"
)

var (
	patternMinConfidence float64
	patternMaxResults    int
	patternType          string
)

func init() {
	registry.Add(func() {
		registry.Register(patternCmd)
	})
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

func init() {
	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered patterns",
		RunE:  runPatternList,
	}
	listCmd.Flags().Float64Var(&patternMinConfidence, "min-confidence", 0.7, "Minimum confidence threshold")
	listCmd.Flags().IntVar(&patternMaxResults, "limit", 50, "Maximum results")
	patternCmd.AddCommand(listCmd)

	// Discover subcommand
	discoverCmd := &cobra.Command{
		Use:   "discover [file]",
		Short: "Analyze file for patterns",
		RunE:  runPatternDiscover,
	}
	patternCmd.AddCommand(discoverCmd)

	// Show subcommand
	showCmd := &cobra.Command{
		Use:   "show <pattern-id>",
		Short: "Show pattern details",
		Args:  cobra.ExactArgs(1),
		RunE:  runPatternShow,
	}
	patternCmd.AddCommand(showCmd)

	// Delete subcommand
	deleteCmd := &cobra.Command{
		Use:   "delete <pattern-id>",
		Short: "Delete a pattern",
		Args:  cobra.ExactArgs(1),
		RunE:  runPatternDelete,
	}
	patternCmd.AddCommand(deleteCmd)
}

func runPatternList(cmd *cobra.Command, args []string) error {
	engine, err := pattern.NewPatternDiscoveryEngine()
	if err != nil {
		return fmt.Errorf("failed to create pattern engine: %w", err)
	}
	defer engine.Close()

	patterns := engine.GetPatterns(patternMinConfidence)

	if len(patterns) == 0 {
		fmt.Println("\nNo patterns discovered yet.")
		fmt.Println("Run 'tokman pattern discover <file>' to analyze content.")
		return nil
	}

	if len(patterns) > patternMaxResults {
		patterns = patterns[:patternMaxResults]
	}

	fmt.Printf("\n%s (%d patterns found)\n\n",
		color.New(color.Bold).Sprint("Discovered Patterns"),
		len(patterns))

	fmt.Printf("%-16s %-20s %-12s %-10s %s\n",
		"ID", "TYPE", "CONFIDENCE", "FREQ", "PATTERN")
	fmt.Println(string(make([]byte, 90)))

	for _, p := range patterns {
		patternStr := p.Pattern
		if len(patternStr) > 40 {
			patternStr = patternStr[:37] + "..."
		}

		fmt.Printf("%-16s %-20s %-12.2f %-10d %s\n",
			p.ID[:16],
			p.Type,
			p.Confidence,
			p.Frequency,
			patternStr)
	}

	fmt.Println()
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

	fmt.Printf("\n%s Analyzed %s\n\n", color.GreenString("✓"), source)
	fmt.Println("Patterns have been submitted for analysis.")
	fmt.Println("Run 'tokman pattern list' to see discovered patterns.")

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

	fmt.Printf("\n%s Pattern Details\n\n", color.New(color.Bold).Sprint("→"))
	fmt.Printf("  ID:          %s\n", p.ID)
	fmt.Printf("  Type:        %s\n", p.Type)
	fmt.Printf("  Pattern:     %s\n", p.Pattern)
	fmt.Printf("  Regex:       %s\n", p.Regex)
	fmt.Printf("  Confidence:  %.2f\n", p.Confidence)
	fmt.Printf("  Frequency:   %d\n", p.Frequency)
	fmt.Printf("  First Seen:  %s\n", p.FirstSeen.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Last Seen:   %s\n", p.LastSeen.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Status:      %s\n", p.Status)
	fmt.Printf("  Sources:     %v\n", p.SourceFiles)
	fmt.Printf("\n  Generated Filter:\n    %s\n\n", p.GenerateFilter())

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

	fmt.Printf("\n%s Deleted pattern %s\n\n", color.GreenString("✓"), patternID)
	return nil
}
