package csvexport

import (
	"strings"
	"testing"
)

func TestNewCSVExporter(t *testing.T) {
	headers := []string{"col1", "col2", "col3"}
	e := NewCSVExporter(headers)
	if e == nil {
		t.Fatal("expected non-nil exporter")
	}
	if len(e.headers) != 3 {
		t.Errorf("headers len = %d, want 3", len(e.headers))
	}
}

func TestExport(t *testing.T) {
	e := NewCSVExporter([]string{"id", "name"})
	e.AddRow([]string{"1", "Alice"})
	e.AddRow([]string{"2", "Bob"})

	output := e.Export()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "id,name" {
		t.Errorf("header = %q, want %q", lines[0], "id,name")
	}
	if lines[1] != "1,Alice" {
		t.Errorf("row1 = %q, want %q", lines[1], "1,Alice")
	}
	if lines[2] != "2,Bob" {
		t.Errorf("row2 = %q, want %q", lines[2], "2,Bob")
	}
}

func TestExportEmpty(t *testing.T) {
	e := NewCSVExporter([]string{"a", "b"})
	output := e.Export()
	if output != "a,b\n" {
		t.Errorf("output = %q, want %q", output, "a,b\n")
	}
}

func TestRowCount(t *testing.T) {
	e := NewCSVExporter([]string{"x"})
	if e.RowCount() != 0 {
		t.Errorf("initial count = %d, want 0", e.RowCount())
	}
	e.AddRow([]string{"a"})
	e.AddRow([]string{"b"})
	if e.RowCount() != 2 {
		t.Errorf("count = %d, want 2", e.RowCount())
	}
}

func TestWebhookSenderEmpty(t *testing.T) {
	s := NewWebhookSender()
	if len(s.GetURLs()) != 0 {
		t.Error("expected empty URLs initially")
	}
}

func TestWebhookSenderAddURL(t *testing.T) {
	s := NewWebhookSender()
	s.AddURL("https://hooks.example.com/a")
	s.AddURL("https://hooks.example.com/b")
	urls := s.GetURLs()
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
}

func TestWebhookSenderNotify(t *testing.T) {
	s := NewWebhookSender()
	s.AddURL("https://a.com/hook")
	s.AddURL("https://b.com/hook")

	sent := s.Notify(`{"msg":"test"}`)
	if len(sent) != 2 {
		t.Fatalf("expected 2 sent notifications, got %d", len(sent))
	}
	if sent[0] != "https://a.com/hook" {
		t.Errorf("first = %q, want %q", sent[0], "https://a.com/hook")
	}
}

func TestWebhookSenderNoURLs(t *testing.T) {
	s := NewWebhookSender()
	sent := s.Notify("payload")
	if len(sent) != 0 {
		t.Errorf("expected 0 sent, got %d", len(sent))
	}
}
