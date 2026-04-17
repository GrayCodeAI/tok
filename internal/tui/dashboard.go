package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Tab definitions for all features
type Tab int

const (
	OverviewTab Tab = iota
	CommandsTab
	CacheTab
	FilterLayersTab
	AnalyticsTab
	SessionsTab
	EconomicsTab
	ConfigTab
	LogsTab
	SystemTab
	TabCount
)

func (t Tab) String() string {
	names := []string{
		"Overview",
		"Commands",
		"Cache",
		"Layers",
		"Analytics",
		"Sessions",
		"Economics",
		"Config",
		"Logs",
		"System",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// FilterLayer represents a filter layer with its state
type FilterLayer struct {
	ID          int
	Name        string
	Description string
	Enabled     bool
	Research    string
	Stats       LayerStats
}

// LayerStats holds statistics for a layer
type LayerStats struct {
	Processed int64
	Saved     int64
	AvgTime   time.Duration
}

// Session represents an active session
type Session struct {
	ID           string
	StartTime    time.Time
	CommandCount int
	TokensSaved  int64
	Status       string
}

// LogEntry represents a log entry
type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

// Model represents the comprehensive TUI state
type Model struct {
	// Core
	tabs      []string
	activeTab Tab
	width     int
	height    int
	ready     bool
	loading   bool

	// Components
	spinner   spinner.Model
	progress  progress.Model
	help      help.Model
	keys      keyMap

	// Data
	stats      Stats
	commands   []CommandEntry
	layers     []FilterLayer
	sessions   []Session
	logs       []LogEntry
	lastUpdate time.Time

	// Tables
	cmdTable   table.Model
	layerList  list.Model
	sessionTable table.Model
	logViewport viewport.Model

	// Inputs
	searchInput textinput.Model

	// Charts data
	dailySavings []DataPoint
	topCommands  []CommandStat
}

// DataPoint for charts
type DataPoint struct {
	Label string
	Value int64
}

// CommandStat for top commands
type CommandStat struct {
	Command string
	Count   int
	Saved   int64
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
	AvgSavings       float64
	TotalCostSaved   float64
	CommandsPerHour  float64
}

// CommandEntry for table
type CommandEntry struct {
	Time    string
	Command string
	Input   string
	Output  string
	Saved   string
	Percent string
	Agent   string
}

// Messages
type tickMsg time.Time
type updateMsg struct {
	stats      Stats
	commands   []CommandEntry
	layers     []FilterLayer
	sessions   []Session
	logs       []LogEntry
	dailyData  []DataPoint
	topCmds    []CommandStat
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Toggle   key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next tab"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Refresh, k.Search, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.ShiftTab, k.Refresh, k.Search},
		{k.Toggle, k.Enter, k.Help, k.Quit},
	}
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab) || key.Matches(msg, m.keys.Right):
			m.activeTab = Tab((int(m.activeTab) + 1) % int(TabCount))
		case key.Matches(msg, m.keys.ShiftTab) || key.Matches(msg, m.keys.Left):
			m.activeTab = Tab((int(m.activeTab) - 1 + int(TabCount)) % int(TabCount))
		case key.Matches(msg, m.keys.Refresh):
			cmds = append(cmds, fetchDataCmd())
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Search):
			if m.activeTab == CommandsTab {
				m.searchInput.Focus()
			}
		case key.Matches(msg, m.keys.Toggle):
			if m.activeTab == FilterLayersTab {
				m.toggleSelectedLayer()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateComponents()

	case tickMsg:
		cmds = append(cmds, tickCmd(), fetchDataCmd())

	case updateMsg:
		m.stats = msg.stats
		m.commands = msg.commands
		m.layers = msg.layers
		m.sessions = msg.sessions
		m.logs = msg.logs
		m.dailySavings = msg.dailyData
		m.topCommands = msg.topCmds
		m.lastUpdate = time.Now()
		m.loading = false
		m.updateComponents()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Update sub-components based on active tab
		switch m.activeTab {
		case CommandsTab:
			m.cmdTable, cmd = m.cmdTable.Update(msg)
			cmds = append(cmds, cmd)
		case FilterLayersTab:
			m.layerList, cmd = m.layerList.Update(msg)
			cmds = append(cmds, cmd)
		case SessionsTab:
			m.sessionTable, cmd = m.sessionTable.Update(msg)
			cmds = append(cmds, cmd)
		case LogsTab:
			m.logViewport, cmd = m.logViewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) toggleSelectedLayer() {
	if len(m.layers) == 0 {
		return
	}
	idx := m.layerList.Index()
	if idx >= 0 && idx < len(m.layers) {
		m.layers[idx].Enabled = !m.layers[idx].Enabled
	}
}

func (m *Model) updateComponents() {
	// Update command table
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

	m.cmdTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Update layer list
	layerItems := []list.Item{}
	for _, layer := range m.layers {
		status := "✓"
		if !layer.Enabled {
			status = "✗"
		}
		layerItems = append(layerItems, layerItem{
			title:       fmt.Sprintf("%s %s", status, layer.Name),
			description: layer.Description,
		})
	}
	m.layerList = list.New(layerItems, list.NewDefaultDelegate(), m.width-4, 15)
	m.layerList.Title = "Filter Layers (Space to toggle)"

	// Update session table
	sessionColumns := []table.Column{
		{Title: "Session ID", Width: 20},
		{Title: "Started", Width: 12},
		{Title: "Commands", Width: 10},
		{Title: "Tokens Saved", Width: 12},
		{Title: "Status", Width: 10},
	}

	sessionRows := []table.Row{}
	for _, s := range m.sessions {
		sessionRows = append(sessionRows, table.Row{
			s.ID,
			s.StartTime.Format("15:04"),
			fmt.Sprintf("%d", s.CommandCount),
			formatTokens(int(s.TokensSaved)),
			s.Status,
		})
	}

	m.sessionTable = table.New(
		table.WithColumns(sessionColumns),
		table.WithRows(sessionRows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Update log viewport
	var logContent strings.Builder
	for _, log := range m.logs {
		levelStyle := TextSecondaryStyle
		switch log.Level {
		case "ERROR":
			levelStyle = ErrorStyle
		case "WARN":
			levelStyle = WarningStyle
		case "INFO":
			levelStyle = InfoStyle
		case "SUCCESS":
			levelStyle = SuccessStyle
		}
		logContent.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			log.Time.Format("15:04:05"),
			levelStyle.Render(log.Level),
			log.Message))
	}
	m.logViewport = viewport.New(m.width-4, 15)
	m.logViewport.SetContent(logContent.String())
}

// View renders the TUI with single accent color
func (m Model) View() string {
	if !m.ready {
		return "\n  " + m.spinner.View() + " Initializing TokMan Dashboard..."
	}

	var b strings.Builder

	// Title with icon (using ASCII art instead of emoji)
	title := TitleStyle.Render(" [+] TOKMAN DASHBOARD ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content based on active tab - wrapped in unified box
	var content string
	switch m.activeTab {
	case OverviewTab:
		content = m.renderOverview()
	case CommandsTab:
		content = m.renderCommands()
	case CacheTab:
		content = m.renderCache()
	case FilterLayersTab:
		content = m.renderFilterLayers()
	case AnalyticsTab:
		content = m.renderAnalytics()
	case SessionsTab:
		content = m.renderSessions()
	case EconomicsTab:
		content = m.renderEconomics()
	case ConfigTab:
		content = m.renderConfig()
	case LogsTab:
		content = m.renderLogs()
	case SystemTab:
		content = m.renderSystem()
	}

	// Wrap content in unified box
	b.WriteString(BoxStyle.Render(content))

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderTabs() string {
	var tabs []string
	for i := 0; i < int(TabCount); i++ {
		tabName := Tab(i).String()
		if i == int(m.activeTab) {
			tabs = append(tabs, TabActiveStyle.Render(" "+tabName+" "))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(" "+tabName+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m Model) renderOverview() string {
	// Single column layout with accent color
	var b strings.Builder

	b.WriteString(AccentStyle.Render("SYSTEM STATS"))
	b.WriteString("\n\n")

	// Stats in a clean list format
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Total Commands:", StatValueStyle.Render(fmt.Sprintf("%d", m.stats.TotalCommands))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Tokens Saved:", StatValueStyle.Render(formatTokens(int(m.stats.TotalSaved)))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Cache Hit Rate:", StatValueStyle.Render(fmt.Sprintf("%.1f%%", m.stats.HitRate))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Active Sessions:", StatValueStyle.Render(fmt.Sprintf("%d", m.stats.ActiveSessions))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Avg Savings:", StatValueStyle.Render(fmt.Sprintf("%.1f%%", m.stats.AvgSavings))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Cost Saved:", StatValueStyle.Render(fmt.Sprintf("$%.2f", m.stats.TotalCostSaved))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Commands/Hour:", StatValueStyle.Render(fmt.Sprintf("%.1f", m.stats.CommandsPerHour))))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", "Today's Savings:", StatValueStyle.Render(formatTokens(int(m.stats.TodaySaved)))))

	if m.stats.TopCommand != "" {
		b.WriteString("\n")
		b.WriteString(AccentStyle.Render("MOST USED"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  > %s", m.stats.TopCommand))
	}

	return b.String()
}

func (m Model) renderCommands() string {
	header := HeaderStyle.Render(" RECENT COMMANDS ")

	if len(m.commands) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			BoxStyle.Render("No commands recorded yet"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.cmdTable.View(),
		"",
		TextMutedStyle.Render("Press '/' to search, 'r' to refresh"),
	)
}

func (m Model) renderCache() string {
	header := HeaderStyle.Render(" CACHE STATISTICS ")

	hits := StatValueStyle.Render(formatNumber(m.stats.CacheHits))
	misses := StatValueStyle.Render(formatNumber(m.stats.CacheMisses))

	hitBox := BoxStyle.Render(
		StatLabelStyle.Render("CACHE HITS") + "\n" + hits,
	)

	missBox := BoxDimStyle.Render(
		StatLabelStyle.Render("CACHE MISSES") + "\n" + misses,
	)

	total := m.stats.CacheHits + m.stats.CacheMisses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(m.stats.CacheHits) / float64(total) * 100
	}

	rateBox := BoxActiveStyle.Render(
		StatLabelStyle.Render("HIT RATE") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%.1f%%", hitRate)),
	)

	// Visual bar
	bar := m.renderProgressBar("Cache Efficiency", int64(hitRate), 100)

	// Cache features - uniform primary accent
	features := []string{
		SuccessStyle.Render("✓") + " Semantic caching enabled",
		PrimaryStyle.Render("✓") + " KV cache alignment",
		PrimaryStyle.Render("✓") + " Fingerprint-based deduplication",
		PrimaryStyle.Render("✓") + " Automatic cleanup (90 days)",
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, hitBox, missBox, rateBox),
		"",
		bar,
		"",
		BoxStyle.Render(strings.Join(features, "\n")),
	)
}

func (m Model) renderFilterLayers() string {
	header := HeaderStyle.Render(" 20-LAYER FILTER PIPELINE ")

	if len(m.layers) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			BoxStyle.Render("Loading filter layers..."),
		)
	}

	// Layer statistics
	var enabledCount int
	var totalSaved int64
	for _, l := range m.layers {
		if l.Enabled {
			enabledCount++
		}
		totalSaved += l.Stats.Saved
	}

	statusBox := BoxStyle.Render(
		StatLabelStyle.Render("LAYERS ACTIVE") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%d/%d", enabledCount, len(m.layers))),
	)

	savedBox := BoxActiveStyle.Render(
		StatLabelStyle.Render("TOKENS SAVED") + "\n" +
			StatValueStyle.Render(formatTokens(int(totalSaved))),
	)

	// Legend
	legend := TextMutedStyle.Render("Space: toggle | Enter: details | r: reset all")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, statusBox, savedBox),
		"",
		m.layerList.View(),
		"",
		legend,
	)
}

func (m Model) renderAnalytics() string {
	header := HeaderStyle.Render(" ANALYTICS & INSIGHTS ")

	// Simple ASCII bar chart for top commands - uniform primary color
	var chart strings.Builder
	chart.WriteString(PrimaryStyle.Render("Top Commands:\n\n"))

	maxCount := 0
	for _, cmd := range m.topCommands {
		if cmd.Count > maxCount {
			maxCount = cmd.Count
		}
	}

	for _, cmd := range m.topCommands {
		barWidth := 0
		if maxCount > 0 {
			barWidth = (cmd.Count * 20) / maxCount
		}
		bar := BarFullStyle.Render(strings.Repeat("█", barWidth))
		chart.WriteString(fmt.Sprintf("%-15s %s %d\n", cmd.Command, bar, cmd.Count))
	}

	// Daily savings chart - uniform primary color
	var dailyChart strings.Builder
	dailyChart.WriteString(PrimaryStyle.Render("\nDaily Savings (last 7 days):\n\n"))

	maxDaily := int64(0)
	for _, d := range m.dailySavings {
		if d.Value > maxDaily {
			maxDaily = d.Value
		}
	}

	for _, d := range m.dailySavings {
		barWidth := 0
		if maxDaily > 0 {
			barWidth = int((d.Value * 20) / maxDaily)
		}
		bar := BarFullStyle.Render(strings.Repeat("█", barWidth))
		dailyChart.WriteString(fmt.Sprintf("%-10s %s %s\n", d.Label, bar, formatTokens(int(d.Value))))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		BoxStyle.Render(chart.String()),
		BoxStyle.Render(dailyChart.String()),
	)
}

func (m Model) renderSessions() string {
	header := HeaderStyle.Render(" SESSION MANAGEMENT ")

	if len(m.sessions) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			BoxDimStyle.Render("No active sessions"),
		)
	}

	stats := BoxActiveStyle.Render(
		StatLabelStyle.Render("ACTIVE SESSIONS") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("%d", len(m.sessions))),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		stats,
		"",
		m.sessionTable.View(),
	)
}

func (m Model) renderEconomics() string {
	header := HeaderStyle.Render(" COST ANALYSIS ")

	// Cost breakdown - uniform style
	totalSaved := BoxStyle.Render(
		StatLabelStyle.Render("TOTAL COST SAVED") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("$%.2f", m.stats.TotalCostSaved)),
	)

	avgPerCmd := 0.0
	if m.stats.TotalCommands > 0 {
		avgPerCmd = m.stats.TotalCostSaved / float64(m.stats.TotalCommands)
	}

	avgCost := BoxActiveStyle.Render(
		StatLabelStyle.Render("AVG PER COMMAND") + "\n" +
			StatValueStyle.Render(fmt.Sprintf("$%.4f", avgPerCmd)),
	)

	// Pricing tiers - uniform primary accent
	tiers := []string{
		HeaderStyle.Render(" PRICING TIERS "),
		"",
		PrimaryStyle.Render("GPT-4/Claude 3:") + " ~$0.03/1K tokens",
		PrimaryStyle.Render("GPT-3.5:") + " ~$0.002/1K tokens",
		SuccessStyle.Render("Local LLM:") + " $0 (compute only)",
		"",
		HeaderStyle.Render(" YOUR SAVINGS "),
		"",
		SuccessStyle.Render(fmt.Sprintf("✓ Saved %s tokens", formatTokens(int(m.stats.TotalSaved)))),
		SuccessStyle.Render(fmt.Sprintf("✓ Equivalent to $%.2f", m.stats.TotalCostSaved)),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, totalSaved, avgCost),
		"",
		BoxStyle.Render(strings.Join(tiers, "\n")),
	)
}

func (m Model) renderConfig() string {
	header := HeaderStyle.Render(" CONFIGURATION ")

	// Config sections - uniform primary accent
	sections := []string{
		HeaderStyle.Render(" PIPELINE SETTINGS "),
		"",
		PrimaryStyle.Render("Max Context:") + " 2M tokens",
		PrimaryStyle.Render("Chunk Size:") + " 100K tokens",
		PrimaryStyle.Render("Stream Threshold:") + " 500K tokens",
		"",
		HeaderStyle.Render(" FEATURE FLAGS "),
		"",
		SuccessStyle.Render("✓") + " Semantic caching",
		PrimaryStyle.Render("✓") + " LLM compaction",
		PrimaryStyle.Render("✓") + " SIMD optimizations",
		PrimaryStyle.Render("✓") + " Telemetry batching",
		"",
		HeaderStyle.Render(" PATHS "),
		"",
		TextSecondaryStyle.Render("Config:") + " ~/.config/tokman/config.toml",
		TextSecondaryStyle.Render("Database:") + " ~/.local/share/tokman/tokman.db",
		TextSecondaryStyle.Render("Cache:") + " ~/.cache/tokman/",
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		BoxStyle.Render(strings.Join(sections, "\n")),
	)
}

func (m Model) renderLogs() string {
	header := HeaderStyle.Render(" REAL-TIME LOGS ")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		BoxStyle.Render(m.logViewport.View()),
	)
}

func (m Model) renderSystem() string {
	header := HeaderStyle.Render(" SYSTEM INFORMATION ")

	// System info - uniform primary accent
	info := []string{
		HeaderStyle.Render(" VERSION "),
		"",
		PrimaryStyle.Render("TokMan:") + " v0.28.0",
		PrimaryStyle.Render("Go Version:") + " 1.26",
		PrimaryStyle.Render("Platform:") + " " + os.Getenv("GOOS") + "/" + os.Getenv("GOARCH"),
		"",
		HeaderStyle.Render(" PERFORMANCE "),
		"",
		SuccessStyle.Render("✓") + " Caching enabled",
		PrimaryStyle.Render("✓") + " Telemetry batching",
		PrimaryStyle.Render("✓") + " SIMD optimizations",
		PrimaryStyle.Render("✓") + " 20-layer pipeline active",
		"",
		HeaderStyle.Render(" RESEARCH FOUNDATION "),
		"",
		TextSecondaryStyle.Render("Based on 120+ papers from:"),
		PrimaryStyle.Render("•") + " Microsoft Research (LLMLingua, LongLLMLingua)",
		PrimaryStyle.Render("•") + " Stanford/Berkeley (Gist Compression)",
		PrimaryStyle.Render("•") + " Princeton/MIT (AutoCompressor)",
		PrimaryStyle.Render("•") + " UC Berkeley (MemGPT)",
		PrimaryStyle.Render("•") + " NeurIPS 2023 (H2O Filter)",
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		BoxStyle.Render(strings.Join(info, "\n")),
	)
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
	// Help view
	helpView := m.help.View(m.keys)

	timeStr := m.lastUpdate.Format("15:04:05")
	if m.lastUpdate.IsZero() {
		timeStr = "--:--:--"
	}

	return FooterStyle.Render(
		helpView +
			"  |  " + TextMutedStyle.Render("Updated: "+timeStr),
	)
}

// layerItem for the list
type layerItem struct {
	title       string
	description string
}

func (i layerItem) Title() string       { return i.title }
func (i layerItem) Description() string { return i.description }
func (i layerItem) FilterValue() string { return i.title }

// Commands
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		// Generate demo data
		stats := Stats{
			TotalCommands:   1234,
			TotalSaved:      5678900,
			TodaySaved:      45000,
			CacheHits:       10000,
			CacheMisses:     500,
			HitRate:         95.2,
			ActiveSessions:  3,
			TopCommand:      "git status",
			AvgSavings:      82.5,
			TotalCostSaved:  170.37,
			CommandsPerHour: 45.2,
		}

		commands := []CommandEntry{
			{Time: "10:42", Command: "git status", Input: "2.1K", Output: "420", Saved: "1.7K", Percent: "81%", Agent: "Claude"},
			{Time: "10:41", Command: "cargo test", Input: "45K", Output: "5K", Saved: "40K", Percent: "89%", Agent: "Claude"},
			{Time: "10:40", Command: "npm ls", Input: "12K", Output: "2K", Saved: "10K", Percent: "83%", Agent: "Cursor"},
			{Time: "10:38", Command: "docker ps", Input: "3.5K", Output: "700", Saved: "2.8K", Percent: "80%", Agent: "Claude"},
			{Time: "10:35", Command: "ls -la", Input: "800", Output: "160", Saved: "640", Percent: "80%", Agent: "Copilot"},
		}

		// Filter layers based on the 20-layer pipeline
		layers := []FilterLayer{
			{ID: 1, Name: "Entropy Filtering", Description: "Remove low-information tokens (Mila 2023)", Enabled: true, Research: "Selective Context"},
			{ID: 2, Name: "Perplexity Pruning", Description: "Iterative token removal (Microsoft/Tsinghua)", Enabled: true, Research: "LLMLingua"},
			{ID: 3, Name: "Goal-Driven Selection", Description: "CRF-style line scoring (Shanghai Jiao Tong)", Enabled: true, Research: "SWE-Pruner"},
			{ID: 4, Name: "AST Preservation", Description: "Syntax-aware compression (NUS)", Enabled: true, Research: "LongCodeZip"},
			{ID: 5, Name: "Contrastive Ranking", Description: "Question-relevance scoring (Microsoft)", Enabled: true, Research: "LongLLMLingua"},
			{ID: 6, Name: "N-gram Abbreviation", Description: "Lossless pattern compression", Enabled: true, Research: "CompactPrompt"},
			{ID: 7, Name: "Evaluator Heads", Description: "Early-layer attention sim (Tsinghua/Huawei)", Enabled: true, Research: "EHPC"},
			{ID: 8, Name: "Gist Compression", Description: "Virtual token embedding (Stanford/Berkeley)", Enabled: true, Research: "Gist"},
			{ID: 9, Name: "Hierarchical Summary", Description: "Recursive summarization (Princeton/MIT)", Enabled: true, Research: "AutoCompressor"},
			{ID: 10, Name: "Budget Enforcement", Description: "Strict token limits", Enabled: true, Research: "Industry"},
			{ID: 11, Name: "Compaction", Description: "Semantic compression (UC Berkeley)", Enabled: true, Research: "MemGPT"},
			{ID: 12, Name: "Attribution Filter", Description: "78% pruning (LinkedIn)", Enabled: true, Research: "ProCut"},
			{ID: 13, Name: "H2O Filter", Description: "30x+ compression (NeurIPS 2023)", Enabled: true, Research: "H2O"},
			{ID: 14, Name: "Attention Sink", Description: "Infinite context stability", Enabled: true, Research: "StreamingLLM"},
			{ID: 15, Name: "Meta-Token", Description: "27% lossless compression", Enabled: true, Research: "arXiv:2506.00307"},
			{ID: 16, Name: "Semantic Chunk", Description: "Context-aware boundaries", Enabled: true, Research: "ChunkKV"},
			{ID: 17, Name: "Semantic Cache", Description: "Reuse similar-context compression", Enabled: true, Research: "KVReviver"},
			{ID: 18, Name: "Lazy Pruner", Description: "2.34x speedup", Enabled: true, Research: "LazyLLM"},
			{ID: 19, Name: "Semantic Anchor", Description: "Context preservation", Enabled: true, Research: "Attention Gradient"},
			{ID: 20, Name: "Agent Memory", Description: "Knowledge graph extraction", Enabled: true, Research: "Focus"},
		}

		sessions := []Session{
			{ID: "sess_abc123", StartTime: time.Now().Add(-2 * time.Hour), CommandCount: 45, TokensSaved: 125000, Status: "active"},
			{ID: "sess_def456", StartTime: time.Now().Add(-5 * time.Hour), CommandCount: 128, TokensSaved: 340000, Status: "active"},
			{ID: "sess_ghi789", StartTime: time.Now().Add(-1 * time.Hour), CommandCount: 12, TokensSaved: 28000, Status: "active"},
		}

		logs := []LogEntry{
			{Time: time.Now().Add(-2 * time.Minute), Level: "INFO", Message: "Compressed git status: 2.1K -> 420 tokens (80% reduction)"},
			{Time: time.Now().Add(-3 * time.Minute), Level: "SUCCESS", Message: "Cache hit for fingerprint: abc123"},
			{Time: time.Now().Add(-5 * time.Minute), Level: "INFO", Message: "Layer 13 (H2O) saved 450 tokens"},
			{Time: time.Now().Add(-8 * time.Minute), Level: "WARN", Message: "High memory usage: 85%"},
			{Time: time.Now().Add(-10 * time.Minute), Level: "INFO", Message: "Session sess_abc123 started"},
		}

		// Daily savings data
		dailyData := []DataPoint{
			{Label: "Mon", Value: 42000},
			{Label: "Tue", Value: 58000},
			{Label: "Wed", Value: 35000},
			{Label: "Thu", Value: 72000},
			{Label: "Fri", Value: 45000},
			{Label: "Sat", Value: 28000},
			{Label: "Sun", Value: 39000},
		}

		// Top commands
		topCmds := []CommandStat{
			{Command: "git status", Count: 234, Saved: 450000},
			{Command: "cargo test", Count: 156, Saved: 890000},
			{Command: "npm ls", Count: 89, Saved: 234000},
			{Command: "docker ps", Count: 67, Saved: 123000},
			{Command: "ls -la", Count: 45, Saved: 34000},
		}

		// Sort by count
		sort.Slice(topCmds, func(i, j int) bool {
			return topCmds[i].Count > topCmds[j].Count
		})

		return updateMsg{
			stats:     stats,
			commands:  commands,
			layers:    layers,
			sessions:  sessions,
			logs:      logs,
			dailyData: dailyData,
			topCmds:   topCmds,
		}
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

	h := help.New()

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search commands..."
	ti.CharLimit = 50

	return Model{
		tabs: []string{
			"Overview", "Commands", "Cache", "Layers",
			"Analytics", "Sessions", "Economics", "Config", "Logs", "System",
		},
		spinner:     s,
		progress:    p,
		help:        h,
		keys:        newKeyMap(),
		loading:     true,
		searchInput: ti,
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
