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

	"github.com/GrayCodeAI/tok/internal/tracking"
)

type snapshotLoadedMsg struct {
	snapshot *tracking.WorkspaceDashboardSnapshot
	err      error
	loadedAt time.Time
}

type refreshTickMsg time.Time

// liveEventMsg is dispatched by the liveSource goroutine whenever the
// underlying tracking data has changed. The TUI reacts by triggering a
// snapshot reload and (when Record is set) popping a toast so the user
// sees instant feedback on their own tok runs.
type liveEventMsg LiveEvent

// quitMsg is dispatched after the loader has been closed so the program can
// exit cleanly without leaking DB handles.
type quitMsg struct{}

type model struct {
	opts          Options
	loader        snapshotLoader
	ctx           context.Context
	cancel        context.CancelFunc
	keys          KeyMap
	theme         theme
	sections      []SectionRenderer
	actions       *ActionRegistry
	toasts        *toastStack
	palette       *Palette
	search        *SearchOverlay
	filter        *FilterOverlay
	confirm       *ConfirmOverlay
	logs          *ringHandler
	prevLogger    *slog.Logger
	env           Environment
	history       *historyStack // navigation history for H/L back/forward
	navIndex      int
	scrollOffsets []int // per-section vertical scroll offset for clipping in renderMain
	width         int
	height        int
	ready         bool
	compact       bool
	helpOpen      bool
	loading       bool
	refreshing    bool
	quitting      bool
	err           error
	lastLoad      time.Time
	lastLive      time.Time // last time any live event arrived (subscribe/fsnotify/tick)
	liveCount     int       // number of Record events since TUI start (for live badge)
	liveFeed      []LiveFeedEntry
	data          *tracking.WorkspaceDashboardSnapshot
	spinner       spinner.Model
	live          liveSource

	// Budget tracking for threshold alerts
	dailySpent    int  // tokens spent today
	budgetAlerted bool // true if we've shown the budget toast already
}

// LiveFeedEntry is one entry in the recent-activity ring buffer rendered
// on Home. Captures just enough to show "+123 saved · git status · 2s ago"
// without holding the full CommandRecord.
type LiveFeedEntry struct {
	At          time.Time
	Command     string
	SavedTokens int
}

// maxLiveFeed bounds the ring buffer. 20 is enough for the Live Feed
// panel (shows 6–8) plus a little history for scrolling later.
const maxLiveFeed = 20

func NewModel(opts Options) tea.Model {
	return newModelWith(opts, newWorkspaceLoader(), newTrackingLiveSource())
}

// NewModelWithLoader is the testable constructor: tests inject a stubLoader
// while production paths go through NewModel which owns the real DB-backed
// workspaceLoader. Tests get a nullLiveSource so no goroutines leak.
func NewModelWithLoader(opts Options, loader snapshotLoader) tea.Model {
	return newModelWith(opts, loader, nullLiveSource{})
}

// NewModelWithLive is a testable constructor that accepts both a stub
// loader and a custom live source — used by the live-update integration
// test to drive synthetic events without opening a real DB.
func NewModelWithLive(opts Options, loader snapshotLoader, live liveSource) tea.Model {
	return newModelWith(opts, loader, live)
}

func newModelWith(opts Options, loader snapshotLoader, live liveSource) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7"))

	ctx, cancel := context.WithCancel(context.Background())
	sections := defaultSections()

	// Route slog through an in-memory ring so the Logs section has
	// something to display. The previous default logger is stashed
	// and restored on shutdown (see shutdownCmd).
	//
	// IMPORTANT: do NOT tee the ring to the prior default handler
	// while the TUI is active. The default handler writes JSON to
	// os.Stderr, and every write over stderr smears the alt-screen
	// frame with log bytes (users report "garbled overlapping headers"
	// — that's stderr leaking through). Logs are preserved in the
	// ring and visible in the Logs section; restore the old delegate
	// on TUI exit so post-TUI CLI code still logs to stderr.
	prev := slog.Default()
	ring := NewRingHandler(512, slog.LevelDebug, nil)
	slog.SetDefault(slog.New(ring))

	normalized := opts.normalized()

	// Load keybindings: merge user overrides with defaults
	km := DefaultKeyMap()
	if hasKeybindings(normalized.Keybindings) {
		if loaded, err := LoadKeyMap(normalized.Keybindings); err == nil {
			km = loaded
		}
	}

	m := model{
		opts:          normalized,
		loader:        loader,
		live:          live,
		ctx:           ctx,
		cancel:        cancel,
		keys:          km,
		theme:         newThemeByName(normalized.Theme),
		sections:      sections,
		scrollOffsets: make([]int, len(sections)),
		history:       newHistoryStack(50),
		toasts:        &toastStack{},
		search:        NewSearchOverlay(),
		filter:        NewFilterOverlay(),
		confirm:       NewConfirmOverlay(),
		logs:          ring,
		prevLogger:    prev,
		env:           DetectEnvironment(),
		loading:       true,
		spinner:       s,
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
		liveEventCmd(m.ctx, m.live),
	)
}

// liveEventCmd pumps the next event from the live source into the tea
// loop. It re-schedules itself from the Update switch so the main loop
// sees a stream of liveEventMsg without us owning a raw goroutine.
func liveEventCmd(ctx context.Context, live liveSource) tea.Cmd {
	if live == nil {
		return nil
	}
	ch := live.Start(ctx)
	return waitForLiveEvent(ch)
}

func waitForLiveEvent(ch <-chan LiveEvent) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		// Carry the channel forward via a follow-up cmd so the next
		// call to Update re-pumps without us starting a new source.
		return liveEventWithChan{ev: liveEventMsg(ev), ch: ch}
	}
}

// liveEventWithChan pairs the event with the channel it came from so we
// can re-subscribe on the next tick without re-Start()ing the source.
type liveEventWithChan struct {
	ev liveEventMsg
	ch <-chan LiveEvent
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
	case liveEventWithChan:
		if m.quitting {
			return m, nil
		}
		m.lastLive = msg.ev.At
		if msg.ev.Source == "subscribe" && msg.ev.Record != nil {
			m.liveCount++
			m.liveFeed = appendLiveFeed(m.liveFeed, LiveFeedEntry{
				At:          msg.ev.At,
				Command:     msg.ev.Record.Command,
				SavedTokens: msg.ev.Record.SavedTokens,
			})
			// Only pop a toast for truly new commands — fsnotify and
			// tick events fire from any write including unrelated ones
			// and would quickly become noise.
			cmds = append(cmds, requestToastCmd(
				ToastSuccess,
				formatLiveRecordToast(msg.ev.Record),
			))
		}
		// Kick off a snapshot reload so the next frame reflects the
		// new data. Coalesce: skip if one's already in-flight.
		if !m.loading && !m.refreshing {
			m.refreshing = true
			cmds = append(cmds, loadSnapshotCmd(m.ctx, m.loader, m.opts))
		}
		// Re-pump the channel for the next event.
		cmds = append(cmds, waitForLiveEvent(msg.ch))
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
			// Update daily spending and check budget thresholds
			m.updateDailySpending()
			cmd := m.checkBudgetAlert()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
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
		if m.filter != nil && m.filter.IsOpen() {
			if cmd := m.filter.Update(msg); cmd != nil {
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
	case filterOverlayMsg:
		// Apply filter to current section's table if it has one
		if m.navIndex < len(m.sections) {
			if t, ok := m.sections[m.navIndex].(interface{ SetFilter(string) }); ok {
				t.SetFilter(msg.Query)
			}
		}
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
			m.history.PushSection(m.navIndex, m.scrollOffsets[m.navIndex])
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

// updateDailySpending calculates today's token usage from the snapshot.
func (m *model) updateDailySpending() {
	if m.data == nil || m.opts.Budget.DailyTokens <= 0 {
		return
	}

	// Use the pre-calculated daily filtered tokens from Budgets
	if m.data.Dashboard != nil {
		m.dailySpent = int(m.data.Dashboard.Budgets.Daily.FilteredTokens)
	}
}

// checkBudgetAlert returns a toast command if budget threshold crossed.
func (m *model) checkBudgetAlert() tea.Cmd {
	if m.opts.Budget.DailyTokens <= 0 {
		return nil
	}

	warningThreshold := m.opts.Budget.WarningThreshold
	if warningThreshold <= 0 {
		warningThreshold = 80
	}

	pct := float64(m.dailySpent) * 100.0 / float64(m.opts.Budget.DailyTokens)

	// Budget exceeded - critical alert
	if pct >= 100 && !m.budgetAlerted {
		m.budgetAlerted = true
		return requestToastCmd(ToastError,
			fmt.Sprintf("Budget exceeded: %d%% (%s/%s tokens)",
				int(pct),
				formatCompactNumber(m.dailySpent),
				formatCompactNumber(m.opts.Budget.DailyTokens)))
	}

	// Warning threshold crossed
	if pct >= float64(warningThreshold) && !m.budgetAlerted {
		m.budgetAlerted = true
		return requestToastCmd(ToastWarning,
			fmt.Sprintf("Budget warning: %d%% used (%s/%s tokens)",
				int(pct),
				formatCompactNumber(m.dailySpent),
				formatCompactNumber(m.opts.Budget.DailyTokens)))
	}

	// Reset alert if we're back under warning threshold
	if pct < float64(warningThreshold)*0.9 {
		m.budgetAlerted = false
	}

	return nil
}

// formatCompactNumber returns compact number display (1.2K, 3.4M, etc).
func formatCompactNumber(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// sectionContext builds the read-only context handed to each section
// during the Update path (no View is being rendered yet — this ctx is
// only used for size-aware key handling and data access). Width must
// match what renderMain eventually passes at render time so cursor
// clamping doesn't drift between Update and View.
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
	// Shrink by Main's frame so sections see the same inner width in
	// Update that they'll receive in View. Without this, cursor / row
	// clamps computed on the wider Update-time width don't match the
	// narrower render-time width, and selection can fall off screen.
	innerWidth := max(1, width-m.theme.Main.GetHorizontalFrameSize())
	innerHeight := max(1, m.height-6-m.theme.Main.GetVerticalFrameSize())
	if innerHeight < 8 {
		innerHeight = 8
	}
	return SectionContext{
		Theme:      m.theme,
		Keys:       m.keys,
		Data:       m.data,
		Opts:       m.opts,
		Width:      innerWidth,
		Height:     innerHeight,
		Compact:    m.compact,
		Focused:    true,
		Logs:       m.logs,
		Env:        m.env,
		LiveFeed:   m.liveFeed,
		DailySpent: m.dailySpent,
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
	if m.filter != nil && m.filter.IsOpen() {
		modal := m.filter.View(m.theme, m.width)
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
	case key.Matches(msg, m.keys.Filter):
		if m.filter != nil {
			return m, m.filter.Open()
		}
	case key.Matches(msg, m.keys.NextSection):
		m.history.PushSection(m.navIndex, m.scrollOffsets[m.navIndex])
		m.navIndex = (m.navIndex + 1) % len(m.sections)
	case key.Matches(msg, m.keys.PrevSection):
		m.history.PushSection(m.navIndex, m.scrollOffsets[m.navIndex])
		m.navIndex = (m.navIndex - 1 + len(m.sections)) % len(m.sections)
	case key.Matches(msg, m.keys.HistoryBack):
		if m.history.CanGoBack() {
			if entry, ok := m.history.Back(); ok {
				m.navIndex = entry.SectionIndex
				m.scrollOffsets[m.navIndex] = entry.ScrollOffset
			}
		}
	case key.Matches(msg, m.keys.HistoryForward):
		if m.history.CanGoForward() {
			if entry, ok := m.history.Forward(); ok {
				m.navIndex = entry.SectionIndex
				m.scrollOffsets[m.navIndex] = entry.ScrollOffset
			}
		}
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
			m.history.PushSection(m.navIndex, m.scrollOffsets[m.navIndex])
			m.navIndex = idx
		}
	case key.Matches(msg, m.keys.PageUp):
		m.scrollBy(-m.pageStep())
	case key.Matches(msg, m.keys.PageDn):
		m.scrollBy(m.pageStep())
	case key.Matches(msg, m.keys.Up):
		if m.sectionWantsScroll() {
			m.scrollBy(-1)
		}
	case key.Matches(msg, m.keys.Down):
		if m.sectionWantsScroll() {
			m.scrollBy(1)
		}
	case key.Matches(msg, m.keys.Top):
		if m.sectionWantsScroll() {
			m.scrollOffsets[m.navIndex] = 0
		}
	case key.Matches(msg, m.keys.Bottom):
		if m.sectionWantsScroll() {
			m.scrollOffsets[m.navIndex] = m.maxOffsetForCurrent()
		}
	}
	return m, nil
}

// scrollableSection marks a section whose View output is a static block
// the root should scroll via clipToHeight (no internal cursor). Sections
// with tables embed their own viewport instead and do NOT implement this
// marker so their arrow/pgup/pgdn keys move the table cursor, not the
// frame.
type scrollableSection interface {
	IsScrollable() bool
}

func (m model) sectionWantsScroll() bool {
	if m.navIndex < 0 || m.navIndex >= len(m.sections) {
		return false
	}
	s, ok := m.sections[m.navIndex].(scrollableSection)
	return ok && s.IsScrollable()
}

func (m *model) scrollBy(delta int) {
	if m.navIndex < 0 || m.navIndex >= len(m.scrollOffsets) {
		return
	}
	offset := m.scrollOffsets[m.navIndex] + delta
	if offset < 0 {
		offset = 0
	}
	if maxOff := m.maxOffsetForCurrent(); offset > maxOff {
		offset = maxOff
	}
	m.scrollOffsets[m.navIndex] = offset
}

// pageStep is the step size for PageUp/PageDn — roughly one screen
// minus two lines of overlap so the reader can keep a landmark in view.
func (m model) pageStep() int {
	_, h := m.mainInnerDims()
	step := h - 2
	if step < 1 {
		step = 1
	}
	return step
}

// mainInnerDims returns the inner (content) width and height of the Main
// pane, matching renderBody's layout math so scroll bookkeeping uses the
// same coordinates the render path uses.
func (m model) mainInnerDims() (int, int) {
	width := max(20, m.width)
	sidebarWidth := 20
	gap := 1
	rightWidth := 0
	if !m.compact {
		showRight := width >= 170 && m.navIndex != 0
		if showRight {
			rightWidth = 26
		}
		width = max(24, width-sidebarWidth-rightWidth-gap)
	}
	innerWidth := max(1, width-m.theme.Main.GetHorizontalFrameSize())

	bodyHeight := max(8, m.height-2-1) // header(1) + footer(1) rough; recomputed below
	// Re-derive body height the way View() does: header + footer + toasts.
	header := m.renderHeader(max(20, m.width))
	footer := m.renderFooter(max(20, m.width))
	toastHeight := 0
	if m.toasts != nil && len(m.toasts.items) > 0 {
		toastHeight = 1
	}
	bodyHeight = max(8, m.height-lipgloss.Height(header)-lipgloss.Height(footer)-toastHeight)
	if m.compact {
		// Compact mode reserves one line for the tabs strip above Main.
		bodyHeight = max(4, bodyHeight-1)
	}
	innerHeight := max(1, bodyHeight-m.theme.Main.GetVerticalFrameSize())
	return innerWidth, innerHeight
}

func (m model) maxOffsetForCurrent() int {
	if m.navIndex < 0 || m.navIndex >= len(m.sections) {
		return 0
	}
	innerW, innerH := m.mainInnerDims()
	ctx := SectionContext{
		Theme:    m.theme,
		Keys:     m.keys,
		Data:     m.data,
		Opts:     m.opts,
		Width:    innerW,
		Height:   innerH,
		Compact:  m.compact,
		Focused:  true,
		Logs:     m.logs,
		Env:      m.env,
		LiveFeed: m.liveFeed,
	}
	rendered := m.sections[m.navIndex].View(ctx)
	lines := strings.Count(rendered, "\n") + 1
	return maxScrollOffset(lines, innerH)
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
	// "0" is a shortcut for section 10, mirroring vim's treatment of 0
	// as a standalone column rather than a digit. Keeps single-stroke
	// navigation viable past section 9 without introducing a prefix
	// state machine.
	if key == "0" {
		if sectionCount >= 10 {
			return 9, true
		}
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
	// Right-pad to 2 chars so 1-digit shortcuts ("1"–"9") line up with
	// 2-digit shortcuts ("10"–"12") in the sidebar. Without this the
	// section name column shifts by one when you scroll past section 9.
	n := index + 1
	if n < 10 {
		return " " + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
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

	// Live badge: "● live" when we've received any event in the last
	// 30 seconds; "○ idle" otherwise. Colored via Positive/HeaderMuted.
	// This is the primary "am I seeing real-time data?" affordance.
	statusParts = append(statusParts, m.renderLiveBadge())

	left := title + "  " + strings.Join(statusParts, m.theme.HeaderMuted.Render(" · "))
	return setWidth(m.theme.Header, width).Render(left)
}

func (m model) renderBody(width, height int) string {
	if m.compact {
		tabs := m.renderCompactTabs(width)
		main := m.renderMain(width, max(4, height-lipgloss.Height(tabs)))
		return lipgloss.JoinVertical(lipgloss.Left, tabs, main)
	}

	sidebarWidth := 20
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
		// Both rows use a plain ASCII 2-char marker so columns line up
		// regardless of (a) 1-digit vs 2-digit shortcut label or (b)
		// active vs inactive state. Active row gets color + bold via
		// the SidebarActive style; the glyph itself is "> " either way
		// so width never shifts. Unicode glyphs have flaky
		// column-width measurement in some terminals, so we avoid them
		// in the sidebar prefix specifically.
		marker := "  "
		style := m.theme.SidebarItem
		if i == m.navIndex {
			marker = "> "
			style = m.theme.SidebarActive
		}
		text := marker + m.theme.SidebarKey.Render(label) + " " + s.Name()
		// The inner width available for each row is the sidebar width
		// minus the Sidebar container's frame (Padding + BorderRight).
		// Using a fixed `-2` underestimated it by the right-border, which
		// caused longer section names (Providers/Sessions/Commands/
		// Pipeline) to wrap onto a second line without their shortcut
		// number.
		inner := width - m.theme.Sidebar.GetHorizontalFrameSize()
		if inner < 1 {
			inner = 1
		}
		lines = append(lines, setWidth(style, inner).Render(text))
	}
	lines = append(lines, "")
	lines = append(lines, m.theme.Muted.Render("Dashboard"))

	return setWidth(m.theme.Sidebar, width).Render(strings.Join(lines, "\n"))
}

func (m model) renderCompactTabs(width int) string {
	// Try to render all labels on one line, bracketing the active one.
	// On narrow terminals (≤100 cols) the full strip wraps, eating main
	// pane height, so fall back to a breadcrumb of "[N/M] Name  ‹ prev ·
	// next ›" which is always one line and makes nav affordances explicit.
	fullParts := make([]string, 0, len(m.sections))
	for i, s := range m.sections {
		label := s.Name()
		if i == m.navIndex {
			fullParts = append(fullParts, m.theme.SidebarActive.Render("["+label+"]"))
		} else {
			fullParts = append(fullParts, m.theme.SidebarItem.Render(" "+label+" "))
		}
	}
	full := strings.Join(fullParts, " ")
	if lipgloss.Width(full) <= width-2 {
		return setWidth(lipgloss.NewStyle().Padding(0, 1), width).Render(full)
	}

	// Breadcrumb fallback: keep the whole line under `width` even on 60-col
	// terminals. Format: "‹ shift-tab · [3/12] Trends · tab ›".
	current := m.sections[m.navIndex].Name()
	breadcrumb := fmt.Sprintf("‹ shift-tab · %s[%d/%d] %s%s · tab ›",
		"",
		m.navIndex+1, len(m.sections),
		current,
		"",
	)
	styled := m.theme.SidebarActive.Render(breadcrumb)
	if lipgloss.Width(styled) > width {
		// Even narrower: drop the arrows.
		styled = m.theme.SidebarActive.Render(
			fmt.Sprintf("[%d/%d] %s", m.navIndex+1, len(m.sections), current),
		)
	}
	return setWidth(lipgloss.NewStyle().Padding(0, 1), width).Render(styled)
}

func (m model) renderMain(width, height int) string {
	// The Main style has Padding(0,1) so its *inner* content area is
	// narrower than `width`. Section renderers lay out panels using the
	// width we hand them — if we pass the outer width, their panels are
	// wider than Main's inner area and lipgloss soft-wraps the borders,
	// producing the dangling `─┘` + "points" on its own line that users
	// report as a "broken" frame. Always pass the inner width.
	innerWidth := max(1, width-m.theme.Main.GetHorizontalFrameSize())
	innerHeight := max(1, height-m.theme.Main.GetVerticalFrameSize())

	if m.helpOpen {
		return setWidth(m.theme.Main, width).Render(m.renderHelp(innerWidth))
	}
	if m.err != nil && m.data == nil {
		body := m.theme.Danger.Render("Failed to load snapshot") + "\n\n" +
			m.err.Error() + "\n\n" +
			m.theme.Muted.Render("Press 'r' to retry, or 'q' to quit.")
		return setWidth(m.theme.Main, width).Render(body)
	}
	// Sections are responsible for their own empty state. Home renders an
	// onboarding panel when Data is nil; other sections render "No data".
	// The header spinner already signals "loading", so a dedicated loading
	// body would just hide the first-run welcome screen behind a spinner.
	ctx := SectionContext{
		Theme:    m.theme,
		Keys:     m.keys,
		Data:     m.data,
		Opts:     m.opts,
		Width:    innerWidth,
		Height:   innerHeight,
		Compact:  m.compact,
		Focused:  true,
		Logs:     m.logs,
		Env:      m.env,
		LiveFeed: m.liveFeed,
	}
	rendered := m.sections[m.navIndex].View(ctx)
	offset := m.scrollOffsets[m.navIndex]
	totalLines := strings.Count(rendered, "\n") + 1

	// Scroll hint: reserve one line at the bottom of the main pane so the
	// indicator never covers a content row. Only reserve when the section
	// is scrollable and the content actually exceeds the viewport.
	wantsHint := m.sectionWantsScroll() && totalLines > innerHeight
	contentHeight := innerHeight
	if wantsHint {
		contentHeight = max(1, innerHeight-1)
	}

	body := clipToHeight(rendered, contentHeight, offset)
	if wantsHint {
		visibleEnd := offset + contentHeight
		if visibleEnd > totalLines {
			visibleEnd = totalLines
		}
		var hint string
		switch {
		case offset > 0 && visibleEnd < totalLines:
			hint = fmt.Sprintf("▲ %d above · %d/%d · ▼ %d below · pgup/pgdn",
				offset, visibleEnd, totalLines, totalLines-visibleEnd)
		case offset > 0:
			hint = fmt.Sprintf("▲ %d above · %d/%d (end) · pgup to scroll back",
				offset, totalLines, totalLines)
		default:
			hint = fmt.Sprintf("%d/%d · ▼ %d below · pgdn or j to scroll",
				contentHeight, totalLines, totalLines-visibleEnd)
		}
		body = body + "\n" + m.theme.Muted.Render(hint)
	}
	return setWidth(m.theme.Main, width).Render(body)
}

// clipToHeight limits the rendered section output to `height` lines,
// shifted by `offset` so content below the viewport can be scrolled
// into view. Splits on "\n" which is safe here because lipgloss emits
// ANSI SGR pairs on a single line (each line self-closes).
func clipToHeight(s string, height, offset int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if offset < 0 {
		offset = 0
	}
	if offset >= len(lines) {
		offset = max(0, len(lines)-1)
	}
	lines = lines[offset:]
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

// maxScrollOffset returns the largest allowed scroll offset for a
// rendered block of lineCount lines inside a viewport of viewportHeight.
func maxScrollOffset(lineCount, viewportHeight int) int {
	if lineCount <= viewportHeight {
		return 0
	}
	return lineCount - viewportHeight
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

	// The footer already advertises every short-help binding; listing
	// them again here wastes the right-pane's screen real estate.
	// Keep the insights panel for data-quality signals only.

	return setWidth(m.theme.RightPane, width).Render(strings.Join(lines, "\n"))
}

func (m model) renderFooter(width int) string {
	// Render the short-help list, shedding entries from the right if
	// the whole strip would exceed the terminal width. Keeps the most
	// useful bindings (nav, quit) visible on narrow terminals instead
	// of letting the strip wrap to a second line.
	bindings := m.keys.ShortHelp()
	for n := len(bindings); n >= 1; n-- {
		parts := make([]string, 0, n*3)
		for i := 0; i < n; i++ {
			if i > 0 {
				parts = append(parts, "  ")
			}
			parts = append(parts, m.theme.FooterKey.Render(bindings[i].Help().Key), " "+bindings[i].Help().Desc)
		}
		line := lipgloss.JoinHorizontal(lipgloss.Left, parts...)
		if lipgloss.Width(line) <= width-2 { // 2 = footer padding
			return setWidth(m.theme.Footer, width).Render(line)
		}
	}
	// Even one binding doesn't fit — render quit alone as a last resort.
	return setWidth(m.theme.Footer, width).Render(m.theme.FooterKey.Render("q") + " quit")
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

// appendLiveFeed prepends entry to feed and caps at maxLiveFeed. Newest
// first so Home's panel shows most-recent on top without re-sorting.
func appendLiveFeed(feed []LiveFeedEntry, entry LiveFeedEntry) []LiveFeedEntry {
	out := make([]LiveFeedEntry, 0, len(feed)+1)
	out = append(out, entry)
	out = append(out, feed...)
	if len(out) > maxLiveFeed {
		out = out[:maxLiveFeed]
	}
	return out
}

// renderLiveBadge returns a short colored tag indicating live-data
// freshness. The TUI is "live" when an event has arrived in the last
// 30s; otherwise it's "idle" (but we're still polling via fallback).
func (m model) renderLiveBadge() string {
	if m.lastLive.IsZero() {
		return m.theme.HeaderMuted.Render("○ connecting")
	}
	age := time.Since(m.lastLive)
	if age < 30*time.Second {
		label := "● live"
		if m.liveCount > 0 {
			label = fmt.Sprintf("● live · %d cmds", m.liveCount)
		}
		return m.theme.Positive.Render(label)
	}
	return m.theme.HeaderMuted.Render(fmt.Sprintf("○ idle %ds", int(age.Seconds())))
}

// formatLiveRecordToast is the one-line summary surfaced when a new
// command gets recorded. Shows tokens saved + the command head so the
// user can see their own action reflected in real time.
func formatLiveRecordToast(rec *tracking.CommandRecord) string {
	if rec == nil {
		return "new command recorded"
	}
	cmd := rec.Command
	if len(cmd) > 40 {
		cmd = cmd[:39] + "…"
	}
	return fmt.Sprintf("+%s saved · %s", formatInt(int64(rec.SavedTokens)), cmd)
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
