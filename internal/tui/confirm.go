package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmOverlay is a y/n modal shown before Confirm=true actions run.
// Opening captures input until the user chooses yes or no; the root
// model's Update forwards KeyMsg here instead of the normal handler
// path while open.
//
// Emits confirmAcceptedMsg or confirmRejectedMsg on resolution. The
// pending Action + args are carried on the overlay so the accept path
// can dispatch the original actionRequestMsg without re-racing through
// the palette.
type ConfirmOverlay struct {
	open    bool
	title   string
	prompt  string
	action  Action
	args    string
	default_ bool // which button is highlighted by default: true=yes, false=no
}

// NewConfirmOverlay returns a closed confirm overlay. Call Open() to
// arm it with a specific action + prompt text.
func NewConfirmOverlay() *ConfirmOverlay {
	return &ConfirmOverlay{}
}

// Open arms the overlay for the given Action. The prompt text is the
// action's description (or a generated fallback if description is
// empty). No command is returned — the modal has no focused child
// widget to tick.
func (c *ConfirmOverlay) Open(a Action, args string) {
	c.open = true
	c.action = a
	c.args = args
	c.title = a.Title
	c.prompt = a.Description
	if strings.TrimSpace(c.prompt) == "" {
		c.prompt = "Run action: " + a.ID
	}
	c.default_ = false // default "no" for safety on destructive actions
}

// Close forces the overlay closed without resolving the pending action.
func (c *ConfirmOverlay) Close() {
	c.open = false
	c.action = Action{}
	c.args = ""
}

// IsOpen reports whether the overlay is capturing input.
func (c *ConfirmOverlay) IsOpen() bool { return c.open }

// Update handles keystrokes while open. y / Enter accepts, n / Esc
// rejects. The accepted branch returns an actionRequestMsg so the root
// model funnels the action through the same runActionCmd path it would
// have used without the modal.
func (c *ConfirmOverlay) Update(msg tea.Msg) tea.Cmd {
	if !c.open {
		return nil
	}
	m, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch m.String() {
	case "y", "Y":
		id := c.action.ID
		args := c.args
		c.Close()
		return func() tea.Msg { return actionRequestMsg{ActionID: id, Args: args} }
	case "enter":
		// Honor whichever button is highlighted. We default to no, so
		// Enter is a safe way to cancel destructive ops the user didn't
		// intend. Users can still type `y` to accept.
		if c.default_ {
			id := c.action.ID
			args := c.args
			c.Close()
			return func() tea.Msg { return actionRequestMsg{ActionID: id, Args: args} }
		}
		c.Close()
		return nil
	case "n", "N", "esc":
		c.Close()
		return nil
	case "left", "right", "tab":
		c.default_ = !c.default_
	}
	return nil
}

// View renders the modal. Caller composites it over the main frame.
func (c *ConfirmOverlay) View(th theme, width int) string {
	if !c.open {
		return ""
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Warning.GetForeground()).
		Background(th.Panel.GetBackground()).
		Padding(1, 2).
		Width(minInt(72, width-4))

	yesLabel := " yes "
	noLabel := " no "
	yesStyle := lipgloss.NewStyle().Padding(0, 1).Background(th.Muted.GetForeground())
	noStyle := lipgloss.NewStyle().Padding(0, 1).Background(th.Muted.GetForeground())
	if c.default_ {
		yesStyle = yesStyle.Background(th.Positive.GetForeground()).Foreground(lipgloss.Color("#000000")).Bold(true)
	} else {
		noStyle = noStyle.Background(th.Danger.GetForeground()).Foreground(lipgloss.Color("#000000")).Bold(true)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top,
		yesStyle.Render(yesLabel), "   ", noStyle.Render(noLabel))

	lines := []string{
		th.Warning.Render("Confirm action"),
		"",
		th.SectionTitle.Render(c.title),
		th.Muted.Render(c.prompt),
		"",
		buttons,
		"",
		th.CardMeta.Render("y accept · n/esc cancel · tab to switch · enter honors default (no)"),
	}
	return box.Render(strings.Join(lines, "\n"))
}
