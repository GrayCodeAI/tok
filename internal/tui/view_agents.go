package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// agentsSection ranks agents (Claude, Copilot, Cursor, etc.) by saved
// tokens. The drill pane surfaces richer attribution (top projects and
// active sessions for the agent) so users can tell whether a "heavy"
// agent is broadly used or concentrated in one project.
type agentsSection struct {
	table       *Table
	drill       string
	snapshotKey string
}

func newAgentsSection() *agentsSection {
	return &agentsSection{
		table: NewTable([]Column{
			{Title: "Agent", MinWidth: 14, Sortable: true},
			{Title: "Commands", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Saved", MinWidth: 8, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
			{Title: "Reduction", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
			{Title: "Cost saved", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
		}),
	}
}

func (s *agentsSection) Name() string                { return "Agents" }
func (s *agentsSection) Short() string               { return "Agent Ops" }
func (s *agentsSection) Init(SectionContext) tea.Cmd { return nil }

func (s *agentsSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect agent")),
		key.NewBinding(key.WithKeys("backspace", "esc"), key.WithHelp("⌫/esc", "back to list")),
	}
}

func (s *agentsSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	if ctx.Data != nil && ctx.Data.Dashboard != nil {
		k := providersSnapshotKey(ctx.Data.Dashboard.TopAgents)
		if k != s.snapshotKey {
			s.snapshotKey = k
			s.table.SetRows(providerRows(ctx.Data.Dashboard.TopAgents))
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

func (s *agentsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	if ctx.Data == nil || ctx.Data.Dashboard == nil {
		return th.Muted.Render("No dashboard data yet.")
	}
	dashboard := ctx.Data.Dashboard
	if s.drill != "" {
		if b, ok := findBreakdown(dashboard.TopAgents, s.drill); ok {
			return s.renderDetail(ctx, b)
		}
		s.drill = ""
	}
	return s.renderList(ctx)
}

func (s *agentsSection) renderList(ctx SectionContext) string {
	th := ctx.Theme
	dashboard := ctx.Data.Dashboard
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Agents"),
		th.Subtitle.Render(fmt.Sprintf("%d agents tracked · %d active sessions · ranked by saved tokens",
			dashboard.Overview.UniqueAgents,
			activeSessionCount(ctx),
		)),
	)
	bodyHeight := max(10, ctx.Height-lipgloss.Height(header)-2)
	tableView := s.table.View(th, ctx.Width, bodyHeight)
	hint := th.CardMeta.Render("enter: drill · /: filter")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", tableView, "", hint)
}

func (s *agentsSection) renderDetail(ctx SectionContext, b tracking.DashboardBreakdown) string {
	th := ctx.Theme
	width := ctx.Width
	title := th.Title.Render("Agent: "+displayKey(b.Key)) + "  " +
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

	sessionLines := []string{th.PanelTitle.Render("Recent sessions for this agent")}
	sessions := agentSessions(ctx, b.Key)
	if len(sessions) == 0 {
		sessionLines = append(sessionLines, th.Muted.Render("no session records"))
	} else {
		sessionLines = append(sessionLines, th.CardMeta.Render(
			fmt.Sprintf("%-16s %-20s %10s %s", "Project", "ID", "Tokens", "Last activity"),
		))
		for _, o := range sessions {
			sessionLines = append(sessionLines, fmt.Sprintf("%-16s %-20s %10d %s",
				truncate(displayPath(o.ProjectPath), 16),
				truncate(o.ID, 20),
				o.TotalTokens,
				formatRelative(o.LastActivity),
			))
		}
	}
	sessionPanel := setWidth(panelStyle(th, 6), width).Render(strings.Join(sessionLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", summary, "", sessionPanel)
}

// --- helpers ---

func activeSessionCount(ctx SectionContext) int64 {
	if ctx.Data == nil || ctx.Data.Sessions == nil {
		return 0
	}
	return ctx.Data.Sessions.StoreSummary.ActiveSessions
}

func agentSessions(ctx SectionContext, agent string) []sessionRowPayload {
	if ctx.Data == nil || ctx.Data.Sessions == nil {
		return nil
	}
	var matches []sessionRowPayload
	for _, o := range ctx.Data.Sessions.RecentSessions {
		if strings.EqualFold(strings.TrimSpace(o.Agent), strings.TrimSpace(agent)) {
			matches = append(matches, sessionRowPayload{
				ID:           o.ID,
				ProjectPath:  o.ProjectPath,
				TotalTokens:  o.TotalTokens,
				LastActivity: o.LastActivity,
			})
		}
	}
	return matches
}

// sessionRowPayload is a narrowed view used only by the Agents detail
// pane so the renderer doesn't have to reach back into session.* types.
type sessionRowPayload struct {
	ID           string
	ProjectPath  string
	TotalTokens  int
	LastActivity time.Time
}
