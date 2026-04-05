package background

import (
	"testing"
	"time"
)

func TestNewWorker(t *testing.T) {
	w := NewWorker()
	if w == nil {
		t.Fatal("NewWorker returned nil")
	}
}

func TestWorker_StartStop(t *testing.T) {
	w := NewWorker()

	if w.IsRunning() {
		t.Error("worker should not be running initially")
	}

	w.Start()
	time.Sleep(50 * time.Millisecond)

	if !w.IsRunning() {
		t.Error("worker should be running after Start()")
	}

	w.Stop()
	time.Sleep(50 * time.Millisecond)

	if w.IsRunning() {
		t.Error("worker should not be running after Stop()")
	}
}

func TestWorker_Status(t *testing.T) {
	w := NewWorker()
	status := w.Status()
	if status == "" {
		t.Error("Status should return a non-empty string")
	}
}

func TestWorker_DoubleStop(t *testing.T) {
	w := NewWorker()
	w.Start()
	time.Sleep(50 * time.Millisecond)
	w.Stop()
	// Should not panic
	w.Stop()
	if w.IsRunning() {
		t.Error("worker should remain stopped after double Stop()")
	}
}
