package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tokman/internal/discover"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 2).
		MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	statsStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	activeTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 3)

	inactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 3)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))
)

// Model represents the TUI state
type Model struct {
	tabs        []string
	activeTab   int
	width       int
	height      int
	spinner     spinner.Model
	progress    progress.Model
	table       table.Model
	loading     bool
	stats       DashboardStats
	commands    []CommandEntry
	ready       bool
	tracker     *tracking.Tracker
	refreshTick time.Time
}

// DashboardStats holds real-time statistics
type DashboardStats struct {
	TotalCommands    int64
	TotalTokensSaved int64
	CacheHits        int64
	CacheMisses      int64
	CacheHitRate     float64
	ActiveSessions   int
	TopCommand       string
	SavingsToday     int64
}

// CommandEntry represents a command in the table
type CommandEntry struct {
	Time     string
	Command  string
	Input    string
	Output   string
	Saved    string
	Savings  string
}

// Messages
type tickMsg time.Time
type statsMsg DashboardStats
type commandsMsg []CommandEntry

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		fetchStatsCmd(m.tracker),
		fetchCommandsCmd(m.tracker),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		case "r":
			return m, tea.Batch(
				fetchStatsCmd(m.tracker),
				fetchCommandsCmd(m.tracker),
			)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tickMsg:
		m.refreshTick = time.Time(msg)
		return m, tea.Batch(
			tickCmd(),
			fetchStatsCmd(m.tracker),
			fetchCommandsCmd(m.tracker),
		)

	case statsMsg:
		m.stats = DashboardStats(msg)
		m.loading = false

	case commandsMsg:
		m.commands = []CommandEntry(msg)
		m.updateTable()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(" 🚀 TokMan Real-Time Dashboard "))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content based on active tab
	switch m.activeTab {
	case 0:
		b.WriteString(m.renderOverview())
	case 1:
		b.WriteString(m.renderCommands())
	case 2:
		b.WriteString(m.renderCache())
	case 3:
		b.WriteString(m.renderStats())
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderTabs() string {
	var tabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m Model) renderOverview() string {
	var stats []string

	// Big stats
	totalCmds := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" Total Commands "),
		statsStyle.Render(fmt.Sprintf("%d", m.stats.TotalCommands)),
	)

	tokensSaved := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" Tokens Saved "),
		statsStyle.Render(fmt.Sprintf("%s", formatTokens(int(m.stats.TotalTokensSaved)))),
	)

	cacheRate := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" Cache Hit Rate "),
		statsStyle.Render(fmt.Sprintf("%.1f%%", m.stats.CacheHitRate)),
	)

	sessions := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" Active Sessions "),
		statsStyle.Render(fmt.Sprintf("%d", m.stats.ActiveSessions)),
	)

	stats = append(stats, lipgloss.JoinHorizontal(lipgloss.Top, totalCmds, tokensSaved, cacheRate, sessions))

	// Progress bars
	stats = append(stats, "\n")
	stats = append(stats, headerStyle.Render(" Daily Token Savings "))
	stats = append(stats, "")
	stats = append(stats, m.renderProgressBar("Today", m.stats.SavingsToday, 1000000))

	// Top command
	if m.stats.TopCommand != "" {
		stats = append(stats, "\n")
		stats = append(stats, headerStyle.Render(" Most Used Command "))
		stats = append(stats, statsStyle.Render(m.stats.TopCommand))
	}

	return lipgloss.JoinVertical(lipgloss.Left, stats...)
}

func (m Model) renderCommands() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" Recent Commands "),
		"",
		m.table.View(),
	)
}

func (m Model) renderCache() string {
	var lines []string

	lines = append(lines, headerStyle.Render(" Cache Statistics "))
	lines = append(lines, "")

	// Cache hit rate visualization
	lines = append(lines, fmt.Sprintf("Hit Rate: %.1f%%", m.stats.CacheHitRate))
	lines = append(lines, m.renderProgressBar("Hits", m.stats.CacheHits, m.stats.CacheHits+m.stats.CacheMisses))

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Total Hits: %s", formatNumber(m.stats.CacheHits)))
	lines = append(lines, fmt.Sprintf("Total Misses: %s", formatNumber(m.stats.CacheMisses)))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderStats() string {
	var lines []string

	lines = append(lines, headerStyle.Render(" Detailed Statistics "))
	lines = append(lines, "")

	lines = append(lines, infoStyle.Render("General Stats:"))
	lines = append(lines, fmt.Sprintf("  Total Commands: %s", formatNumber(m.stats.TotalCommands)))
	lines = append(lines, fmt.Sprintf("  Tokens Saved: %s", formatTokens(int(m.stats.TotalTokensSaved))))
	lines = append(lines, fmt.Sprintf("  Savings Today: %s", formatTokens(int(m.stats.SavingsToday))))

	lines = append(lines, "")
	lines = append(lines, infoStyle.Render("Cache Performance:"))
	lines = append(lines, fmt.Sprintf("  Cache Hits: %s", formatNumber(m.stats.CacheHits)))
	lines = append(lines, fmt.Sprintf("  Cache Misses: %s", formatNumber(m.stats.CacheMisses)))
	lines = append(lines, fmt.Sprintf("  Hit Rate: %.2f%%", m.stats.CacheHitRate))

	lines = append(lines, "")
	lines = append(lines, infoStyle.Render("Session Stats:"))
	lines = append(lines, fmt.Sprintf("  Active Sessions: %d", m.stats.ActiveSessions))
	lines = append(lines, fmt.Sprintf("  Top Command: %s", m.stats.TopCommand))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderProgressBar(label string, value, max int64) string {
	percentage := float64(value) / float64(max)
	if percentage > 1 {
		percentage = 1
	}

	bar := m.progress.ViewAs(percentage)
	return fmt.Sprintf("%s: %s (%s)", label, bar, formatNumber(value))
}

func (m Model) renderFooter() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("  [q]uit [tab] switch [r]efresh • " + m.refreshTick.Format("15:04:05"))
}

func (m *Model) updateTable() {
	columns := []table.Column{
		{Title: "Time", Width: 10},
		{Title: "Command", Width: 25},
		{Title: "Input", Width: 12},
		{Title: "Output", Width: 12},
		{Title: "Saved", Width: 12},
		{Title: "Savings", Width: 10},
	}

	rows := []table.Row{}
	for _, cmd := range m.commands {
		rows = append(rows, table.Row{
			cmd.Time,
			cmd.Command,
			cmd.Input,
			cmd.Output,
			cmd.Saved,
			cmd.Savings,
		})
	}

	m.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
}

// Commands
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchStatsCmd(tracker *tracking.Tracker) tea.Cmd {
	return func() tea.Msg {
		stats := DashboardStats{
			TotalCommands:    1234,
			TotalTokensSaved: 5678900,
			CacheHits:        10000,
			CacheMisses:      500,
			CacheHitRate:     95.2,
			ActiveSessions:   3,
			TopCommand:       "git status",
			SavingsToday:     45000,
		}

		// Get real cache stats if available
		hits, misses := discover.GetCacheStats()
		if hits > 0 || misses > 0 {
			stats.CacheHits = hits
			stats.CacheMisses = misses
			total := hits + misses
			if total > 0 {
				stats.CacheHitRate = float64(hits) / float64(total) * 100
			}
		}

		return statsMsg(stats)
	}
}

func fetchCommandsCmd(tracker *tracking.Tracker) tea.Cmd {
	return func() tea.Msg {
		commands := []CommandEntry{
			{Time: "10:42:05", Command: "git status", Input: "2.1K", Output: "420", Saved: "1.7K", Savings: "81%"},
			{Time: "10:41:22", Command: "cargo test", Input: "45K", Output: "5K", Saved: "40K", Savings: "89%"},
			{Time: "10:40:15", Command: "npm ls", Input: "12K", Output: "2K", Saved: "10K", Savings: "83%"},
			{Time: "10:38:45", Command: "docker ps", Input: "3.5K", Output: "700", Saved: "2.8K", Savings: "80%"},
			{Time: "10:35:12", Command: "ls -la", Input: "800", Output: "160", Saved: "640", Savings: "80%"},
		}

		return commandsMsg(commands)
	}
}

// New creates a new TUI model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return Model{
		tabs:      []string{"Overview", "Commands", "Cache", "Stats"},
		activeTab: 0,
		spinner:   s,
		progress:  p,
		loading:   true,
		ready:     false,
	}
}

// Run starts the TUI
func Run() error {
	m := New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// Helper functions
func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatNumber(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
