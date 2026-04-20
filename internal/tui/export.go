package tui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ExportableTable is implemented by sections that expose their current
// table state to the export subsystem. Sections opt in by returning
// their columns + visible rows (post-filter, post-sort). Rows stay raw
// strings — the exporter formats them; we don't export rendered ANSI.
type ExportableTable interface {
	ExportColumns() []Column
	ExportRows() []Row
	ExportName() string // section name used in the output filename
}

// exportBaseDir is ~/.tok/exports by default; overridable via env for
// tests. Created lazily on first export.
func exportBaseDir() string {
	if d := os.Getenv("TOK_EXPORT_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "tok-exports")
	}
	return filepath.Join(home, ".tok", "exports")
}

// ExportTableCmd writes the table's current visible rows to the
// requested format and dispatches a toast with the destination path.
// Errors bubble through actionResultMsg, not toast, so section Update
// loops can layer their own handling if desired.
func ExportTableCmd(t ExportableTable, format ExportFormat) tea.Cmd {
	return func() tea.Msg {
		path, err := writeExport(t, format)
		if err != nil {
			return tea.Batch(
				requestToastCmd(ToastError, "export failed: "+err.Error()),
			)()
		}
		return tea.Batch(
			requestToastCmd(ToastSuccess, "exported → "+path),
		)()
	}
}

func writeExport(t ExportableTable, format ExportFormat) (string, error) {
	dir := exportBaseDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create export dir: %w", err)
	}
	ts := time.Now().Format("20060102-150405")
	name := strings.ToLower(t.ExportName())
	if name == "" {
		name = "section"
	}
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.%s", name, ts, format))

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	cols := t.ExportColumns()
	rows := t.ExportRows()

	switch format {
	case ExportJSON:
		return path, writeJSON(f, cols, rows)
	case ExportCSV:
		return path, writeCSV(f, cols, rows)
	case ExportMD:
		return path, writeMarkdown(f, cols, rows)
	default:
		return "", fmt.Errorf("unknown export format: %s", format)
	}
}

type jsonRecord map[string]any

func writeJSON(f *os.File, cols []Column, rows []Row) error {
	records := make([]jsonRecord, 0, len(rows))
	for _, r := range rows {
		rec := jsonRecord{}
		for i, c := range cols {
			if i < len(r.Cells) {
				rec[c.Title] = r.Cells[i]
			}
		}
		if r.Payload != nil {
			rec["_payload"] = r.Payload
		}
		records = append(records, rec)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

func writeCSV(f *os.File, cols []Column, rows []Row) error {
	w := csv.NewWriter(f)
	defer w.Flush()

	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.Title
	}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, r := range rows {
		record := make([]string, len(cols))
		for i := range cols {
			if i < len(r.Cells) {
				record[i] = r.Cells[i]
			}
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func writeMarkdown(f *os.File, cols []Column, rows []Row) error {
	// Title row + separator, then rows. Use pipe escaping for cells
	// that contain literal pipes.
	b := &strings.Builder{}
	b.WriteString("|")
	for _, c := range cols {
		b.WriteString(" " + c.Title + " |")
	}
	b.WriteString("\n|")
	for range cols {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")
	for _, r := range rows {
		b.WriteString("|")
		for i := range cols {
			cell := ""
			if i < len(r.Cells) {
				cell = strings.ReplaceAll(r.Cells[i], "|", "\\|")
			}
			b.WriteString(" " + cell + " |")
		}
		b.WriteString("\n")
	}
	_, err := f.WriteString(b.String())
	return err
}
