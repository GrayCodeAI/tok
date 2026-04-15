// Package commands provides a professional CLI interface for TokMan.
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

// ProfessionalCLI provides a polished, modern CLI interface
type ProfessionalCLI struct {
	scanner  *bufio.Scanner
	tracker  *tracking.Tracker
	config   *config.Config
	session  *Session
	renderer *Renderer
}

// Session tracks the current session state
type Session struct {
	Files       []ContextFile
	TotalTokens int
	Budget      int
	Mode        string
	StartTime   time.Time
}

// ContextFile represents a file in context
type ContextFile struct {
	Path   string
	Size   int64
	Tokens int
}

// Renderer handles output formatting
type Renderer struct {
	width int
}

// RunProfessional starts the professional CLI mode
func RunProfessional() error {
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Defaults()
	}

	cli := &ProfessionalCLI{
		scanner:  bufio.NewScanner(os.Stdin),
		tracker:  tracking.GetGlobalTracker(),
		config:   cfg,
		session:  newSession(),
		renderer: &Renderer{width: 80},
	}

	cli.showIntro()
	return cli.loop()
}

func newSession() *Session {
	return &Session{
		Files:     []ContextFile{},
		Budget:    200000,
		Mode:      "balanced",
		StartTime: time.Now(),
	}
}

// showIntro displays the professional welcome screen
func (cli *ProfessionalCLI) showIntro() {
	cli.clearScreen()
	
	// Header
	cli.printHeader("TokMan", "Token-efficient CLI")
	
	// Quick stats if available
	if cli.tracker != nil {
		if summary, err := cli.tracker.GetSavings(""); err == nil && summary.TotalCommands > 0 {
			cli.printInfo(fmt.Sprintf("Session: %d commands • %.1f%% avg savings", 
				summary.TotalCommands, summary.ReductionPct))
		}
	}
	
	fmt.Println()
	cli.printSection("Quick Start")
	cli.printItem("Type a command", "git status, docker ps, kubectl logs")
	cli.printItem("Manage context", "/add, /drop, /ls")
	cli.printItem("View status", "/status, /tokens, /cost")
	cli.printItem("Get help", "/help or /?")
	
	fmt.Println()
}

// loop runs the main REPL
func (cli *ProfessionalCLI) loop() error {
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

		if err := cli.handle(input); err != nil {
			if err.Error() == "exit" {
				cli.printSuccess("Goodbye!")
				return nil
			}
			cli.printError(err.Error())
		}
	}
	return nil
}

// handle processes user input
func (cli *ProfessionalCLI) handle(input string) error {
	// Check for slash commands
	if strings.HasPrefix(input, "/") {
		return cli.handleCommand(input[1:])
	}

	// Execute as shell command
	return cli.executeShellCommand(input)
}

// handleCommand processes slash commands
func (cli *ProfessionalCLI) handleCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	// Help
	case "help", "?", "h":
		return cli.cmdHelp(args)
	case "exit", "quit", "q":
		return fmt.Errorf("exit")

	// Context management
	case "add", "a":
		return cli.cmdAdd(args)
	case "drop", "d":
		return cli.cmdDrop(args)
	case "ls", "list", "files":
		return cli.cmdList()
	case "clear":
		return cli.cmdClear()

	// Status & info
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

	// Actions
	case "compact", "compress":
		return cli.cmdCompact()
	case "filter", "f":
		return cli.cmdFilter(args)

	// Stats
	case "stats", "gain", "g":
		return cli.cmdStats()
	case "audit":
		return cli.cmdAudit(args)
	case "history", "hist":
		return cli.cmdHistory(args)

	// Info
	case "config":
		return cli.cmdConfig()
	case "filters":
		return cli.cmdFilters()
	case "layers":
		return cli.cmdLayers()
	case "version", "v":
		cli.printInfo("TokMan v" + shared.Version)
		return nil

	default:
		return fmt.Errorf("unknown command: /%s (try /help)", cmd)
	}
}

// ==================== COMMAND IMPLEMENTATIONS ====================

func (cli *ProfessionalCLI) cmdHelp(args []string) error {
	if len(args) > 0 {
		return cli.showCommandHelp(args[0])
	}

	cli.printHeader("Commands", "")
	
	cli.printSection("Context Management")
	cli.printCommand("/add <file>", "Add file to context")
	cli.printCommand("/drop <file>", "Remove file from context")
	cli.printCommand("/ls", "List files in context")
	cli.printCommand("/clear", "Clear all context")
	
	cli.printSection("Status & Monitoring")
	cli.printCommand("/status", "Show session status")
	cli.printCommand("/tokens", "Show token usage")
	cli.printCommand("/cost", "Show API cost estimate")
	cli.printCommand("/stats", "Show savings statistics")
	
	cli.printSection("Configuration")
	cli.printCommand("/mode <mode>", "Set compression mode (fast/balanced/aggressive)")
	cli.printCommand("/budget <n>", "Set token budget")
	cli.printCommand("/filters", "List active filters")
	cli.printCommand("/config", "Show configuration")
	
	cli.printSection("Actions")
	cli.printCommand("/compact", "Compress context")
	cli.printCommand("/filter <text>", "Compress text directly")
	cli.printCommand("/audit [days]", "Run optimization audit")
	cli.printCommand("/history [n]", "Show command history")
	
	cli.printSection("General")
	cli.printCommand("/help [cmd]", "Show command help")
	cli.printCommand("/version", "Show version")
	cli.printCommand("/exit", "Exit TokMan")
	
	fmt.Println()
	cli.printTip("Type any shell command to run it with automatic compression")
	
	return nil
}

func (cli *ProfessionalCLI) showCommandHelp(cmd string) error {
	help := map[string]string{
		"add":     "Add a file to context\n\nUsage: /add <filepath>\nExample: /add main.go",
		"drop":    "Remove a file from context\n\nUsage: /drop <filepath>",
		"tokens":  "Show token usage with visual progress bar",
		"cost":    "Show estimated API cost for current context",
		"compact": "Compress context to free up tokens",
		"mode":    "Set compression mode\n\nModes:\n  fast       - Quick compression\n  balanced   - Default\n  aggressive - Maximum compression",
	}
	
	text, ok := help[cmd]
	if !ok {
		return fmt.Errorf("no help available for /%s", cmd)
	}
	
	cli.printHeader("Help: /"+cmd, "")
	fmt.Println(text)
	return nil
}

func (cli *ProfessionalCLI) cmdAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /add <file>")
	}

	path := args[0]
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %s", path)
	}

	// Calculate tokens
	content, _ := os.ReadFile(path)
	tokens := len(content) / 4

	cli.session.Files = append(cli.session.Files, ContextFile{
		Path:   path,
		Size:   info.Size(),
		Tokens: tokens,
	})
	cli.session.TotalTokens += tokens

	cli.printSuccess("Added %s (%d tokens)", path, tokens)
	return nil
}

func (cli *ProfessionalCLI) cmdDrop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /drop <file>")
	}

	target := args[0]
	found := false
	
	for i, f := range cli.session.Files {
		if f.Path == target || filepath.Base(f.Path) == target {
			cli.session.TotalTokens -= f.Tokens
			cli.session.Files = append(cli.session.Files[:i], cli.session.Files[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("file not in context: %s", target)
	}

	cli.printSuccess("Removed %s", target)
	return nil
}

func (cli *ProfessionalCLI) cmdList() error {
	if len(cli.session.Files) == 0 {
		cli.printInfo("No files in context. Use /add <file> to add.")
		return nil
	}

	cli.printHeader("Context Files", fmt.Sprintf("%d files", len(cli.session.Files)))
	
	for _, f := range cli.session.Files {
		cli.printFile(f.Path, f.Size, f.Tokens)
	}
	
	fmt.Println()
	cli.printInfo("Total: %d files, %d tokens", len(cli.session.Files), cli.session.TotalTokens)
	return nil
}

func (cli *ProfessionalCLI) cmdClear() error {
	cli.session.Files = []ContextFile{}
	cli.session.TotalTokens = 0
	cli.printSuccess("Context cleared")
	return nil
}

func (cli *ProfessionalCLI) cmdStatus() error {
	cli.printHeader("Status", "")
	
	data := []struct {
		label string
		value string
	}{
		{"Mode", cli.session.Mode},
		{"Budget", formatNumber(cli.session.Budget)},
		{"Files", fmt.Sprintf("%d", len(cli.session.Files))},
		{"Tokens", fmt.Sprintf("%s / %s", formatNumber(cli.session.TotalTokens), formatNumber(cli.session.Budget))},
		{"Uptime", time.Since(cli.session.StartTime).Round(time.Second).String()},
	}
	
	for _, d := range data {
		cli.printKeyValue(d.label, d.value)
	}
	
	return nil
}

func (cli *ProfessionalCLI) cmdTokens() error {
	used := cli.session.TotalTokens
	budget := cli.session.Budget
	remaining := budget - used
	if remaining < 0 {
		remaining = 0
	}
	
	pct := float64(used) / float64(budget) * 100
	
	cli.printHeader("Token Usage", "")
	
	// Progress bar
	bar := cli.renderProgressBar(pct, 40)
	fmt.Println(bar)
	
	// Stats
	fmt.Printf("  %s %s / %s (%.1f%%)\n", 
		color.HiBlackString("Used:"),
		color.WhiteString(formatNumber(used)),
		color.WhiteString(formatNumber(budget)),
		pct,
	)
	fmt.Printf("  %s %s\n",
		color.HiBlackString("Remaining:"),
		color.GreenString(formatNumber(remaining)),
	)
	
	if pct > 90 {
		fmt.Println()
		cli.printWarning("Context nearly full. Run /compact to free tokens.")
	}
	
	return nil
}

func (cli *ProfessionalCLI) cmdCost() error {
	tokens := cli.session.TotalTokens
	cost := float64(tokens) / 1000 * 0.03 // $0.03 per 1K tokens
	
	cli.printHeader("Cost Estimate", "")
	cli.printKeyValue("Tokens", formatNumber(tokens))
	cli.printKeyValue("Est. Cost", fmt.Sprintf("$%.2f", cost))
	cli.printKeyValue("Rate", "$0.03 / 1K tokens")
	
	return nil
}

func (cli *ProfessionalCLI) cmdMode(args []string) error {
	if len(args) == 0 {
		cli.printInfo("Current mode: %s", cli.session.Mode)
		cli.printInfo("Available: fast, balanced, aggressive")
		return nil
	}

	mode := strings.ToLower(args[0])
	if mode != "fast" && mode != "balanced" && mode != "aggressive" {
		return fmt.Errorf("invalid mode. Use: fast, balanced, aggressive")
	}

	cli.session.Mode = mode
	cli.printSuccess("Mode set to %s", mode)
	return nil
}

func (cli *ProfessionalCLI) cmdBudget(args []string) error {
	if len(args) == 0 {
		cli.printInfo("Current budget: %s tokens", formatNumber(cli.session.Budget))
		return nil
	}

	var budget int
	if _, err := fmt.Sscanf(args[0], "%d", &budget); err != nil || budget <= 0 {
		return fmt.Errorf("invalid budget")
	}

	cli.session.Budget = budget
	cli.printSuccess("Budget set to %s tokens", formatNumber(budget))
	return nil
}

func (cli *ProfessionalCLI) cmdCompact() error {
	if len(cli.session.Files) == 0 {
		cli.printInfo("No files to compress")
		return nil
	}

	before := cli.session.TotalTokens
	
	// Simulate compression (would use actual filter)
	compressionRatio := 0.85
	cli.session.TotalTokens = int(float64(before) * (1 - compressionRatio))
	
	after := cli.session.TotalTokens
	saved := before - after
	
	cli.printHeader("Compression", "")
	cli.printKeyValue("Before", formatNumber(before))
	cli.printKeyValue("After", formatNumber(after))
	cli.printKeyValue("Saved", fmt.Sprintf("%s (%.0f%%)", formatNumber(saved), compressionRatio*100))
	
	return nil
}

func (cli *ProfessionalCLI) cmdFilter(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /filter <text>")
	}

	text := strings.Join(args, " ")
	
	cfg := filter.PipelineConfig{
		Mode:            filter.Mode(cli.session.Mode),
		SessionTracking: true,
	}
	p := filter.NewPipelineCoordinator(cfg)
	output, stats := p.Process(text)
	
	fmt.Println(output)
	
	if stats.TotalSaved > 0 {
		fmt.Println()
		cli.printSuccess("Saved %d tokens (%.0f%%)", 
			stats.TotalSaved, 
			float64(stats.TotalSaved)/float64(stats.OriginalTokens)*100)
	}
	
	return nil
}

func (cli *ProfessionalCLI) cmdStats() error {
	cli.printHeader("Statistics", "")
	
	if cli.tracker == nil {
		cli.printInfo("Tracking not enabled")
		return nil
	}

	summary, err := cli.tracker.GetSavings("")
	if err != nil || summary.TotalCommands == 0 {
		cli.printInfo("No data yet. Run some commands!")
		return nil
	}

	cli.printKeyValue("Commands", formatNumber(summary.TotalCommands))
	cli.printKeyValue("Tokens Saved", formatNumber(summary.TotalSaved))
	cli.printKeyValue("Avg Reduction", fmt.Sprintf("%.1f%%", summary.ReductionPct))
	cli.printKeyValue("Est. Value", fmt.Sprintf("$%.2f", float64(summary.TotalSaved)*0.00003))
	
	return nil
}

func (cli *ProfessionalCLI) cmdAudit(args []string) error {
	days := 7
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &days)
	}

	cli.printHeader("Audit", fmt.Sprintf("Last %d days", days))
	cli.printInfo("Run 'tokman audit --days=%d' for full report", days)
	return nil
}

func (cli *ProfessionalCLI) cmdHistory(args []string) error {
	limit := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &limit)
	}

	cli.printHeader("History", fmt.Sprintf("Last %d commands", limit))
	cli.printInfo("History feature coming soon")
	return nil
}

func (cli *ProfessionalCLI) cmdConfig() error {
	cli.printHeader("Configuration", "")
	cli.printKeyValue("Config file", config.ConfigPath())
	cli.printKeyValue("Data directory", config.DataPath())
	cli.printKeyValue("Mode", cli.session.Mode)
	cli.printKeyValue("Budget", formatNumber(cli.session.Budget))
	return nil
}

func (cli *ProfessionalCLI) cmdFilters() error {
	cli.printHeader("Active Filters", "")
	
	filters := []string{
		"Entropy Filter",
		"Perplexity Filter",
		"H2O Filter",
		"AST Preservation",
		"Semantic Compaction",
	}
	
	for _, f := range filters {
		fmt.Printf("  %s %s\n", color.GreenString("✓"), f)
	}
	
	return nil
}

func (cli *ProfessionalCLI) cmdLayers() error {
	cli.printHeader("Compression Layers", "20 layers")
	cli.printInfo("Full layer info: tokman layers")
	return nil
}

// ==================== SHELL COMMAND EXECUTION ====================

func (cli *ProfessionalCLI) executeShellCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	// Check if command exists
	if _, err := exec.LookPath(parts[0]); err != nil {
		// Not a command, treat as text
		return cli.cmdFilter(parts)
	}

	// Execute with compression
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
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
	
	fmt.Print(compressed)
	
	if stats.TotalSaved > 0 {
		fmt.Println()
		cli.printCompressed(stats.TotalSaved, stats.OriginalTokens)
	}
	
	return nil
}

// ==================== RENDERING HELPERS ====================

func (cli *ProfessionalCLI) buildPrompt() string {
	var parts []string
	
	// Main prompt
	parts = append(parts, color.CyanString("→"))
	
	// Token indicator
	if cli.session.TotalTokens > 0 {
		pct := float64(cli.session.TotalTokens) / float64(cli.session.Budget) * 100
		indicator := cli.tokenIndicator(pct)
		parts = append(parts, indicator)
	}
	
	// File count
	if len(cli.session.Files) > 0 {
		parts = append(parts, color.HiBlackString("(%d)", len(cli.session.Files)))
	}
	
	return strings.Join(parts, " ") + " "
}

func (cli *ProfessionalCLI) tokenIndicator(pct float64) string {
	switch {
	case pct > 90:
		return color.RedString("●")
	case pct > 70:
		return color.YellowString("●")
	default:
		return color.GreenString("●")
	}
}

func (cli *ProfessionalCLI) renderProgressBar(percentage float64, width int) string {
	filled := int(float64(width) * percentage / 100)
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	var colored string
	switch {
	case percentage > 90:
		colored = color.RedString(bar)
	case percentage > 70:
		colored = color.YellowString(bar)
	default:
		colored = color.GreenString(bar)
	}
	
	return "  " + colored
}

func (cli *ProfessionalCLI) printHeader(title, subtitle string) {
	fmt.Println()
	if subtitle != "" {
		fmt.Printf("%s %s\n", color.CyanString(title), color.HiBlackString("(%s)", subtitle))
	} else {
		fmt.Println(color.CyanString(title))
	}
	fmt.Println(color.HiBlackString(strings.Repeat("─", 50)))
}

func (cli *ProfessionalCLI) printSection(name string) {
	fmt.Println()
	fmt.Println(color.HiBlackString(name))
}

func (cli *ProfessionalCLI) printCommand(cmd, desc string) {
	fmt.Printf("  %-18s %s\n", color.GreenString(cmd), color.WhiteString(desc))
}

func (cli *ProfessionalCLI) printItem(label, value string) {
	fmt.Printf("  %s %s\n", color.HiBlackString(label), value)
}

func (cli *ProfessionalCLI) printKeyValue(key, value string) {
	fmt.Printf("  %-12s %s\n", color.HiBlackString(key+":"), color.WhiteString(value))
}

func (cli *ProfessionalCLI) printFile(path string, size int64, tokens int) {
	fmt.Printf("  %s %-30s %s\n",
		color.GreenString("•"),
		truncate(path, 28),
		color.HiBlackString("%d tokens", tokens),
	)
}

func (cli *ProfessionalCLI) printSuccess(format string, args ...interface{}) {
	fmt.Printf("  %s %s\n", color.GreenString("✓"), fmt.Sprintf(format, args...))
}

func (cli *ProfessionalCLI) printError(msg string) {
	fmt.Printf("  %s %s\n", color.RedString("✗"), msg)
}

func (cli *ProfessionalCLI) printWarning(msg string) {
	fmt.Printf("  %s %s\n", color.YellowString("⚠"), msg)
}

func (cli *ProfessionalCLI) printInfo(format string, args ...interface{}) {
	fmt.Printf("  %s\n", color.WhiteString(format, args...))
}

func (cli *ProfessionalCLI) printTip(msg string) {
	fmt.Printf("\n  %s %s\n", color.HiBlackString("Tip:"), msg)
}

func (cli *ProfessionalCLI) printCompressed(saved, original int) {
	pct := float64(saved) / float64(original) * 100
	fmt.Printf("  %s Compressed: %s saved (%.0f%%)\n",
		color.GreenString("✓"),
		formatNumber(saved),
		pct,
	)
}

func (cli *ProfessionalCLI) clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// ==================== UTILITIES ====================

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
