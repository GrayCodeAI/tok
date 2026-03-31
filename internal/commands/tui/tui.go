package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/tee"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI dashboard",
	Long:  `Launch an interactive terminal dashboard for TokMan with real-time analytics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

func init() {
	registry.Add(func() { registry.Register(tuiCmd) })
}

const refreshInterval = 2 * time.Second

type tab int

const (
	tabOverview tab = iota
	tabCommands
	tabLayers
	tabTimeline
	tabDiscover
	tabTee
	tabLive
)

type model struct {
	width     int
	height    int
	activeTab tab
	tracker   *tracking.Tracker
	ready     bool

	summary    *Summary
	cmdTable   table.Model
	layerTable table.Model
	viewport   viewport.Model

	quitting bool
	showHelp bool
	lastTick time.Time
}

type Summary struct {
	TotalCommands int
	TotalInput    int
	TotalOutput   int
	TotalSaved    int
	AvgSavings    float64
	Period        string
	LastUpdated   time.Time
}

type DayData struct {
	Date  string
	Saved int
	Count int
}

type tickMsg time.Time
type dataUpdatedMsg struct {
	summary   *Summary
	cmdRows   []table.Row
	layerRows []table.Row
	dailyData []DayData
}

func runTUI() error {
	dbPath := tracking.DatabasePath()
	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		tracker = nil
	}
	p := tea.NewProgram(initialModel(tracker), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModel(tracker *tracking.Tracker) model {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Command", Width: 26},
			{Title: "Count", Width: 7},
			{Title: "Saved", Width: 12},
			{Title: "Avg%", Width: 8},
			{Title: "Last Seen", Width: 14},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	l := table.New(
		table.WithColumns([]table.Column{
			{Title: "Layer", Width: 22},
			{Title: "Paper", Width: 24},
			{Title: "Status", Width: 12},
		}),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color("235")).
		Background(lipgloss.Color("252")).
		Bold(false)
	t.SetStyles(ts)
	l.SetStyles(ts)

	v := viewport.New(80, 20)

	return model{
		tracker:    tracker,
		cmdTable:   t,
		layerTable: l,
		viewport:   v,
		lastTick:   time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) }),
		fetchData(m.tracker),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
			m.cmdTable.SetWidth(msg.Width - 6)
			m.layerTable.SetWidth(msg.Width - 6)
			m.viewport.Width = msg.Width - 6
			m.viewport.Height = msg.Height - 16
		}
		return m, nil

	case tickMsg:
		m.lastTick = time.Now()
		cmds = append(cmds, tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) }))
		cmds = append(cmds, fetchData(m.tracker))
		return m, tea.Batch(cmds...)

	case dataUpdatedMsg:
		m.summary = msg.summary
		m.cmdTable.SetRows(msg.cmdRows)
		m.layerTable.SetRows(msg.layerRows)
		m.viewport.SetContent(m.renderTimeline(msg.dailyData))
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % 7
		case "left", "h":
			m.activeTab = (m.activeTab - 1 + 7) % 7
		case "1":
			m.activeTab = tabOverview
		case "2":
			m.activeTab = tabCommands
		case "3":
			m.activeTab = tabLayers
		case "4":
			m.activeTab = tabTimeline
		case "5":
			m.activeTab = tabDiscover
		case "6":
			m.activeTab = tabTee
		case "7":
			m.activeTab = tabLive
		case "r":
			cmds = append(cmds, fetchData(m.tracker))
		}
	}

	var cmd tea.Cmd
	switch m.activeTab {
	case tabCommands:
		m.cmdTable, cmd = m.cmdTable.Update(msg)
	case tabLayers:
		m.layerTable, cmd = m.layerTable.Update(msg)
	case tabTimeline:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return "\n\n  TokMan TUI closed.\n"
	}
	if m.showHelp {
		return m.helpView()
	}
	if !m.ready {
		return "\n\n  Initializing..."
	}

	var content string
	switch m.activeTab {
	case tabOverview:
		content = m.overviewView()
	case tabCommands:
		content = m.commandsView()
	case tabLayers:
		content = m.layersView()
	case tabTimeline:
		content = m.viewport.View()
	case tabDiscover:
		content = m.discoverView()
	case tabTee:
		content = m.teeView()
	case tabLive:
		content = m.liveView()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		"",
		m.statusBar(),
		m.tabBar(),
		"",
		content,
		"",
		m.helpBar(),
	)
}

func (m model) statusBar() string {
	var center string
	if m.summary != nil {
		center = fmt.Sprintf("  %d commands  ·  %s saved  ·  %.1f%% avg",
			m.summary.TotalCommands,
			formatTokens(m.summary.TotalSaved),
			m.summary.AvgSavings)
	} else {
		center = "  Loading..."
	}

	right := fmt.Sprintf("%s  ", time.Now().Format("15:04:05"))

	title := lipgloss.NewStyle().Bold(true).Render("tokman")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		center,
		lipgloss.NewStyle().Width(m.width-len(title)-len(center)-len(right)).Render(""),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(right),
	)
}

func (m model) tabBar() string {
	tabs := []string{"overview", "commands", "layers", "timeline", "discover", "tee", "live"}
	var rendered []string
	for i, t := range tabs {
		if tab(i) == m.activeTab {
			rendered = append(rendered, lipgloss.NewStyle().
				Foreground(lipgloss.Color("235")).
				Background(lipgloss.Color("252")).
				Bold(true).
				Padding(0, 2).
				Render(t))
		} else {
			rendered = append(rendered, lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2).
				Render(t))
		}
	}
	return strings.Join(rendered, "")
}

func (m model) helpBar() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("  1-7: tabs  ↑↓: scroll  r: refresh  ?: help  q: quit")
}

func (m model) helpView() string {
	title := lipgloss.NewStyle().Bold(true).Render("  Keyboard Shortcuts")

	help := strings.Join([]string{
		"",
		"  Navigation",
		"    1-7 / Tab / ←→    Switch tabs",
		"    ↑↓ / j/k          Scroll list",
		"    g / G             Go to top / bottom",
		"",
		"  Actions",
		"    r                 Refresh data",
		"    ?                 Toggle this help",
		"    q / Ctrl+C        Quit",
		"",
		"  Tabs",
		"    1  Overview       Dashboard with KPIs",
		"    2  Commands       Command history",
		"    3  Layers         Compression layers",
		"    4  Timeline       Daily savings trend",
		"    5  Discover       Missed savings opportunities",
		"    6  Tee            Full output recovery",
		"    7  Live           Real-time command monitor",
		"",
	}, "\n")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		"",
		title,
		"  "+lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", 40)),
		help,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  Press ? to close"),
	)
}

func (m model) overviewView() string {
	if m.summary == nil {
		return "  Loading data..."
	}

	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	kpis := []string{
		fmt.Sprintf("  commands   %s", bold.Render(fmt.Sprintf("%d", m.summary.TotalCommands))),
		fmt.Sprintf("  input      %s", bold.Render(formatTokens(m.summary.TotalInput))),
		fmt.Sprintf("  output     %s", bold.Render(formatTokens(m.summary.TotalOutput))),
		fmt.Sprintf("  saved      %s", green.Render(fmt.Sprintf("%s  (%.1f%%)", formatTokens(m.summary.TotalSaved), m.summary.AvgSavings))),
		fmt.Sprintf("  period     %s", dim.Render(m.summary.Period)),
		fmt.Sprintf("  updated    %s", dim.Render(m.summary.LastUpdated.Format("15:04:05"))),
	}

	meter := buildEfficiencyMeter(m.summary.AvgSavings)

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(kpis, "\n"),
		"",
		fmt.Sprintf("  efficiency  %s", meter),
	)
}

func (m model) commandsView() string {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	return lipgloss.JoinVertical(lipgloss.Left,
		bold.Render("  command history"),
		dim.Render("  "+strings.Repeat("─", m.width-8)),
		m.cmdTable.View(),
	)
}

func (m model) layersView() string {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	return lipgloss.JoinVertical(lipgloss.Left,
		bold.Render("  compression layers"),
		dim.Render("  "+strings.Repeat("─", m.width-8)),
		m.layerTable.View(),
	)
}

func (m model) renderTimeline(data []DayData) string {
	if len(data) == 0 {
		return "  No timeline data available."
	}

	var lines []string
	maxSaved := 1
	for _, d := range data {
		if d.Saved > maxSaved {
			maxSaved = d.Saved
		}
	}

	for _, d := range data {
		barLen := int(math.Round(float64(d.Saved) / float64(maxSaved) * 36))
		bar := strings.Repeat("▸", barLen)
		var color string
		if d.Saved > 100000 {
			color = "42"
		} else if d.Saved > 10000 {
			color = "220"
		} else {
			color = "240"
		}
		coloredBar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(bar)
		dateShort := d.Date
		if len(dateShort) > 10 {
			dateShort = dateShort[5:10]
		}
		lines = append(lines, fmt.Sprintf("  %s  %6s  %s  %s",
			dateShort,
			formatTokens(d.Saved),
			coloredBar,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%d cmds", d.Count))))
	}

	return strings.Join(lines, "\n")
}

func (m model) discoverView() string {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	analyzer := core.NewDiscoverAnalyzer()
	results := analyzer.AnalyzeBatch([]string{
		"cat file.txt", "ls -la", "grep pattern .", "docker ps",
		"kubectl get pods", "curl http://api", "env", "npm test",
	})

	lines := []string{
		bold.Render("  missed savings opportunities"),
		dim.Render("  " + strings.Repeat("─", m.width-8)),
		"",
	}
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("  %s  %s  %s",
			green.Render(fmt.Sprintf("+%d", r.EstSavings)),
			bold.Render(r.Command),
			dim.Render(r.Suggestion)))
	}
	return strings.Join(lines, "\n")
}

func (m model) teeView() string {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	entries, _ := tee.List(tee.DefaultConfig())
	lines := []string{
		bold.Render("  full output recovery (tee)"),
		dim.Render("  " + strings.Repeat("─", m.width-8)),
		"",
	}
	if len(entries) == 0 {
		lines = append(lines, "  No saved outputs. Tee saves output on command failure.")
	} else {
		for _, e := range entries {
			lines = append(lines, fmt.Sprintf("  %s  %s",
				dim.Render(e.Timestamp.Format("01-02 15:04")),
				bold.Render(e.Command)))
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) liveView() string {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	lines := []string{
		bold.Render("  live command monitor"),
		dim.Render("  " + strings.Repeat("─", m.width-8)),
		"",
		fmt.Sprintf("  %s  %s  %s  %s  %s",
			dim.Render("PID"),
			dim.Render("Command"),
			dim.Render("Duration"),
			dim.Render("Input"),
			dim.Render("Status")),
		dim.Render("  " + strings.Repeat("─", m.width-8)),
	}

	if m.summary != nil {
		lines = append(lines, fmt.Sprintf("  %s  %s  %s  %s  %s",
			dim.Render("---"),
			bold.Render("tokman tui"),
			dim.Render(time.Since(m.lastTick).Round(time.Second).String()),
			green.Render(formatTokens(m.summary.TotalInput)),
			green.Render("active")))
	}

	lines = append(lines, "", dim.Render("  Real-time monitoring requires hook integration."))
	return strings.Join(lines, "\n")
}

func fetchData(tracker *tracking.Tracker) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if tracker == nil {
			return dataUpdatedMsg{}
		}

		summary := &Summary{LastUpdated: time.Now()}
		savings, err := tracker.GetSavings("")
		if err == nil {
			summary.TotalCommands = savings.TotalCommands
			summary.TotalSaved = savings.TotalSaved
			summary.TotalInput = savings.TotalOriginal
			summary.TotalOutput = savings.TotalFiltered
			if savings.TotalOriginal > 0 {
				summary.AvgSavings = float64(savings.TotalSaved) / float64(savings.TotalOriginal) * 100
			}
		}

		daily, _ := tracker.GetDailySavings("", 30)
		if len(daily) > 0 {
			summary.Period = daily[len(daily)-1].Date + " → " + daily[0].Date
		}

		var dayData []DayData
		for _, d := range daily {
			dayData = append(dayData, DayData{Date: d.Date, Saved: d.Saved, Count: d.Commands})
		}

		stats, _ := tracker.GetCommandStats("")
		recent, _ := tracker.GetRecentCommands("", 100)
		tsMap := make(map[string]string)
		for _, r := range recent {
			if _, exists := tsMap[r.Command]; !exists {
				tsMap[r.Command] = r.Timestamp.Format("01-02 15:04")
			}
		}

		var cmdRows []table.Row
		for _, cs := range stats {
			cmdRows = append(cmdRows, table.Row{
				cs.Command,
				fmt.Sprintf("%d", cs.ExecutionCount),
				formatTokens(cs.TotalSaved),
				fmt.Sprintf("%.1f%%", cs.ReductionPct),
				tsMap[cs.Command],
			})
		}

		layerRows := getLayerRows()

		return dataUpdatedMsg{
			summary:   summary,
			cmdRows:   cmdRows,
			layerRows: layerRows,
			dailyData: dayData,
		}
	})
}

func buildEfficiencyMeter(pct float64) string {
	width := 40
	filled := int(pct / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	pctStr := fmt.Sprintf("%.1f%%", pct)

	if pct >= 70 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(bar) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render(pctStr)
	} else if pct >= 40 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(bar) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render(pctStr)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(bar) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(pctStr)
}

func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func getLayerRows() []table.Row {
	return []table.Row{
		{"1_entropy", "Selective Context (Mila)", "enabled"},
		{"2_perplexity", "LLMLingua (Microsoft)", "enabled"},
		{"3_goal_driven", "SWE-Pruner (SJTU)", "query-only"},
		{"4_ast_preserve", "LongCodeZip (NUS)", "enabled"},
		{"5_contrastive", "LongLLMLingua (MS)", "query-only"},
		{"6_ngram", "CompactPrompt", "enabled"},
		{"7_evaluator", "EHPC (Tsinghua)", "enabled"},
		{"8_gist", "Gisting (Stanford)", "enabled"},
		{"9_hierarchical", "AutoCompressor (Princeton)", "enabled"},
		{"11_compaction", "MemGPT (Berkeley)", "optional"},
		{"13_h2o", "Heavy-Hitter Oracle", "enabled"},
		{"14_attention_sink", "StreamingLLM (MIT)", "enabled"},
		{"15_meta_token", "Meta-Tokens (arXiv)", "enabled"},
		{"23_swezze", "SWEzze (PKU/UCL 2026)", "optional"},
		{"24_mixed_dim", "MixedDimKV (2026)", "optional"},
		{"25_beaver", "BEAVER (2026)", "optional"},
		{"26_poc", "PoC (2026)", "optional"},
		{"27_token_quant", "TurboQuant (Google)", "optional"},
		{"28_token_retention", "Token Retention (Yale)", "optional"},
		{"29_acon", "ACON (ICLR 2026)", "optional"},
	}
}
