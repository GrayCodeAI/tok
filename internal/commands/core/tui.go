package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch beautiful real-time TUI dashboard",
	Long: `Launch an interactive terminal UI with real-time metrics and statistics.

The TUI provides a beautiful, real-time view of:
- Command usage statistics
- Token savings metrics
- Cache performance
- Active sessions
- Recent commands with filtering

Features:
- Real-time updates every 5 seconds
- Multiple tabs: Overview, Commands, Cache, Stats
- Beautiful color schemes and progress bars
- Interactive navigation with keyboard shortcuts

Keyboard Shortcuts:
  tab         Switch to next tab
  shift+tab   Switch to previous tab
  r           Refresh data manually
  q/esc       Quit

Examples:
  tokman tui                    # Launch the dashboard
  tokman tui --refresh 10       # Refresh every 10 seconds`,
	RunE: runTUI,
}

var tuiRefreshRate int

func init() {
	registry.Add(func() { registry.Register(tuiCmd) })
	tuiCmd.Flags().IntVarP(&tuiRefreshRate, "refresh", "r", 5, "Refresh rate in seconds")
}

func runTUI(cmd *cobra.Command, args []string) error {
	if err := tui.RunDashboard(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
