package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/discover"
)

// Exit codes for rewrite command (protocol for hook scripts)
const (
	ExitRewriteAllow      = 0 // Rewrite found, auto-allow
	ExitNoRewrite         = 1 // No tok equivalent, pass-through
	ExitDeny              = 2 // Deny rule matched
	ExitRewriteAsk        = 3 // Rewrite found, ask user for confirmation
	ExitInvalidInput      = 4 // Invalid input
	ExitCommandDisabled   = 5 // Command is in disabled list
	ExitUnsafeOperation   = 6 // Unsafe operation detected
	ExitResourceIntensive = 7 // Resource-intensive operation
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite <command>",
	Short: "Rewrite a command to use tok wrappers (for hook scripts)",
	Long: `Check if a command should be rewritten and output the tok version.
Used by shell hooks to automatically intercept commands.

This is the single source of truth for command rewriting logic.
Hook scripts delegate to this command instead of implementing
rewrite rules in shell code.

Exit codes:
  0 - Rewrite found, auto-allow
  1 - No tok equivalent, pass-through unchanged
  2 - Deny rule matched (dangerous command)
  3 - Rewrite found, ask user for confirmation
  4 - Invalid input
  5 - Command is in disabled list
  6 - Unsafe operation detected
  7 - Resource-intensive operation

Examples:
  tok rewrite "git status"          # Output: tok git status, exit 0
  tok rewrite "echo hello"          # No output, exit 1
  tok rewrite "rm -rf /"            # No output, exit 2
  tok rewrite "sudo apt upgrade"    # Output: tok sudo apt upgrade, exit 3`,
	Args:               cobra.MinimumNArgs(1),
	RunE:               runRewrite,
	SilenceUsage:       true,
	SilenceErrors:      true,
	DisableFlagParsing: true,
}

func runRewrite(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		os.Exit(ExitInvalidInput)
	}

	fullCmd := strings.Join(args, " ")
	parts := strings.Fields(fullCmd)

	if len(parts) == 0 {
		os.Exit(ExitInvalidInput)
	}

	baseCmd := parts[0]

	// Check if command is already using tok
	if baseCmd == "tok" {
		os.Exit(ExitNoRewrite)
	}

	// Check deny rules (dangerous commands)
	if isDenied(baseCmd, parts) {
		os.Exit(ExitDeny)
	}

	// Check if command is disabled
	if isDisabled(baseCmd) {
		os.Exit(ExitCommandDisabled)
	}

	// Check unsafe operations
	if isUnsafe(baseCmd, parts) {
		os.Exit(ExitUnsafeOperation)
	}

	// Try to rewrite using discover package
	rewritten, changed := discover.RewriteCommand(fullCmd, nil)

	if !changed {
		// No rewrite available
		os.Exit(ExitNoRewrite)
	}

	// Check if requires user confirmation (ask rules)
	if requiresConfirmation(baseCmd, parts) {
		fmt.Println(rewritten)
		os.Exit(ExitRewriteAsk)
	}

	// Check if resource-intensive
	if isResourceIntensive(baseCmd, parts) {
		fmt.Println(rewritten)
		os.Exit(ExitResourceIntensive)
	}

	// Rewrite and auto-allow
	fmt.Println(rewritten)

	if shared.Verbose > 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(cmd.ErrOrStderr(), "%s → %s\n", cyan(fullCmd), green(rewritten))
	}

	os.Exit(ExitRewriteAllow)
	return nil
}

// Safety check functions

func isDenied(baseCmd string, parts []string) bool {
	// Dangerous commands that should never be rewritten
	denyList := []string{
		"rm",
		"dd",
		"mkfs",
		"fdisk",
		"parted",
		":(){:|:&};:", // Fork bomb
		">/dev/sda",
	}

	for _, denied := range denyList {
		if baseCmd == denied {
			return true
		}
	}

	// Check for dangerous patterns
	cmdStr := strings.Join(parts, " ")
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=",
		">/dev/",
		"chmod -R 777",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdStr, pattern) {
			return true
		}
	}

	return false
}

func isDisabled(baseCmd string) bool {
	// Read disabled commands from config if available
	cfg, err := config.Load("")
	if err == nil && cfg != nil && cfg.Hooks.ExcludedCommands != nil {
		for _, cmd := range cfg.Hooks.ExcludedCommands {
			if baseCmd == cmd {
				return true
			}
		}
	}
	return false
}

func isUnsafe(baseCmd string, parts []string) bool {
	unsafeCommands := []string{
		"curl",
		"wget",
		"ssh",
		"scp",
		"rsync",
	}

	for _, unsafe := range unsafeCommands {
		if baseCmd == unsafe {
			// Check if piping to shell
			cmdStr := strings.Join(parts, " ")
			if strings.Contains(cmdStr, "| sh") ||
				strings.Contains(cmdStr, "| bash") ||
				strings.Contains(cmdStr, "| zsh") {
				return true
			}
		}
	}

	return false
}

func requiresConfirmation(baseCmd string, parts []string) bool {
	askList := []string{
		"sudo",      // Privileged operations
		"su",        // Switch user
		"systemctl", // System control
	}

	for _, ask := range askList {
		if baseCmd == ask {
			return true
		}
	}

	// Check for destructive flags
	cmdStr := strings.Join(parts, " ")
	destructivePatterns := []string{
		"--force",
		"-f",
		"--yes",
		"-y",
		"--delete",
	}

	for _, pattern := range destructivePatterns {
		if strings.Contains(cmdStr, pattern) {
			// Some commands are safe with these flags
			safeCmds := []string{"git", "cargo", "npm", "go"}
			isSafe := false
			for _, safe := range safeCmds {
				if baseCmd == safe {
					isSafe = true
					break
				}
			}
			if !isSafe {
				return true
			}
		}
	}

	return false
}

func isResourceIntensive(baseCmd string, parts []string) bool {
	intensiveCommands := []string{
		"find",
		"grep",
		"ag",
		"rg",
		"fd",
	}

	for _, intensive := range intensiveCommands {
		if baseCmd == intensive {
			// Check if operating on large directories
			cmdStr := strings.Join(parts, " ")
			if strings.Contains(cmdStr, "/") ||
				strings.Contains(cmdStr, "-r") ||
				strings.Contains(cmdStr, "--recursive") {
				return true
			}
		}
	}

	return false
}

func init() {
	registry.Add(func() { registry.Register(rewriteCmd) })
}
