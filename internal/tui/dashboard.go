package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ============================================================================
// WORLD-CLASS TUI DASHBOARD
// Inspired by: k9s, lazygit, grafana, htop
// Features: Purpose-driven colors, rich visualizations, professional layout
// ============================================================================

// Tab definitions
type Tab int

const (
	OverviewTab Tab = iota
	MetricsTab
	LayersTab
	AnalyticsTab
	SessionsTab
	EconomicsTab
	ConfigTab
	LogsTab
	TabCount
)

func (t Tab) String() string {
	names := []string{
		"[O]verview",
		"[M]etrics",
		"[L]ayers",
		"[A]nalytics",
		"[S]essions",
		"[E]conomics",
		"[C]onfig",
		"Lo[G]s",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

func (t Tab) Color() lipgloss.Style {
	colors := []lipgloss.Style{
		TabActive,      // Overview - Primary
		TabSuccess,     // Metrics - Success
		TabInfo,        // Layers - Info
		TabWarning,     // Analytics - Warning
		TabSuccessDim,  // Sessions - SuccessDim
		TabWarningDim,  // Economics - WarningDim
		TabInfoDim,     // Config - InfoDim
		TabErrorDim,    // Logs - ErrorDim
	}
	return colors[t]
}

// ============================================================================
// DATA MODELS
// ============================================================================

type DashboardStats struct {
	// Core metrics
	TotalCommands   int64
	TotalSaved      int64
	TodaySaved      int64
	AvgSavings      float64
	CommandsPerHour float64

	// Cache metrics
	CacheHits      int64
	CacheMisses    int64
	CacheHitRate   float64
	CacheSize      int64

	// Performance
	ActiveSessions int
	TopCommand     string
	TotalCostSaved float64

	// Real-time
	LastCommand    string
	LastSaved      int
	Uptime         time.Duration
}

type TimeSeriesPoint struct {
	Time  time.Time
	Value int64
}

type CommandStat struct {
	Command    string
	Count      int
	TokensSaved int64
	AvgInput   int
	AvgOutput  int
	Trend      float64 // -1 to 1
}

type FilterLayerInfo struct {
	ID          int
	Name        string
	Description string
	Research    string
	Enabled     bool
	Efficiency  float64 // 0-100%
	TokensSaved int64
	Status      string // active, idle, error
}

type SessionInfo struct {
	ID           string
	StartTime    time.Time
	Duration     time.Duration
	CommandCount int
	TokensSaved  int64
	Agent        string
	Status       string
}

type LogEntryInfo struct {
	Time    time.Time
	Level   string // INFO, SUCCESS, WARN, ERROR
	Source  string
	Message string
}

// ============================================================================
// MAIN MODEL
// ============================================================================

type DashboardModel struct {
	// Core state
	width       int
	height      int
	activeTab   Tab
	sidebarSel  int // Sidebar selection index
	ready       bool
	loading     bool
	showWelcome bool // Show welcome screen
	lastUpdate  time.Time

	// Components
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     keyMap

	// Data
	stats         DashboardStats
	dailyTrend    []TimeSeriesPoint
	hourlyTrend   []TimeSeriesPoint
	topCommands   []CommandStat
	layers        []FilterLayerInfo
	sessions      []SessionInfo
	logs          []LogEntryInfo

	// Tables
	cmdTable    table.Model
	layerList   list.Model
	sessionList list.Model
	logViewport viewport.Model

	// Views
	sidebar viewport.Model
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
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
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

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Refresh, k.Search, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.Refresh},
		{k.Search, k.Help, k.Quit},
	}
}

// ============================================================================
// TEA METHODS
// ============================================================================

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		fetchDashboardDataCmd(),
	)
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dismiss welcome screen on any key
		if m.showWelcome {
			m.showWelcome = false
			return m, nil
		}
		
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			// Navigate up in sidebar
			if m.sidebarSel > 0 {
				m.sidebarSel--
				m.activeTab = Tab(m.sidebarSel)
			}
		case key.Matches(msg, m.keys.Down):
			// Navigate down in sidebar
			if m.sidebarSel < int(TabCount)-1 {
				m.sidebarSel++
				m.activeTab = Tab(m.sidebarSel)
			}
		case key.Matches(msg, m.keys.Right):
			m.activeTab = Tab((int(m.activeTab) + 1) % int(TabCount))
			m.sidebarSel = int(m.activeTab)
		case key.Matches(msg, m.keys.Left):
			m.activeTab = Tab((int(m.activeTab) - 1 + int(TabCount)) % int(TabCount))
			m.sidebarSel = int(m.activeTab)
		case key.Matches(msg, m.keys.Tab):
			m.activeTab = Tab((int(m.activeTab) + 1) % int(TabCount))
			m.sidebarSel = int(m.activeTab)
		case key.Matches(msg, m.keys.ShiftTab):
			m.activeTab = Tab((int(m.activeTab) - 1 + int(TabCount)) % int(TabCount))
			m.sidebarSel = int(m.activeTab)
		case key.Matches(msg, m.keys.Refresh):
			cmds = append(cmds, fetchDashboardDataCmd())
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		default:
			// Direct tab access
			switch msg.String() {
			case "o":
				m.activeTab = OverviewTab
				m.sidebarSel = 0
			case "m":
				m.activeTab = MetricsTab
				m.sidebarSel = 1
			case "l":
				m.activeTab = LayersTab
				m.sidebarSel = 2
			case "a":
				m.activeTab = AnalyticsTab
				m.sidebarSel = 3
			case "s":
				m.activeTab = SessionsTab
				m.sidebarSel = 4
			case "e":
				m.activeTab = EconomicsTab
				m.sidebarSel = 5
			case "c":
				m.activeTab = ConfigTab
				m.sidebarSel = 6
			case "g":
				m.activeTab = LogsTab
				m.sidebarSel = 7
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateComponents()

	case tickMsg:
		cmds = append(cmds, tickCmd(), fetchDashboardDataCmd())

	case dashboardUpdateMsg:
		m.stats = msg.stats
		m.dailyTrend = msg.dailyTrend
		m.hourlyTrend = msg.hourlyTrend
		m.topCommands = msg.topCommands
		m.layers = msg.layers
		m.sessions = msg.sessions
		m.logs = msg.logs
		m.lastUpdate = time.Now()
		m.loading = false
		m.updateComponents()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *DashboardModel) updateComponents() {
	// Update command table
	columns := []table.Column{
		{Title: "Command", Width: 20},
		{Title: "Count", Width: 8},
		{Title: "Saved", Width: 12},
		{Title: "Trend", Width: 8},
	}
	rows := []table.Row{}
	for _, cmd := range m.topCommands {
		trend := "→"
		if cmd.Trend > 0.1 {
			trend = "↑"
		} else if cmd.Trend < -0.1 {
			trend = "↓"
		}
		rows = append(rows, table.Row{
			cmd.Command,
			fmt.Sprintf("%d", cmd.Count),
			formatTokens(int(cmd.TokensSaved)),
			trend,
		})
	}
	m.cmdTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)
}

// ============================================================================
// VIEW RENDERING
// ============================================================================

func (m DashboardModel) View() string {
	if !m.ready {
		return m.renderLoading()
	}

	// Show welcome screen first
	if m.showWelcome {
		return m.renderWelcome()
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Main content area
	b.WriteString(m.renderMainContent())

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m DashboardModel) renderLoading() string {
	return "\n  " + m.spinner.View() + " Loading TokMan Dashboard..."
}

func (m DashboardModel) renderWelcome() string {
	var b strings.Builder

	// Empty lines for vertical centering
	for i := 0; i < m.height/3; i++ {
		b.WriteString("\n")
	}

	// Simple centered welcome
	welcome := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimaryBright)).
		Align(lipgloss.Center).
		Width(m.width)

	b.WriteString(welcome.Render("Welcome to Tokman"))
	b.WriteString("\n\n")

	// Press any key hint
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Align(lipgloss.Center).
		Width(m.width)

	b.WriteString(hint.Render("Press any key to start..."))

	return b.String()
}

func (m DashboardModel) renderHeader() string {
	// Status indicator
	statusColor := SuccessStyle
	statusText := "ONLINE"
	if m.stats.CacheHitRate < 50 {
		statusColor = WarningStyle
		statusText = "WARN"
	}

	// Compact status bar exactly as requested
	status := fmt.Sprintf("%s│◉ %dsessions│▼%stoday│∑%stotal│⚡ %.0f%%cache",
		statusColor.Render("["+statusText+"]"),
		m.stats.ActiveSessions,
		AccentStyle.Render(formatTokens(int(m.stats.TodaySaved))),
		AccentStyle.Render(formatTokens(int(m.stats.TotalSaved))),
		m.stats.CacheHitRate,
	)

	return status
}

func (m DashboardModel) renderMainContent() string {
	// Sidebar + Content layout
	sidebar := m.renderSidebar()
	content := m.renderContent()

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
}

func (m DashboardModel) renderSidebar() string {
	var items []string

	// Navigation menu
	items = append(items, TextMutedStyle.Render("NAVIGATION"))
	items = append(items, "")

	for i := 0; i < int(TabCount); i++ {
		tab := Tab(i)
		label := tab.String()

		if i == int(m.activeTab) {
			items = append(items, tab.Color().Render(" "+label+" "))
		} else {
			items = append(items, "  "+TextDimStyle.Render(label))
		}
	}

	items = append(items, "")
	items = append(items, TextMutedStyle.Render("QUICK STATS"))
	items = append(items, "")

	// Mini stats
	items = append(items, fmt.Sprintf("  %s %s",
		TextMutedStyle.Render("Commands:"),
		StatValueStyle.Render(fmt.Sprintf("%d", m.stats.TotalCommands))))

	items = append(items, fmt.Sprintf("  %s %s",
		TextMutedStyle.Render("Saved:"),
		SuccessStyle.Render(formatTokens(int(m.stats.TotalSaved)))))

	items = append(items, fmt.Sprintf("  %s %s",
		TextMutedStyle.Render("Cache:"),
		InfoStyle.Render(fmt.Sprintf("%.1f%%", m.stats.CacheHitRate))))

	content := strings.Join(items, "\n")

	return BoxDim.Render(content)
}

func (m DashboardModel) renderContent() string {
	switch m.activeTab {
	case OverviewTab:
		return m.renderOverview()
	case MetricsTab:
		return m.renderMetrics()
	case LayersTab:
		return m.renderLayers()
	case AnalyticsTab:
		return m.renderAnalytics()
	case SessionsTab:
		return m.renderSessions()
	case EconomicsTab:
		return m.renderEconomics()
	case ConfigTab:
		return m.renderConfig()
	case LogsTab:
		return m.renderLogs()
	default:
		return m.renderOverview()
	}
}

// ============================================================================
// TAB VIEWS
// ============================================================================

func (m DashboardModel) renderOverview() string {
	var sections []string

	// Header
	sections = append(sections, HeaderPrimary.Render(" SYSTEM OVERVIEW "))
	sections = append(sections, "")

	// Main stats grid
	statsRow1 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderStatBox("Total Commands", fmt.Sprintf("%d", m.stats.TotalCommands), ColorPrimary),
		renderStatBox("Tokens Saved", formatTokens(int(m.stats.TotalSaved)), ColorSuccess),
		renderStatBox("Today's Savings", formatTokens(int(m.stats.TodaySaved)), ColorWarning),
	)

	statsRow2 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderStatBox("Avg Savings", fmt.Sprintf("%.1f%%", m.stats.AvgSavings), ColorInfo),
		renderStatBox("Cache Hit Rate", fmt.Sprintf("%.1f%%", m.stats.CacheHitRate), ColorData1),
		renderStatBox("Commands/Hour", fmt.Sprintf("%.1f", m.stats.CommandsPerHour), ColorData2),
	)

	sections = append(sections, statsRow1)
	sections = append(sections, "")
	sections = append(sections, statsRow2)
	sections = append(sections, "")

	// Top command
	if m.stats.TopCommand != "" {
		sections = append(sections, HeaderSuccess.Render(" MOST USED COMMAND "))
		sections = append(sections, "")
		sections = append(sections, BoxSuccess.Render(
			AccentStyle.Render(m.stats.TopCommand)))
		sections = append(sections, "")
	}

	// Recent activity sparkline
	if len(m.hourlyTrend) > 0 {
		sections = append(sections, HeaderInfo.Render(" HOURLY ACTIVITY "))
		sections = append(sections, "")
		sections = append(sections, renderSparkline(m.hourlyTrend, 50))
	}

	return BoxPrimary.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderMetrics() string {
	var sections []string

	sections = append(sections, HeaderSuccess.Render(" PERFORMANCE METRICS "))
	sections = append(sections, "")

	// Cache metrics
	sections = append(sections, TextSecondaryStyle.Render("CACHE PERFORMANCE"))
	sections = append(sections, "")

	cacheMetrics := lipgloss.JoinHorizontal(lipgloss.Top,
		renderMetricBox("Hits", fmt.Sprintf("%s", formatNumber(m.stats.CacheHits)), SuccessStyle),
		renderMetricBox("Misses", fmt.Sprintf("%s", formatNumber(m.stats.CacheMisses)), WarningStyle),
		renderMetricBox("Hit Rate", fmt.Sprintf("%.1f%%", m.stats.CacheHitRate), InfoStyle),
	)
	sections = append(sections, cacheMetrics)
	sections = append(sections, "")

	// Progress bar for hit rate
	sections = append(sections, TextMutedStyle.Render("Cache Efficiency:"))
	sections = append(sections, renderProgressBar(m.stats.CacheHitRate, 100, 60, ColorSuccess))
	sections = append(sections, "")

	// Command distribution
	if len(m.topCommands) > 0 {
		sections = append(sections, TextSecondaryStyle.Render("COMMAND DISTRIBUTION"))
		sections = append(sections, "")
		sections = append(sections, m.cmdTable.View())
	}

	return BoxSuccess.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderLayers() string {
	var sections []string

	sections = append(sections, HeaderInfo.Render(" 20-LAYER FILTER PIPELINE "))
	sections = append(sections, "")

	if len(m.layers) == 0 {
		sections = append(sections, TextMutedStyle.Render("Loading filter layers..."))
		return BoxInfo.Render(strings.Join(sections, "\n"))
	}

	// Layer stats
	activeCount := 0
	totalSaved := int64(0)
	for _, l := range m.layers {
		if l.Enabled {
			activeCount++
		}
		totalSaved += l.TokensSaved
	}

	statusLine := lipgloss.JoinHorizontal(lipgloss.Left,
		TextMutedStyle.Render("Active Layers: "),
		AccentStyle.Render(fmt.Sprintf("%d/%d", activeCount, len(m.layers))),
		TextMutedStyle.Render("  |  "),
		TextMutedStyle.Render("Total Saved: "),
		SuccessStyle.Render(formatTokens(int(totalSaved))),
	)
	sections = append(sections, statusLine)
	sections = append(sections, "")

	// Layer table
	sections = append(sections, TextMutedStyle.Render("Press [space] to toggle layers"))
	sections = append(sections, "")

	var layerRows []string
	for _, layer := range m.layers {
		status := "○"
		statusColor := TextDimStyle
		if layer.Enabled {
			status = "●"
			statusColor = SuccessStyle
		}

		efficiency := renderMiniBar(layer.Efficiency, 20)

		row := fmt.Sprintf("%s %s %s %s %s %s",
			statusColor.Render(status),
			TextPrimaryStyle.Render(fmt.Sprintf("%-3d", layer.ID)),
			AccentStyle.Render(fmt.Sprintf("%-20s", layer.Name)),
			TextMutedStyle.Render(fmt.Sprintf("%-15s", layer.Research)),
			efficiency,
			SuccessStyle.Render(fmt.Sprintf("%10s", formatTokens(int(layer.TokensSaved)))),
		)
		layerRows = append(layerRows, row)
	}

	sections = append(sections, strings.Join(layerRows, "\n"))

	return BoxInfo.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderAnalytics() string {
	var sections []string

	sections = append(sections, HeaderWarning.Render(" ANALYTICS & INSIGHTS "))
	sections = append(sections, "")

	// Daily trend bar chart
	if len(m.dailyTrend) > 0 {
		sections = append(sections, TextSecondaryStyle.Render("DAILY SAVINGS TREND"))
		sections = append(sections, "")
		sections = append(sections, renderBarChart(m.dailyTrend, 40))
		sections = append(sections, "")
	}

	// Top commands with bars
	if len(m.topCommands) > 0 {
		sections = append(sections, TextSecondaryStyle.Render("TOP COMMANDS BY USAGE"))
		sections = append(sections, "")

		maxCount := 0
		for _, cmd := range m.topCommands {
			if cmd.Count > maxCount {
				maxCount = cmd.Count
			}
		}

		for _, cmd := range m.topCommands {
			barWidth := 0
			if maxCount > 0 {
				barWidth = (cmd.Count * 30) / maxCount
			}
			bar := BarSuccess.Render(strings.Repeat("█", barWidth))

			row := fmt.Sprintf("%-15s %s %5d  %s",
				TextPrimaryStyle.Render(cmd.Command),
				bar,
				cmd.Count,
				SuccessStyle.Render(formatTokens(int(cmd.TokensSaved))),
			)
			sections = append(sections, row)
		}
	}

	return BoxWarning.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderSessions() string {
	var sections []string

	sections = append(sections, HeaderSuccessDim.Render(" SESSION MANAGEMENT "))
	sections = append(sections, "")

	if len(m.sessions) == 0 {
		sections = append(sections, TextMutedStyle.Render("No active sessions"))
		return BoxDim.Render(strings.Join(sections, "\n"))
	}

	// Session count
	sections = append(sections, fmt.Sprintf("%s %s",
		TextMutedStyle.Render("Active Sessions:"),
		AccentStyle.Render(fmt.Sprintf("%d", len(m.sessions)))))
	sections = append(sections, "")

	// Session list
	for _, s := range m.sessions {
		statusColor := SuccessStyle
		if s.Status == "idle" {
			statusColor = WarningStyle
		}

		row := fmt.Sprintf("%s %s  %s  %s  %s  %s",
			statusColor.Render("●"),
			TextPrimaryStyle.Render(s.ID),
			TextMutedStyle.Render(s.StartTime.Format("15:04")),
			TextSecondaryStyle.Render(fmt.Sprintf("%3d cmds", s.CommandCount)),
			SuccessStyle.Render(fmt.Sprintf("%8s", formatTokens(int(s.TokensSaved)))),
			TextDimStyle.Render(fmt.Sprintf("(%s)", s.Agent)),
		)
		sections = append(sections, row)
	}

	return BoxSuccess.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderEconomics() string {
	var sections []string

	sections = append(sections, HeaderWarningDim.Render(" COST ANALYSIS "))
	sections = append(sections, "")

	// Cost savings
	sections = append(sections, TextSecondaryStyle.Render("COST SAVINGS"))
	sections = append(sections, "")

	costBox := lipgloss.JoinHorizontal(lipgloss.Top,
		renderStatBox("Total Saved", fmt.Sprintf("$%.2f", m.stats.TotalCostSaved), ColorSuccess),
		renderStatBox("Per Command", fmt.Sprintf("$%.4f", m.stats.TotalCostSaved/float64(max(1, m.stats.TotalCommands))), ColorInfo),
	)
	sections = append(sections, costBox)
	sections = append(sections, "")

	// Pricing tiers
	sections = append(sections, TextSecondaryStyle.Render("PRICING REFERENCE"))
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("  %s ~$0.03/1K tokens", TextMutedStyle.Render("GPT-4/Claude:")))
	sections = append(sections, fmt.Sprintf("  %s ~$0.002/1K tokens", TextMutedStyle.Render("GPT-3.5:")))
	sections = append(sections, fmt.Sprintf("  %s $0 (compute only)", TextMutedStyle.Render("Local LLM:")))
	sections = append(sections, "")

	// Savings summary
	sections = append(sections, SuccessStyle.Render(
		fmt.Sprintf("✓ Saved %s tokens worth $%.2f",
			formatTokens(int(m.stats.TotalSaved)),
			m.stats.TotalCostSaved)))

	return BoxWarning.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderConfig() string {
	var sections []string

	sections = append(sections, HeaderInfoDim.Render(" CONFIGURATION "))
	sections = append(sections, "")

	// Pipeline settings
	sections = append(sections, TextSecondaryStyle.Render("PIPELINE SETTINGS"))
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("  %s %s", TextMutedStyle.Render("Max Context:"), AccentStyle.Render("2M tokens")))
	sections = append(sections, fmt.Sprintf("  %s %s", TextMutedStyle.Render("Chunk Size:"), AccentStyle.Render("100K tokens")))
	sections = append(sections, fmt.Sprintf("  %s %s", TextMutedStyle.Render("Stream Threshold:"), AccentStyle.Render("500K tokens")))
	sections = append(sections, "")

	// Feature flags
	sections = append(sections, TextSecondaryStyle.Render("FEATURES"))
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("  %s %s", SuccessStyle.Render("[✓]"), TextPrimaryStyle.Render("Semantic caching")))
	sections = append(sections, fmt.Sprintf("  %s %s", SuccessStyle.Render("[✓]"), TextPrimaryStyle.Render("LLM compaction")))
	sections = append(sections, fmt.Sprintf("  %s %s", SuccessStyle.Render("[✓]"), TextPrimaryStyle.Render("SIMD optimizations")))
	sections = append(sections, fmt.Sprintf("  %s %s", SuccessStyle.Render("[✓]"), TextPrimaryStyle.Render("20-layer pipeline")))
	sections = append(sections, "")

	// Paths
	sections = append(sections, TextSecondaryStyle.Render("PATHS"))
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("  %s %s", TextMutedStyle.Render("Config:"), TextDimStyle.Render("~/.config/tokman/config.toml")))
	sections = append(sections, fmt.Sprintf("  %s %s", TextMutedStyle.Render("Database:"), TextDimStyle.Render("~/.local/share/tokman/tokman.db")))

	return BoxInfo.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderLogs() string {
	var sections []string

	sections = append(sections, HeaderErrorDim.Render(" SYSTEM LOGS "))
	sections = append(sections, "")

	if len(m.logs) == 0 {
		sections = append(sections, TextMutedStyle.Render("No recent log entries"))
		return BoxDim.Render(strings.Join(sections, "\n"))
	}

	for _, log := range m.logs {
		levelColor := TextDimStyle
		switch log.Level {
		case "SUCCESS":
			levelColor = SuccessStyle
		case "WARN":
			levelColor = WarningStyle
		case "ERROR":
			levelColor = ErrorStyle
		case "INFO":
			levelColor = InfoStyle
		}

		row := fmt.Sprintf("%s %s %s: %s",
			TextDimStyle.Render(log.Time.Format("15:04:05")),
			levelColor.Render(fmt.Sprintf("[%s]", log.Level)),
			TextMutedStyle.Render(log.Source),
			TextPrimaryStyle.Render(log.Message),
		)
		sections = append(sections, row)
	}

	return BoxDim.Render(strings.Join(sections, "\n"))
}

func (m DashboardModel) renderFooter() string {
	helpView := m.help.View(m.keys)
	timeStr := m.lastUpdate.Format("15:04:05")
	if m.lastUpdate.IsZero() {
		timeStr = "--:--:--"
	}

	status := fmt.Sprintf("%s  |  Updated: %s",
		helpView,
		TextDimStyle.Render(timeStr),
	)

	return FooterStyle.Render(status)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func renderStatBox(label, value, color string) string {
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2).
		Width(20).
		Align(lipgloss.Center)

	content := fmt.Sprintf("%s\n%s",
		TextMutedStyle.Render(label),
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(value),
	)

	return box.Render(content)
}

func renderMetricBox(label, value string, style lipgloss.Style) string {
	return fmt.Sprintf("%-12s %s", TextMutedStyle.Render(label), style.Render(value))
}

func renderProgressBar(value, max float64, width int, color string) string {
	pct := value / max
	if pct > 1 {
		pct = 1
	}
	if pct < 0 {
		pct = 0
	}

	filled := int(pct * float64(width))
	empty := width - filled

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", filled)) +
		TextDimStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("[%s] %.1f%%", bar, value)
}

func renderMiniBar(percentage float64, width int) string {
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	color := ColorSuccess
	if percentage < 50 {
		color = ColorWarning
	}
	if percentage < 25 {
		color = ColorError
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", filled)) +
		TextDimStyle.Render(strings.Repeat("░", empty))
}

func renderSparkline(data []TimeSeriesPoint, width int) string {
	if len(data) == 0 {
		return ""
	}

	// Find max
	maxVal := int64(0)
	for _, p := range data {
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Characters for sparkline (low to high)
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var result strings.Builder
	for _, p := range data {
		idx := int(float64(p.Value) / float64(maxVal) * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		result.WriteString(chars[idx])
	}

	return BarSuccess.Render(result.String())
}

func renderBarChart(data []TimeSeriesPoint, maxWidth int) string {
	if len(data) == 0 {
		return ""
	}

	// Find max
	maxVal := int64(0)
	for _, p := range data {
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	var rows []string
	for _, p := range data {
		label := p.Time.Format("Mon")
		barWidth := int(float64(p.Value) / float64(maxVal) * float64(maxWidth))
		bar := BarPrimary.Render(strings.Repeat("█", barWidth))

		row := fmt.Sprintf("%-4s %s %s",
			TextMutedStyle.Render(label),
			bar,
			TextSecondaryStyle.Render(formatTokens(int(p.Value))),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// ============================================================================
// COMMANDS & MESSAGES
// ============================================================================

type tickMsg time.Time
type dashboardUpdateMsg struct {
	stats         DashboardStats
	dailyTrend    []TimeSeriesPoint
	hourlyTrend   []TimeSeriesPoint
	topCommands   []CommandStat
	layers        []FilterLayerInfo
	sessions      []SessionInfo
	logs          []LogEntryInfo
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchDashboardDataCmd() tea.Cmd {
	return func() tea.Msg {
		// Generate demo data
		stats := DashboardStats{
			TotalCommands:   12345,
			TotalSaved:      5678900,
			TodaySaved:      45000,
			AvgSavings:      82.5,
			CommandsPerHour: 45.2,
			CacheHits:       10000,
			CacheMisses:     500,
			CacheHitRate:    95.2,
			ActiveSessions:  3,
			TopCommand:      "git status",
			TotalCostSaved:  170.37,
		}

		// Daily trend
		dailyTrend := []TimeSeriesPoint{
			{Time: time.Now().AddDate(0, 0, -6), Value: 42000},
			{Time: time.Now().AddDate(0, 0, -5), Value: 58000},
			{Time: time.Now().AddDate(0, 0, -4), Value: 35000},
			{Time: time.Now().AddDate(0, 0, -3), Value: 72000},
			{Time: time.Now().AddDate(0, 0, -2), Value: 45000},
			{Time: time.Now().AddDate(0, 0, -1), Value: 28000},
			{Time: time.Now(), Value: 45000},
		}

		// Hourly trend
		hourlyTrend := []TimeSeriesPoint{}
		for i := 0; i < 24; i++ {
			hourlyTrend = append(hourlyTrend, TimeSeriesPoint{
				Time:  time.Now().Add(time.Duration(-23+i) * time.Hour),
				Value: int64(1000 + i*100),
			})
		}

		// Top commands
		topCommands := []CommandStat{
			{Command: "git status", Count: 234, TokensSaved: 450000, Trend: 0.2},
			{Command: "cargo test", Count: 156, TokensSaved: 890000, Trend: 0.1},
			{Command: "npm ls", Count: 89, TokensSaved: 234000, Trend: -0.1},
			{Command: "docker ps", Count: 67, TokensSaved: 123000, Trend: 0.0},
			{Command: "ls -la", Count: 45, TokensSaved: 34000, Trend: -0.2},
		}

		// Filter layers
		layers := []FilterLayerInfo{
			{ID: 1, Name: "Entropy Filtering", Research: "Mila 2023", Enabled: true, Efficiency: 85, TokensSaved: 450000},
			{ID: 2, Name: "Perplexity Pruning", Research: "MS/Tsinghua", Enabled: true, Efficiency: 92, TokensSaved: 890000},
			{ID: 3, Name: "Goal-Driven Selection", Research: "SJTU 2025", Enabled: true, Efficiency: 78, TokensSaved: 234000},
			{ID: 4, Name: "AST Preservation", Research: "NUS 2025", Enabled: true, Efficiency: 88, TokensSaved: 123000},
			{ID: 5, Name: "Contrastive Ranking", Research: "MS 2024", Enabled: true, Efficiency: 90, TokensSaved: 34000},
		}

		// Sessions
		sessions := []SessionInfo{
			{ID: "sess_abc123", StartTime: time.Now().Add(-2 * time.Hour), Duration: 2 * time.Hour, CommandCount: 45, TokensSaved: 125000, Agent: "Claude", Status: "active"},
			{ID: "sess_def456", StartTime: time.Now().Add(-5 * time.Hour), Duration: 5 * time.Hour, CommandCount: 128, TokensSaved: 340000, Agent: "Cursor", Status: "active"},
			{ID: "sess_ghi789", StartTime: time.Now().Add(-1 * time.Hour), Duration: 1 * time.Hour, CommandCount: 12, TokensSaved: 28000, Agent: "Copilot", Status: "idle"},
		}

		// Logs
		logs := []LogEntryInfo{
			{Time: time.Now().Add(-2 * time.Minute), Level: "INFO", Source: "filter", Message: "Compressed git status: 2.1K → 420 tokens"},
			{Time: time.Now().Add(-3 * time.Minute), Level: "SUCCESS", Source: "cache", Message: "Cache hit for fingerprint: abc123"},
			{Time: time.Now().Add(-5 * time.Minute), Level: "INFO", Source: "layer13", Message: "H2O filter saved 450 tokens"},
			{Time: time.Now().Add(-8 * time.Minute), Level: "WARN", Source: "memory", Message: "High memory usage: 85%"},
			{Time: time.Now().Add(-10 * time.Minute), Level: "INFO", Source: "session", Message: "Session sess_abc123 started"},
		}

		return dashboardUpdateMsg{
			stats:       stats,
			dailyTrend:  dailyTrend,
			hourlyTrend: hourlyTrend,
			topCommands: topCommands,
			layers:      layers,
			sessions:    sessions,
			logs:        logs,
		}
	}
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

func NewDashboard() DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)

	h := help.New()

	return DashboardModel{
		spinner:     s,
		progress:    p,
		help:        h,
		keys:        newKeyMap(),
		loading:     true,
		showWelcome: true,
	}
}

func RunDashboard() error {
	if termenv.NewOutput(os.Stdout).ColorProfile() == termenv.Ascii {
		fmt.Println("TokMan Dashboard (Basic Mode)")
		fmt.Println("=============================")
		fmt.Println()
		fmt.Println("Total Commands: 12,345")
		fmt.Println("Tokens Saved: 5.7M")
		fmt.Println("Cache Hit Rate: 95.2%")
		fmt.Println()
		fmt.Println("Press any key to exit...")
		fmt.Scanln()
		return nil
	}

	m := NewDashboard()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

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

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
