// Package tui provides a Terminal User Interface using BubbleTea.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI application state.
type Model struct {
	tabs      []string
	activeTab int
	savings   int64
	commands  int64
	width     int
	height    int
	history   []CommandEntry
	err       error
}

// CommandEntry represents a command in history.
type CommandEntry struct {
	Command      string
	OriginalSize int
	Compressed   int
	Saved        int
	Timestamp    time.Time
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{
		tabs: []string{"Dashboard", "History", "Stats", "Settings"},
		history: []CommandEntry{
			{Command: "cargo build", OriginalSize: 15000, Compressed: 2000, Saved: 13000, Timestamp: time.Now().Add(-5 * time.Minute)},
			{Command: "git status", OriginalSize: 500, Compressed: 150, Saved: 350, Timestamp: time.Now().Add(-10 * time.Minute)},
			{Command: "npm test", OriginalSize: 8000, Compressed: 800, Saved: 7200, Timestamp: time.Now().Add(-15 * time.Minute)},
		},
		savings:  20550,
		commands: 3,
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
	)
}

// Update handles messages and updates state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		}

	case tickMsg:
		return m, tickCmd()

	case errorMsg:
		m.err = msg.err
	}

	return m, nil
}

// View renders the TUI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n\n")

	// Tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	// Content based on active tab
	s.WriteString(m.renderContent())
	s.WriteString("\n\n")

	// Footer
	s.WriteString(m.renderFooter())

	return s.String()
}

func (m Model) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D4AA")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(1, 2).
		Width(m.width)

	return style.Render("  TOKMAN - Token Optimizer Dashboard  ")
}

func (m Model) renderTabs() string {
	var tabs []string
	for i, tab := range m.tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1)

		if i == m.activeTab {
			style = style.
				Foreground(lipgloss.Color("#1a1a2e")).
				Background(lipgloss.Color("#00D4AA")).
				Bold(true)
		} else {
			style = style.
				Foreground(lipgloss.Color("#888888")).
				Background(lipgloss.Color("#2a2a3e"))
		}

		tabs = append(tabs, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case 0:
		return m.renderDashboard()
	case 1:
		return m.renderHistory()
	case 2:
		return m.renderStats()
	case 3:
		return m.renderSettings()
	default:
		return "Unknown tab"
	}
}

func (m Model) renderDashboard() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1).
		Width(m.width - 4)

	var content strings.Builder

	// Stats cards row
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1, 2).
		Width((m.width - 8) / 3)

	card1 := cardStyle.Render(fmt.Sprintf("Tokens Saved\n\n%s%d",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")).Bold(true).Render("→ "),
		m.savings))

	card2 := cardStyle.Render(fmt.Sprintf("Commands\n\n%s%d",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")).Bold(true).Render("→ "),
		m.commands))

	savingsRate := float64(m.savings) / float64(m.savings+5000) * 100
	card3 := cardStyle.Render(fmt.Sprintf("Savings Rate\n\n%s%.1f%%",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")).Bold(true).Render("→ "),
		savingsRate))

	row := lipgloss.JoinHorizontal(lipgloss.Top, card1, card2, card3)
	content.WriteString(row)
	content.WriteString("\n\n")

	// Recent activity title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Recent Activity"))
	content.WriteString("\n\n")

	// Recent commands table
	for _, entry := range m.history[:min(3, len(m.history))] {
		line := fmt.Sprintf("• %-20s %6d → %6d tokens  %s %d saved\n",
			entry.Command,
			entry.OriginalSize,
			entry.Compressed,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")).Render("▲"),
			entry.Saved,
		)
		content.WriteString(line)
	}

	return style.Render(content.String())
}

func (m Model) renderHistory() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1).
		Width(m.width - 4)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Command History\n\n"))

	// Table header
	header := fmt.Sprintf("%-20s %10s %10s %10s %15s\n",
		"Command", "Original", "Compressed", "Saved", "Time")
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888888")).Render(header))
	content.WriteString(strings.Repeat("─", m.width-8) + "\n")

	// Table rows
	for _, entry := range m.history {
		line := fmt.Sprintf("%-20s %10d %10d %10d %15s\n",
			truncate(entry.Command, 20),
			entry.OriginalSize,
			entry.Compressed,
			entry.Saved,
			entry.Timestamp.Format("15:04"),
		)
		content.WriteString(line)
	}

	return style.Render(content.String())
}

func (m Model) renderStats() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1).
		Width(m.width - 4)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Detailed Statistics\n\n"))

	// Layer statistics
	layers := []struct {
		Name      string
		Savings   int64
		Calls     int
		AvgTimeMs int
	}{
		{"Entropy Filter", 5200, 15, 5},
		{"Perplexity Filter", 4800, 12, 8},
		{"AST Parser", 6100, 8, 15},
		{"Budget Enforcement", 4450, 20, 2},
	}

	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888888")).
		Render(fmt.Sprintf("%-25s %10s %10s %10s\n", "Layer", "Savings", "Calls", "Avg Time")))
	content.WriteString(strings.Repeat("─", m.width-8) + "\n")

	for _, layer := range layers {
		line := fmt.Sprintf("%-25s %10d %10d %10dms\n",
			layer.Name,
			layer.Savings,
			layer.Calls,
			layer.AvgTimeMs,
		)
		content.WriteString(line)
	}

	return style.Render(content.String())
}

func (m Model) renderSettings() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1).
		Width(m.width - 4)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Settings\n\n"))

	settings := []struct {
		Name  string
		Value string
	}{
		{"Max Tokens", "4000"},
		{"Budget Mode", "adaptive"},
		{"Auto-detect Language", "true"},
		{"Compress on Save", "true"},
		{"Show Savings Tip", "true"},
	}

	for _, s := range settings {
		line := fmt.Sprintf("%-25s %s\n", s.Name, lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")).Render(s.Value))
		content.WriteString(line)
	}

	return style.Render(content.String())
}

func (m Model) renderFooter() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	return style.Render("Press q to quit • tab/arrow keys to switch tabs")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Messages
type tickMsg time.Time
type errorMsg struct{ err error }

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Run starts the TUI application.
func Run() error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}
