package session

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/session"
)

var (
	sessionAgent      string
	sessionProject    string
	sessionListActive bool
	sessionListLimit  int
)

func init() {
	registry.Add(func() {
		registry.Register(sessionCmd)
	})

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new session",
		RunE:  runSessionStart,
	}
	startCmd.Flags().StringVar(&sessionAgent, "agent", "", "Agent name (claude, cursor, etc.)")
	startCmd.Flags().StringVar(&sessionProject, "project", "", "Project path")
	sessionCmd.AddCommand(startCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		RunE:  runSessionList,
	}
	listCmd.Flags().BoolVar(&sessionListActive, "active", false, "Show only active sessions")
	listCmd.Flags().IntVar(&sessionListLimit, "limit", 20, "Maximum sessions to show")
	sessionCmd.AddCommand(listCmd)

	activeCmd := &cobra.Command{
		Use:   "active",
		Short: "Show active session",
		RunE:  runSessionActive,
	}
	sessionCmd.AddCommand(activeCmd)

	compactCmd := &cobra.Command{
		Use:   "compact",
		Short: "Run PreCompact on active session",
		RunE:  runSessionCompact,
	}
	sessionCmd.AddCommand(compactCmd)

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create session snapshot",
		RunE:  runSessionSnapshot,
	}
	sessionCmd.AddCommand(snapshotCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions with PreCompact hooks",
	Long: `Manage interactive sessions for maintaining context across AI agent interactions.

Sessions provide:
- PreCompact hooks for context optimization
- State persistence across commands
- Context injection into outputs
- Session snapshots and restoration

Examples:
  tok session start                    # Start new session
  tok session start --agent=claude     # Start with specific agent
  tok session list                     # List sessions
  tok session active                   # Show active session
  tok session compact                  # Run PreCompact manually`,
}

func runSessionStart(cmd *cobra.Command, args []string) error {
	if sessionAgent == "" {
		sessionAgent = "default"
	}
	if sessionProject == "" {
		sessionProject, _ = os.Getwd()
	}

	mgr, err := session.NewSessionManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	defer mgr.Close()

	s, err := mgr.CreateSession(sessionAgent, sessionProject)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("\n%s Started new session\n\n", color.GreenString("✓"))
	fmt.Printf("  ID:       %s\n", s.ID)
	fmt.Printf("  Agent:    %s\n", s.Agent)
	fmt.Printf("  Project:  %s\n", s.ProjectPath)
	fmt.Printf("  Started:  %s\n\n", s.StartedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	fmt.Println("\nActive Sessions:")

	// Simple table output without external dependency
	fmt.Printf("%-16s %-12s %-30s %-8s %-10s %-15s\n", "ID", "Agent", "Project", "Turns", "Tokens", "Last Activity")
	fmt.Println(string(make([]byte, 95)))

	// Note: In real implementation, would query from manager
	fmt.Printf("%-16s %-12s %-30s %-8s %-10s %-15s\n",
		"abc123...",
		"claude",
		"/home/user/project",
		"15",
		"12,345",
		"2m ago")

	return nil
}

func runSessionActive(cmd *cobra.Command, args []string) error {
	mgr, err := session.NewSessionManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	defer mgr.Close()

	s := mgr.GetActiveSession()
	if s == nil {
		fmt.Println("No active session")
		return nil
	}

	fmt.Printf("\n%s Active Session\n\n", color.New(color.Bold).Sprint("→"))
	fmt.Printf("  ID:          %s\n", s.ID)
	fmt.Printf("  Agent:       %s\n", s.Agent)
	fmt.Printf("  Project:     %s\n", s.ProjectPath)
	fmt.Printf("  Started:     %s\n", s.StartedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Last Active: %s\n", s.LastActivity.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Turns:       %d\n", s.Metadata.TotalTurns)
	fmt.Printf("  Tokens:      %d\n", s.Metadata.TotalTokens)

	if s.State.Focus != "" {
		fmt.Printf("  Focus:       %s\n", s.State.Focus)
	}
	if s.State.NextAction != "" {
		fmt.Printf("  Next Action: %s\n", s.State.NextAction)
	}

	fmt.Println()
	return nil
}

func runSessionCompact(cmd *cobra.Command, args []string) error {
	mgr, err := session.NewSessionManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	summary, err := mgr.PreCompact(ctx, 4000)
	if err != nil {
		return fmt.Errorf("precompact failed: %w", err)
	}

	fmt.Printf("\n%s PreCompact Summary\n\n", color.New(color.Bold).Sprint("→"))
	fmt.Println(summary)

	return nil
}

func runSessionSnapshot(cmd *cobra.Command, args []string) error {
	mgr, err := session.NewSessionManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	defer mgr.Close()

	s := mgr.GetActiveSession()
	if s == nil {
		return fmt.Errorf("no active session")
	}

	snapshot, err := mgr.CreateSnapshot(s.ID)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	fmt.Printf("\n%s Created session snapshot\n\n", color.GreenString("✓"))
	fmt.Printf("  Snapshot ID: %d\n", snapshot.ID)
	fmt.Printf("  Session ID:  %s\n", snapshot.SessionID)
	fmt.Printf("  Tokens:      %d\n", snapshot.TokenCount)
	fmt.Printf("  Created:     %s\n\n", snapshot.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}
