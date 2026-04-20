package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchOverlay is an in-pane search field toggled with "/". When open
// it captures every keystroke except Esc / Enter. The root model blocks
// its key dispatch and section Update delivery while the overlay is
// focused so typing in the field doesn't trigger nav or palette keys.
type SearchOverlay struct {
	input   textinput.Model
	open    bool
	visible bool
}

// NewSearchOverlay returns an initialized search overlay. The prompt is
// a single "/" styled by the footer-key color.
func NewSearchOverlay() *SearchOverlay {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.Placeholder = "filter rows"
	ti.CharLimit = 64
	return &SearchOverlay{input: ti}
}

// Open focuses the overlay. Returns the textinput's focus command so the
// cursor starts blinking immediately.
func (s *SearchOverlay) Open() tea.Cmd {
	s.open = true
	s.visible = true
	return s.input.Focus()
}

// Close hides and clears the overlay.
func (s *SearchOverlay) Close() {
	s.open = false
	s.visible = false
	s.input.Blur()
	s.input.Reset()
}

// IsOpen reports whether the overlay is currently accepting keystrokes.
func (s *SearchOverlay) IsOpen() bool { return s.open }

// Query returns the current query text.
func (s *SearchOverlay) Query() string { return s.input.Value() }

// Update feeds the message to the textinput and returns a searchMsg
// whenever the query changes. The caller (root model or a section)
// is responsible for routing that message to the table or similar
// consumer.
func (s *SearchOverlay) Update(msg tea.Msg) tea.Cmd {
	if !s.open {
		return nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			s.Close()
			return func() tea.Msg { return searchCloseMsg{} }
		case "enter":
			// Keep the query applied but release focus so cursor keys
			// can navigate the filtered results.
			s.open = false
			s.input.Blur()
			return func() tea.Msg { return searchMsg{Query: s.input.Value()} }
		}
	}
	prev := s.input.Value()
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	if s.input.Value() != prev {
		value := s.input.Value()
		cmd = tea.Batch(cmd, func() tea.Msg { return searchMsg{Query: value} })
	}
	return cmd
}

// View renders the overlay as a single line. Sections should render it
// near their top so users see what's being filtered.
func (s *SearchOverlay) View(th theme, width int) string {
	if !s.visible {
		return ""
	}
	style := lipgloss.NewStyle().
		Foreground(th.Focus.GetForeground()).
		Background(th.Panel.GetBackground()).
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(th.Focus.GetForeground())
	return setWidth(style, width).Render(s.input.View())
}
