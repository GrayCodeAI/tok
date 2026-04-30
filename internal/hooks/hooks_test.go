package hooks

import (
	"path/filepath"
	"testing"
)

func TestGetFlagPath(t *testing.T) {
	// Test with env var
	testDir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", testDir)

	path := GetFlagPath()
	expected := filepath.Join(testDir, ".tok-active")
	if path != expected {
		t.Errorf("GetFlagPath() = %q, want %q", path, expected)
	}
}

func TestActivateDeactivate(t *testing.T) {
	testDir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", testDir)

	// Test activation
	if err := Activate("full"); err != nil {
		t.Errorf("Activate() error = %v", err)
	}

	if !IsActive() {
		t.Error("IsActive() = false after Activate()")
	}

	// Test get mode
	mode := GetMode()
	if mode != "full" {
		t.Errorf("GetMode() = %q, want full", mode)
	}

	// Test deactivation
	if err := Deactivate(); err != nil {
		t.Errorf("Deactivate() error = %v", err)
	}

	if IsActive() {
		t.Error("IsActive() = true after Deactivate()")
	}
}

func TestGetStatusLine(t *testing.T) {
	testDir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", testDir)

	// Not active
	if status := GetStatusLine(); status != "" {
		t.Errorf("GetStatusLine() inactive = %q, want empty", status)
	}

	// Active with full mode
	Activate("full")
	if status := GetStatusLine(); status != "[TOK]" {
		t.Errorf("GetStatusLine() full = %q, want [TOK]", status)
	}

	// Active with ultra mode
	Activate("ultra")
	if status := GetStatusLine(); status != "[TOK:ULTRA]" {
		t.Errorf("GetStatusLine() ultra = %q, want [TOK:ULTRA]", status)
	}
}

func TestAutoActivateOnStartup(t *testing.T) {
	testDir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", testDir)

	// Without env var
	Deactivate()
	if err := AutoActivateOnStartup(); err != nil {
		t.Errorf("AutoActivateOnStartup() error = %v", err)
	}
	if IsActive() {
		t.Error("AutoActivateOnStartup() activated without env var")
	}

	// With env var
	t.Setenv("TOK_AUTO_ACTIVATE", "1")
	t.Setenv("TOK_DEFAULT_MODE", "lite")

	if err := AutoActivateOnStartup(); err != nil {
		t.Errorf("AutoActivateOnStartup() error = %v", err)
	}
	if !IsActive() {
		t.Error("AutoActivateOnStartup() did not activate with env var")
	}
	if mode := GetMode(); mode != "lite" {
		t.Errorf("GetMode() = %q, want lite", mode)
	}
}

func TestResolveDefaultMode(t *testing.T) {
	testDir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", testDir)

	_ = Deactivate()
	t.Setenv("TOK_DEFAULT_MODE", "ultra")
	if got := ResolveDefaultMode(); got != "ultra" {
		t.Errorf("ResolveDefaultMode() = %q, want ultra", got)
	}

	if err := Activate("lite"); err != nil {
		t.Fatalf("Activate(lite) error = %v", err)
	}
	if got := ResolveDefaultMode(); got != "lite" {
		t.Errorf("ResolveDefaultMode() active = %q, want lite", got)
	}
}
