package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Palette is the command palette overlay toggled with ":". It fuzzy-
// matches against registered actions + section names and dispatches
// paletteExecMsg on Enter.
//
// The palette draws itself as a modal card over the main view; while
// open the root model forwards every KeyMsg here instead of the keymap
// registry. Esc closes without executing.
type Palette struct {
	input    textinput.Model
	entries  []paletteEntry
	matches  []paletteEntry
	cursor   int
	open     bool
	registry *ActionRegistry
	sections []SectionRenderer
}

type paletteEntry struct {
	ID          string // the actionID (or "section:N" pseudo-id)
	Title       string
	Description string
	Category    string
	Score       int    // populated during fuzzy match; higher = better
	Args        string // the user-typed tail (e.g. ":section.jump 5" → "5")
}

// NewPalette returns a closed palette seeded from the action registry
// and section list. Sections are synthesized as pseudo-entries so users
// can type "home" or "sessions" without knowing the action namespace.
func NewPalette(reg *ActionRegistry, sections []SectionRenderer) *Palette {
	ti := textinput.New()
	ti.Prompt = ":"
	ti.Placeholder = "command or section"
	ti.CharLimit = 128

	p := &Palette{
		input:    ti,
		registry: reg,
		sections: sections,
	}
	p.rebuildEntries()
	return p
}

func (p *Palette) rebuildEntries() {
	entries := make([]paletteEntry, 0, 32)
	if p.registry != nil {
		for _, a := range p.registry.All() {
			entries = append(entries, paletteEntry{
				ID:          a.ID,
				Title:       a.Title,
				Description: a.Description,
				Category:    a.Category,
			})
		}
	}
	// Synthesize one entry per section so "Home", "Trends" etc. are
	// reachable by name without memorizing section.jump.
	for i, s := range p.sections {
		entries = append(entries, paletteEntry{
			ID:          "section.jump",
			Title:       "Go to " + s.Name(),
			Description: s.Short(),
			Category:    "Navigation",
			Args:        itoaPalette(i + 1),
		})
	}
	p.entries = entries
	p.applyFilter("")
}

// itoaPalette avoids strconv just to keep this file dependency-light.
func itoaPalette(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// Open focuses the palette input.
func (p *Palette) Open() tea.Cmd {
	p.open = true
	p.input.Reset()
	p.cursor = 0
	p.applyFilter("")
	return p.input.Focus()
}

// Close hides the palette.
func (p *Palette) Close() {
	p.open = false
	p.input.Blur()
}

// IsOpen reports whether the palette is currently handling input.
func (p *Palette) IsOpen() bool { return p.open }

// Update feeds a message to the palette. On Enter it returns a
// paletteExecMsg command; on Esc a paletteCloseMsg; otherwise typing
// refilters the match list.
func (p *Palette) Update(msg tea.Msg) tea.Cmd {
	if !p.open {
		return nil
	}
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "esc":
			p.Close()
			return func() tea.Msg { return paletteCloseMsg{} }
		case "enter":
			if len(p.matches) == 0 {
				return nil
			}
			entry := p.matches[p.cursor]
			p.Close()
			// If the user typed free-text args after the command ID,
			// prefer those; otherwise fall back to the synthesized
			// Args set when the entry was created (e.g. section index).
			args := paletteExtractArgs(p.input.Value())
			if args == "" {
				args = entry.Args
			}
			id := entry.ID
			return func() tea.Msg { return paletteExecMsg{ActionID: id, Args: args} }
		case "up":
			if p.cursor > 0 {
				p.cursor--
			}
			return nil
		case "down":
			if p.cursor < len(p.matches)-1 {
				p.cursor++
			}
			return nil
		}
	}
	prev := p.input.Value()
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	if p.input.Value() != prev {
		p.applyFilter(p.input.Value())
	}
	return cmd
}

// paletteExtractArgs splits ":cmd arg tail" → "arg tail".
func paletteExtractArgs(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}
	parts := strings.SplitN(query, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (p *Palette) applyFilter(query string) {
	query = strings.TrimSpace(strings.ToLower(query))
	// Match on the head (everything before the first space) so the user
	// can type ":section.jump 5" without the "5" breaking the ranking.
	head := query
	if idx := strings.Index(query, " "); idx >= 0 {
		head = query[:idx]
	}

	if head == "" {
		p.matches = append(p.matches[:0], p.entries...)
		sort.SliceStable(p.matches, func(i, j int) bool {
			if p.matches[i].Category != p.matches[j].Category {
				return p.matches[i].Category < p.matches[j].Category
			}
			return p.matches[i].Title < p.matches[j].Title
		})
		p.clampCursor()
		return
	}

	scored := make([]paletteEntry, 0, len(p.entries))
	for _, e := range p.entries {
		score := fuzzyScore(head, strings.ToLower(e.ID), strings.ToLower(e.Title))
		if score > 0 {
			e.Score = score
			scored = append(scored, e)
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		return scored[i].Title < scored[j].Title
	})
	p.matches = scored
	p.clampCursor()
}

func (p *Palette) clampCursor() {
	if p.cursor >= len(p.matches) {
		p.cursor = 0
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// fuzzyScore is a tiny substring + subsequence scorer. Returns 0 on no
// match, positive scores with higher = better. Prefix matches on the ID
// or Title are boosted.
func fuzzyScore(needle, id, title string) int {
	if needle == "" {
		return 1
	}
	// Whole-word substring on either ID or Title wins outright.
	if strings.Contains(id, needle) {
		if strings.HasPrefix(id, needle) {
			return 100 + len(needle)
		}
		return 60 + len(needle)
	}
	if strings.Contains(title, needle) {
		if strings.HasPrefix(title, needle) {
			return 80 + len(needle)
		}
		return 40 + len(needle)
	}
	// Subsequence match on title.
	if subsequence(title, needle) {
		return 20
	}
	return 0
}

func subsequence(haystack, needle string) bool {
	i := 0
	for _, r := range haystack {
		if i >= len(needle) {
			return true
		}
		if rune(needle[i]) == r {
			i++
		}
	}
	return i >= len(needle)
}

// View renders the palette as a centered modal. The caller is
// responsible for compositing it over the main view (see app.go View()).
func (p *Palette) View(th theme, width int) string {
	if !p.open {
		return ""
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Focus.GetForeground()).
		Background(th.Panel.GetBackground()).
		Padding(1, 2).
		Width(minInt(80, width-4))

	header := th.SectionTitle.Render("Command palette")
	lines := []string{
		header,
		th.Muted.Render("type to fuzzy-match · enter runs · esc cancels"),
		"",
		p.input.View(),
		"",
	}

	if len(p.matches) == 0 {
		lines = append(lines, th.Muted.Render("no matches"))
	} else {
		maxList := 8
		end := len(p.matches)
		if end > maxList {
			end = maxList
		}
		for i := 0; i < end; i++ {
			entry := p.matches[i]
			label := entry.Title
			if entry.Category != "" {
				label = th.CardMeta.Render("["+entry.Category+"] ") + label
			}
			line := "  " + label
			if i == p.cursor {
				line = th.Focus.Render("▸ ") + label
			}
			lines = append(lines, line)
			if entry.Description != "" {
				hint := th.CardMeta.Render("    " + entry.Description)
				lines = append(lines, hint)
			}
		}
	}

	return box.Render(strings.Join(lines, "\n"))
}

// minInt avoids pulling math; keeps this file leaf-clean.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
