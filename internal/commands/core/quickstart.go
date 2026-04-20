package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
)

var quickstartAll bool

// quickstartCmd provides one-command setup for tok
var quickstartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "One-command setup for tok",
	Long: `Automatically detect installed AI agents, install hooks,
apply sensible defaults, and verify the setup works.

This is the fastest way to get started with tok:
  - Detects Claude Code, Cursor, Windsurf, Cline, etc.
  - Installs appropriate hooks for each agent
  - Creates default configuration
  - Runs doctor to verify setup`,
	RunE: runQuickstart,
}

func init() {
	quickstartCmd.Flags().BoolVarP(&quickstartAll, "all", "a", false, "setup for all detected agents")
	registry.Add(func() { registry.Register(quickstartCmd) })
}

type agentInfo struct {
	Name       string
	DetectDir  string
	ConfigDir  string
	MarkerPath string
	Detected   bool
	Configured bool
}

func runQuickstart(cmd *cobra.Command, args []string) error {
	if shared.IsVerbose() {
		out.Global().Println("tok Quickstart")
		out.Global().Println("=================")
		out.Global().Println()
	}

	// Step 1: Detect agents
	out.Global().Println("Detecting AI agents...")
	agents := detectAgents()

	detectedCount := 0
	for _, agent := range agents {
		if agent.Detected {
			detectedCount++
			out.Global().Printf("   ✓ %s detected\n", agent.Name)
		}
	}

	if detectedCount == 0 {
		out.Global().Println("   ℹ No AI agents detected in standard locations")
		out.Global().Println()
		out.Global().Println("You can manually run:")
		out.Global().Println("  tok init --claude     # For Claude Code")
		out.Global().Println("  tok init --cursor     # For Cursor")
		out.Global().Println("  tok init --windsurf   # For Windsurf")
		return nil
	}
	out.Global().Println()

	// Step 2: Install hooks
	out.Global().Println("Installing hooks...")
	installedCount := 0
	for _, agent := range agents {
		if agent.Detected {
			if quickstartAll || detectedCount == 1 {
				if err := installHookForAgent(agent); err != nil {
					out.Global().Printf("   ✗ %s: %v\n", agent.Name, err)
				} else {
					installedCount++
					out.Global().Printf("   ✓ %s hook installed\n", agent.Name)
				}
			}
		}
	}

	if installedCount == 0 && detectedCount > 0 && !quickstartAll {
		out.Global().Println("   ℹ Run 'tok quickstart --all' to install hooks for all detected agents")
	}
	out.Global().Println()

	// Step 3: Create default config
	out.Global().Println("Setting up configuration...")
	if err := createDefaultConfig(); err != nil {
		out.Global().Printf("   ✗ Config setup failed: %v\n", err)
	} else {
		out.Global().Println("   ✓ Default configuration applied")
	}
	out.Global().Println()

	// Step 4: Run doctor
	out.Global().Println("Running diagnostics...")
	doctorCmd := exec.Command(tokExecutablePath(), "doctor")
	doctorCmd.Stdout = os.Stdout
	doctorCmd.Stderr = os.Stderr
	if err := doctorCmd.Run(); err != nil {
		out.Global().Println()
		out.Global().Println("WARNING Some issues detected. See above for details.")
		return fmt.Errorf("doctor command failed: %w", err)
	}

	out.Global().Println()
	out.Global().Println("Quickstart complete!")
	out.Global().Println()
	out.Global().Println("tok is now active and will compress CLI output automatically.")
	out.Global().Println()
	out.Global().Println("Quick commands:")
	out.Global().Println("  tok status          # View current stats")
	out.Global().Println("  tok gain            # See token savings")
	out.Global().Println("  tok discover        # Find optimization opportunities")
	return nil
}

func detectAgents() []agentInfo {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}
	cwd, _ := os.Getwd()
	if cwd == "" {
		return nil
	}

	agents := []agentInfo{
		{
			Name:       "Claude Code",
			DetectDir:  home + "/.claude",
			ConfigDir:  home + "/.claude",
			MarkerPath: home + "/.claude/hooks/tok-rewrite.sh",
		},
		{
			Name:       "Cursor",
			DetectDir:  home + "/.cursor",
			ConfigDir:  home + "/.cursor",
			MarkerPath: home + "/.cursor/hooks/tok-rewrite.sh",
		},
		{
			Name:       "Windsurf",
			DetectDir:  home + "/.windsurf",
			ConfigDir:  cwd,
			MarkerPath: filepath.Join(cwd, ".windsurfrules"),
		},
		{
			Name:       "Cline",
			DetectDir:  home + "/.cline",
			ConfigDir:  cwd,
			MarkerPath: filepath.Join(cwd, ".clinerules"),
		},
		{
			Name:       "OpenCode",
			DetectDir:  filepath.Join(home, ".config", "opencode"),
			ConfigDir:  filepath.Join(home, ".config", "opencode"),
			MarkerPath: filepath.Join(home, ".config", "opencode", "plugins", "tok.ts"),
		},
		{
			Name:       "OpenClaw",
			DetectDir:  home + "/.openclaw",
			ConfigDir:  home + "/.openclaw",
			MarkerPath: home + "/.openclaw/hooks/tok-rewrite.sh",
		},
	}

	for i := range agents {
		if _, err := os.Stat(agents[i].DetectDir); err == nil {
			agents[i].Detected = true
		}
		// Check if hook already exists
		if _, err := os.Stat(agents[i].MarkerPath); err == nil || legacyQuickstartHookExists(agents[i].MarkerPath) {
			agents[i].Configured = true
		}
	}

	return agents
}

func installHookForAgent(agent agentInfo) error {
	return setupAgent(AgentInfo{
		Name:      agent.Name,
		ConfigDir: agent.ConfigDir,
		HookDir:   filepath.Dir(agent.MarkerPath),
	}, true)
}

func createDefaultConfig() error {
	configPath := shared.GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Create directory if needed
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists, don't overwrite
	}

	// Write default config
	defaultConfig := `# tok Configuration
# Auto-generated by 'tok quickstart'

[tracking]
enabled = true

[filter]
mode = "minimal"

[pipeline]
# Maximum context tokens
max_context_tokens = 2000000
`
	return os.WriteFile(configPath, []byte(defaultConfig), 0600)
}

func tokExecutablePath() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		return exe
	}
	return "tok"
}

func legacyQuickstartHookExists(hookPath string) bool {
	legacyPath := strings.TrimSuffix(hookPath, "tok-rewrite.sh") + "tok.sh"
	_, err := os.Stat(legacyPath)
	return err == nil
}
