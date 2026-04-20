package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

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
	sections   []SectionRenderer
	actions    *ActionRegistry
	toasts     *toastStack
	palette    *Palette
	search     *SearchOverlay
	confirm    *ConfirmOverlay
	logs       *ringHandler
	prevLogger *slog.Logger
	env        Environment
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
	sections := defaultSections()

	// Route slog through an in-memory ring so the Logs section has
	// something to display. We keep the original default logger aside
	// and restore it on Close so CLI commands after TUI exit (e.g.
	// in test harnesses) don't lose stderr output.
	prev := slog.Default()
	ring := NewRingHandler(512, slog.LevelDebug, prev.Handler())
	slog.SetDefault(slog.New(ring))

	normalized := opts.normalized()
	m := model{
		opts:       normalized,
		loader:     loader,
		ctx:        ctx,
		cancel:     cancel,
		keys:       DefaultKeyMap(),
		theme:      newThemeByName(normalized.Theme),
		sections:   sections,
		toasts:     &toastStack{},
		search:     NewSearchOverlay(),
		confirm:    NewConfirmOverlay(),
		logs:       ring,
		prevLogger: prev,
		env:        DetectEnvironment(),
		loading:    true,
		spinner:    s,
	}

	// Build the action registry with callbacks that close over the model
	// via message-dispatch, not pointer capture — a SectionRenderer or
	// action handler must never touch the model directly.
	m.actions = DefaultActionRegistry(ActionDeps{
		SectionCount: len(sections),
		RequestRefresh: func() tea.Cmd {
			return func() tea.Msg { return actionRequestMsg{ActionID: "view.refresh"} }
		},
		RequestJump: func(idx int) tea.Cmd {
			return func() tea.Msg { return drillDownMsg{Section: idx} }
		},
		RequestToast: requestToastCmd,
		RequestTheme: func(name ThemeName) tea.Cmd {
			return func() tea.Msg { return themeChangedMsg{Name: name} }
		},
		RequestThemeCycle: func() tea.Cmd {
			return func() tea.Msg { return themeCycleMsg{} }
		},
		ClearLogRing: func() tea.Cmd {
			return func() tea.Msg {
				if ring != nil {
					ring.Clear()
				}
				return nil
			}
		},
	})

	// Build the palette last so it can reference both the registry and
	// the section list built above.
	m.palette = NewPalette(m.actions, sections)

	return m
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
		// Overlays capture input while open. Palette, search, and
		// confirm are mutually exclusive — only one can be focused at
		// a time. Confirm wins over the others because it's armed by
		// an action that was itself routed through palette/keybind.
		if m.confirm != nil && m.confirm.IsOpen() {
			if cmd := m.confirm.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}
		if m.palette != nil && m.palette.IsOpen() {
			if cmd := m.palette.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}
		if m.search != nil && m.search.IsOpen() {
			if cmd := m.search.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}
		var cmd tea.Cmd
		m, cmd = m.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case toastAddMsg:
		if m.toasts != nil {
			if cmd := m.toasts.add(msg.Kind, msg.Text, msg.TTL); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case toastExpireMsg:
		if m.toasts != nil {
			m.toasts.expire(msg.ID)
		}
	case actionRequestMsg:
		if m.actions != nil {
			// Intercept Confirm=true actions: open the modal and wait
			// for the user's decision. The confirm overlay re-emits
			// actionRequestMsg on accept, which then falls through to
			// runActionCmd via this same case (but with Confirm already
			// resolved — short-circuit on that by consulting a sentinel
			// flag on the modal? simpler: overlay clears itself before
			// emitting, so when the accept message arrives IsOpen is
			// false and we drop straight into Run).
			if a, ok := m.actions.Get(msg.ActionID); ok && a.Confirm && m.confirm != nil && !m.confirm.IsOpen() {
				m.confirm.Open(a, msg.Args)
				return m, tea.Batch(cmds...)
			}
			cmds = append(cmds, runActionCmd(m.ctx, m.actions, msg.ActionID, msg.Args))
		}
	case actionResultMsg:
		// For Phase 1 we report outcomes via toast; Phase 2 sections can
		// intercept these to refresh themselves selectively.
		if msg.Err != nil {
			cmds = append(cmds, requestToastCmd(ToastError, fmt.Sprintf("%s: %s", msg.ActionID, msg.Err)))
		}
	case paletteExecMsg:
		if m.actions != nil {
			cmds = append(cmds, runActionCmd(m.ctx, m.actions, msg.ActionID, msg.Args))
		}
	case paletteCloseMsg:
		// No-op; palette already closed itself before emitting the msg.
	case searchCloseMsg:
		// No-op; search overlay already cleared state.
	case themeChangedMsg:
		m.theme = newThemeByName(msg.Name)
		m.opts.Theme = msg.Name
	case themeCycleMsg:
		next := AvailableThemes[0]
		for i, t := range AvailableThemes {
			if t == m.opts.Theme {
				next = AvailableThemes[(i+1)%len(AvailableThemes)]
				break
			}
		}
		m.theme = newThemeByName(next)
		m.opts.Theme = next
		cmds = append(cmds, requestToastCmd(ToastInfo, "theme: "+string(next)))
	case drillDownMsg:
		// Section index from a drill-down acts as a jump for Phase 1
		// where no section yet drills into a detail pane.
		if msg.Section >= 0 && msg.Section < len(m.sections) {
			m.navIndex = msg.Section
		}
	}

	// Forward the message to the currently-focused section so it can
	// react (cursor moves, row selection, search typing, etc.). This
	// runs *after* the root handlers so root-level decisions (quit,
	// nav) take priority. Section updates never receive tea.QuitMsg or
	// quitMsg because those short-circuit the switch above.
	if m.ready && len(m.sections) > 0 && !m.quitting {
		ctx := m.sectionContext()
		focused := m.sections[m.navIndex]
		next, cmd := focused.Update(ctx, msg)
		if next != nil {
			m.sections[m.navIndex] = next
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// sectionContext builds the read-only context handed to each section.
func (m model) sectionContext() SectionContext {
	// Reserve space for sidebar + right pane the same way renderBody does.
	width := max(24, m.width)
	if !m.compact {
		width -= 19 // sidebar + gap
		if m.width >= 170 && m.navIndex != 0 {
			width -= 27 // right pane + gap
		}
		if width < 24 {
			width = 24
		}
	}
	return SectionContext{
		Theme:   m.theme,
		Keys:    m.keys,
		Data:    m.data,
		Opts:    m.opts,
		Width:   width,
		Height:  max(8, m.height-6), // header + footer
		Compact: m.compact,
		Focused: true,
		Logs:    m.logs,
		Env:     m.env,
	}
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading tok TUI…"
	}

	contentWidth := max(20, m.width)
	header := m.renderHeader(contentWidth)
	footer := m.renderFooter(contentWidth)

	// Toast region sits between the body and the footer so it never
	// clobbers the main view. Right-aligned to mimic a conventional
	// notification stack. Phase 3 can replace this with a true overlay
	// (lipgloss.Place composited on top of the body).
	toastView := ""
	toastHeight := 0
	if m.toasts != nil && len(m.toasts.items) > 0 {
		block := m.toasts.render(m.theme, contentWidth)
		toastView = lipgloss.PlaceHorizontal(contentWidth, lipgloss.Right, block)
		toastHeight = lipgloss.Height(toastView)
	}

	bodyHeight := max(8, m.height-lipgloss.Height(header)-lipgloss.Height(footer)-toastHeight)
	body := m.renderBody(contentWidth, bodyHeight)

	parts := []string{header, body}
	if toastView != "" {
		parts = append(parts, toastView)
	}
	parts = append(parts, footer)
	view := lipgloss.JoinVertical(lipgloss.Left, parts...)
	rendered := setWidth(m.theme.App, m.width).Render(view)

	// Modals: confirm wins over palette wins over base frame. Both use
	// the same centered lipgloss.Place composition so the dimensions
	// and transition feel consistent.
	if m.confirm != nil && m.confirm.IsOpen() {
		modal := m.confirm.View(m.theme, m.width)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	}
	if m.palette != nil && m.palette.IsOpen() {
		modal := m.palette.View(m.theme, m.width)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	}
	return rendered
}

// handleKey dispatches a key press against the keymap registry, mutates the
// model as needed, and returns any resulting tea.Cmd. Kept off Update so the
// message switch stays readable and so section-local handlers can layer on
// in Phase 1 by calling back into this from their own Update methods.
func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, shutdownCmd(m.cancel, m.loader, m.prevLogger)
	case key.Matches(msg, m.keys.Help):
		m.helpOpen = !m.helpOpen
	case key.Matches(msg, m.keys.Esc):
		m.helpOpen = false
	case key.Matches(msg, m.keys.Palette):
		if m.palette != nil {
			return m, m.palette.Open()
		}
	case key.Matches(msg, m.keys.Search):
		if m.search != nil {
			return m, m.search.Open()
		}
	case key.Matches(msg, m.keys.NextSection):
		m.navIndex = (m.navIndex + 1) % len(m.sections)
	case key.Matches(msg, m.keys.PrevSection):
		m.navIndex = (m.navIndex - 1 + len(m.sections)) % len(m.sections)
	case key.Matches(msg, m.keys.Refresh):
		m.loading = m.data == nil
		m.refreshing = m.data != nil
		return m, loadSnapshotCmd(m.ctx, m.loader, m.opts)
	case key.Matches(msg, m.keys.Export):
		if m.navIndex < len(m.sections) {
			if exp, ok := m.sections[m.navIndex].(ExportableTable); ok {
				return m, ExportTableCmd(exp, ExportJSON)
			}
		}
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

// shutdownCmd cancels in-flight loads, closes loader-held DB handles,
// and restores the previous default slog logger, then dispatches
// quitMsg so tea.Quit runs only after teardown is complete.
func shutdownCmd(cancel context.CancelFunc, loader snapshotLoader, prevLogger *slog.Logger) tea.Cmd {
	return func() tea.Msg {
		if cancel != nil {
			cancel()
		}
		if loader != nil {
			_ = loader.Close()
		}
		if prevLogger != nil {
			slog.SetDefault(prevLogger)
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
	sectionTitle := m.theme.Focus.Render(m.sections[m.navIndex].Name())

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
		text := lipgloss.JoinHorizontal(lipgloss.Left, m.theme.SidebarKey.Render(label), " ", s.Name())
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
		label := s.Name()
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
	// Dispatch to the focused section via the SectionRenderer interface.
	// We rebuild a SectionContext here with the computed main-pane width
	// so sections don't have to re-derive layout from the raw model.
	ctx := SectionContext{
		Theme:   m.theme,
		Keys:    m.keys,
		Data:    m.data,
		Opts:    m.opts,
		Width:   width,
		Height:  height,
		Compact: m.compact,
		Focused: true,
		Logs:    m.logs,
		Env:     m.env,
	}
	return setWidth(m.theme.Main, width).Render(m.sections[m.navIndex].View(ctx))
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
