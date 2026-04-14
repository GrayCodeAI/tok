package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var (
	ccEconDaily   bool
	ccEconWeekly  bool
	ccEconMonthly bool
	ccEconAll     bool
	ccEconFormat  string
)

var ccEconCmd = &cobra.Command{
	Use:   "cc-economics",
	Short: "Claude Code economics: spending vs savings analysis",
	Long: `Analyze the economics of using Claude Code with TokMan.

Compares your Claude Code API spending (via ccusage) with the token savings
achieved by using TokMan's compression pipeline.

Shows:
- Claude Code API costs (input/output/cache tokens)
- TokMan token savings (tokens filtered/compressed)
- Effective cost reduction percentage
- ROI analysis

Examples:
  tokman cc-economics --daily
  tokman cc-economics --weekly
  tokman cc-economics --monthly
  tokman cc-economics --all`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
	},
	RunE: runCcEconomics,
}

func init() {
	registry.Add(func() { registry.Register(ccEconCmd) })
	ccEconCmd.Flags().BoolVarP(&ccEconDaily, "daily", "d", false, "Show daily breakdown")
	ccEconCmd.Flags().BoolVarP(&ccEconWeekly, "weekly", "w", false, "Show weekly breakdown")
	ccEconCmd.Flags().BoolVarP(&ccEconMonthly, "monthly", "m", false, "Show monthly breakdown")
	ccEconCmd.Flags().BoolVarP(&ccEconAll, "all", "a", false, "Show all breakdowns")
	ccEconCmd.Flags().StringVarP(&ccEconFormat, "format", "f", "text", "Output format (text, json, csv)")
}

// CcUsagePeriod represents usage data for a time period
type CcUsagePeriod struct {
	Date                string  `json:"date"`
	InputTokens         uint64  `json:"inputTokens"`
	OutputTokens        uint64  `json:"outputTokens"`
	CacheCreationTokens uint64  `json:"cacheCreationTokens"`
	CacheReadTokens     uint64  `json:"cacheReadTokens"`
	TotalTokens         uint64  `json:"totalTokens"`
	TotalCost           float64 `json:"totalCost"`
}

// TokManSavings represents TokMan savings for a period
type TokManSavings struct {
	Date         string `json:"date"`
	Commands     int    `json:"commands"`
	SavedTokens  uint64 `json:"savedTokens"`
	OriginalSize uint64 `json:"originalSize"`
	FilteredSize uint64 `json:"filteredSize"`
}

// EconomicsReport combines ccusage and TokMan data
type EconomicsReport struct {
	Period         string        `json:"period"`
	CcUsage        CcUsagePeriod `json:"ccUsage"`
	TokManSavings  TokManSavings `json:"tokManSavings"`
	EffectiveCost  float64       `json:"effectiveCost"`
	SavingsPercent float64       `json:"savingsPercent"`
}

func runCcEconomics(cmd *cobra.Command, args []string) error {
	// Determine granularity
	granularities := []string{"daily"}
	if ccEconWeekly {
		granularities = []string{"weekly"}
	} else if ccEconMonthly {
		granularities = []string{"monthly"}
	} else if ccEconAll {
		granularities = []string{"daily", "weekly", "monthly"}
	}

	// Get ccusage data
	ccusageData, err := fetchCcusageData(granularities)
	if err != nil {
		// Don't fail if ccusage is not available, just show TokMan data
		if shared.Verbose > 0 {
			fmt.Fprintf(os.Stderr, "Note: ccusage not available: %v\n", err)
		}
	}

	// Get TokMan savings data
	tokmanData, err := fetchTokManSavings(granularities)
	if err != nil {
		return fmt.Errorf("failed to fetch TokMan savings: %w", err)
	}

	// Generate reports
	reports := generateEconomicsReports(ccusageData, ccusageData, tokmanData)

	// Output
	switch ccEconFormat {
	case "json":
		return outputJson(reports)
	case "csv":
		return outputCsv(reports)
	default:
		return outputText(reports, granularities)
	}
}

func fetchCcusageData(granularities []string) (map[string][]CcUsagePeriod, error) {
	result := make(map[string][]CcUsagePeriod)

	// Check if ccusage is available
	ccusagePath, err := findCcusage()
	if err != nil {
		return nil, err
	}

	for _, g := range granularities {
		cmd := exec.Command(ccusagePath, g, "--json", "--since", "20250101")
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue // Skip unavailable granularities
		}

		periods, err := parseCcusageOutput(string(output), g)
		if err != nil {
			continue
		}
		result[g] = periods
	}

	return result, nil
}

func findCcusage() (string, error) {
	if _, err := exec.LookPath("ccusage"); err == nil {
		return "ccusage", nil
	}
	// Try npx
	npxCheck := exec.Command("npx", "ccusage", "--help")
	npxCheck.Stdout = nil
	npxCheck.Stderr = nil
	if err := npxCheck.Run(); err == nil {
		return "npx ccusage", nil
	}
	return "", fmt.Errorf("ccusage not found")
}

func parseCcusageOutput(jsonStr, granularity string) ([]CcUsagePeriod, error) {
	var periods []CcUsagePeriod

	switch granularity {
	case "daily":
		var resp struct {
			Daily []struct {
				Date string `json:"date"`
				CcUsagePeriod
			} `json:"daily"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			return nil, err
		}
		for _, d := range resp.Daily {
			p := d.CcUsagePeriod
			p.Date = d.Date
			periods = append(periods, p)
		}

	case "weekly":
		var resp struct {
			Weekly []struct {
				Week string `json:"week"`
				CcUsagePeriod
			} `json:"weekly"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			return nil, err
		}
		for _, w := range resp.Weekly {
			p := w.CcUsagePeriod
			p.Date = w.Week
			periods = append(periods, p)
		}

	case "monthly":
		var resp struct {
			Monthly []struct {
				Month string `json:"month"`
				CcUsagePeriod
			} `json:"monthly"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			return nil, err
		}
		for _, m := range resp.Monthly {
			p := m.CcUsagePeriod
			p.Date = m.Month
			periods = append(periods, p)
		}
	}

	return periods, nil
}

func fetchTokManSavings(granularities []string) (map[string][]TokManSavings, error) {
	result := make(map[string][]TokManSavings)

	dbPath := config.DatabasePath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// No database yet, return empty
		return result, nil
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Check if commands table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='commands'").Scan(&tableName)
	if err != nil {
		// Table doesn't exist yet
		return result, nil
	}

	// Daily savings
	if contains(granularities, "daily") {
		rows, err := db.Query(`
			SELECT 
				DATE(timestamp) as date,
				COUNT(*) as commands,
				SUM(saved_tokens) as saved_tokens,
				SUM(original_tokens) as original_size,
				SUM(filtered_tokens) as filtered_size
			FROM commands
			WHERE timestamp >= date('now', '-30 days')
			GROUP BY DATE(timestamp)
			ORDER BY date DESC
		`)
		if err == nil {
			defer rows.Close()
			var daily []TokManSavings
			for rows.Next() {
				var s TokManSavings
				rows.Scan(&s.Date, &s.Commands, &s.SavedTokens, &s.OriginalSize, &s.FilteredSize)
				daily = append(daily, s)
			}
			result["daily"] = daily
		}
	}

	// Weekly savings
	if contains(granularities, "weekly") {
		rows, err := db.Query(`
			SELECT 
				strftime('%Y-W%W', timestamp) as week,
				COUNT(*) as commands,
				SUM(saved_tokens) as saved_tokens,
				SUM(original_tokens) as original_size,
				SUM(filtered_tokens) as filtered_size
			FROM commands
			WHERE timestamp >= date('now', '-90 days')
			GROUP BY strftime('%Y-W%W', timestamp)
			ORDER BY week DESC
		`)
		if err == nil {
			defer rows.Close()
			var weekly []TokManSavings
			for rows.Next() {
				var s TokManSavings
				rows.Scan(&s.Date, &s.Commands, &s.SavedTokens, &s.OriginalSize, &s.FilteredSize)
				weekly = append(weekly, s)
			}
			result["weekly"] = weekly
		}
	}

	// Monthly savings
	if contains(granularities, "monthly") {
		rows, err := db.Query(`
			SELECT 
				strftime('%Y-%m', timestamp) as month,
				COUNT(*) as commands,
				SUM(saved_tokens) as saved_tokens,
				SUM(original_tokens) as original_size,
				SUM(filtered_tokens) as filtered_size
			FROM commands
			WHERE timestamp >= date('now', '-365 days')
			GROUP BY strftime('%Y-%m', timestamp)
			ORDER BY month DESC
		`)
		if err == nil {
			defer rows.Close()
			var monthly []TokManSavings
			for rows.Next() {
				var s TokManSavings
				rows.Scan(&s.Date, &s.Commands, &s.SavedTokens, &s.OriginalSize, &s.FilteredSize)
				monthly = append(monthly, s)
			}
			result["monthly"] = monthly
		}
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateEconomicsReports(ccusageData, tokmanData map[string][]CcUsagePeriod, tokmanSavings map[string][]TokManSavings) []EconomicsReport {
	var reports []EconomicsReport

	// Combine data by period
	for granularity, ccPeriods := range ccusageData {
		tokmanPeriods := tokmanSavings[granularity]

		// Create lookup map for TokMan data
		tokmanMap := make(map[string]TokManSavings)
		for _, t := range tokmanPeriods {
			tokmanMap[t.Date] = t
		}

		for _, cc := range ccPeriods {
			tokman := tokmanMap[cc.Date]

			// Calculate effective cost (what cost would be without TokMan)
			tokmanCompressionRatio := 0.0
			if tokman.OriginalSize > 0 {
				tokmanCompressionRatio = float64(tokman.SavedTokens) / float64(tokman.OriginalSize)
			}

			// Estimate cost savings based on compression
			// Assume input tokens would scale similarly
			estimatedInputWithoutTokMan := float64(cc.InputTokens) / (1 - tokmanCompressionRatio)
			if tokmanCompressionRatio >= 1 || tokmanCompressionRatio <= 0 {
				estimatedInputWithoutTokMan = float64(cc.InputTokens)
			}

			tokensSaved := uint64(estimatedInputWithoutTokMan - float64(cc.InputTokens))
			costSaved := float64(tokensSaved) / float64(cc.TotalTokens) * cc.TotalCost
			if cc.TotalTokens == 0 {
				costSaved = 0
			}

			report := EconomicsReport{
				Period:         cc.Date,
				CcUsage:        cc,
				TokManSavings:  tokman,
				EffectiveCost:  costSaved,
				SavingsPercent: tokmanCompressionRatio * 100,
			}
			reports = append(reports, report)
		}
	}

	// If no ccusage data, just show TokMan savings
	if len(ccusageData) == 0 {
		for granularity, periods := range tokmanSavings {
			_ = granularity
			for _, t := range periods {
				report := EconomicsReport{
					Period:         t.Date,
					TokManSavings:  t,
					SavingsPercent: 0,
				}
				if t.OriginalSize > 0 {
					report.SavingsPercent = float64(t.SavedTokens) / float64(t.OriginalSize) * 100
				}
				reports = append(reports, report)
			}
		}
	}

	return reports
}

func outputText(reports []EconomicsReport, granularities []string) error {
	if len(reports) == 0 {
		fmt.Println("No economics data available yet.")
		fmt.Println("Run some commands through TokMan and use ccusage to track Claude Code spending.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println()
	fmt.Println(cyan("╔══════════════════════════════════════════════════════════╗"))
	fmt.Println(cyan("║           Claude Code Economics Report                   ║"))
	fmt.Println(cyan("╚══════════════════════════════════════════════════════════╝"))
	fmt.Println()

	// Group by granularity
	for _, g := range granularities {
		fmt.Printf("\n📊 %s Breakdown\n", strings.ToUpper(g[:1])+g[1:])
		fmt.Println(strings.Repeat("─", 70))

		var totalCcCost float64
		var totalTokManTokens uint64
		var totalCcTokens uint64

		for _, r := range reports {
			// Simple period matching
			if (g == "daily" && len(r.Period) == 10) ||
				(g == "weekly" && strings.Contains(r.Period, "W")) ||
				(g == "monthly" && len(r.Period) == 7 && strings.Contains(r.Period, "-")) {

				if r.CcUsage.TotalCost > 0 {
					totalCcCost += r.CcUsage.TotalCost
					totalCcTokens += r.CcUsage.TotalTokens
				}
				totalTokManTokens += r.TokManSavings.SavedTokens

				fmt.Printf("\n%s:\n", yellow(r.Period))
				if r.CcUsage.TotalCost > 0 {
					fmt.Printf("  Claude: %s tokens, $%.2f\n",
						formatTokens(r.CcUsage.TotalTokens), r.CcUsage.TotalCost)
				}
				if r.TokManSavings.SavedTokens > 0 {
					fmt.Printf("  TokMan: %s saved (%.1f%% compression)\n",
						formatTokens(r.TokManSavings.SavedTokens), r.SavingsPercent)
				}
			}
		}

		fmt.Println(strings.Repeat("─", 70))
		fmt.Printf("TOTALS:\n")
		if totalCcCost > 0 {
			fmt.Printf("  CC Spend: $%.2f (%s tokens)\n", totalCcCost, formatTokens(totalCcTokens))
		}
		fmt.Printf("  TokMan Savings: %s tokens\n", formatTokens(totalTokManTokens))
		if totalCcCost > 0 && totalTokManTokens > 0 {
			efficiency := float64(totalTokManTokens) / float64(totalCcTokens) * 100
			fmt.Printf("  Efficiency: %s tokens saved per token spent\n",
				green(fmt.Sprintf("%.2fx", efficiency)))
		}
	}

	fmt.Println()
	fmt.Println(green("✓ TokMan reduces your Claude Code token consumption"))
	fmt.Println()

	return nil
}

func outputJson(reports []EconomicsReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(reports)
}

func outputCsv(reports []EconomicsReport) error {
	fmt.Println("period,cc_input_tokens,cc_output_tokens,cc_cache_tokens,cc_total_tokens,cc_cost,tokman_commands,tokman_saved_tokens,tokman_original_size,tokman_filtered_size,effective_cost,savings_percent")

	for _, r := range reports {
		fmt.Printf("%s,%d,%d,%d,%d,%.2f,%d,%d,%d,%d,%.2f,%.2f\n",
			r.Period,
			r.CcUsage.InputTokens,
			r.CcUsage.OutputTokens,
			r.CcUsage.CacheCreationTokens+r.CcUsage.CacheReadTokens,
			r.CcUsage.TotalTokens,
			r.CcUsage.TotalCost,
			r.TokManSavings.Commands,
			r.TokManSavings.SavedTokens,
			r.TokManSavings.OriginalSize,
			r.TokManSavings.FilteredSize,
			r.EffectiveCost,
			r.SavingsPercent,
		)
	}

	return nil
}

func formatTokens(n uint64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
