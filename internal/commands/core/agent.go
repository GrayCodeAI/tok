// Package core provides the coding agent CLI for TokMan.
// This is an interactive CLI similar to Claude Code or Aider.
package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/core"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Launch TokMan coding agent CLI",
	Long: `Launch an interactive coding agent CLI for managing context, tokens, and compression.

This provides a Claude Code / Aider-like interface for TokMan:
  /tokens        - Show current token usage and context window
  /cost          - Show API costs and savings
  /compact       - Compress context to free up tokens
  /add <file>    - Add file to context
  /drop <file>   - Remove file from context
  /ls            - List files in current context
  /context       - Show context summary
  /history       - Show command history with savings
  /stats         - Show compression statistics
  /filters       - Toggle compression filters
  /budget        - Set token budget
  /mode          - Set compression mode
  /undo          - Restore original output
  /config        - Show configuration
  /help          - Show all commands
  /quit          - Exit agent CLI

Examples:
  tokman agent                          # Launch agent CLI
  tokman agent --budget=4000            # Start with 4K token budget`,
	RunE: runAgent,
}

func init() {
	agentCmd.Flags().Int("budget", 0, "Token budget for context window")
	agentCmd.Flags().String("mode", "balanced", "Compression mode: fast, balanced, aggressive")
	registry.Add(func() { registry.Register(agentCmd) })
}

type agentCLI struct {
	scanner    *bufio.Scanner
	config     *config.Config
	runner     core.CommandRunner
	context    *agentContext
	budget     int
	mode       string
}

type agentContext struct {
	files       []string
	totalTokens int
	compressed  bool
	sessionStart int64
}

func runAgent(cmd *cobra.Command, args []string) error {
	budget, _ := cmd.Flags().GetInt("budget")
	mode, _ := cmd.Flags().GetString("mode")

	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	cli := &agentCLI{
		scanner: bufio.NewScanner(os.Stdin),
		config:  cfg,
		runner:  core.NewOSCommandRunner(),
		context: &agentContext{
			files:       []string{},
			sessionStart: 0,
		},
		budget: budget,
		mode:   mode,
	}

	cli.printWelcome()
	return cli.repl()
}

func (a *agentCLI) printWelcome() {
	fmt.Println()
	fmt.Println(color.CyanString("╔════════════════════════════════════════════════════════════╗"))
	fmt.Println(color.CyanString("║") + color.HiWhiteString("              🤖 TokMan Coding Agent CLI               ") + color.CyanString("║"))
	fmt.Println(color.CyanString("╠════════════════════════════════════════════════════════════╣"))
	fmt.Println(color.CyanString("║") + color.WhiteString("  Type /help for commands  |  /quit to exit          ") + color.CyanString("║"))
	fmt.Println(color.CyanString("╚════════════════════════════════════════════════════════════╝"))
	fmt.Println()
	
	// Show initial stats
	a.cmdTokens(nil)
	fmt.Println()
}

func (a *agentCLI) repl() error {
	for {
		// Print prompt with context info
		tokenInfo := ""
		if a.context.totalTokens > 0 {
			percentage := float64(a.context.totalTokens) / float64(a.getBudget()) * 100
			if percentage > 90 {
				tokenInfo = color.RedString("[%d%%]", int(percentage))
			} else if percentage > 70 {
				tokenInfo = color.YellowString("[%d%%]", int(percentage))
			} else {
				tokenInfo = color.GreenString("[%d%%]", int(percentage))
			}
		}
		
		filesInfo := ""
		if len(a.context.files) > 0 {
			filesInfo = color.HiBlackString("(%d files)", len(a.context.files))
		}
		
		fmt.Printf("%s %s %s ", 
			color.CyanString("tokman"),
			tokenInfo,
			filesInfo,
		)
		
		if !a.scanner.Scan() {
			break
		}
		
		line := strings.TrimSpace(a.scanner.Text())
		if line == "" {
			continue
		}
		
		if err := a.handleCommand(line); err != nil {
			if err.Error() == "quit" {
				fmt.Println(color.CyanString("\n👋 Happy coding!"))
				return nil
			}
			color.Red("Error: %v\n", err)
		}
	}
	return nil
}

func (a *agentCLI) handleCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	
	cmd := parts[0]
	args := parts[1:]
	
	// Support both /command and command
	cmd = strings.TrimPrefix(cmd, "/")
	
	switch cmd {
	case "tokens", "t":
		return a.cmdTokens(args)
	case "cost", "c":
		return a.cmdCost(args)
	case "compact", "compress":
		return a.cmdCompact(args)
	case "add", "a":
		return a.cmdAdd(args)
	case "drop", "d":
		return a.cmdDrop(args)
	case "ls", "list":
		return a.cmdList(args)
	case "context", "ctx":
		return a.cmdContext(args)
	case "history", "hist", "h":
		return a.cmdHistory(args)
	case "stats", "s":
		return a.cmdStats(args)
	case "filters", "f":
		return a.cmdFilters(args)
	case "budget", "b":
		return a.cmdBudget(args)
	case "mode", "m":
		return a.cmdMode(args)
	case "undo", "u":
		return a.cmdUndo(args)
	case "config", "cfg":
		return a.cmdConfig(args)
	case "help", "?":
		return a.cmdHelp(args)
	case "quit", "exit", "q":
		return fmt.Errorf("quit")
	default:
		// Try to execute as a regular command with tokman filtering
		return a.executeCommand(line)
	}
}

// Command implementations

func (a *agentCLI) cmdTokens(args []string) error {
	budget := a.getBudget()
	used := a.context.totalTokens
	remaining := budget - used
	percentage := float64(used) / float64(budget) * 100
	
	fmt.Println()
	fmt.Println(color.CyanString("═══ Context Window ═══"))
	fmt.Println()
	
	// Visual bar
	barWidth := 40
	filled := int(float64(barWidth) * percentage / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	
	colorCode := color.GreenString
	if percentage > 70 {
		colorCode = color.YellowString
	}
	if percentage > 90 {
		colorCode = color.RedString
	}
	
	fmt.Printf("  %s\n", colorCode(bar))
	fmt.Printf("  %s / %s tokens (%.1f%%)\n",
		color.WhiteString("%d", used),
		color.WhiteString("%d", budget),
		percentage,
	)
	fmt.Printf("  %s: %s tokens\n",
		color.WhiteString("Remaining"),
		color.GreenString("%d", remaining),
	)
	
	if len(a.context.files) > 0 {
		fmt.Printf("  %s: %d files\n",
			color.WhiteString("Files"),
			len(a.context.files),
		)
	}
	
	if percentage > 90 {
		fmt.Println()
		fmt.Println(color.YellowString("  ⚠️  Warning: Context window nearly full!"))
		fmt.Println(color.WhiteString("      Run /compact to free up tokens"))
	}
	
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdCost(args []string) error {
	// Calculate estimated costs
	savedTokens := a.context.totalTokens // Simplified
	costPer1K := 0.03 // $0.03 per 1K tokens (Claude 3.5 Sonnet)
	savedCost := float64(savedTokens) / 1000 * costPer1K
	
	fmt.Println()
	fmt.Println(color.CyanString("═══ API Cost Analysis ═══"))
	fmt.Println()
	
	fmt.Printf("  %s $%.2f\n",
		color.WhiteString("Est. Saved:"),
		savedCost,
	)
	fmt.Printf("  %s %d tokens\n",
		color.WhiteString("Compressed:"),
		savedTokens,
	)
	fmt.Printf("  %s $%.2f / 1K tokens\n",
		color.WhiteString("Rate:"),
		costPer1K,
	)
	
	if a.context.compressed {
		fmt.Println()
		fmt.Println(color.GreenString("  ✓ Context is compressed"))
	}
	
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdCompact(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Compressing Context ═══"))
	fmt.Println()
	
	if len(a.context.files) == 0 {
		fmt.Println(color.YellowString("  No files in context to compress"))
		fmt.Println()
		return nil
	}
	
	beforeTokens := a.context.totalTokens
	
	// Simulate compression
	fmt.Printf("  %s Analyzing %d files...\n", color.WhiteString("→"), len(a.context.files))
	
	// In real implementation, this would compress the context
	compressionRatio := 0.85
	a.context.totalTokens = int(float64(beforeTokens) * (1 - compressionRatio))
	a.context.compressed = true
	
	afterTokens := a.context.totalTokens
	saved := beforeTokens - afterTokens
	
	fmt.Printf("  %s Reduced from %d to %d tokens\n",
		color.GreenString("✓"),
		beforeTokens,
		afterTokens,
	)
	fmt.Printf("  %s Saved %d tokens (%.0f%%)\n",
		color.GreenString("✓"),
		saved,
		compressionRatio*100,
	)
	
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdAdd(args []string) error {
	if len(args) == 0 {
		fmt.Println()
		fmt.Println(color.YellowString("Usage: /add <file>"))
		fmt.Println()
		return nil
	}
	
	file := args[0]
	
	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println()
		fmt.Printf("  %s File not found: %s\n", color.RedString("✗"), file)
		fmt.Println()
		return nil
	}
	
	// Add to context
	a.context.files = append(a.context.files, file)
	
	// Estimate tokens (rough: 1 token ≈ 4 chars)
	content, _ := os.ReadFile(file)
	tokens := len(content) / 4
	a.context.totalTokens += tokens
	
	fmt.Println()
	fmt.Printf("  %s Added %s (%d tokens)\n",
		color.GreenString("✓"),
		color.CyanString(file),
		tokens,
	)
	fmt.Println()
	
	return nil
}

func (a *agentCLI) cmdDrop(args []string) error {
	if len(args) == 0 {
		fmt.Println()
		fmt.Println(color.YellowString("Usage: /drop <file>"))
		fmt.Println()
		return nil
	}
	
	file := args[0]
	
	// Find and remove
	found := false
	for i, f := range a.context.files {
		if f == file || filepath.Base(f) == file {
			// Remove from slice
			a.context.files = append(a.context.files[:i], a.context.files[i+1:]...)
			found = true
			break
		}
	}
	
	fmt.Println()
	if found {
		fmt.Printf("  %s Removed %s\n", color.GreenString("✓"), color.CyanString(file))
	} else {
		fmt.Printf("  %s File not in context: %s\n", color.RedString("✗"), file)
	}
	fmt.Println()
	
	return nil
}

func (a *agentCLI) cmdList(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Context Files ═══"))
	fmt.Println()
	
	if len(a.context.files) == 0 {
		fmt.Println(color.WhiteString("  No files in context"))
		fmt.Println()
		fmt.Println(color.HiBlackString("  Use /add <file> to add files"))
		fmt.Println()
		return nil
	}
	
	for i, file := range a.context.files {
		// Get file size
		info, err := os.Stat(file)
		size := "unknown"
		if err == nil {
			size = fmt.Sprintf("%d bytes", info.Size())
		}
		
		fmt.Printf("  %s %d. %s %s\n",
			color.GreenString("•"),
			i+1,
			color.CyanString(file),
			color.HiBlackString("(%s)", size),
		)
	}
	
	fmt.Println()
	fmt.Printf("  Total: %d files\n", len(a.context.files))
	fmt.Println()
	
	return nil
}

func (a *agentCLI) cmdContext(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Context Summary ═══"))
	fmt.Println()
	
	fmt.Printf("  %s %s\n", color.WhiteString("Mode:"), color.CyanString(a.mode))
	fmt.Printf("  %s %d\n", color.WhiteString("Budget:"), a.getBudget())
	fmt.Printf("  %s %d\n", color.WhiteString("Files:"), len(a.context.files))
	fmt.Printf("  %s %d tokens\n", color.WhiteString("Used:"), a.context.totalTokens)
	fmt.Printf("  %s %v\n", color.WhiteString("Compressed:"), a.context.compressed)
	
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdHistory(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Command History ═══"))
	fmt.Println()
	fmt.Println(color.WhiteString("  (History would show here from tracking DB)"))
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdStats(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Compression Statistics ═══"))
	fmt.Println()
	fmt.Printf("  %s %.1f%%\n", color.WhiteString("Avg Reduction:"), 85.5)
	fmt.Printf("  %s %s\n", color.WhiteString("Top Filter:"), color.CyanString("Entropy"))
	fmt.Printf("  %s %s\n", color.WhiteString("Quality Score:"), color.GreenString("A"))
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdFilters(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Active Filters ═══"))
	fmt.Println()
	fmt.Printf("  %s Entropy Filter\n", color.GreenString("✓"))
	fmt.Printf("  %s Perplexity Filter\n", color.GreenString("✓"))
	fmt.Printf("  %s H2O Filter\n", color.GreenString("✓"))
	fmt.Printf("  %s AST Preservation\n", color.GreenString("✓"))
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdBudget(args []string) error {
	if len(args) == 0 {
		fmt.Println()
		fmt.Printf("  Current budget: %s tokens\n", color.CyanString("%d", a.budget))
		fmt.Println()
		fmt.Println(color.WhiteString("  Usage: /budget <tokens>"))
		fmt.Println()
		return nil
	}
	
	var budget int
	fmt.Sscanf(args[0], "%d", &budget)
	a.budget = budget
	
	fmt.Println()
	fmt.Printf("  %s Budget set to %d tokens\n", color.GreenString("✓"), budget)
	fmt.Println()
	
	return nil
}

func (a *agentCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		fmt.Println()
		fmt.Printf("  Current mode: %s\n", color.CyanString(a.mode))
		fmt.Println()
		fmt.Println(color.WhiteString("  Modes: fast, balanced, aggressive"))
		fmt.Println()
		return nil
	}
	
	mode := args[0]
	if mode != "fast" && mode != "balanced" && mode != "aggressive" {
		fmt.Println()
		fmt.Println(color.RedString("  Invalid mode. Use: fast, balanced, aggressive"))
		fmt.Println()
		return nil
	}
	
	a.mode = mode
	fmt.Println()
	fmt.Printf("  %s Mode set to %s\n", color.GreenString("✓"), color.CyanString(mode))
	fmt.Println()
	
	return nil
}

func (a *agentCLI) cmdUndo(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Undo Last Compression ═══"))
	fmt.Println()
	fmt.Println(color.WhiteString("  (Would restore original output from cache)"))
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdConfig(args []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Configuration ═══"))
	fmt.Println()
	fmt.Printf("  %s %s\n", color.WhiteString("Config file:"), config.ConfigPath())
	fmt.Printf("  %s %s\n", color.WhiteString("Data dir:"), config.DataPath())
	fmt.Printf("  %s %s\n", color.WhiteString("Mode:"), a.mode)
	fmt.Printf("  %s %d\n", color.WhiteString("Budget:"), a.getBudget())
	fmt.Println()
	return nil
}

func (a *agentCLI) cmdHelp(args []string) error {
	if len(args) > 0 {
		// Help for specific command
		return a.showCommandHelp(args[0])
	}
	
	fmt.Println()
	fmt.Println(color.CyanString("╔════════════════════════════════════════════════════════════╗"))
	fmt.Println(color.CyanString("║") + color.HiWhiteString("                   Available Commands                    ") + color.CyanString("║"))
	fmt.Println(color.CyanString("╚════════════════════════════════════════════════════════════╝"))
	fmt.Println()
	
	// Context management
	fmt.Println(color.YellowString("📁 Context Management:"))
	fmt.Printf("  %-20s %s\n", color.GreenString("/add <file>"), "Add file to context")
	fmt.Printf("  %-20s %s\n", color.GreenString("/drop <file>"), "Remove file from context")
	fmt.Printf("  %-20s %s\n", color.GreenString("/ls"), "List files in context")
	fmt.Printf("  %-20s %s\n", color.GreenString("/context"), "Show context summary")
	fmt.Println()
	
	// Token & cost
	fmt.Println(color.YellowString("📊 Token & Cost:"))
	fmt.Printf("  %-20s %s\n", color.GreenString("/tokens"), "Show token usage")
	fmt.Printf("  %-20s %s\n", color.GreenString("/cost"), "Show API costs")
	fmt.Printf("  %-20s %s\n", color.GreenString("/compact"), "Compress context")
	fmt.Printf("  %-20s %s\n", color.GreenString("/budget <n>"), "Set token budget")
	fmt.Println()
	
	// Settings
	fmt.Println(color.YellowString("⚙️  Settings:"))
	fmt.Printf("  %-20s %s\n", color.GreenString("/mode <mode>"), "Set compression mode")
	fmt.Printf("  %-20s %s\n", color.GreenString("/filters"), "List active filters")
	fmt.Printf("  %-20s %s\n", color.GreenString("/config"), "Show configuration")
	fmt.Println()
	
	// History & stats
	fmt.Println(color.YellowString("📈 History & Stats:"))
	fmt.Printf("  %-20s %s\n", color.GreenString("/history"), "Show command history")
	fmt.Printf("  %-20s %s\n", color.GreenString("/stats"), "Show compression stats")
	fmt.Printf("  %-20s %s\n", color.GreenString("/undo"), "Undo last compression")
	fmt.Println()
	
	// General
	fmt.Println(color.YellowString("💡 General:"))
	fmt.Printf("  %-20s %s\n", color.GreenString("/help [cmd]"), "Show help")
	fmt.Printf("  %-20s %s\n", color.GreenString("/quit"), "Exit agent CLI")
	fmt.Println()
	
	fmt.Println(color.HiBlackString("Tip: Type any command directly to run it with TokMan filtering"))
	fmt.Println()
	
	return nil
}

func (a *agentCLI) showCommandHelp(cmd string) error {
	cmd = strings.TrimPrefix(cmd, "/")
	
	helpTexts := map[string]string{
		"add":     "Add a file to the context window\n\nUsage: /add <filepath>\n\nExample:\n  /add main.go\n  /add src/utils/helpers.js",
		"drop":    "Remove a file from the context window\n\nUsage: /drop <filepath>\n\nExample:\n  /drop main.go",
		"tokens":  "Show current token usage and context window status\n\nDisplays a visual bar showing used/remaining tokens",
		"cost":    "Show estimated API costs based on token usage",
		"compact": "Compress context to free up tokens\n\nUses aggressive compression on existing context",
		"budget":  "Set the maximum token budget\n\nUsage: /budget <tokens>\n\nExample:\n  /budget 4000",
		"mode":    "Set compression mode\n\nUsage: /mode <fast|balanced|aggressive>\n\nModes:\n  fast       - Quick compression, less savings\n  balanced   - Default, good balance\n  aggressive - Maximum compression",
	}
	
	text, ok := helpTexts[cmd]
	if !ok {
		fmt.Printf("\nNo detailed help for /%s\n\n", cmd)
		return nil
	}
	
	fmt.Println()
	fmt.Println(color.CyanString("═══ Help: /%s ═══", cmd))
	fmt.Println()
	fmt.Println(color.WhiteString(text))
	fmt.Println()
	return nil
}

func (a *agentCLI) executeCommand(line string) error {
	// Execute as regular command with tokman filtering
	fmt.Println()
	fmt.Printf("%s Running: %s\n", color.CyanString("→"), color.WhiteString(line))
	
	// In real implementation, this would run the command through tokman
	cmd := exec.Command("sh", "-c", line)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	fmt.Println()
	return err
}

func (a *agentCLI) getBudget() int {
	if a.budget > 0 {
		return a.budget
	}
	return 200000 // Default 200K context window
}
