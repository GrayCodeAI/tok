package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// todaySection surfaces the current day's activity in one glance:
// headline metrics for the newest DailyTrends point, streak status,
// and a 7-day trailing sparkline so users can see today's bar in
// context.
type todaySection struct{}

func newTodaySection() *todaySection { return &todaySection{} }

func (s *todaySection) Name() string                      { return "Today" }
func (s *todaySection) Short() string                     { return "Easy Day" }
func (s *todaySection) Init(SectionContext) tea.Cmd       { return nil }
func (s *todaySection) KeyBindings() []key.Binding        { return nil }
func (s *todaySection) Update(_ SectionContext, _ tea.Msg) (SectionRenderer, tea.Cmd) {
	return s, nil
}

func (s *todaySection) View(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	snapshot := ctx.Data.Dashboard
	today := latestTrendPoint(snapshot.DailyTrends)
	yesterday := penultimateTrendPoint(snapshot.DailyTrends)

	columns := 3
	if width < 96 {
		columns = 2
	}
	if width < 64 {
		columns = 1
	}
	cards := []string{
		renderMetricCard(th, "Today saved", formatInt(today.SavedTokens),
			deltaLabel(today.SavedTokens, yesterday.SavedTokens), splitWidth(width, columns, 1), 0, th.ValuePositive),
		renderMetricCard(th, "Today reduction", fmt.Sprintf("%.1f%%", today.ReductionPct),
			deltaLabelFloat(today.ReductionPct, yesterday.ReductionPct), splitWidth(width, columns, 1), 1, th.ValueGold),
		renderMetricCard(th, "Today commands", formatInt(today.Commands),
			deltaLabel(today.Commands, yesterday.Commands), splitWidth(width, columns, 1), 2, th.ValueFocus),
		renderMetricCard(th, "Today cost saved", fmt.Sprintf("$%.4f", today.EstimatedSavingsUSD),
			"vs raw provider invoice", splitWidth(width, columns, 1), 3, th.ValuePositive),
		renderMetricCard(th, "Streak", fmt.Sprintf("%d days", snapshot.Streaks.SavingsDays),
			fmt.Sprintf("goal %d @ %.0f%%", snapshot.Streaks.GoalDays, snapshot.Streaks.GoalReductionPct),
			splitWidth(width, columns, 1), 4, th.ValueWarning),
		renderMetricCard(th, "Daily budget", fmt.Sprintf("%s / %s",
			formatInt(snapshot.Budgets.Daily.FilteredTokens), formatInt(snapshot.Budgets.Daily.TokenBudget)),
			fmt.Sprintf("%.1f%% used", snapshot.Budgets.Daily.TokenUtilizationPct),
			splitWidth(width, columns, 1), 5, th.ValueFocus),
	}
	cardGrid := renderCardGrid(cards, columns)

	trailing := trailingWindow(snapshot.DailyTrends, 7)
	sparkBlock := setWidth(panelStyle(th, 4), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Trailing 7 days"),
		"",
		th.CardLabel.Render("Saved") + "     " + th.ValuePositive.Render(sparklineSaved(trailing)) +
			"   " + th.CardMeta.Render(labelRange(trailing)),
		th.CardLabel.Render("Commands") + "  " + th.ValueFocus.Render(commandSparkline(trailing)) +
			"   " + th.CardMeta.Render(fmt.Sprintf("%d tracked total", sumCommands(trailing))),
	}, "\n"))

	bestLine := ""
	if snapshot.Streaks.BestDay != "" {
		bestLine = fmt.Sprintf("Best day: %s (%s saved · %.1f%%)",
			snapshot.Streaks.BestDay,
			formatInt(snapshot.Streaks.BestDaySavedTokens),
			snapshot.Streaks.BestDayReductionPct,
		)
	}
	notes := setWidth(panelStyle(th, 8), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Context"),
		renderHealthLine("Window", fmt.Sprintf("%d days", ctx.Opts.Days)),
		renderHealthLine("Active days (30d)", fmt.Sprintf("%d", snapshot.Lifecycle.ActiveDays30d)),
		renderHealthLine("Projects touched", fmt.Sprintf("%d", snapshot.Lifecycle.ProjectsCount)),
		renderHealthLine("Avg saved / exec", fmt.Sprintf("%.0f tokens", snapshot.Lifecycle.AvgSavedTokensPerExec)),
		th.Muted.Render(bestLine),
	}, "\n"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Today"),
		th.Subtitle.Render("Focused view of the newest daily bucket with deltas vs yesterday."),
		"",
		cardGrid,
		"",
		sparkBlock,
		"",
		notes,
	)
}

// --- small helpers ---

func latestTrendPoint(points []tracking.DashboardTrendPoint) tracking.DashboardTrendPoint {
	if len(points) == 0 {
		return tracking.DashboardTrendPoint{}
	}
	return points[len(points)-1]
}

func penultimateTrendPoint(points []tracking.DashboardTrendPoint) tracking.DashboardTrendPoint {
	if len(points) < 2 {
		return tracking.DashboardTrendPoint{}
	}
	return points[len(points)-2]
}

// trailingWindow returns the last n points, or all of them if fewer exist.
func trailingWindow(points []tracking.DashboardTrendPoint, n int) []tracking.DashboardTrendPoint {
	if n <= 0 || len(points) <= n {
		return points
	}
	return points[len(points)-n:]
}

// deltaLabel renders "+12% vs yesterday" or "first tracked day" when
// the prior bucket is empty. Using percentages keeps scale-invariant
// signal — raw deltas on a quiet day are noisy.
func deltaLabel(today, prior int64) string {
	if prior == 0 {
		if today == 0 {
			return "no activity"
		}
		return "first tracked day"
	}
	pct := (float64(today-prior) / float64(prior)) * 100
	sign := "+"
	if pct < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.1f%% vs yesterday", sign, pct)
}

func deltaLabelFloat(today, prior float64) string {
	if prior == 0 {
		if today == 0 {
			return "no reduction yet"
		}
		return "first tracked day"
	}
	delta := today - prior
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.1f pts vs yesterday", sign, delta)
}

func commandSparkline(points []tracking.DashboardTrendPoint) string {
	values := make([]int64, 0, len(points))
	for _, p := range points {
		values = append(values, p.Commands)
	}
	return sparkline(values)
}

func sumCommands(points []tracking.DashboardTrendPoint) int64 {
	var s int64
	for _, p := range points {
		s += p.Commands
	}
	return s
}

func labelRange(points []tracking.DashboardTrendPoint) string {
	if len(points) == 0 {
		return ""
	}
	return fmt.Sprintf("%s → %s", points[0].Period, points[len(points)-1].Period)
}
