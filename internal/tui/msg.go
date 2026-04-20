package tui

import "time"

// This file is the single source of truth for the TUI's typed message
// vocabulary. Every Update() switch across the package references these
// types rather than declaring ad-hoc struct literals, so a new section or
// overlay can be added without touching unrelated message handlers.
//
// Keep these messages *value types* (or small pointers to immutable data).
// Bubble Tea fans them out across the program, and shared mutable state in
// a message is a race waiting to happen.

// --- Filtering ------------------------------------------------------------

// FilterField enumerates the known filter dimensions. Centralized so the
// palette, filter modal, and each section's Update dispatch agree on what
// "agent" or "provider" means.
type FilterField string

const (
	FilterDays     FilterField = "days"
	FilterProject  FilterField = "project"
	FilterAgent    FilterField = "agent"
	FilterProvider FilterField = "provider"
	FilterModel    FilterField = "model"
	FilterSession  FilterField = "session"
)

// filterChangedMsg is dispatched when the user edits a filter. The root
// model applies it to Options, then kicks a reload.
type filterChangedMsg struct {
	Field FilterField
	Value string // empty string clears the filter
}

// --- Drill-down / navigation ---------------------------------------------

// drillDownMsg is emitted by a section when the user presses Enter on a
// row. The section itself interprets Key + Payload; the root model only
// routes the message back to that section.
type drillDownMsg struct {
	Section int    // navIndex of the target section
	Key     string // row identifier (opaque to the router)
	Payload any    // optional typed payload the section set
}

// drillBackMsg pops one level of drill-down on the current section.
type drillBackMsg struct{}

// --- Command palette ------------------------------------------------------

type paletteOpenMsg struct{}
type paletteCloseMsg struct{}

// paletteExecMsg fires when the user selects an entry in the palette. The
// entry is always an actionID plus optional free-text args (e.g. ":jump 5"
// → action="section.jump", args="5").
type paletteExecMsg struct {
	ActionID string
	Args     string
}

// --- In-pane search -------------------------------------------------------

type searchOpenMsg struct{}
type searchCloseMsg struct{}

// searchMsg carries the live query string as the user types. An empty
// Query means "show all rows again".
type searchMsg struct {
	Query string
}

// searchNextMsg / searchPrevMsg jump between matches without reopening.
type searchNextMsg struct{}
type searchPrevMsg struct{}

// --- Actions --------------------------------------------------------------

// actionRequestMsg asks the action registry to run an action. The registry
// responds with actionResultMsg on completion.
type actionRequestMsg struct {
	ActionID string
	Args     string
}

// actionResultMsg reports the outcome of an action run. The section that
// triggered the action is responsible for rendering any follow-up toast.
type actionResultMsg struct {
	ActionID string
	Result   any
	Err      error
}

// --- Toasts ---------------------------------------------------------------

// ToastKind drives the toast's accent color and icon. Keep the set small.
type ToastKind int

const (
	ToastInfo ToastKind = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// toastAddMsg enqueues a toast. TTL<=0 means use the package default.
type toastAddMsg struct {
	Kind ToastKind
	Text string
	TTL  time.Duration
}

// toastExpireMsg is dispatched from tea.Tick after a toast's TTL elapses.
type toastExpireMsg struct {
	ID uint64
}

// --- Export ---------------------------------------------------------------

// ExportFormat is the on-disk format the user asked for.
type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportCSV  ExportFormat = "csv"
	ExportMD   ExportFormat = "md"
)

type exportMsg struct {
	Format  ExportFormat
	Section int // -1 means "current section"
}

// --- Theme ----------------------------------------------------------------

// themeChangedMsg swaps the active theme on the root model. Emitted by
// the theme.set action and consumed by Update to rebuild the model's
// theme field without tearing down the tea.Program.
type themeChangedMsg struct {
	Name ThemeName
}

// themeCycleMsg is emitted by the theme.cycle action. The root model
// resolves "next" against its *current* theme (not a snapshot captured
// when the registry was built), so repeated cycles actually advance.
type themeCycleMsg struct{}
