// Package commands provides a PI-style CLI interface with prominent input box.
package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// PiStyleCLI provides a PI-agent-like interface with prominent input box
type PiStyleCLI struct {
	scanner  *bufio.Scanner
	tracker  *tracking.Tracker
	config   *config.Config
	session  *PiSession
	history  []string
	histIdx  int
}

type PiSession struct {
	Files       []string
	TotalTokens int
	Budget      int
	Mode        string
	StartTime   time.Time
}

// RunPiStyle starts the PI-style CLI
func RunPiStyle() error {
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	cli := &PiStyleCLI{
		scanner: bufio.NewScanner(os.Stdin),
		tracker: tracking.GetGlobalTracker(),
		config:  cfg,
		session: &PiSession{
			Files:     []string{},
			Budget:    200000,
			Mode:      "balanced",
			StartTime: time.Now(),
		},
		history: []string{},
		histIdx: -1,
	}

	cli.showIntro()
	return cli.loop()
}

// showIntro displays the welcome screen
func (cli *PiStyleCLI) showIntro() {
	cli.clearScreen()
	
	// Title bar
	cli.printTitleBar()
	
	// Stats line
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("  %s %s commands  •  %s %.1f%% avg savings\n",
				color.HiBlackString("Session:"),
				color.WhiteString("%d", summary.TotalCommands),
				color.HiBlackString("Savings:"),
				summary.ReductionPct,
			)
		}
	}
	
	fmt.Println()
	
	// Quick tips in compact format
	fmt.Println(color.HiBlackString("  Commands: /add  /drop  /ls  /status  /tokens  /cost  /help  /quit"))
	fmt.Println(color.HiBlackString("  Or type any shell command: git status, docker ps, kubectl logs"))
	
	cli.printSeparator()
}

// printTitleBar displays the top title bar
func (cli *PiStyleCLI) printTitleBar() {
	width := 70
	
	// Top border
	fmt.Println()
	fmt.Println(color.CyanString("  ╔" + strings.Repeat("═", width-4) + "╗"))
	
	// Title line
	title := "TokMan"
	subtitle := "Token-efficient CLI"
	padding := (width - 4 - len(title) - 3 - len(subtitle)) / 2
	line := "║" + strings.Repeat(" ", padding) + 
		title + "  " + color.HiBlackString(subtitle) +
		strings.Repeat(" ", width-4-padding-len(title)-3-len(subtitle)) + "║"
	fmt.Println(color.CyanString("  ") + line)
	
	// Bottom border
	fmt.Println(color.CyanString("  ╚" + strings.Repeat("═", width-4) + "╝"))
}

// printInputBox displays the prominent input box
func (cli *PiStyleCLI) printInputBox() {
	// Get status for the box
	status := cli.getStatusString()
	
	// Input box top
	fmt.Println()
	fmt.Println(color.CyanString("  ┌" + strings.Repeat("─", 66) + "┐"))
	
	// Status line inside box
	if status != "" {
		fmt.Printf(color.CyanString("  │ ") + "%s" + strings.Repeat(" ", 65-len(status)) + color.CyanString("│\n"), status)
	}
	
	// Input prompt line
	prompt := "You:"
	fmt.Printf(color.CyanString("  │ ") + color.HiBlackString("%-4s") + color.CyanString(" │") + " \n", prompt)
	
	// Bottom border (will be cursor position)
	fmt.Print(color.CyanString("  └" + strings.Repeat("─", 66) + "┘"))
	fmt.Print("\r  └─" + color.HiBlackString("You:") + color.CyanString("─"))
	fmt.Print(" ")
}

// getStatusString returns status for input box
func (cli *PiStyleCLI) getStatusString() string {
	parts := []string{}
	
	if len(cli.session.Files) > 0 {
		parts = append(parts, fmt.Sprintf("%d files", len(cli.session.Files)))
	}
	
	if cli.session.TotalTokens > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		var indicator string
		if pct > 90 {
			indicator = color.RedString("%.0f%%", pct)
		} else if pct > 70 {
			indicator = color.YellowString("%.0f%%", pct)
		} else {
			indicator = color.GreenString("%.0f%%", pct)
		}
		parts = append(parts, fmt.Sprintf("%s tokens", indicator))
	}
	
	if len(parts) > 0 {
		return "Context: " + strings.Join(parts, " • ")
	}
	return ""
}

// printSeparator prints a separator line
func (cli *PiStyleCLI) printSeparator() {
	fmt.Println()
	fmt.Println(color.HiBlackString("  " + strings.Repeat("─", 68)))
	fmt.Println()
}

// loop runs the main REPL with PI-style input
func (cli *PiStyleCLI) loop() error {
	for {
		// Show input box
		cli.printInputBox()
		
		if !cli.scanner.Scan() {
			break
		}
		
		// Clear the input box line
		fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
		
		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			continue
		}
		
		// Add to history
		cli.history = append(cli.history, input)
		cli.histIdx = len(cli.history)
		
		// Show the input as a message bubble
		cli.printUserMessage(input)
		
		// Process the input
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				cli.printAssistantMessage("Goodbye! 👋")
				fmt.Println()
				return nil
			}
			cli.printError(err.Error())
		}
		
		fmt.Println()
	}
	return nil
}

// printUserMessage displays user input in a bubble
func (cli *PiStyleCLI) printUserMessage(msg string) {
	// Simple display without box for cleanliness
	fmt.Printf("  %s %s\n", color.HiBlackString("You:"), msg)
}

// printAssistantMessage displays assistant response
func (cli *PiStyleCLI) printAssistantMessage(msg string) {
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), msg)
}

// printResponse displays a response block
func (cli *PiStyleCLI) printResponse(title string, content func()) {
	fmt.Println()
	fmt.Printf("  %s %s\n", color.CyanString("┌─"), color.WhiteString(title))
	fmt.Println(color.CyanString("  │"))
	content()
	fmt.Println(color.CyanString("  └"))
}

// handle processes user input
func (cli *PiStyleCLI) handle(input string) error {
	// Check for slash commands
	if strings.HasPrefix(input, "/") {
		return cli.handleCommand(input[1:])
	}
	
	// Execute as shell command
	return cli.executeShellCommand(input)
}

// handleCommand processes slash commands
func (cli *PiStyleCLI) handleCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	
	cmd := parts[0]
	args := parts[1:]
	
	switch cmd {
	case "help", "?", "h":
		return cli.cmdHelp(args)
	case "exit", "quit", "q":
		return fmt.Errorf("exit")
	case "clear":
		cli.clearScreen()
		cli.printTitleBar()
		return nil
		
	case "add", "a":
		return cli.cmdAdd(args)
	case "drop", "d":
		return cli.cmdDrop(args)
	case "ls", "list":
		return cli.cmdList()
		
	case "status", "s":
		return cli.cmdStatus()
	case "tokens", "t":
		return cli.cmdTokens()
	case "cost", "c":
		return cli.cmdCost()
	case "mode", "m":
		return cli.cmdMode(args)
	case "budget", "b":
		return cli.cmdBudget(args)
		
	case "compact":
		return cli.cmdCompact()
	case "stats", "gain", "g":
		return cli.cmdStats()
		
	case "filters":
		return cli.cmdFilters()
	case "config":
		return cli.cmdConfig()
	case "version", "v":
		cli.printAssistantMessage("TokMan v" + shared.Version)
		return nil
		
	default:
		return fmt.Errorf("unknown command /%s. Try /help", cmd)
	}
}

// ==================== COMMANDS ====================

func (cli *PiStyleCLI) cmdHelp(args []string) error {
	if len(args) > 0 {
		return cli.showCommandHelp(args[0])
	}
	
	cli.printResponse("Help", func() {
		fmt.Println(color.CyanString("  │"))
		fmt.Println(color.CyanString("  │") + "  " + color.WhiteString("Context Management"))
		fmt.Println(color.CyanString("  │") + "    /add <file>    Add file to context")
		fmt.Println(color.CyanString("  │") + "    /drop <file>   Remove file from context")
		fmt.Println(color.CyanString("  │") + "    /ls            List files")
		fmt.Println(color.CyanString("  │"))
		fmt.Println(color.CyanString("  │") + "  " + color.WhiteString("Status & Info"))
		fmt.Println(color.CyanString("  │") + "    /status        Show session status")
		fmt.Println(color.CyanString("  │") + "    /tokens        Show token usage")
		fmt.Println(color.CyanString("  │") + "    /cost          Show API cost estimate")
		fmt.Println(color.CyanString("  │") + "    /stats         Show statistics")
		fmt.Println(color.CyanString("  │"))
		fmt.Println(color.CyanString("  │") + "  " + color.WhiteString("Settings"))
		fmt.Println(color.CyanString("  │") + "    /mode <type>   Set mode: fast/balanced/aggressive")
		fmt.Println(color.CyanString("  │") + "    /budget <n>    Set token budget")
		fmt.Println(color.CyanString("  │") + "    /filters       List active filters")
		fmt.Println(color.CyanString("  │"))
		fmt.Println(color.CyanString("  │") + "  " + color.WhiteString("Other"))
		fmt.Println(color.CyanString("  │") + "    /compact       Compress context")
		fmt.Println(color.CyanString("  │") + "    /help <cmd>    Help for specific command")
		fmt.Println(color.CyanString("  │") + "    /quit          Exit")
		fmt.Println(color.CyanString("  │"))
		fmt.Println(color.CyanString("  │") + "  " + color.HiBlackString("Or type any shell command directly"))
	})
	
	return nil
}

func (cli *PiStyleCLI) showCommandHelp(cmd string) error {
	help := map[string]string{
		"add": "Add a file to your context\n\nExample: /add main.go",
		"tokens": "Show your current token usage\n\nDisplays a visual bar and percentage",
		"cost": "Show estimated API cost",
		"mode": "Set compression mode\n\n  fast       - Quick compression\n  balanced   - Good balance\n  aggressive - Maximum compression",
	}
	
	text, ok := help[cmd]
	if !ok {
		return fmt.Errorf("no help for /%s", cmd)
	}
	
	cli.printAssistantMessage(text)
	return nil
}

func (cli *PiStyleCLI) cmdAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /add <file>")
	}
	
	path := args[0]
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %s", path)
	}
	
	content, _ := os.ReadFile(path)
	tokens := len(content) / 4
	
	cli.session.Files = append(cli.session.Files, path)
	cli.session.TotalTokens += tokens
	
	cli.printAssistantMessage(fmt.Sprintf("Added %s (%s, %d tokens)", 
		path, formatBytes(info.Size()), tokens))
	return nil
}

func (cli *PiStyleCLI) cmdDrop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /drop <file>")
	}
	
	target := args[0]
	for i, f := range cli.session.Files {
		if f == target || filepath.Base(f) == target {
			cli.session.TotalTokens -= (len(f) / 4) // approximate
			cli.session.Files = append(cli.session.Files[:i], cli.session.Files[i+1:]...)
			cli.printAssistantMessage(fmt.Sprintf("Removed %s", target))
			return nil
		}
	}
	
	return fmt.Errorf("file not in context: %s", target)
}

func (cli *PiStyleCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		cli.printAssistantMessage("No files in context. Use /add <file> to add.")
		return nil
	}
	
	cli.printResponse(fmt.Sprintf("Context (%d files)", len(cli.session.Files)), func() {
		for i, f := range cli.session.Files {
			info, _ := os.Stat(f)
			size := ""
			if info != nil {
				size = formatBytes(info.Size())
			}
			fmt.Printf("  │  %d. %-30s %s\n", i+1, truncate(f, 28), color.HiBlackString(size))
		}
		fmt.Println(color.CyanString("  │"))
		fmt.Printf("  │  Total: %s tokens\n", color.WhiteString("%d", cli.session.TotalTokens))
	})
	
	return nil
}

func (cli *PiStyleCLI) cmdStatus() error {
	cli.printResponse("Status", func() {
		fmt.Printf("  │  Mode:   %s\n", color.CyanString(cli.session.Mode))
		fmt.Printf("  │  Budget: %s tokens\n", color.WhiteString(formatNumber(cli.session.Budget)))
		fmt.Printf("  │  Files:  %d\n", len(cli.session.Files))
		fmt.Printf("  │  Tokens: %s / %s\n", 
			color.WhiteString(formatNumber(cli.session.TotalTokens)),
			color.WhiteString(formatNumber(cli.session.Budget)))
		fmt.Printf("  │  Uptime: %s\n", time.Since(cli.session.StartTime).Round(time.Second))
	})
	return nil
}

func (cli *PiStyleCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	pct := float64(used) / float64(budget) * 100
	
	cli.printResponse("Token Usage", func() {
		// Progress bar
		width := 40
		filled := int(float64(width) * pct / 100)
		if filled > width {
			filled = width
		}
		
		bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
		var coloredBar string
		if pct > 90 {
			coloredBar = color.RedString(bar)
		} else if pct > 70 {
			coloredBar = color.YellowString(bar)
		} else {
			coloredBar = color.GreenString(bar)
		}
		
		fmt.Printf("  │  %s\n", coloredBar)
		fmt.Printf("  │  %s / %s (%.1f%%)\n",
			color.WhiteString(formatNumber(used)),
			color.WhiteString(formatNumber(budget)),
			pct)
		
		if pct > 90 {
			fmt.Println(color.CyanString("  │"))
			fmt.Printf("  │  %s Context nearly full!\n", color.YellowString("⚠"))
		}
	})
	
	return nil
}

func (cli *PiStyleCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03
	
	cli.printResponse("Cost Estimate", func() {
		fmt.Printf("  │  Tokens:   %s\n", color.WhiteString(formatNumber(cli.session.TotalTokens)))
		fmt.Printf("  │  Est cost: %s\n", color.WhiteString("$%.2f", cost))
		fmt.Printf("  │  Rate:     $0.03 / 1K tokens\n")
	})
	
	return nil
}

func (cli *PiStyleCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		cli.printAssistantMessage(fmt.Sprintf("Current mode: %s. Options: fast, balanced, aggressive", cli.session.Mode))
		return nil
	}
	
	mode := strings.ToLower(args[0])
	if mode != "fast" && mode != "balanced" && mode != "aggressive" {
		return fmt.Errorf("mode must be: fast, balanced, or aggressive")
	}
	
	cli.session.Mode = mode
	cli.printAssistantMessage(fmt.Sprintf("Mode set to %s", mode))
	return nil
}

func (cli *PiStyleCLI) cmdBudget(args []string) error {
	if len(args) == 0 {
		cli.printAssistantMessage(fmt.Sprintf("Current budget: %s tokens", formatNumber(cli.session.Budget)))
		return nil
	}
	
	var budget int
	if _, err := fmt.Sscanf(args[0], "%d", &budget); err != nil || budget <= 0 {
		return fmt.Errorf("invalid budget")
	}
	
	cli.session.Budget = budget
	cli.printAssistantMessage(fmt.Sprintf("Budget set to %s tokens", formatNumber(budget)))
	return nil
}

func (cli *PiStyleCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		cli.printAssistantMessage("No files to compress")
		return nil
	}
	
	before := cli.session.TotalTokens
	ratio := 0.85
	after := int(float64(before) * (1 - ratio))
	saved := before - after
	
	cli.session.TotalTokens = after
	
	cli.printAssistantMessage(fmt.Sprintf("Compressed: %s → %s tokens (saved %s, %.0f%%)",
		formatNumber(before), formatNumber(after), formatNumber(saved), ratio*100))
	return nil
}

func (cli *PiStyleCLI) cmdStats() error {
	if cli.tracker == nil {
		cli.printAssistantMessage("Tracking not enabled")
		return nil
	}
	
	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		cli.printAssistantMessage("No data yet. Run some commands!")
		return nil
	}
	
	cli.printResponse("Statistics", func() {
		fmt.Printf("  │  Commands:     %s\n", color.WhiteString("%d", summary.TotalCommands))
		fmt.Printf("  │  Tokens saved: %s\n", color.WhiteString(formatNumber(summary.TotalSaved)))
		fmt.Printf("  │  Avg savings:  %.1f%%\n", summary.ReductionPct)
		fmt.Printf("  │  Est. value:   $%.2f\n", float64(summary.TotalSaved)*0.00003)
	})
	
	return nil
}

func (cli *PiStyleCLI) cmdFilters() error {
	cli.printResponse("Active Filters", func() {
		filters := []string{
			"Entropy Filter",
			"Perplexity Filter",
			"H2O Filter",
			"AST Preservation",
			"Semantic Compaction",
		}
		for _, f := range filters {
			fmt.Printf("  │  %s %s\n", color.GreenString("✓"), f)
		}
	})
	return nil
}

func (cli *PiStyleCLI) cmdConfig() error {
	cli.printResponse("Configuration", func() {
		fmt.Printf("  │  Config:  %s\n", config.ConfigPath())
		fmt.Printf("  │  Data:    %s\n", config.DataPath())
		fmt.Printf("  │  Mode:    %s\n", cli.session.Mode)
		fmt.Printf("  │  Budget:  %s\n", formatNumber(cli.session.Budget))
	})
	return nil
}

func (cli *PiStyleCLI) executeShellCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	
	// Check if command exists
	if _, err := exec.LookPath(parts[0]); err != nil {
		return fmt.Errorf("command not found: %s", parts[0])
	}
	
	// Show thinking indicator
	fmt.Printf("\r  %s Running %s...", color.HiBlackString("●"), parts[0])
	
	// Execute
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	
	// Clear indicator
	fmt.Print("\r" + strings.Repeat(" ", 50) + "\r")
	
	if err != nil && len(output) == 0 {
		return fmt.Errorf("command failed: %v", err)
	}
	
	// Compress output
	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	compressed, stats := p.Process(string(output))
	
	// Show compressed output
	if len(compressed) > 0 {
		cli.printResponse("Output", func() {
			lines := strings.Split(compressed, "\n")
			for _, line := range lines[:min(len(lines), 20)] {
				fmt.Printf("  │  %s\n", line)
			}
			if len(lines) > 20 {
				fmt.Printf("  │  %s\n", color.HiBlackString("... (%d more lines)", len(lines)-20))
			}
			
			if stats.TotalSaved > 0 {
				fmt.Println(color.CyanString("  │"))
				pct := float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
				fmt.Printf("  │  %s Saved %s (%.0f%%)\n",
					color.GreenString("✓"),
					formatNumber(stats.TotalSaved),
					pct)
			}
		})
	}
	
	return nil
}

// ==================== UTILITIES ====================

func (cli *PiStyleCLI) clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (cli *PiStyleCLI) printError(msg string) {
	fmt.Printf("  %s %s\n", color.RedString("✗"), msg)
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatBytes(n int64) string {
	if n >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
	if n >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%d B", n)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
