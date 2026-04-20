package core

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	appui "github.com/lakshmanpatel/tok/internal/tui"
)

var (
	tuiRefresh  string
	tuiDays     int
	tuiProject  string
	tuiAgent    string
	tuiProvider string
	tuiModel    string
	tuiSession  string
	tuiTheme    string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the tok token intelligence TUI",
	Long: `Launch the new tok terminal dashboard.

Phase 1 includes:
- shared Bubble Tea shell
- real workspace snapshot loading
- Home dashboard
- section navigation and refresh

Additional screens are being implemented incrementally on top of the same data layer.`,
	RunE: runTUI,
}

func init() {
	registry.Add(func() { registry.Register(tuiCmd) })

	tuiCmd.Flags().StringVar(&tuiRefresh, "refresh", "20s", "Refresh interval, e.g. 10s, 30s, 1m")
	tuiCmd.Flags().IntVar(&tuiDays, "days", 30, "Active dashboard window in days")
	tuiCmd.Flags().StringVar(&tuiProject, "project", "", "Filter to a project path")
	tuiCmd.Flags().StringVar(&tuiAgent, "agent", "", "Filter to an agent")
	tuiCmd.Flags().StringVar(&tuiProvider, "provider", "", "Filter to a provider")
	tuiCmd.Flags().StringVar(&tuiModel, "model", "", "Filter to a model")
	tuiCmd.Flags().StringVar(&tuiSession, "session", "", "Filter to a session ID")
	tuiCmd.Flags().StringVar(&tuiTheme, "theme", "dark", "Theme: dark, light, high-contrast, colorblind")
}

func runTUI(cmd *cobra.Command, args []string) error {
	refreshInterval, err := time.ParseDuration(tuiRefresh)
	if err != nil {
		return err
	}

	// Refuse to run when stdout or stdin aren't a terminal. Without this
	// guard the tea.Program writes alt-screen escape sequences straight
	// to a pipe or file, which corrupts the downstream consumer and
	// leaves the user's terminal in a weird state.
	env := appui.DetectEnvironment()
	if !env.IsStdoutTTY || !env.IsStdinTTY {
		return fmt.Errorf("tok tui requires an interactive terminal; stdin/stdout must both be TTYs")
	}

	model := appui.NewModel(appui.Options{
		RefreshInterval: refreshInterval,
		Days:            tuiDays,
		ProjectPath:     tuiProject,
		AgentName:       tuiAgent,
		Provider:        tuiProvider,
		ModelName:       tuiModel,
		SessionID:       tuiSession,
		Theme:           appui.ThemeName(tuiTheme),
	})

	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = program.Run()
	return err
}
