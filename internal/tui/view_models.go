package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// modelsSection ranks models by saved tokens. Drill shows the provider
// partners for the selected model (derived from TopProviderModels via
// "<provider> / <model>" suffix match) so users can see which upstream
// providers contribute most to the model's savings.
type modelsSection struct {
	table       *Table
	drill       string
	snapshotKey string
}

func newModelsSection() *modelsSection {
	return &modelsSection{
		table: NewTable([]Column{
			{Title: "Model", MinWidth: 18, Sortable: true},
			{Title: "Commands", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Saved", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Reduction", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Cost saved", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *modelsSection) Name() string                      { return "Models" }
func (s *modelsSection) Short() string                     { return "Model Cost" }
func (s *modelsSection) Init(SectionContext) tea.Cmd       { return nil }

func (s *modelsSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect model")),
		key.NewBinding(key.WithKeys("backspace", "esc"), key.WithHelp("⌫/esc", "back to list")),
	}
}

func (s *modelsSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	if ctx.Data != nil && ctx.Data.Dashboard != nil {
		k := providersSnapshotKey(ctx.Data.Dashboard.TopModels)
		if k != s.snapshotKey {
			s.snapshotKey = k
			s.table.SetRows(providerRows(ctx.Data.Dashboard.TopModels))
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
		handleTableNav(s.table, m)
		if m.String() == "enter" {
			if row, ok := s.table.Selected(); ok {
				if key, castOk := row.Payload.(string); castOk {
					s.drill = key
				}
			}
		}
	}
	return s, nil
}

func (s *modelsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	dashboard := ctx.Data.Dashboard
	if s.drill != "" {
		if b, ok := findBreakdown(dashboard.TopModels, s.drill); ok {
			return s.renderDetail(ctx, b)
		}
		s.drill = ""
	}
	return s.renderList(ctx)
}

func (s *modelsSection) renderList(ctx SectionContext) string {
	th := ctx.Theme
	dashboard := ctx.Data.Dashboard
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Models"),
		th.Subtitle.Render(fmt.Sprintf("%d unique models tracked · ranked by saved tokens",
			dashboard.Overview.UniqueModels)),
	)
	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-2)
	tableView := s.table.View(th, ctx.Width, bodyHeight)
	hint := th.CardMeta.Render("enter: drill · /: filter")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", tableView, "", hint)
}

func (s *modelsSection) renderDetail(ctx SectionContext, b tracking.DashboardBreakdown) string {
	th := ctx.Theme
	width := ctx.Width

	title := th.Title.Render("Model: "+displayKey(b.Key)) + "  " +
		th.Muted.Render("(esc/backspace to return)")

	summary := setWidth(panelStyle(th, 5), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Summary"),
		renderHealthLine("Commands", fmt.Sprintf("%d", b.Commands)),
		renderHealthLine("Saved tokens", formatInt(b.SavedTokens)),
		renderHealthLine("Reduction", fmt.Sprintf("%.1f%%", b.ReductionPct)),
		renderHealthLine("Original cost", fmt.Sprintf("$%.4f", b.EstimatedOriginalCostUSD)),
		renderHealthLine("Filtered cost", fmt.Sprintf("$%.4f", b.EstimatedFilteredCostUSD)),
		renderHealthLine("Cost saved", fmt.Sprintf("$%.4f", b.EstimatedSavingsUSD)),
	}, "\n"))

	providers := filterProviderPartners(ctx.Data.Dashboard.TopProviderModels, b.Key)
	providerLines := []string{th.PanelTitle.Render("Providers serving this model")}
	if len(providers) == 0 {
		providerLines = append(providerLines, th.Muted.Render("no provider attribution"))
	} else {
		providerLines = append(providerLines, th.CardMeta.Render(
			fmt.Sprintf("%-20s %10s %10s  %s", "Provider", "Commands", "Saved", "Reduction"),
		))
		for _, p := range providers {
			providerLines = append(providerLines, fmt.Sprintf("%-20s %10d %10s  %5.1f%%",
				truncate(extractProviderFromKey(p.Key), 20),
				p.Commands,
				formatInt(p.SavedTokens),
				p.ReductionPct,
			))
		}
	}
	providerPanel := setWidth(panelStyle(th, 6), width).Render(strings.Join(providerLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", summary, "", providerPanel)
}

// --- helpers shared with other breakdown sections ---

// handleTableNav is the small cursor helper that every section with an
// embedded Table wires into its Update. Kept out of Table itself so
// sections can suppress or layer their own keys (e.g. enter → drill).
func handleTableNav(t *Table, m tea.KeyMsg) {
	switch m.String() {
	case "j", "down":
		t.MoveDown()
	case "k", "up":
		t.MoveUp()
	case "g", "home":
		t.Top()
	case "G", "end":
		t.Bottom()
	case "pgdown", "ctrl+f":
		t.PageDown()
	case "pgup", "ctrl+b":
		t.PageUp()
	}
}

// filterProviderPartners is the mirror of filterProviderModels: given
// "model", return every "<any> / model" entry.
func filterProviderPartners(items []tracking.DashboardBreakdown, model string) []tracking.DashboardBreakdown {
	suffix := " / " + model
	out := make([]tracking.DashboardBreakdown, 0, len(items))
	for _, it := range items {
		if strings.HasSuffix(it.Key, suffix) {
			out = append(out, it)
		}
	}
	return out
}

func extractProviderFromKey(composite string) string {
	idx := strings.Index(composite, " / ")
	if idx < 0 {
		return composite
	}
	return composite[:idx]
}
