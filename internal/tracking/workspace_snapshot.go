package tracking

import "github.com/lakshmanpatel/tok/internal/session"

// WorkspaceDashboardSnapshot joins token analytics with persisted session analytics.
type WorkspaceDashboardSnapshot struct {
	Dashboard   *DashboardSnapshot                `json:"dashboard"`
	DataQuality DashboardDataQuality              `json:"data_quality"`
	Sessions    *session.SessionAnalyticsSnapshot `json:"sessions"`
}

// GetWorkspaceDashboardSnapshot returns the canonical pre-TUI aggregate payload.
func GetWorkspaceDashboardSnapshot(
	tracker *Tracker,
	manager *session.SessionManager,
	dashboardOpts DashboardQueryOptions,
	sessionOpts session.SessionListOptions,
	snapshotLimit int,
) (*WorkspaceDashboardSnapshot, error) {
	dashboard, err := tracker.GetDashboardSnapshot(dashboardOpts)
	if err != nil {
		return nil, err
	}
	quality, err := tracker.GetDashboardDataQuality(dashboardOpts)
	if err != nil {
		return nil, err
	}
	sessionAnalytics, err := manager.GetAnalyticsSnapshot(sessionOpts, snapshotLimit)
	if err != nil {
		return nil, err
	}

	return &WorkspaceDashboardSnapshot{
		Dashboard:   dashboard,
		DataQuality: quality,
		Sessions:    sessionAnalytics,
	}, nil
}
