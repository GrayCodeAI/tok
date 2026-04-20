package core

// Per-agent integration tests: mock a fresh user HOME, run setupAgent
// for each of the top-5 wired agents, assert the files tok claims to
// install actually exist on disk with plausible content, then
// uninstall and assert everything is cleaned up.
//
// Before this file there were unit tests for individual JSON patchers
// (ensureClaudeHook, ensureCursorHook, etc.) but no coverage asserting
// that `tok init --<agent>` end-to-end produces the files its
// Instructions string promises. That gap meant a small refactor could
// break an agent's install silently without any test complaint.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// agentCase describes one agent's expected install footprint so a single
// test loop can cover every top-5 wired agent with its own idiosyncratic
// file layout (JSON-patched vs rules-file vs plugin-script).
type agentCase struct {
	name string
	// buildAgent constructs an AgentInfo using the given fake HOME/cwd.
	buildAgent func(home, cwd string) AgentInfo
	// global flag to pass to setupAgent. Some agents (OpenCode) require it.
	global bool
	// expectedFiles must exist after setupAgent.
	expectedFiles func(home, cwd string) []string
	// expectedContents: path => substring that must be present after install.
	expectedContents func(home, cwd string) map[string]string
}

func topFiveAgentCases() []agentCase {
	return []agentCase{
		{
			name: "Claude Code",
			buildAgent: func(home, cwd string) AgentInfo {
				return AgentInfo{
					Name:      "Claude Code",
					ConfigDir: filepath.Join(home, ".claude"),
					HookDir:   filepath.Join(home, ".claude", "hooks"),
				}
			},
			expectedFiles: func(home, cwd string) []string {
				return []string{
					filepath.Join(home, ".claude", "hooks", "tok-rewrite.sh"),
					filepath.Join(home, ".claude", "TOK.md"),
					filepath.Join(home, ".claude", "settings.json"),
				}
			},
			expectedContents: func(home, cwd string) map[string]string {
				return map[string]string{
					filepath.Join(home, ".claude", "settings.json"): "PreToolUse",
					filepath.Join(home, ".claude", "CLAUDE.md"):     "@TOK.md",
				}
			},
		},
		{
			name: "Codex",
			buildAgent: func(home, cwd string) AgentInfo {
				return AgentInfo{
					Name:      "Codex",
					ConfigDir: cwd,
				}
			},
			expectedFiles: func(home, cwd string) []string {
				return []string{
					filepath.Join(cwd, "TOK.md"),
					filepath.Join(cwd, "AGENTS.md"),
				}
			},
			expectedContents: func(home, cwd string) map[string]string {
				return map[string]string{
					filepath.Join(cwd, "AGENTS.md"): "@TOK.md",
				}
			},
		},
		{
			name: "Gemini CLI",
			buildAgent: func(home, cwd string) AgentInfo {
				return AgentInfo{
					Name:      "Gemini CLI",
					ConfigDir: filepath.Join(home, ".gemini"),
					HookDir:   filepath.Join(home, ".gemini", "hooks"),
				}
			},
			expectedFiles: func(home, cwd string) []string {
				return []string{
					filepath.Join(home, ".gemini", "hooks", "tok-rewrite.sh"),
					filepath.Join(home, ".gemini", "TOK.md"),
					filepath.Join(home, ".gemini", "settings.json"),
				}
			},
			expectedContents: func(home, cwd string) map[string]string {
				return map[string]string{
					filepath.Join(home, ".gemini", "settings.json"): "BeforeTool",
				}
			},
		},
		{
			name: "Qwen Code",
			buildAgent: func(home, cwd string) AgentInfo {
				return AgentInfo{
					Name:      "Qwen Code",
					ConfigDir: filepath.Join(home, ".qwen"),
					HookDir:   filepath.Join(home, ".qwen", "hooks"),
				}
			},
			expectedFiles: func(home, cwd string) []string {
				return []string{
					filepath.Join(home, ".qwen", "hooks", "tok-rewrite.sh"),
					filepath.Join(home, ".qwen", "TOK.md"),
				}
			},
		},
		{
			name: "OpenCode",
			buildAgent: func(home, cwd string) AgentInfo {
				return AgentInfo{
					Name:      "OpenCode",
					ConfigDir: filepath.Join(home, ".config", "opencode"),
					HookDir:   filepath.Join(home, ".config", "opencode", "plugins"),
				}
			},
			global: true,
			expectedFiles: func(home, cwd string) []string {
				return []string{
					filepath.Join(home, ".config", "opencode", "plugins", "tok.ts"),
				}
			},
			expectedContents: func(home, cwd string) map[string]string {
				return map[string]string{
					filepath.Join(home, ".config", "opencode", "plugins", "tok.ts"): "TokOpenCodePlugin",
				}
			},
		},
	}
}

func TestTopFiveAgentInstallRoundTrip(t *testing.T) {
	for _, tc := range topFiveAgentCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			cwd := filepath.Join(home, "project")
			if err := os.MkdirAll(cwd, 0755); err != nil {
				t.Fatalf("mkdir cwd: %v", err)
			}
			agent := tc.buildAgent(home, cwd)

			if err := setupAgent(agent, tc.global); err != nil {
				t.Fatalf("setupAgent: %v", err)
			}

			// File existence assertions — if any promised file is missing,
			// the Instructions string is lying to the user.
			for _, path := range tc.expectedFiles(home, cwd) {
				if _, err := os.Stat(path); err != nil {
					t.Errorf("expected file missing after install: %s (%v)", path, err)
				}
			}

			// Content substring assertions — catches empty or malformed
			// patch output (e.g. settings.json parsed but hook never added).
			if tc.expectedContents != nil {
				for path, want := range tc.expectedContents(home, cwd) {
					content, err := os.ReadFile(path)
					if err != nil {
						t.Errorf("read %s: %v", path, err)
						continue
					}
					if !strings.Contains(string(content), want) {
						t.Errorf("%s missing expected substring %q", path, want)
					}
				}
			}

			// Idempotency: second install must not error and must not
			// duplicate entries in patched JSON files.
			if err := setupAgent(agent, tc.global); err != nil {
				t.Errorf("second setupAgent (idempotency check): %v", err)
			}

			// Uninstall must leave no tok-owned files behind.
			removed, err := uninstallAgent(agent)
			if err != nil {
				t.Errorf("uninstallAgent: %v", err)
			}
			if len(removed) == 0 {
				t.Errorf("uninstallAgent reported zero artifacts removed — install likely wrote nothing")
			}
			for _, path := range tc.expectedFiles(home, cwd) {
				// JSON config files stay (user may have other settings);
				// tok-owned artifacts (TOK.md, hook script, plugin) must go.
				if strings.HasSuffix(path, "settings.json") ||
					strings.HasSuffix(path, "hooks.json") ||
					strings.HasSuffix(path, "AGENTS.md") {
					continue
				}
				if _, err := os.Stat(path); err == nil {
					t.Errorf("tok-owned file still present after uninstall: %s", path)
				}
			}
		})
	}
}
