package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// filterOverlayMsg signals the filter overlay closed with/without a query.
type filterOverlayMsg struct {
	Query string // empty if cancelled
}

// FilterOverlay provides real-time fuzzy filtering for table rows.
type FilterOverlay struct {
	input textinput.Model
	open  bool
}

// NewFilterOverlay creates a new filter input overlay.
func NewFilterOverlay() *FilterOverlay {
	ti := textinput.New()
	ti.Placeholder = "filter rows..."
	ti.CharLimit = 100
	ti.Width = 40
	return &FilterOverlay{input: ti}
}

// IsOpen returns true if the overlay is currently active.
func (f *FilterOverlay) IsOpen() bool {
	return f.open
}

// Open activates the filter overlay with focus.
func (f *FilterOverlay) Open() tea.Cmd {
	f.open = true
	f.input.SetValue("")
	f.input.Focus()
	return textinput.Blink
}

// Close deactivates the filter overlay.
func (f *FilterOverlay) Close() {
	f.open = false
	f.input.Blur()
}

// Value returns the current filter query.
func (f *FilterOverlay) Value() string {
	return f.input.Value()
}

// SetValue sets the filter query programmatically.
func (f *FilterOverlay) SetValue(v string) {
	f.input.SetValue(v)
}

// Update handles key events while the overlay is open.
// Returns a command that emits filterOverlayMsg on enter/esc.
func (f *FilterOverlay) Update(msg tea.Msg) tea.Cmd {
	if !f.open {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			query := f.input.Value()
			f.Close()
			return func() tea.Msg { return filterOverlayMsg{Query: query} }
		case tea.KeyEsc:
			f.Close()
			return func() tea.Msg { return filterOverlayMsg{Query: ""} }
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd
}

// View renders the filter overlay as a floating input bar.
func (f *FilterOverlay) View(th theme, width int) string {
	if !f.open {
		return ""
	}
	// Center the input with some padding
	inputWidth := min(50, width-4)
	f.input.Width = inputWidth

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Focus.GetForeground()).
		Padding(0, 1).
		Width(inputWidth + 4)

	return style.Render(
		th.PanelTitle.Render("Filter") + "\n" +
			f.input.View() + "\n" +
			th.CardMeta.Render("enter: apply  ·  esc: clear"),
	)
}

// rowMatchesFilter checks if row matches the filter query using substring matching.
// This is a simple case-insensitive check across all cells.
func rowMatchesFilter(row Row, query string) bool {
	if query == "" {
		return true
	}
	needle := strings.ToLower(query)
	for _, cell := range row.Cells {
		if strings.Contains(strings.ToLower(cell), needle) {
			return true
		}
	}
	return false
}

// filterRows returns only rows matching the filter query.
func filterRows(rows []Row, query string) []Row {
	if query == "" {
		return rows
	}

	var filtered []Row
	for _, row := range rows {
		if rowMatchesFilter(row, query) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
