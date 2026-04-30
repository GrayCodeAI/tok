package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/discover"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Hook processors for AI coding agents",
	Long: `Native hook processors that read JSON from stdin and output
rewritten commands for various AI coding agent platforms.

Supported platforms:
  claude   — Claude Code PreToolUse hook
  cursor   — Cursor preToolUse hook
  gemini   — Gemini CLI BeforeTool hook
  copilot  — GitHub Copilot (VS Code Chat + Copilot CLI)`,
}

func init() {
	registry.Add(func() { registry.Register(hookCmd) })
	hookCmd.AddCommand(hookClaudeCmd)
	hookCmd.AddCommand(hookCursorCmd)
	hookCmd.AddCommand(hookGeminiCmd)
	hookCmd.AddCommand(hookCopilotCmd)
}

var hookClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Process Claude Code PreToolUse hook",
	Run: func(cmd *cobra.Command, args []string) {
		runClaudeHook()
	},
}

var hookCursorCmd = &cobra.Command{
	Use:   "cursor",
	Short: "Process Cursor preToolUse hook",
	Run: func(cmd *cobra.Command, args []string) {
		runCursorHook()
	},
}

var hookGeminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Process Gemini CLI BeforeTool hook",
	Long: `Reads JSON from stdin (Gemini CLI hook format), rewrites
run_shell_command tool calls to tok equivalents, and outputs
Gemini CLI JSON format to stdout.

Used as a Gemini CLI BeforeTool hook — install with: tok init -g --gemini`,
	Run: func(cmd *cobra.Command, args []string) {
		runGeminiHook()
	},
}

type copilotHookFormat int

const (
	copilotFormatVsCode copilotHookFormat = iota
	copilotFormatCli
	copilotFormatPassThrough
)

var hookCopilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Process Copilot preToolUse hook",
	Long: `Reads JSON from stdin, auto-detects VS Code Copilot Chat format
(snake_case) vs Copilot CLI format (camelCase), and outputs the
appropriate response.

Used as a Copilot preToolUse hook — install with: tok init -g --copilot`,
	Run: func(cmd *cobra.Command, args []string) {
		runCopilotHook()
	},
}

func runClaudeHook() {
	input, ok := readHookJSON()
	if !ok {
		return
	}

	output, action, rewritten, ok := processClaudePayload(input)
	if !ok {
		return
	}

	recordHookAudit(action, hookCommandFromPayload(input), rewritten)
	writeHookJSON(output)
}

func runCursorHook() {
	input, ok := readHookJSON()
	if !ok {
		out.Global().Println("{}")
		return
	}

	output, action, rewritten := processCursorPayload(input)
	recordHookAudit(action, hookCommandFromPayload(input), rewritten)
	if output == nil {
		out.Global().Println("{}")
		return
	}
	writeHookJSON(output)
}

func runGeminiHook() {
	input, ok := readHookJSON()
	if !ok {
		printGeminiAllow()
		return
	}

	output, action, original, rewritten := processGeminiPayload(input)
	recordHookAudit(action, original, rewritten)
	if output == nil {
		printGeminiAllow()
		return
	}
	writeHookJSON(output)
}

func runCopilotHook() {
	input, ok := readHookJSON()
	if !ok {
		return
	}

	format, command := detectCopilotFormat(input)
	switch format {
	case copilotFormatVsCode:
		output, action, rewritten, ok := processVSCodeCommand(command)
		if !ok {
			return
		}
		recordHookAudit(action, command, rewritten)
		writeHookJSON(output)
	case copilotFormatCli:
		output, action, rewritten := processCopilotCLICommand(command)
		recordHookAudit(action, command, rewritten)
		if output == nil {
			out.Global().Println("{}")
			return
		}
		writeHookJSON(output)
	}
}

func readHookJSON() (map[string]any, bool) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, false
	}
	inputStr := strings.TrimSpace(string(input))
	if inputStr == "" {
		return nil, false
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(inputStr), &payload); err != nil {
		out.Global().Errorf("[tok hook] Failed to parse JSON input: %v", err)
		return nil, false
	}
	return payload, true
}

func detectCopilotFormat(v map[string]any) (copilotHookFormat, string) {
	if toolName, ok := v["tool_name"].(string); ok {
		switch toolName {
		case "runTerminalCommand", "Bash", "bash":
			if toolInput, ok := v["tool_input"].(map[string]any); ok {
				if cmd, ok := toolInput["command"].(string); ok && cmd != "" {
					return copilotFormatVsCode, cmd
				}
			}
		}
		return copilotFormatPassThrough, ""
	}

	if toolName, ok := v["toolName"].(string); ok && toolName == "bash" {
		if toolArgsStr, ok := v["toolArgs"].(string); ok {
			var toolArgs map[string]any
			if err := json.Unmarshal([]byte(toolArgsStr), &toolArgs); err == nil {
				if cmd, ok := toolArgs["command"].(string); ok && cmd != "" {
					return copilotFormatCli, cmd
				}
			}
		}
		return copilotFormatPassThrough, ""
	}

	return copilotFormatPassThrough, ""
}

func processClaudePayload(payload map[string]any) (map[string]any, string, string, bool) {
	command := hookCommandFromPayload(payload)
	if command == "" {
		return nil, "", "", false
	}
	return processVSCodeCommand(command)
}

func processCursorPayload(payload map[string]any) (map[string]any, string, string) {
	command := hookCommandFromPayload(payload)
	if command == "" {
		return nil, "skip:ignored", ""
	}

	decision := checkCommandPermissions(command)
	if decision == PermissionDeny {
		return nil, "skip:deny_rule", ""
	}

	rewritten, changed := rewriteHookCommand(command)
	if !changed {
		return nil, "skip:no_match", ""
	}

	permission := "ask"
	if decision == PermissionAllow {
		permission = "allow"
	}
	return map[string]any{
		"permission": permission,
		"updated_input": map[string]any{
			"command": rewritten,
		},
	}, "rewrite", rewritten
}

func processGeminiPayload(payload map[string]any) (map[string]any, string, string, string) {
	toolName, _ := payload["tool_name"].(string)
	if toolName != "run_shell_command" {
		return nil, "skip:ignored", "", ""
	}

	command := hookCommandFromPayload(payload)
	if command == "" {
		return nil, "skip:ignored", "", ""
	}

	if checkCommandPermissions(command) == PermissionDeny {
		return map[string]any{
			"decision": "deny",
			"reason":   "Blocked by tok permission rule",
		}, "skip:deny_rule", command, ""
	}

	rewritten, changed := rewriteHookCommand(command)
	if !changed {
		return nil, "skip:no_match", command, ""
	}

	return map[string]any{
		"decision": "allow",
		"hookSpecificOutput": map[string]any{
			"tool_input": map[string]any{
				"command": rewritten,
			},
		},
	}, "rewrite", command, rewritten
}

func processVSCodeCommand(command string) (map[string]any, string, string, bool) {
	decision := checkCommandPermissions(command)
	if decision == PermissionDeny {
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": "Command denied by tok permission rules",
			},
		}, "skip:deny_rule", "", true
	}

	rewritten, changed := rewriteHookCommand(command)
	if !changed {
		return nil, "skip:no_match", "", false
	}

	return buildCopilotVSCodeResponse(command, rewritten), "rewrite", rewritten, true
}

func processCopilotCLICommand(command string) (map[string]any, string, string) {
	if checkCommandPermissions(command) == PermissionDeny {
		return map[string]any{
			"permissionDecision":       "deny",
			"permissionDecisionReason": "Blocked by tok permission rule",
		}, "skip:deny_rule", ""
	}

	rewritten, changed := rewriteHookCommand(command)
	if !changed {
		return nil, "skip:no_match", ""
	}

	return map[string]any{
		"permissionDecision":       "deny",
		"permissionDecisionReason": fmt.Sprintf("Token savings: use `%s` instead (tok saves 60-90%% tokens)", rewritten),
	}, "rewrite", rewritten
}

func rewriteHookCommand(command string) (string, bool) {
	if isHookExcluded(command) {
		return "", false
	}
	rewritten, changed := discover.RewriteCommand(command, nil)
	if !changed || rewritten == command {
		return "", false
	}
	return rewritten, true
}

func isHookExcluded(command string) bool {
	cfg, err := config.Load("")
	if err != nil || cfg == nil {
		return false
	}
	trimmed := strings.TrimSpace(command)
	for _, excluded := range cfg.Hooks.ExcludedCommands {
		excluded = strings.TrimSpace(excluded)
		if excluded == "" {
			continue
		}
		if trimmed == excluded || strings.HasPrefix(trimmed, excluded+" ") {
			return true
		}
	}
	return false
}

func hookCommandFromPayload(payload map[string]any) string {
	toolInput, _ := payload["tool_input"].(map[string]any)
	command, _ := toolInput["command"].(string)
	return strings.TrimSpace(command)
}

func buildCopilotVSCodeResponse(originalCmd, rewritten string) map[string]any {
	decision := checkCommandPermissions(originalCmd)
	reason := "tok auto-rewrite"

	switch decision {
	case PermissionDeny:
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": "Command denied by tok permission rules",
			},
		}
	case PermissionAsk, PermissionDefault:
		if decision == PermissionDefault {
			reason = "tok rewrite prepared; confirm command under default least-privilege policy"
		} else {
			reason = "tok rewrite prepared; command matches ask rule"
		}
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "ask",
				"permissionDecisionReason": reason,
				"updatedInput": map[string]any{
					"command": rewritten,
				},
			},
		}
	default:
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "allow",
				"permissionDecisionReason": reason,
				"updatedInput": map[string]any{
					"command": rewritten,
				},
			},
		}
	}
}

func printGeminiAllow() {
	out.Global().Println(`{"decision":"allow"}`)
}

func writeHookJSON(output map[string]any) {
	data, err := json.Marshal(output)
	if err != nil {
		out.Global().Errorf("[tok hook] Failed to marshal output: %v\n", err)
		return
	}
	out.Global().Println(string(data))
}

func recordHookAudit(action, originalCmd, rewrittenCmd string) {
	auditEnabled := strings.EqualFold(os.Getenv("TOK_HOOK_AUDIT"), "1") ||
		strings.EqualFold(os.Getenv("TOK_HOOK_AUDIT"), "true")
	if !auditEnabled || action == "" || originalCmd == "" {
		return
	}

	logPath := getAuditLogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	line := fmt.Sprintf(
		"%s | %s | %s | %s\n",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		sanitizeAuditField(action),
		sanitizeAuditField(originalCmd),
		sanitizeAuditField(rewrittenCmd),
	)
	_, _ = file.WriteString(line)
}

// secretPatterns matches common credential patterns in command strings and output.
// This is a best-effort heuristic; it does not guarantee complete secret scrubbing.
var secretPatterns = []struct {
	re      *regexp.Regexp
	replace string
}{
	// API keys / tokens
	{regexp.MustCompile(`\b(sk-[a-zA-Z0-9]{20,})\b`), `[REDACTED_SK]`},
	{regexp.MustCompile(`\b(ghp_[a-zA-Z0-9]{36,})\b`), `[REDACTED_GH]`},
	{regexp.MustCompile(`\b(glpat-[a-zA-Z0-9\-]{20,})\b`), `[REDACTED_GL]`},
	{regexp.MustCompile(`\b(AKIA[0-9A-Z]{16})\b`), `[REDACTED_AWS_AK]`},
	// Passwords / secrets in key=value forms
	{regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api_key|apikey)\s*[=:]\s*\S+`), `[REDACTED]`},
	// Authorization headers
	{regexp.MustCompile(`(?i)(authorization|x-api-key)\s*[:=]\s*\S+`), `[REDACTED]`},
	// Private keys
	{regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----[\s\S]*?-----END (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), `[REDACTED_KEY]`},
}

func sanitizeAuditField(value string) string {
	v := strings.NewReplacer("\\", "\\\\", "|", "\\|", "\n", "\\n", "\r", "\\r").Replace(value)
	for _, p := range secretPatterns {
		v = p.re.ReplaceAllString(v, p.replace)
	}
	return v
}

func runClaudeInner(input string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return ""
	}
	output, _, _, ok := processClaudePayload(payload)
	if !ok || output == nil {
		return ""
	}
	data, err := json.Marshal(output)
	if err != nil {
		return ""
	}
	return string(data)
}

func runCursorInner(input string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return "{}"
	}
	output, _, _ := processCursorPayload(payload)
	if output == nil {
		return "{}"
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "{}"
	}
	return string(data)
}
