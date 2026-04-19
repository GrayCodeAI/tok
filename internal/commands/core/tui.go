package core

import (
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
}

func runTUI(cmd *cobra.Command, args []string) error {
	refreshInterval, err := time.ParseDuration(tuiRefresh)
	if err != nil {
		return err
	}

	model := appui.NewModel(appui.Options{
		RefreshInterval: refreshInterval,
		Days:            tuiDays,
		ProjectPath:     tuiProject,
		AgentName:       tuiAgent,
		Provider:        tuiProvider,
		ModelName:       tuiModel,
		SessionID:       tuiSession,
	})

	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}
