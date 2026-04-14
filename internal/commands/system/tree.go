package system

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var treeCmd = &cobra.Command{
	Use:   "tree [args...]",
	Short: "Directory tree with token-optimized output",
	Long: `Proxy to native tree with token-optimized output.

Supports all native tree flags like -L, -d, -a.
Filters noise directories (node_modules, .git, target, etc.) for cleaner output.

Examples:
  tokman tree
  tokman tree -L 2
  tokman tree -a -I 'node_modules|.git'`,
	RunE: runTree,
}

func init() {
	registry.Add(func() { registry.Register(treeCmd) })
	treeCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
}

func runTree(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Build tree command
	treeArgs := append([]string{}, args...)

	output, _, err := shared.RunAndCapture("tree", treeArgs)

	if shared.UltraCompact {
		for _, line := range strings.Split(output, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "directories") && strings.Contains(trimmed, "files") {
				fmt.Println(trimmed)
				return nil
			}
		}
		lines := strings.Split(output, "\n")
		fmt.Printf("%d entries\n", len(lines))
		return nil
	}

	// Apply filtering
	engine := filter.NewEngine(filter.ModeMinimal)
	filtered, _ := engine.Process(output)

	if err != nil {
		if hint := shared.TeeOnFailure(output, "tree", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Print(filtered)

	// Track
	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("tree %s", strings.Join(args, " ")), "tokman tree", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		return err
	}
	return nil
}
