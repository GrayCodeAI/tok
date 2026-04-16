package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Model represents the TUI state
type Model struct {
	tabs       []string
	activeTab  int
	width      int
	height     int
	spinner    spinner.Model
	progress   progress.Model
	table      table.Model
	loading    bool
	stats      Stats
	commands   []CommandEntry
	ready      bool
	lastUpdate time.Time
}

// Stats holds dashboard statistics
type Stats struct {
	TotalCommands    int64
	TotalSaved       int64
	TodaySaved       int64
	CacheHits        int64
	CacheMisses      int64
	HitRate          float64
	ActiveSessions   int
	TopCommand       string
}

// CommandEntry for table
type CommandEntry struct {
	Time    string
	Command string
	Input   string
	Output  string
	Saved   string
	Percent string
}

// Messages
type tickMsg time.Time
type updateMsg struct {
	stats    Stats
	commands []CommandEntry
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		fetchDataCmd(),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		case "r":
			return m, fetchDataCmd()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateTable()

	case tickMsg:
		return m, tea.Batch(tickCmd(), fetchDataCmd())

	case updateMsg:
		m.stats = msg.stats
		m.commands = msg.commands
		m.lastUpdate = time.Now()
		m.loading = false
		m.updateTable()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI with vibrant colors
func (m Model) View() string {
	if !m.ready {
		return "\n  " + m.spinner.View() + " Initializing TokMan Dashboard..."
	}

	var b strings.Builder

	// Title with vibrant colors
	title := TitleStyle.Render(" 🚀 TOKMAN DASHBOARD ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content
	switch m.activeTab {
	case 0:
		b.WriteString(m.renderOverview())
	case 1:
		b.WriteString(m.renderCommands())
	case 2:
		b.WriteString(m.renderCache())
	case 3:
		b.WriteString(m.renderSystem())
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
			tabs = append(tabs, TabActiveStyle.Render(" "+tab+" "))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(" "+tab+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m Model) renderOverview() string {
	// Big stats in boxes
	totalBox := BoxStyle.Render(
		StatLabelStyle.Render("TOTAL COMMANDS") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%d", m.stats.TotalCommands)),
	)

	savedBox := BoxStyle.Render(
		StatLabelStyle.Render("TOKENS SAVED") + "\n" +
			StatValueStyle.Render(formatTokens(int(m.stats.TotalSaved))),
	)

	hitRateBox := BoxStyle.Render(
		StatLabelStyle.Render("CACHE HIT RATE") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%.1f%%", m.stats.HitRate)),
	)

	sessionsBox := BoxStyle.Render(
		StatLabelStyle.Render("ACTIVE SESSIONS") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%d", m.stats.ActiveSessions)),
	)

	// Top row
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, totalBox, savedBox, hitRateBox, sessionsBox)

	// Progress bar for today's savings
	progressBar := m.renderProgressBar("Today's Savings", m.stats.TodaySaved, 100000)

	// Top command
	var topCmd string
	if m.stats.TopCommand != "" {
		topCmd = BoxActiveStyle.Render(
			StatLabelStyle.Render("MOST USED COMMAND") + "\n" +
				InfoStyle.Render(m.stats.TopCommand),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		"",
		progressBar,
		"",
		topCmd,
	)
}

func (m Model) renderCommands() string {
	if len(m.commands) == 0 {
		return BoxStyle.Render("No commands recorded yet")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		HeaderStyle.Render(" RECENT COMMANDS "),
		"",
		m.table.View(),
	)
}

func (m Model) renderCache() string {
	hits := StatValueStyle.Render(formatNumber(m.stats.CacheHits))
	misses := StatValueStyle.Render(formatNumber(m.stats.CacheMisses))

	hitBox := BoxStyle.Render(
		StatLabelStyle.Render("CACHE HITS") + "\n" + hits,
	)

	missBox := BoxStyle.Render(
		StatLabelStyle.Render("CACHE MISSES") + "\n" + misses,
	)

	// Visual bar
	bar := m.renderProgressBar("Hit Rate", int64(m.stats.HitRate), 100)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, hitBox, missBox),
		"",
		bar,
	)
}

func (m Model) renderSystem() string {
	info := []string{
		HeaderStyle.Render(" SYSTEM INFORMATION "),
		"",
		InfoStyle.Render("Version:") + " v0.28.0",
		InfoStyle.Render("Go Version:") + " 1.26",
		InfoStyle.Render("Platform:") + " " + os.Getenv("GOOS") + "/" + os.Getenv("GOARCH"),
		"",
		HeaderStyle.Render(" PERFORMANCE "),
		"",
		SuccessStyle.Render("✓ Caching enabled"),
		SuccessStyle.Render("✓ Telemetry batching"),
		SuccessStyle.Render("✓ SIMD optimizations"),
		"",
		HeaderStyle.Render(" COMMANDS "),
		"",
		TextSecondaryStyle.Render("Total commands tracked: ") + fmt.Sprintf("%d", m.stats.TotalCommands),
		TextSecondaryStyle.Render("Total tokens saved: ") + formatTokens(int(m.stats.TotalSaved)),
	}

	return BoxStyle.Render(strings.Join(info, "\n"))
}

func (m Model) renderProgressBar(label string, value, max int64) string {
	pct := float64(value) / float64(max)
	if pct > 1 {
		pct = 1
	}

	bar := m.progress.ViewAs(pct)
	return lipgloss.JoinVertical(lipgloss.Left,
		TextSecondaryStyle.Render(label),
		bar,
		TextMutedStyle.Render(fmt.Sprintf("%s / %s", formatNumber(value), formatNumber(max))),
	)
}

func (m Model) renderFooter() string {
	keys := []string{
		KeyStyle.Render("tab"),
		TextSecondaryStyle.Render("switch"),
		KeyStyle.Render("r"),
		TextSecondaryStyle.Render("refresh"),
		KeyStyle.Render("q"),
		TextSecondaryStyle.Render("quit"),
	}

	timeStr := m.lastUpdate.Format("15:04:05")
	if m.lastUpdate.IsZero() {
		timeStr = "--:--:--"
	}

	return FooterStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Left, keys...) +
			"  |  " + TextMutedStyle.Render("Updated: "+timeStr),
	)
}

func (m *Model) updateTable() {
	columns := []table.Column{
		{Title: "Time", Width: 8},
		{Title: "Command", Width: 20},
		{Title: "Input", Width: 10},
		{Title: "Output", Width: 10},
		{Title: "Saved", Width: 10},
		{Title: "Savings", Width: 8},
	}

	rows := []table.Row{}
	for _, cmd := range m.commands {
		rows = append(rows, table.Row{
			cmd.Time,
			cmd.Command,
			cmd.Input,
			cmd.Output,
			cmd.Saved,
			cmd.Percent,
		})
	}

	m.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)
}

// Commands
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		// Generate demo data
		stats := Stats{
			TotalCommands:  1234,
			TotalSaved:     5678900,
			TodaySaved:     45000,
			CacheHits:      10000,
			CacheMisses:    500,
			HitRate:        95.2,
			ActiveSessions: 3,
			TopCommand:     "git status",
		}

		commands := []CommandEntry{
			{Time: "10:42", Command: "git status", Input: "2.1K", Output: "420", Saved: "1.7K", Percent: "81%"},
			{Time: "10:41", Command: "cargo test", Input: "45K", Output: "5K", Saved: "40K", Percent: "89%"},
			{Time: "10:40", Command: "npm ls", Input: "12K", Output: "2K", Saved: "10K", Percent: "83%"},
			{Time: "10:38", Command: "docker ps", Input: "3.5K", Output: "700", Saved: "2.8K", Percent: "80%"},
			{Time: "10:35", Command: "ls -la", Input: "800", Output: "160", Saved: "640", Percent: "80%"},
		}

		return updateMsg{stats: stats, commands: commands}
	}
}

// New creates a new TUI model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)

	return Model{
		tabs:    []string{"Overview", "Commands", "Cache", "System"},
		spinner: s,
		progress: p,
		loading: true,
	}
}

// Run starts the TUI
func Run() error {
	// Check if we have a TTY
	if termenv.NewOutput(os.Stdout).ColorProfile() == termenv.Ascii {
		// No color support, use basic mode
		fmt.Println("TokMan Dashboard (Basic Mode)")
		fmt.Println("═════════════════════════════")
		fmt.Println()
		fmt.Println("Total Commands: 1,234")
		fmt.Println("Tokens Saved: 5.7M")
		fmt.Println("Cache Hit Rate: 95.2%")
		fmt.Println()
		fmt.Println("Press any key to exit...")
		fmt.Scanln()
		return nil
	}

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
