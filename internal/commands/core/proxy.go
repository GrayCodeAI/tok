package core

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// proxyCmd executes commands without filtering but tracks usage
var proxyCmd = &cobra.Command{
	Use:   "proxy <command> [args...]",
	Short: "Execute command without filtering but track usage",
	Long: `Execute any command without applying tok filters, but still track
usage metrics for analytics.

This is useful when you want raw command output but still want to:
- Track command usage in tok analytics
- Record token counts for reporting
- Maintain a history of executed commands

Examples:
  tok proxy ls -la
  tok proxy cargo build --release
  tok proxy ./custom-script.sh`,
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
	timer.Track(cmdStr, "tok proxy", tokenCount, tokenCount)

	return err
}
