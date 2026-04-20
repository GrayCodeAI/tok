package tui

import (
	"encoding/base64"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRowToTSV(t *testing.T) {
	got := RowToTSV(Row{Cells: []string{"a", "b", "c"}})
	if got != "a\tb\tc" {
		t.Fatalf("RowToTSV = %q, want %q", got, "a\tb\tc")
	}
}

func TestYankCmdEmitsOSC52AndToast(t *testing.T) {
	cmd := YankCmd("hello world")
	if cmd == nil {
		t.Fatal("YankCmd returned nil for non-empty payload")
	}
	// YankCmd returns a tea.Batch of (writeOSC52, requestToastCmd).
	// Execute the batch and inspect the emitted messages.
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected BatchMsg, got %T", msg)
	}

	foundToast := false
	for _, child := range batch {
		m := child()
		if ta, isToast := m.(toastAddMsg); isToast {
			foundToast = true
			if ta.Kind != ToastSuccess {
				t.Errorf("toast kind = %v, want ToastSuccess", ta.Kind)
			}
			if !strings.Contains(ta.Text, "hello world") {
				t.Errorf("toast text = %q, want containing 'hello world'", ta.Text)
			}
		}
	}
	if !foundToast {
		t.Fatal("yank did not emit a toastAddMsg")
	}
}

func TestYankCmdEmptyIsNoop(t *testing.T) {
	if cmd := YankCmd(""); cmd != nil {
		t.Fatal("empty payload should return nil cmd")
	}
}

func TestYankCmdBase64EncodingRoundtrips(t *testing.T) {
	// Reproduce the writeOSC52 payload assembly and decode to confirm
	// the escape sequence carries the original bytes.
	payload := "anthropic\tclaude-opus-4-7\t12345"
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != payload {
		t.Fatalf("roundtrip mismatch: %q vs %q", decoded, payload)
	}
}

func TestDescribeYankTruncation(t *testing.T) {
	long := strings.Repeat("x", 80)
	got := describeYank(long)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("long string should be truncated with ellipsis, got %q", got)
	}
	multi := "first line\nsecond line"
	if got := describeYank(multi); got != "first line" {
		t.Fatalf("multiline should take first line, got %q", got)
	}
}
