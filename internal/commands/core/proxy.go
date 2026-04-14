package core

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// proxyCmd executes commands without filtering but tracks usage
var proxyCmd = &cobra.Command{
	Use:   "proxy <command> [args...]",
	Short: "Execute command without filtering but track usage",
	Long: `Execute any command without applying TokMan filters, but still track
usage metrics for analytics.

This is useful when you want raw command output but still want to:
- Track command usage in TokMan analytics
- Record token counts for reporting
- Maintain a history of executed commands

Examples:
  tokman proxy ls -la
  tokman proxy cargo build --release
  tokman proxy ./custom-script.sh`,
	RunE: runProxy,
}

func init() {
	registry.Add(func() { registry.Register(proxyCmd) })
}

func runProxy(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	timer := tracking.Start()

	command := args[0]
	commandArgs := args[1:]

	// Execute the command directly without filtering
	execCmd := exec.Command(command, commandArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Env = os.Environ()

	// Run the command
	err := execCmd.Run()

	// Estimate tokens from the command itself (output is already streamed)
	cmdStr := strings.Join(args, " ")
	tokenCount := filter.EstimateTokens(cmdStr)

	// Track usage (even though we didn't filter, we track that it was executed)
	timer.Track(cmdStr, "tokman proxy", tokenCount, tokenCount)

	return err
}
