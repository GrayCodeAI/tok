package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// homeSection is the overview cockpit. Stateless for now — all layout
// decisions live in View() and are driven purely by the SectionContext.
type homeSection struct{}

func newHomeSection() *homeSection { return &homeSection{} }

func (h *homeSection) Name() string                { return "Home" }
func (h *homeSection) Short() string               { return "Overview" }
func (h *homeSection) Init(SectionContext) tea.Cmd { return nil }
func (h *homeSection) KeyBindings() []key.Binding  { return nil }

// IsScrollable marks Home as a scrollable block — the root model
// handles up/down/pgup/pgdn/g/G by shifting a viewport offset.
func (h *homeSection) IsScrollable() bool { return true }

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
		// First-run empty state: the user installed tok but no
		// commands have been wrapped yet. Guide them to the two
		// fastest paths to produce data (install the hook, or run
		// a wrapped command) and point at the docs for more.
		return renderOnboarding(th, width)
	}

	snapshot := ctx.Data.Dashboard
	overview := snapshot.Overview
	store := ctx.Data.Sessions.StoreSummary
	quality := ctx.Data.DataQuality

	// Attribution banner — shown at the top of Home when a significant
	// share of commands are missing agent/provider/model/session tags.
	// This is a product-level warning: the hook is capturing commands
	// but not the context around them, so provider-level cost numbers
	// and per-model breakdowns are all pooled into "Unattributed".
	// Surfacing it here is the difference between "tok works" and
	// "tok's numbers look wrong, I don't know why."
	attributionBanner := renderAttributionBanner(th, quality, width)

	// Budget badge — shows when daily budget threshold is crossed
	budgetBadge := renderBudgetBadge(th, ctx.Opts.Budget, ctx.DailySpent, width)

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
		leaderboards = joinPanelsEqualHeight(
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
		healthBlock = joinPanelsEqualHeight(
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

	blocks := []string{
		th.Title.Render("Home"),
		th.Subtitle.Render("Token intelligence cockpit with live savings, costs, attribution, and quality telemetry."),
	}
	if attributionBanner != "" {
		blocks = append(blocks, "", attributionBanner)
	}
	if budgetBadge != "" {
		blocks = append(blocks, "", budgetBadge)
	}
	blocks = append(blocks,
		"",
		cardGrid,
		"",
		trendsBlock,
	)
	// Live Feed: only render when we have events. Before any `tok <cmd>`
	// has run, the panel would be empty and distract from onboarding.
	if feed := renderLiveFeedPanel(th, ctx.LiveFeed, width); feed != "" {
		blocks = append(blocks, "", feed)
	}
	blocks = append(blocks,
		"",
		leaderboards,
		"",
		healthBlock,
	)
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

// renderLiveFeedPanel shows the most recent command events captured
// from the in-process subscribe stream. Newest first, one line per
// entry. Returns empty string when the feed is empty so the panel
// doesn't show before the first event.
func renderLiveFeedPanel(th theme, feed []LiveFeedEntry, width int) string {
	if len(feed) == 0 {
		return ""
	}
	rows := min(len(feed), 6)
	lines := []string{
		th.PanelTitle.Render("Live Feed") + "  " +
			th.CardMeta.Render(fmt.Sprintf("last %d command(s) in this session", rows)),
	}
	now := nowFunc()
	for i := 0; i < rows; i++ {
		e := feed[i]
		age := formatAge(now.Sub(e.At))
		saved := fmt.Sprintf("+%s", formatInt(int64(e.SavedTokens)))
		cmd := e.Command
		// Leave room for marker (2) + saved (10) + age (8) + separators (6).
		cmdWidth := max(12, width-28)
		if len(cmd) > cmdWidth {
			cmd = cmd[:cmdWidth-1] + "…"
		}
		// "Just arrived" flash: entries younger than 2s get a leading
		// marker + focus color to draw the eye. After 2s the flash
		// decays to a plain dot so the feed reads as a timeline rather
		// than a constant highlight.
		marker := " ·"
		cmdStyle := th.CardMeta
		if now.Sub(e.At) < 2*time.Second {
			marker = th.Focus.Render("▸·")
			cmdStyle = th.Focus
		}
		line := fmt.Sprintf("%s %-10s  %-*s  %s",
			marker,
			th.ValuePositive.Render(saved),
			cmdWidth, cmdStyle.Render(cmd),
			th.CardMeta.Render(age))
		lines = append(lines, line)
	}
	return setWidth(panelStyle(th, 3), width).Render(strings.Join(lines, "\n"))
}

// formatAge is the human-readable relative time used in the Live Feed.
// Mirrors formatRelative in view_sessions but trimmed down for the
// Home panel where space is tight.
func formatAge(d time.Duration) string {
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// renderAttributionBanner returns a warning banner — rendered above
// the headline cards — when more than half of the window's tracked
// commands are missing agent / provider / model / session metadata.
// Returns empty string when coverage is healthy so the banner only
// appears when it has something to say.
func renderAttributionBanner(th theme, q tracking.DashboardDataQuality, width int) string {
	if q.TotalCommands == 0 {
		return ""
	}
	worst := q.CommandsMissingAgent
	for _, n := range []int64{
		q.CommandsMissingProvider,
		q.CommandsMissingModel,
		q.CommandsMissingSession,
	} {
		if n > worst {
			worst = n
		}
	}
	// Threshold: warn when ≥50% of commands are missing at least one
	// attribution dimension. Below that, treat it as normal drift.
	if worst*2 < q.TotalCommands {
		return ""
	}
	pct := float64(worst) / float64(q.TotalCommands) * 100
	lines := []string{
		th.Warning.Render("⚠  Attribution coverage is low"),
		fmt.Sprintf("%d of %d commands (%.0f%%) are missing context metadata.",
			worst, q.TotalCommands, pct),
		"Provider/model/cost numbers are pooled into 'Unattributed'.",
		"",
		th.CardMeta.Render("Fix: run  :hooks.diagnose  from the palette to see what's missing,"),
		th.CardMeta.Render("     or reinstall the hook with  tok init -g --force  from a shell."),
	}
	border := panelStyle(th, 0).BorderForeground(th.Warning.GetForeground())
	return setWidth(border, width).Render(strings.Join(lines, "\n"))
}

// renderBudgetBadge returns a warning/danger banner when daily token
// budget threshold is crossed. Returns empty string when under budget.
func renderBudgetBadge(th theme, budget config.BudgetConfig, dailySpent int, width int) string {
	if budget.DailyTokens <= 0 || dailySpent <= 0 {
		return ""
	}

	warningThreshold := budget.WarningThreshold
	if warningThreshold <= 0 {
		warningThreshold = 80
	}

	pct := float64(dailySpent) * 100.0 / float64(budget.DailyTokens)

	// Only show when at warning threshold or above
	if pct < float64(warningThreshold) {
		return ""
	}

	isDanger := pct >= 100
	style := th.Warning
	icon := "⚠"
	label := "Budget warning"
	if isDanger {
		style = th.Danger
		icon = "▲"
		label = "Budget exceeded"
	}

	spentStr := formatCompactNumber(dailySpent)
	budgetStr := formatCompactNumber(int(budget.DailyTokens))

	lines := []string{
		style.Render(icon + "  " + label),
		fmt.Sprintf("Daily: %s / %s tokens (%d%%)", spentStr, budgetStr, int(pct)),
	}

	border := panelStyle(th, 0).BorderForeground(style.GetForeground())
	return setWidth(border, width).Render(strings.Join(lines, "\n"))
}

// renderOnboarding is the empty-state shown on first launch, before
// any commands have been tracked. Everything else in Home assumes a
// snapshot exists — this screen tells the user how to make one.
func renderOnboarding(th theme, width int) string {
	lines := []string{
		th.Title.Render("Welcome to tok"),
		th.Subtitle.Render("No commands tracked yet — here's how to start:"),
		"",
		th.PanelTitle.Render("Option 1: install the global hook"),
		"  $ tok init -g",
		th.CardMeta.Render("  Every bash command your AI agent runs gets wrapped automatically."),
		"",
		th.PanelTitle.Render("Option 2: run a wrapped command yourself"),
		"  $ tok git status",
		"  $ tok npm test",
		th.CardMeta.Render("  Filtered output, token counts tracked. Refresh this screen with 'r'."),
		"",
		th.PanelTitle.Render("Quick tour"),
		th.CardMeta.Render("  1-9,0   jump between sections"),
		th.CardMeta.Render("  :       command palette"),
		th.CardMeta.Render("  /       search within the current view"),
		th.CardMeta.Render("  ?       full keybinding help"),
		th.CardMeta.Render("  q       quit"),
		"",
		th.Muted.Render("Full docs: docs/TUI.md"),
	}
	return setWidth(panelStyle(th, 1), width).Render(strings.Join(lines, "\n"))
}

// joinPanelsEqualHeight renders two side-by-side panels padded to the
// taller of the two, so a short panel doesn't leave a visible gap
// next to a long one. If content panels differ in height we add blank
// lines to the shorter — matching each panel's background style so the
// filler doesn't look like an escape-code gap.
//
// Callers pass pre-rendered panel strings; this function does not
// attempt to re-style them. It's a layout helper, not a style helper.
func joinPanelsEqualHeight(gap string, panels ...string) string {
	if len(panels) == 0 {
		return ""
	}
	maxH := 0
	widths := make([]int, len(panels))
	for i, p := range panels {
		h := lipgloss.Height(p)
		if h > maxH {
			maxH = h
		}
		widths[i] = lipgloss.Width(p)
	}
	padded := make([]string, len(panels))
	for i, p := range panels {
		h := lipgloss.Height(p)
		if h >= maxH {
			padded[i] = p
			continue
		}
		// Append `maxH - h` blank lines, each padded to the panel's
		// existing visible width so JoinHorizontal doesn't treat the
		// shorter panel as ragged.
		blank := strings.Repeat(" ", widths[i])
		extra := make([]string, maxH-h)
		for j := range extra {
			extra[j] = blank
		}
		padded[i] = p + "\n" + strings.Join(extra, "\n")
	}
	return joinHorizontalGap(gap, padded...)
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

// Column layout for the two-line breakdown entry:
//
//   <keyWidth>  <savedCol>  <rateCol>
//   <barWidth>  <shareCol>
//
// savedCol/rateCol/shareCol widths are fixed so numerics right-align
// predictably. keyWidth and barWidth are derived from the panel width.
// The header is built from the SAME constants as the data rows so
// "Saved" sits over the saved numbers and "Rate" sits over the
// percentages.
const (
	breakdownSavedCol = 9 // "999.9k" width
	breakdownRateCol  = 7 // "100.0%" width
	breakdownShareCol = 6 // "100.0%" width for the bar line
)

func breakdownKeyWidth(panelWidth int) int {
	// Panel width minus fixed numeric columns and their 2 single-space
	// separators, bounded so the key never becomes unreadable.
	inner := panelWidth - breakdownSavedCol - breakdownRateCol - 2
	if inner < 12 {
		inner = 12
	}
	if inner > 36 {
		inner = 36
	}
	return inner
}

func renderBreakdownPanel(th theme, title, keyLabel string, items []tracking.DashboardBreakdown, width, accentIndex int) string {
	keyWidth := breakdownKeyWidth(width)
	header := fmt.Sprintf("%-*s %*s %*s",
		keyWidth, keyLabel,
		breakdownSavedCol, "Saved",
		breakdownRateCol, "Rate")
	lines := []string{
		th.PanelTitle.Render(title),
		th.CardMeta.Render(header),
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
	keyWidth := breakdownKeyWidth(width)
	// Bar line sits under the key + gets the share %. Its width is
	// whatever's left after the share column + its leading space.
	barWidth := max(8, width-breakdownShareCol-3)
	keyStr := truncate(displayKey(item.Key), keyWidth)
	head := fmt.Sprintf("%-*s %*s %*s",
		keyWidth, keyStr,
		breakdownSavedCol, formatInt(item.SavedTokens),
		breakdownRateCol, fmt.Sprintf("%.1f%%", item.ReductionPct))
	bar := renderBar(th, item.SavedTokens, maxSaved, barWidth, index)
	style := th.TableRow
	if index == 0 {
		style = th.TableRowAccent
	}
	share := 0.0
	if maxSaved > 0 {
		share = (float64(item.SavedTokens) / float64(maxSaved)) * 100
	}
	barLine := bar + " " + th.CardMeta.Render(fmt.Sprintf("%*s", breakdownShareCol, fmt.Sprintf("%.1f%%", share)))
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
