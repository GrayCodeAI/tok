package tui

import (
	"strings"
	"testing"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

func TestRewardsSectionView(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.Gamification = tracking.DashboardGamification{
		Points: 1240, Level: 3, NextLevelPoints: 2000, Badges: []string{"early-bird", "clean-diff"},
	}
	s := newRewardsSection()
	view := s.View(ctx)
	for _, want := range []string{"Rewards", "CURRENT STREAK", "POINTS", "LEVEL", "Last "} {
		if !strings.Contains(view, want) {
			t.Fatalf("rewards view missing %q:\n%s", want, view)
		}
	}
}

func TestStreakBannerStages(t *testing.T) {
	cases := []struct {
		s    tracking.DashboardStreaks
		want string
	}{
		{tracking.DashboardStreaks{SavingsDays: 0, GoalDays: 7}, "Compress a few"},
		{tracking.DashboardStreaks{SavingsDays: 3, GoalDays: 7}, "4 more"},
		{tracking.DashboardStreaks{SavingsDays: 7, GoalDays: 7}, "Goal reached"},
		{tracking.DashboardStreaks{SavingsDays: 10, GoalDays: 7}, "past the goal"},
	}
	for _, tc := range cases {
		if got := streakBanner(tc.s); !strings.Contains(got, tc.want) {
			t.Errorf("streakBanner(%+v) = %q, want substring %q", tc.s, got, tc.want)
		}
	}
}

func TestProgressLabel(t *testing.T) {
	if got := progressLabel(0, 0); got != "no goal set" {
		t.Errorf("zero goal: %q", got)
	}
	if got := progressLabel(3, 7); !strings.Contains(got, "% of goal") {
		t.Errorf("partial progress: %q", got)
	}
	if got := progressLabel(7, 7); got != "goal hit — maintain" {
		t.Errorf("met goal: %q", got)
	}
	if got := progressLabel(10, 7); got != "goal hit — maintain" {
		t.Errorf("past goal: %q", got)
	}
}
