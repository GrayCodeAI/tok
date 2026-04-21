package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

// providersSection ranks providers by token savings and lets the user
// drill into a provider to see the per-model breakdown scoped to it.
// Provider→model association comes from Dashboard.TopProviderModels,
// whose Key is formatted "<provider> / <model>", so we derive the
// scoped model list by prefix match.
type providersSection struct {
	table        *Table
	drill        string // provider key when non-empty
	snapshotKey  string
}

func newProvidersSection() *providersSection {
	return &providersSection{
		table: NewTable([]Column{
			{Title: "Provider", MinWidth: 14, Sortable: true},
			{Title: "Commands", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Saved", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Reduction", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Cost saved", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *providersSection) Name() string  { return "Providers" }
func (s *providersSection) Short() string { return "Economics" }

func (s *providersSection) ExportColumns() []Column { return s.table.Columns() }
func (s *providersSection) ExportRows() []Row       { return s.table.VisibleRows() }
func (s *providersSection) ExportName() string      { return "providers" }

func (s *providersSection) Init(SectionContext) tea.Cmd { return nil }

func (s *providersSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect provider")),
		key.NewBinding(key.WithKeys("backspace", "esc"), key.WithHelp("⌫/esc", "back to list")),
	}
}

func (s *providersSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	if ctx.Data != nil && ctx.Data.Dashboard != nil {
		newKey := providersSnapshotKey(ctx.Data.Dashboard.TopProviders)
		if newKey != s.snapshotKey {
			s.snapshotKey = newKey
			s.table.SetRows(providerRows(ctx.Data.Dashboard.TopProviders))
		}
	} else if s.snapshotKey != "" {
		s.snapshotKey = ""
		s.table.SetRows(nil)
	}

	switch m := msg.(type) {
	case searchMsg:
		s.table.SetFilter(m.Query)
	case tea.KeyMsg:
		if s.drill != "" {
			switch m.String() {
			case "backspace", "esc":
				s.drill = ""
			}
			return s, nil
		}
		switch m.String() {
		case "j", "down":
			s.table.MoveDown()
		case "k", "up":
			s.table.MoveUp()
		case "g", "home":
			s.table.Top()
		case "G", "end":
			s.table.Bottom()
		case "pgdown", "ctrl+f":
			s.table.PageDown()
		case "pgup", "ctrl+b":
			s.table.PageUp()
		case "enter":
			if row, ok := s.table.Selected(); ok {
				if key, castOk := row.Payload.(string); castOk {
					s.drill = key
				}
			}
		case "y":
			if row, ok := s.table.Selected(); ok {
				return s, YankCmd(RowToTSV(row))
			}
		}
	}
	return s, nil
}

func (s *providersSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	dashboard := ctx.Data.Dashboard
	if s.drill != "" {
		if b, ok := findBreakdown(dashboard.TopProviders, s.drill); ok {
			return s.renderDetail(ctx, b)
		}
		s.drill = ""
	}
	return s.renderList(ctx)
}

func (s *providersSection) renderList(ctx SectionContext) string {
	th := ctx.Theme
	dashboard := ctx.Data.Dashboard

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Providers"),
		th.Subtitle.Render(fmt.Sprintf("%d providers tracked · ranked by saved tokens",
			dashboard.Overview.UniqueProviders)),
	)
	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-2)
	tableView := s.table.View(th, ctx.Width, bodyHeight)
	hint := th.CardMeta.Render("enter: drill into provider-models · /: filter")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", tableView, "", hint)
}

func (s *providersSection) renderDetail(ctx SectionContext, b tracking.DashboardBreakdown) string {
	th := ctx.Theme
	width := ctx.Width

	title := th.Title.Render("Provider: "+displayKey(b.Key)) + "  " +
		th.Muted.Render("(esc/backspace to return)")

	metrics := []string{
		th.PanelTitle.Render("Summary"),
		renderHealthLine("Commands", fmt.Sprintf("%d", b.Commands)),
		renderHealthLine("Saved tokens", formatInt(b.SavedTokens)),
		renderHealthLine("Reduction", fmt.Sprintf("%.1f%%", b.ReductionPct)),
		renderHealthLine("Original cost", fmt.Sprintf("$%.4f", b.EstimatedOriginalCostUSD)),
		renderHealthLine("Filtered cost", fmt.Sprintf("$%.4f", b.EstimatedFilteredCostUSD)),
		renderHealthLine("Cost saved", fmt.Sprintf("$%.4f", b.EstimatedSavingsUSD)),
	}
	summary := setWidth(panelStyle(th, 2), width).Render(strings.Join(metrics, "\n"))

	models := filterProviderModels(ctx.Data.Dashboard.TopProviderModels, b.Key)
	modelLines := []string{th.PanelTitle.Render("Models on this provider")}
	if len(models) == 0 {
		modelLines = append(modelLines, th.Muted.Render("no per-model data"))
	} else {
		modelLines = append(modelLines, th.CardMeta.Render(
			fmt.Sprintf("%-24s %10s %10s  %s", "Model", "Commands", "Saved", "Reduction"),
		))
		for _, m := range models {
			modelLines = append(modelLines, fmt.Sprintf("%-24s %10d %10s  %5.1f%%",
				truncate(extractModelFromKey(m.Key), 24),
				m.Commands,
				formatInt(m.SavedTokens),
				m.ReductionPct,
			))
		}
	}
	modelPanel := setWidth(panelStyle(th, 3), width).Render(strings.Join(modelLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", summary, "", modelPanel)
}

// --- helpers ---

func providersSnapshotKey(items []tracking.DashboardBreakdown) string {
	var b strings.Builder
	fmt.Fprintf(&b, "n=%d", len(items))
	if len(items) > 0 {
		fmt.Fprintf(&b, ";top=%s@%d", items[0].Key, items[0].SavedTokens)
	}
	return b.String()
}

func providerRows(items []tracking.DashboardBreakdown) []Row {
	rows := make([]Row, 0, len(items))
	for _, it := range items {
		rows = append(rows, Row{
			Cells: []string{
				displayKey(it.Key),
				fmt.Sprintf("%d", it.Commands),
				fmt.Sprintf("%d", it.SavedTokens),
				fmt.Sprintf("%.1f", it.ReductionPct),
				fmt.Sprintf("%.4f", it.EstimatedSavingsUSD),
			},
			Payload: it.Key,
		})
	}
	return rows
}

func findBreakdown(items []tracking.DashboardBreakdown, key string) (tracking.DashboardBreakdown, bool) {
	for _, it := range items {
		if it.Key == key {
			return it, true
		}
	}
	return tracking.DashboardBreakdown{}, false
}

// filterProviderModels returns the subset of per-provider-model entries
// whose key starts with "<provider> / ". The dashboard query builds the
// composite key that way, so a prefix match is reliable.
func filterProviderModels(items []tracking.DashboardBreakdown, provider string) []tracking.DashboardBreakdown {
	prefix := provider + " / "
	out := make([]tracking.DashboardBreakdown, 0, len(items))
	for _, it := range items {
		if strings.HasPrefix(it.Key, prefix) {
			out = append(out, it)
		}
	}
	return out
}

func extractModelFromKey(composite string) string {
	idx := strings.Index(composite, " / ")
	if idx < 0 {
		return composite
	}
	return composite[idx+3:]
}
