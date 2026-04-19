package core

import (
	"fmt"
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
		fmt.Println("tok: Enabled")
		fmt.Printf("Project: %s\n", config.ProjectPath())
		fmt.Printf("Config path: %s\n", config.ConfigPath())
		fmt.Printf("Data path: %s\n", config.DataPath())
		fmt.Printf("Tracking DB: %s\n", config.DatabasePath())

		if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
			fmt.Println("Config: Not found (run 'tok config --create')")
		} else {
			fmt.Println("Config: Found")
		}

		if result, err := integrity.VerifyHook(); err != nil {
			fmt.Printf("Hook integrity: error (%v)\n", err)
		} else {
			fmt.Printf("Hook integrity: %s\n", result.Status.String())
		}

		if manager, err := session.NewSessionManager(); err == nil {
			defer manager.Close()
			if summary, err := manager.GetSummary(); err == nil {
				fmt.Printf("Sessions: %d total, %d active, %d snapshots\n", summary.TotalSessions, summary.ActiveSessions, summary.SnapshotCount)
			}
		}

		tracker, err := shared.OpenTracker()
		if err != nil {
			fmt.Printf("Tracking: unavailable (%v)\n", err)
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
			fmt.Printf("Tracking analytics: unavailable (%v)\n", err)
			return nil
		}

		fmt.Println()
		fmt.Println("Token Intelligence")
		fmt.Printf("30d commands: %d\n", snapshot.Overview.TotalCommands)
		fmt.Printf("30d saved: %d tokens (%.1f%%)\n", snapshot.Overview.TotalSavedTokens, snapshot.Overview.ReductionPct)
		fmt.Printf("30d estimated savings: $%.4f\n", snapshot.Overview.EstimatedSavingsUSD)
		fmt.Printf("Active days (30d): %d\n", snapshot.Lifecycle.ActiveDays30d)
		fmt.Printf("Projects tracked: %d\n", snapshot.Lifecycle.ProjectsCount)
		fmt.Printf("Avg saved / exec: %.1f tokens\n", snapshot.Lifecycle.AvgSavedTokensPerExec)
		fmt.Printf("Savings streak: %d days\n", snapshot.Streaks.SavingsDays)
		fmt.Printf("Goal streak: %d days @ %.0f%% reduction\n", snapshot.Streaks.GoalDays, snapshot.Streaks.GoalReductionPct)
		fmt.Printf("Gamification: %d pts, level %d\n", snapshot.Gamification.Points, snapshot.Gamification.Level)
		if len(snapshot.TopProviders) > 0 {
			fmt.Printf("Top provider: %s (%d saved)\n", snapshot.TopProviders[0].Key, snapshot.TopProviders[0].SavedTokens)
		}
		if len(snapshot.TopModels) > 0 {
			fmt.Printf("Top model: %s (%d saved)\n", snapshot.TopModels[0].Key, snapshot.TopModels[0].SavedTokens)
		}
		if len(snapshot.TopProviderModels) > 0 {
			fmt.Printf("Top provider/model: %s ($%.4f saved)\n", snapshot.TopProviderModels[0].Key, snapshot.TopProviderModels[0].EstimatedSavingsUSD)
		}
		if len(snapshot.LowSavingsCommands) > 0 {
			fmt.Printf("Weakest command: %s (%.1f%% reduction)\n", snapshot.LowSavingsCommands[0].Key, snapshot.LowSavingsCommands[0].ReductionPct)
		}
		fmt.Printf("Daily budget: %d / %d filtered tokens\n", snapshot.Budgets.Daily.FilteredTokens, snapshot.Budgets.Daily.TokenBudget)

		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(statusCmd) })
}
