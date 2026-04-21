package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap is the registry of every global TUI keybinding. One source of truth
// for (a) matching in Update(), (b) rendering the help overlay, (c) driving
// the command palette, and (d) documenting keys in generated docs.
//
// Section-local keymaps live next to their view (see SectionKeyMap once the
// SectionRenderer interface lands in Phase 1). Sections embed a reference to
// this global map so their help overlays show both global and local bindings
// without duplication.
type KeyMap struct {
	// Navigation
	NextSection key.Binding
	PrevSection key.Binding
	JumpSection key.Binding // numeric 1-9; matched separately, documented here

	// History navigation (back/forward across sections)
	HistoryBack    key.Binding
	HistoryForward key.Binding

	// Cursor movement inside a section (Phase 1 sections will honor these)
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Top     key.Binding
	Bottom  key.Binding
	PageUp  key.Binding
	PageDn  key.Binding

	// Actions
	Refresh key.Binding
	Enter   key.Binding
	Back    key.Binding
	Yank    key.Binding
	Export  key.Binding

	// Overlays (Phase 1)
	Palette key.Binding // ":"
	Search  key.Binding // "/"
	Filter  key.Binding // "f"
	Help    key.Binding // "?"

	// App lifecycle
	Quit key.Binding
	Esc  key.Binding
}

// DefaultKeyMap returns the canonical key bindings. Copy and mutate only if
// a user-facing rebind feature ever lands; for now this is the only source.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextSection: key.NewBinding(
			key.WithKeys("tab", "right", "l"),
			key.WithHelp("tab/→/l", "next section"),
		),
		PrevSection: key.NewBinding(
			key.WithKeys("shift+tab", "left", "h"),
			key.WithHelp("shift+tab/←/h", "prev section"),
		),
		HistoryBack: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "history back"),
		),
		HistoryForward: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "history forward"),
		),
		JumpSection: key.NewBinding(
			// 1–9 jump to the matching section. 0 maps to section 10.
			// Sections 11–12 are still reachable via tab/shift+tab or
			// the palette (`:section.jump 11`) — we don't bind a
			// second digit because bubbletea's key system has no
			// built-in prefix support and a bespoke state machine for
			// two rare shortcuts would be more code than value.
			key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9", "0"),
			key.WithHelp("1-9,0", "jump to section"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "cursor up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "cursor down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "cursor left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "cursor right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/home", "jump to top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/end", "jump to bottom"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("pgup", "page up"),
		),
		PageDn: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+f"),
			key.WithHelp("pgdn", "page down"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "drill into row"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("⌫", "back"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank row"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export view"),
		),
		Palette: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command palette"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter rows"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close overlay"),
		),
	}
}

// ShortHelp returns the bindings shown in the always-visible footer strip.
// Keep this to ≤6 bindings — footer width is precious.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.JumpSection,
		k.NextSection,
		k.Refresh,
		k.Palette,
		k.Help,
		k.Quit,
	}
}

// FullHelp groups bindings into columns for the "?" overlay.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.JumpSection, k.NextSection, k.PrevSection, k.HistoryBack, k.HistoryForward, k.Refresh},
		{k.Up, k.Down, k.Top, k.Bottom, k.PageUp, k.PageDn},
		{k.Enter, k.Back, k.Yank, k.Export},
		{k.Palette, k.Search, k.Help, k.Esc, k.Quit},
	}
}
