package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
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
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize TokMan for AI agents",
	Long: `Set up TokMan integration with AI coding assistants.

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
  tokman init                    # Interactive setup (detects agents)
  tokman init --claude           # Setup for Claude Code only
  tokman init --cursor --windsurf # Setup for multiple agents
  tokman init --all              # Setup for all detected agents
  tokman init --global --all     # Global installation for all agents`,
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
	initCmd.Flags().BoolVar(&initShow, "show", false, "Show current configuration")
}

// AgentInfo holds information about an AI agent
type AgentInfo struct {
	Name         string
	Flag         *bool
	ConfigDir    string
	HookDir      string
	Detected     bool
	Instructions string
}

func runInit(cmd *cobra.Command, args []string) error {
	if initShow {
		return showInitConfig()
	}

	home, _ := os.UserHomeDir()
	if home == "" {
		return fmt.Errorf("cannot determine user home directory")
	}

	// Define all supported agents
	agents := []AgentInfo{
		{
			Name:         "Claude Code",
			Flag:         &initClaude,
			ConfigDir:    filepath.Join(home, ".claude"),
			HookDir:      filepath.Join(home, ".claude", "hooks"),
			Instructions: "Add to ~/.claude/CLAUDE.md or use --global for global config",
		},
		{
			Name:         "Cursor",
			Flag:         &initCursor,
			ConfigDir:    filepath.Join(home, ".cursor"),
			HookDir:      filepath.Join(home, ".cursor", "hooks"),
			Instructions: "Add to ~/.cursor/rules/tokman or use --global",
		},
		{
			Name:         "Windsurf",
			Flag:         &initWindsurf,
			ConfigDir:    filepath.Join(home, ".windsurf"),
			HookDir:      filepath.Join(home, ".windsurf", "hooks"),
			Instructions: "Add to ~/.windsurf/settings.json",
		},
		{
			Name:         "Cline",
			Flag:         &initCline,
			ConfigDir:    filepath.Join(home, ".cline"),
			HookDir:      filepath.Join(home, ".cline", "hooks"),
			Instructions: "Add to VS Code: Cline settings > Custom Instructions",
		},
		{
			Name:         "Gemini CLI",
			Flag:         &initGemini,
			ConfigDir:    filepath.Join(home, ".gemini"),
			HookDir:      filepath.Join(home, ".gemini", "hooks"),
			Instructions: "Add to ~/.gemini/settings.json",
		},
		{
			Name:         "Codex",
			Flag:         &initCodex,
			ConfigDir:    filepath.Join(home, ".codex"),
			HookDir:      filepath.Join(home, ".codex", "hooks"),
			Instructions: "Add AGENTS.md with TokMan instructions",
		},
		{
			Name:         "GitHub Copilot",
			Flag:         &initCopilot,
			ConfigDir:    filepath.Join(home, ".github-copilot"),
			HookDir:      filepath.Join(home, ".github-copilot", "hooks"),
			Instructions: "Add to VS Code: Copilot settings + .github/copilot-instructions.md",
		},
		{
			Name:         "OpenCode",
			Flag:         &initOpencode,
			ConfigDir:    filepath.Join(home, ".opencode"),
			HookDir:      filepath.Join(home, ".opencode", "hooks"),
			Instructions: "Add to ~/.opencode/config.json",
		},
		{
			Name:         "OpenClaw",
			Flag:         &initOpenclaw,
			ConfigDir:    filepath.Join(home, ".openclaw"),
			HookDir:      filepath.Join(home, ".openclaw", "hooks"),
			Instructions: "Add to ~/.openclaw/config.json",
		},
		{
			Name:         "Kilo Code",
			Flag:         &initKilocode,
			ConfigDir:    filepath.Join(home, ".kilocode"),
			HookDir:      filepath.Join(home, ".kilocode", "hooks"),
			Instructions: "Add to VS Code: Kilo Code settings",
		},
		{
			Name:         "Google Antigravity",
			Flag:         &initAntigravity,
			ConfigDir:    filepath.Join(home, ".antigravity"),
			HookDir:      filepath.Join(home, ".antigravity", "hooks"),
			Instructions: "Add to ~/.antigravity/config.json",
		},
	}

	// Detect which agents are installed
	for i := range agents {
		if _, err := os.Stat(agents[i].ConfigDir); err == nil {
			agents[i].Detected = true
		}
	}

	// Determine which agents to setup
	var toSetup []AgentInfo

	if initAll {
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
		fmt.Println("No agents selected or detected.")
		fmt.Println("\nTo setup a specific agent, use:")
		fmt.Println("  tokman init --claude     # For Claude Code")
		fmt.Println("  tokman init --cursor     # For Cursor")
		fmt.Println("  tokman init --windsurf   # For Windsurf")
		fmt.Println("  tokman init --opencode   # For OpenCode")
		fmt.Println("  tokman init --openclaw   # For OpenClaw")
		fmt.Println("  tokman init --kilocode   # For Kilo Code")
		fmt.Println("  tokman init --antigravity # For Google Antigravity")
		fmt.Println("\nOr detect all installed agents:")
		fmt.Println("  tokman init --all")
		return nil
	}

	// Setup each agent
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Println("Setting up TokMan for AI agents...")
	fmt.Println()

	for _, agent := range toSetup {
		fmt.Printf("📦 %s\n", agent.Name)

		if err := setupAgent(agent, initGlobal); err != nil {
			fmt.Printf("   %s %v\n", yellow("⚠"), err)
		} else {
			fmt.Printf("   %s Hook installed\n", green("✓"))
			fmt.Printf("   %s %s\n", yellow("ℹ"), agent.Instructions)
		}
		fmt.Println()
	}

	// Create default config if it doesn't exist
	configPath := config.ConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultTokManConfig(); err == nil {
			fmt.Printf("%s Created default config at %s\n", green("✓"), configPath)
		}
	}

	fmt.Println()
	fmt.Println(green("🎉 Setup complete!"))
	fmt.Println()
	fmt.Println("TokMan is now integrated with your AI agents.")
	fmt.Println("Token compression will be applied automatically.")

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
		fmt.Println("No AI agents detected in standard locations.")
		fmt.Println("\nPlease specify which agent to setup:")
		fmt.Println("  tokman init --claude     # For Claude Code")
		fmt.Println("  tokman init --cursor     # For Cursor")
		fmt.Println("  tokman init --windsurf   # For Windsurf")
		fmt.Println("  tokman init --opencode   # For OpenCode")
		fmt.Println("  tokman init --openclaw   # For OpenClaw")
		fmt.Println("  tokman init --kilocode   # For Kilo Code")
		fmt.Println("  tokman init --antigravity # For Google Antigravity")
		return nil
	}

	fmt.Println("Detected AI agents:")
	for i, agent := range detected {
		fmt.Printf("  %d. %s\n", i+1, agent.Name)
	}
	fmt.Println()

	// For single agent, auto-setup
	if len(detected) == 1 {
		fmt.Printf("Auto-setting up for %s...\n", detected[0].Name)
		return setupAgent(detected[0], initGlobal)
	}

	// For multiple agents, ask user
	fmt.Print("Setup for all detected agents? [Y/n]: ")
	var response string
	fmt.Scanln(&response)

	if response == "" || response == "y" || response == "Y" {
		for _, agent := range detected {
			if err := setupAgent(agent, initGlobal); err != nil {
				fmt.Printf("Warning: failed to setup %s: %v\n", agent.Name, err)
			}
		}
	}

	return nil
}

func setupAgent(agent AgentInfo, global bool) error {
	// Create hooks directory
	if err := os.MkdirAll(agent.HookDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	// Create hook script
	hookPath := filepath.Join(agent.HookDir, "tokman-rewrite.sh")
	hookScript := generateAgentHookScript(agent.Name)

	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("cannot write hook script: %w", err)
	}

	// Create instructions file
	instructionsPath := filepath.Join(agent.ConfigDir, "TOKMAN.md")
	instructions := generateInstructions(agent.Name)

	if err := os.WriteFile(instructionsPath, []byte(instructions), 0644); err != nil {
		return fmt.Errorf("cannot write instructions: %w", err)
	}

	return nil
}

func generateAgentHookScript(agentName string) string {
	return fmt.Sprintf(`#!/bin/bash
# TokMan hook for %s
# Auto-generated by 'tokman init'

# Rewrite command through tokman for compression
if command -v tokman &> /dev/null; then
    exec tokman "$@"
else
    exec "$@"
fi
`, agentName)
}

func generateInstructions(agentName string) string {
	return fmt.Sprintf(`# TokMan Integration for %s

TokMan is a token-aware CLI proxy that reduces LLM token consumption.

## Quick Start

1. TokMan hooks are installed in the hooks/ directory
2. Commands are automatically compressed before being sent to the LLM
3. View savings with: tokman gain

## Available Commands

- tokman status    - Check TokMan status
- tokman gain      - View token savings
- tokman doctor    - Run diagnostics
- tokman cc-economics - Compare Claude Code costs vs savings

## Configuration

Edit ~/.config/tokman/config.toml to customize behavior.

## Documentation

https://github.com/GrayCodeAI/tokman
`, agentName)
}

func createDefaultTokManConfig() error {
	configPath := config.ConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	defaultConfig := `# TokMan Configuration
# https://github.com/GrayCodeAI/tokman

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

[hocks]
excluded_commands = []
`
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

func showInitConfig() error {
	home, _ := os.UserHomeDir()
	if home == "" {
		return fmt.Errorf("cannot determine user home directory")
	}

	agents := []struct {
		Name      string
		ConfigDir string
	}{
		{"Claude Code", filepath.Join(home, ".claude")},
		{"Cursor", filepath.Join(home, ".cursor")},
		{"Windsurf", filepath.Join(home, ".windsurf")},
		{"Cline", filepath.Join(home, ".cline")},
		{"Gemini CLI", filepath.Join(home, ".gemini")},
		{"Codex", filepath.Join(home, ".codex")},
		{"GitHub Copilot", filepath.Join(home, ".github-copilot")},
		{"OpenCode", filepath.Join(home, ".opencode")},
		{"OpenClaw", filepath.Join(home, ".openclaw")},
		{"Kilo Code", filepath.Join(home, ".kilocode")},
		{"Google Antigravity", filepath.Join(home, ".antigravity")},
	}

	fmt.Println("TokMan Agent Configuration")
	fmt.Println("==========================")
	fmt.Println()

	for _, agent := range agents {
		status := "not detected"
		if _, err := os.Stat(agent.ConfigDir); err == nil {
			hookPath := filepath.Join(agent.ConfigDir, "hooks", "tokman-rewrite.sh")
			if _, err := os.Stat(hookPath); err == nil {
				status = "configured"
			} else {
				status = "detected (not configured)"
			}
		}
		fmt.Printf("  %-20s %s\n", agent.Name+":", status)
	}

	fmt.Println()
	fmt.Println("Run 'tokman init --all' to setup all detected agents")

	return nil
}
