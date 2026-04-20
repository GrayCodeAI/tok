package core

import (
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/integrity"
)

var (
	initGlobal      bool
	initClaude      bool
	initCursor      bool
	initWindsurf    bool
	initCline       bool
	initGemini      bool
	initCodex       bool
	initCopilot     bool
	initOpencode    bool
	initOpenclaw    bool
	initKilocode    bool
	initAntigravity bool
	initAll         bool
	initShow        bool
	initUninstall   bool
	initAgent       string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tok for AI agents",
	Long: `Set up tok integration with AI coding assistants.

Installs hooks and configuration for various AI agents to enable
token compression and optimization.

Supported agents:
  --claude         Claude Code (default)
  --cursor         Cursor IDE
  --windsurf       Windsurf IDE
  --cline          Cline / Roo Code
  --gemini         Google Gemini CLI
  --codex          OpenAI Codex CLI
  --copilot        GitHub Copilot
  --opencode       OpenCode
  --openclaw       OpenClaw
  --kilocode       Kilo Code
  --antigravity    Google Antigravity
  --all            All detected agents

Examples:
  tok init                    # Interactive setup (detects agents)
  tok init --claude           # Setup for Claude Code only
  tok init --cursor --windsurf # Setup for multiple agents
  tok init --all              # Setup for all detected agents
  tok init --global --all     # Global installation for all agents
  tok init --claude --uninstall # Remove Claude integration`,
	RunE: runInit,
}

func init() {
	registry.Add(func() { registry.Register(initCmd) })

	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "Install to global agent config directory")
	initCmd.Flags().BoolVar(&initClaude, "claude", false, "Setup for Claude Code")
	initCmd.Flags().BoolVar(&initCursor, "cursor", false, "Setup for Cursor")
	initCmd.Flags().BoolVar(&initWindsurf, "windsurf", false, "Setup for Windsurf")
	initCmd.Flags().BoolVar(&initCline, "cline", false, "Setup for Cline")
	initCmd.Flags().BoolVar(&initGemini, "gemini", false, "Setup for Gemini CLI")
	initCmd.Flags().BoolVar(&initCodex, "codex", false, "Setup for Codex")
	initCmd.Flags().BoolVar(&initCopilot, "copilot", false, "Setup for GitHub Copilot")
	initCmd.Flags().BoolVar(&initOpencode, "opencode", false, "Setup for OpenCode")
	initCmd.Flags().BoolVar(&initOpenclaw, "openclaw", false, "Setup for OpenClaw")
	initCmd.Flags().BoolVar(&initKilocode, "kilocode", false, "Setup for Kilo Code")
	initCmd.Flags().BoolVar(&initAntigravity, "antigravity", false, "Setup for Google Antigravity")
	initCmd.Flags().BoolVarP(&initAll, "all", "a", false, "Setup for all detected agents")
	initCmd.Flags().StringVar(&initAgent, "agent", "", "Setup for specific agent by name (claude-code, copilot, cursor, windsurf, cline, roo-code, codex, gemini, kilocode, antigravity, opencode, openclaw)")
	initCmd.Flags().BoolVar(&initShow, "show", false, "Show current configuration")
	initCmd.Flags().BoolVar(&initUninstall, "uninstall", false, "Remove tok integration for selected agents")
}

// AgentInfo holds information about an AI agent
type AgentInfo struct {
	Name         string
	Flag         *bool
	DetectDir    string
	ConfigDir    string
	HookDir      string
	Detected     bool
	Instructions string
}

// agentNameMap maps CLI agent names to AgentInfo names
var agentNameMap = map[string]string{
	"claude-code": "Claude Code",
	"claude":      "Claude Code",
	"copilot":     "GitHub Copilot",
	"cursor":      "Cursor",
	"windsurf":    "Windsurf",
	"cline":       "Cline",
	"roo-code":    "Cline",
	"codex":       "Codex",
	"gemini":      "Gemini CLI",
	"kilocode":    "Kilo Code",
	"antigravity": "Google Antigravity",
	"opencode":    "OpenCode",
	"openclaw":    "OpenClaw",
}

// SupportedAgents lists all supported agent names
var SupportedAgents = []string{
	"claude-code", "copilot", "cursor", "windsurf", "cline",
	"roo-code", "codex", "gemini", "kilocode", "antigravity",
	"opencode", "openclaw",
}

func currentAgentInfos(global bool) ([]AgentInfo, error) {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil, fmt.Errorf("cannot determine user home directory")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot determine current directory: %w", err)
	}

	codexDir := cwd
	if global {
		codexDir = resolveCodexConfigDir(home)
	}

	agents := []AgentInfo{
		{
			Name:         "Claude Code",
			Flag:         &initClaude,
			DetectDir:    filepath.Join(home, ".claude"),
			ConfigDir:    filepath.Join(home, ".claude"),
			HookDir:      filepath.Join(home, ".claude", "hooks"),
			Instructions: "Patched ~/.claude/settings.json and ensured ~/.claude/CLAUDE.md references @TOK.md",
		},
		{
			Name:         "Cursor",
			Flag:         &initCursor,
			DetectDir:    filepath.Join(home, ".cursor"),
			ConfigDir:    filepath.Join(home, ".cursor"),
			HookDir:      filepath.Join(home, ".cursor", "hooks"),
			Instructions: "Patched ~/.cursor/hooks.json with a preToolUse hook",
		},
		{
			Name:         "Windsurf",
			Flag:         &initWindsurf,
			DetectDir:    filepath.Join(home, ".windsurf"),
			ConfigDir:    cwd,
			Instructions: "Patched project .windsurfrules with tok instructions",
		},
		{
			Name:         "Cline",
			Flag:         &initCline,
			DetectDir:    filepath.Join(home, ".cline"),
			ConfigDir:    cwd,
			Instructions: "Patched project .clinerules with tok instructions",
		},
		{
			Name:         "Gemini CLI",
			Flag:         &initGemini,
			DetectDir:    filepath.Join(home, ".gemini"),
			ConfigDir:    filepath.Join(home, ".gemini"),
			HookDir:      filepath.Join(home, ".gemini", "hooks"),
			Instructions: "Patched ~/.gemini/settings.json with a BeforeTool hook",
		},
		{
			Name:         "Codex",
			Flag:         &initCodex,
			DetectDir:    resolveCodexConfigDir(home),
			ConfigDir:    codexDir,
			Instructions: "Patched AGENTS.md with a tok instructions reference",
		},
		{
			Name:         "GitHub Copilot",
			Flag:         &initCopilot,
			DetectDir:    filepath.Join(cwd, ".github"),
			ConfigDir:    cwd,
			Instructions: "Installed .github/hooks/tok-rewrite.json and .github/copilot-instructions.md",
		},
		{
			Name:         "OpenCode",
			Flag:         &initOpencode,
			DetectDir:    filepath.Join(home, ".config", "opencode"),
			ConfigDir:    filepath.Join(home, ".config", "opencode"),
			HookDir:      filepath.Join(home, ".config", "opencode", "plugins"),
			Instructions: "Installed ~/.config/opencode/plugins/tok.ts",
		},
		{
			Name:         "OpenClaw",
			Flag:         &initOpenclaw,
			DetectDir:    filepath.Join(home, ".openclaw"),
			ConfigDir:    filepath.Join(home, ".openclaw"),
			HookDir:      filepath.Join(home, ".openclaw", "hooks"),
			Instructions: "Add to ~/.openclaw/config.json",
		},
		{
			Name:         "Kilo Code",
			Flag:         &initKilocode,
			DetectDir:    filepath.Join(home, ".kilocode"),
			ConfigDir:    cwd,
			Instructions: "Installed .kilocode/rules/tok-rules.md",
		},
		{
			Name:         "Google Antigravity",
			Flag:         &initAntigravity,
			DetectDir:    filepath.Join(home, ".antigravity"),
			ConfigDir:    cwd,
			Instructions: "Installed .agents/rules/antigravity-tok-rules.md",
		},
	}

	for i := range agents {
		detectDir := agents[i].DetectDir
		if detectDir == "" {
			detectDir = agents[i].ConfigDir
		}
		if _, err := os.Stat(detectDir); err == nil {
			agents[i].Detected = true
		}
	}

	return agents, nil
}

// createAgentInfoByName creates an AgentInfo struct for a given agent name
func createAgentInfoByName(name string) *AgentInfo {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}
	cwd, _ := os.Getwd()

	switch name {
	case "Claude Code":
		return &AgentInfo{
			Name:         "Claude Code",
			DetectDir:    filepath.Join(home, ".claude"),
			ConfigDir:    filepath.Join(home, ".claude"),
			HookDir:      filepath.Join(home, ".claude", "hooks"),
			Instructions: "Installed transparent rewrite hook via DEBUG trap",
		}
	case "Cursor":
		return &AgentInfo{
			Name:         "Cursor",
			DetectDir:    filepath.Join(home, ".cursor"),
			ConfigDir:    filepath.Join(home, ".cursor"),
			HookDir:      filepath.Join(home, ".cursor", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Windsurf":
		return &AgentInfo{
			Name:         "Windsurf",
			DetectDir:    filepath.Join(home, ".windsurf"),
			ConfigDir:    cwd,
			HookDir:      filepath.Join(home, ".windsurf", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Cline":
		return &AgentInfo{
			Name:         "Cline",
			DetectDir:    filepath.Join(home, ".cline"),
			ConfigDir:    cwd,
			HookDir:      filepath.Join(home, ".cline", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Gemini CLI":
		return &AgentInfo{
			Name:         "Gemini CLI",
			DetectDir:    filepath.Join(home, ".gemini"),
			ConfigDir:    filepath.Join(home, ".gemini"),
			HookDir:      filepath.Join(home, ".gemini", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Codex":
		return &AgentInfo{
			Name:         "Codex",
			DetectDir:    resolveCodexConfigDir(home),
			ConfigDir:    resolveCodexConfigDir(home),
			HookDir:      filepath.Join(resolveCodexConfigDir(home), "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "GitHub Copilot":
		return &AgentInfo{
			Name:         "GitHub Copilot",
			DetectDir:    filepath.Join(cwd, ".github"),
			ConfigDir:    cwd,
			HookDir:      filepath.Join(cwd, ".github", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "OpenCode":
		return &AgentInfo{
			Name:         "OpenCode",
			DetectDir:    filepath.Join(home, ".config", "opencode"),
			ConfigDir:    filepath.Join(home, ".config", "opencode"),
			HookDir:      filepath.Join(home, ".config", "opencode", "plugins"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "OpenClaw":
		return &AgentInfo{
			Name:         "OpenClaw",
			DetectDir:    filepath.Join(home, ".openclaw"),
			ConfigDir:    filepath.Join(home, ".openclaw"),
			HookDir:      filepath.Join(home, ".openclaw", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Kilo Code":
		return &AgentInfo{
			Name:         "Kilo Code",
			DetectDir:    filepath.Join(home, ".kilocode"),
			ConfigDir:    cwd,
			HookDir:      filepath.Join(home, ".kilocode", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	case "Google Antigravity":
		return &AgentInfo{
			Name:         "Google Antigravity",
			DetectDir:    filepath.Join(home, ".antigravity"),
			ConfigDir:    cwd,
			HookDir:      filepath.Join(home, ".antigravity", "hooks"),
			Instructions: "Installed transparent rewrite hook for bash tool calls",
		}
	default:
		return nil
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	if initShow {
		return showInitConfig()
	}

	agents, err := currentAgentInfos(initGlobal)
	if err != nil {
		return err
	}

	// Determine which agents to setup
	var toSetup []AgentInfo

	if initAgent != "" {
		// Setup specific agent by name
		agentName, ok := agentNameMap[initAgent]
		if !ok {
			out.Global().Printf("Unknown agent: %s\n", initAgent)
			out.Global().Println("\nSupported agents:")
			for _, name := range SupportedAgents {
				out.Global().Printf("  %s\n", name)
			}
			return fmt.Errorf("unknown agent: %s", initAgent)
		}

		found := false
		for _, agent := range agents {
			if agent.Name == agentName {
				toSetup = append(toSetup, agent)
				found = true
				break
			}
		}

		if !found {
			// Create agent info for agents not in currentAgentInfos
			agent := createAgentInfoByName(agentName)
			if agent != nil {
				toSetup = append(toSetup, *agent)
			} else {
				return fmt.Errorf("could not create agent info for: %s", agentName)
			}
		}
	} else if initAll {
		// Setup all detected agents
		for _, agent := range agents {
			if agent.Detected {
				toSetup = append(toSetup, agent)
			}
		}
	} else {
		// Check if any specific agent flags were set
		anyFlag := initClaude || initCursor || initWindsurf || initCline ||
			initGemini || initCodex || initCopilot || initOpencode ||
			initOpenclaw || initKilocode || initAntigravity

		if !anyFlag {
			// Interactive mode - detect and ask
			return runInteractiveInit(agents)
		}

		// Setup selected agents
		for _, agent := range agents {
			if *agent.Flag {
				toSetup = append(toSetup, agent)
			}
		}
	}

	if len(toSetup) == 0 {
		out.Global().Println("No agents selected or detected.")
		out.Global().Println("\nTo setup a specific agent, use:")
		out.Global().Println("  tok init --claude     # For Claude Code")
		out.Global().Println("  tok init --cursor     # For Cursor")
		out.Global().Println("  tok init --windsurf   # For Windsurf")
		out.Global().Println("  tok init --opencode   # For OpenCode")
		out.Global().Println("  tok init --openclaw   # For OpenClaw")
		out.Global().Println("  tok init --kilocode   # For Kilo Code")
		out.Global().Println("  tok init --antigravity # For Google Antigravity")
		out.Global().Println("\nOr detect all installed agents:")
		out.Global().Println("  tok init --all")
		return nil
	}

	if initUninstall {
		return runInitUninstall(toSetup)
	}

	// Setup each agent
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	out.Global().Println()
	out.Global().Println("Setting up tok for AI agents...")
	out.Global().Println()

	for _, agent := range toSetup {
		out.Global().Printf("📦 %s\n", agent.Name)

		if err := setupAgent(agent, installUsesGlobal(agent, initGlobal)); err != nil {
			out.Global().Printf("   %s %v\n", yellow("⚠"), err)
		} else {
			out.Global().Printf("   %s Hook installed\n", green("✓"))
			out.Global().Printf("   %s %s\n", yellow("ℹ"), agent.Instructions)
		}
		out.Global().Println()
	}

	// Create default config if it doesn't exist
	configPath := config.ConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaulttokConfig(); err == nil {
			out.Global().Printf("%s Created default config at %s\n", green("✓"), configPath)
		}
	}

	out.Global().Println()
	out.Global().Println(green("🎉 Setup complete!"))
	out.Global().Println()
	out.Global().Println("tok is now integrated with your AI agents.")
	out.Global().Println("Token compression will be applied automatically.")

	return nil
}

func runInitUninstall(toSetup []AgentInfo) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	out.Global().Println()
	out.Global().Println("Removing tok integration...")
	out.Global().Println()

	for _, agent := range toSetup {
		out.Global().Printf("🧹 %s\n", agent.Name)
		removed, err := uninstallAgent(agent)
		if err != nil {
			out.Global().Printf("   %s %v\n", yellow("⚠"), err)
		} else if len(removed) == 0 {
			out.Global().Printf("   %s nothing to remove\n", yellow("ℹ"))
		} else {
			out.Global().Printf("   %s removed %d artifact(s)\n", green("✓"), len(removed))
		}
		out.Global().Println()
	}

	out.Global().Println(green("Cleanup complete."))
	return nil
}

func runInteractiveInit(agents []AgentInfo) error {
	// Count detected agents
	detected := []AgentInfo{}
	for _, agent := range agents {
		if agent.Detected {
			detected = append(detected, agent)
		}
	}

	if len(detected) == 0 {
		out.Global().Println("No AI agents detected in standard locations.")
		out.Global().Println("\nPlease specify which agent to setup:")
		out.Global().Println("  tok init --claude     # For Claude Code")
		out.Global().Println("  tok init --cursor     # For Cursor")
		out.Global().Println("  tok init --windsurf   # For Windsurf")
		out.Global().Println("  tok init --opencode   # For OpenCode")
		out.Global().Println("  tok init --openclaw   # For OpenClaw")
		out.Global().Println("  tok init --kilocode   # For Kilo Code")
		out.Global().Println("  tok init --antigravity # For Google Antigravity")
		return nil
	}

	out.Global().Println("Detected AI agents:")
	for i, agent := range detected {
		out.Global().Printf("  %d. %s\n", i+1, agent.Name)
	}
	out.Global().Println()

	// For single agent, auto-setup
	if len(detected) == 1 {
		out.Global().Printf("Auto-setting up for %s...\n", detected[0].Name)
		return setupAgent(detected[0], installUsesGlobal(detected[0], initGlobal))
	}

	// For multiple agents, ask user
	out.Global().Print("Setup for all detected agents? [Y/n]: ")
	var response string
	fmt.Scanln(&response)

	if response == "" || response == "y" || response == "Y" {
		for _, agent := range detected {
			if err := setupAgent(agent, installUsesGlobal(agent, initGlobal)); err != nil {
				out.Global().Printf("Warning: failed to setup %s: %v\n", agent.Name, err)
			}
		}
	}

	return nil
}

func setupAgent(agent AgentInfo, global bool) error {
	switch agent.Name {
	case "Codex":
		return setupCodexAgent(agent, global)
	case "GitHub Copilot":
		return setupCopilotAgent(agent)
	case "OpenCode":
		return setupOpenCodeAgent(agent, global)
	case "Windsurf":
		return upsertManagedBlockFile(filepath.Join(agent.ConfigDir, ".windsurfrules"), "tok:windsurf", generateWorkspaceRules(agent.Name))
	case "Cline":
		return upsertManagedBlockFile(filepath.Join(agent.ConfigDir, ".clinerules"), "tok:cline", generateWorkspaceRules(agent.Name))
	case "Kilo Code":
		return writeOwnedFile(filepath.Join(agent.ConfigDir, ".kilocode", "rules", "tok-rules.md"), generateWorkspaceRules(agent.Name), 0644)
	case "Google Antigravity":
		return writeOwnedFile(filepath.Join(agent.ConfigDir, ".agents", "rules", "antigravity-tok-rules.md"), generateWorkspaceRules(agent.Name), 0644)
	}

	// Create hooks directory
	if err := os.MkdirAll(agent.HookDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	// Create hook script
	hookPath := filepath.Join(agent.HookDir, "tok-rewrite.sh")
	hookScript := generateAgentHookScript(agent.Name)

	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("cannot write hook script: %w", err)
	}
	if err := integrity.StoreHash(hookPath); err != nil {
		return fmt.Errorf("cannot store hook integrity baseline: %w", err)
	}

	// Create instructions file
	instructionsPath := filepath.Join(agent.ConfigDir, "TOK.md")
	instructions := generateInstructions(agent.Name)

	if err := os.WriteFile(instructionsPath, []byte(instructions), 0644); err != nil {
		return fmt.Errorf("cannot write instructions: %w", err)
	}

	if err := patchAgentIntegration(agent, hookPath); err != nil {
		return err
	}

	return nil
}

func generateAgentHookScript(agentName string) string {
	handler := hookHandlerForAgent(agentName)
	return fmt.Sprintf(`#!/bin/bash
# tok hook for %s
# Auto-generated by 'tok init'
# tok-hook-version: %d

# Delegate structured hook payloads to tok hook processors.
# If stdin is a TTY, fall back to direct CLI execution.
if ! command -v tok >/dev/null 2>&1; then
    if [ ! -t 0 ]; then
        cat >/dev/null
    fi
    exit 0
fi

if [ -t 0 ]; then
    exec tok "$@"
fi

exec tok hook %s "$@"
`, agentName, integrity.CurrentHookVersion, handler)
}

func hookHandlerForAgent(agentName string) string {
	switch {
	case strings.EqualFold(agentName, "Claude Code"):
		return "claude"
	case strings.EqualFold(agentName, "Cursor"):
		return "cursor"
	case strings.EqualFold(agentName, "Gemini CLI"):
		return "gemini"
	default:
		return "copilot"
	}
}

func patchAgentIntegration(agent AgentInfo, hookPath string) error {
	switch agent.Name {
	case "Claude Code":
		if err := patchClaudeSettingsFile(filepath.Join(agent.ConfigDir, "settings.json"), hookPath); err != nil {
			return fmt.Errorf("cannot patch Claude settings: %w", err)
		}
		if err := ensureReferenceFileContains(filepath.Join(agent.ConfigDir, "CLAUDE.md"), "@TOK.md"); err != nil {
			return fmt.Errorf("cannot patch CLAUDE.md: %w", err)
		}
	case "Cursor":
		if err := patchCursorHooksFile(filepath.Join(agent.ConfigDir, "hooks.json"), hookPath); err != nil {
			return fmt.Errorf("cannot patch Cursor hooks.json: %w", err)
		}
	case "Gemini CLI":
		if err := patchGeminiSettingsFile(filepath.Join(agent.ConfigDir, "settings.json"), hookPath); err != nil {
			return fmt.Errorf("cannot patch Gemini settings: %w", err)
		}
	}
	return nil
}

func uninstallAgent(agent AgentInfo) ([]string, error) {
	switch agent.Name {
	case "Codex":
		return uninstallCodexAgent(agent)
	case "GitHub Copilot":
		return uninstallCopilotAgent(agent)
	case "OpenCode":
		return uninstallOpenCodeAgent(agent)
	case "Windsurf":
		return uninstallManagedBlockFile(filepath.Join(agent.ConfigDir, ".windsurfrules"), "tok:windsurf")
	case "Cline":
		return uninstallManagedBlockFile(filepath.Join(agent.ConfigDir, ".clinerules"), "tok:cline")
	case "Kilo Code":
		return uninstallOwnedFiles(filepath.Join(agent.ConfigDir, ".kilocode", "rules", "tok-rules.md"))
	case "Google Antigravity":
		return uninstallOwnedFiles(filepath.Join(agent.ConfigDir, ".agents", "rules", "antigravity-tok-rules.md"))
	}

	var removed []string
	hookPath := filepath.Join(agent.HookDir, "tok-rewrite.sh")
	legacyHookPath := filepath.Join(agent.HookDir, "tok.sh")

	if ok, err := removeFile(hookPath); err != nil {
		return removed, err
	} else if ok {
		removed = append(removed, hookPath)
	}
	if ok, err := integrity.RemoveHash(hookPath); err != nil {
		return removed, err
	} else if ok {
		removed = append(removed, integrity.HashPath(hookPath))
	}

	if ok, err := removeFile(legacyHookPath); err != nil {
		return removed, err
	} else if ok {
		removed = append(removed, legacyHookPath)
	}
	if ok, err := integrity.RemoveHash(legacyHookPath); err != nil {
		return removed, err
	} else if ok {
		removed = append(removed, integrity.HashPath(legacyHookPath))
	}

	instructionsPath := filepath.Join(agent.ConfigDir, "TOK.md")
	if ok, err := removeFile(instructionsPath); err != nil {
		return removed, err
	} else if ok {
		removed = append(removed, instructionsPath)
	}

	changed, err := removeAgentIntegration(agent, hookPath, legacyHookPath)
	if err != nil {
		return removed, err
	}
	removed = append(removed, changed...)
	return removed, nil
}

func removeAgentIntegration(agent AgentInfo, hookPath, legacyHookPath string) ([]string, error) {
	var removed []string
	switch agent.Name {
	case "Claude Code":
		settingsPath := filepath.Join(agent.ConfigDir, "settings.json")
		changed, err := removeClaudeHook(settingsPath, hookPath, legacyHookPath)
		if err != nil {
			return removed, err
		}
		if changed {
			removed = append(removed, settingsPath)
		}
		claudeMD := filepath.Join(agent.ConfigDir, "CLAUDE.md")
		changed, err = removeReferenceFromFile(claudeMD, "@TOK.md")
		if err != nil {
			return removed, err
		}
		if changed {
			removed = append(removed, claudeMD)
		}
	case "Cursor":
		hooksPath := filepath.Join(agent.ConfigDir, "hooks.json")
		changed, err := removeCursorHook(hooksPath, hookPath, legacyHookPath)
		if err != nil {
			return removed, err
		}
		if changed {
			removed = append(removed, hooksPath)
		}
	case "Gemini CLI":
		settingsPath := filepath.Join(agent.ConfigDir, "settings.json")
		changed, err := removeGeminiHook(settingsPath, hookPath, legacyHookPath)
		if err != nil {
			return removed, err
		}
		if changed {
			removed = append(removed, settingsPath)
		}
	}
	return removed, nil
}

func patchClaudeSettingsFile(path, hookPath string) error {
	root, err := loadJSONObject(path)
	if err != nil {
		return err
	}
	changed, err := ensureClaudeHook(root, hookPath)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return writeJSONObject(path, root)
}

func patchCursorHooksFile(path, hookPath string) error {
	root, err := loadJSONObject(path)
	if err != nil {
		return err
	}
	changed, err := ensureCursorHook(root, hookPath)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return writeJSONObject(path, root)
}

func patchGeminiSettingsFile(path, hookPath string) error {
	root, err := loadJSONObject(path)
	if err != nil {
		return err
	}
	changed, err := ensureGeminiHook(root, hookPath)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return writeJSONObject(path, root)
}

func loadJSONObject(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(string(content)) == "" {
		return map[string]any{}, nil
	}

	var root map[string]any
	if err := json.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if root == nil {
		root = map[string]any{}
	}
	return root, nil
}

func writeJSONObject(path string, root map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tok-json-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func ensureReferenceFileContains(path, reference string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	trimmed := strings.TrimSpace(string(content))
	if strings.Contains(trimmed, reference) {
		return nil
	}
	if trimmed != "" {
		trimmed += "\n\n"
	}
	trimmed += reference + "\n"
	return os.WriteFile(path, []byte(trimmed), 0644)
}

func removeReferenceFromFile(path, reference string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	lines := strings.Split(string(content), "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		if strings.TrimSpace(line) == reference {
			changed = true
			continue
		}
		out = append(out, line)
	}
	if !changed {
		return false, nil
	}
	updated := strings.TrimRight(strings.Join(out, "\n"), "\n")
	if updated != "" {
		updated += "\n"
	}
	return true, os.WriteFile(path, []byte(updated), 0644)
}

func ensureClaudeHook(root map[string]any, hookPath string) (bool, error) {
	if hasClaudeHook(root, hookPath) {
		return false, nil
	}
	hooks, err := ensureObject(root, "hooks")
	if err != nil {
		return false, err
	}
	preToolUse, err := ensureArray(hooks, "PreToolUse")
	if err != nil {
		return false, err
	}
	preToolUse = append(preToolUse, map[string]any{
		"matcher": "Bash",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookPath,
			},
		},
	})
	hooks["PreToolUse"] = preToolUse
	return true, nil
}

func hasClaudeHook(root map[string]any, hookPath string) bool {
	return commandHookExists(root, hookPath, "hooks", "PreToolUse")
}

func removeClaudeHook(path string, hookPaths ...string) (bool, error) {
	root, err := loadJSONObject(path)
	if err != nil {
		return false, err
	}
	changed := removeNestedCommandHook(root, "hooks", "PreToolUse", hookPaths...)
	if !changed {
		return false, nil
	}
	return true, writeJSONObject(path, root)
}

func ensureCursorHook(root map[string]any, hookPath string) (bool, error) {
	if hasCursorHook(root, hookPath) {
		return false, nil
	}
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}
	hooks, err := ensureObject(root, "hooks")
	if err != nil {
		return false, err
	}
	preToolUse, err := ensureArray(hooks, "preToolUse")
	if err != nil {
		return false, err
	}
	preToolUse = append(preToolUse, map[string]any{
		"matcher": "Shell",
		"command": hookPath,
	})
	hooks["preToolUse"] = preToolUse
	return true, nil
}

func hasCursorHook(root map[string]any, hookPath string) bool {
	hooks, ok := root["hooks"].(map[string]any)
	if !ok {
		return false
	}
	preToolUse, ok := hooks["preToolUse"].([]any)
	if !ok {
		return false
	}
	for _, raw := range preToolUse {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		if command, _ := entry["command"].(string); command == hookPath {
			return true
		}
	}
	return false
}

func installUsesGlobal(agent AgentInfo, global bool) bool {
	return global || agent.Name == "OpenCode"
}

func removeCursorHook(path string, hookPaths ...string) (bool, error) {
	root, err := loadJSONObject(path)
	if err != nil {
		return false, err
	}
	changed := removeDirectCommandHook(root, "hooks", "preToolUse", hookPaths...)
	if !changed {
		return false, nil
	}
	return true, writeJSONObject(path, root)
}

func ensureGeminiHook(root map[string]any, hookPath string) (bool, error) {
	if hasGeminiHook(root, hookPath) {
		return false, nil
	}
	hooks, err := ensureObject(root, "hooks")
	if err != nil {
		return false, err
	}
	beforeTool, err := ensureArray(hooks, "BeforeTool")
	if err != nil {
		return false, err
	}
	beforeTool = append(beforeTool, map[string]any{
		"matcher": "run_shell_command",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookPath,
			},
		},
	})
	hooks["BeforeTool"] = beforeTool
	return true, nil
}

func hasGeminiHook(root map[string]any, hookPath string) bool {
	return commandHookExists(root, hookPath, "hooks", "BeforeTool")
}

func removeGeminiHook(path string, hookPaths ...string) (bool, error) {
	root, err := loadJSONObject(path)
	if err != nil {
		return false, err
	}
	changed := removeNestedCommandHook(root, "hooks", "BeforeTool", hookPaths...)
	if !changed {
		return false, nil
	}
	return true, writeJSONObject(path, root)
}

func commandHookExists(root map[string]any, hookPath string, outerKey, arrayKey string) bool {
	outer, ok := root[outerKey].(map[string]any)
	if !ok {
		return false
	}
	items, ok := outer[arrayKey].([]any)
	if !ok {
		return false
	}
	for _, raw := range items {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		hooks, _ := entry["hooks"].([]any)
		for _, hookRaw := range hooks {
			hook, _ := hookRaw.(map[string]any)
			if hook == nil {
				continue
			}
			if command, _ := hook["command"].(string); command == hookPath {
				return true
			}
		}
	}
	return false
}

func removeNestedCommandHook(root map[string]any, outerKey, arrayKey string, hookPaths ...string) bool {
	outer, ok := root[outerKey].(map[string]any)
	if !ok {
		return false
	}
	items, ok := outer[arrayKey].([]any)
	if !ok {
		return false
	}
	originalLen := len(items)
	filtered := make([]any, 0, len(items))
	for _, raw := range items {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			filtered = append(filtered, raw)
			continue
		}
		hooks, _ := entry["hooks"].([]any)
		keep := true
		for _, hookRaw := range hooks {
			hook, _ := hookRaw.(map[string]any)
			if hook == nil {
				continue
			}
			command, _ := hook["command"].(string)
			if stringInSlice(command, hookPaths) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, raw)
		}
	}
	if len(filtered) == originalLen {
		return false
	}
	outer[arrayKey] = filtered
	return true
}

func removeDirectCommandHook(root map[string]any, outerKey, arrayKey string, hookPaths ...string) bool {
	outer, ok := root[outerKey].(map[string]any)
	if !ok {
		return false
	}
	items, ok := outer[arrayKey].([]any)
	if !ok {
		return false
	}
	originalLen := len(items)
	filtered := make([]any, 0, len(items))
	for _, raw := range items {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			filtered = append(filtered, raw)
			continue
		}
		command, _ := entry["command"].(string)
		if stringInSlice(command, hookPaths) {
			continue
		}
		filtered = append(filtered, raw)
	}
	if len(filtered) == originalLen {
		return false
	}
	outer[arrayKey] = filtered
	return true
}

func ensureObject(root map[string]any, key string) (map[string]any, error) {
	if existing, ok := root[key]; ok {
		object, ok := existing.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an object", key)
		}
		return object, nil
	}
	object := map[string]any{}
	root[key] = object
	return object, nil
}

func ensureArray(root map[string]any, key string) ([]any, error) {
	if existing, ok := root[key]; ok {
		array, ok := existing.([]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an array", key)
		}
		return array, nil
	}
	array := []any{}
	root[key] = array
	return array, nil
}

func removeFile(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := os.Remove(path); err != nil {
		return false, err
	}
	return true, nil
}

func stringInSlice(target string, values []string) bool {
	for _, value := range values {
		if target == value {
			return true
		}
	}
	return false
}

func resolveCodexConfigDir(home string) string {
	if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
		return codexHome
	}
	return filepath.Join(home, ".codex")
}

func setupCodexAgent(agent AgentInfo, global bool) error {
	configDir := agent.ConfigDir
	if global {
		home, _ := os.UserHomeDir()
		configDir = resolveCodexConfigDir(home)
	}

	tokPath := filepath.Join(configDir, "TOK.md")
	if err := writeOwnedFile(tokPath, generateInstructions(agent.Name), 0644); err != nil {
		return err
	}

	reference := "@TOK.md"
	if global {
		reference = "@" + tokPath
	}
	if err := ensureReferenceFileContains(filepath.Join(configDir, "AGENTS.md"), reference); err != nil {
		return err
	}
	return nil
}

func setupCopilotAgent(agent AgentInfo) error {
	hooksDir := filepath.Join(agent.ConfigDir, ".github", "hooks")
	hookConfigPath := filepath.Join(hooksDir, "tok-rewrite.json")
	instructionsPath := filepath.Join(agent.ConfigDir, ".github", "copilot-instructions.md")

	if err := writeOwnedFile(hookConfigPath, generateCopilotHookConfig(), 0644); err != nil {
		return err
	}
	return writeOwnedFile(instructionsPath, generateCopilotInstructions(), 0644)
}

func setupOpenCodeAgent(agent AgentInfo, global bool) error {
	if !global {
		return fmt.Errorf("OpenCode plugin is global-only; run 'tok init --global --opencode'")
	}
	pluginPath := filepath.Join(agent.ConfigDir, "plugins", "tok.ts")
	return writeOwnedFile(pluginPath, generateOpenCodePlugin(), 0644)
}

func uninstallCodexAgent(agent AgentInfo) ([]string, error) {
	home, _ := os.UserHomeDir()
	globalDir := resolveCodexConfigDir(home)
	var removed []string

	dirs := []struct {
		dir       string
		reference string
	}{
		{dir: agent.ConfigDir, reference: "@TOK.md"},
	}
	if globalDir != agent.ConfigDir {
		dirs = append(dirs, struct {
			dir       string
			reference string
		}{
			dir:       globalDir,
			reference: "@" + filepath.Join(globalDir, "TOK.md"),
		})
	}

	for _, item := range dirs {
		tokPath := filepath.Join(item.dir, "TOK.md")
		if ok, err := removeFile(tokPath); err != nil {
			return removed, err
		} else if ok {
			removed = append(removed, tokPath)
		}

		agentsPath := filepath.Join(item.dir, "AGENTS.md")
		changed, err := removeReferencesFromFile(agentsPath, item.reference, "@TOK.md")
		if err != nil {
			return removed, err
		}
		if changed {
			removed = append(removed, agentsPath)
		}
	}

	return removed, nil
}

func uninstallCopilotAgent(agent AgentInfo) ([]string, error) {
	return uninstallOwnedFiles(
		filepath.Join(agent.ConfigDir, ".github", "hooks", "tok-rewrite.json"),
		filepath.Join(agent.ConfigDir, ".github", "copilot-instructions.md"),
	)
}

func uninstallOpenCodeAgent(agent AgentInfo) ([]string, error) {
	return uninstallOwnedFiles(filepath.Join(agent.ConfigDir, "plugins", "tok.ts"))
}

func uninstallOwnedFiles(paths ...string) ([]string, error) {
	removed := make([]string, 0, len(paths))
	for _, path := range paths {
		ok, err := removeFile(path)
		if err != nil {
			return removed, err
		}
		if ok {
			removed = append(removed, path)
		}
	}
	return removed, nil
}

func upsertManagedBlockFile(path, marker, body string) error {
	begin := managedBlockStart(marker)
	end := managedBlockEnd(marker)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existing := strings.TrimSpace(string(content))
	block := begin + "\n" + body + "\n" + end

	var updated string
	switch {
	case strings.Contains(existing, begin) && strings.Contains(existing, end):
		start := strings.Index(existing, begin)
		finish := strings.Index(existing[start:], end)
		if finish < 0 {
			updated = strings.TrimSpace(existing + "\n\n" + block)
		} else {
			finish += start + len(end)
			updated = strings.TrimSpace(existing[:start] + block + existing[finish:])
		}
	case existing == "":
		updated = block
	default:
		updated = strings.TrimSpace(existing + "\n\n" + block)
	}
	return os.WriteFile(path, []byte(updated+"\n"), 0644)
}

func uninstallManagedBlockFile(path, marker string) ([]string, error) {
	changed, err := removeManagedBlockFile(path, marker)
	if err != nil {
		return nil, err
	}
	if changed {
		return []string{path}, nil
	}
	return nil, nil
}

func removeManagedBlockFile(path, marker string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	text := string(content)
	begin := managedBlockStart(marker)
	end := managedBlockEnd(marker)
	start := strings.Index(text, begin)
	if start < 0 {
		return false, nil
	}
	finish := strings.Index(text[start:], end)
	if finish < 0 {
		return false, nil
	}
	finish += start + len(end)
	updated := strings.TrimSpace(text[:start] + text[finish:])
	if updated == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		return true, nil
	}
	return true, os.WriteFile(path, []byte(updated+"\n"), 0644)
}

func managedBlockStart(marker string) string {
	return fmt.Sprintf("<!-- %s:start -->", marker)
}

func managedBlockEnd(marker string) string {
	return fmt.Sprintf("<!-- %s:end -->", marker)
}

func writeOwnedFile(path, content string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), perm)
}

func removeReferencesFromFile(path string, references ...string) (bool, error) {
	changed := false
	for _, reference := range references {
		updated, err := removeReferenceFromFile(path, reference)
		if err != nil {
			return changed, err
		}
		changed = changed || updated
	}
	return changed, nil
}

func generateInstructions(agentName string) string {
	return fmt.Sprintf(`# tok Integration for %s

tok is a token-aware CLI proxy that reduces LLM token consumption.

## Quick Start

1. tok hooks are installed in the hooks/ directory
2. Commands are automatically compressed before being sent to the LLM
3. View savings with: tok gain

## Available Commands

- tok status    - Check tok status
- tok gain      - View token savings
- tok doctor    - Run diagnostics
- tok cc-economics - Compare Claude Code costs vs savings

## Configuration

Edit ~/.config/tok/config.toml to customize behavior.

## Documentation

https://github.com/lakshmanpatel/tok
`, agentName)
}

func generateWorkspaceRules(agentName string) string {
	return fmt.Sprintf(
		"# tok Rules for %s\n\n"+
			"tok sits between your coding agent and the LLM provider. Prefer `tok`-prefixed shell commands so large terminal output is reduced before it reaches the model.\n\n"+
			"Use:\n"+
			"- `tok git status`\n"+
			"- `tok git diff --stat`\n"+
			"- `tok npm test`\n"+
			"- `tok go test ./...`\n"+
			"- `tok docker ps`\n\n"+
			"Use direct tok commands for analysis:\n"+
			"- `tok status`\n"+
			"- `tok gain`\n"+
			"- `tok discover`\n"+
			"- `tok cc-economics`\n",
		agentName,
	)
}

func generateCopilotHookConfig() string {
	return `{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "tok hook copilot",
        "cwd": ".",
        "timeout": 5
      }
    ]
  }
}
`
}

func generateCopilotInstructions() string {
	return "# tok for GitHub Copilot\n\n" +
		"tok rewrites shell commands before Copilot executes them so large terminal output is compressed before it is sent to an LLM.\n\n" +
		"Prefer tok-prefixed commands:\n\n" +
		"```bash\n" +
		"tok git status\n" +
		"tok git diff --stat\n" +
		"tok npm test\n" +
		"tok go test ./...\n" +
		"```\n\n" +
		"Useful tok commands:\n\n" +
		"```bash\n" +
		"tok status\n" +
		"tok gain\n" +
		"tok discover\n" +
		"tok cc-economics\n" +
		"```\n"
}

func generateOpenCodePlugin() string {
	return `import type { Plugin } from "@opencode-ai/plugin"

// tok OpenCode plugin — rewrites commands to use tok for token savings.

export const TokOpenCodePlugin: Plugin = async ({ $ }) => {
  try {
    await $` + "`which tok`" + `.quiet()
  } catch {
    console.warn("[tok] tok binary not found in PATH — plugin disabled")
    return {}
  }

  return {
    "tool.execute.before": async (input, output) => {
      const tool = String(input?.tool ?? "").toLowerCase()
      if (tool !== "bash" && tool !== "shell") return
      const args = output?.args
      if (!args || typeof args !== "object") return

      const command = (args as Record<string, unknown>).command
      if (typeof command !== "string" || !command) return

      try {
        const result = await $` + "`tok rewrite ${command}`" + `.quiet().nothrow()
        const rewritten = String(result.stdout).trim()
        if (rewritten && rewritten !== command) {
          ;(args as Record<string, unknown>).command = rewritten
        }
      } catch {
        // tok rewrite failed — pass through unchanged
      }
    },
  }
}
`
}

func createDefaulttokConfig() error {
	configPath := config.ConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	defaultConfig := `# tok Configuration
# https://github.com/lakshmanpatel/tok

[tracking]
enabled = true
database_path = ""
retention_days = 90

[filter]
mode = "minimal"
max_width = 0

[pipeline]
max_context_tokens = 2000000
default_budget = 0
entropy_threshold = 0.3
perplexity_threshold = 0.5

[hooks]
excluded_commands = []
`
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

func showInitConfig() error {
	agents, err := currentAgentInfos(false)
	if err != nil {
		return err
	}

	out.Global().Println("tok Agent Configuration")
	out.Global().Println("==========================")
	out.Global().Println()

	for _, agent := range agents {
		status, detail := describeAgentStatus(agent)
		if detail != "" {
			out.Global().Printf("  %-20s %s (%s)\n", agent.Name+":", status, detail)
		} else {
			out.Global().Printf("  %-20s %s\n", agent.Name+":", status)
		}
	}

	out.Global().Println()
	out.Global().Println("Run 'tok init --all' to setup all detected agents")

	return nil
}

func describeAgentStatus(agent AgentInfo) (string, string) {
	detectDir := agent.DetectDir
	if detectDir == "" {
		detectDir = agent.ConfigDir
	}
	detected := fileExists(detectDir)
	hookPath := filepath.Join(agent.ConfigDir, "hooks", "tok-rewrite.sh")
	legacyHookPath := filepath.Join(agent.ConfigDir, "hooks", "tok.sh")
	switch agent.Name {
	case "Windsurf":
		rulesPath := filepath.Join(agent.ConfigDir, ".windsurfrules")
		switch {
		case managedBlockPresent(rulesPath, "tok:windsurf"):
			return "configured", ".windsurfrules patched"
		case fileExists(rulesPath):
			return "partial", ".windsurfrules exists without tok block"
		case detected:
			return "detected", "not configured"
		default:
			return "not detected", ""
		}
	case "Cline":
		rulesPath := filepath.Join(agent.ConfigDir, ".clinerules")
		switch {
		case managedBlockPresent(rulesPath, "tok:cline"):
			return "configured", ".clinerules patched"
		case fileExists(rulesPath):
			return "partial", ".clinerules exists without tok block"
		case detected:
			return "detected", "not configured"
		default:
			return "not detected", ""
		}
	case "Claude Code":
		if !detected {
			return "not detected", ""
		}
		settingsPath := filepath.Join(agent.ConfigDir, "settings.json")
		claudeMD := filepath.Join(agent.ConfigDir, "CLAUDE.md")
		if _, err := loadJSONObject(settingsPath); err != nil && fileExists(settingsPath) {
			return "broken", "invalid settings.json"
		}
		switch {
		case claudeHookConfigured(settingsPath, hookPath):
			detail := "settings.json patched"
			if fileExists(claudeMD) {
				if content, err := os.ReadFile(claudeMD); err == nil && strings.Contains(string(content), "@TOK.md") {
					detail += ", CLAUDE.md linked"
				}
			}
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "configured", detail
		case fileExists(hookPath):
			detail := "hook exists but settings.json not patched"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "partial", detail
		case fileExists(legacyHookPath):
			return "legacy", "old hook script present"
		default:
			return "detected", "not configured"
		}
	case "Cursor":
		if !detected {
			return "not detected", ""
		}
		hooksPath := filepath.Join(agent.ConfigDir, "hooks.json")
		if _, err := loadJSONObject(hooksPath); err != nil && fileExists(hooksPath) {
			return "broken", "invalid hooks.json"
		}
		switch {
		case cursorHookConfigured(hooksPath, hookPath):
			detail := "hooks.json patched"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "configured", detail
		case fileExists(hookPath):
			detail := "hook exists but hooks.json not patched"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "partial", detail
		case fileExists(legacyHookPath):
			return "legacy", "old hook script present"
		default:
			return "detected", "not configured"
		}
	case "Gemini CLI":
		if !detected {
			return "not detected", ""
		}
		settingsPath := filepath.Join(agent.ConfigDir, "settings.json")
		if _, err := loadJSONObject(settingsPath); err != nil && fileExists(settingsPath) {
			return "broken", "invalid settings.json"
		}
		switch {
		case geminiHookConfigured(settingsPath, hookPath):
			detail := "settings.json patched"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "configured", detail
		case fileExists(hookPath):
			detail := "hook exists but settings.json not patched"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "partial", detail
		default:
			return "detected", "not configured"
		}
	case "Codex":
		return describeCodexStatus(agent)
	case "GitHub Copilot":
		return describeCopilotStatus(agent)
	case "OpenCode":
		pluginPath := filepath.Join(agent.ConfigDir, "plugins", "tok.ts")
		switch {
		case fileExists(pluginPath):
			return "configured", "plugin installed"
		case detected:
			return "detected", "not configured"
		default:
			return "not detected", ""
		}
	case "Kilo Code":
		rulesPath := filepath.Join(agent.ConfigDir, ".kilocode", "rules", "tok-rules.md")
		switch {
		case fileExists(rulesPath):
			return "configured", ".kilocode/rules/tok-rules.md installed"
		case detected:
			return "detected", "not configured"
		default:
			return "not detected", ""
		}
	case "Google Antigravity":
		rulesPath := filepath.Join(agent.ConfigDir, ".agents", "rules", "antigravity-tok-rules.md")
		switch {
		case fileExists(rulesPath):
			return "configured", ".agents/rules/antigravity-tok-rules.md installed"
		case detected:
			return "detected", "not configured"
		default:
			return "not detected", ""
		}
	default:
		if !detected {
			return "not detected", ""
		}
		if fileExists(hookPath) {
			detail := "hook installed"
			if health := hookHealthDetail(hookPath); health != "" {
				detail += ", " + health
			}
			return "configured", detail
		}
		if fileExists(legacyHookPath) {
			return "legacy", "old hook script present"
		}
		return "detected", "not configured"
	}
}

func describeCodexStatus(agent AgentInfo) (string, string) {
	home, _ := os.UserHomeDir()
	globalDir := resolveCodexConfigDir(home)
	localDir := agent.ConfigDir
	globalConfigured := codexScopeConfigured(globalDir, "@"+filepath.Join(globalDir, "TOK.md"))
	localConfigured := codexScopeConfigured(localDir, "@TOK.md")

	switch {
	case globalConfigured && localConfigured:
		return "configured", "global and local AGENTS.md linked"
	case globalConfigured:
		return "configured", "global AGENTS.md linked"
	case localConfigured:
		return "configured", "local AGENTS.md linked"
	}

	switch {
	case codexScopePartial(globalDir, "@"+filepath.Join(globalDir, "TOK.md")) && codexScopePartial(localDir, "@TOK.md"):
		return "partial", "global and local Codex files need repair"
	case codexScopePartial(globalDir, "@"+filepath.Join(globalDir, "TOK.md")):
		return "partial", "global Codex files need repair"
	case codexScopePartial(localDir, "@TOK.md"):
		return "partial", "local Codex files need repair"
	case fileExists(globalDir) || fileExists(filepath.Join(localDir, "AGENTS.md")) || fileExists(filepath.Join(localDir, "TOK.md")):
		return "detected", "not configured"
	default:
		return "not detected", ""
	}
}

func describeCopilotStatus(agent AgentInfo) (string, string) {
	hookConfigPath := filepath.Join(agent.ConfigDir, ".github", "hooks", "tok-rewrite.json")
	instructionsPath := filepath.Join(agent.ConfigDir, ".github", "copilot-instructions.md")
	switch {
	case fileExists(hookConfigPath) && fileExists(instructionsPath):
		return "configured", ".github hooks and instructions installed"
	case fileExists(hookConfigPath) || fileExists(instructionsPath):
		return "partial", "Copilot project files incomplete"
	case fileExists(filepath.Join(agent.ConfigDir, ".github")):
		return "detected", "not configured"
	default:
		return "not detected", ""
	}
}

func codexScopeConfigured(baseDir, reference string) bool {
	return fileContains(filepath.Join(baseDir, "AGENTS.md"), reference) && fileExists(filepath.Join(baseDir, "TOK.md"))
}

func codexScopePartial(baseDir, reference string) bool {
	return fileExists(filepath.Join(baseDir, "TOK.md")) || fileContains(filepath.Join(baseDir, "AGENTS.md"), reference)
}

func managedBlockPresent(path, marker string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(content)
	return strings.Contains(text, managedBlockStart(marker)) && strings.Contains(text, managedBlockEnd(marker))
}

func fileContains(path, needle string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), needle)
}

func hookHealthDetail(hookPath string) string {
	result, err := integrity.VerifyHookAt(hookPath)
	if err != nil {
		return ""
	}
	switch result.Status {
	case integrity.StatusVerified:
		return "integrity verified"
	case integrity.StatusOutdated:
		return "outdated hook"
	case integrity.StatusNoBaseline:
		return "no integrity baseline"
	case integrity.StatusTampered:
		return "hook modified"
	default:
		return ""
	}
}

func claudeHookConfigured(path, hookPath string) bool {
	root, err := loadJSONObject(path)
	if err != nil {
		return false
	}
	return hasClaudeHook(root, hookPath)
}

func cursorHookConfigured(path, hookPath string) bool {
	root, err := loadJSONObject(path)
	if err != nil {
		return false
	}
	return hasCursorHook(root, hookPath)
}

func geminiHookConfigured(path, hookPath string) bool {
	root, err := loadJSONObject(path)
	if err != nil {
		return false
	}
	return hasGeminiHook(root, hookPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
