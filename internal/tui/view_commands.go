package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// commandsSection shows a dual view: the top command patterns by saved
// tokens on one side, and the weakest commands (where compression is
// underperforming) on the other. Toggle between "top" and "weak" with
// 't' / 'w'. Drill-down surfaces the full command text plus the global
// layer breakdown as supporting context — per-command layer data isn't
// in the snapshot yet, so the caption calls out the scope.
type commandsSection struct {
	table       *Table
	mode        commandsMode
	drill       string
	snapshotKey string
}

type commandsMode int

const (
	commandsTop commandsMode = iota
	commandsWeak
)

func newCommandsSection() *commandsSection {
	return &commandsSection{
		table: NewTable([]Column{
			{Title: "Command", MinWidth: 28, Sortable: true},
			{Title: "Count", MinWidth: 7, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Saved", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Reduction", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *commandsSection) Name() string                { return "Commands" }
func (s *commandsSection) Short() string               { return "Command Mix" }
func (s *commandsSection) Init(SectionContext) tea.Cmd { return nil }

func (s *commandsSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "top commands")),
		key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "weak commands")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect command")),
		key.NewBinding(key.WithKeys("backspace", "esc"), key.WithHelp("⌫/esc", "back to list")),
	}
}

func (s *commandsSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	s.syncTable(ctx)

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
		case "t":
			if s.mode != commandsTop {
				s.mode = commandsTop
				s.snapshotKey = "" // force resync to new source
				s.syncTable(ctx)
			}
			return s, nil
		case "w":
			if s.mode != commandsWeak {
				s.mode = commandsWeak
				s.snapshotKey = ""
				s.syncTable(ctx)
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

func (s *commandsSection) syncTable(ctx SectionContext) {
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		if s.snapshotKey != "" {
			s.snapshotKey = ""
			s.table.SetRows(nil)
		}
		return
	}
	src := ctx.Data.Dashboard.TopCommands
	if s.mode == commandsWeak {
		src = ctx.Data.Dashboard.LowSavingsCommands
	}
	key := fmt.Sprintf("mode=%d;%s", s.mode, providersSnapshotKey(src))
	if key != s.snapshotKey {
		s.snapshotKey = key
		s.table.SetRows(commandRows(src))
	}
}

func (s *commandsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	dashboard := ctx.Data.Dashboard
	if s.drill != "" {
		src := dashboard.TopCommands
		if s.mode == commandsWeak {
			src = dashboard.LowSavingsCommands
		}
		if b, ok := findBreakdown(src, s.drill); ok {
			return s.renderDetail(ctx, b)
		}
		s.drill = ""
	}
	return s.renderList(ctx)
}

func (s *commandsSection) renderList(ctx SectionContext) string {
	th := ctx.Theme
	subtitle := "Top commands by saved tokens"
	if s.mode == commandsWeak {
		subtitle = "Weakest commands (lowest reduction in the window)"
	}
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Commands"),
		th.Subtitle.Render(subtitle+" · press t/w to switch"),
	)
	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-2)
	tableView := s.table.View(th, ctx.Width, bodyHeight)
	hint := th.CardMeta.Render("enter: drill · /: filter · t: top · w: weak")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", tableView, "", hint)
}

func (s *commandsSection) renderDetail(ctx SectionContext, b tracking.DashboardBreakdown) string {
	th := ctx.Theme
	width := ctx.Width

	title := th.Title.Render("Command") + "  " +
		th.Muted.Render("(esc/backspace to return)")

	summary := setWidth(panelStyle(th, 4), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Command"),
		th.Muted.Render(b.Key),
		"",
		renderHealthLine("Count", fmt.Sprintf("%d", b.Commands)),
		renderHealthLine("Saved tokens", formatInt(b.SavedTokens)),
		renderHealthLine("Reduction", fmt.Sprintf("%.1f%%", b.ReductionPct)),
		renderHealthLine("Original cost", fmt.Sprintf("$%.4f", b.EstimatedOriginalCostUSD)),
		renderHealthLine("Filtered cost", fmt.Sprintf("$%.4f", b.EstimatedFilteredCostUSD)),
		renderHealthLine("Cost saved", fmt.Sprintf("$%.4f", b.EstimatedSavingsUSD)),
	}, "\n"))

	layers := ctx.Data.Dashboard.TopLayers
	layerLines := []string{
		th.PanelTitle.Render("Top pipeline layers"),
		th.Muted.Render("(aggregate across all commands — per-command layer data not yet indexed)"),
	}
	if len(layers) == 0 {
		layerLines = append(layerLines, th.Muted.Render("no layer stats"))
	} else {
		layerLines = append(layerLines, th.CardMeta.Render(
			fmt.Sprintf("%-26s %10s %10s %10s", "Layer", "Calls", "Total saved", "Avg saved"),
		))
		for _, l := range layers {
			layerLines = append(layerLines, fmt.Sprintf("%-26s %10d %10s %10.1f",
				truncate(l.LayerName, 26),
				l.CallCount,
				formatInt(l.TotalSaved),
				l.AvgSaved,
			))
		}
	}
	layerPanel := setWidth(panelStyle(th, 7), width).Render(strings.Join(layerLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", summary, "", layerPanel)
}

// --- helpers ---

func commandRows(items []tracking.DashboardBreakdown) []Row {
	rows := make([]Row, 0, len(items))
	for _, it := range items {
		rows = append(rows, Row{
			Cells: []string{
				truncate(it.Key, 60),
				fmt.Sprintf("%d", it.Commands),
				fmt.Sprintf("%d", it.SavedTokens),
				fmt.Sprintf("%.1f", it.ReductionPct),
			},
			Payload: it.Key,
		})
	}
	return rows
}
