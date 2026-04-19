package core

import (
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/integrity"
	"github.com/lakshmanpatel/tok/internal/session"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show tok status",
	Long:  `Display tok status and configuration`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		out.Global().Println("tok: Enabled")
		out.Global().Printf("Project: %s\n", config.ProjectPath())
		out.Global().Printf("Config path: %s\n", config.ConfigPath())
		out.Global().Printf("Data path: %s\n", config.DataPath())
		out.Global().Printf("Tracking DB: %s\n", config.DatabasePath())

		if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
			out.Global().Println("Config: Not found (run 'tok config --create')")
		} else {
			out.Global().Println("Config: Found")
		}

		if result, err := integrity.VerifyHook(); err != nil {
			out.Global().Printf("Hook integrity: error (%v)\n", err)
		} else {
			out.Global().Printf("Hook integrity: %s\n", result.Status.String())
		}

		if manager, err := session.NewSessionManager(); err == nil {
			defer manager.Close()
			if summary, err := manager.GetSummary(); err == nil {
				out.Global().Printf("Sessions: %d total, %d active, %d snapshots\n", summary.TotalSessions, summary.ActiveSessions, summary.SnapshotCount)
			}
		}

		tracker, err := shared.OpenTracker()
		if err != nil {
			out.Global().Printf("Tracking: unavailable (%v)\n", err)
			return nil
		}
		defer tracker.Close()

		snapshot, err := tracker.GetDashboardSnapshot(tracking.DashboardQueryOptions{
			Days:               30,
			DailyTokenBudget:   100_000,
			WeeklyTokenBudget:  500_000,
			MonthlyTokenBudget: 2_000_000,
			ReductionGoalPct:   40,
		})
		if err != nil {
			out.Global().Printf("Tracking analytics: unavailable (%v)\n", err)
			return nil
		}

		out.Global().Println()
		out.Global().Println("Token Intelligence")
		out.Global().Printf("30d commands: %d\n", snapshot.Overview.TotalCommands)
		out.Global().Printf("30d saved: %d tokens (%.1f%%)\n", snapshot.Overview.TotalSavedTokens, snapshot.Overview.ReductionPct)
		out.Global().Printf("30d estimated savings: $%.4f\n", snapshot.Overview.EstimatedSavingsUSD)
		out.Global().Printf("Active days (30d): %d\n", snapshot.Lifecycle.ActiveDays30d)
		out.Global().Printf("Projects tracked: %d\n", snapshot.Lifecycle.ProjectsCount)
		out.Global().Printf("Avg saved / exec: %.1f tokens\n", snapshot.Lifecycle.AvgSavedTokensPerExec)
		out.Global().Printf("Savings streak: %d days\n", snapshot.Streaks.SavingsDays)
		out.Global().Printf("Goal streak: %d days @ %.0f%% reduction\n", snapshot.Streaks.GoalDays, snapshot.Streaks.GoalReductionPct)
		out.Global().Printf("Gamification: %d pts, level %d\n", snapshot.Gamification.Points, snapshot.Gamification.Level)
		if len(snapshot.TopProviders) > 0 {
			out.Global().Printf("Top provider: %s (%d saved)\n", snapshot.TopProviders[0].Key, snapshot.TopProviders[0].SavedTokens)
		}
		if len(snapshot.TopModels) > 0 {
			out.Global().Printf("Top model: %s (%d saved)\n", snapshot.TopModels[0].Key, snapshot.TopModels[0].SavedTokens)
		}
		if len(snapshot.TopProviderModels) > 0 {
			out.Global().Printf("Top provider/model: %s ($%.4f saved)\n", snapshot.TopProviderModels[0].Key, snapshot.TopProviderModels[0].EstimatedSavingsUSD)
		}
		if len(snapshot.LowSavingsCommands) > 0 {
			out.Global().Printf("Weakest command: %s (%.1f%% reduction)\n", snapshot.LowSavingsCommands[0].Key, snapshot.LowSavingsCommands[0].ReductionPct)
		}
		out.Global().Printf("Daily budget: %d / %d filtered tokens\n", snapshot.Budgets.Daily.FilteredTokens, snapshot.Budgets.Daily.TokenBudget)

		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(statusCmd) })
}
