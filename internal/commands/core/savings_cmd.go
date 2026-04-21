package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/telemetry"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var (
	gainProject   bool
	gainGraph     bool
	gainHistory   bool
	gainDaily     bool
	gainWeekly    bool
	gainMonthly   bool
	gainAll       bool
	gainFormat    string
	gainFailures  bool
	gainSinceDays int
	gainQuota     string // quota estimation tier (pro, 5x, 20x)
)

var gainCmd = &cobra.Command{
	Use:   "gain",
	Short: "Show token savings analytics",
	Long: `Display comprehensive token savings statistics with various views.

Examples:
  tok gain                    # Default summary view
  tok gain --graph            # ASCII bar chart of last 30 days
  tok gain --daily            # Day-by-day breakdown
  tok gain --history          # Recent command history
  tok gain --format json      # JSON export
  tok gain --project          # Show only current project
  tok gain --quota pro        # Estimate quota usage (pro, 5x, 20x)`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: runGain,
}

func runGain(cmd *cobra.Command, args []string) error {
	dbPath := config.DatabasePath()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		out.Global().Println("No tracking data found.")
		out.Global().Println("Run some commands through tok to start tracking token savings!")
		out.Global().Printf("\nExample: %s git status\n", os.Args[0])
		return nil
	}

	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize tracker: %w", err)
	}
	defer tracker.Close()

	// Handle failures-only view
	if gainFailures {
		return showFailures(tracker)
	}

	// Determine project scope
	projectPath := ""
	if gainProject {
		projectPath = config.ProjectPath()
	}

	// Build options
	opts := tracking.GainSummaryOptions{
		ProjectPath:    projectPath,
		IncludeDaily:   gainDaily || gainAll,
		IncludeWeekly:  gainWeekly || gainAll,
		IncludeMonthly: gainMonthly || gainAll,
		IncludeHistory: gainHistory,
	}

	summary, err := tracker.GetFullGainSummary(opts)
	if err != nil {
		return fmt.Errorf("failed to get gain summary: %w", err)
	}

	// Handle quota estimation
	if gainQuota != "" {
		return printQuotaEstimation(summary, gainQuota)
	}

	// Handle export formats
	switch gainFormat {
	case "json":
		return exportGainJSON(summary)
	case "csv":
		return exportGainCSV(summary, opts)
	default:
		return printGainText(summary, opts)
	}
}

func printGainText(summary *tracking.GainSummary, opts tracking.GainSummaryOptions) error {
	// Print header
	title := "tok Token Savings"
	if opts.ProjectPath != "" {
		title = "tok Token Savings (Project Scope)"
	}

	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint(title))
	out.Global().Println(strings.Repeat("═", 60))
	if opts.ProjectPath != "" {
		out.Global().Printf("Scope: %s\n", shortenPath(opts.ProjectPath))
	}
	out.Global().Println()

	// Print KPIs
	printKPI("Total commands", fmt.Sprintf("%d", summary.TotalCommands))
	printKPI("Input tokens", formatTokensInt(summary.TotalInput))
	printKPI("Output tokens", formatTokensInt(summary.TotalOutput))
	printKPI("Tokens saved", fmt.Sprintf("%s (%.1f%%)",
		formatTokensInt(summary.TotalSaved), summary.AvgSavingsPct))

	// Calculate cost savings (using default Claude Sonnet pricing)
	if summary.TotalSaved > 0 {
		estimator := tracking.NewCostEstimator("claude-3-sonnet")
		costSaved := estimator.EstimateSavings(summary.TotalSaved).EstimatedSavings
		printKPI("Cost saved", fmt.Sprintf("$%.2f USD", costSaved))
	}

	printKPI("Total exec time", formatDuration(summary.TotalExecTimeMs))
	printEfficiencyMeter(summary.AvgSavingsPct)
	out.Global().Println()

	// Print command breakdown
	if len(summary.ByCommand) > 0 {
		out.Global().Println(color.New(color.Bold).Sprint("By Command"))
		out.Global().Println(strings.Repeat("─", 60))

		for _, cmd := range summary.ByCommand {
			impact := ""
			if cmd.SavingsPct >= 80 {
				impact = color.GreenString("high")
			} else if cmd.SavingsPct >= 50 {
				impact = color.YellowString("med")
			} else {
				impact = color.RedString("low")
			}

			out.Global().Printf("  %-24s %4d  %8s  %s\n",
				truncate(cmd.Command, 24),
				cmd.Count,
				formatTokensInt(cmd.SavedTokens),
				impact,
			)
		}
		out.Global().Println()
	}

	// Print daily stats if requested
	if gainGraph && len(summary.DailyStats) > 0 {
		printASCIIGraph(summary.DailyStats)
	}

	// Print daily table if requested
	if gainDaily && len(summary.DailyStats) > 0 {
		printDailyTable(summary.DailyStats)
	}

	// Print history if requested
	if gainHistory && len(summary.RecentCommands) > 0 {
		printHistoryTable(summary.RecentCommands)
	}

	return nil
}

func printKPI(label, value string) {
	out.Global().Printf("  %-20s %s\n", label+":", value)
}

func printEfficiencyMeter(pct float64) {
	width := 40
	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	out.Global().Printf("  Efficiency: [%s] %.0f%%\n", bar, pct)
}

func printASCIIGraph(stats []tracking.PeriodStats) {
	if len(stats) == 0 {
		return
	}

	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint("Token Savings (Last 30 Days)"))
	out.Global().Println("┌" + strings.Repeat("─", 62) + "┐")

	// Find max for scaling
	maxSaved := 0
	totalOriginal := 0
	totalFiltered := 0
	for _, s := range stats {
		if s.SavedTokens > maxSaved {
			maxSaved = s.SavedTokens
		}
		totalOriginal += s.InputTokens
		totalFiltered += s.OutputTokens
	}
	if maxSaved == 0 {
		maxSaved = 1
	}

	// Print graph (last 30 days)
	startIdx := 0
	if len(stats) > 30 {
		startIdx = len(stats) - 30
	}

	for i := len(stats) - 1; i >= startIdx; i-- {
		s := stats[i]
		barLen := int((float64(s.SavedTokens) / float64(maxSaved)) * 30)
		if barLen < 0 {
			barLen = 0
		}
		bar := strings.Repeat("█", barLen)

		dateLabel := s.Period
		if len(dateLabel) > 10 {
			dateLabel = dateLabel[5:] // Remove year
		}

		pctStr := ""
		if s.InputTokens > 0 {
			pct := float64(s.SavedTokens) / float64(s.InputTokens) * 100
			pctStr = fmt.Sprintf(" %.0f%%", pct)
		}

		out.Global().Printf("│ %-10s %-30s %6s%s │\n", dateLabel, bar, formatTokensInt(s.SavedTokens), pctStr)
	}
	out.Global().Println("└" + strings.Repeat("─", 62) + "┘")

	totalSaved := totalOriginal - totalFiltered
	if totalSaved < 0 {
		totalSaved = 0
	}
	overallPct := 0.0
	if totalOriginal > 0 {
		overallPct = float64(totalSaved) / float64(totalOriginal) * 100
	}
	out.Global().Printf("Total: %s → %s tokens (%.1f%% saved)\n",
		formatTokensInt(totalOriginal), formatTokensInt(totalFiltered), overallPct)
	out.Global().Println()
}

func exportGainJSON(summary *tracking.GainSummary) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

func exportGainCSV(summary *tracking.GainSummary, opts tracking.GainSummaryOptions) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write summary header
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Commands", fmt.Sprintf("%d", summary.TotalCommands)})
	writer.Write([]string{"Input Tokens", fmt.Sprintf("%d", summary.TotalInput)})
	writer.Write([]string{"Output Tokens", fmt.Sprintf("%d", summary.TotalOutput)})
	writer.Write([]string{"Tokens Saved", fmt.Sprintf("%d", summary.TotalSaved)})
	writer.Write([]string{"Savings %", fmt.Sprintf("%.2f", summary.AvgSavingsPct)})
	writer.Write([]string{})

	// Write command breakdown
	writer.Write([]string{"Command", "Count", "Input", "Output", "Saved", "Savings %"})
	for _, cmd := range summary.ByCommand {
		writer.Write([]string{
			cmd.Command,
			fmt.Sprintf("%d", cmd.Count),
			fmt.Sprintf("%d", cmd.InputTokens),
			fmt.Sprintf("%d", cmd.OutputTokens),
			fmt.Sprintf("%d", cmd.SavedTokens),
			fmt.Sprintf("%.2f", cmd.SavingsPct),
		})
	}

	return nil
}

func showFailures(tracker *tracking.Tracker) error {
	// Get recent commands with failures
	commands, err := tracker.GetRecentCommands("", 50)
	if err != nil {
		return err
	}

	out.Global().Println(color.New(color.Bold).Sprint("Recent Failures"))
	out.Global().Println(strings.Repeat("─", 60))

	found := false
	for _, cmd := range commands {
		if !cmd.ParseSuccess {
			found = true
			out.Global().Printf("  %s  %s\n",
				cmd.Timestamp.Format("Jan 02 15:04"),
				cmd.Command)
		}
	}

	if !found {
		out.Global().Println("  No recent failures found.")
	}

	return nil
}

// Helper functions
func formatTokensInt(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	if ms < 60_000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}
	mins := ms / 60_000
	secs := (ms % 60_000) / 1000
	return fmt.Sprintf("%dm%ds", mins, secs)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func printHistoryTable(cmds []tracking.CommandRecord) {
	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint("Recent Commands"))
	out.Global().Println("┌" + strings.Repeat("─", 20) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 8) + "┐")
	out.Global().Printf("│ %-18s │ %8s │ %8s │ %6s │\n", "Command", "Original", "Filtered", "Saved")
	out.Global().Println("├" + strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 8) + "┤")

	for _, cmd := range cmds {
		pct := 0.0
		if cmd.OriginalTokens > 0 {
			pct = float64(cmd.SavedTokens) / float64(cmd.OriginalTokens) * 100
		}
		out.Global().Printf("│ %-18s │ %8s │ %8s │ %5.1f%% │\n",
			truncate(cmd.Command, 18),
			formatTokensInt(cmd.OriginalTokens),
			formatTokensInt(cmd.FilteredTokens),
			pct,
		)
	}
	out.Global().Println("└" + strings.Repeat("─", 20) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 8) + "┘")
	out.Global().Println()
}

func printDailyTable(stats []tracking.PeriodStats) {
	if len(stats) == 0 {
		return
	}

	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint("Day-by-Day Breakdown"))
	out.Global().Println("┌" + strings.Repeat("─", 12) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 8) + "┐")
	out.Global().Printf("│ %-10s │ %8s │ %8s │ %8s │ %6s │\n", "Date", "Commands", "Original", "Filtered", "Saved")
	out.Global().Println("├" + strings.Repeat("─", 12) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 8) + "┤")

	for i := len(stats) - 1; i >= 0; i-- {
		s := stats[i]
		pct := 0.0
		if s.InputTokens > 0 {
			pct = float64(s.SavedTokens) / float64(s.InputTokens) * 100
		}
		out.Global().Printf("│ %-10s │ %8d │ %8s │ %8s │ %5.1f%% │\n",
			s.Period,
			s.Commands,
			formatTokensInt(s.InputTokens),
			formatTokensInt(s.OutputTokens),
			pct,
		)
	}
	out.Global().Println("└" + strings.Repeat("─", 12) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 8) + "┘")
	out.Global().Println()
}

// printQuotaEstimation shows subscription tier usage estimation.
func printQuotaEstimation(summary *tracking.GainSummary, tier string) error {
	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint("Quota Estimation"))
	out.Global().Println(strings.Repeat("═", 60))

	// Define tier limits (approximate token limits per tier)
	tierLimits := map[string]int{
		"free":      1_000_000,   // 1M tokens
		"pro":       5_000_000,   // 5M tokens
		"5x":        25_000_000,  // 25M tokens
		"20x":       100_000_000, // 100M tokens
		"unlimited": 999_999_999, // Effectively unlimited
	}

	limit, ok := tierLimits[tier]
	if !ok {
		// Try to parse as number
		out.Global().Printf("Unknown tier '%s'. Using 'pro' as default.\n", tier)
		limit = tierLimits["pro"]
		tier = "pro"
	}

	// Calculate usage
	inputTokens := summary.TotalInput
	outputTokens := summary.TotalOutput
	totalTokens := inputTokens + outputTokens

	// Estimate daily usage from available data
	days := 30 // Assume 30 days if no daily data
	if len(summary.DailyStats) > 0 {
		days = len(summary.DailyStats)
	}

	avgDaily := totalTokens / days
	monthlyProjection := avgDaily * 30

	// Calculate percentage of tier
	usagePct := float64(monthlyProjection) / float64(limit) * 100

	// Display results
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	out.Global().Printf("Subscription Tier:  %s\n", bold(tier))
	out.Global().Printf("Monthly Limit:      %s tokens\n", formatTokensInt(limit))
	out.Global().Println()

	out.Global().Printf("Current Period (%d days):\n", days)
	out.Global().Printf("  Input tokens:     %s\n", cyan(formatTokensInt(inputTokens)))
	out.Global().Printf("  Output tokens:    %s\n", cyan(formatTokensInt(outputTokens)))
	out.Global().Printf("  Total processed:  %s\n", bold(formatTokensInt(totalTokens)))
	out.Global().Println()

	out.Global().Printf("Projected Monthly Usage:\n")
	out.Global().Printf("  Daily average:    %s tokens/day\n", formatTokensInt(avgDaily))
	out.Global().Printf("  30-day estimate:  %s tokens\n", bold(formatTokensInt(monthlyProjection)))

	// Usage bar
	barWidth := 40
	filled := int((usagePct / 100.0) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	var bar string
	if usagePct < 50 {
		bar = green(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
	} else if usagePct < 80 {
		bar = yellow(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
	} else {
		bar = red(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
	}

	out.Global().Printf("  Tier usage:       [%s] %.1f%%\n", bar, usagePct)

	// Savings from tok
	savingsTokens := summary.TotalSaved
	savingsPct := float64(0)
	if inputTokens > 0 {
		savingsPct = float64(savingsTokens) / float64(inputTokens) * 100
	}

	out.Global().Println()
	out.Global().Println(bold("tok Savings Impact:"))
	out.Global().Printf("  Tokens saved:     %s (%.1f%%)\n", green(formatTokensInt(savingsTokens)), savingsPct)

	// Calculate effective cost without tok
	if savingsTokens > 0 {
		effectiveWithouttok := monthlyProjection + (savingsTokens / days * 30)
		out.Global().Printf("  Without tok:   ~%s tokens/month\n", formatTokensInt(effectiveWithouttok))
		out.Global().Printf("  Effective tier:   %s\n", cyan(getTierForTokens(effectiveWithouttok, tierLimits)))
	}

	// Recommendation
	out.Global().Println()
	out.Global().Println(bold("Recommendation:"))
	if usagePct < 50 {
		out.Global().Println(green("  ✓ Current tier is sufficient"))
	} else if usagePct < 80 {
		out.Global().Println(yellow("  ⚠ Approaching tier limit, monitor usage"))
	} else {
		out.Global().Println(red("  ✗ Consider upgrading tier or increasing compression"))
	}

	out.Global().Println()

	// Track telemetry
	telemetry.TrackQuotaUsage(tier, usagePct)

	return nil
}

// getTierForTokens returns the recommended tier for a given token count
func getTierForTokens(tokens int, tierLimits map[string]int) string {
	if tokens < tierLimits["free"] {
		return "free"
	}
	if tokens < tierLimits["pro"] {
		return "pro"
	}
	if tokens < tierLimits["5x"] {
		return "5x"
	}
	if tokens < tierLimits["20x"] {
		return "20x"
	}
	return "unlimited"
}

func init() {
	registry.Add(func() { registry.Register(gainCmd) })

	gainCmd.Flags().BoolVarP(&gainProject, "project", "p", false, "Show only current project stats")
	gainCmd.Flags().BoolVar(&gainGraph, "graph", false, "Show ASCII graph of daily savings")
	gainCmd.Flags().BoolVar(&gainHistory, "history", false, "Show recent command history")
	gainCmd.Flags().BoolVar(&gainDaily, "daily", false, "Show day-by-day breakdown")
	gainCmd.Flags().BoolVar(&gainWeekly, "weekly", false, "Show week-by-week breakdown")
	gainCmd.Flags().BoolVar(&gainMonthly, "monthly", false, "Show month-by-month breakdown")
	gainCmd.Flags().BoolVarP(&gainAll, "all", "a", false, "Show all breakdowns")
	gainCmd.Flags().StringVarP(&gainFormat, "format", "f", "text", "Output format: text, json, csv")
	gainCmd.Flags().BoolVar(&gainFailures, "failures", false, "Show only parse failures")
	gainCmd.Flags().IntVarP(&gainSinceDays, "since", "s", 30, "Limit to last N days")
	gainCmd.Flags().StringVarP(&gainQuota, "quota", "t", "", "Estimate quota usage (pro, 5x, 20x)")
}
