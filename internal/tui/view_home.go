package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func (m model) renderHome(width int) string {
	if m.data == nil || m.data.Dashboard == nil {
		return m.theme.Muted.Render("No dashboard data yet.")
	}

	snapshot := m.data.Dashboard
	overview := snapshot.Overview
	store := m.data.Sessions.StoreSummary
	quality := m.data.DataQuality

	metricColumns := 3
	if width < 96 {
		metricColumns = 2
	}
	if width < 64 {
		metricColumns = 1
	}
	cards := []string{
		m.renderMetricCard("Saved Tokens", formatInt(overview.TotalSavedTokens), fmt.Sprintf("%d day window", m.opts.Days), splitWidth(width, metricColumns, 1), 0, m.theme.ValuePositive),
		m.renderMetricCard("Cost Saved", fmt.Sprintf("$%.4f", overview.EstimatedSavingsUSD), "estimated reduction value", splitWidth(width, metricColumns, 1), 1, m.theme.ValueFocus),
		m.renderMetricCard("Reduction", fmt.Sprintf("%.1f%%", overview.ReductionPct), "overall compression rate", splitWidth(width, metricColumns, 1), 2, m.theme.ValueGold),
		m.renderMetricCard("Commands", formatInt(overview.TotalCommands), "tracked commands", splitWidth(width, metricColumns, 1), 3, m.theme.Title),
		m.renderMetricCard("Active Days", fmt.Sprintf("%d / %d", snapshot.Lifecycle.ActiveDays30d, m.opts.Days), "days with tracked activity", splitWidth(width, metricColumns, 1), 4, m.theme.ValuePositive),
		m.renderMetricCard("Current Streak", fmt.Sprintf("%d days", snapshot.Streaks.SavingsDays), fmt.Sprintf("%d pts · level %d", snapshot.Gamification.Points, snapshot.Gamification.Level), splitWidth(width, metricColumns, 1), 5, m.theme.ValueWarning),
	}
	cardGrid := m.renderCardGrid(cards, metricColumns)

	dailySpark := sparklineSaved(snapshot.DailyTrends)
	weeklySpark := sparklineSaved(snapshot.WeeklyTrends)
	trendsBlock := setWidth(m.panelStyle(8), width).Render(strings.Join([]string{
		m.theme.PanelTitle.Render("Activity & Trends"),
		"",
		m.theme.CardLabel.Render("Daily sparkline") + "  " + m.theme.ValuePositive.Render(dailySpark) + "  " + m.theme.CardMeta.Render(fmt.Sprintf("%d points", len(snapshot.DailyTrends))),
		m.theme.CardLabel.Render("Weekly sparkline") + " " + m.theme.ValueFocus.Render(weeklySpark) + "  " + m.theme.CardMeta.Render(fmt.Sprintf("%d points", len(snapshot.WeeklyTrends))),
		m.theme.CardLabel.Render("Budget") + "  " + m.theme.ValueWarning.Render(formatInt(snapshot.Budgets.Daily.FilteredTokens)+" / "+formatInt(snapshot.Budgets.Daily.TokenBudget)) + "  " + m.theme.CardMeta.Render("daily filtered tokens"),
	}, "\n"))

	leaderboards := ""
	if width >= 100 {
		leftWidth := splitWidth(width, 2, 1)
		rightWidth := width - leftWidth - 1
		leaderboards = joinHorizontalGap(
			" ",
			m.renderBreakdownPanel("Top Providers", "Provider", snapshot.TopProviders, leftWidth, 9),
			m.renderBreakdownPanel("Weak Commands", "Command", snapshot.LowSavingsCommands, rightWidth, 10),
		)
	} else {
		leaderboards = lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderBreakdownPanel("Top Providers", "Provider", snapshot.TopProviders, width, 9),
			"",
			m.renderBreakdownPanel("Weak Commands", "Command", snapshot.LowSavingsCommands, width, 10),
		)
	}

	healthLines := []string{
		m.theme.PanelTitle.Render("Health"),
		renderHealthLine("Attribution gaps", fmt.Sprintf("%d agent, %d provider, %d model, %d session",
			quality.CommandsMissingAgent,
			quality.CommandsMissingProvider,
			quality.CommandsMissingModel,
			quality.CommandsMissingSession,
		), quality.CommandsMissingAgent > 0 || quality.CommandsMissingProvider > 0 || quality.CommandsMissingModel > 0 || quality.CommandsMissingSession > 0),
		renderHealthLine("Pricing coverage", fmt.Sprintf("%.1f%%", quality.PricingCoverage.CoveragePct()), quality.PricingCoverage.FallbackPricingCommands > 0),
		renderHealthLine("Parse failures", fmt.Sprintf("%d", quality.ParseFailures), quality.ParseFailures > 0),
	}
	if store.TopAgent != "" {
		healthLines = append(healthLines, renderHealthLine("Top session agent", displayKey(store.TopAgent), false))
	}
	insightLines := []string{
		m.theme.PanelTitle.Render("Snapshot"),
		renderHealthLine("Pricing coverage", fmt.Sprintf("%.1f%% explicit pricing", quality.PricingCoverage.CoveragePct()), quality.PricingCoverage.FallbackPricingCommands > 0),
	}
	if weak := firstBreakdown(snapshot.LowSavingsCommands); weak != nil {
		insightLines = append(insightLines, renderHealthLine("Weakest command", fmt.Sprintf("%s at %.1f%%", displayKey(weak.Key), weak.ReductionPct), weak.ReductionPct < 10))
	}
	if provider := firstBreakdown(snapshot.TopProviders); provider != nil {
		insightLines = append(insightLines, renderHealthLine("Top provider", fmt.Sprintf("%s saved %s", displayKey(provider.Key), formatInt(provider.SavedTokens)), false))
	}

	healthBlock := ""
	if width >= 100 {
		leftWidth := splitWidth(width, 2, 1)
		rightWidth := width - leftWidth - 1
		healthBlock = joinHorizontalGap(
			" ",
			setWidth(m.panelStyle(11), leftWidth).Render(strings.Join(healthLines, "\n")),
			setWidth(m.panelStyle(12), rightWidth).Render(strings.Join(insightLines, "\n")),
		)
	} else {
		healthBlock = lipgloss.JoinVertical(
			lipgloss.Left,
			setWidth(m.panelStyle(11), width).Render(strings.Join(healthLines, "\n")),
			"",
			setWidth(m.panelStyle(12), width).Render(strings.Join(insightLines, "\n")),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Title.Render("Home"),
		m.theme.Subtitle.Render("Token intelligence cockpit with live savings, costs, attribution, and quality telemetry."),
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

func (m model) renderMetricCard(title, value, detail string, width int, accentIndex int, valueStyle lipgloss.Style) string {
	return setWidth(m.accentCardStyle(accentIndex), width).Render(strings.Join([]string{
		m.theme.CardLabel.Render(strings.ToUpper(title)),
		valueStyle.Render(value),
		m.theme.CardMeta.Render(detail),
	}, "\n"))
}

func (m model) renderCardGrid(cards []string, columns int) string {
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

func (m model) renderBreakdownPanel(title, keyLabel string, items []tracking.DashboardBreakdown, width int, accentIndex int) string {
	lines := []string{
		m.theme.PanelTitle.Render(title),
		m.theme.CardMeta.Render(fmt.Sprintf("%-18s %9s %7s  %s", keyLabel, "Saved", "Rate", "Share")),
	}
	if len(items) == 0 {
		lines = append(lines, m.theme.Muted.Render("No data"))
		return setWidth(m.panelStyle(accentIndex), width).Render(strings.Join(lines, "\n"))
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
		lines = append(lines, m.renderBreakdownEntry(item, width, i, maxSaved))
	}

	return setWidth(m.panelStyle(accentIndex), width).Render(strings.Join(lines, "\n"))
}

func (m model) renderPlaceholder(width int) string {
	current := m.sections[m.navIndex]
	lines := []string{
		m.theme.Title.Render(current.Title),
		m.theme.Muted.Render(current.Short),
		"",
		m.theme.Warning.Render("Planned next in the phased build."),
		"",
		"This screen is intentionally held behind the shell/data foundation.",
		"The new TUI is being built one slice at a time to avoid the old layout and architecture failures.",
	}
	return setWidth(m.accentCardStyle(13), width).Render(strings.Join(lines, "\n"))
}

func sparklineSaved(points []tracking.DashboardTrendPoint) string {
	if len(points) == 0 {
		return "no trend data"
	}
	values := make([]int64, 0, len(points))
	for _, point := range points {
		values = append(values, point.SavedTokens)
	}
	return sparkline(values)
}

func sparkline(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	blocks := []rune("▁▂▃▄▅▆▇█")
	var maxValue int64
	for _, v := range values {
		if v > maxValue {
			maxValue = v
		}
	}
	if maxValue <= 0 {
		return strings.Repeat(string(blocks[0]), len(values))
	}

	var b strings.Builder
	for _, v := range values {
		idx := int((float64(v) / float64(maxValue)) * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteRune(blocks[idx])
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

func renderHealthLine(label, value string, warn bool) string {
	prefix := "• "
	return prefix + label + ": " + value
}

func displayKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "(unknown)" {
		return "Unattributed"
	}
	return s
}

func (m model) accentCardStyle(index int) lipgloss.Style {
	base := m.theme.Card
	if len(m.theme.AccentColors) == 0 {
		return base
	}
	color := m.theme.AccentColors[index%len(m.theme.AccentColors)]
	return base.BorderForeground(color)
}

func (m model) panelStyle(index int) lipgloss.Style {
	base := m.theme.Panel
	if len(m.theme.AccentColors) == 0 {
		return base
	}
	color := m.theme.AccentColors[index%len(m.theme.AccentColors)]
	return base.BorderForeground(color)
}

func (m model) renderBreakdownEntry(item tracking.DashboardBreakdown, width, index int, maxSaved int64) string {
	keyWidth := max(12, width/2-6)
	barWidth := max(8, width-keyWidth-22)
	key := truncate(displayKey(item.Key), keyWidth)
	head := fmt.Sprintf("%-*s %9s %6.1f%%", keyWidth, key, formatInt(item.SavedTokens), item.ReductionPct)
	bar := m.renderBar(item.SavedTokens, maxSaved, barWidth, index)
	style := m.theme.TableRow
	if index == 0 {
		style = m.theme.TableRowAccent
	}
	share := 0.0
	if maxSaved > 0 {
		share = (float64(item.SavedTokens) / float64(maxSaved)) * 100
	}
	barLine := bar + " " + m.theme.CardMeta.Render(fmt.Sprintf("%5.1f%%", share))
	return style.Render(head) + "\n" + barLine
}

func (m model) renderBar(value, maxValue int64, width, index int) string {
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
	color := m.theme.AccentColors[index%len(m.theme.AccentColors)]
	filledStyle := lipgloss.NewStyle().Foreground(color)
	return filledStyle.Render(strings.Repeat("█", filled)) + m.theme.BarEmpty.Render(strings.Repeat("░", width-filled))
}
