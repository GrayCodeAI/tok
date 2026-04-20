package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// trendsSection renders the daily/weekly trend series as Braille line
// charts plus a headline delta vs the prior equivalent window.
//
// The chart is intentionally axis-free — the caption below it gives the
// numeric range. This keeps the Braille glyph density as the only thing
// the user has to read.
type trendsSection struct {
	granularity trendGranularity
}

type trendGranularity int

const (
	trendDaily trendGranularity = iota
	trendWeekly
)

func newTrendsSection() *trendsSection { return &trendsSection{granularity: trendDaily} }

func (s *trendsSection) Name() string  { return "Trends" }
func (s *trendsSection) Short() string { return "Analytics" }
func (s *trendsSection) Init(SectionContext) tea.Cmd { return nil }

func (s *trendsSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "daily granularity")),
		key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "weekly granularity")),
	}
}

func (s *trendsSection) Update(_ SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "d":
			s.granularity = trendDaily
		case "w":
			s.granularity = trendWeekly
		}
	}
	return s, nil
}

func (s *trendsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	snapshot := ctx.Data.Dashboard

	points := snapshot.DailyTrends
	label := "daily"
	if s.granularity == trendWeekly {
		points = snapshot.WeeklyTrends
		label = "weekly"
	}
	if len(points) == 0 {
		return th.Muted.Render(fmt.Sprintf("No %s trend data yet.", label))
	}

	saved := trendValues(points, func(p tracking.DashboardTrendPoint) float64 { return float64(p.SavedTokens) })
	commands := trendValues(points, func(p tracking.DashboardTrendPoint) float64 { return float64(p.Commands) })
	reduction := trendValues(points, func(p tracking.DashboardTrendPoint) float64 { return p.ReductionPct })

	chartWidth := max(24, width-6)
	chartHeight := 6
	if ctx.Height < 24 {
		chartHeight = 3
	}

	savedChart := LineChart(saved, chartWidth, chartHeight, ctx.Env.UTF8)
	commandsChart := LineChart(commands, chartWidth, chartHeight, ctx.Env.UTF8)
	reductionChart := LineChart(reduction, chartWidth, chartHeight, ctx.Env.UTF8)

	head := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Trends"),
		th.Subtitle.Render(fmt.Sprintf("%s granularity · %d buckets · press d / w to switch",
			label, len(points))),
	)

	savedPanel := renderTrendPanel(th, "Saved tokens", th.ValuePositive.Render(savedChart),
		fmt.Sprintf("range %s – %s · total %s",
			formatInt(int64(chartMin(saved))), formatInt(int64(chartMax(saved))),
			formatInt(sumValues(saved))), width, 2)

	reductionPanel := renderTrendPanel(th, "Reduction %", th.ValueGold.Render(reductionChart),
		fmt.Sprintf("range %.1f%% – %.1f%% · avg %.1f%%",
			chartMin(reduction), chartMax(reduction), avgValues(reduction)), width, 4)

	commandsPanel := renderTrendPanel(th, "Commands", th.ValueFocus.Render(commandsChart),
		fmt.Sprintf("range %d – %d · total %d",
			int64(chartMin(commands)), int64(chartMax(commands)), sumValues(commands)),
		width, 6)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		head,
		"",
		savedPanel,
		"",
		reductionPanel,
		"",
		commandsPanel,
	)
}

func renderTrendPanel(th theme, title, chart, caption string, width, accentIdx int) string {
	return setWidth(panelStyle(th, accentIdx), width).Render(strings.Join([]string{
		th.PanelTitle.Render(title),
		"",
		chart,
		"",
		th.CardMeta.Render(caption),
	}, "\n"))
}

func trendValues(points []tracking.DashboardTrendPoint, pick func(tracking.DashboardTrendPoint) float64) []float64 {
	out := make([]float64, len(points))
	for i, p := range points {
		out[i] = pick(p)
	}
	return out
}

func chartMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func chartMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func sumValues(values []float64) int64 {
	var s float64
	for _, v := range values {
		s += v
	}
	return int64(s)
}

func avgValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var s float64
	for _, v := range values {
		s += v
	}
	return s / float64(len(values))
}
