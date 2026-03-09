package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/discover"
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite <command>",
	Short: "Rewrite a command to use TokMan wrappers",
	Long: `Check if a command should be rewritten and output the TokMan version.
Used by shell hooks to automatically intercept commands.

Exit codes:
  0 - Command was rewritten (output to stdout)
  1 - No rewrite available (no output)

Example:
  tokman rewrite "git status"     # Output: tokman git status, exit 0
  tokman rewrite "ls -la"         # Output: tokman ls, exit 0
  tokman rewrite "cat file.txt"   # No output, exit 1 (no rewrite)`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Join all args as the command
		fullCmd := args[0]
		for i := 1; i < len(args); i++ {
			fullCmd += " " + args[i]
		}

		// Rewrite the command using the new registry
		rewritten, changed := discover.RewriteCommand(fullCmd, nil)

		// TokMan-style: exit 1 without output if no rewrite
		if !changed {
			os.Exit(1)
		}

		// Output the rewritten command (for shell hooks)
		fmt.Println(rewritten)

		// If verbose, show what happened
		if verbose > 0 {
			cyan := color.New(color.FgCyan).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Fprintf(cmd.ErrOrStderr(), "%s → %s\n", cyan(fullCmd), green(rewritten))
		}
	},
}

var rewriteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered command rewrites",
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dim := color.New(color.FgHiBlack).SprintFunc()

		fmt.Println(cyan("Registered Command Rewrites"))
		fmt.Println(dim("─────────────────────────────────────"))

		rewrites := discover.ListRewrites()
		for _, mapping := range rewrites {
			fmt.Printf("  %s → %s\n", green(mapping.Original), cyan(mapping.TokManCmd))
		}

		fmt.Println(dim("─────────────────────────────────────"))
		fmt.Printf("  %d commands registered\n", len(rewrites))
	},
}

func init() {
	rootCmd.AddCommand(rewriteCmd)
	rewriteCmd.AddCommand(rewriteListCmd)
}
