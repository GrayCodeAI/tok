package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

// testClaudeDir redirects CLAUDE_CONFIG_DIR to a fresh temp dir for the test.
func testClaudeDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	return dir
}

func TestWriteFlag_RejectsInvalidMode(t *testing.T) {
	testClaudeDir(t)
	if err := writeFlag("not-a-mode"); err == nil {
		t.Errorf("expected error for invalid mode")
	}
}

func TestWriteFlag_PersistsValidMode(t *testing.T) {
	dir := testClaudeDir(t)
	if err := writeFlag("ultra"); err != nil {
		t.Fatalf("writeFlag: %v", err)
	}
	path := filepath.Join(dir, ".tok-active")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "ultra" {
		t.Errorf("flag body = %q, want ultra", string(data))
	}
	// Permission check — must be 0600
	st, _ := os.Stat(path)
	if st.Mode().Perm() != 0o600 {
		t.Errorf("flag perms = %o, want 0600", st.Mode().Perm())
	}
}

func TestReadFlag_ReturnsEmptyForMissingFile(t *testing.T) {
	testClaudeDir(t)
	if got := readFlag(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestReadFlag_RoundTripsValidMode(t *testing.T) {
	testClaudeDir(t)
	if err := writeFlag("wenyan-full"); err != nil {
		t.Fatal(err)
	}
	if got := readFlag(); got != "wenyan-full" {
		t.Errorf("round-trip = %q, want wenyan-full", got)
	}
}

func TestReadFlag_RejectsOversizedFile(t *testing.T) {
	dir := testClaudeDir(t)
	path := filepath.Join(dir, ".tok-active")
	// Write a file > 64 bytes
	big := make([]byte, maxFlagBytes+10)
	for i := range big {
		big[i] = 'f'
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, big, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readFlag(); got != "" {
		t.Errorf("oversized flag should yield empty, got %q", got)
	}
}

func TestReadFlag_RejectsSymlink(t *testing.T) {
	dir := testClaudeDir(t)
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("ultra"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, ".tok-active")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported: " + err.Error())
	}
	if got := readFlag(); got != "" {
		t.Errorf("symlinked flag should yield empty, got %q", got)
	}
}

func TestReadFlag_RejectsUnknownMode(t *testing.T) {
	dir := testClaudeDir(t)
	path := filepath.Join(dir, ".tok-active")
	if err := os.WriteFile(path, []byte("super-duper-mode"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readFlag(); got != "" {
		t.Errorf("unknown mode should yield empty, got %q", got)
	}
}

func TestClearFlag_Removes(t *testing.T) {
	dir := testClaudeDir(t)
	_ = writeFlag("full")
	clearFlag()
	if _, err := os.Stat(filepath.Join(dir, ".tok-active")); !os.IsNotExist(err) {
		t.Errorf("clearFlag did not remove file: %v", err)
	}
}

func TestResolveDefaultMode_FallbackFull(t *testing.T) {
	t.Setenv("TOK_DEFAULT_MODE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // empty config dir
	t.Setenv("HOME", t.TempDir())
	if got := resolveDefaultMode(); got != "full" {
		t.Errorf("fallback = %q, want full", got)
	}
}

func TestResolveDefaultMode_EnvWins(t *testing.T) {
	t.Setenv("TOK_DEFAULT_MODE", "ULTRA") // case-insensitive
	if got := resolveDefaultMode(); got != "ultra" {
		t.Errorf("env = %q, want ultra", got)
	}
}

func TestResolveDefaultMode_InvalidEnvIgnored(t *testing.T) {
	t.Setenv("TOK_DEFAULT_MODE", "nonsense")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	if got := resolveDefaultMode(); got != "full" {
		t.Errorf("invalid env should fall through, got %q", got)
	}
}

func TestResolveDefaultMode_ConfigFile(t *testing.T) {
	t.Setenv("TOK_DEFAULT_MODE", "")
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	cfgDir := filepath.Join(xdg, "tok")
	_ = os.MkdirAll(cfgDir, 0o700)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"defaultMode":"lite"}`), 0o600)
	if got := resolveDefaultMode(); got != "lite" {
		t.Errorf("config-file = %q, want lite", got)
	}
}

func TestContainsAll(t *testing.T) {
	if !containsAll("activate tok please", []string{"tok", "activate"}) {
		t.Error("should match both tokens")
	}
	if containsAll("activate please", []string{"tok", "activate"}) {
		t.Error("should not match — tok missing")
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("please turn on now", []string{"enable", "turn on"}) {
		t.Error("should match turn on")
	}
	if containsAny("nope", []string{"enable", "start"}) {
		t.Error("should not match")
	}
}
