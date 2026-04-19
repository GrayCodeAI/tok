package state

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/config"
)

func TestGlobal(t *testing.T) {
	// Reset global
	global = nil

	g := Global()
	if g == nil {
		t.Fatal("expected non-nil global manager")
	}

	// Second call should return same instance
	g2 := Global()
	if g != g2 {
		t.Error("expected same global instance")
	}
}

func TestManager_SetRootCmd(t *testing.T) {
	m := &Manager{}
	cmd := &cobra.Command{Use: "test"}

	m.SetRootCmd(cmd)

	retrieved := m.GetRootCmd()
	if retrieved != cmd {
		t.Error("expected to retrieve same command")
	}
}

func TestManager_SetConfig(t *testing.T) {
	m := &Manager{}
	cfg := &config.Config{}

	m.SetConfig(cfg)

	retrieved := m.GetConfig()
	if retrieved != cfg {
		t.Error("expected to retrieve same config")
	}
}

func TestManager_SetFlags(t *testing.T) {
	m := &Manager{}

	m.SetFlags(2, true, false, "debug", 1000)

	verbose, dryRun, ultraCompact, queryIntent, budget := m.GetFlags()

	if verbose != 2 {
		t.Errorf("expected verbose=2, got %d", verbose)
	}
	if !dryRun {
		t.Error("expected dryRun=true")
	}
	if ultraCompact {
		t.Error("expected ultraCompact=false")
	}
	if queryIntent != "debug" {
		t.Errorf("expected queryIntent='debug', got '%s'", queryIntent)
	}
	if budget != 1000 {
		t.Errorf("expected budget=1000, got %d", budget)
	}
}

func TestManager_IsVerbose(t *testing.T) {
	tests := []struct {
		name    string
		verbose int
		want    bool
	}{
		{"verbose 0", 0, false},
		{"verbose 1", 1, true},
		{"verbose 2", 2, true},
		{"verbose 3", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			m.SetFlags(tt.verbose, false, false, "", 0)

			if got := m.IsVerbose(); got != tt.want {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Version(t *testing.T) {
	m := &Manager{}

	// Default version should be empty initially
	if v := m.GetVersion(); v != "" {
		t.Errorf("expected default version='', got '%s'", v)
	}

	// Set custom version
	m.SetVersion("v1.2.3")
	if v := m.GetVersion(); v != "v1.2.3" {
		t.Errorf("expected version='v1.2.3', got '%s'", v)
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := &Manager{}

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	// Writers
	for i := 0; i < 5; i++ {
		go func(v int) {
			m.SetFlags(v, v%2 == 0, v%2 != 0, "test", v*100)
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		go func() {
			_, _, _, _, _ = m.GetFlags()
			_ = m.IsVerbose()
			_ = m.GetVersion()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	_, _, _, _, _ = m.GetFlags()
}

func TestManager_NilSafety(t *testing.T) {
	m := &Manager{}

	// Should not panic when accessing unset values
	_ = m.GetRootCmd()
	_ = m.GetConfig()
	_ = m.GetVersion()
}
