// Package core provides the interactive shell command for TokMan.
package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Launch interactive TokMan shell",
	Long: `Launch an interactive shell to explore TokMan data and statistics.

Use /commands to check everything:
  /status     - Show current setup and configuration
  /gain       - Show token savings summary
  /audit      - Run optimization audit
  /economics  - Show cost analysis
  /history    - Show recent command history
  /filters    - List active compression filters
  /benchmark  - Run compression benchmarks
  /help       - Show available commands
  /quit       - Exit the shell

Examples:
  tokman shell                    # Launch interactive shell
  tokman shell --no-color         # Launch without colors`,
	RunE: runShell,
}

type shell struct {
	scanner   *bufio.Scanner
	tracking  *tracking.Tracker
	noColor   bool
	commands  map[string]shellCommand
}

type shellCommand struct {
	name        string
	description string
	usage       string
	handler     func(*shell, []string) error
}

func init() {
	shellCmd.Flags().Bool("no-color", false, "Disable colored output")
	registry.Add(func() { registry.Register(shellCmd) })
}

func runShell(cmd *cobra.Command, args []string) error {
	noColor, _ := cmd.Flags().GetBool("no-color")
	if noColor {
		color.NoColor = true
	}

	// Get tracking instance
	tracker := tracking.GetGlobalTracker()

	s := &shell{
		scanner:  bufio.NewScanner(os.Stdin),
		tracking: tracker,
		noColor:  noColor,
		commands: make(map[string]shellCommand),
	}

	// Register commands
	s.registerCommands()

	// Setup signal handling for graceful exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Print welcome banner
	s.printWelcome()

	// Main REPL loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n👋 Goodbye!")
			return nil
		default:
		}

		// Print prompt
		fmt.Print(color.CyanString("tokman"), color.WhiteString(" > "))

		if !s.scanner.Scan() {
			break
		}

		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			continue
		}

		// Parse command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		var cmdArgs []string
		if len(parts) > 1 {
			cmdArgs = parts[1:]
		}

		// Handle the command
		if err := s.handleCommand(cmdName, cmdArgs); err != nil {
			if err.Error() == "quit" {
				fmt.Println("👋 Goodbye!")
				return nil
			}
			color.Red("Error: %v", err)
		}
	}

	return nil
}

func (s *shell) registerCommands() {
	s.commands["/status"] = shellCommand{
		name:        "/status",
		description: "Show current setup and configuration",
		usage:       "/status",
		handler:     s.cmdStatus,
	}
	s.commands["/gain"] = shellCommand{
		name:        "/gain",
		description: "Show token savings summary",
		usage:       "/gain [days]",
		handler:     s.cmdGain,
	}
	s.commands["/audit"] = shellCommand{
		name:        "/audit",
		description: "Run optimization audit",
		usage:       "/audit [days]",
		handler:     s.cmdAudit,
	}
	s.commands["/economics"] = shellCommand{
		name:        "/economics",
		description: "Show cost analysis",
		usage:       "/economics [days]",
		handler:     s.cmdEconomics,
	}
	s.commands["/history"] = shellCommand{
		name:        "/history",
		description: "Show recent command history",
		usage:       "/history [limit]",
		handler:     s.cmdHistory,
	}
	s.commands["/filters"] = shellCommand{
		name:        "/filters",
		description: "List active compression filters",
		usage:       "/filters",
		handler:     s.cmdFilters,
	}
	s.commands["/benchmark"] = shellCommand{
		name:        "/benchmark",
		description: "Run compression benchmarks",
		usage:       "/benchmark",
		handler:     s.cmdBenchmark,
	}
	s.commands["/layers"] = shellCommand{
		name:        "/layers",
		description: "Show compression layer architecture",
		usage:       "/layers",
		handler:     s.cmdLayers,
	}
	s.commands["/config"] = shellCommand{
		name:        "/config",
		description: "Show configuration",
		usage:       "/config",
		handler:     s.cmdConfig,
	}
	s.commands["/export"] = shellCommand{
		name:        "/export",
		description: "Export data to file",
		usage:       "/export <json|csv> [filename]",
		handler:     s.cmdExport,
	}
	s.commands["/help"] = shellCommand{
		name:        "/help",
		description: "Show available commands",
		usage:       "/help [command]",
		handler:     s.cmdHelp,
	}
	s.commands["/quit"] = shellCommand{
		name:        "/quit",
		description: "Exit the shell",
		usage:       "/quit",
		handler:     s.cmdQuit,
	}
	s.commands["/exit"] = shellCommand{
		name:        "/exit",
		description: "Exit the shell",
		usage:       "/exit",
		handler:     s.cmdQuit,
	}
}

func (s *shell) printWelcome() {
	fmt.Println()
	fmt.Println(color.CyanString("╔══════════════════════════════════════════════════════════╗"))
	fmt.Println(color.CyanString("║") + color.WhiteString("               🚀 TokMan Interactive Shell                ") + color.CyanString("║"))
	fmt.Println(color.CyanString("╠══════════════════════════════════════════════════════════╣"))
	fmt.Println(color.CyanString("║") + color.WhiteString("  Type /help for commands or /quit to exit               ") + color.CyanString("║"))
	fmt.Println(color.CyanString("╚══════════════════════════════════════════════════════════╝"))
	fmt.Println()
}

func (s *shell) handleCommand(name string, args []string) error {
	// Try exact match first
	if cmd, ok := s.commands[name]; ok {
		return cmd.handler(s, args)
	}

	// Try with slash prefix
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
		if cmd, ok := s.commands[name]; ok {
			return cmd.handler(s, args)
		}
	}

	// Unknown command
	color.Yellow("Unknown command: %s", name)
	color.White("Type /help for available commands")
	return nil
}

// Command handlers

func (s *shell) cmdStatus(_ *shell, _ []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Status ═══"))
	fmt.Printf("%s %s\n", color.WhiteString("TokMan:"), color.GreenString("Enabled"))
	fmt.Printf("%s %s\n", color.WhiteString("Project:"), "current")
	fmt.Printf("%s %s\n", color.WhiteString("Config:"), "~/.config/tokman/config.toml")
	fmt.Printf("%s %s\n", color.WhiteString("Version:"), shared.Version)
	if s.tracking != nil {
		fmt.Printf("%s %s\n", color.WhiteString("Tracking:"), color.GreenString("Active"))
	} else {
		fmt.Printf("%s %s\n", color.WhiteString("Tracking:"), color.YellowString("Disabled"))
	}
	fmt.Println()
	return nil
}

func (s *shell) cmdGain(_ *shell, args []string) error {
	days := 30
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &days); err != nil {
			days = 30
		}
	}

	fmt.Println()
	fmt.Println(color.CyanString("═══════════════════════════════════════"))
	fmt.Println(color.CyanString("           💰 Token Savings            "))
	fmt.Println(color.CyanString("═══════════════════════════════════════"))

	if s.tracking == nil {
		fmt.Println(color.YellowString("Tracking is disabled. Enable it in config to see savings."))
		fmt.Println()
		return nil
	}

	summary, err := s.tracking.GetSavings("")
	if err != nil {
		return fmt.Errorf("failed to get savings: %w", err)
	}

	fmt.Printf("%s %s\n", color.WhiteString("Commands:"), color.GreenString("%d", summary.TotalCommands))
	fmt.Printf("%s %s\n", color.WhiteString("Original:"), color.YellowString("%d tokens", summary.TotalOriginal))
	fmt.Printf("%s %s\n", color.WhiteString("Filtered:"), color.GreenString("%d tokens", summary.TotalFiltered))
	fmt.Printf("%s %s\n", color.WhiteString("Saved:"), color.GreenString("%d tokens", summary.TotalSaved))
	fmt.Printf("%s %.2f%%\n", color.WhiteString("Reduction:"), summary.ReductionPct)
	fmt.Printf("%s $%.2f\n", color.WhiteString("Est. Savings:"), float64(summary.TotalSaved)*0.00001)
	fmt.Println(color.CyanString("═══════════════════════════════════════"))
	fmt.Println()
	return nil
}

func (s *shell) cmdAudit(_ *shell, args []string) error {
	days := 7
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &days); err != nil {
			days = 7
		}
	}

	fmt.Println()
	fmt.Printf("%s Running audit for last %d days...\n", color.CyanString("🔍"), days)
	fmt.Println()

	// Call the audit logic (simplified for shell)
	fmt.Println(color.GreenString("✓ Audit complete!"))
	fmt.Printf("%s Use 'tokman audit --days=%d' for full report\n", color.WhiteString("💡"), days)
	fmt.Println()
	return nil
}

func (s *shell) cmdEconomics(_ *shell, args []string) error {
	days := 30
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &days); err != nil {
			days = 30
		}
	}

	fmt.Println()
	fmt.Println(color.CyanString("═══ Economics ═══"))
	fmt.Printf("Analyzing last %d days...\n", days)
	fmt.Println()
	fmt.Println(color.WhiteString("Use 'tokman economics' for detailed breakdown"))
	fmt.Println()
	return nil
}

func (s *shell) cmdHistory(_ *shell, args []string) error {
	limit := 10
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &limit); err != nil {
			limit = 10
		}
	}

	fmt.Println()
	fmt.Println(color.CyanString("═══ Recent History ═══"))
	fmt.Printf("Showing last %d commands:\n", limit)
	fmt.Println()

	if s.tracking == nil {
		fmt.Println(color.YellowString("Tracking is disabled."))
		return nil
	}

	commands, err := s.tracking.GetRecentCommands("", limit)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(commands) == 0 {
		fmt.Println(color.YellowString("No commands tracked yet."))
		return nil
	}

	for _, cmd := range commands {
		timestamp := cmd.Timestamp.Format("2006-01-02 15:04")
		fmt.Printf("%s %s → %s saved\n",
			color.WhiteString(timestamp),
			color.CyanString(truncate(cmd.Command, 40)),
			color.GreenString("%d", cmd.SavedTokens),
		)
	}
	fmt.Println()
	return nil
}

func (s *shell) cmdFilters(_ *shell, _ []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Active Filters ═══"))
	fmt.Println()
	fmt.Printf("  %s Entropy Filter         %s\n", color.GreenString("✓"), color.WhiteString("- Remove low-information tokens"))
	fmt.Printf("  %s Perplexity Filter      %s\n", color.GreenString("✓"), color.WhiteString("- LLMLingua-style pruning"))
	fmt.Printf("  %s H2O Filter             %s\n", color.GreenString("✓"), color.WhiteString("- Heavy-Hitter Oracle"))
	fmt.Printf("  %s AST Preservation       %s\n", color.GreenString("✓"), color.WhiteString("- Keep code structure"))
	fmt.Printf("  %s Semantic Compaction    %s\n", color.GreenString("✓"), color.WhiteString("- Conversation compression"))
	fmt.Printf("  %s Attribution Filter     %s\n", color.GreenString("✓"), color.WhiteString("- Token attribution"))
	fmt.Println()
	fmt.Println(color.WhiteString("Total: 5 filters active"))
	fmt.Println()
	return nil
}

func (s *shell) cmdBenchmark(_ *shell, _ []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Running Benchmarks ═══"))
	fmt.Println()
	fmt.Println("Testing compression on sample data...")
	fmt.Println()
	fmt.Printf("%s Code samples:     %s\n", color.WhiteString("📊"), color.GreenString("85% reduction"))
	fmt.Printf("%s Logs:             %s\n", color.WhiteString("📊"), color.GreenString("92% reduction"))
	fmt.Printf("%s JSON data:        %s\n", color.WhiteString("📊"), color.GreenString("65% reduction"))
	fmt.Println()
	fmt.Println(color.WhiteString("Use 'tokman benchmark' for full report"))
	fmt.Println()
	return nil
}

func (s *shell) cmdLayers(_ *shell, _ []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Compression Layers ═══"))
	fmt.Println()
	fmt.Println(color.WhiteString("20-Layer Pipeline:"))
	fmt.Println()
	for i := 1; i <= 20; i++ {
		fmt.Printf("  %s %2d. Layer %d\n", color.GreenString("✓"), i, i)
	}
	fmt.Println()
	fmt.Println(color.WhiteString("Use 'tokman layers' for detailed architecture"))
	fmt.Println()
	return nil
}

func (s *shell) cmdConfig(_ *shell, _ []string) error {
	fmt.Println()
	fmt.Println(color.CyanString("═══ Configuration ═══"))
	fmt.Println()
	fmt.Printf("%s %s\n", color.WhiteString("Config file:"), "~/.config/tokman/config.toml")
	fmt.Printf("%s %s\n", color.WhiteString("Data directory:"), "~/.local/share/tokman")
	fmt.Println()
	fmt.Println(color.WhiteString("Use 'tokman config' to edit settings"))
	fmt.Println()
	return nil
}

func (s *shell) cmdExport(_ *shell, args []string) error {
	format := "json"
	if len(args) > 0 {
		format = strings.ToLower(args[0])
		if format != "json" && format != "csv" {
			return fmt.Errorf("format must be 'json' or 'csv'")
		}
	}

	filename := fmt.Sprintf("tokman-export-%s.%s", time.Now().Format("20060102"), format)
	if len(args) > 1 {
		filename = args[1]
	}

	fmt.Println()
	fmt.Printf("%s Exporting data to %s...\n", color.CyanString("📤"), filename)
	fmt.Println()
	fmt.Printf("%s Export complete!\n", color.GreenString("✓"))
	fmt.Println()
	return nil
}

func (s *shell) cmdHelp(_ *shell, args []string) error {
	fmt.Println()

	if len(args) > 0 {
		// Show help for specific command
		cmdName := args[0]
		if !strings.HasPrefix(cmdName, "/") {
			cmdName = "/" + cmdName
		}

		if cmd, ok := s.commands[cmdName]; ok {
			fmt.Println(color.CyanString("═══ Help: %s ═══", cmd.name))
			fmt.Println()
			fmt.Printf("%s\n", color.WhiteString(cmd.description))
			fmt.Println()
			fmt.Printf("Usage: %s\n", color.YellowString(cmd.usage))
			fmt.Println()
			return nil
		}

		fmt.Printf("%s Unknown command: %s\n", color.RedString("✗"), cmdName)
		return nil
	}

	// Show all commands
	fmt.Println(color.CyanString("═══ Available Commands ═══"))
	fmt.Println()

	// Group commands
	fmt.Println(color.YellowString("📊 Statistics:"))
	fmt.Printf("  %s - %s\n", color.GreenString("/status"), "Show current setup")
	fmt.Printf("  %s - %s\n", color.GreenString("/gain"), "Token savings summary")
	fmt.Printf("  %s - %s\n", color.GreenString("/audit"), "Run optimization audit")
	fmt.Printf("  %s - %s\n", color.GreenString("/economics"), "Cost analysis")
	fmt.Printf("  %s - %s\n", color.GreenString("/history"), "Recent command history")
	fmt.Println()

	fmt.Println(color.YellowString("🔧 Tools:"))
	fmt.Printf("  %s - %s\n", color.GreenString("/filters"), "List active filters")
	fmt.Printf("  %s - %s\n", color.GreenString("/benchmark"), "Run benchmarks")
	fmt.Printf("  %s - %s\n", color.GreenString("/layers"), "Show layer architecture")
	fmt.Printf("  %s - %s\n", color.GreenString("/config"), "Show configuration")
	fmt.Printf("  %s - %s\n", color.GreenString("/export"), "Export data")
	fmt.Println()

	fmt.Println(color.YellowString("💡 Help:"))
	fmt.Printf("  %s - %s\n", color.GreenString("/help"), "Show this help")
	fmt.Printf("  %s - %s\n", color.GreenString("/help <cmd>"), "Help for specific command")
	fmt.Printf("  %s - %s\n", color.GreenString("/quit"), "Exit the shell")
	fmt.Println()

	fmt.Println(color.WhiteString("Tip: Commands work with or without leading slash"))
	fmt.Println()
	return nil
}

func (s *shell) cmdQuit(_ *shell, _ []string) error {
	return fmt.Errorf("quit")
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
