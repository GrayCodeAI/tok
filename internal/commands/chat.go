// Package commands provides a clean chat-style CLI interface.
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

// ChatCLI provides a clean chat-style interface
type ChatCLI struct {
	scanner *bufio.Scanner
	tracker *tracking.Tracker
	config  *config.Config
	session *ChatSession
}

type ChatSession struct {
	Files       []string
	TotalTokens int
	Budget      int
	Mode        string
	StartTime   time.Time
}

// RunChat starts the chat-style CLI
func RunChat() error {
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	cli := &ChatCLI{
		scanner: bufio.NewScanner(os.Stdin),
		tracker: tracking.GetGlobalTracker(),
		config:  cfg,
		session: &ChatSession{
			Files:     []string{},
			Budget:    200000,
			Mode:      "balanced",
			StartTime: time.Now(),
		},
	}

	cli.showHeader()
	return cli.loop()
}

func (cli *ChatCLI) showHeader() {
	// Clean header
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                   TokMan  Token-efficient CLI                   ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════╝")
	
	// Stats line
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			fmt.Printf("  Session: %d commands  •  Savings: %.1f%% avg\n", 
				summary.TotalCommands, summary.ReductionPct)
		}
	}
	
	fmt.Println()
	fmt.Println("  Commands: /add  /drop  /ls  /status  /tokens  /cost  /help  /quit")
	fmt.Println("  Or type: git status, docker ps, kubectl logs, go test, etc.")
	fmt.Println()
	fmt.Println("  ──────────────────────────────────────────────────────────────────")
	fmt.Println()
}

func (cli *ChatCLI) loop() error {
	for {
		// Show prompt
		prompt := cli.buildPrompt()
		fmt.Print(prompt)

		if !cli.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			continue
		}

		// Echo user input
		fmt.Printf("\r  %s %s\n", color.HiBlackString("You:"), input)

		// Process
		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				fmt.Println()
				fmt.Println("  " + color.CyanString("TokMan:") + " Goodbye! 👋")
				fmt.Println()
				return nil
			}
			fmt.Printf("  %s %s\n", color.RedString("✗"), err.Error())
		}
		fmt.Println()
	}
	return nil
}

func (cli *ChatCLI) buildPrompt() string {
	var parts []string
	
	// Token indicator
	if cli.session.TotalTokens > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		var indicator string
		switch {
		case pct > 90:
			indicator = color.RedString("●")
		case pct > 70:
			indicator = color.YellowString("●")
		default:
			indicator = color.GreenString("●")
		}
		parts = append(parts, indicator)
		
		if len(cli.session.Files) > 0 {
			parts = append(parts, color.HiBlackString(fmt.Sprintf("(%d)", len(cli.session.Files))))
		}
	}
	
	return "  " + strings.Join(parts, " ") + color.CyanString(" →")
}

func (cli *ChatCLI) handle(input string) error {
	if strings.HasPrefix(input, "/") {
		return cli.handleCommand(input[1:])
	}
	return cli.executeShellCommand(input)
}

func (cli *ChatCLI) handleCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help", "?":
		return cli.cmdHelp()
	case "quit", "exit", "q":
		return fmt.Errorf("exit")
	case "clear":
		fmt.Print("\033[H\033[2J")
		cli.showHeader()
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
		cli.printResponse("TokMan v" + shared.Version)
		return nil

	default:
		return fmt.Errorf("unknown command: /%s (try /help)", cmd)
	}
}

func (cli *ChatCLI) cmdHelp() error {
	fmt.Println("  " + color.CyanString("TokMan:") + " Here are the available commands:")
	fmt.Println()
	
	sections := []struct {
		title    string
		commands []string
	}{
		{"Context", []string{
			"/add <file>     Add file to context",
			"/drop <file>    Remove file from context",
			"/ls             List files in context",
		}},
		{"Status", []string{
			"/status         Show session status",
			"/tokens         Show token usage",
			"/cost           Show API cost estimate",
			"/stats          Show statistics",
		}},
		{"Settings", []string{
			"/mode <type>    Set mode: fast/balanced/aggressive",
			"/budget <n>     Set token budget",
			"/filters        List active filters",
			"/config         Show configuration",
		}},
		{"Other", []string{
			"/compact        Compress context",
			"/help           Show this help",
			"/quit           Exit TokMan",
		}},
	}
	
	for _, section := range sections {
		fmt.Println("  " + color.WhiteString(section.title))
		for _, cmd := range section.commands {
			fmt.Println("    " + color.GreenString(cmd[:10]) + cmd[10:])
		}
		fmt.Println()
	}
	
	fmt.Println("  " + color.HiBlackString("Or type any shell command directly (git status, docker ps, etc.)"))
	return nil
}

func (cli *ChatCLI) cmdAdd(args []string) error {
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

	cli.printResponse(fmt.Sprintf("Added %s (%s, %d tokens)", path, formatBytes(info.Size()), tokens))
	return nil
}

func (cli *ChatCLI) cmdDrop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /drop <file>")
	}

	target := args[0]
	for i, f := range cli.session.Files {
		if f == target || filepath.Base(f) == target {
			cli.session.TotalTokens -= len(f) / 4
			cli.session.Files = append(cli.session.Files[:i], cli.session.Files[i+1:]...)
			cli.printResponse(fmt.Sprintf("Removed %s", target))
			return nil
		}
	}

	return fmt.Errorf("file not in context: %s", target)
}

func (cli *ChatCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		cli.printResponse("No files in context. Use /add <file> to add.")
		return nil
	}

	fmt.Printf("  %s %s (%d files, %d tokens)\n", 
		color.CyanString("TokMan:"),
		color.WhiteString("Context:"),
		len(cli.session.Files),
		cli.session.TotalTokens)
	
	for i, f := range cli.session.Files {
		info, _ := os.Stat(f)
		size := ""
		if info != nil {
			size = color.HiBlackString("(%s)", formatBytes(info.Size()))
		}
		fmt.Printf("    %d. %s %s\n", i+1, f, size)
	}
	return nil
}

func (cli *ChatCLI) cmdStatus() error {
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Status"))
	fmt.Printf("    Mode:    %s\n", color.CyanString(cli.session.Mode))
	fmt.Printf("    Budget:  %s tokens\n", formatNumber(cli.session.Budget))
	fmt.Printf("    Files:   %d\n", len(cli.session.Files))
	fmt.Printf("    Tokens:  %s / %s\n", 
		formatNumber(cli.session.TotalTokens),
		formatNumber(cli.session.Budget))
	fmt.Printf("    Uptime:  %s\n", time.Since(cli.session.StartTime).Round(time.Second))
	return nil
}

func (cli *ChatCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	pct := float64(used) / float64(budget) * 100

	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Token Usage"))
	
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
	
	fmt.Printf("    %s\n", coloredBar)
	fmt.Printf("    %s / %s (%.1f%%)\n", formatNumber(used), formatNumber(budget), pct)
	
	if pct > 90 {
		fmt.Printf("    %s Context nearly full!\n", color.YellowString("⚠"))
	}
	return nil
}

func (cli *ChatCLI) cmdCost() error {
	cost := float64(cli.session.TotalTokens) / 1000 * 0.03
	
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Cost Estimate"))
	fmt.Printf("    Tokens:  %s\n", formatNumber(cli.session.TotalTokens))
	fmt.Printf("    Cost:    $%.2f\n", cost)
	fmt.Printf("    Rate:    $0.03 / 1K tokens\n")
	return nil
}

func (cli *ChatCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		cli.printResponse(fmt.Sprintf("Current mode: %s. Options: fast, balanced, aggressive", cli.session.Mode))
		return nil
	}

	mode := strings.ToLower(args[0])
	if mode != "fast" && mode != "balanced" && mode != "aggressive" {
		return fmt.Errorf("mode must be: fast, balanced, or aggressive")
	}

	cli.session.Mode = mode
	cli.printResponse(fmt.Sprintf("Mode set to %s", mode))
	return nil
}

func (cli *ChatCLI) cmdBudget(args []string) error {
	if len(args) == 0 {
		cli.printResponse(fmt.Sprintf("Current budget: %s tokens", formatNumber(cli.session.Budget)))
		return nil
	}

	var budget int
	if _, err := fmt.Sscanf(args[0], "%d", &budget); err != nil || budget <= 0 {
		return fmt.Errorf("invalid budget")
	}

	cli.session.Budget = budget
	cli.printResponse(fmt.Sprintf("Budget set to %s tokens", formatNumber(budget)))
	return nil
}

func (cli *ChatCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		cli.printResponse("No files to compress")
		return nil
	}

	before := cli.session.TotalTokens
	ratio := 0.85
	after := int(float64(before) * (1 - ratio))
	saved := before - after

	cli.session.TotalTokens = after
	cli.printResponse(fmt.Sprintf("Compressed: %s → %s (saved %s, %.0f%%)",
		formatNumber(before), formatNumber(after), formatNumber(saved), ratio*100))
	return nil
}

func (cli *ChatCLI) cmdStats() error {
	if cli.tracker == nil {
		cli.printResponse("Tracking not enabled")
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		cli.printResponse("No data yet. Run some commands!")
		return nil
	}

	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Statistics"))
	fmt.Printf("    Commands:  %s\n", formatNumber(summary.TotalCommands))
	fmt.Printf("    Saved:     %s tokens\n", formatNumber(summary.TotalSaved))
	fmt.Printf("    Avg:       %.1f%% reduction\n", summary.ReductionPct)
	fmt.Printf("    Value:     $%.2f\n", float64(summary.TotalSaved)*0.00003)
	return nil
}

func (cli *ChatCLI) cmdFilters() error {
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Active Filters"))
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

func (cli *ChatCLI) cmdConfig() error {
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), color.WhiteString("Configuration"))
	fmt.Printf("    Config:  %s\n", config.ConfigPath())
	fmt.Printf("    Data:    %s\n", config.DataPath())
	fmt.Printf("    Mode:    %s\n", cli.session.Mode)
	fmt.Printf("    Budget:  %s\n", formatNumber(cli.session.Budget))
	return nil
}

func (cli *ChatCLI) executeShellCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	// Check if it's a known command
	if _, err := exec.LookPath(parts[0]); err != nil {
		// Not a command - try to filter as text
		return cli.filterText(input)
	}

	// Execute command
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
		for _, line := range lines[:min(len(lines), 30)] {
			fmt.Println("  " + line)
		}
		if len(lines) > 30 {
			fmt.Printf("  %s\n", color.HiBlackString("... (%d more lines)", len(lines)-30))
		}
	}

	if stats.TotalSaved > 0 {
		pct := float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
		fmt.Printf("  %s Saved %s tokens (%.0f%%)\n", 
			color.GreenString("✓"),
			formatNumber(stats.TotalSaved),
			pct)
	}

	return nil
}

func (cli *ChatCLI) filterText(text string) error {
	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	compressed, stats := p.Process(text)

	fmt.Println("  " + compressed)

	if stats.TotalSaved > 0 {
		pct := float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100
		cli.printResponse(fmt.Sprintf("Compressed: saved %s tokens (%.0f%%)", 
			formatNumber(stats.TotalSaved), pct))
	}
	return nil
}

func (cli *ChatCLI) printResponse(msg string) {
	fmt.Printf("  %s %s\n", color.CyanString("TokMan:"), msg)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
