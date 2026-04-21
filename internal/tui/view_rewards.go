package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

// rewardsSection surfaces gamification: streak status vs goal, points,
// level, badges, plus a calendar strip of the last N days colored by
// activity intensity. The calendar tells users at a glance whether the
// streak is about to break.
type rewardsSection struct{}

func newRewardsSection() *rewardsSection { return &rewardsSection{} }

func (s *rewardsSection) Name() string                { return "Rewards" }
func (s *rewardsSection) Short() string               { return "Streaks" }
func (s *rewardsSection) Init(SectionContext) tea.Cmd { return nil }
func (s *rewardsSection) KeyBindings() []key.Binding  { return nil }
func (s *rewardsSection) IsScrollable() bool          { return true }
func (s *rewardsSection) Update(_ SectionContext, _ tea.Msg) (SectionRenderer, tea.Cmd) {
	return s, nil
}

func (s *rewardsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	snapshot := ctx.Data.Dashboard
	streaks := snapshot.Streaks
	gam := snapshot.Gamification

	columns := 3
	if width < 96 {
		columns = 2
	}
	if width < 64 {
		columns = 1
	}

	progressBadges := strings.Join(gam.Badges, ", ")
	if progressBadges == "" {
		progressBadges = "none yet"
	}

	cards := []string{
		renderMetricCard(th, "Current streak", fmt.Sprintf("%d days", streaks.SavingsDays),
			fmt.Sprintf("goal %d @ %.0f%% reduction", streaks.GoalDays, streaks.GoalReductionPct),
			splitWidth(width, columns, 1), 0, th.ValueWarning),
		renderMetricCard(th, "Points", fmt.Sprintf("%d", gam.Points),
			fmt.Sprintf("next level @ %d pts", gam.NextLevelPoints),
			splitWidth(width, columns, 1), 1, th.ValueGold),
		renderMetricCard(th, "Level", fmt.Sprintf("%d", gam.Level),
			fmt.Sprintf("%d badges earned", len(gam.Badges)),
			splitWidth(width, columns, 1), 2, th.ValueFocus),
		renderMetricCard(th, "Best day", fallback(streaks.BestDay, "—"),
			fmt.Sprintf("%s saved · %.1f%% reduction",
				formatInt(streaks.BestDaySavedTokens), streaks.BestDayReductionPct),
			splitWidth(width, columns, 1), 3, th.ValuePositive),
		renderMetricCard(th, "Goal streak", fmt.Sprintf("%d / %d", streaks.SavingsDays, streaks.GoalDays),
			progressLabel(streaks.SavingsDays, streaks.GoalDays),
			splitWidth(width, columns, 1), 4, th.ValuePositive),
		renderMetricCard(th, "Badges", progressBadges,
			"earned achievements",
			splitWidth(width, columns, 1), 5, th.ValueFocus),
	}
	cardGrid := renderCardGrid(cards, columns)

	calendar := renderStreakCalendar(th, snapshot.DailyTrends, width, ctx.Env.UTF8)

	bestLineStyle := th.Muted
	if streaks.SavingsDays >= streaks.GoalDays {
		bestLineStyle = th.Positive
	}
	footer := setWidth(panelStyle(th, 6), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Motivation"),
		bestLineStyle.Render(streakBanner(streaks)),
		th.Muted.Render("Points accrue as daily reduction holds above the goal."),
	}, "\n"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Rewards"),
		th.Subtitle.Render("Streaks, points, and daily momentum at a glance."),
		"",
		cardGrid,
		"",
		calendar,
		"",
		footer,
	)
}

// --- helpers ---

// renderStreakCalendar draws a row of colored cells, one per DailyTrends
// point, where the fill color tracks reduction-pct thresholds. This is
// the TUI equivalent of a GitHub contribution heatmap.
func renderStreakCalendar(th theme, points []tracking.DashboardTrendPoint, width int, utf8 bool) string {
	if len(points) == 0 {
		return th.Muted.Render("No activity yet.")
	}

	cells := make([]string, 0, len(points))
	for _, p := range points {
		cells = append(cells, streakCell(th, p, utf8))
	}
	// Row wraps when the rendered width would exceed the pane width.
	// Each cell is 2 runes wide to keep a square-ish aspect in terminal.
	rows := wrapCells(cells, max(10, (width-4)/2))

	legend := "legend: ░ none  ▒ low  ▓ mid  █ goal+"
	if !utf8 {
		legend = "legend: . none  : low  + mid  # goal+"
	}
	lines := []string{th.PanelTitle.Render("Last " + fmt.Sprintf("%d", len(points)) + " days")}
	lines = append(lines, rows...)
	lines = append(lines, th.CardMeta.Render(legend))
	return setWidth(panelStyle(th, 5), width).Render(strings.Join(lines, "\n"))
}

func streakCell(th theme, p tracking.DashboardTrendPoint, utf8 bool) string {
	var glyph string
	var color lipgloss.Color
	switch {
	case p.SavedTokens == 0:
		glyph = pickGlyph(utf8, "░ ", ". ")
		color = lipgloss.Color("#2B3442")
	case p.ReductionPct < 20:
		glyph = pickGlyph(utf8, "▒ ", ": ")
		color = lipgloss.Color("#7AB8FF")
	case p.ReductionPct < 40:
		glyph = pickGlyph(utf8, "▓ ", "+ ")
		color = lipgloss.Color("#53D18D")
	default:
		glyph = pickGlyph(utf8, "█ ", "# ")
		color = lipgloss.Color("#F2CC70")
	}
	_ = th
	return lipgloss.NewStyle().Foreground(color).Render(glyph)
}

func pickGlyph(utf8 bool, unicode, ascii string) string {
	if utf8 {
		return unicode
	}
	return ascii
}

func wrapCells(cells []string, perRow int) []string {
	if perRow <= 0 {
		return []string{strings.Join(cells, "")}
	}
	rows := make([]string, 0, (len(cells)+perRow-1)/perRow)
	for i := 0; i < len(cells); i += perRow {
		end := i + perRow
		if end > len(cells) {
			end = len(cells)
		}
		rows = append(rows, strings.Join(cells[i:end], ""))
	}
	return rows
}

func progressLabel(current, goal int) string {
	if goal <= 0 {
		return "no goal set"
	}
	pct := (float64(current) / float64(goal)) * 100
	if pct >= 100 {
		return "goal hit — maintain"
	}
	return fmt.Sprintf("%.0f%% of goal", pct)
}

func streakBanner(s tracking.DashboardStreaks) string {
	switch {
	case s.SavingsDays == 0:
		return "Compress a few commands today to start a streak."
	case s.SavingsDays < s.GoalDays:
		remaining := s.GoalDays - s.SavingsDays
		return fmt.Sprintf("%d more day(s) to hit the goal streak.", remaining)
	case s.SavingsDays == s.GoalDays:
		return "Goal reached — hold the line to keep compounding."
	default:
		return fmt.Sprintf("%d days past the goal — building a compound habit.",
			s.SavingsDays-s.GoalDays)
	}
}

func fallback(s, alt string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return alt
	}
	return s
}
