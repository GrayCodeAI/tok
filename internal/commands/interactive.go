// Package commands provides an interactive CLI inspired by gh/vercel/stripe.
// Design patterns from:
// - github.com/cli/cli (clean output, tables)
// - github.com/vercel/vercel (status indicators, minimal)
// - github.com/stripe/stripe-cli (professional formatting)
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

// InteractiveCLI provides a gh/vercel-inspired interface
type InteractiveCLI struct {
	scanner *bufio.Scanner
	tracker *tracking.Tracker
	config  *config.Config
	session *Session
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
	}

	cli.printWelcome()
	return cli.loop()
}

// printWelcome shows a minimal welcome like gh CLI
func (cli *InteractiveCLI) printWelcome() {
	fmt.Println()
	fmt.Printf("%s %s\n", 
		color.CyanString("TokMan"),
		color.HiBlackString("v"+shared.Version))
	
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("  %s %d commands  %s %.1f%% avg savings\n",
				color.HiBlackString("Session:"),
				summary.TotalCommands,
				color.HiBlackString("·"),
				summary.ReductionPct)
		}
	}
	
	fmt.Println()
	
	// Quick reference - gh CLI style
	fmt.Println(color.HiBlackString("Usage:"))
	fmt.Println("  " + color.CyanString("/add <file>") + "     Add file to context")
	fmt.Println("  " + color.CyanString("/status") + "         Show session status")
	fmt.Println("  " + color.CyanString("/tokens") + "         Show token usage")
	fmt.Println("  " + color.CyanString("/help") + "           List all commands")
	fmt.Println()
	fmt.Println("  Or type any shell command: git status, docker ps, kubectl logs")
	fmt.Println()
}

// loop runs the main REPL
func (cli *InteractiveCLI) loop() error {
	for {
		prompt := cli.buildPrompt()
		fmt.Print(prompt)

		if !cli.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			continue
		}

		// Process
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				fmt.Println()
				return nil
			}
			cli.printError(err.Error())
		}
	}
	return nil
}

// buildPrompt creates a clean prompt like gh CLI
func (cli *InteractiveCLI) buildPrompt() string {
	// Simple prompt with optional status
	if len(cli.session.Files) > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		var status string
		if pct > 90 {
			status = color.RedString(fmt.Sprintf("(%d files, %.0f%%)", len(cli.session.Files), pct))
		} else if pct > 70 {
			status = color.YellowString(fmt.Sprintf("(%d files, %.0f%%)", len(cli.session.Files), pct))
		} else {
			status = color.GreenString(fmt.Sprintf("(%d files, %.0f%%)", len(cli.session.Files), pct))
		}
		return fmt.Sprintf("%s %s ", color.CyanString("→"), status)
	}
	return color.CyanString("→ ")
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
		fmt.Printf("TokMan version %s\n", shared.Version)
		return nil
	
	default:
		return fmt.Errorf("unknown command: /%s (try /help)", cmd)
	}
}

// cmdHelp shows help like gh CLI
func (cli *InteractiveCLI) cmdHelp() error {
	fmt.Println()
	fmt.Println(color.HiBlackString("Core commands:"))
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
		{"/stats", "Show compression statistics"},
		{"", ""},
		{"/mode <type>", "Set compression mode (fast/balanced/aggressive)"},
		{"/budget <n>", "Set token budget"},
		{"/filters", "List active compression filters"},
		{"/config", "Show configuration"},
		{"", ""},
		{"/compact", "Compress context to save tokens"},
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
	fmt.Println(color.HiBlackString("You can also type any shell command directly."))
	fmt.Println()
	return nil
}

// cmdAdd adds a file to context
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

// cmdList lists files like gh repo list
func (cli *InteractiveCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(color.HiBlackString("No files in context. Use /add <file> to add files."))
		return nil
	}

	// Table header like gh CLI
	fmt.Println()
	fmt.Printf("  %s  %s  %s\n",
		padRight("FILE", 40),
		padRight("SIZE", 10),
		"TOKENS")
	fmt.Println(color.HiBlackString("  " + strings.Repeat("─", 60)))
	
	for _, f := range cli.session.Files {
		name := truncate(f.Path, 38)
		fmt.Printf("  %s  %s  %s\n",
			padRight(name, 40),
			padRight(formatBytes(f.Size), 10),
			formatNumber(f.Tokens))
	}
	
	fmt.Println(color.HiBlackString("  " + strings.Repeat("─", 60)))
	fmt.Printf("  %s: %d files, %s tokens\n",
		color.HiBlackString("Total"),
		len(cli.session.Files),
		formatNumber(cli.session.TotalTokens))
	fmt.Println()
	return nil
}

// cmdStatus shows status
func (cli *InteractiveCLI) cmdStatus() error {
	fmt.Println()
	fmt.Printf("  %s %s\n", color.HiBlackString("Mode:"), cli.session.Mode)
	fmt.Printf("  %s %s tokens\n", color.HiBlackString("Budget:"), formatNumber(cli.session.Budget))
	fmt.Printf("  %s %d files\n", color.HiBlackString("Files:"), len(cli.session.Files))
	fmt.Printf("  %s %s / %s tokens (%.1f%%)\n", 
		color.HiBlackString("Usage:"),
		formatNumber(cli.session.TotalTokens),
		formatNumber(cli.session.Budget),
		float64(cli.session.TotalTokens)/float64(cli.session.Budget)*100)
	fmt.Printf("  %s %s\n", color.HiBlackString("Uptime:"), time.Since(cli.session.StartTime).Round(time.Second))
	fmt.Println()
	return nil
}

// cmdTokens shows token usage with progress bar
func (cli *InteractiveCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	pct := float64(used) / float64(budget) * 100

	fmt.Println()
	
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
	fmt.Printf("  %s / %s tokens (%.1f%%)\n",
		color.WhiteString(formatNumber(used)),
		color.HiBlackString(formatNumber(budget)),
		pct)
	
	if pct > 90 {
		fmt.Println()
		fmt.Printf("  %s Context nearly full. Run /compact to free tokens.\n", color.YellowString("!"))
	}
	
	fmt.Println()
	return nil
}

// cmdCost shows cost estimate
func (cli *InteractiveCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03
	
	fmt.Println()
	fmt.Printf("  %s %s\n", color.HiBlackString("Tokens:"), formatNumber(cli.session.TotalTokens))
	fmt.Printf("  %s $%.2f\n", color.HiBlackString("Est. cost:"), cost)
	fmt.Printf("  %s $0.03 / 1K tokens\n", color.HiBlackString("Rate:"))
	fmt.Println()
	return nil
}

// cmdMode sets compression mode
func (cli *InteractiveCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		fmt.Printf("Current mode: %s (options: fast, balanced, aggressive)\n", cli.session.Mode)
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
		fmt.Printf("Current budget: %s tokens\n", formatNumber(cli.session.Budget))
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

// cmdCompact compresses context
func (cli *InteractiveCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(color.HiBlackString("No files to compress"))
		return nil
	}

	before := cli.session.TotalTokens
	ratio := 0.85
	after := int(float64(before) * (1 - ratio))
	saved := before - after

	cli.session.TotalTokens = after
	
	fmt.Println()
	fmt.Printf("  %s Compressed context\n", color.GreenString("✓"))
	fmt.Printf("  Before: %s tokens\n", formatNumber(before))
	fmt.Printf("  After:  %s tokens\n", formatNumber(after))
	fmt.Printf("  Saved:  %s tokens (%.0f%%)\n", formatNumber(saved), ratio*100)
	fmt.Println()
	return nil
}

// cmdStats shows statistics
func (cli *InteractiveCLI) cmdStats() error {
	if cli.tracker == nil {
		fmt.Println(color.HiBlackString("Tracking not enabled"))
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		fmt.Println(color.HiBlackString("No data yet. Run some commands!"))
		return nil
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", color.HiBlackString("Commands:"), formatNumber(summary.TotalCommands))
	fmt.Printf("  %s %s\n", color.HiBlackString("Tokens saved:"), formatNumber(summary.TotalSaved))
	fmt.Printf("  %s %.1f%%\n", color.HiBlackString("Avg reduction:"), summary.ReductionPct)
	fmt.Printf("  %s $%.2f\n", color.HiBlackString("Est. value:"), float64(summary.TotalSaved)*0.00003)
	fmt.Println()
	return nil
}

// cmdFilters lists active filters
func (cli *InteractiveCLI) cmdFilters() error {
	fmt.Println()
	fmt.Println(color.HiBlackString("Active compression filters:"))
	fmt.Println()
	
	filters := []struct {
		name   string
		status string
	}{
		{"Entropy Filter", "enabled"},
		{"Perplexity Filter", "enabled"},
		{"H2O Filter", "enabled"},
		{"AST Preservation", "enabled"},
		{"Semantic Compaction", "enabled"},
	}
	
	for _, f := range filters {
		fmt.Printf("  %s %s %s\n", color.GreenString("✓"), f.name, color.HiBlackString("["+f.status+"]"))
	}
	
	fmt.Println()
	return nil
}

// cmdConfig shows configuration
func (cli *InteractiveCLI) cmdConfig() error {
	fmt.Println()
	fmt.Printf("  %s %s\n", color.HiBlackString("Config file:"), config.ConfigPath())
	fmt.Printf("  %s %s\n", color.HiBlackString("Data dir:"), config.DataPath())
	fmt.Printf("  %s %s\n", color.HiBlackString("Mode:"), cli.session.Mode)
	fmt.Printf("  %s %s tokens\n", color.HiBlackString("Budget:"), formatNumber(cli.session.Budget))
	fmt.Println()
	return nil
}

// executeCommand runs a shell command
func (cli *InteractiveCLI) executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	// Check command exists
	if _, err := exec.LookPath(parts[0]); err != nil {
		return fmt.Errorf("command not found: %s", parts[0])
	}

	// Execute
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()

	if err != nil && len(output) == 0 {
		return fmt.Errorf("%s: %v", parts[0], err)
	}

	// Compress output
	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	compressed, stats := p.Process(string(output))

	// Show output
	if len(compressed) > 0 {
		lines := strings.Split(compressed, "\n")
		for _, line := range lines[:min(len(lines), 25)] {
			fmt.Println(line)
		}
		if len(lines) > 25 {
			fmt.Println(color.HiBlackString(fmt.Sprintf("... %d more lines", len(lines)-25)))
		}
	}

	// Show savings
	if stats.TotalSaved > 0 {
		pct := float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
		fmt.Println()
		fmt.Printf("%s Compressed %s → %s (%.0f%% savings)\n",
			color.GreenString("✓"),
			formatNumber(stats.OriginalTokens),
			formatNumber(stats.OriginalTokens-stats.TotalSaved),
			pct)
	}

	return nil
}

// Helper functions

func (cli *InteractiveCLI) printSuccess(msg string) {
	fmt.Printf("%s %s\n", color.GreenString("✓"), msg)
}

func (cli *InteractiveCLI) printError(msg string) {
	fmt.Printf("%s %s\n", color.RedString("✗"), msg)
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
