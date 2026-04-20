package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// pipelineSection visualizes the filter pipeline's per-layer impact.
// Unlike the breakdown sections (Providers/Models/Agents) whose rows
// are external entities, pipeline rows are internal components — so
// the view leans harder on visual bars and total-share labels instead
// of drill-downs (there isn't another level of data to drill into).
type pipelineSection struct {
	table       *Table
	snapshotKey string
}

func newPipelineSection() *pipelineSection {
	return &pipelineSection{
		table: NewTable([]Column{
			{Title: "Layer", MinWidth: 22, Sortable: true},
			{Title: "Calls", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Total saved", MinWidth: 11, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Avg / call", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Share", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *pipelineSection) Name() string  { return "Pipeline" }
func (s *pipelineSection) Short() string { return "Layer View" }

func (s *pipelineSection) ExportColumns() []Column { return s.table.Columns() }
func (s *pipelineSection) ExportRows() []Row       { return s.table.VisibleRows() }
func (s *pipelineSection) ExportName() string      { return "pipeline" }
func (s *pipelineSection) Init(SectionContext) tea.Cmd { return nil }

func (s *pipelineSection) KeyBindings() []key.Binding {
	// No drill — pipeline rows are leaf data. Surface only the table
	// cursor keys so the help overlay reads accurately.
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "cursor")),
	}
}

func (s *pipelineSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	s.sync(ctx)
	switch m := msg.(type) {
	case searchMsg:
		s.table.SetFilter(m.Query)
	case tea.KeyMsg:
		handleTableNav(s.table, m)
		if m.String() == "y" {
			if row, ok := s.table.Selected(); ok {
				return s, YankCmd(RowToTSV(row))
			}
		}
	}
	return s, nil
}

func (s *pipelineSection) sync(ctx SectionContext) {
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		if s.snapshotKey != "" {
			s.snapshotKey = ""
			s.table.SetRows(nil)
		}
		return
	}
	layers := ctx.Data.Dashboard.TopLayers
	key := layersSnapshotKey(layers)
	if key == s.snapshotKey {
		return
	}
	s.snapshotKey = key
	total := totalLayerSaved(layers)
	rows := make([]Row, 0, len(layers))
	for _, l := range layers {
		share := 0.0
		if total > 0 {
			share = (float64(l.TotalSaved) / float64(total)) * 100
		}
		rows = append(rows, Row{
			Cells: []string{
				l.LayerName,
				fmt.Sprintf("%d", l.CallCount),
				fmt.Sprintf("%d", l.TotalSaved),
				fmt.Sprintf("%.1f", l.AvgSaved),
				fmt.Sprintf("%.1f", share),
			},
			Payload: l.LayerName,
		})
	}
	s.table.SetRows(rows)
}

func (s *pipelineSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No pipeline data yet.")
	}
	layers := ctx.Data.Dashboard.TopLayers
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Pipeline"),
		th.Subtitle.Render(fmt.Sprintf("%d layers · %s tokens saved across the pipeline",
			len(layers), formatInt(totalLayerSaved(layers)))),
	)

	// Top contributors panel: horizontal bars for the 6 layers with the
	// highest TotalSaved. This complements the table below by making
	// relative contribution visible at a glance.
	barsPanel := renderLayerBars(th, layers, ctx.Width)

	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-lipgloss.Height(barsPanel)-4)
	tableView := s.table.View(th, ctx.Width, bodyHeight)

	hint := th.CardMeta.Render("j/k: move · /: filter · columns are sortable")

	return lipgloss.JoinVertical(lipgloss.Left, header, "", barsPanel, "", tableView, "", hint)
}

func renderLayerBars(th theme, layers []tracking.DashboardLayerSummary, width int) string {
	if len(layers) == 0 {
		return th.Muted.Render("No layer stats yet.")
	}
	maxSaved := int64(0)
	for _, l := range layers {
		if l.TotalSaved > maxSaved {
			maxSaved = l.TotalSaved
		}
	}
	lines := []string{th.PanelTitle.Render("Top contributors")}
	take := 6
	if take > len(layers) {
		take = len(layers)
	}
	nameWidth := 22
	barWidth := max(8, width-nameWidth-30)
	for i := 0; i < take; i++ {
		l := layers[i]
		name := truncate(l.LayerName, nameWidth)
		bar := renderBar(th, l.TotalSaved, maxSaved, barWidth, i)
		lines = append(lines, fmt.Sprintf("%-*s  %s  %s",
			nameWidth, name, bar, th.CardMeta.Render(
				fmt.Sprintf("%s  %.1f / call", formatInt(l.TotalSaved), l.AvgSaved),
			)))
	}
	return setWidth(panelStyle(th, 9), width).Render(strings.Join(lines, "\n"))
}

func totalLayerSaved(layers []tracking.DashboardLayerSummary) int64 {
	var total int64
	for _, l := range layers {
		total += l.TotalSaved
	}
	return total
}

func layersSnapshotKey(layers []tracking.DashboardLayerSummary) string {
	if len(layers) == 0 {
		return ""
	}
	return fmt.Sprintf("n=%d;top=%s@%d", len(layers), layers[0].LayerName, layers[0].TotalSaved)
}
