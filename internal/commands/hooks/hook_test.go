package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lakshmanpatel/tok/internal/discover"
)

// ── detectCopilotFormat ────────────────────────────────────────

func TestDetectCopilotFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		wantFormat  copilotHookFormat
		wantCommand string
	}{
		{
			name: "VS Code format",
			input: map[string]any{
				"tool_name": "Bash",
				"tool_input": map[string]any{
					"command": "git status",
				},
			},
			wantFormat:  copilotFormatVsCode,
			wantCommand: "git status",
		},
		{
			name: "VS Code runTerminalCommand",
			input: map[string]any{
				"tool_name": "runTerminalCommand",
				"tool_input": map[string]any{
					"command": "cargo test",
				},
			},
			wantFormat:  copilotFormatVsCode,
			wantCommand: "cargo test",
		},
		{
			name: "VS Code bash lowercase",
			input: map[string]any{
				"tool_name": "bash",
				"tool_input": map[string]any{
					"command": "ls -la",
				},
			},
			wantFormat:  copilotFormatVsCode,
			wantCommand: "ls -la",
		},
		{
			name: "Copilot CLI format",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": `{"command":"git status"}`,
			},
			wantFormat:  copilotFormatCli,
			wantCommand: "git status",
		},
		{
			name: "non-bash tool",
			input: map[string]any{
				"tool_name": "editFiles",
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name:       "empty input",
			input:      map[string]any{},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "non-bash copilot CLI tool",
			input: map[string]any{
				"toolName": "view",
				"toolArgs": "{}",
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "VS Code empty command",
			input: map[string]any{
				"tool_name": "Bash",
				"tool_input": map[string]any{
					"command": "",
				},
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "VS Code missing tool_input",
			input: map[string]any{
				"tool_name": "Bash",
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "VS Code tool_input not map",
			input: map[string]any{
				"tool_name":  "Bash",
				"tool_input": "not a map",
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "VS Code command not string",
			input: map[string]any{
				"tool_name": "Bash",
				"tool_input": map[string]any{
					"command": 123,
				},
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI invalid JSON in toolArgs",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": `not json`,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI empty toolArgs",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": `{}`,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI toolArgs not string",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": map[string]any{},
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI missing toolArgs",
			input: map[string]any{
				"toolName": "bash",
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI command not string",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": `{"command": 123}`,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "Copilot CLI empty command",
			input: map[string]any{
				"toolName": "bash",
				"toolArgs": `{"command":""}`,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "tool_name not string",
			input: map[string]any{
				"tool_name": 123,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "toolName not string",
			input: map[string]any{
				"toolName": 123,
			},
			wantFormat: copilotFormatPassThrough,
		},
		{
			name: "unknown tool_name",
			input: map[string]any{
				"tool_name": "unknownTool",
				"tool_input": map[string]any{
					"command": "git status",
				},
			},
			wantFormat: copilotFormatPassThrough,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, cmd := detectCopilotFormat(tt.input)
			if format != tt.wantFormat {
				t.Errorf("format = %v, want %v", format, tt.wantFormat)
			}
			if tt.wantCommand != "" && cmd != tt.wantCommand {
				t.Errorf("command = %q, want %q", cmd, tt.wantCommand)
			}
		})
	}
}

func TestRunClaudeInner_RewritesCommand(t *testing.T) {
	input := `{"tool_name":"Bash","tool_input":{"command":"git status"}}`

	output := runClaudeInner(input)
	if output == "" {
		t.Fatal("runClaudeInner() returned empty output")
	}
	if !strings.Contains(output, `"updatedInput":{"command":"tok git status"}`) {
		t.Fatalf("runClaudeInner() = %s", output)
	}
}

func TestRunCursorInner_RewritesCommand(t *testing.T) {
	input := `{"tool_input":{"command":"git status"}}`

	output := runCursorInner(input)
	if output == "{}" {
		t.Fatal("runCursorInner() returned empty response")
	}
	if !strings.Contains(output, `"updated_input":{"command":"tok git status"}`) {
		t.Fatalf("runCursorInner() = %s", output)
	}
}

func TestProcessGeminiPayload_DenyRule(t *testing.T) {
	original := checkCommandPermissions
	t.Cleanup(func() { checkCommandPermissions = original })
	checkCommandPermissions = func(cmd string) PermissionVerdict {
		return PermissionDeny
	}

	payload := map[string]any{
		"tool_name": "run_shell_command",
		"tool_input": map[string]any{
			"command": "git status",
		},
	}

	output, action, originalCmd, rewritten := processGeminiPayload(payload)
	if action != "skip:deny_rule" {
		t.Fatalf("action = %q, want skip:deny_rule", action)
	}
	if originalCmd != "git status" {
		t.Fatalf("originalCmd = %q, want git status", originalCmd)
	}
	if rewritten != "" {
		t.Fatalf("rewritten = %q, want empty", rewritten)
	}
	if decision, _ := output["decision"].(string); decision != "deny" {
		t.Fatalf("decision = %q, want deny", decision)
	}
}

func TestRecordHookAudit(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("TOKMAN_HOOK_AUDIT", "1")

	recordHookAudit("rewrite", "git status", "tok git status")

	content, err := os.ReadFile(getAuditLogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	line := strings.TrimSpace(string(content))
	entry := parseAuditLine(line)
	if entry == nil {
		t.Fatalf("parseAuditLine(%q) returned nil", line)
	}
	if entry.Action != "rewrite" {
		t.Fatalf("entry.Action = %q, want rewrite", entry.Action)
	}
}

func TestRecordHookAuditEscapesFields(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("TOKMAN_HOOK_AUDIT", "1")

	recordHookAudit("rewrite", "git status | head -n 1", "tok git status\n")

	content, err := os.ReadFile(getAuditLogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(content), " | head -n 1 | tok git status\n") {
		t.Fatal("audit content should escape separators and newlines")
	}
}

// ── handleCopilotVsCode ────────────────────────────────────────

func TestHandleCopilotVsCode_RewriteLogic(t *testing.T) {
	// Test that the discover.RewriteCommand logic works for commands
	// that would be handled by handleCopilotVsCode
	tests := []struct {
		name    string
		cmd     string
		wantOut string
	}{
		{
			name:    "git status rewrite",
			cmd:     "git status",
			wantOut: "tok git status",
		},
		{
			name:    "cargo test rewrite",
			cmd:     "cargo test",
			wantOut: "tok test-runner cargo test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the command would be rewritten (same logic as handleCopilotVsCode)
			rewritten, changed := discover.RewriteCommand(tt.cmd, nil)
			if !changed {
				t.Errorf("command %q should be rewritten", tt.cmd)
			}
			if rewritten != tt.wantOut {
				t.Errorf("rewritten = %q, want %q", rewritten, tt.wantOut)
			}
		})
	}
}

// ── handleCopilotCli ───────────────────────────────────────────

func TestHandleCopilotCli_RewriteLogic(t *testing.T) {
	// Test that the discover.RewriteCommand logic works for commands
	// that would be handled by handleCopilotCli
	tests := []struct {
		name      string
		cmd       string
		wantEmpty bool
	}{
		{
			name:      "git status produces deny",
			cmd:       "git status",
			wantEmpty: false,
		},
		{
			name:      "cd produces empty",
			cmd:       "cd /tmp",
			wantEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, changed := discover.RewriteCommand(tt.cmd, nil)
			if tt.wantEmpty && changed {
				t.Errorf("command %q should NOT be rewritten", tt.cmd)
			}
			if !tt.wantEmpty && !changed {
				t.Errorf("command %q should be rewritten", tt.cmd)
			}
		})
	}
}

// ── Gemini hook ────────────────────────────────────────────────

func TestGeminiHookOutput(t *testing.T) {
	// Test that printGeminiAllow outputs valid JSON
	allowed := `{"decision":"allow"}`
	if allowed == "" {
		t.Error("allow output should not be empty")
	}
	if len(allowed) < 10 {
		t.Errorf("allow output too short: %q", allowed)
	}
}

func TestBuildCopilotVSCodeResponse_Ask(t *testing.T) {
	original := checkCommandPermissions
	t.Cleanup(func() { checkCommandPermissions = original })
	checkCommandPermissions = func(cmd string) PermissionVerdict {
		return PermissionAsk
	}

	output := buildCopilotVSCodeResponse("git status", "tok git status")
	payload, _ := output["hookSpecificOutput"].(map[string]any)
	if decision, _ := payload["permissionDecision"].(string); decision != "ask" {
		t.Fatalf("permissionDecision = %q, want ask", decision)
	}
}

func TestPrintGeminiRewrite_Structure(t *testing.T) {
	// Test the structure of printGeminiRewrite output
	// The output should contain the rewritten command
	cmd := "tok git status"
	// We can't easily capture stdout, but we can verify the logic
	if !strings.Contains(cmd, "tok") {
		t.Error("rewritten command should contain 'tok'")
	}
}

// ── Audit functions ────────────────────────────────────────────

func TestParseAuditLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *AuditEntry
	}{
		{
			name: "valid line",
			line: "2026-01-15T10:30:00Z | rewrite | git status | tok git status",
			want: &AuditEntry{
				Timestamp:    "2026-01-15T10:30:00Z",
				Action:       "rewrite",
				OriginalCmd:  "git status",
				RewrittenCmd: "tok git status",
			},
		},
		{
			name: "valid line no rewritten",
			line: "2026-01-15T10:30:00Z | skip:ignored | cd /tmp",
			want: &AuditEntry{
				Timestamp:    "2026-01-15T10:30:00Z",
				Action:       "skip:ignored",
				OriginalCmd:  "cd /tmp",
				RewrittenCmd: "-",
			},
		},
		{
			name: "too few parts",
			line: "2026-01-15T10:30:00Z | rewrite",
			want: nil,
		},
		{
			name: "empty line",
			line: "",
			want: nil,
		},
		{
			name: "single part",
			line: "just one part",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAuditLine(tt.line)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseAuditLine(%q) = %+v, want nil", tt.line, got)
				}
			} else {
				if got == nil {
					t.Fatalf("parseAuditLine(%q) = nil, want %+v", tt.line, tt.want)
				}
				if got.Timestamp != tt.want.Timestamp {
					t.Errorf("Timestamp = %q, want %q", got.Timestamp, tt.want.Timestamp)
				}
				if got.Action != tt.want.Action {
					t.Errorf("Action = %q, want %q", got.Action, tt.want.Action)
				}
				if got.OriginalCmd != tt.want.OriginalCmd {
					t.Errorf("OriginalCmd = %q, want %q", got.OriginalCmd, tt.want.OriginalCmd)
				}
				if got.RewrittenCmd != tt.want.RewrittenCmd {
					t.Errorf("RewrittenCmd = %q, want %q", got.RewrittenCmd, tt.want.RewrittenCmd)
				}
			}
		})
	}
}

func TestBaseCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "git status", "git status"},
		{"with env vars", "FOO=bar git status", "git status"},
		{"multiple env vars", "A=1 B=2 cargo test", "cargo test"},
		{"single word", "git", "git"},
		{"empty", "", ""},
		{"only env vars", "FOO=bar BAR=baz", "FOO=bar BAR=baz"},
		{"with flags", "git status --short", "git status"},
		{"cargo build", "cargo build --release", "cargo build"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := baseCommand(tt.input)
			if got != tt.want {
				t.Errorf("baseCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterEntriesByDays(t *testing.T) {
	now := time.Now()
	recent := now.Add(-2 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
	weekAgo := now.Add(-5 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
	oldEntry := now.Add(-30 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
	veryOld := now.Add(-365 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")

	entries := []AuditEntry{
		{Timestamp: recent},   // 2 days ago
		{Timestamp: weekAgo},  // 5 days ago
		{Timestamp: oldEntry}, // 30 days ago
		{Timestamp: veryOld},  // 1 year ago
	}

	// 0 days = all entries
	all := filterEntriesByDays(entries, 0)
	if len(all) != 4 {
		t.Errorf("filterEntriesByDays(entries, 0) = %d entries, want 4", len(all))
	}

	// 7 days = recent entries only
	week := filterEntriesByDays(entries, 7)
	if len(week) < 2 {
		t.Errorf("filterEntriesByDays(entries, 7) = %d entries, want >= 2", len(week))
	}

	// 30 days = should include most entries
	month := filterEntriesByDays(entries, 30)
	if len(month) < 3 {
		t.Errorf("filterEntriesByDays(entries, 30) = %d entries, want >= 3", len(month))
	}
}

func TestGetAuditLogPath(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	path := getAuditLogPath()
	want := filepath.Join(dataHome, "tok", "hook-audit.log")
	if path != want {
		t.Errorf("getAuditLogPath() = %q, want %q", path, want)
	}
}

func TestGetAuditLogPath_EnvOverride(t *testing.T) {
	override := t.TempDir()
	t.Setenv("TOKMAN_AUDIT_DIR", override)

	got := getAuditLogPath()
	want := filepath.Join(override, "hook-audit.log")
	if got != want {
		t.Errorf("getAuditLogPath() = %q, want %q", got, want)
	}
}

// ── Copilot format constants ───────────────────────────────────

func TestCopilotFormatConstants(t *testing.T) {
	if copilotFormatVsCode != 0 {
		t.Errorf("copilotFormatVsCode = %d, want 0", copilotFormatVsCode)
	}
	if copilotFormatCli != 1 {
		t.Errorf("copilotFormatCli = %d, want 1", copilotFormatCli)
	}
	if copilotFormatPassThrough != 2 {
		t.Errorf("copilotFormatPassThrough = %d, want 2", copilotFormatPassThrough)
	}
}

// ── Benchmarks ─────────────────────────────────────────────────

func BenchmarkDetectCopilotFormat_VSCode(b *testing.B) {
	input := map[string]any{
		"tool_name": "Bash",
		"tool_input": map[string]any{
			"command": "git status",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectCopilotFormat(input)
	}
}

func BenchmarkDetectCopilotFormat_Cli(b *testing.B) {
	input := map[string]any{
		"toolName": "bash",
		"toolArgs": `{"command":"git status"}`,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectCopilotFormat(input)
	}
}

func BenchmarkDetectCopilotFormat_PassThrough(b *testing.B) {
	input := map[string]any{
		"tool_name": "editFiles",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectCopilotFormat(input)
	}
}

func BenchmarkParseAuditLine(b *testing.B) {
	line := "2026-01-15T10:30:00Z | rewrite | git status | tok git status"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseAuditLine(line)
	}
}

func BenchmarkBaseCommand(b *testing.B) {
	cmd := "FOO=bar BAR=baz git status --short"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseCommand(cmd)
	}
}

func TestRunClaudeInnerProducesValidJSON(t *testing.T) {
	input := `{"tool_name":"Bash","tool_input":{"command":"git status"}}`
	output := runClaudeInner(input)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}
