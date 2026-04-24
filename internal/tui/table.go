package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Column describes one column of a Table. Widths are advisory: the
// renderer shrinks columns proportionally when the total exceeds the
// available width.
//
// An accent column renders values as a bar scaled against the column's
// maximum value. Intended for "share" or "reduction %" columns where
// visual weight conveys magnitude better than the digits alone.
type Column struct {
	Title    string
	MinWidth int   // clamps below this; 0 means no clamp
	MaxWidth int   // 0 means no clamp
	Align    Align // default Left
	Accent   bool  // render as bar (cells should be numeric strings)
	Sortable bool  // header responds to a sort keybinding
	Numeric  bool  // right-align and use number comparison for sorting
}

// Align is the horizontal alignment of a cell.
type Align int

const (
	AlignLeft Align = iota
	AlignRight
	AlignCenter
)

// Row is one entry in a Table. Payload is opaque to the table — it's
// passed through to drillDownMsg.Payload so sections can recover their
// domain object from a selected row.
type Row struct {
	Cells   []string
	Payload any
}

// Table is a reusable, scrollable, filterable row widget. Owned by a
// SectionRenderer; the section calls Update() from its own Update() and
// View() from its own View().
type Table struct {
	columns []Column
	rows    []Row
	visible []int // indices into rows after filtering + sorting

	cursor     int
	offset     int
	filter     string
	sortCol    int // -1 = no sort, respect insertion order
	sortDesc   bool
	selectable bool
}

// NewTable builds a table with the given columns. Rows default to empty.
func NewTable(columns []Column) *Table {
	return &Table{
		columns:    columns,
		sortCol:    -1,
		selectable: true,
	}
}

// VisibleRows returns the filtered + sorted rows in display order. The
// exporter uses this so "exported view" matches "what's on screen" —
// if a user filters down to 5 rows then hits `e`, only those 5 land
// in the file.
func (t *Table) VisibleRows() []Row {
	out := make([]Row, 0, len(t.visible))
	for _, i := range t.visible {
		out = append(out, t.rows[i])
	}
	return out
}

// Columns returns the column definitions (copy-safe: the slice is
// shared but columns are value types, not mutated by the caller).
func (t *Table) Columns() []Column { return t.columns }

// SetRows replaces the row set and resets cursor position. Preserves the
// current filter and sort so that a live refresh doesn't jump the user's
// selection off screen.
func (t *Table) SetRows(rows []Row) {
	t.rows = rows
	t.rebuild()
	// Clamp cursor if the new row count is shorter.
	if t.cursor >= len(t.visible) {
		t.cursor = max(0, len(t.visible)-1)
	}
}

// Rows returns the underlying rows (unfiltered, unsorted).
func (t *Table) Rows() []Row { return t.rows }

// Selected returns the row currently under the cursor. ok is false if
// the table is empty.
func (t *Table) Selected() (Row, bool) {
	if len(t.visible) == 0 {
		return Row{}, false
	}
	return t.rows[t.visible[t.cursor]], true
}

// Cursor returns the cursor position among visible rows (0-based).
func (t *Table) Cursor() int { return t.cursor }

// Len returns the number of visible rows.
func (t *Table) Len() int { return len(t.visible) }

// SetFilter applies a substring filter (case-insensitive) across all
// cells. Empty string clears the filter.
func (t *Table) SetFilter(q string) {
	t.filter = strings.TrimSpace(q)
	t.rebuild()
	if t.cursor >= len(t.visible) {
		t.cursor = max(0, len(t.visible)-1)
	}
}

// ToggleSort cycles the sort for the given column between asc → desc →
// off. Non-sortable columns are ignored.
func (t *Table) ToggleSort(col int) {
	if col < 0 || col >= len(t.columns) || !t.columns[col].Sortable {
		return
	}
	switch {
	case t.sortCol != col:
		t.sortCol = col
		t.sortDesc = false
	case !t.sortDesc:
		t.sortDesc = true
	default:
		t.sortCol = -1
	}
	t.rebuild()
}

// SortAccent toggles sort on the first accent column (the "Saved" column
// in most sections). Cycles through ascending → descending → off.
func (t *Table) SortAccent() {
	for i, c := range t.columns {
		if c.Accent {
			t.ToggleSort(i)
			return
		}
	}
}

// MoveUp, MoveDown, Top, Bottom, PageUp, PageDown are the cursor
// primitives a section's Update wires to KeyMap entries.
func (t *Table) MoveUp()   { t.moveBy(-1) }
func (t *Table) MoveDown() { t.moveBy(1) }
func (t *Table) Top()      { t.moveTo(0) }
func (t *Table) Bottom()   { t.moveTo(len(t.visible) - 1) }
func (t *Table) PageUp()   { t.moveBy(-10) }
func (t *Table) PageDown() { t.moveBy(10) }

func (t *Table) moveBy(delta int) {
	if len(t.visible) == 0 {
		t.cursor = 0
		return
	}
	n := t.cursor + delta
	if n < 0 {
		n = 0
	}
	if n >= len(t.visible) {
		n = len(t.visible) - 1
	}
	t.cursor = n
}

func (t *Table) moveTo(pos int) {
	if len(t.visible) == 0 {
		t.cursor = 0
		return
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(t.visible) {
		pos = len(t.visible) - 1
	}
	t.cursor = pos
}

// rebuild refreshes the visible index after a filter or sort change.
func (t *Table) rebuild() {
	visible := make([]int, 0, len(t.rows))
	filter := strings.ToLower(t.filter)
	for i, r := range t.rows {
		if filter == "" || rowMatches(r, filter) {
			visible = append(visible, i)
		}
	}
	if t.sortCol >= 0 && t.sortCol < len(t.columns) {
		col := t.sortCol
		numeric := t.columns[col].Numeric
		desc := t.sortDesc
		sort.SliceStable(visible, func(a, b int) bool {
			av := cellAt(t.rows[visible[a]], col)
			bv := cellAt(t.rows[visible[b]], col)
			less := compareCells(av, bv, numeric)
			if desc {
				return !less
			}
			return less
		})
	}
	t.visible = visible
}

func rowMatches(r Row, needle string) bool {
	// Use simple substring matching for filtering
	for _, c := range r.Cells {
		if strings.Contains(strings.ToLower(c), needle) {
			return true
		}
	}
	return false
}

func cellAt(r Row, col int) string {
	if col < 0 || col >= len(r.Cells) {
		return ""
	}
	return r.Cells[col]
}

func compareCells(a, b string, numeric bool) bool {
	if numeric {
		var af, bf float64
		fmt.Sscanf(a, "%f", &af)
		fmt.Sscanf(b, "%f", &bf)
		return af < bf
	}
	return a < b
}

// --- Rendering ------------------------------------------------------------

// View renders the table to the given width and height. Height includes
// the header row + separator + body.
func (t *Table) View(th theme, width, height int) string {
	if width < 12 {
		return th.Muted.Render("table too narrow")
	}
	widths := t.computeWidths(width)
	header := t.renderHeader(th, widths)
	sep := strings.Repeat("─", width)
	body := t.renderBody(th, widths, max(1, height-2))
	footer := t.renderStatus(th, width)
	return strings.Join([]string{header, th.Muted.Render(sep), body, footer}, "\n")
}

func (t *Table) computeWidths(total int) []int {
	n := len(t.columns)
	if n == 0 {
		return nil
	}
	// Allocate in three passes so columns never oversubscribe the row:
	//
	//  1. Seed every column with its MinWidth (or 0). MaxWidth columns
	//     are fixed.
	//  2. Any surplus beyond the seed is distributed evenly across
	//     non-MaxWidth columns.
	//  3. If the seed already exceeds the total, proportionally shrink
	//     non-MaxWidth columns down to fit. Shrinking below MinWidth
	//     is allowed as a last resort — losing a few chars off a cell
	//     is preferable to a wrapping header that breaks the whole
	//     table layout.
	widths := make([]int, n)
	gutters := n - 1
	if gutters < 0 {
		gutters = 0
	}

	seed := 0
	flexibleIdx := make([]int, 0, n)
	for i, c := range t.columns {
		if c.MaxWidth > 0 {
			widths[i] = c.MaxWidth
			seed += c.MaxWidth
			continue
		}
		widths[i] = c.MinWidth
		seed += c.MinWidth
		flexibleIdx = append(flexibleIdx, i)
	}

	budget := total - gutters
	surplus := budget - seed

	switch {
	case surplus > 0 && len(flexibleIdx) > 0:
		// Distribute extra room round-robin so the remainder doesn't
		// fall entirely on one column.
		for surplus > 0 {
			progress := false
			for _, i := range flexibleIdx {
				if surplus == 0 {
					break
				}
				widths[i]++
				surplus--
				progress = true
			}
			if !progress {
				break
			}
		}
	case surplus < 0:
		// Seed exceeds budget — steal from flexible columns first,
		// largest ones first so narrow columns stay readable.
		deficit := -surplus
		for deficit > 0 {
			progress := false
			// Find the widest flexible column that's still >1.
			widest := -1
			for _, i := range flexibleIdx {
				if widths[i] <= 1 {
					continue
				}
				if widest == -1 || widths[i] > widths[widest] {
					widest = i
				}
			}
			if widest == -1 {
				break
			}
			widths[widest]--
			deficit--
			progress = true
			if !progress {
				break
			}
		}
	}
	return widths
}

func (t *Table) renderHeader(th theme, widths []int) string {
	cells := make([]string, 0, len(t.columns))
	for i, c := range t.columns {
		label := c.Title
		if t.sortCol == i {
			if t.sortDesc {
				label += " ↓"
			} else {
				label += " ↑"
			}
		}
		cells = append(cells, alignCell(label, widths[i], c.Align))
	}
	return th.TableHeader.Render(strings.Join(cells, " "))
}

func (t *Table) renderBody(th theme, widths []int, height int) string {
	if len(t.visible) == 0 {
		return th.Muted.Render("no rows")
	}
	// Keep the cursor in view by scrolling the offset window.
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+height {
		t.offset = t.cursor - height + 1
	}
	end := t.offset + height
	if end > len(t.visible) {
		end = len(t.visible)
	}
	lines := make([]string, 0, end-t.offset)

	// Track column max-values for accent bars (only for numeric columns
	// flagged Accent). Computed over the filtered visible set so bars
	// reflect the currently-shown distribution, not the raw data.
	accentMax := make(map[int]float64)
	for i, c := range t.columns {
		if c.Accent {
			var m float64
			for _, idx := range t.visible {
				var v float64
				fmt.Sscanf(cellAt(t.rows[idx], i), "%f", &v)
				if v > m {
					m = v
				}
			}
			accentMax[i] = m
		}
	}

	for i := t.offset; i < end; i++ {
		row := t.rows[t.visible[i]]
		cells := make([]string, 0, len(t.columns))
		for ci, col := range t.columns {
			raw := cellAt(row, ci)
			cell := alignCell(raw, widths[ci], col.Align)
			if col.Accent && accentMax[ci] > 0 {
				var v float64
				fmt.Sscanf(raw, "%f", &v)
				cell = renderAccentCell(th, cell, v, accentMax[ci], widths[ci])
			}
			cells = append(cells, cell)
		}
		line := strings.Join(cells, " ")
		style := th.TableRow
		if t.selectable && i == t.cursor {
			style = th.SidebarActive
		} else if i == t.offset && t.sortCol >= 0 {
			style = th.TableRowAccent
		}
		lines = append(lines, style.Render(line))
	}
	return strings.Join(lines, "\n")
}

func (t *Table) renderStatus(th theme, width int) string {
	total := len(t.rows)
	shown := len(t.visible)
	pos := 0
	if shown > 0 {
		pos = t.cursor + 1
	}
	status := fmt.Sprintf("%d/%d", pos, shown)
	if t.filter != "" {
		status += fmt.Sprintf(" · /%s", t.filter)
	}
	if shown < total {
		status += fmt.Sprintf(" · %d hidden", total-shown)
	}
	return th.CardMeta.Render(lipgloss.PlaceHorizontal(width, lipgloss.Right, status))
}

func alignCell(s string, width int, align Align) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) > width {
		s = truncate(s, width)
	}
	switch align {
	case AlignRight:
		return lipgloss.PlaceHorizontal(width, lipgloss.Right, s)
	case AlignCenter:
		return lipgloss.PlaceHorizontal(width, lipgloss.Center, s)
	default:
		return lipgloss.PlaceHorizontal(width, lipgloss.Left, s)
	}
}

// renderAccentCell renders a numeric value + a trailing mini-bar. The
// value occupies the left half of the cell, the bar the right half.
func renderAccentCell(th theme, valueText string, v, maxV float64, width int) string {
	half := width / 2
	if half < 4 {
		return valueText
	}
	value := lipgloss.PlaceHorizontal(half, lipgloss.Right, valueText)
	bar := renderBar(th, int64(v), int64(maxV), width-half-1, 0)
	return value + " " + bar
}
