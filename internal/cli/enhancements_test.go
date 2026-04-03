package cli

import (
	"strings"
	"testing"
)

func TestShellCompletionUnsupported(t *testing.T) {
	sc := NewShellCompletion("powershell")
	_, err := sc.Generate([]CommandInfo{{Name: "test"}})
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestShellCompletionBash(t *testing.T) {
	sc := NewShellCompletion("bash")
	cmds := []CommandInfo{
		{Name: "build", Description: "Build project"},
		{Name: "test", Description: "Run tests"},
	}
	script, err := sc.Generate(cmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(script, "build") {
		t.Error("expected 'build' in script")
	}
	if !strings.Contains(script, "test") {
		t.Error("expected 'test' in script")
	}
	if !strings.Contains(script, "_tokman_completion") {
		t.Error("expected completion function")
	}
}

func TestShellCompletionZsh(t *testing.T) {
	sc := NewShellCompletion("zsh")
	cmds := []CommandInfo{
		{Name: "build", Description: "Build project"},
	}
	script, err := sc.Generate(cmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(script, "#compdef tokman") {
		t.Error("expected zsh compdef header")
	}
	if !strings.Contains(script, "build:Build project") {
		t.Error("expected command with description")
	}
}

func TestShellCompletionFish(t *testing.T) {
	sc := NewShellCompletion("fish")
	cmds := []CommandInfo{{Name: "build", Description: "Build project"}}
	script, err := sc.Generate(cmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(script, "complete -c tokman") {
		t.Error("expected fish completion syntax")
	}
}

func TestAliasManager(t *testing.T) {
	am := NewAliasManager()
	am.AddAlias("bm", "benchmark run")
	am.AddAlias("cost", "cost summary")

	if am.ResolveAlias("bm") != "benchmark run" {
		t.Errorf("resolve(bm) = %q, want %q", am.ResolveAlias("bm"), "benchmark run")
	}
	if am.ResolveAlias("unknown") != "unknown" {
		t.Errorf("resolve(unknown) = %q, want %q", am.ResolveAlias("unknown"), "unknown")
	}
}

func TestAliasManagerList(t *testing.T) {
	am := NewAliasManager()
	am.AddAlias("a", "cmd1")
	am.AddAlias("b", "cmd2")

	aliases := am.ListAliases()
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
}

func TestStandardAliases(t *testing.T) {
	aliases := StandardAliases()
	if aliases["bm"] != "benchmark run" {
		t.Errorf("bm alias = %q, want %q", aliases["bm"], "benchmark run")
	}
	if aliases["cost"] != "cost summary" {
		t.Errorf("cost alias = %q, want %q", aliases["cost"], "cost summary")
	}
}

func TestDryRunMode(t *testing.T) {
	dr := NewDryRunMode(true)
	if !dr.IsEnabled() {
		t.Error("expected dry-run enabled")
	}

	dr.Record(DryRunAction{
		Type:        "delete",
		Description: "Remove temp files",
		Parameters:  map[string]interface{}{"path": "/tmp"},
	})

	actions := dr.GetActions()
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != "delete" {
		t.Errorf("action type = %q, want %q", actions[0].Type, "delete")
	}
}

func TestDryRunModeDisabled(t *testing.T) {
	dr := NewDryRunMode(false)
	if dr.IsEnabled() {
		t.Error("expected dry-run disabled")
	}
	// PrintActions should return immediately when disabled (no output check)
	dr.Record(DryRunAction{Type: "noop", Description: "test"})
	dr.PrintActions() // should not panic, but not print
}

func TestCommandChain(t *testing.T) {
	cc := NewCommandChain()
	cc.AddCommand(ChainCommand{Command: "echo", Args: []string{"hello"}})
	cc.AddCommand(ChainCommand{Command: "echo", Args: []string{"world"}})

	if err := cc.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatchOperation(t *testing.T) {
	bo := NewBatchOperation(4, ErrorHandlingContinue)
	bo.AddItem(BatchItem{ID: "1", Action: "build"})
	bo.AddItem(BatchItem{ID: "2", Action: "test"})

	results := bo.Execute()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for %s", r.ID)
		}
	}
}

func TestBatchOperationStopOnError(t *testing.T) {
	bo := NewBatchOperation(1, ErrorHandlingStop)
	if bo.onError != ErrorHandlingStop {
		t.Errorf("onError = %q, want %q", bo.onError, ErrorHandlingStop)
	}
}

func TestBatchOperationRetry(t *testing.T) {
	bo := NewBatchOperation(1, ErrorHandlingRetry)
	if bo.onError != ErrorHandlingRetry {
		t.Errorf("onError = %q, want %q", bo.onError, ErrorHandlingRetry)
	}
}

func TestStandardThemes(t *testing.T) {
	themes := StandardThemes()
	if _, ok := themes["dark"]; !ok {
		t.Error("expected 'dark' theme")
	}
	if _, ok := themes["light"]; !ok {
		t.Error("expected 'light' theme")
	}
	if _, ok := themes["monokai"]; !ok {
		t.Error("expected 'monokai' theme")
	}
}
