package tui

import (
	"time"

	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/session"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

type Options struct {
	RefreshInterval time.Duration
	Days            int
	ProjectPath     string
	AgentName       string
	Provider        string
	ModelName       string
	SessionID       string
}

func (o Options) normalized() Options {
	if o.RefreshInterval <= 0 {
		o.RefreshInterval = 20 * time.Second
	}
	if o.Days <= 0 {
		o.Days = 30
	}
	return o
}

type snapshotLoader interface {
	Load(Options) (*tracking.WorkspaceDashboardSnapshot, error)
}

type workspaceLoader struct{}

func (workspaceLoader) Load(opts Options) (*tracking.WorkspaceDashboardSnapshot, error) {
	opts = opts.normalized()

	tracker, err := shared.OpenTracker()
	if err != nil {
		return nil, err
	}
	defer tracker.Close()

	manager, err := session.NewSessionManager()
	if err != nil {
		return nil, err
	}
	defer manager.Close()

	return tracking.GetWorkspaceDashboardSnapshot(
		tracker,
		manager,
		tracking.DashboardQueryOptions{
			Days:                 opts.Days,
			ProjectPath:          opts.ProjectPath,
			AgentName:            opts.AgentName,
			Provider:             opts.Provider,
			ModelName:            opts.ModelName,
			SessionID:            opts.SessionID,
			Limit:                8,
			ReductionGoalPct:     40,
			DailyTokenBudget:     100_000,
			WeeklyTokenBudget:    500_000,
			MonthlyTokenBudget:   2_000_000,
			DailyCostBudgetUSD:   5,
			WeeklyCostBudgetUSD:  25,
			MonthlyCostBudgetUSD: 100,
		},
		session.SessionListOptions{
			Agent:       opts.AgentName,
			ProjectPath: opts.ProjectPath,
			Limit:       10,
		},
		10,
	)
}
