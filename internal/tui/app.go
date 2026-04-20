package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

type section struct {
	Title       string
	Short       string
	Implemented bool
}

type snapshotLoadedMsg struct {
	snapshot *tracking.WorkspaceDashboardSnapshot
	err      error
	loadedAt time.Time
}

type refreshTickMsg time.Time

// quitMsg is dispatched after the loader has been closed so the program can
// exit cleanly without leaking DB handles.
type quitMsg struct{}

type model struct {
	opts       Options
	loader     snapshotLoader
	ctx        context.Context
	cancel     context.CancelFunc
	keys       KeyMap
	theme      theme
	sections   []section
	navIndex   int
	width      int
	height     int
	ready      bool
	compact    bool
	helpOpen   bool
	loading    bool
	refreshing bool
	quitting   bool
	err        error
	lastLoad   time.Time
	data       *tracking.WorkspaceDashboardSnapshot
	spinner    spinner.Model
}

func NewModel(opts Options) tea.Model {
	return NewModelWithLoader(opts, newWorkspaceLoader())
}

// NewModelWithLoader is the testable constructor: tests inject a stubLoader
// while production paths go through NewModel which owns the real DB-backed
// workspaceLoader.
func NewModelWithLoader(opts Options, loader snapshotLoader) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7"))

	ctx, cancel := context.WithCancel(context.Background())

	return model{
		opts:   opts.normalized(),
		loader: loader,
		ctx:    ctx,
		cancel: cancel,
		keys:   DefaultKeyMap(),
		theme:  newTheme(),
		sections: []section{
			{Title: "Home", Short: "Overview", Implemented: true},
			{Title: "Today", Short: "Easy Day"},
			{Title: "Trends", Short: "Analytics"},
			{Title: "Providers", Short: "Economics"},
			{Title: "Models", Short: "Model Cost"},
			{Title: "Agents", Short: "Agent Ops"},
			{Title: "Sessions", Short: "Session Ops"},
			{Title: "Commands", Short: "Command Mix"},
			{Title: "Pipeline", Short: "Layer View"},
			{Title: "Rewards", Short: "Streaks"},
			{Title: "Logs", Short: "Runtime"},
			{Title: "Config", Short: "Health"},
		},
		loading: true,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadSnapshotCmd(m.ctx, m.loader, m.opts),
		refreshTickCmd(m.opts.RefreshInterval),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.compact = msg.Width < 112 || msg.Height < 26
	case spinner.TickMsg:
		if m.loading || m.refreshing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case refreshTickMsg:
		if m.quitting {
			return m, nil
		}
		if !m.loading {
			m.refreshing = true
			cmds = append(cmds, loadSnapshotCmd(m.ctx, m.loader, m.opts))
		}
		cmds = append(cmds, refreshTickCmd(m.opts.RefreshInterval))
	case snapshotLoadedMsg:
		m.loading = false
		m.refreshing = false
		// Suppress cancellation errors triggered by our own Quit path so the
		// user never sees a red banner flash on exit.
		if msg.err != nil && isCancellationErr(msg.err) {
			break
		}
		m.err = msg.err
		if msg.err == nil {
			m.data = msg.snapshot
			m.lastLoad = msg.loadedAt
		}
	case quitMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		if m.quitting {
			return m, nil
		}
		var cmd tea.Cmd
		m, cmd = m.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading tok TUI…"
	}

	contentWidth := max(20, m.width)
	header := m.renderHeader(contentWidth)
	footer := m.renderFooter(contentWidth)

	bodyHeight := max(8, m.height-lipgloss.Height(header)-lipgloss.Height(footer))
	body := m.renderBody(contentWidth, bodyHeight)

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	return setWidth(m.theme.App, m.width).Render(view)
}

// handleKey dispatches a key press against the keymap registry, mutates the
// model as needed, and returns any resulting tea.Cmd. Kept off Update so the
// message switch stays readable and so section-local handlers can layer on
// in Phase 1 by calling back into this from their own Update methods.
func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, shutdownCmd(m.cancel, m.loader)
	case key.Matches(msg, m.keys.Help):
		m.helpOpen = !m.helpOpen
	case key.Matches(msg, m.keys.Esc):
		m.helpOpen = false
	case key.Matches(msg, m.keys.NextSection):
		m.navIndex = (m.navIndex + 1) % len(m.sections)
	case key.Matches(msg, m.keys.PrevSection):
		m.navIndex = (m.navIndex - 1 + len(m.sections)) % len(m.sections)
	case key.Matches(msg, m.keys.Refresh):
		m.loading = m.data == nil
		m.refreshing = m.data != nil
		return m, loadSnapshotCmd(m.ctx, m.loader, m.opts)
	case key.Matches(msg, m.keys.JumpSection):
		if idx, ok := sectionShortcutIndex(msg.String(), len(m.sections)); ok {
			m.navIndex = idx
		}
	}
	return m, nil
}

func loadSnapshotCmd(ctx context.Context, loader snapshotLoader, opts Options) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := loader.Load(ctx, opts)
		return snapshotLoadedMsg{
			snapshot: snapshot,
			err:      err,
			loadedAt: time.Now(),
		}
	}
}

// shutdownCmd cancels in-flight loads and closes loader-held DB handles,
// then dispatches quitMsg so tea.Quit runs only after teardown is complete.
func shutdownCmd(cancel context.CancelFunc, loader snapshotLoader) tea.Cmd {
	return func() tea.Msg {
		if cancel != nil {
			cancel()
		}
		if loader != nil {
			_ = loader.Close()
		}
		return quitMsg{}
	}
}

// isCancellationErr reports whether err is (or wraps) a context cancellation.
func isCancellationErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func refreshTickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func sectionShortcutIndex(key string, sectionCount int) (int, bool) {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(key); err == nil {
		if n <= 0 || n > sectionCount {
			return 0, false
		}
		return n - 1, true
	}
	return 0, false
}

func sectionShortcutLabel(index int) string {
	return strconv.Itoa(index + 1)
}

func (m model) renderHeader(width int) string {
	title := m.theme.Title.Render("tok")
	sectionTitle := m.theme.Focus.Render(m.sections[m.navIndex].Title)

	statusParts := []string{
		sectionTitle,
		m.theme.HeaderMuted.Render(fmt.Sprintf("%dd window", m.opts.Days)),
	}
	if m.opts.ProjectPath != "" {
		statusParts = append(statusParts, m.theme.HeaderMuted.Render("project filtered"))
	}
	if m.refreshing {
		statusParts = append(statusParts, m.theme.Focus.Render(m.spinner.View()+" refreshing"))
	} else if m.loading {
		statusParts = append(statusParts, m.theme.Focus.Render(m.spinner.View()+" loading"))
	} else if !m.lastLoad.IsZero() {
		statusParts = append(statusParts, m.theme.HeaderMuted.Render("updated "+m.lastLoad.Format("15:04:05")))
	}

	left := title + "  " + strings.Join(statusParts, m.theme.HeaderMuted.Render(" · "))
	return setWidth(m.theme.Header, width).Render(left)
}

func (m model) renderBody(width, height int) string {
	if m.compact {
		tabs := m.renderCompactTabs(width)
		main := m.renderMain(width, max(4, height-lipgloss.Height(tabs)))
		return lipgloss.JoinVertical(lipgloss.Left, tabs, main)
	}

	sidebarWidth := 18
	showRightPane := width >= 170 && m.navIndex != 0
	rightWidth := 0
	if showRightPane {
		rightWidth = 26
	}
	gap := 1
	mainWidth := max(24, width-sidebarWidth-rightWidth-gap)

	sidebar := m.renderSidebar(sidebarWidth, height)
	main := m.renderMain(mainWidth, height)
	if !showRightPane {
		return joinHorizontalGap(" ", sidebar, main)
	}
	right := m.renderInsights(rightWidth, height)

	return joinHorizontalGap(" ", sidebar, main, right)
}

func (m model) renderSidebar(width, height int) string {
	lines := make([]string, 0, len(m.sections)+4)
	lines = append(lines, m.theme.SectionTitle.Render("Sections"))
	for i, s := range m.sections {
		label := sectionShortcutLabel(i)
		text := lipgloss.JoinHorizontal(lipgloss.Left, m.theme.SidebarKey.Render(label), " ", s.Title)
		if i == m.navIndex {
			lines = append(lines, setWidth(m.theme.SidebarActive, width-2).Render(text))
		} else {
			lines = append(lines, setWidth(m.theme.SidebarItem, width-2).Render(text))
		}
	}
	lines = append(lines, "")
	lines = append(lines, m.theme.Muted.Render("Dashboard"))

	return setWidth(m.theme.Sidebar, width).Render(strings.Join(lines, "\n"))
}

func (m model) renderCompactTabs(width int) string {
	parts := make([]string, 0, len(m.sections))
	for i, s := range m.sections {
		label := s.Title
		if i == m.navIndex {
			parts = append(parts, m.theme.SidebarActive.Render(label))
		} else {
			parts = append(parts, m.theme.SidebarItem.Render(label))
		}
	}
	return setWidth(lipgloss.NewStyle().Padding(0, 1), width).Render(strings.Join(parts, "  "))
}

func (m model) renderMain(width, height int) string {
	if m.helpOpen {
		return setWidth(m.theme.Main, width).Render(m.renderHelp(width))
	}
	if m.loading && m.data == nil {
		return setWidth(m.theme.Main, width).Render("\n" + m.spinner.View() + " Loading workspace snapshot…")
	}
	if m.err != nil && m.data == nil {
		return setWidth(m.theme.Main, width).Render(m.theme.Danger.Render("Failed to load snapshot") + "\n\n" + m.err.Error())
	}
	if m.navIndex == 0 {
		return setWidth(m.theme.Main, width).Render(m.renderHome(width))
	}
	return setWidth(m.theme.Main, width).Render(m.renderPlaceholder(width))
}

func (m model) renderInsights(width, height int) string {
	if m.compact {
		return ""
	}
	lines := []string{m.theme.SectionTitle.Render("Insights")}

	if m.err != nil {
		lines = append(lines, m.theme.Danger.Render("Data load issue"))
		lines = append(lines, m.theme.Insight.Render(m.err.Error()))
	}

	if m.data != nil {
		quality := m.data.DataQuality
		switch {
		case quality.PricingCoverage.FallbackPricingCommands > 0:
			lines = append(lines, m.theme.Warning.Render("Pricing coverage"))
			lines = append(lines, m.theme.Insight.Render(
				fmt.Sprintf("%.1f%% explicit pricing", quality.PricingCoverage.CoveragePct()),
			))
		case quality.ParseFailures > 0:
			lines = append(lines, m.theme.Warning.Render("Parse failures"))
			lines = append(lines, m.theme.Insight.Render(fmt.Sprintf("%d failures in active window", quality.ParseFailures)))
		default:
			lines = append(lines, m.theme.Positive.Render("Data quality healthy"))
			lines = append(lines, m.theme.Insight.Render("Attribution and pricing look stable"))
		}

		if weak := firstBreakdown(m.data.Dashboard.LowSavingsCommands); weak != nil {
			lines = append(lines, "")
			lines = append(lines, m.theme.Warning.Render("Weakest command"))
			lines = append(lines, m.theme.Insight.Render(fmt.Sprintf("%s at %.1f%% reduction", weak.Key, weak.ReductionPct)))
		}
		if provider := firstBreakdown(m.data.Dashboard.TopProviders); provider != nil {
			lines = append(lines, "")
			lines = append(lines, m.theme.Positive.Render("Top provider"))
			lines = append(lines, m.theme.Insight.Render(fmt.Sprintf("%s saved %s tokens", provider.Key, formatInt(provider.SavedTokens))))
		}
	}

	lines = append(lines, "")
	lines = append(lines, m.theme.SectionTitle.Render("Controls"))
	lines = append(lines, m.theme.Insight.Render(fmt.Sprintf("1-%d jump sections", len(m.sections))))
	lines = append(lines, m.theme.Insight.Render("tab switch focus"))
	lines = append(lines, m.theme.Insight.Render("r refresh"))
	lines = append(lines, m.theme.Insight.Render("? help"))

	return setWidth(m.theme.RightPane, width).Render(strings.Join(lines, "\n"))
}

func (m model) renderFooter(width int) string {
	parts := make([]string, 0, len(m.keys.ShortHelp())*3)
	for i, b := range m.keys.ShortHelp() {
		if i > 0 {
			parts = append(parts, "  ")
		}
		parts = append(parts, m.theme.FooterKey.Render(b.Help().Key), " "+b.Help().Desc)
	}
	return setWidth(m.theme.Footer, width).Render(lipgloss.JoinHorizontal(lipgloss.Left, parts...))
}

func (m model) renderHelp(width int) string {
	lines := []string{
		m.theme.SectionTitle.Render("tok TUI Help"),
		m.theme.Muted.Render("Keybindings are generated from the registry; see internal/tui/keys.go."),
		"",
	}
	columns := m.keys.FullHelp()
	// Render each column vertically, then stack columns horizontally so the
	// overlay scales to terminal width without truncation logic here.
	rendered := make([]string, 0, len(columns))
	for _, col := range columns {
		var b strings.Builder
		for _, binding := range col {
			h := binding.Help()
			b.WriteString(m.theme.FooterKey.Render(h.Key))
			b.WriteString("  ")
			b.WriteString(m.theme.Insight.Render(h.Desc))
			b.WriteString("\n")
		}
		rendered = append(rendered, b.String())
	}
	help := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	return strings.Join(lines, "\n") + "\n" + help
}

func firstBreakdown(items []tracking.DashboardBreakdown) *tracking.DashboardBreakdown {
	if len(items) == 0 {
		return nil
	}
	return &items[0]
}

func setWidth(style lipgloss.Style, total int) lipgloss.Style {
	return style.Width(max(0, total-style.GetHorizontalFrameSize()))
}

func joinHorizontalGap(gap string, items ...string) string {
	if len(items) == 0 {
		return ""
	}
	out := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			out = append(out, gap)
		}
		out = append(out, item)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}
