package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// homeSection is the overview cockpit. Stateless for now — all layout
// decisions live in View() and are driven purely by the SectionContext.
type homeSection struct{}

func newHomeSection() *homeSection { return &homeSection{} }

func (h *homeSection) Name() string                { return "Home" }
func (h *homeSection) Short() string               { return "Overview" }
func (h *homeSection) Init(SectionContext) tea.Cmd { return nil }
func (h *homeSection) KeyBindings() []key.Binding  { return nil }
func (h *homeSection) Update(_ SectionContext, _ tea.Msg) (SectionRenderer, tea.Cmd) {
	return h, nil
}

func (h *homeSection) View(ctx SectionContext) string {
	return renderHomeView(ctx)
}

// renderHomeView is the raw rendering logic, factored out so tests can
// call it with a fabricated SectionContext without going through the
// whole tea.Program.
func renderHomeView(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}

	snapshot := ctx.Data.Dashboard
	overview := snapshot.Overview
	store := ctx.Data.Sessions.StoreSummary
	quality := ctx.Data.DataQuality

	metricColumns := 3
	if width < 96 {
		metricColumns = 2
	}
	if width < 64 {
		metricColumns = 1
	}
	cards := []string{
		renderMetricCard(th, "Saved Tokens", formatInt(overview.TotalSavedTokens), fmt.Sprintf("%d day window", ctx.Opts.Days), splitWidth(width, metricColumns, 1), 0, th.ValuePositive),
		renderMetricCard(th, "Cost Saved", fmt.Sprintf("$%.4f", overview.EstimatedSavingsUSD), "estimated reduction value", splitWidth(width, metricColumns, 1), 1, th.ValueFocus),
		renderMetricCard(th, "Reduction", fmt.Sprintf("%.1f%%", overview.ReductionPct), "overall compression rate", splitWidth(width, metricColumns, 1), 2, th.ValueGold),
		renderMetricCard(th, "Commands", formatInt(overview.TotalCommands), "tracked commands", splitWidth(width, metricColumns, 1), 3, th.Title),
		renderMetricCard(th, "Active Days", fmt.Sprintf("%d / %d", snapshot.Lifecycle.ActiveDays30d, ctx.Opts.Days), "days with tracked activity", splitWidth(width, metricColumns, 1), 4, th.ValuePositive),
		renderMetricCard(th, "Current Streak", fmt.Sprintf("%d days", snapshot.Streaks.SavingsDays), fmt.Sprintf("%d pts · level %d", snapshot.Gamification.Points, snapshot.Gamification.Level), splitWidth(width, metricColumns, 1), 5, th.ValueWarning),
	}
	cardGrid := renderCardGrid(cards, metricColumns)

	dailySpark := sparklineSavedFor(snapshot.DailyTrends, ctx.Env.UTF8)
	weeklySpark := sparklineSavedFor(snapshot.WeeklyTrends, ctx.Env.UTF8)
	trendsBlock := setWidth(panelStyle(th, 8), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Activity & Trends"),
		"",
		th.CardLabel.Render("Daily sparkline") + "  " + th.ValuePositive.Render(dailySpark) + "  " + th.CardMeta.Render(fmt.Sprintf("%d points", len(snapshot.DailyTrends))),
		th.CardLabel.Render("Weekly sparkline") + " " + th.ValueFocus.Render(weeklySpark) + "  " + th.CardMeta.Render(fmt.Sprintf("%d points", len(snapshot.WeeklyTrends))),
		th.CardLabel.Render("Budget") + "  " + th.ValueWarning.Render(formatInt(snapshot.Budgets.Daily.FilteredTokens)+" / "+formatInt(snapshot.Budgets.Daily.TokenBudget)) + "  " + th.CardMeta.Render("daily filtered tokens"),
	}, "\n"))

	leaderboards := ""
	if width >= 100 {
		leftWidth := splitWidth(width, 2, 1)
		rightWidth := width - leftWidth - 1
		leaderboards = joinHorizontalGap(
			" ",
			renderBreakdownPanel(th, "Top Providers", "Provider", snapshot.TopProviders, leftWidth, 9),
			renderBreakdownPanel(th, "Weak Commands", "Command", snapshot.LowSavingsCommands, rightWidth, 10),
		)
	} else {
		leaderboards = lipgloss.JoinVertical(
			lipgloss.Left,
			renderBreakdownPanel(th, "Top Providers", "Provider", snapshot.TopProviders, width, 9),
			"",
			renderBreakdownPanel(th, "Weak Commands", "Command", snapshot.LowSavingsCommands, width, 10),
		)
	}

	healthLines := []string{
		th.PanelTitle.Render("Health"),
		renderHealthLine("Attribution gaps", fmt.Sprintf("%d agent, %d provider, %d model, %d session",
			quality.CommandsMissingAgent,
			quality.CommandsMissingProvider,
			quality.CommandsMissingModel,
			quality.CommandsMissingSession,
		)),
		renderHealthLine("Pricing coverage", fmt.Sprintf("%.1f%%", quality.PricingCoverage.CoveragePct())),
		renderHealthLine("Parse failures", fmt.Sprintf("%d", quality.ParseFailures)),
	}
	if store.TopAgent != "" {
		healthLines = append(healthLines, renderHealthLine("Top session agent", displayKey(store.TopAgent)))
	}
	insightLines := []string{
		th.PanelTitle.Render("Snapshot"),
		renderHealthLine("Pricing coverage", fmt.Sprintf("%.1f%% explicit pricing", quality.PricingCoverage.CoveragePct())),
	}
	if weak := firstBreakdown(snapshot.LowSavingsCommands); weak != nil {
		insightLines = append(insightLines, renderHealthLine("Weakest command", fmt.Sprintf("%s at %.1f%%", displayKey(weak.Key), weak.ReductionPct)))
	}
	if provider := firstBreakdown(snapshot.TopProviders); provider != nil {
		insightLines = append(insightLines, renderHealthLine("Top provider", fmt.Sprintf("%s saved %s", displayKey(provider.Key), formatInt(provider.SavedTokens))))
	}

	healthBlock := ""
	if width >= 100 {
		leftWidth := splitWidth(width, 2, 1)
		rightWidth := width - leftWidth - 1
		healthBlock = joinHorizontalGap(
			" ",
			setWidth(panelStyle(th, 11), leftWidth).Render(strings.Join(healthLines, "\n")),
			setWidth(panelStyle(th, 12), rightWidth).Render(strings.Join(insightLines, "\n")),
		)
	} else {
		healthBlock = lipgloss.JoinVertical(
			lipgloss.Left,
			setWidth(panelStyle(th, 11), width).Render(strings.Join(healthLines, "\n")),
			"",
			setWidth(panelStyle(th, 12), width).Render(strings.Join(insightLines, "\n")),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Home"),
		th.Subtitle.Render("Token intelligence cockpit with live savings, costs, attribution, and quality telemetry."),
		"",
		cardGrid,
		"",
		trendsBlock,
		"",
		leaderboards,
		"",
		healthBlock,
	)
}

// --- shared rendering primitives (used by Home now, other sections later) ---

func renderMetricCard(th theme, title, value, detail string, width, accentIndex int, valueStyle lipgloss.Style) string {
	return setWidth(accentCardStyle(th, accentIndex), width).Render(strings.Join([]string{
		th.CardLabel.Render(strings.ToUpper(title)),
		valueStyle.Render(value),
		th.CardMeta.Render(detail),
	}, "\n"))
}

func renderCardGrid(cards []string, columns int) string {
	if len(cards) == 0 {
		return ""
	}
	if columns <= 1 {
		return lipgloss.JoinVertical(lipgloss.Left, cards...)
	}
	rows := make([]string, 0, (len(cards)+columns-1)/columns)
	for i := 0; i < len(cards); i += columns {
		end := min(i+columns, len(cards))
		rows = append(rows, joinHorizontalGap(" ", cards[i:end]...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderBreakdownPanel(th theme, title, keyLabel string, items []tracking.DashboardBreakdown, width, accentIndex int) string {
	lines := []string{
		th.PanelTitle.Render(title),
		th.CardMeta.Render(fmt.Sprintf("%-18s %9s %7s  %s", keyLabel, "Saved", "Rate", "Share")),
	}
	if len(items) == 0 {
		lines = append(lines, th.Muted.Render("No data"))
		return setWidth(panelStyle(th, accentIndex), width).Render(strings.Join(lines, "\n"))
	}

	maxSaved := int64(0)
	for _, item := range items {
		if item.SavedTokens > maxSaved {
			maxSaved = item.SavedTokens
		}
	}

	maxItems := min(len(items), 5)
	for i := 0; i < maxItems; i++ {
		item := items[i]
		lines = append(lines, renderBreakdownEntry(th, item, width, i, maxSaved))
	}

	return setWidth(panelStyle(th, accentIndex), width).Render(strings.Join(lines, "\n"))
}

func renderBreakdownEntry(th theme, item tracking.DashboardBreakdown, width, index int, maxSaved int64) string {
	keyWidth := max(12, width/2-6)
	barWidth := max(8, width-keyWidth-22)
	keyStr := truncate(displayKey(item.Key), keyWidth)
	head := fmt.Sprintf("%-*s %9s %6.1f%%", keyWidth, keyStr, formatInt(item.SavedTokens), item.ReductionPct)
	bar := renderBar(th, item.SavedTokens, maxSaved, barWidth, index)
	style := th.TableRow
	if index == 0 {
		style = th.TableRowAccent
	}
	share := 0.0
	if maxSaved > 0 {
		share = (float64(item.SavedTokens) / float64(maxSaved)) * 100
	}
	barLine := bar + " " + th.CardMeta.Render(fmt.Sprintf("%5.1f%%", share))
	return style.Render(head) + "\n" + barLine
}

func renderBar(th theme, value, maxValue int64, width, index int) string {
	return renderBarGlyphs(th, value, maxValue, width, index, true)
}

// renderBarGlyphs is the utf8-aware form. Unicode block glyphs degrade
// to '#' / '-' when the terminal can't render them. Both glyphs still
// compose visually (solid mass on the left, shaded region on the right)
// so the bar reads correctly.
func renderBarGlyphs(th theme, value, maxValue int64, width, index int, utf8 bool) string {
	if width <= 0 {
		return ""
	}
	filled := 0
	if maxValue > 0 {
		filled = int(math.Round((float64(value) / float64(maxValue)) * float64(width)))
	}
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	solidGlyph := "█"
	emptyGlyph := "░"
	if !utf8 {
		solidGlyph = "#"
		emptyGlyph = "-"
	}
	color := th.AccentColors[index%len(th.AccentColors)]
	filledStyle := lipgloss.NewStyle().Foreground(color)
	return filledStyle.Render(strings.Repeat(solidGlyph, filled)) +
		th.BarEmpty.Render(strings.Repeat(emptyGlyph, width-filled))
}

func accentCardStyle(th theme, index int) lipgloss.Style {
	base := th.Card
	if len(th.AccentColors) == 0 {
		return base
	}
	color := th.AccentColors[index%len(th.AccentColors)]
	return base.BorderForeground(color)
}

func panelStyle(th theme, index int) lipgloss.Style {
	base := th.Panel
	if len(th.AccentColors) == 0 {
		return base
	}
	color := th.AccentColors[index%len(th.AccentColors)]
	return base.BorderForeground(color)
}

// --- theme-agnostic helpers ------------------------------------------------

func sparklineSaved(points []tracking.DashboardTrendPoint) string {
	return sparklineSavedFor(points, true)
}

func sparklineSavedFor(points []tracking.DashboardTrendPoint, utf8 bool) string {
	if len(points) == 0 {
		return "no trend data"
	}
	values := make([]int64, 0, len(points))
	for _, point := range points {
		values = append(values, point.SavedTokens)
	}
	return sparklineGlyphs(values, utf8)
}

func sparkline(values []int64) string {
	return sparklineGlyphs(values, true)
}

// sparklineGlyphs renders a block-sparkline when utf8=true, falling back
// to an ASCII bucketed substitute (".-=#") when the terminal can't
// display the Unicode block glyphs. The ASCII bucket is coarser but
// still conveys trend shape — good enough when the alternative is
// garbage.
func sparklineGlyphs(values []int64, utf8 bool) string {
	if len(values) == 0 {
		return ""
	}
	glyphs := []rune("▁▂▃▄▅▆▇█")
	if !utf8 {
		glyphs = []rune(".-=#")
	}
	var maxValue int64
	for _, v := range values {
		if v > maxValue {
			maxValue = v
		}
	}
	if maxValue <= 0 {
		return strings.Repeat(string(glyphs[0]), len(values))
	}
	var b strings.Builder
	for _, v := range values {
		idx := int((float64(v) / float64(maxValue)) * float64(len(glyphs)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(glyphs) {
			idx = len(glyphs) - 1
		}
		b.WriteRune(glyphs[idx])
	}
	return b.String()
}

func formatInt(v int64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fm", float64(v)/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.1fk", float64(v)/1_000)
	default:
		return fmt.Sprintf("%d", v)
	}
}

func truncate(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	if width <= 1 {
		return s[:width]
	}
	return s[:width-1] + "…"
}

func splitWidth(total, parts, gap int) int {
	if parts <= 0 {
		return total
	}
	return max(8, (total-((parts-1)*gap))/parts)
}

func renderHealthLine(label, value string) string {
	return "• " + label + ": " + value
}

func displayKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "(unknown)" {
		return "Unattributed"
	}
	return s
}
