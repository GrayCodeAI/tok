package tui

import (
	"strings"
	"testing"
)

func newTestTable(rowCount int) *Table {
	t := NewTable([]Column{
		{Title: "Key", MinWidth: 8, Sortable: true},
		{Title: "Value", Numeric: true, Sortable: true, Align: AlignRight},
	})
	rows := make([]Row, rowCount)
	for i := 0; i < rowCount; i++ {
		rows[i] = Row{Cells: []string{
			"key-" + string(rune('a'+i)),
			itoa(rowCount - i), // descending to verify sort flip
		}}
	}
	t.SetRows(rows)
	return t
}

func itoa(i int) string {
	s := ""
	if i == 0 {
		return "0"
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

func TestTableCursorMovement(t *testing.T) {
	tbl := newTestTable(5)
	if tbl.Cursor() != 0 {
		t.Fatalf("initial cursor = %d, want 0", tbl.Cursor())
	}
	tbl.MoveDown()
	tbl.MoveDown()
	if tbl.Cursor() != 2 {
		t.Fatalf("after 2 MoveDown cursor = %d, want 2", tbl.Cursor())
	}
	tbl.Top()
	if tbl.Cursor() != 0 {
		t.Fatalf("Top() cursor = %d, want 0", tbl.Cursor())
	}
	tbl.Bottom()
	if tbl.Cursor() != 4 {
		t.Fatalf("Bottom() cursor = %d, want 4", tbl.Cursor())
	}
	tbl.MoveDown() // clamped
	if tbl.Cursor() != 4 {
		t.Fatalf("MoveDown beyond bottom cursor = %d, want 4 (clamped)", tbl.Cursor())
	}
}

func TestTableFilter(t *testing.T) {
	tbl := newTestTable(5)
	tbl.SetFilter("key-c")
	if got := tbl.Len(); got != 1 {
		t.Fatalf("filter len = %d, want 1", got)
	}
	tbl.SetFilter("")
	if got := tbl.Len(); got != 5 {
		t.Fatalf("cleared filter len = %d, want 5", got)
	}
}

func TestTableSortToggle(t *testing.T) {
	tbl := newTestTable(4)
	// Rows start in descending-by-value order. Sort asc by col 1.
	tbl.ToggleSort(1)
	if r, ok := tbl.Selected(); !ok || r.Cells[1] != "1" {
		t.Fatalf("asc sort: selected row first cell = %q, want '1'", r.Cells[1])
	}
	// Toggle to desc.
	tbl.ToggleSort(1)
	if r, ok := tbl.Selected(); !ok || r.Cells[1] != "4" {
		t.Fatalf("desc sort: selected row first cell = %q, want '4'", r.Cells[1])
	}
	// Third toggle clears sort.
	tbl.ToggleSort(1)
	if tbl.sortCol != -1 {
		t.Fatalf("third toggle: sortCol = %d, want -1", tbl.sortCol)
	}
}

func TestTableSelectedOnEmpty(t *testing.T) {
	tbl := newTestTable(0)
	if _, ok := tbl.Selected(); ok {
		t.Fatal("Selected() should be false on empty table")
	}
}

func TestTableViewRenders(t *testing.T) {
	tbl := newTestTable(3)
	th := newTheme()
	view := tbl.View(th, 60, 10)
	// Header should include both column titles.
	if !strings.Contains(view, "Key") || !strings.Contains(view, "Value") {
		t.Fatalf("view missing headers:\n%s", view)
	}
	if !strings.Contains(view, "key-a") {
		t.Fatalf("view missing first row:\n%s", view)
	}
}
