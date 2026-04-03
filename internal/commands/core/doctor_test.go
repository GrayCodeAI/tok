package core

import (
	"strings"
	"testing"
)

func TestCheckBinary(t *testing.T) {
	r := checkBinary()
	if r.Name != "Binary" {
		t.Errorf("name = %q, want %q", r.Name, "Binary")
	}
	if r.Status == "error" && strings.Contains(r.Message, "cannot determine") {
		t.Logf("binary check returned error: %s (may be OK in test env)", r.Message)
	} else if r.Status == "ok" {
		if r.Message == "" {
			t.Error("expected non-empty message for ok status")
		}
	}
}

func TestCheckConfigDir(t *testing.T) {
	r := checkConfigDir()
	if r.Name != "Config Dir" {
		t.Errorf("name = %q, want %q", r.Name, "Config Dir")
	}
	// Should be either ok or warn (warn if config dir doesn't exist, which is common in tests)
	if r.Status != "ok" && r.Status != "warn" {
		t.Errorf("status = %q, want ok or warn", r.Status)
	}
}

func TestCheckDatabase(t *testing.T) {
	r := checkDatabase()
	if r.Name != "Database" {
		t.Errorf("name = %q, want %q", r.Name, "Database")
	}
	// Database may not exist yet in test env
	if r.Status != "ok" && r.Status != "warn" {
		t.Logf("database check: %s: %s", r.Status, r.Message)
	}
}

func TestCheckShellHook(t *testing.T) {
	r := checkShellHook()
	if r.Name != "Shell Hook" {
		t.Errorf("name = %q, want %q", r.Name, "Shell Hook")
	}
	// Shell hook likely not present in test env
	if r.Status == "warn" && strings.Contains(r.Message, "no shell hook found") {
		// Expected in test environment
		return
	}
	if r.Status == "ok" && r.Message != "" {
		// OK if hook exists
		return
	}
	t.Errorf("unexpected status: %s: %s", r.Status, r.Message)
}

func TestCheckPath(t *testing.T) {
	r := checkPath()
	if r.Name != "PATH" {
		t.Errorf("name = %q, want %q", r.Name, "PATH")
	}
	if r.Status != "ok" && r.Status != "warn" {
		t.Errorf("status = %q, want ok or warn", r.Status)
	}
}

func TestCheckPlatform(t *testing.T) {
	r := checkPlatform()
	if r.Name != "Platform" {
		t.Errorf("name = %q, want %q", r.Name, "Platform")
	}
	if r.Status != "ok" {
		t.Errorf("platform check should always be ok, got %q", r.Status)
	}
	if r.Message == "" {
		t.Error("expected non-empty platform info")
	}
}

func TestCheckTokenizer(t *testing.T) {
	r := checkTokenizer()
	if r.Name != "Tokenizer" {
		t.Errorf("name = %q, want %q", r.Name, "Tokenizer")
	}
	// Tokenizer may or may not be available
	if r.Status != "ok" && r.Status != "warn" {
		t.Errorf("status = %q, want ok or warn", r.Status)
	}
}

func TestCheckTierSystem(t *testing.T) {
	r := checkTierSystem()
	if r.Status != "ok" && r.Status != "warn" {
		t.Errorf("status = %q, want ok or warn", r.Status)
	}
}
