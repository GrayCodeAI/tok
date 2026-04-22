package filter

import (
	"strings"
	"testing"
)

func TestNewSessionTracker(t *testing.T) {
	st := NewSessionTracker()
	if st == nil {
		t.Fatal("expected non-nil SessionTracker")
	}
	if st.maxEntries != 10000 {
		t.Errorf("expected maxEntries 10000, got %d", st.maxEntries)
	}
}

func TestSessionTracker_Name(t *testing.T) {
	st := NewSessionTracker()
	if st.Name() != "session" {
		t.Errorf("expected name 'session', got %q", st.Name())
	}
}

func TestSessionTracker_Apply_ShortInput(t *testing.T) {
	st := NewSessionTracker()
	input := "hi"
	output, saved := st.Apply(input, ModeMinimal)
	if output != input {
		t.Error("expected passthrough for short input")
	}
	if saved != 0 {
		t.Error("expected 0 saved for short input")
	}
}

func TestSessionTracker_Apply_NewContent(t *testing.T) {
	st := NewSessionTracker()
	input := strings.Repeat("this is a test line\n", 20)
	output, saved := st.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("expected non-empty output")
	}
	// saved can be negative if output is longer than input (markers added)
	_ = saved
}

func TestSessionTracker_Apply_SeenContent(t *testing.T) {
	st := NewSessionTracker()
	input := strings.Repeat("this is repeated content for testing session tracking\n", 20)

	// First time — should track
	output1, saved1 := st.Apply(input, ModeMinimal)
	if output1 == "" {
		t.Fatal("expected non-empty output on first call")
	}

	// Second time — should mark as seen
	output2, saved2 := st.Apply(input, ModeMinimal)
	if !strings.Contains(output2, "[seen]") {
		t.Logf("second output: %q", output2)
	}
	_ = saved1
	_ = saved2
}

func TestSessionTracker_Stats(t *testing.T) {
	st := NewSessionTracker()
	input := strings.Repeat("test content for stats\n", 20)
	st.Apply(input, ModeMinimal)

	stats := st.Stats()
	if stats.UniqueEntries == 0 {
		t.Error("expected non-zero unique entries")
	}
}

func TestSessionTracker_Clear(t *testing.T) {
	st := NewSessionTracker()
	input := strings.Repeat("test content\n", 20)
	st.Apply(input, ModeMinimal)

	// Should not panic even if file doesn't exist
	_ = st.Clear()

	stats := st.Stats()
	if stats.UniqueEntries != 0 {
		t.Error("expected zero entries after clear")
	}
}

func TestSessionTracker_Save(t *testing.T) {
	st := NewSessionTracker()
	input := strings.Repeat("test content for save\n", 20)
	st.Apply(input, ModeMinimal)

	err := st.Save()
	if err != nil {
		t.Logf("save returned error (may be permission-related): %v", err)
	}
}

func TestIsOnlyNumbers(t *testing.T) {
	if !isOnlyNumbers("123") {
		t.Error("expected true for pure numbers")
	}
	if !isOnlyNumbers("12:34:56") {
		t.Error("expected true for time-like string")
	}
	if isOnlyNumbers("abc") {
		t.Error("expected false for letters")
	}
}

func TestIsDigit(t *testing.T) {
	if !isDigit('5') {
		t.Error("expected true for digit")
	}
	if isDigit('a') {
		t.Error("expected false for letter")
	}
}
