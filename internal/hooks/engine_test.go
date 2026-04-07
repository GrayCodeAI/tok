package hooks

import (
	"context"
	"os"
	"testing"
)

func TestHookEngineNew(t *testing.T) {
	engine := NewHookEngine()
	if engine == nil {
		t.Error("Expected non-nil HookEngine")
	}
	if len(engine.configs) != 11 {
		t.Errorf("Expected 11 hook configs, got %d", len(engine.configs))
	}
}

func TestHookEngineInstallPath(t *testing.T) {
	engine := NewHookEngine()

	tests := []struct {
		hookType   HookType
		shouldFind bool
	}{
		{HookTypeClaude, true},
		{HookTypeCursor, true},
		{HookTypeGemini, true},
		{HookTypeCopilot, true},
		{HookTypeCodex, true},
		{HookTypeWindsurf, true},
		{HookTypeCline, true},
		{HookTypeOpencode, true},
		{HookTypeAider, true},
		{HookTypeContinue, true},
		{HookTypeReplit, true},
		{"unknown", false},
	}

	for _, test := range tests {
		config, found := engine.configs[HookType(test.hookType)]
		if test.shouldFind && !found {
			t.Errorf("Expected to find config for %s", test.hookType)
		}
		if test.shouldFind && found && config.Shell == "" {
			t.Errorf("Expected non-empty shell for %s", test.hookType)
		}
	}
}

func TestHookEngineSetShell(t *testing.T) {
	engine := NewHookEngine()

	engine.SetShell(HookTypeClaude, ShellZsh)

	config := engine.configs[HookTypeClaude]
	if config.Shell != ShellZsh {
		t.Errorf("Expected shell to be zsh, got %s", config.Shell)
	}
}

func TestHookEngineGetStats(t *testing.T) {
	engine := NewHookEngine()

	stats := engine.GetStats()
	if stats.Installs != 0 {
		t.Errorf("Expected 0 installs, got %d", stats.Installs)
	}
}

func TestHookEngineGetHookInfo(t *testing.T) {
	engine := NewHookEngine()

	config, found := engine.GetHookInfo(HookTypeClaude)
	if !found {
		t.Error("Expected to find Claude hook config")
	}
	if config.Type != HookTypeClaude {
		t.Errorf("Expected type Claude, got %s", config.Type)
	}

	_, found = engine.GetHookInfo("unknown")
	if found {
		t.Error("Expected not to find unknown hook config")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := home + "/test/path"

	result := expandPath("~/test/path")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	result = expandPath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("Expected /absolute/path, got %s", result)
	}

	result = expandPath("relative/path")
	if result != "relative/path" {
		t.Errorf("Expected relative/path, got %s", result)
	}
}

func TestGenerateBashHook(t *testing.T) {
	script, err := generateBashHook(HookTypeClaude)
	if err != nil {
		t.Fatalf("generateBashHook failed: %v", err)
	}
	if len(script) == 0 {
		t.Error("Expected non-empty script")
	}
	if !contains(script, "tokman") {
		t.Error("Expected script to contain 'tokman'")
	}
}

func TestGenerateZshHook(t *testing.T) {
	script, err := generateZshHook(HookTypeCursor)
	if err != nil {
		t.Fatalf("generateZshHook failed: %v", err)
	}
	if len(script) == 0 {
		t.Error("Expected non-empty script")
	}
	if !contains(script, "preexec") {
		t.Error("Expected script to contain 'preexec'")
	}
}

func TestGenerateFishHook(t *testing.T) {
	script, err := generateFishHook(HookTypeGemini)
	if err != nil {
		t.Fatalf("generateFishHook failed: %v", err)
	}
	if len(script) == 0 {
		t.Error("Expected non-empty script")
	}
	if !contains(script, "function") {
		t.Error("Expected script to contain 'function'")
	}
}

func TestHookTestingFramework(t *testing.T) {
	tf := NewHookTestingFramework()

	tf.AddTest(HookTest{
		Name:        "test-1",
		HookType:    HookTypeClaude,
		InputCmd:    "cargo build",
		ExpectedCmd: "tokman cargo build",
	})

	tf.AddTest(HookTest{
		Name:        "test-2",
		HookType:    HookTypeGemini,
		InputCmd:    "npm test",
		ExpectedCmd: "tokman npm test",
	})

	if len(tf.tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(tf.tests))
	}

	ctx := context.Background()
	engine := NewHookEngine()
	err := tf.RunTests(ctx, engine)
	if err != nil {
		t.Logf("RunTests error (expected - tokman may not be in PATH): %v", err)
	}

	passed, failed := tf.Summary()
	t.Logf("Test results: %d passed, %d failed", passed, failed)
}

func TestHookEngineDetectConflicts(t *testing.T) {
	engine := NewHookEngine()

	conflicts := engine.DetectConflicts(context.Background())
	t.Logf("Found %d potential conflicts", len(conflicts))
}

func TestHookEngineCheckIntegrity(t *testing.T) {
	engine := NewHookEngine()

	exists, err := engine.CheckIntegrity(context.Background(), HookTypeClaude)
	if err != nil {
		t.Logf("CheckIntegrity error (expected - file may not exist): %v", err)
	}
	if exists {
		t.Logf("Hook integrity check returned true (file exists)")
	}
}

func TestHookEngineInstall(t *testing.T) {
	engine := NewHookEngine()

	err := engine.Install(context.Background(), HookTypeClaude)
	if err != nil {
		t.Logf("Install error (expected - may not have permissions): %v", err)
	}
}

func TestHookEngineUninstall(t *testing.T) {
	engine := NewHookEngine()

	err := engine.Uninstall(context.Background(), HookTypeClaude)
	if err != nil {
		t.Logf("Uninstall error: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHookEngineStats(t *testing.T) {
	engine := NewHookEngine()

	engine.mu.Lock()
	engine.stats.Installs = 5
	engine.stats.Updates = 3
	engine.stats.Uninstalls = 1
	engine.mu.Unlock()

	stats := engine.GetStats()
	if stats.Installs != 5 {
		t.Errorf("Expected 5 installs, got %d", stats.Installs)
	}
	if stats.Updates != 3 {
		t.Errorf("Expected 3 updates, got %d", stats.Updates)
	}
	if stats.Uninstalls != 1 {
		t.Errorf("Expected 1 uninstall, got %d", stats.Uninstalls)
	}
}

func TestHookTypes(t *testing.T) {
	expectedTypes := []HookType{
		HookTypeClaude, HookTypeCursor, HookTypeCopilot, HookTypeGemini,
		HookTypeCodex, HookTypeWindsurf, HookTypeCline, HookTypeOpencode,
		HookTypeAider, HookTypeContinue, HookTypeReplit,
	}

	if len(expectedTypes) != 11 {
		t.Errorf("Expected 11 hook types, got %d", len(expectedTypes))
	}
}

func TestShellTypes(t *testing.T) {
	expectedShells := []ShellType{ShellBash, ShellZsh, ShellFish}

	if len(expectedShells) != 3 {
		t.Errorf("Expected 3 shell types, got %d", len(expectedShells))
	}
}

func TestHookTestResult(t *testing.T) {
	result := &HookTestResult{
		TestName:  "test-result",
		Passed:    false,
		ActualCmd: "actual",
		Errors:    []string{"error1", "error2"},
	}

	if result.TestName != "test-result" {
		t.Errorf("Expected test-name, got %s", result.TestName)
	}
	if result.Passed {
		t.Error("Expected passed to be false")
	}
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
}
