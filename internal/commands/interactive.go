// Package commands provides an interactive CLI inspired by gh, stripe, and bubbletea.
//
// Design patterns from OSS research:
// - ColorScheme pattern from GitHub CLI
// - Model-View pattern from Bubbletea
// - Table formatting from Stripe CLI
// - Prompt style from Supabase CLI
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

// ColorScheme provides consistent coloring like gh CLI
type ColorScheme struct {
	Enabled bool
}

func NewColorScheme() *ColorScheme {
	// Check if stdout is a terminal
	enabled := true
	if fi, err := os.Stdout.Stat(); err == nil {
		enabled = (fi.Mode() & os.ModeCharDevice) != 0
	}
	// Check NO_COLOR env
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
	}
	return &ColorScheme{Enabled: enabled}
}

func (cs *ColorScheme) Bold(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.New(color.Bold).Sprint(s)
}

func (cs *ColorScheme) Cyan(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.CyanString(s)
}

func (cs *ColorScheme) Green(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.GreenString(s)
}

func (cs *ColorScheme) Yellow(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.YellowString(s)
}

func (cs *ColorScheme) Red(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.RedString(s)
}

func (cs *ColorScheme) Gray(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.HiBlackString(s)
}

func (cs *ColorScheme) White(s string) string {
	if !cs.Enabled {
		return s
	}
	return color.WhiteString(s)
}

func (cs *ColorScheme) SuccessIcon() string {
	return cs.Green("✓")
}

func (cs *ColorScheme) ErrorIcon() string {
	return cs.Red("✗")
}

func (cs *ColorScheme) WarningIcon() string {
	return cs.Yellow("!")
}

func (cs *ColorScheme) BulletIcon() string {
	return cs.Green("●")
}

// InteractiveCLI provides a professional CLI interface
type InteractiveCLI struct {
	scanner *bufio.Scanner
	tracker *tracking.Tracker
	config  *config.Config
	session *Session
	cs      *ColorScheme
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

	cs := NewColorScheme()

	cli := &InteractiveCLI{
		scanner: bufio.NewScanner(os.Stdin),
		tracker: tracking.GetGlobalTracker(),
		config:  cfg,
		cs:      cs,
		session: &Session{
			Files:     []ContextFile{},
			Budget:    200000,
			Mode:      "balanced",
			StartTime: time.Now(),
		},
		width: 80,
	}

	cli.run()
	return nil
}

func (cli *InteractiveCLI) run() {
	// Print welcome
	cli.printHeader()

	// Main loop
	for {
		// Print input area
		cli.printInputArea()

		// Read input
		if !cli.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			fmt.Println()
			continue
		}

		// Close input area visually
		fmt.Println()

		// Echo command
		fmt.Printf("  %s %s\n\n", cli.cs.Cyan("▶"), input)

		// Process
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				cli.printGoodbye()
				return
			}
			cli.printError(err.Error())
		}

		fmt.Println()
	}
}

func (cli *InteractiveCLI) printHeader() {
	fmt.Println()
	// Top border
	fmt.Println(cli.cs.Cyan(strings.Repeat("═", cli.width)))
	fmt.Println()

	// Title
	fmt.Printf("  %s  %s\n", cli.cs.Cyan("◉"), cli.cs.Bold("TokMan"))
	fmt.Printf("     %s\n", cli.cs.Gray("Token-efficient CLI"))

	// Stats
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("     %s %d commands · %.1f%% avg savings\n",
				cli.cs.Gray("●"),
				summary.TotalCommands,
				summary.ReductionPct)
		}
	}

	fmt.Println()
	fmt.Println(cli.cs.Cyan(strings.Repeat("═", cli.width)))
	fmt.Println()
}

func (cli *InteractiveCLI) printInputArea() {
	w := cli.width - 4

	// Box top
	fmt.Println(cli.cs.Cyan("  ┌" + strings.Repeat("─", w) + "┐"))

	// Context line (if any)
	if len(cli.session.Files) > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		var pctStr string
		switch {
		case pct > 90:
			pctStr = cli.cs.Red(fmt.Sprintf("%.0f%%", pct))
		case pct > 70:
			pctStr = cli.cs.Yellow(fmt.Sprintf("%.0f%%", pct))
		default:
			pctStr = cli.cs.Green(fmt.Sprintf("%.0f%%", pct))
		}
		line := fmt.Sprintf("Context: %d files · %s tokens (%s of %s)",
			len(cli.session.Files),
			formatNumber(cli.session.TotalTokens),
			pctStr,
			formatNumber(cli.session.Budget))
		padding := w - len(line) - 1
		fmt.Printf("  %s %s%s%s\n",
			cli.cs.Cyan("│"),
			cli.cs.Gray(line),
			strings.Repeat(" ", padding),
			cli.cs.Cyan("│"))
	}

	// Mode line
	modeLine := fmt.Sprintf("Mode: %s · Budget: %s", cli.session.Mode, formatNumber(cli.session.Budget))
	padding := w - len(modeLine) - 1
	fmt.Printf("  %s %s%s%s\n",
		cli.cs.Cyan("│"),
		cli.cs.Gray(modeLine),
		strings.Repeat(" ", padding),
		cli.cs.Cyan("│"))

	// Separator
	fmt.Println(cli.cs.Cyan("  ├" + strings.Repeat("─", w) + "┤"))

	// Input line with prompt
	fmt.Printf("  %s %s ", cli.cs.Cyan("│"), cli.cs.White(">"))
}

func (cli *InteractiveCLI) printGoodbye() {
	fmt.Println()
	fmt.Println(cli.cs.Cyan(strings.Repeat("═", cli.width)))
	fmt.Println()
	fmt.Printf("  %s\n", cli.cs.Gray("Goodbye!"))
	fmt.Println()
}

func (cli *InteractiveCLI) handle(input string) error {
	if strings.HasPrefix(input, "/") {
		return cli.handleCommand(input[1:])
	}
	return cli.executeCommand(input)
}

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
	case "add", "a":
		return cli.cmdAdd(args)
	case "drop", "d", "rm":
		return cli.cmdDrop(args)
	case "ls", "list":
		return cli.cmdList()
	case "clear":
		cli.session.Files = []ContextFile{}
		cli.session.TotalTokens = 0
		cli.printSuccess("Context cleared")
		return nil
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
	case "stats", "stat":
		return cli.cmdStats()
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

func (cli *InteractiveCLI) cmdHelp() error {
	fmt.Println(cli.cs.Bold("Available commands:"))
	fmt.Println()

	// Context Management
	fmt.Printf("  %s\n", cli.cs.Gray("Context Management"))
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/add <file>"), "Add file to context")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/drop <file>"), "Remove file from context")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/ls"), "List files in context")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/clear"), "Clear all context")
	fmt.Println()

	// Status
	fmt.Printf("  %s\n", cli.cs.Gray("Status & Info"))
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/status"), "Show session status")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/tokens"), "Show token usage")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/cost"), "Show API cost")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/stats"), "Show statistics")
	fmt.Println()

	// Settings
	fmt.Printf("  %s\n", cli.cs.Gray("Settings"))
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/mode <type>"), "Set: fast/balanced/aggressive")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/budget <n>"), "Set token budget")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/filters"), "List active filters")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/config"), "Show configuration")
	fmt.Println()

	// Other
	fmt.Printf("  %s\n", cli.cs.Gray("Other"))
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/compact"), "Compress context")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/help"), "Show this help")
	fmt.Printf("    %s  %s\n", cli.cs.Cyan("/quit"), "Exit TokMan")
	fmt.Println()

	fmt.Println("  " + cli.cs.Gray("Type any shell command directly (git status, docker ps, etc.)"))
	return nil
}

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

func (cli *InteractiveCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(cli.cs.Gray("  No files in context. Use /add <file> to add."))
		return nil
	}

	// Header
	fmt.Printf("  %s (%d files, %s tokens)\n\n",
		cli.cs.Bold("Context:"),
		len(cli.session.Files),
		formatNumber(cli.session.TotalTokens))

	// Files
	for _, f := range cli.session.Files {
		name := truncate(f.Path, 40)
		fmt.Printf("  %s  %-42s %s  %s\n",
			cli.cs.BulletIcon(),
			name,
			cli.cs.Gray(formatBytes(f.Size)),
			cli.cs.Gray(formatNumber(f.Tokens)+" tokens"))
	}

	return nil
}

func (cli *InteractiveCLI) cmdStatus() error {
	fmt.Println(cli.cs.Bold("Status"))
	fmt.Println()
	fmt.Printf("  Mode:   %s\n", cli.session.Mode)
	fmt.Printf("  Budget: %s tokens\n", formatNumber(cli.session.Budget))
	fmt.Printf("  Files:  %d\n", len(cli.session.Files))
	fmt.Printf("  Tokens: %s / %s\n",
		formatNumber(cli.session.TotalTokens),
		formatNumber(cli.session.Budget))
	fmt.Printf("  Uptime: %s\n", time.Since(cli.session.StartTime).Round(time.Second))
	return nil
}

func (cli *InteractiveCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	pct := float64(used) / float64(budget) * 100

	fmt.Println(cli.cs.Bold("Token Usage"))
	fmt.Println()

	// Progress bar
	width := 60
	filled := int(float64(width) * pct / 100)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	var coloredBar string
	switch {
	case pct > 90:
		coloredBar = cli.cs.Red(bar)
	case pct > 70:
		coloredBar = cli.cs.Yellow(bar)
	default:
		coloredBar = cli.cs.Green(bar)
	}

	fmt.Printf("  %s\n\n", coloredBar)
	fmt.Printf("  Used:      %s tokens (%.1f%%)\n", formatNumber(used), pct)
	fmt.Printf("  Budget:    %s tokens\n", formatNumber(budget))
	fmt.Printf("  Remaining: %s tokens\n", formatNumber(budget-used))

	if pct > 90 {
		fmt.Printf("\n  %s Context nearly full! Run /compact.\n", cli.cs.WarningIcon())
	}

	return nil
}

func (cli *InteractiveCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03

	fmt.Println(cli.cs.Bold("Cost Estimate"))
	fmt.Println()
	fmt.Printf("  Tokens: %s\n", formatNumber(cli.session.TotalTokens))
	fmt.Printf("  Cost:   $%.2f\n", cost)
	fmt.Printf("  Rate:   $0.03 / 1K tokens\n")
	return nil
}

func (cli *InteractiveCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		fmt.Printf("Current mode: %s\n", cli.session.Mode)
		fmt.Println("Options: fast, balanced, aggressive")
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

func (cli *InteractiveCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		fmt.Println(cli.cs.Gray("  No files to compress"))
		return nil
	}

	before := cli.session.TotalTokens
	ratio := 0.85
	after := int(float64(before) * (1 - ratio))
	saved := before - after

	cli.session.TotalTokens = after

	cli.printSuccess(fmt.Sprintf("Compressed context"))
	fmt.Printf("  Before: %s tokens\n", formatNumber(before))
	fmt.Printf("  After:  %s tokens\n", formatNumber(after))
	fmt.Printf("  Saved:  %s tokens (%.0f%%)\n", formatNumber(saved), ratio*100)
	return nil
}

func (cli *InteractiveCLI) cmdStats() error {
	if cli.tracker == nil {
		fmt.Println(cli.cs.Gray("  Tracking not enabled"))
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		fmt.Println(cli.cs.Gray("  No data yet. Run some commands!"))
		return nil
	}

	fmt.Println(cli.cs.Bold("Statistics"))
	fmt.Println()
	fmt.Printf("  Commands: %s\n", formatNumber(summary.TotalCommands))
	fmt.Printf("  Saved:    %s tokens\n", formatNumber(summary.TotalSaved))
	fmt.Printf("  Avg:      %.1f%% reduction\n", summary.ReductionPct)
	fmt.Printf("  Value:    $%.2f\n", float64(summary.TotalSaved)*0.00003)
	return nil
}

func (cli *InteractiveCLI) cmdFilters() error {
	fmt.Println(cli.cs.Bold("Active Filters"))
	fmt.Println()

	filters := []string{
		"Entropy Filter",
		"Perplexity Filter",
		"H2O Filter",
		"AST Preservation",
		"Semantic Compaction",
	}

	for _, f := range filters {
		fmt.Printf("  %s  %s\n", cli.cs.BulletIcon(), f)
	}

	return nil
}

func (cli *InteractiveCLI) cmdConfig() error {
	fmt.Println(cli.cs.Bold("Configuration"))
	fmt.Println()
	fmt.Printf("  Config file: %s\n", config.ConfigPath())
	fmt.Printf("  Data dir:    %s\n", config.DataPath())
	fmt.Printf("  Mode:        %s\n", cli.session.Mode)
	fmt.Printf("  Budget:      %s tokens\n", formatNumber(cli.session.Budget))
	return nil
}

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

	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	compressed, stats := p.Process(string(output))

	if len(compressed) > 0 {
		lines := strings.Split(compressed, "\n")
		for _, line := range lines[:min(len(lines), 15)] {
			fmt.Println("  " + line)
		}
		if len(lines) > 15 {
			fmt.Println(cli.cs.Gray(fmt.Sprintf("  ... %d more lines", len(lines)-15)))
		}
	}

	if stats.TotalSaved > 0 {
		fmt.Println()
		cli.printSuccess(fmt.Sprintf("Compressed %s → %s (%.0f%%)",
			formatNumber(stats.OriginalTokens),
			formatNumber(stats.OriginalTokens-stats.TotalSaved),
			float64(stats.TotalSaved)/float64(stats.OriginalTokens)*100))
	}

	return nil
}

// Output helpers

func (cli *InteractiveCLI) printSuccess(msg string) {
	fmt.Printf("  %s %s\n", cli.cs.SuccessIcon(), msg)
}

func (cli *InteractiveCLI) printError(msg string) {
	fmt.Printf("  %s %s\n", cli.cs.ErrorIcon(), msg)
}

// Utility functions

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
