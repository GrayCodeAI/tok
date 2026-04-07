// Package recoverycmd provides CLI commands for session recovery.
package recoverycmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/session_recovery"
)

var (
	headerColor = color.New(color.FgCyan, color.Bold)
	warnColor   = color.New(color.FgYellow)
	okColor     = color.New(color.FgGreen)
	dimColor    = color.New(color.Faint)
)

var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Session crash recovery and resume",
	Long: `Session recovery stores and manages session state to allow
resuming after crashes, power loss, or network disconnections.

Inspired by OMNI's transcript & recovery architecture.`,
	Example: `  tokman recovery --status        # Check for recoverable sessions
  tokman recovery list          # List all sessions
  tokman recovery resume <id>   # Resume a specific session
  tokman recovery close <id>    # Mark session as completed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showStatus, _ := cmd.Flags().GetBool("status")
		if showStatus || len(args) == 0 {
			return checkRecoveryStatus()
		}
		return cmd.Help()
	},
}

var recoveryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStore()
		if err != nil {
			return err
		}
		if store == nil {
			dimColor.Println("Session recovery is not enabled.")
			return nil
		}

		sessions, err := store.ListSessions()
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No active sessions.")
			return nil
		}

		headerColor.Println("Active Sessions")
		fmt.Println(strings.Repeat("─", 70))
		fmt.Printf("%-10s  %-20s  %5s  %8s  %s\n",
			"ID", "Started", "Cmds", "Hot F.", "Last Updated")
		fmt.Println(strings.Repeat("─", 70))

		for _, s := range sessions {
			fmt.Printf("%-10s  %-20s  %5d  %8d  %s\n",
				s.ID,
				s.StartedAt.Format("15:04:05"),
				len(s.Commands),
				len(s.HotFiles),
				s.LastUpdate.Format("15:04:05"),
			)
		}

		fmt.Println(strings.Repeat("─", 70))
		return nil
	},
}

var recoveryResumeCmd = &cobra.Command{
	Use:   "resume [session-id]",
	Short: "Resume a previous session",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStore()
		if err != nil {
			return err
		}
		if store == nil {
			dimColor.Println("Session recovery is not enabled.")
			return nil
		}

		if len(args) == 0 {
			// Auto-find most recent
			info, err := store.CheckRecovery()
			if err != nil {
				return fmt.Errorf("check recovery: %w", err)
			}
			if !info.CanResume {
				fmt.Println("No sessions to resume.")
				return nil
			}

			warnColor.Printf("Recovering session %s...\n", info.Session.ID)
			fmt.Println(info.Summary)
			fmt.Println()

			if len(info.InterruptedCommands) > 0 {
				headerColor.Println("Last commands before interruption:")
				for _, c := range info.InterruptedCommands {
					fmt.Printf("  $ %s\n", c.Command)
					if c.Output != "" {
						fmt.Printf("    %s\n", truncate(c.Output, 60))
					}
				}
			}

			return nil
		}

		sessionID := args[0]
		sessions, err := store.ListSessions()
		if err != nil {
			return err
		}

		for _, s := range sessions {
			if s.ID == sessionID {
				okColor.Printf("Resuming session %s\n", s.ID)
				fmt.Printf("Started: %s\n", s.StartedAt.Format("15:04:05"))
				fmt.Printf("Commands executed: %d\n", len(s.Commands))
				fmt.Printf("Hot files: %d\n", len(s.HotFiles))
				fmt.Printf("Checkpoint: %d\n", s.Checkpoint)

				if len(s.HotFiles) > 0 {
					fmt.Println()
					headerColor.Println("Hot files (most recent):")
					for path, stat := range s.HotFiles {
						fmt.Printf("  %s (%d accesses)\n", path, stat.AccessCount)
					}
				}

				return nil
			}
		}

		return fmt.Errorf("session not found: %s", sessionID)
	},
}

var recoveryCloseCmd = &cobra.Command{
	Use:   "close [session-id]",
	Short: "Close a session (mark as completed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStore()
		if err != nil {
			return err
		}
		if store == nil {
			dimColor.Println("Session recovery is not enabled.")
			return nil
		}

		if len(args) == 0 {
			// Close most recent
			sessions, err := store.ListSessions()
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				fmt.Println("No active sessions to close.")
				return nil
			}
			args = []string{sessions[0].ID}
		}

		if err := store.CloseSession(args[0]); err != nil {
			return fmt.Errorf("close session: %w", err)
		}

		okColor.Printf("Session %s closed and archived.\n", args[0])
		return nil
	},
}

var recoveryStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check for recoverable sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkRecoveryStatus()
	},
}

func checkRecoveryStatus() error {
	store, err := getStore()
	if err != nil {
		return err
	}
	if store == nil {
		dimColor.Println("Session recovery is not enabled.")
		return nil
	}

	info, err := store.CheckRecovery()
	if err != nil {
		return fmt.Errorf("check recovery: %w", err)
	}

	if !info.CanResume {
		okColor.Println("✓ No interrupted sessions to recover")
		return nil
	}

	warnColor.Println("⚠ Interrupted session detected")
	fmt.Println()
	fmt.Println(info.Summary)
	fmt.Println()
	fmt.Println("To resume:")
	fmt.Println("  tokman recovery resume")
	fmt.Println()
	fmt.Println("To list all sessions:")
	fmt.Println("  tokman recovery list")

	return nil
}

func getStore() (*session_recovery.RecoveryStore, error) {
	cfg := session_recovery.DefaultConfig()
	return session_recovery.New(cfg)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func init() {
	recoveryCmd.AddCommand(recoveryListCmd)
	recoveryCmd.AddCommand(recoveryResumeCmd)
	recoveryCmd.AddCommand(recoveryCloseCmd)
	recoveryCmd.AddCommand(recoveryStatusCmd)

	recoveryCmd.Flags().Bool("status", true, "Check for recoverable sessions")

	registry.Add(func() { registry.Register(recoveryCmd) })
}
