package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type HookType string

const (
	HookTypeClaude   HookType = "claude"
	HookTypeCursor   HookType = "cursor"
	HookTypeCopilot  HookType = "copilot"
	HookTypeGemini   HookType = "gemini"
	HookTypeCodex    HookType = "codex"
	HookTypeWindsurf HookType = "windsurf"
	HookTypeCline    HookType = "cline"
	HookTypeOpencode HookType = "opencode"
	HookTypeAider    HookType = "aider"
	HookTypeContinue HookType = "continue"
	HookTypeReplit   HookType = "replit"
)

type ShellType string

const (
	ShellBash ShellType = "bash"
	ShellZsh  ShellType = "zsh"
	ShellFish ShellType = "fish"
)

type HookConfig struct {
	Type        HookType
	Shell       ShellType
	InstallPath string
	AutoInstall bool
}

type HookEngine struct {
	mu      sync.RWMutex
	configs map[HookType]*HookConfig
	stats   HookStats
}

type HookStats struct {
	Installs   int64
	Uninstalls int64
	Updates    int64
	LastCheck  time.Time
}

func NewHookEngine() *HookEngine {
	return &HookEngine{
		configs: map[HookType]*HookConfig{
			HookTypeClaude:   {Type: HookTypeClaude, Shell: ShellBash, InstallPath: "~/.claude/settings.json"},
			HookTypeCursor:   {Type: HookTypeCursor, Shell: ShellBash, InstallPath: "~/.cursor/settings.json"},
			HookTypeCopilot:  {Type: HookTypeCopilot, Shell: ShellBash, InstallPath: "~/.github-copilot/settings.json"},
			HookTypeGemini:   {Type: HookTypeGemini, Shell: ShellBash, InstallPath: "~/.gemini/settings.json"},
			HookTypeCodex:    {Type: HookTypeCodex, Shell: ShellBash, InstallPath: "~/.codex/settings.json"},
			HookTypeWindsurf: {Type: HookTypeWindsurf, Shell: ShellBash, InstallPath: "~/.windsurf/settings.json"},
			HookTypeCline:    {Type: HookTypeCline, Shell: ShellBash, InstallPath: "~/.cline/settings.json"},
			HookTypeOpencode: {Type: HookTypeOpencode, Shell: ShellBash, InstallPath: "~/.opencode/settings.json"},
			HookTypeAider:    {Type: HookTypeAider, Shell: ShellBash, InstallPath: "~/.aider/settings.json"},
			HookTypeContinue: {Type: HookTypeContinue, Shell: ShellBash, InstallPath: "~/.continue/settings.json"},
			HookTypeReplit:   {Type: HookTypeReplit, Shell: ShellBash, InstallPath: "~/.replit/settings.json"},
		},
		stats: HookStats{},
	}
}

func (e *HookEngine) Install(ctx context.Context, hookType HookType) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	config, ok := e.configs[hookType]
	if !ok {
		return fmt.Errorf("unknown hook type: %s", hookType)
	}

	if config.Shell == "" {
		return fmt.Errorf("no shell configured for %s", hookType)
	}

	hookScript, err := e.generateHookScript(hookType, config.Shell)
	if err != nil {
		return fmt.Errorf("generating hook script: %w", err)
	}

	installPath := expandPath(config.InstallPath)
	dir := filepath.Dir(installPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(installPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("writing hook file: %w", err)
	}

	e.stats.Installs++
	return nil
}

func (e *HookEngine) Uninstall(ctx context.Context, hookType HookType) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	config, ok := e.configs[hookType]
	if !ok {
		return fmt.Errorf("unknown hook type: %s", hookType)
	}

	installPath := expandPath(config.InstallPath)
	if _, err := os.Stat(installPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.Remove(installPath); err != nil {
		return fmt.Errorf("removing hook file: %w", err)
	}

	e.stats.Uninstalls++
	return nil
}

func (e *HookEngine) CheckIntegrity(ctx context.Context, hookType HookType) (bool, error) {
	config, ok := e.configs[hookType]
	if !ok {
		return false, fmt.Errorf("unknown hook type: %s", hookType)
	}

	installPath := expandPath(config.InstallPath)
	data, err := os.ReadFile(installPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	expectedScript, _ := e.generateHookScript(hookType, config.Shell)
	return string(data) == expectedScript, nil
}

func (e *HookEngine) Update(ctx context.Context, hookType HookType) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	config, ok := e.configs[hookType]
	if !ok {
		return fmt.Errorf("unknown hook type: %s", hookType)
	}

	hookScript, err := e.generateHookScript(hookType, config.Shell)
	if err != nil {
		return fmt.Errorf("generating hook script: %w", err)
	}

	installPath := expandPath(config.InstallPath)
	if err := os.WriteFile(installPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("updating hook file: %w", err)
	}

	e.stats.Updates++
	return nil
}

func (e *HookEngine) generateHookScript(hookType HookType, shell ShellType) (string, error) {
	switch shell {
	case ShellBash:
		return generateBashHook(hookType)
	case ShellZsh:
		return generateZshHook(hookType)
	case ShellFish:
		return generateFishHook(hookType)
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func generateBashHook(hookType HookType) (string, error) {
	return `#!/bin/bash
# TokMan Hook - Auto-generated for ` + string(hookType) + `
# Version: 2.0

if ! command -v tokman &>/dev/null; then
    return 0 2>/dev/null || exit 0
fi

__tokman_rewrite() {
    local cmd="$1"
    tokman rewrite "$cmd" 2>/dev/null
}

__tokman_hook_init() {
    if [[ -n "$BASH_VERSION" ]]; then
        export PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND;} __tokman_preexec"
    fi
}

__tokman_preexec() {
    local last_cmd="$BASH_COMMAND"
    local rewritten
    rewritten=$(__tokman_rewrite "$last_cmd")
    if [[ -n "$rewritten" && "$rewritten" != "$last_cmd" ]]; then
        eval "$rewritten"
    fi
}

__tokman_hook_init
`, nil
}

func generateZshHook(hookType HookType) (string, error) {
	return `#!/bin/zsh
# TokMan Hook - Auto-generated for ` + string(hookType) + `
# Version: 2.0

if ! command -v tokman &>/dev/null; then
    return 0 2>/dev/null || exit 0
fi

__tokman_rewrite() {
    local cmd="$1"
    tokman rewrite "$cmd" 2>/dev/null
}

__tokman_preexec() {
    local last_cmd="$BUFFER"
    local rewritten
    rewritten=$(__tokman_rewrite "$last_cmd")
    if [[ -n "$rewritten" && "$rewritten" != "$last_cmd" ]]; then
        BUFFER="$rewritten"
    fi
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec __tokman_preexec
`, nil
}

func generateFishHook(hookType HookType) (string, error) {
	return `#!/usr/bin/env fish
# TokMan Hook - Auto-generated for ` + string(hookType) + `
# Version: 2.0

if not command -v tokman &>/dev/null
    exit 0
end

function __tokman_rewrite
    set -l cmd $argv[1]
    tokman rewrite "$cmd" 2>/dev/null
end

function __tokman_preexec
    set -l cmd (commandline)
    set -l rewritten (__tokman_rewrite "$cmd")
    if test -n "$rewritten" && test "$rewritten" != "$cmd"
        commandline -r "$rewritten"
    end
end

if not functions -q __tokman_preexec
    functions -c fish_preexec __tokman_preexec
end
`, nil
}

func (e *HookEngine) DetectConflicts(ctx context.Context) []string {
	conflicts := make([]string, 0)

	hooksDirs := []string{
		"~/.claude",
		"~/.cursor",
		"~/.github-copilot",
		"~/.gemini",
		"~/.codex",
		"~/.windsurf",
		"~/.cline",
		"~/.opencode",
		"~/.aider",
		"~/.continue",
		"~/.replit",
	}

	for _, dir := range hooksDirs {
		expanded := expandPath(dir)
		entries, err := os.ReadDir(expanded)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "tokman") || strings.HasPrefix(entry.Name(), ".tokman") {
				conflicts = append(conflicts, filepath.Join(dir, entry.Name()))
			}
		}
	}

	return conflicts
}

func (e *HookEngine) SetShell(hookType HookType, shell ShellType) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if config, ok := e.configs[hookType]; ok {
		config.Shell = shell
	}
}

func (e *HookEngine) GetStats() HookStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

type HookTestingFramework struct {
	mu      sync.Mutex
	tests   []HookTest
	results map[string]*HookTestResult
}

type HookTest struct {
	Name        string
	HookType    HookType
	InputCmd    string
	ExpectedCmd string
}

type HookTestResult struct {
	TestName  string
	Passed    bool
	ActualCmd string
	Errors    []string
	Duration  time.Duration
}

func NewHookTestingFramework() *HookTestingFramework {
	return &HookTestingFramework{
		tests:   make([]HookTest, 0),
		results: make(map[string]*HookTestResult),
	}
}

func (tf *HookTestingFramework) AddTest(test HookTest) {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	tf.tests = append(tf.tests, test)
}

func (tf *HookTestingFramework) RunTests(ctx context.Context, engine *HookEngine) error {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	for _, test := range tf.tests {
		start := time.Now()

		result := &HookTestResult{
			TestName: test.Name,
			Passed:   false,
		}

		rewritten, err := exec.CommandContext(ctx, "tokman", "rewrite", test.InputCmd).Output()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("exec error: %v", err))
		} else {
			result.ActualCmd = strings.TrimSpace(string(rewritten))
			if result.ActualCmd == test.ExpectedCmd {
				result.Passed = true
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("expected %q, got %q", test.ExpectedCmd, result.ActualCmd))
			}
		}

		result.Duration = time.Since(start)
		tf.results[test.Name] = result
	}

	return nil
}

func (tf *HookTestingFramework) GetResults() map[string]*HookTestResult {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	return tf.results
}

func (tf *HookTestingFramework) Summary() (passed, failed int) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	for _, result := range tf.results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}

func (e *HookEngine) AutoInstallShellHook(ctx context.Context, shell ShellType) error {
	var shellConfig string

	switch shell {
	case ShellBash:
		shellConfig = "~/.bashrc"
	case ShellZsh:
		shellConfig = "~/.zshrc"
	case ShellFish:
		shellConfig = "~/.config/fish/config.fish"
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	expanded := expandPath(shellConfig)

	marker := "# TokMan Shell Hook"
	hookLine := "[ -f ~/.tokman/tokman-hook.sh ] && source ~/.tokman/tokman-hook.sh"

	data, err := os.ReadFile(expanded)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading shell config: %w", err)
	}

	if os.IsNotExist(err) || !strings.Contains(string(data), marker) {
		f, err := os.OpenFile(expanded, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening shell config: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString("\n" + marker + "\n"); err != nil {
			return fmt.Errorf("writing marker: %w", err)
		}
		if _, err := f.WriteString(hookLine + "\n"); err != nil {
			return fmt.Errorf("writing hook line: %w", err)
		}
	}

	hookDir := expandPath("~/.tokman")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("creating hook dir: %w", err)
	}

	hookScript, _ := generateBashHook(HookTypeClaude)
	hookPath := filepath.Join(hookDir, "tokman-hook.sh")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("writing shell hook: %w", err)
	}

	return nil
}

func (e *HookEngine) GetHookInfo(hookType HookType) (HookConfig, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	config, ok := e.configs[hookType]
	if !ok {
		return HookConfig{}, false
	}

	return *config, true
}
