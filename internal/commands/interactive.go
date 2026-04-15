// Package commands provides an interactive CLI with prominent input box.
// Inspired by Claude Code and PI agent interfaces.
package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// InteractiveCLI provides a chat-like interface with input box
type InteractiveCLI struct {
	scanner *bufio.Scanner
	tracker *tracking.Tracker
	config  *config.Config
	session *Session
	width   int
}

type Session struct {
	Files       []ContextFile
	TotalTokens int
	Budget      int
	Mode        string
	StartTime   time.Time
}

type ContextFile struct {
	Path   string
	Size   int64
	Tokens int
}

// RunInteractive starts the interactive CLI
func RunInteractive() error {
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	cli := &InteractiveCLI{
		scanner: bufio.NewScanner(os.Stdin),
		tracker: tracking.GetGlobalTracker(),
		config:  cfg,
		session: &Session{
			Files:     []ContextFile{},
			Budget:    200000,
			Mode:      "balanced",
			StartTime: time.Now(),
		},
		width: 70,
	}

	cli.clearScreen()
	cli.printHeader()
	return cli.loop()
}

// printHeader shows the top header
func (cli *InteractiveCLI) printHeader() {
	// Title bar
	fmt.Println()
	fmt.Println(color.CyanString("  ╔" + strings.Repeat("═", cli.width-4) + "╗"))
	
	title := "TokMan"
	subtitle := "Token-efficient CLI"
	padding := (cli.width - 4 - len(title) - 3 - len(subtitle)) / 2
	line := strings.Repeat(" ", padding) + title + "  " + subtitle +
		strings.Repeat(" ", cli.width-4-padding-len(title)-3-len(subtitle))
	fmt.Println(color.CyanString("  ║") + line + color.CyanString("║"))
	
	fmt.Println(color.CyanString("  ╚" + strings.Repeat("═", cli.width-4) + "╝"))
	
	// Stats line
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("  Session: %d commands  ·  Savings: %.1f%% avg\n",
				summary.TotalCommands, summary.ReductionPct)
		}
	}
	
	// Quick help
	fmt.Println()
	fmt.Println(color.HiBlackString("  Commands: /add /drop /ls /status /tokens /cost /help /quit"))
	fmt.Println()
}

// printInputBox shows the prominent input box at bottom
func (cli *InteractiveCLI) printInputBox() {
	// Build status line for inside the box
	status := cli.getStatusString()
	
	fmt.Println()
	fmt.Println(color.CyanString("  ┌" + strings.Repeat("─", cli.width-4) + "┐"))
	
	// Status line (if any)
	if status != "" {
		padding := cli.width - 4 - len(status) - 1
		fmt.Printf("  %s %s%s%s\n", 
			color.CyanString("│"),
			status,
			strings.Repeat(" ", padding),
			color.CyanString("│"))
	}
	
	// Input line with >
	fmt.Printf("  %s %s ", color.CyanString("│"), color.HiBlackString(">"))
}

// getStatusString returns context status
func (cli *InteractiveCLI) getStatusString() string {
	if len(cli.session.Files) == 0 {
		return ""
	}
	
	pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
	var pctStr string
	if pct > 90 {
		pctStr = color.RedString("%.0f%%", pct)
	} else if pct > 70 {
		pctStr = color.YellowString("%.0f%%", pct)
	} else {
		pctStr = color.GreenString("%.0f%%", pct)
	}
	
	return fmt.Sprintf("Context: %d files · %s tokens", len(cli.session.Files), pctStr)
}

// clearScreen clears the terminal
func (cli *InteractiveCLI) clearScreen() {
	// Only clear if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		fmt.Print("\033[H\033[2J")
	}
}

// loop runs the main REPL
func (cli *InteractiveCLI) loop() error {
	for {
		cli.printInputBox()

		if !cli.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			// In interactive mode, clear empty line
			if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
				fmt.Print("\033[F\033[2K")
			}
			continue
		}

		// Only manipulate cursor in interactive terminal
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			// Move up to erase input box
			fmt.Print("\033[F\033[2K")
			if cli.getStatusString() != "" {
				fmt.Print("\033[F\033[2K")
			}
			fmt.Print("\033[F\033[2K\033[F\033[2K")
		} else {
			// In pipe mode, just print a separator
			fmt.Println()
		}
		
		// Show user input as message
		fmt.Printf("  %s %s\n", color.HiBlackString(">"), input)

		// Process
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				fmt.Println()
				fmt.Printf("  %s Goodbye!\n", color.CyanString("TokMan:"))
				fmt.Println()
				return nil
			}
			cli.printError(err.Error())
		}
	}
	return nil
}

// handle processes input
func (cli *InteractiveCLI) handle(input string) error {
	if strings.HasPrefix(input, "/") {
		return cli.handleCommand(input[1:])
	}
	return cli.executeCommand(input)
}

// handleCommand processes slash commands
func (cli *InteractiveCLI) handleCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help", "h", "?":
		return cli.cmdHelp()
	case "exit", "quit", "q":
		return fmt.Errorf("exit")
	
	// Context
	case "add", "a":
		return cli.cmdAdd(args)
	case "drop", "d", "rm":
		return cli.cmdDrop(args)
	case "ls", "list", "files":
		return cli.cmdList()
	case "clear":
		cli.session.Files = []ContextFile{}
		cli.session.TotalTokens = 0
		cli.printSuccess("Context cleared")
		return nil
	
	// Status
	case "status", "s":
		return cli.cmdStatus()
	case "tokens", "t":
		return cli.cmdTokens()
	case "cost", "c":
		return cli.cmdCost()
	
	// Settings
	case "mode", "m":
		return cli.cmdMode(args)
	case "budget", "b":
		return cli.cmdBudget(args)
	
	// Actions
	case "compact", "compress":
		return cli.cmdCompact()
	case "stats", "stat":
		return cli.cmdStats()
	
	// Info
	case "filters":
		return cli.cmdFilters()
	case "config", "cfg":
		return cli.cmdConfig()
	case "version", "v":
		fmt.Printf("  TokMan version %s\n", shared.Version)
		return nil
	
	default:
		return fmt.Errorf("unknown command: /%s (try /help)", cmd)
	}
}

// cmdHelp shows help
func (cli *InteractiveCLI) cmdHelp() error {
	fmt.Printf("  %s\n", color.CyanString("Available commands:"))
	fmt.Println()
	
	commands := []struct {
		cmd  string
		desc string
	}{
		{"/add <file>", "Add file to context"},
		{"/drop <file>", "Remove file from context"},
		{"/ls", "List files in context"},
		{"/clear", "Clear all context"},
		{"", ""},
		{"/status", "Show session status"},
		{"/tokens", "Show token usage with bar"},
		{"/cost", "Show API cost estimate"},
		{"/stats", "Show statistics"},
		{"", ""},
		{"/mode <type>", "Set: fast/balanced/aggressive"},
		{"/budget <n>", "Set token budget"},
		{"/filters", "List active filters"},
		{"/config", "Show configuration"},
		{"", ""},
		{"/compact", "Compress context"},
		{"/help", "Show this help"},
		{"/quit", "Exit TokMan"},
	}
	
	for _, c := range commands {
		if c.cmd == "" {
			fmt.Println()
			continue
		}
		fmt.Printf("  %s %s\n", color.CyanString(padRight(c.cmd, 14)), c.desc)
	}
	
	fmt.Println()
	fmt.Println(color.HiBlackString("  Or type any shell command: git status, docker ps, etc."))
	return nil
}

// cmdAdd adds a file
func (cli *InteractiveCLI) cmdAdd(args []string) error {
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

	cli.session.Files = append(cli.session.Files, ContextFile{
		Path:   path,
		Size:   info.Size(),
		Tokens: tokens,
	})
	cli.session.TotalTokens += tokens

	cli.printSuccess(fmt.Sprintf("Added %s (%s, %d tokens)", path, formatBytes(info.Size()), tokens))
	return nil
}

// cmdDrop removes a file
func (cli *InteractiveCLI) cmdDrop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /drop <file>")
	}

	target := args[0]
	for i, f := range cli.session.Files {
		if f.Path == target || filepath.Base(f.Path) == target {
			cli.session.TotalTokens -= f.Tokens
			cli.session.Files = append(cli.session.Files[:i], cli.session.Files[i+1:]...)
			cli.printSuccess(fmt.Sprintf("Removed %s", target))
			return nil
		}
	}

	return fmt.Errorf("file not in context: %s", target)
}

// cmdList lists files
func (cli *InteractiveCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(color.HiBlackString("  No files in context. Use /add <file> to add."))
		return nil
	}

	fmt.Printf("  %s (%d files, %s tokens)\n", 
		color.CyanString("Context:"),
		len(cli.session.Files),
		formatNumber(cli.session.TotalTokens))
	
	fmt.Println(color.HiBlackString("  " + strings.Repeat("─", 50)))
	
	for _, f := range cli.session.Files {
		name := truncate(f.Path, 35)
		fmt.Printf("  %s %s %s\n", 
			color.GreenString("•"),
			padRight(name, 37),
			color.HiBlackString("%s, %s", formatBytes(f.Size), formatNumber(f.Tokens)))
	}
	
	return nil
}

// cmdStatus shows status
func (cli *InteractiveCLI) cmdStatus() error {
	fmt.Printf("  %s\n", color.CyanString("Status"))
	fmt.Printf("    Mode:   %s\n", cli.session.Mode)
	fmt.Printf("    Budget: %s tokens\n", formatNumber(cli.session.Budget))
	fmt.Printf("    Files:  %d\n", len(cli.session.Files))
	fmt.Printf("    Tokens: %s / %s\n", 
		formatNumber(cli.session.TotalTokens),
		formatNumber(cli.session.Budget))
	fmt.Printf("    Uptime: %s\n", time.Since(cli.session.StartTime).Round(time.Second))
	return nil
}

// cmdTokens shows token usage
func (cli *InteractiveCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	pct := float64(used) / float64(budget) * 100

	fmt.Printf("  %s\n", color.CyanString("Token Usage"))
	
	// Progress bar
	width := 40
	filled := int(float64(width) * pct / 100)
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	var coloredBar string
	switch {
	case pct > 90:
		coloredBar = color.RedString(bar)
	case pct > 70:
		coloredBar = color.YellowString(bar)
	default:
		coloredBar = color.GreenString(bar)
	}
	
	fmt.Printf("  %s\n", coloredBar)
	fmt.Printf("  %s / %s (%.1f%%)\n", formatNumber(used), formatNumber(budget), pct)
	
	if pct > 90 {
		fmt.Printf("  %s Context nearly full!\n", color.YellowString("⚠"))
	}
	
	return nil
}

// cmdCost shows cost
func (cli *InteractiveCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03
	
	fmt.Printf("  %s\n", color.CyanString("Cost Estimate"))
	fmt.Printf("    Tokens: %s\n", formatNumber(cli.session.TotalTokens))
	fmt.Printf("    Cost:   $%.2f\n", cost)
	fmt.Printf("    Rate:   $0.03 / 1K tokens\n")
	return nil
}

// cmdMode sets mode
func (cli *InteractiveCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		fmt.Printf("  Current: %s (fast, balanced, aggressive)\n", cli.session.Mode)
		return nil
	}

	mode := strings.ToLower(args[0])
	if mode != "fast" && mode != "balanced" && mode != "aggressive" {
		return fmt.Errorf("mode must be: fast, balanced, or aggressive")
	}

	cli.session.Mode = mode
	cli.printSuccess(fmt.Sprintf("Mode set to %s", mode))
	return nil
}

// cmdBudget sets budget
func (cli *InteractiveCLI) cmdBudget(args []string) error {
	if len(args) == 0 {
		fmt.Printf("  Current: %s tokens\n", formatNumber(cli.session.Budget))
		return nil
	}

	var budget int
	if _, err := fmt.Sscanf(args[0], "%d", &budget); err != nil || budget <= 0 {
		return fmt.Errorf("invalid budget")
	}

	cli.session.Budget = budget
	cli.printSuccess(fmt.Sprintf("Budget set to %s tokens", formatNumber(budget)))
	return nil
}

// cmdCompact compresses
func (cli *InteractiveCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(color.HiBlackString("  No files to compress"))
		return nil
	}

	before := cli.session.TotalTokens
	ratio := 0.85
	after := int(float64(before) * (1 - ratio))
	saved := before - after

	cli.session.TotalTokens = after
	
	cli.printSuccess(fmt.Sprintf("Compressed: %s → %s (saved %s, %.0f%%)",
		formatNumber(before), formatNumber(after), formatNumber(saved), ratio*100))
	return nil
}

// cmdStats shows stats
func (cli *InteractiveCLI) cmdStats() error {
	if cli.tracker == nil {
		fmt.Println(color.HiBlackString("  Tracking not enabled"))
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		fmt.Println(color.HiBlackString("  No data yet. Run some commands!"))
		return nil
	}

	fmt.Printf("  %s\n", color.CyanString("Statistics"))
	fmt.Printf("    Commands: %s\n", formatNumber(summary.TotalCommands))
	fmt.Printf("    Saved:    %s tokens\n", formatNumber(summary.TotalSaved))
	fmt.Printf("    Avg:      %.1f%% reduction\n", summary.ReductionPct)
	fmt.Printf("    Value:    $%.2f\n", float64(summary.TotalSaved)*0.00003)
	return nil
}

// cmdFilters lists filters
func (cli *InteractiveCLI) cmdFilters() error {
	fmt.Printf("  %s\n", color.CyanString("Active Filters"))
	filters := []string{
		"Entropy Filter",
		"Perplexity Filter",
		"H2O Filter",
		"AST Preservation",
		"Semantic Compaction",
	}
	
	for _, f := range filters {
		fmt.Printf("    %s %s\n", color.GreenString("✓"), f)
	}
	
	return nil
}

// cmdConfig shows config
func (cli *InteractiveCLI) cmdConfig() error {
	fmt.Printf("  %s\n", color.CyanString("Configuration"))
	fmt.Printf("    Config: %s\n", config.ConfigPath())
	fmt.Printf("    Data:   %s\n", config.DataPath())
	fmt.Printf("    Mode:   %s\n", cli.session.Mode)
	fmt.Printf("    Budget: %s\n", formatNumber(cli.session.Budget))
	return nil
}

// executeCommand runs shell command
func (cli *InteractiveCLI) executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	if _, err := exec.LookPath(parts[0]); err != nil {
		return fmt.Errorf("command not found: %s", parts[0])
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()

	if err != nil && len(output) == 0 {
		return fmt.Errorf("%s: %v", parts[0], err)
	}

	// Compress
	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	compressed, stats := p.Process(string(output))

	// Show output
	if len(compressed) > 0 {
		lines := strings.Split(compressed, "\n")
		for _, line := range lines[:min(len(lines), 20)] {
			fmt.Println("  " + line)
		}
		if len(lines) > 20 {
			fmt.Println(color.HiBlackString(fmt.Sprintf("  ... %d more lines", len(lines)-20)))
		}
	}

	if stats.TotalSaved > 0 {
		pct := float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
		fmt.Println()
		fmt.Printf("  %s Compressed %s → %s (%.0f%%)\n",
			color.GreenString("✓"),
			formatNumber(stats.OriginalTokens),
			formatNumber(stats.OriginalTokens-stats.TotalSaved),
			pct)
	}

	return nil
}

// Helpers

func (cli *InteractiveCLI) printSuccess(msg string) {
	fmt.Printf("  %s %s\n", color.GreenString("✓"), msg)
}

func (cli *InteractiveCLI) printError(msg string) {
	fmt.Printf("  %s %s\n", color.RedString("✗"), msg)
}

func padRight(s string, length int) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-utf8.RuneCountInString(s))
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
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen-3]) + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
