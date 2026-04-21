package tui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/hooks"
)

// configSection surfaces the TUI's view of configuration, hook state,
// and data-quality health. This is the first section with a mutating
// keybinding (`t` toggles the tok hook). The action runs through the
// shared registry so the palette can invoke the same thing via
// `:hooks.toggle`.
type configSection struct{}

func newConfigSection() *configSection { return &configSection{} }

func (s *configSection) Name() string                { return "Config" }
func (s *configSection) Short() string               { return "Health" }
func (s *configSection) Init(SectionContext) tea.Cmd { return nil }
func (s *configSection) IsScrollable() bool          { return true }

func (s *configSection) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "toggle tok hook")),
	}
}

func (s *configSection) Update(_ SectionContext, msg tea.Msg) (SectionRenderer, tea.Cmd) {
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "t":
			return s, func() tea.Msg {
				return actionRequestMsg{ActionID: "hooks.toggle"}
			}
		}
	}
	return s, nil
}

func (s *configSection) View(ctx SectionContext) string {
	th := ctx.Theme
	width := ctx.Width

	// Hook status card — the main interactive element on this screen.
	active := hooks.IsActive()
	hookTitle := th.PanelTitle.Render("Tok hook")
	hookStatus := th.Danger.Render("● inactive")
	hookMode := "—"
	if active {
		hookStatus = th.Positive.Render("● active")
		hookMode = hooks.GetMode()
	}
	hookPanel := setWidth(panelStyle(th, 0), width).Render(strings.Join([]string{
		hookTitle,
		renderHealthLine("Status", hookStatus),
		renderHealthLine("Mode", hookMode),
		renderHealthLine("Flag path", hooks.GetFlagPath()),
		renderHealthLine("Default mode (if activated)", hooks.ResolveDefaultMode()),
		"",
		th.CardMeta.Render("press t to toggle · :hooks.toggle in the palette"),
	}, "\n"))

	// Config paths panel — helpful when users need to inspect or edit
	// the files tok reads. We don't parse the config; just show paths.
	configPanel := setWidth(panelStyle(th, 2), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Paths"),
		renderHealthLine("Tracking database", shared.GetDatabasePath()),
		renderHealthLine("Refresh interval", ctx.Opts.RefreshInterval.String()),
		renderHealthLine("Active window", fmt.Sprintf("%d days", ctx.Opts.Days)),
		renderHealthLine("Go runtime", fmt.Sprintf("%s on %s/%s",
			runtime.Version(), runtime.GOOS, runtime.GOARCH)),
	}, "\n"))

	// Data quality panel — the DataQuality block in the snapshot is the
	// same one surfaced on Home, but here we put it front-and-center so
	// users can diagnose why their numbers look off.
	dqPanel := setWidth(panelStyle(th, 4), width).Render(strings.Join(dataQualityLines(ctx), "\n"))

	// Filter panel — shows which dimensions the TUI is scoped to. If
	// everything is empty, users know they're looking at workspace-wide
	// numbers.
	filterPanel := setWidth(panelStyle(th, 6), width).Render(strings.Join([]string{
		th.PanelTitle.Render("Active filters"),
		renderHealthLine("Project", fallback(ctx.Opts.ProjectPath, "(all)")),
		renderHealthLine("Agent", fallback(ctx.Opts.AgentName, "(all)")),
		renderHealthLine("Provider", fallback(ctx.Opts.Provider, "(all)")),
		renderHealthLine("Model", fallback(ctx.Opts.ModelName, "(all)")),
		renderHealthLine("Session", fallback(ctx.Opts.SessionID, "(all)")),
	}, "\n"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		th.Title.Render("Config"),
		th.Subtitle.Render("Runtime configuration, hook state, and data health."),
		"",
		hookPanel,
		"",
		configPanel,
		"",
		dqPanel,
		"",
		filterPanel,
	)
}

func dataQualityLines(ctx SectionContext) []string {
	th := ctx.Theme
	lines := []string{th.PanelTitle.Render("Data quality")}
	if ctx.Data == nil {
		lines = append(lines, th.Muted.Render("no dashboard data yet"))
		return lines
	}
	q := ctx.Data.DataQuality
	lines = append(lines,
		renderHealthLine("Pricing coverage",
			fmt.Sprintf("%.1f%% explicit · %d commands fell back",
				q.PricingCoverage.CoveragePct(), q.PricingCoverage.FallbackPricingCommands)),
		renderHealthLine("Attribution gaps",
			fmt.Sprintf("agent=%d provider=%d model=%d session=%d",
				q.CommandsMissingAgent, q.CommandsMissingProvider,
				q.CommandsMissingModel, q.CommandsMissingSession)),
		renderHealthLine("Parse failures", fmt.Sprintf("%d in active window", q.ParseFailures)),
		renderHealthLine("Total commands", fmt.Sprintf("%d", q.TotalCommands)),
	)
	return lines
}
