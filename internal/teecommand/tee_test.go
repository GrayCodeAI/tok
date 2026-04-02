package teecommand

import "testing"

func TestNewTeeExecutor(t *testing.T) {
	executor := NewTeeExecutor()
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
	if executor.mode != TeeCapture {
		t.Errorf("Expected capture mode, got %s", executor.mode)
	}
}

func TestSavingsTracker(t *testing.T) {
	st := NewSavingsTracker()

	st.Record(100)
	st.Record(200)
	st.Record(300)

	if st.Total() != 600 {
		t.Errorf("Expected 600 total, got %d", st.Total())
	}

	graph := st.RenderASCIIGraph(40)
	if graph == "" {
		t.Error("Expected non-empty graph")
	}
}

func TestSavingsTrackerEmpty(t *testing.T) {
	st := NewSavingsTracker()
	if st.Total() != 0 {
		t.Error("Expected 0 total for empty tracker")
	}
	graph := st.RenderASCIIGraph(40)
	if graph == "" {
		t.Error("Expected non-empty graph even for empty tracker")
	}
}
