package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// logsSection renders the in-memory slog ring that the TUI root
// installs at startup. Filtering by level (d/i/w/e keys) + free-text
// filter via / lets users narrow to the events that matter.
//
// The section pulls a fresh snapshot from the ring every Update — the
// ring is thread-safe and O(N) to snapshot where N≤capacity (default
// 512), so polling per frame is cheap.
type logsSection struct {
	filter     string
	minLevel   slog.Level
	scroll     int // offset from bottom; 0 = latest entries visible
	lastSize   int
	lastFilter string
}

func newLogsSection() *logsSection {
	return &logsSection{minLevel: slog.LevelInfo}
}

func (s *logsSection) Name() string                { return "Logs" }
func (s *logsSection) Short() string               { return "Runtime" }
func (s *logsSection) Init(SectionContext) tea.Cmd { return nil }

func (s *logsSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "debug+")),
		key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "info+ (default)")),
		key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "warn+")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "error only")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear ring")),
		key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "scroll")),
		key.NewBinding(key.WithKeys("g", "G"), key.WithHelp("g/G", "top / bottom")),
	}
}

func (s *logsSection) Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	switch m := msg.(type) {
	case searchMsg:
		s.filter = m.Query
		s.scroll = 0
	case tea.KeyMsg:
		switch m.String() {
		case "d":
			s.minLevel = slog.LevelDebug
		case "i":
			s.minLevel = slog.LevelInfo
		case "w":
			s.minLevel = slog.LevelWarn
		case "e":
			s.minLevel = slog.LevelError
		case "c":
			if ctx.Logs != nil {
				ctx.Logs.Clear()
			}
		case "j", "down":
			if s.scroll > 0 {
				s.scroll--
			}
		case "k", "up":
			s.scroll++
		case "g", "home":
			// Scroll to the oldest (top of the ring).
			s.scroll = 100000
		case "G", "end":
			s.scroll = 0
		case "pgdown", "ctrl+f":
			s.scroll -= 10
			if s.scroll < 0 {
				s.scroll = 0
			}
		case "pgup", "ctrl+b":
			s.scroll += 10
		}
	}
	return s, nil
}

func (s *logsSection) View(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width
	if ctx.Logs == nil {
		return th.Muted.Render("Log ring not installed (TUI not active).")
	}

	entries := ctx.Logs.Snapshot()
	entries = filterEntries(entries, s.minLevel, s.filter)

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Logs"),
		th.Subtitle.Render(fmt.Sprintf("%d events shown · level ≥ %s · filter %q",
			len(entries), formatLevel(s.minLevel), s.filter)),
	)

	visibleRows := max(6, ctx.Height-lipgloss.Height(header)-4)
	if s.scroll > len(entries)-visibleRows {
		s.scroll = len(entries) - visibleRows
	}
	if s.scroll < 0 {
		s.scroll = 0
	}

	// Slice the window we render. Entries are oldest→newest so the
	// bottom of the display is "now". scroll=0 pins to the latest; a
	// positive scroll walks backwards in time.
	end := len(entries) - s.scroll
	if end < 0 {
		end = 0
	}
	start := end - visibleRows
	if start < 0 {
		start = 0
	}

	body := ""
	if end == 0 {
		body = th.Muted.Render("No events captured yet. The ring starts empty on TUI launch " +
			"and fills as tok logs events from refresh ticks, action runs, or internal " +
			"warnings. Try pressing 'r' to refresh — the loader emits a log entry per run.")
	} else {
		lines := make([]string, 0, end-start)
		for i := start; i < end; i++ {
			lines = append(lines, renderLogLine(th, entries[i], width))
		}
		body = strings.Join(lines, "\n")
	}

	bodyPanel := setWidth(panelStyle(th, 5), width).Render(body)

	footer := th.CardMeta.Render(
		fmt.Sprintf("d/i/w/e level · c clear · /%s filter · j/k scroll · g top · G bottom",
			fallback(s.filter, "…")),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", bodyPanel, "", footer)
}

func filterEntries(entries []LogEntry, minLevel slog.Level, query string) []LogEntry {
	q := strings.ToLower(strings.TrimSpace(query))
	out := entries[:0]
	for _, e := range entries {
		if e.Level < minLevel {
			continue
		}
		if q != "" && !entryMatches(e, q) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func entryMatches(e LogEntry, q string) bool {
	if strings.Contains(strings.ToLower(e.Message), q) {
		return true
	}
	for _, a := range e.Attrs {
		if strings.Contains(strings.ToLower(a), q) {
			return true
		}
	}
	return false
}

func renderLogLine(th theme, e LogEntry, width int) string {
	level := formatLevel(e.Level)
	levelStyle := th.Muted
	switch level {
	case "ERR":
		levelStyle = th.Danger
	case "WRN":
		levelStyle = th.Warning
	case "INF":
		levelStyle = th.Focus
	}
	ts := th.CardMeta.Render(e.Time.Format("15:04:05"))
	lvl := levelStyle.Render(level)
	// Reserve ~20 chars for timestamp+level+spacing, leave the rest for
	// the message+attrs. Truncate to prevent per-line width blowouts.
	remaining := width - 20
	if remaining < 20 {
		remaining = 20
	}
	tail := e.Message
	if len(e.Attrs) > 0 {
		tail += "  " + strings.Join(e.Attrs, " ")
	}
	tail = truncate(tail, remaining)
	return fmt.Sprintf("%s %s  %s", ts, lvl, tail)
}
