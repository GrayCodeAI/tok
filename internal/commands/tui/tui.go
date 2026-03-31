package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI",
	Long:  `Launch an interactive terminal dashboard for TokMan.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

func init() {
	registry.Add(func() { registry.Register(tuiCmd) })
}

type tab int

const (
	tabDashboard tab = iota
	tabCommands
	tabLayers
	tabConfig
)

type model struct {
	activeTab  tab
	tabs       []string
	summary    *Summary
	cmdTable   table.Model
	layerTable table.Model
	width      int
	height     int
	quitting   bool
}

type Summary struct {
	TotalCommands int
	TotalInput    int
	TotalOutput   int
	TotalSaved    int
	AvgSavings    float64
	Period        string
}

func runTUI() error {
	dbPath := tracking.DatabasePath()
	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return fmt.Errorf("tracking error: %w", err)
	}
	defer tracker.Close()

	summary := getSummary(tracker)
	cmdRows := getCommandRows(tracker)
	layerRows := getLayerRows()

	p := tea.NewProgram(initialModel(summary, cmdRows, layerRows), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModel(summary *Summary, cmdRows []table.Row, layerRows []table.Row) model {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Command", Width: 28},
			{Title: "Count", Width: 8},
			{Title: "Saved", Width: 12},
			{Title: "Avg%", Width: 8},
			{Title: "Last Seen", Width: 14},
		}),
		table.WithRows(cmdRows),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	l := table.New(
		table.WithColumns([]table.Column{
			{Title: "Layer", Width: 24},
			{Title: "Paper", Width: 22},
			{Title: "Status", Width: 12},
		}),
		table.WithRows(layerRows),
		table.WithFocused(false),
		table.WithHeight(12),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	l.SetStyles(s)

	return model{
		tabs:       []string{"Dashboard", "Commands", "Layers", "Config"},
		summary:    summary,
		cmdTable:   t,
		layerTable: l,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.cmdTable.SetWidth(msg.Width - 4)
		m.layerTable.SetWidth(msg.Width - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % tab(len(m.tabs))
		case "left":
			m.activeTab = (m.activeTab - 1 + tab(len(m.tabs))) % tab(len(m.tabs))
		case "1":
			m.activeTab = tabDashboard
		case "2":
			m.activeTab = tabCommands
		case "3":
			m.activeTab = tabLayers
		case "4":
			m.activeTab = tabConfig
		}

		if m.activeTab == tabCommands {
			m.cmdTable, cmd = m.cmdTable.Update(msg)
		} else if m.activeTab == tabLayers {
			m.layerTable, cmd = m.layerTable.Update(msg)
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "TokMan TUI closed.\n"
	}

	var content string
	switch m.activeTab {
	case tabDashboard:
		content = m.dashboardView()
	case tabCommands:
		content = m.commandsView()
	case tabLayers:
		content = m.layersView()
	case tabConfig:
		content = m.configView()
	}

	return m.tabBar() + "\n\n" + content
}

func (m model) tabBar() string {
	var tabs []string
	for i, t := range m.tabs {
		if tab(i) == m.activeTab {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Padding(0, 2).
				Render(" "+t+" "))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2).
				Render(" "+t+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m model) dashboardView() string {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	title := green.Render("TokMan Dashboard")
	divider := dim.Render(strings.Repeat("─", 60))

	kpis := []string{
		fmt.Sprintf("  Commands:  %s", cyan.Render(fmt.Sprintf("%d", m.summary.TotalCommands))),
		fmt.Sprintf("  Input:     %s", cyan.Render(formatTokens(m.summary.TotalInput))),
		fmt.Sprintf("  Output:    %s", cyan.Render(formatTokens(m.summary.TotalOutput))),
		fmt.Sprintf("  Saved:     %s", green.Render(fmt.Sprintf("%s (%.1f%%)", formatTokens(m.summary.TotalSaved), m.summary.AvgSavings))),
		fmt.Sprintf("  Period:    %s", dim.Render(m.summary.Period)),
		fmt.Sprintf("  Generated: %s", dim.Render(time.Now().Format("2006-01-02 15:04:05"))),
	}

	meter := buildEfficiencyMeter(m.summary.AvgSavings)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		strings.Join(kpis, "\n"),
		"",
		"  Efficiency: "+meter,
		"",
		dim.Render("  [1] Dashboard  [2] Commands  [3] Layers  [4] Config  [q] Quit"),
	)
}

func (m model) commandsView() string {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	title := cyan.Render("Command History")
	divider := dim.Render(strings.Repeat("─", 60))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		m.cmdTable.View(),
		"",
		dim.Render("  ↑↓ navigate  [1] Dashboard  [2] Commands  [3] Layers  [4] Config  [q] Quit"),
	)
}

func (m model) layersView() string {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	title := cyan.Render("Compression Layers (37 total)")
	divider := dim.Render(strings.Repeat("─", 60))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		m.layerTable.View(),
		"",
		dim.Render("  ↑↓ navigate  [1] Dashboard  [2] Commands  [3] Layers  [4] Config  [q] Quit"),
	)
}

func (m model) configView() string {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	title := cyan.Render("Configuration")
	divider := dim.Render(strings.Repeat("─", 60))

	configs := []string{
		"  Tier:           trim (12 layers)",
		"  Mode:           minimal",
		"  Budget:         unlimited",
		"  LLM:            disabled",
		"  Cache:          enabled",
		"  Session:        enabled",
		"  Database:       ~/.local/share/tokman/tracking.db",
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		strings.Join(configs, "\n"),
		"",
		dim.Render("  [1] Dashboard  [2] Commands  [3] Layers  [4] Config  [q] Quit"),
	)
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

func getSummary(tracker *tracking.Tracker) *Summary {
	s := &Summary{}
	savings, err := tracker.GetSavings("")
	if err == nil {
		s.TotalCommands = savings.TotalCommands
		s.TotalSaved = savings.TotalSaved
		s.TotalInput = savings.TotalOriginal
		s.TotalOutput = savings.TotalFiltered
		if savings.TotalOriginal > 0 {
			s.AvgSavings = float64(savings.TotalSaved) / float64(savings.TotalOriginal) * 100
		}
	}

	daily, err := tracker.GetDailySavings("", 30)
	if err == nil && len(daily) > 0 {
		s.Period = daily[len(daily)-1].Date + " → " + daily[0].Date
	}

	return s
}

func getCommandRows(tracker *tracking.Tracker) []table.Row {
	stats, err := tracker.GetCommandStats("")
	if err != nil {
		return nil
	}

	recent, _ := tracker.GetRecentCommands("", 100)
	tsMap := make(map[string]string)
	for _, r := range recent {
		if _, exists := tsMap[r.Command]; !exists {
			tsMap[r.Command] = r.Timestamp.Format("01-02 15:04")
		}
	}

	var rows []table.Row
	for _, cs := range stats {
		rows = append(rows, table.Row{
			cs.Command,
			fmt.Sprintf("%d", cs.ExecutionCount),
			formatTokens(cs.TotalSaved),
			fmt.Sprintf("%.1f%%", cs.ReductionPct),
			tsMap[cs.Command],
		})
	}
	return rows
}

func getLayerRows() []table.Row {
	return []table.Row{
		{"1_entropy", "Selective Context", "enabled"},
		{"2_perplexity", "LLMLingua", "enabled"},
		{"3_goal_driven", "SWE-Pruner", "query-only"},
		{"4_ast_preserve", "LongCodeZip", "enabled"},
		{"5_contrastive", "LongLLMLingua", "query-only"},
		{"6_ngram", "CompactPrompt", "enabled"},
		{"7_evaluator", "EHPC", "enabled"},
		{"8_gist", "Gisting", "enabled"},
		{"9_hierarchical", "AutoCompressor", "enabled"},
		{"11_compaction", "MemGPT", "optional"},
		{"13_h2o", "Heavy-Hitter Oracle", "enabled"},
		{"14_attention_sink", "StreamingLLM", "enabled"},
		{"15_meta_token", "Meta-Tokens", "enabled"},
		{"23_swezze", "SWEzze (2026)", "optional"},
		{"24_mixed_dim", "MixedDimKV (2026)", "optional"},
		{"25_beaver", "BEAVER (2026)", "optional"},
		{"26_poc", "PoC (2026)", "optional"},
		{"27_token_quant", "TurboQuant (2026)", "optional"},
		{"28_token_retention", "Token Retention (2026)", "optional"},
		{"29_acon", "ACON (ICLR 2026)", "optional"},
	}
}
