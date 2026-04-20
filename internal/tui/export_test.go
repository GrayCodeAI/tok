package tui

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeExport is a tiny ExportableTable used so the test doesn't
// depend on the real section types.
type fakeExport struct {
	name string
	cols []Column
	rows []Row
}

func (f fakeExport) ExportColumns() []Column { return f.cols }
func (f fakeExport) ExportRows() []Row       { return f.rows }
func (f fakeExport) ExportName() string      { return f.name }

func fakeTable() fakeExport {
	return fakeExport{
		name: "fake",
		cols: []Column{{Title: "Key"}, {Title: "Value"}},
		rows: []Row{
			{Cells: []string{"alpha", "1"}},
			{Cells: []string{"beta|bar", "2"}}, // pipe to verify md escaping
		},
	}
}

func TestExportJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOK_EXPORT_DIR", dir)

	path, err := writeExport(fakeTable(), ExportJSON)
	if err != nil {
		t.Fatalf("writeExport: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("expected file under %s, got %s", dir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	if records[0]["Key"] != "alpha" || records[0]["Value"] != "1" {
		t.Fatalf("row 0 mismatch: %+v", records[0])
	}
}

func TestExportCSV(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOK_EXPORT_DIR", dir)

	path, err := writeExport(fakeTable(), ExportCSV)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("csv parse: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records = %d, want 3 (header + 2 data)", len(records))
	}
	if records[0][0] != "Key" || records[1][0] != "alpha" {
		t.Fatalf("csv contents: %+v", records)
	}
}

func TestExportMarkdownEscapesPipes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOK_EXPORT_DIR", dir)

	path, err := writeExport(fakeTable(), ExportMD)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `beta\|bar`) {
		t.Fatalf("markdown export should escape |: got\n%s", content)
	}
	if !strings.Contains(content, "| Key | Value |") {
		t.Fatalf("markdown header missing: %s", content)
	}
}

func TestExportBaseDirFromEnv(t *testing.T) {
	t.Setenv("TOK_EXPORT_DIR", "/tmp/tok-exports-test")
	if got := exportBaseDir(); got != "/tmp/tok-exports-test" {
		t.Fatalf("exportBaseDir = %q, want /tmp/tok-exports-test", got)
	}
}
