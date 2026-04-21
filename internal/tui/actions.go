package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/hooks"
)

// Action is a single registered operation the TUI can perform, whether
// triggered by the palette, a keybinding, or another message handler.
//
// An Action can be read-only (refresh, export, jump) or mutating (toggle
// hook, vacuum DB, run compress). Mutating actions MUST set Confirm=true
// so the root model shows a confirm-modal before invoking Run. Phase 1
// only registers read-only actions; Phase 2 adds the management set.
type Action struct {
	// ID is the stable identifier used by paletteExecMsg.ActionID. Use a
	// dotted namespace: "section.jump", "view.refresh", "hooks.toggle".
	ID string

	// Title is what the palette shows (concise, human-readable).
	Title string

	// Description is a one-line hint rendered under the title in the
	// palette. Keep under 60 chars so even narrow terminals fit it.
	Description string

	// Category groups actions in the palette ("Navigation", "View",
	// "Data"). Categories render as section headers.
	Category string

	// Confirm requests a yes/no modal before Run. Leave false for
	// read-only actions.
	Confirm bool

	// Run executes the action. It returns an arbitrary Result (usually
	// nil) and an error. The registry will wrap the return in an
	// actionResultMsg so sections can react.
	//
	// Args is the free-text tail the user typed after the action ID in
	// the palette (":section.jump 5" → "5"). Parse it as needed.
	Run func(ctx context.Context, args string) (any, error)
}

// ActionRegistry is the central store of registered actions. Safe for
// concurrent reads during TUI rendering.
type ActionRegistry struct {
	mu      sync.RWMutex
	byID    map[string]Action
	ordered []string
}

// NewActionRegistry returns an empty registry. Use DefaultActionRegistry
// for the TUI's canonical set.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{byID: map[string]Action{}}
}

// Register adds an action. Re-registering the same ID replaces the prior
// entry — callers are responsible for ensuring IDs are unique.
func (r *ActionRegistry) Register(a Action) {
	if a.ID == "" || a.Run == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[a.ID]; !exists {
		r.ordered = append(r.ordered, a.ID)
	}
	r.byID[a.ID] = a
}

// Get returns the action by ID; ok is false if not registered.
func (r *ActionRegistry) Get(id string) (Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byID[id]
	return a, ok
}

// All returns every registered action sorted by category then title.
// Used by the palette to render its list.
func (r *ActionRegistry) All() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Action, 0, len(r.byID))
	for _, id := range r.ordered {
		out = append(out, r.byID[id])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// runActionCmd produces a tea.Cmd that executes the action in a goroutine
// and dispatches an actionResultMsg. The ctx is the model-owned context
// so a Quit cancels in-flight actions cleanly.
func runActionCmd(ctx context.Context, reg *ActionRegistry, id, args string) tea.Cmd {
	return func() tea.Msg {
		a, ok := reg.Get(id)
		if !ok {
			return actionResultMsg{ActionID: id, Err: fmt.Errorf("unknown action: %s", id)}
		}
		// Respect context cancellation before we kick off work so a rapid
		// Quit doesn't start an action we're about to tear down.
		if err := ctx.Err(); err != nil {
			return actionResultMsg{ActionID: id, Err: err}
		}
		result, err := a.Run(ctx, args)
		return actionResultMsg{ActionID: id, Result: result, Err: err}
	}
}

// DefaultActionRegistry returns the registry pre-populated with Phase 1's
// read-only actions. The closures capture the model's context and key
// dependencies via the provided ActionDeps. Phase 2 extends this with
// management actions (hooks toggle, DB vacuum, run compress).
type ActionDeps struct {
	RequestRefresh    func() tea.Cmd
	RequestJump       func(sectionIndex int) tea.Cmd
	RequestToast      func(kind ToastKind, text string) tea.Cmd
	RequestTheme      func(name ThemeName) tea.Cmd
	RequestThemeCycle func() tea.Cmd
	// ClearLogRing wipes the in-memory slog ring. Exposed as an action
	// dependency (not a direct ctx.Logs.Clear call) so the action can
	// be dispatched from the palette without needing SectionContext.
	ClearLogRing func() tea.Cmd
	SectionCount int
}

func DefaultActionRegistry(deps ActionDeps) *ActionRegistry {
	r := NewActionRegistry()

	r.Register(Action{
		ID:          "view.refresh",
		Title:       "Refresh",
		Description: "Reload the workspace snapshot now",
		Category:    "View",
		Run: func(context.Context, string) (any, error) {
			if deps.RequestRefresh != nil {
				_ = deps.RequestRefresh()
			}
			return nil, nil
		},
	})

	r.Register(Action{
		ID:          "section.jump",
		Title:       "Jump to section",
		Description: "Usage: :section.jump <n>  (1–" + fmt.Sprint(deps.SectionCount) + ")",
		Category:    "Navigation",
		Run: func(_ context.Context, args string) (any, error) {
			args = strings.TrimSpace(args)
			if args == "" {
				return nil, fmt.Errorf("missing section number")
			}
			var n int
			if _, err := fmt.Sscanf(args, "%d", &n); err != nil || n < 1 || n > deps.SectionCount {
				return nil, fmt.Errorf("invalid section: %s", args)
			}
			if deps.RequestJump != nil {
				_ = deps.RequestJump(n - 1)
			}
			return nil, nil
		},
	})

	r.Register(Action{
		ID:          "toast.info",
		Title:       "Show info toast",
		Description: "Diagnostic: verify the toast layer renders",
		Category:    "Debug",
		Run: func(_ context.Context, args string) (any, error) {
			msg := strings.TrimSpace(args)
			if msg == "" {
				msg = "hello from the action registry"
			}
			if deps.RequestToast != nil {
				_ = deps.RequestToast(ToastInfo, msg)
			}
			return nil, nil
		},
	})

	r.Register(Action{
		ID:          "theme.set",
		Title:       "Set theme",
		Description: "Usage: :theme.set dark|light|high-contrast|colorblind",
		Category:    "View",
		Run: func(_ context.Context, args string) (any, error) {
			name := ThemeName(strings.TrimSpace(args))
			if name == "" {
				return nil, fmt.Errorf("theme name required")
			}
			valid := false
			for _, t := range AvailableThemes {
				if t == name {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("unknown theme: %s", name)
			}
			if deps.RequestTheme != nil {
				_ = deps.RequestTheme(name)
			}
			return name, nil
		},
	})

	r.Register(Action{
		ID:          "theme.cycle",
		Title:       "Cycle theme",
		Description: "Advance to the next bundled theme",
		Category:    "View",
		Run: func(_ context.Context, _ string) (any, error) {
			// The model resolves "current" at message delivery time, so
			// repeated cycles actually advance (see themeCycleMsg
			// handler in app.go Update).
			if deps.RequestThemeCycle == nil {
				return nil, fmt.Errorf("theme cycling unavailable")
			}
			_ = deps.RequestThemeCycle()
			return nil, nil
		},
	})

	// logs.clear is destructive enough to warrant a confirm modal —
	// the ring holds diagnostic context that users might still need
	// for an ongoing debugging session. Confirm=true routes it
	// through ConfirmOverlay before Run fires.
	r.Register(Action{
		ID:          "logs.clear",
		Title:       "Clear log ring",
		Description: "Discard every in-memory log event. This cannot be undone.",
		Category:    "System",
		Confirm:     true,
		Run: func(_ context.Context, _ string) (any, error) {
			if deps.ClearLogRing != nil {
				_ = deps.ClearLogRing()
			}
			if deps.RequestToast != nil {
				_ = deps.RequestToast(ToastSuccess, "log ring cleared")
			}
			return nil, nil
		},
	})

	// hooks.diagnose inspects the hook flag file + runtime env to
	// explain why attribution might be failing. Read-only action:
	// it just reports, doesn't mutate. Output goes to a toast chain
	// so users can review it in context without leaving the Home view.
	r.Register(Action{
		ID:          "hooks.diagnose",
		Title:       "Diagnose hook",
		Description: "Check why commands might be missing attribution",
		Category:    "System",
		Run: func(_ context.Context, _ string) (any, error) {
			report := diagnoseHooks()
			kind := ToastSuccess
			if !report.OK {
				kind = ToastWarning
			}
			if deps.RequestToast != nil {
				// Split multi-line reports across several toasts so
				// each line stays readable in the stack.
				for _, line := range strings.Split(report.Summary, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					_ = deps.RequestToast(kind, line)
				}
			}
			return report, nil
		},
	})

	// First mutating action: toggle the tok hook activation flag. The
	// flag file is the same one `tok hook mode` and the shell
	// statusline read, so flipping it here changes whether new shell
	// sessions pick tok up. No confirm modal yet (see Phase 3 plan);
	// the toast reports the post-toggle state so the user can undo
	// immediately if the effect surprises them.
	r.Register(Action{
		ID:          "hooks.toggle",
		Title:       "Toggle tok hook",
		Description: "Activate or deactivate the global tok shell hook",
		Category:    "System",
		Run: func(_ context.Context, args string) (any, error) {
			_ = args
			var newState string
			if hooks.IsActive() {
				if err := hooks.Deactivate(); err != nil {
					return nil, err
				}
				newState = "deactivated"
			} else {
				mode := hooks.ResolveDefaultMode()
				if err := hooks.Activate(mode); err != nil {
					return nil, err
				}
				newState = "activated (" + mode + ")"
			}
			if deps.RequestToast != nil {
				_ = deps.RequestToast(ToastSuccess, "tok hook "+newState)
			}
			return newState, nil
		},
	})

	return r
}
