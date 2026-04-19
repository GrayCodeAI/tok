package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureClaudeHook(t *testing.T) {
	root := map[string]any{}

	changed, err := ensureClaudeHook(root, "/tmp/tok-rewrite.sh")
	if err != nil {
		t.Fatalf("ensureClaudeHook() error = %v", err)
	}
	if !changed {
		t.Fatal("ensureClaudeHook() should report a change")
	}
	if !hasClaudeHook(root, "/tmp/tok-rewrite.sh") {
		t.Fatal("Claude hook should be present after patch")
	}

	changed, err = ensureClaudeHook(root, "/tmp/tok-rewrite.sh")
	if err != nil {
		t.Fatalf("ensureClaudeHook() second call error = %v", err)
	}
	if changed {
		t.Fatal("ensureClaudeHook() should be idempotent")
	}
}

func TestEnsureCursorHook(t *testing.T) {
	root := map[string]any{}

	changed, err := ensureCursorHook(root, "/tmp/tok-rewrite.sh")
	if err != nil {
		t.Fatalf("ensureCursorHook() error = %v", err)
	}
	if !changed {
		t.Fatal("ensureCursorHook() should report a change")
	}
	if version, ok := root["version"].(int); ok && version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}
	if !hasCursorHook(root, "/tmp/tok-rewrite.sh") {
		t.Fatal("Cursor hook should be present after patch")
	}
}

func TestEnsureGeminiHook(t *testing.T) {
	root := map[string]any{}

	changed, err := ensureGeminiHook(root, "/tmp/tok-rewrite.sh")
	if err != nil {
		t.Fatalf("ensureGeminiHook() error = %v", err)
	}
	if !changed {
		t.Fatal("ensureGeminiHook() should report a change")
	}
	if !hasGeminiHook(root, "/tmp/tok-rewrite.sh") {
		t.Fatal("Gemini hook should be present after patch")
	}
}

func TestEnsureReferenceFileContains(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	if err := os.WriteFile(path, []byte("# Existing\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := ensureReferenceFileContains(path, "@TOKMAN.md"); err != nil {
		t.Fatalf("ensureReferenceFileContains() error = %v", err)
	}
	if err := ensureReferenceFileContains(path, "@TOKMAN.md"); err != nil {
		t.Fatalf("ensureReferenceFileContains() second call error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if count := strings.Count(string(content), "@TOKMAN.md"); count != 1 {
		t.Fatalf("reference count = %d, want 1", count)
	}
}

func TestGenerateAgentHookScriptSelectsHandler(t *testing.T) {
	tests := []struct {
		name    string
		agent   string
		handler string
	}{
		{name: "Claude", agent: "Claude Code", handler: "claude"},
		{name: "Cursor", agent: "Cursor", handler: "cursor"},
		{name: "Gemini", agent: "Gemini CLI", handler: "gemini"},
		{name: "Copilot", agent: "GitHub Copilot", handler: "copilot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := generateAgentHookScript(tt.agent)
			want := "exec tok hook " + tt.handler
			if !strings.Contains(script, want) {
				t.Fatalf("generateAgentHookScript(%q) missing %q", tt.agent, want)
			}
		})
	}
}

func TestUninstallAgentClaudeRemovesArtifacts(t *testing.T) {
	home := t.TempDir()
	agent := AgentInfo{
		Name:      "Claude Code",
		ConfigDir: filepath.Join(home, ".claude"),
		HookDir:   filepath.Join(home, ".claude", "hooks"),
	}
	hookPath := filepath.Join(agent.HookDir, "tok-rewrite.sh")
	if err := os.MkdirAll(agent.HookDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(agent.ConfigDir, "TOKMAN.md"), []byte("tok"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(agent.ConfigDir, "CLAUDE.md"), []byte("# Notes\n\n@TOKMAN.md\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := patchClaudeSettingsFile(filepath.Join(agent.ConfigDir, "settings.json"), hookPath); err != nil {
		t.Fatalf("patchClaudeSettingsFile() error = %v", err)
	}

	removed, err := uninstallAgent(agent)
	if err != nil {
		t.Fatalf("uninstallAgent() error = %v", err)
	}
	if len(removed) == 0 {
		t.Fatal("uninstallAgent() should remove artifacts")
	}
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatalf("hook still exists: err=%v", err)
	}
	content, err := os.ReadFile(filepath.Join(agent.ConfigDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(content), "@TOKMAN.md") {
		t.Fatal("CLAUDE.md should no longer reference @TOKMAN.md")
	}
	root, err := loadJSONObject(filepath.Join(agent.ConfigDir, "settings.json"))
	if err != nil {
		t.Fatalf("loadJSONObject() error = %v", err)
	}
	if hasClaudeHook(root, hookPath) {
		t.Fatal("Claude hook entry should be removed from settings.json")
	}
}

func TestRemoveCursorHook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hooks.json")
	root := map[string]any{}
	if _, err := ensureCursorHook(root, "/tmp/tok-rewrite.sh"); err != nil {
		t.Fatalf("ensureCursorHook() error = %v", err)
	}
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	changed, err := removeCursorHook(path, "/tmp/tok-rewrite.sh")
	if err != nil {
		t.Fatalf("removeCursorHook() error = %v", err)
	}
	if !changed {
		t.Fatal("removeCursorHook() should report a change")
	}
	root, err = loadJSONObject(path)
	if err != nil {
		t.Fatalf("loadJSONObject() error = %v", err)
	}
	if hasCursorHook(root, "/tmp/tok-rewrite.sh") {
		t.Fatal("Cursor hook should be removed")
	}
}

func TestSetupAgentCodexLocal(t *testing.T) {
	projectDir := t.TempDir()
	agent := AgentInfo{
		Name:      "Codex",
		ConfigDir: projectDir,
	}

	if err := setupAgent(agent, false); err != nil {
		t.Fatalf("setupAgent() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, "TOKMAN.md")); err != nil {
		t.Fatalf("TOKMAN.md missing: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "@TOKMAN.md") {
		t.Fatal("AGENTS.md should reference @TOKMAN.md")
	}
}

func TestUninstallAgentCodexRemovesLocalArtifacts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := filepath.Join(home, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	agent := AgentInfo{
		Name:      "Codex",
		ConfigDir: projectDir,
	}
	if err := setupAgent(agent, false); err != nil {
		t.Fatalf("setupAgent() error = %v", err)
	}

	removed, err := uninstallAgent(agent)
	if err != nil {
		t.Fatalf("uninstallAgent() error = %v", err)
	}
	if len(removed) == 0 {
		t.Fatal("uninstallAgent() should remove Codex artifacts")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "TOKMAN.md")); !os.IsNotExist(err) {
		t.Fatalf("TOKMAN.md still exists: %v", err)
	}
	if content, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md")); err == nil && strings.Contains(string(content), "@TOKMAN.md") {
		t.Fatal("AGENTS.md should not reference @TOKMAN.md after uninstall")
	}
}

func TestSetupAgentCopilotWritesProjectFiles(t *testing.T) {
	projectDir := t.TempDir()
	agent := AgentInfo{
		Name:      "GitHub Copilot",
		ConfigDir: projectDir,
	}

	if err := setupAgent(agent, false); err != nil {
		t.Fatalf("setupAgent() error = %v", err)
	}

	hookConfig := filepath.Join(projectDir, ".github", "hooks", "tok-rewrite.json")
	instructions := filepath.Join(projectDir, ".github", "copilot-instructions.md")
	if !fileExists(hookConfig) || !fileExists(instructions) {
		t.Fatal("Copilot files should be written in .github/")
	}
	content, err := os.ReadFile(hookConfig)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "tok hook copilot") {
		t.Fatal("Copilot hook config should invoke 'tok hook copilot'")
	}
}

func TestSetupAgentOpenCodeRequiresGlobal(t *testing.T) {
	configDir := t.TempDir()
	agent := AgentInfo{
		Name:      "OpenCode",
		ConfigDir: configDir,
		HookDir:   filepath.Join(configDir, "plugins"),
	}

	if err := setupAgent(agent, false); err == nil {
		t.Fatal("setupAgent() should reject local OpenCode installs")
	}
	if err := setupAgent(agent, true); err != nil {
		t.Fatalf("setupAgent() global error = %v", err)
	}
	pluginPath := filepath.Join(configDir, "plugins", "tok.ts")
	if !fileExists(pluginPath) {
		t.Fatal("OpenCode plugin should be installed")
	}
}

func TestSetupAgentClineWritesManagedRules(t *testing.T) {
	projectDir := t.TempDir()
	rulesPath := filepath.Join(projectDir, ".clinerules")
	if err := os.WriteFile(rulesPath, []byte("# Existing rules\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	agent := AgentInfo{
		Name:      "Cline",
		ConfigDir: projectDir,
	}
	if err := setupAgent(agent, false); err != nil {
		t.Fatalf("setupAgent() error = %v", err)
	}
	if !managedBlockPresent(rulesPath, "tok:cline") {
		t.Fatal(".clinerules should contain tok managed block")
	}

	removed, err := uninstallAgent(agent)
	if err != nil {
		t.Fatalf("uninstallAgent() error = %v", err)
	}
	if len(removed) == 0 {
		t.Fatal("uninstallAgent() should remove managed Cline block")
	}
	content, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(content), "tok:cline") {
		t.Fatal(".clinerules should not contain tok block after uninstall")
	}
}
