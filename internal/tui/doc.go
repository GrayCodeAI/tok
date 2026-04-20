// Package tui implements the tok terminal dashboard — a long-lived Bubble
// Tea app launched via `tok tui` (wired from internal/commands/core/tui.go).
//
// # Architecture
//
// The shell is phased: the root model in app.go owns a navigation bar, a
// themed layout (sidebar / main / optional insights pane), a KeyMap-driven
// dispatcher (keys.go), and a snapshotLoader (loader.go) that fetches a
// WorkspaceDashboardSnapshot from the SQLite trackers on a refresh tick.
//
// # Invariants
//
// These invariants matter because the TUI is a long-lived process in a
// binary whose other subcommands are short-lived:
//
//  1. internal/state.Global() is populated exactly once per process in
//     root.go's PersistentPreRunE before any subcommand runs. The TUI does
//     NOT invoke cobra subcommand handlers, so global flags are effectively
//     immutable for the TUI's lifetime. Do not call shared.SetFlags from
//     inside a TUI message handler.
//
//  2. The snapshotLoader owns all DB handles. On tea.Quit, shutdownCmd
//     cancels the model's context and calls loader.Close() before emitting
//     quitMsg (which then issues tea.Quit). Do not open a Tracker or
//     SessionManager directly from a section view; go through the loader.
//
//  3. Compressor / filter code paths invoked from future management actions
//     MUST route output through internal/output.Global(). Raw fmt.Print to
//     os.Stdout will corrupt the alt-screen frame.
//
// # Adding a section
//
// Section renderers are being extracted to a SectionRenderer interface in
// Phase 1. Until that lands, view_home.go is the canonical reference and
// view_placeholder.go stands in for the rest.
package tui
