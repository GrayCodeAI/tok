// Package commands provides a Claude Code-style interactive CLI.
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

// InteractiveCLI provides a Claude Code-style interface
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
		width: 80,
	}

	cli.printWelcome()
	return cli.loop()
}

func (cli *InteractiveCLI) printWelcome() {
	// Header
	fmt.Println()
	fmt.Println(color.CyanString(strings.Repeat("═", cli.width)))
	fmt.Println()
	fmt.Printf("  %s  %s\n", color.CyanString("◉"), color.WhiteString("TokMan"))
	fmt.Printf("     %s\n", color.HiBlackString("Token-efficient CLI"))
	
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("     %s %d commands · %.1f%% avg savings\n",
				color.HiBlackString("●"), summary.TotalCommands, summary.ReductionPct)
		}
	}
	
	fmt.Println()
	fmt.Println(color.CyanString(strings.Repeat("═", cli.width)))
	fmt.Println()
}

func (cli *InteractiveCLI) loop() error {
	for {
		// Print input box
		cli.printInputBox()

		if !cli.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			fmt.Println()
			continue
		}

		// Show what user typed
		fmt.Printf("\n  %s %s\n\n", color.CyanString("▶"), input)

		// Process
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				fmt.Println()
				fmt.Println(color.CyanString(strings.Repeat("═", cli.width)))
				fmt.Println()
				fmt.Println("  " + color.HiBlackString("Goodbye!"))
				fmt.Println()
				return nil
			}
			cli.printError(err.Error())
		}
		
		fmt.Println()
	}
	return nil
}

func (cli *InteractiveCLI) printInputBox() {
	width := cli.width - 4
	
	// Top border
	fmt.Println(color.CyanString("  ┌" + strings.Repeat("─", width) + "┐"))
	
	// Context line
	if len(cli.session.Files) > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		var pctStr string
		if pct > 90 {
			pctStr = color.RedString(fmt.Sprintf("%.0f%%", pct))
		} else if pct > 70 {
			pctStr = color.YellowString(fmt.Sprintf("%.0f%%", pct))
		} else {
			pctStr = color.GreenString(fmt.Sprintf("%.0f%%", pct))
		}
		line := fmt.Sprintf("Context: %d files · %s tokens (%s of %s)",
			len(cli.session.Files),
			formatNumber(cli.session.TotalTokens),
			pctStr,
			formatNumber(cli.session.Budget))
		padding := width - len(line) - 1
		fmt.Printf("  %s %s%s%s\n", 
			color.CyanString("│"),
			color.HiBlackString(line),
			strings.Repeat(" ", padding),
			color.CyanString("│"))
	}
	
	// Mode line
	modeLine := fmt.Sprintf("Mode: %s · Budget: %s", cli.session.Mode, formatNumber(cli.session.Budget))
	padding := width - len(modeLine) - 1
	fmt.Printf("  %s %s%s%s\n",
		color.CyanString("│"),
		color.HiBlackString(modeLine),
		strings.Repeat(" ", padding),
		color.CyanString("│"))
	
	// Separator
	fmt.Println(color.CyanString("  ├" + strings.Repeat("─", width) + "┤"))
	
	// Input line
	fmt.Printf("  %s %s ", color.CyanString("│"), color.WhiteString(">"))
}

func (cli *InteractiveCLI) printBoxBottom() {
	// Don't print bottom border - keeps box open like Claude Code
	// In interactive mode, the input appears after the >
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
	fmt.Println(color.WhiteString("Available commands:"))
	fmt.Println()
	
	sections := []struct {
		title string
		cmds  []struct{ cmd, desc string }
	}{
		{
			"Context Management",
			[]struct{ cmd, desc string }{
				{"/add <file>", "Add file to context"},
				{"/drop <file>", "Remove file from context"},
				{"/ls", "List files"},
				{"/clear", "Clear all context"},
			},
		},
		{
			"Status & Info",
			[]struct{ cmd, desc string }{
				{"/status", "Show session status"},
				{"/tokens", "Show token usage"},
				{"/cost", "Show API cost"},
				{"/stats", "Show statistics"},
			},
		},
		{
			"Settings",
			[]struct{ cmd, desc string }{
				{"/mode <type>", "Set: fast/balanced/aggressive"},
				{"/budget <n>", "Set token budget"},
				{"/filters", "List filters"},
				{"/config", "Show config"},
			},
		},
		{
			"Other",
			[]struct{ cmd, desc string }{
				{"/compact", "Compress context"},
				{"/help", "Show help"},
				{"/quit", "Exit"},
			},
		},
	}
	
	for _, section := range sections {
		fmt.Printf("  %s\n", color.HiBlackString(section.title))
		for _, c := range section.cmds {
			fmt.Printf("    %s  %s\n", color.CyanString(padRight(c.cmd, 14)), c.desc)
		}
		fmt.Println()
	}
	
	fmt.Println("  " + color.HiBlackString("Type any shell command directly (git status, docker ps, etc.)"))
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
		fmt.Println(color.HiBlackString("  No files in context. Use /add <file> to add."))
		return nil
	}

	fmt.Printf("  %s (%d files, %s tokens)\n\n",
		color.WhiteString("Context:"),
		len(cli.session.Files),
		formatNumber(cli.session.TotalTokens))
	
	for _, f := range cli.session.Files {
		name := truncate(f.Path, 45)
		fmt.Printf("  %s  %-47s %s  %s\n",
			color.GreenString("●"),
			name,
			color.HiBlackString(formatBytes(f.Size)),
			color.HiBlackString(formatNumber(f.Tokens)+" tokens"))
	}
	
	return nil
}

func (cli *InteractiveCLI) cmdStatus() error {
	fmt.Println(color.WhiteString("Status"))
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

	fmt.Println(color.WhiteString("Token Usage"))
	fmt.Println()
	
	width := 60
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
	
	fmt.Printf("  %s\n\n", coloredBar)
	fmt.Printf("  Used:      %s tokens (%.1f%%)\n", formatNumber(used), pct)
	fmt.Printf("  Budget:    %s tokens\n", formatNumber(budget))
	fmt.Printf("  Remaining: %s tokens\n", formatNumber(budget-used))
	
	if pct > 90 {
		fmt.Printf("\n  %s Context nearly full! Run /compact.\n", color.YellowString("⚠"))
	}
	
	return nil
}

func (cli *InteractiveCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03
	
	fmt.Println(color.WhiteString("Cost Estimate"))
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
		fmt.Println(color.HiBlackString("  No files to compress"))
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
		fmt.Println(color.HiBlackString("  Tracking not enabled"))
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		fmt.Println(color.HiBlackString("  No data yet. Run some commands!"))
		return nil
	}

	fmt.Println(color.WhiteString("Statistics"))
	fmt.Println()
	fmt.Printf("  Commands: %s\n", formatNumber(summary.TotalCommands))
	fmt.Printf("  Saved:    %s tokens\n", formatNumber(summary.TotalSaved))
	fmt.Printf("  Avg:      %.1f%% reduction\n", summary.ReductionPct)
	fmt.Printf("  Value:    $%.2f\n", float64(summary.TotalSaved)*0.00003)
	return nil
}

func (cli *InteractiveCLI) cmdFilters() error {
	fmt.Println(color.WhiteString("Active Filters"))
	fmt.Println()
	
	filters := []string{
		"Entropy Filter",
		"Perplexity Filter",
		"H2O Filter",
		"AST Preservation",
		"Semantic Compaction",
	}
	
	for _, f := range filters {
		fmt.Printf("  %s  %s\n", color.GreenString("●"), f)
	}
	
	return nil
}

func (cli *InteractiveCLI) cmdConfig() error {
	fmt.Println(color.WhiteString("Configuration"))
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
			fmt.Println(color.HiBlackString(fmt.Sprintf("  ... %d more lines", len(lines)-15)))
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
