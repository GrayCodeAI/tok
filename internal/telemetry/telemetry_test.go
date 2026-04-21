package telemetry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GrayCodeAI/tok/internal/config"
)

func TestGetLocalEventStats(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	eventsPath := filepath.Join(config.DataPath(), "telemetry", LocalEventsFile)
	if err := os.MkdirAll(filepath.Dir(eventsPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := strings.Join([]string{
		`{"feature":"command_invocation","category":"meta","command_path":"tok status","timestamp":"2026-04-18T10:00:00Z"}`,
		`{"feature":"command_invocation","category":"operational","command_path":"tok gain","timestamp":"2026-04-18T12:00:00Z"}`,
		`{"feature":"test_runner","runner_type":"go test","timestamp":"2026-04-17T09:00:00Z"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(eventsPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stats, err := GetLocalEventStats()
	if err != nil {
		t.Fatalf("GetLocalEventStats() error = %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Fatalf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.ByFeature["command_invocation"] != 2 {
		t.Fatalf("ByFeature[command_invocation] = %d, want 2", stats.ByFeature["command_invocation"])
	}
	if stats.ByCategory["meta"] != 1 || stats.ByCategory["operational"] != 1 {
		t.Fatalf("ByCategory = %#v", stats.ByCategory)
	}
	if stats.CommandInvoked != 2 || stats.MetaCommands != 1 || stats.OperationalCmds != 1 {
		t.Fatalf("command/meta/operational counts = %d/%d/%d", stats.CommandInvoked, stats.MetaCommands, stats.OperationalCmds)
	}
	if stats.LastEventAt != "2026-04-18T12:00:00Z" {
		t.Fatalf("LastEventAt = %q", stats.LastEventAt)
	}
	if len(stats.TopCommands) == 0 || !strings.Contains(stats.TopCommands[0], "tok gain") {
		t.Fatalf("TopCommands = %#v", stats.TopCommands)
	}
	if len(stats.TopTestRunners) == 0 || !strings.Contains(stats.TopTestRunners[0], "go test") {
		t.Fatalf("TopTestRunners = %#v", stats.TopTestRunners)
	}
}

func TestForgetConsentRemovesLocalFiles(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	if err := os.MkdirAll(filepath.Join(config.DataPath(), "telemetry"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(getConsentPath(), []byte("enabled"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(localEventsPath(), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := ForgetConsent(); err != nil {
		t.Fatalf("ForgetConsent() error = %v", err)
	}
	if _, err := os.Stat(getConsentPath()); !os.IsNotExist(err) {
		t.Fatalf("consent file still exists: %v", err)
	}
	if _, err := os.Stat(localEventsPath()); !os.IsNotExist(err) {
		t.Fatalf("events file still exists: %v", err)
	}
}
