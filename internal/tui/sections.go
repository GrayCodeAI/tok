package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// SectionContext is the read-only bundle the root model hands to a
// section every time Init, Update, or View is called. Keeping the
// context separate from the section's own state is what lets the root
// own layout (width/height), theming, keybindings, and data fetching,
// while sections own purely section-local state (cursor, sort, filter).
//
// Sections must NOT mutate the context. If a section needs to signal
// the root — e.g. request a refresh, emit a toast, trigger an action —
// it does so by returning a tea.Cmd that produces the appropriate
// message (see internal/tui/msg.go).
type SectionContext struct {
	Theme   theme
	Keys    KeyMap
	Data    *tracking.WorkspaceDashboardSnapshot
	Opts    Options
	Width   int
	Height  int
	Compact bool
	Focused bool // true when this section is currently visible
	// Logs is the in-memory slog ring populated by the TUI root model.
	// Nil outside the TUI (e.g. in unit tests that don't need the
	// Logs section). Pass through to sections that need it; other
	// sections ignore it.
	Logs *ringHandler
	// Env reports terminal capabilities detected at startup. Sections
	// that render Braille/block glyphs should degrade to ASCII when
	// Env.UTF8 is false. Zero-value (all false) is safe but produces
	// the most conservative glyphs.
	Env Environment
}

// SectionRenderer is implemented by each screen in the TUI. The root
// model holds one instance per section in the sidebar, and the
// currently-focused section's Update and View are driven every frame.
//
// Return contract:
//   - Update returns the (possibly mutated) section so the root can
//     store it back. Sections ARE NOT pointer-free — a section may
//     carry table state, textinputs, etc., but they must be copyable
//     (embed bubbles components directly, don't hide them behind a
//     pointer that would surprise the caller on reassignment).
//   - KeyBindings returns section-local bindings that are:
//       1. merged into the ? overlay
//       2. documented in auto-generated help
//     Global bindings (nav, refresh, palette, quit) must NOT be
//     redeclared here.
type SectionRenderer interface {
	Name() string
	Short() string
	Init(ctx SectionContext) tea.Cmd
	Update(ctx SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd)
	View(ctx SectionContext) string
	KeyBindings() []key.Binding
}

// defaultSections returns the canonical ordered list of SectionRenderers
// the TUI ships with. Phase 1 registers Home as the only fully-rendered
// section; every other entry is a placeholderSection that documents
// what's coming. Phase 2 replaces them one-by-one with real renderers.
func defaultSections() []SectionRenderer {
	return []SectionRenderer{
		newHomeSection(),
		newTodaySection(),
		newTrendsSection(),
		newProvidersSection(),
		newModelsSection(),
		newAgentsSection(),
		newSessionsSection(),
		newCommandsSection(),
		newPipelineSection(),
		newRewardsSection(),
		newLogsSection(),
		newConfigSection(),
	}
}

// --- Placeholder section --------------------------------------------------

// placeholderSection is the default renderer for sections that haven't
// been implemented yet. It keeps the sidebar entry live, explains the
// phased rollout, and participates in the Update loop so navigation
// doesn't break.
type placeholderSection struct {
	title string
	short string
}

func newPlaceholderSection(title, short string) *placeholderSection {
	return &placeholderSection{title: title, short: short}
}

func (s *placeholderSection) Name() string                         { return s.title }
func (s *placeholderSection) Short() string                        { return s.short }
func (s *placeholderSection) Init(SectionContext) tea.Cmd          { return nil }
func (s *placeholderSection) KeyBindings() []key.Binding           { return nil }
func (s *placeholderSection) Update(_ SectionContext, _ tea.Msg) (SectionRenderer, tea.Cmd) {
	return s, nil
}

func (s *placeholderSection) View(ctx SectionContext) string {
	lines := []string{
		ctx.Theme.Title.Render(s.title),
		ctx.Theme.Muted.Render(s.short),
		"",
		ctx.Theme.Warning.Render("Planned next in the phased build."),
		"",
		"This screen is intentionally held behind the shell/data foundation.",
		"The TUI is being built one slice at a time to avoid the old layout and architecture failures.",
	}
	// Pick a deterministic accent so each placeholder has a distinct
	// border color — matches the Phase 0 look and feel.
	base := ctx.Theme.Card
	if len(ctx.Theme.AccentColors) > 0 {
		color := ctx.Theme.AccentColors[len(s.title)%len(ctx.Theme.AccentColors)]
		base = base.BorderForeground(color)
	}
	return setWidth(base, ctx.Width).Render(joinLines(lines))
}

// joinLines is a tiny helper kept local so section implementations don't
// have to depend on strings.Join directly; it sidesteps the temptation
// to import "strings" in every new section file.
func joinLines(ls []string) string {
	out := ""
	for i, l := range ls {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}
