package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tok/internal/session"
)

// sessionsSection lists recent sessions from the workspace snapshot and
// offers drill-down into a detail pane. The section owns its Table and
// its drill-selection state; the root model owns search/palette.
type sessionsSection struct {
	table       *Table
	drillID     string // when non-empty, render the detail pane
	snapshotKey string // which snapshot payload populated the table (used to avoid re-syncing every frame)
}

func newSessionsSection() *sessionsSection {
	return &sessionsSection{
		table: NewTable([]Column{
			{Title: "Agent", MinWidth: 10, Sortable: true},
			{Title: "Project", MinWidth: 14, Sortable: true},
			{Title: "Started", MinWidth: 10, Sortable: true, Numeric: true},
			// "Last" is shorter and less ambiguous than "Active" —
			// users reported confusion over whether "Active" meant
			// "currently active" or "last active".
			{Title: "Last", MinWidth: 8, Sortable: true, Numeric: true},
			{Title: "Tokens", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Turns", MinWidth: 6, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Snaps", MinWidth: 6, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *sessionsSection) Name() string  { return "Sessions" }
func (s *sessionsSection) Short() string { return "Session Ops" }

// Export* implementations satisfy ExportableTable so Phase 3 e-export
// can serialize this view without the section reaching into the
// internal table from the outside.
func (s *sessionsSection) ExportColumns() []Column { return s.table.Columns() }
func (s *sessionsSection) ExportRows() []Row       { return s.table.VisibleRows() }
func (s *sessionsSection) ExportName() string      { return "sessions" }

func (s *sessionsSection) Init(SectionContext) tea.Cmd { return nil }

func (s *sessionsSection) KeyBindings() []key.Binding {
	// Surface only section-local keys — the root's global keymap (nav,
	// refresh, palette, quit) is rendered alongside this in the help
	// overlay and does not need to be re-declared here.
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect session")),
		key.NewBinding(key.WithKeys("backspace", "esc"), key.WithHelp("⌫/esc", "back to list")),
	}
}

func (s *sessionsSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	// Sync the table from the snapshot whenever the data payload pointer
	// changes. We key on "who is top of the first entry + count" rather
	// than on the pointer directly so that an unchanged payload doesn't
	// repaint the table on every frame (which would reset scroll offset).
	if ctx.Data != nil && ctx.Data.Sessions != nil {
		newKey := snapshotKeyFor(ctx.Data.Sessions)
		if newKey != s.snapshotKey {
			s.snapshotKey = newKey
			s.table.SetRows(sessionRows(ctx.Data.Sessions.RecentSessions))
		}
	} else if s.snapshotKey != "" {
		s.snapshotKey = ""
		s.table.SetRows(nil)
	}

	switch m := msg.(type) {
	case searchMsg:
		s.table.SetFilter(m.Query)
	case tea.KeyMsg:
		// If we're in drill-down, esc/backspace returns to the list and
		// consumes the key so it doesn't cascade to the root handlers.
		if s.drillID != "" {
			switch m.String() {
			case "backspace", "esc":
				s.drillID = ""
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
				if id, castOk := row.Payload.(string); castOk {
					s.drillID = id
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

func (s *sessionsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Sessions == nil {
		return th.Muted.Render("No session data yet.")
	}
	if s.drillID != "" {
		if overview, ok := findSessionOverview(ctx.Data.Sessions.RecentSessions, s.drillID); ok {
			return s.renderDetail(ctx, overview)
		}
		// Drill target vanished (session was closed mid-view) — snap back.
		s.drillID = ""
	}
	return s.renderList(ctx)
}

func (s *sessionsSection) renderList(ctx SectionContext) string {
	th := ctx.Theme
	summary := ctx.Data.Sessions.StoreSummary

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Sessions"),
		th.Subtitle.Render(fmt.Sprintf("%d total · %d active · top agent: %s",
			summary.TotalSessions,
			summary.ActiveSessions,
			displayKey(summary.TopAgent),
		)),
	)

	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-2)
	tableView := s.table.View(th, ctx.Width, bodyHeight)

	hint := th.CardMeta.Render("enter: inspect · j/k: move · /: filter")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		tableView,
		"",
		hint,
	)
}

func (s *sessionsSection) renderDetail(ctx SectionContext, o session.SessionOverview) string {
	th := ctx.Theme
	width := ctx.Width

	title := th.Title.Render("Session detail") + "  " +
		th.Muted.Render("(esc/backspace to return)")

	status := th.Muted.Render("inactive")
	if o.IsActive {
		status = th.Positive.Render("● active")
	}

	lines := []string{
		th.PanelTitle.Render("Identity"),
		renderHealthLine("Session ID", o.ID),
		renderHealthLine("Agent", displayKey(o.Agent)),
		renderHealthLine("Project", fullPath(o.ProjectPath)),
		renderHealthLine("Status", status),
		"",
		th.PanelTitle.Render("Activity"),
		renderHealthLine("Started", formatRelative(o.StartedAt)),
		renderHealthLine("Last activity", formatRelative(o.LastActivity)),
		renderHealthLine("Turns", fmt.Sprintf("%d", o.TotalTurns)),
		renderHealthLine("Tokens", formatInt(int64(o.TotalTokens))),
		renderHealthLine("Compression ratio", fmt.Sprintf("%.2fx", o.CompressionRatio)),
		renderHealthLine("Context blocks", fmt.Sprintf("%d", o.ContextBlockCount)),
		"",
		th.PanelTitle.Render("Snapshots"),
		renderHealthLine("Count", fmt.Sprintf("%d", o.SnapshotCount)),
	}
	if o.LastSnapshotAt != nil {
		lines = append(lines, renderHealthLine("Last snapshot", formatRelative(*o.LastSnapshotAt)))
	} else {
		lines = append(lines, renderHealthLine("Last snapshot", "never"))
	}

	pane := setWidth(panelStyle(th, 6), width).Render(strings.Join(lines, "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, title, "", pane)
}

// --- helpers ---

// snapshotKeyFor derives a coarse identity for a session analytics
// payload so the table only re-syncs when something actually changed.
func snapshotKeyFor(snap *session.SessionAnalyticsSnapshot) string {
	if snap == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "t%d;a%d;r%d",
		snap.StoreSummary.TotalSessions,
		snap.StoreSummary.ActiveSessions,
		len(snap.RecentSessions),
	)
	if len(snap.RecentSessions) > 0 {
		first := snap.RecentSessions[0]
		fmt.Fprintf(&b, ";top=%s@%d", first.ID, first.LastActivity.Unix())
	}
	return b.String()
}

func sessionRows(overviews []session.SessionOverview) []Row {
	rows := make([]Row, 0, len(overviews))
	for _, o := range overviews {
		rows = append(rows, Row{
			Cells: []string{
				displayKey(o.Agent),
				displayPath(o.ProjectPath),
				formatRelative(o.StartedAt),
				formatRelative(o.LastActivity),
				fmt.Sprintf("%d", o.TotalTokens),
				fmt.Sprintf("%d", o.TotalTurns),
				fmt.Sprintf("%d", o.SnapshotCount),
			},
			Payload: o.ID,
		})
	}
	return rows
}

func findSessionOverview(list []session.SessionOverview, id string) (session.SessionOverview, bool) {
	for _, o := range list {
		if o.ID == id {
			return o, true
		}
	}
	return session.SessionOverview{}, false
}

// displayPath shortens long absolute paths to just the basename — enough
// to identify a project without blowing out the Project column width.
// The detail pane shows the full path so nothing is lost.
func displayPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "—"
	}
	return filepath.Base(p)
}

func fullPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "—"
	}
	return p
}

// nowFunc is the package-level clock used by relative-time renderers.
// Tests override it to pin the golden output. Production callers see
// time.Now.
var nowFunc = time.Now

// formatRelative renders a timestamp as "3m ago", "2h ago", "just now".
// Absolute fallback on very old timestamps keeps the table scannable.
func formatRelative(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := nowFunc().Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
