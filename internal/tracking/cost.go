// Package tracking provides cost estimation and reporting.
package tracking

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// CostEstimator estimates API costs based on token usage.
type CostEstimator struct {
	inputCostPer1M  float64 // Cost per 1M input tokens
	outputCostPer1M float64 // Cost per 1M output tokens (if applicable)
	currency        string
}

// ModelPricing contains pricing for different LLM models.
var ModelPricing = map[string]CostEstimator{
	"claude-3-opus":     {inputCostPer1M: 15.00, outputCostPer1M: 75.00, currency: "USD"},
	"claude-3-sonnet":   {inputCostPer1M: 3.00, outputCostPer1M: 15.00, currency: "USD"},
	"claude-3-haiku":    {inputCostPer1M: 0.25, outputCostPer1M: 1.25, currency: "USD"},
	"claude-3.5-sonnet": {inputCostPer1M: 3.00, outputCostPer1M: 15.00, currency: "USD"},
	"claude-3.5-haiku":  {inputCostPer1M: 0.80, outputCostPer1M: 4.00, currency: "USD"},
	"gpt-4-turbo":       {inputCostPer1M: 10.00, outputCostPer1M: 30.00, currency: "USD"},
	"gpt-4":             {inputCostPer1M: 30.00, outputCostPer1M: 60.00, currency: "USD"},
	"gpt-3.5-turbo":     {inputCostPer1M: 0.50, outputCostPer1M: 1.50, currency: "USD"},
	"default":           {inputCostPer1M: 3.00, outputCostPer1M: 15.00, currency: "USD"},
}

// NewCostEstimator creates a new cost estimator.
func NewCostEstimator(model string) *CostEstimator {
	if pricing, ok := ModelPricing[model]; ok {
		return &CostEstimator{pricing.inputCostPer1M, pricing.outputCostPer1M, pricing.currency}
	}
	p := ModelPricing["default"]
	return &CostEstimator{p.inputCostPer1M, p.outputCostPer1M, p.currency}
}

// EstimateSavings estimates cost savings from token reduction.
func (ce *CostEstimator) EstimateSavings(tokensSaved int) CostEstimate {
	savings := float64(tokensSaved) * ce.inputCostPer1M / 1_000_000
	return CostEstimate{
		TokensSaved:    tokensSaved,
		EstimatedSavings: savings,
		Currency:       ce.currency,
		CostPer1MTokens: ce.inputCostPer1M,
	}
}

// EstimateCost estimates the cost for a number of tokens.
func (ce *CostEstimator) EstimateCost(tokens int) float64 {
	return float64(tokens) * ce.inputCostPer1M / 1_000_000
}

// CostEstimate contains cost estimation results.
type CostEstimate struct {
	TokensSaved     int
	EstimatedSavings float64
	Currency        string
	CostPer1MTokens float64
}

// Format returns a formatted string representation.
func (ce CostEstimate) Format() string {
	return fmt.Sprintf("$%.2f %s (saved %d tokens @ $%.2f/1M)",
		ce.EstimatedSavings, ce.Currency, ce.TokensSaved, ce.CostPer1MTokens)
}

// CostReport contains comprehensive cost reporting.
type CostReport struct {
	GeneratedAt      time.Time
	Period           string
	TotalTokensSaved int64
	EstimatedSavings float64
	Currency         string
	Model            string
	DailyBreakdown   []DailyCost
	TopCommands      []CommandCost
	Projections      CostProjection
}

// DailyCost represents cost data for a single day.
type DailyCost struct {
	Date        string
	TokensSaved int
	CostSaved   float64
	CommandCount int
}

// CommandCost represents cost data for a command.
type CommandCost struct {
	Command     string
	TimesRun    int
	TokensSaved int
	CostSaved   float64
}

// CostProjection contains future cost projections.
type CostProjection struct {
	MonthlyEstimate float64
	YearlyEstimate  float64
	GrowthRate      float64 // Monthly growth rate
}

// GenerateCostReport generates a comprehensive cost report.
func (t *Tracker) GenerateCostReport(model string, days int) (*CostReport, error) {
	estimator := NewCostEstimator(model)

	// Get daily savings
	dailySavings, err := t.GetDailySavings("", days)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily savings: %w", err)
	}

	var totalSaved int64
	dailyBreakdown := make([]DailyCost, 0, len(dailySavings))

	for _, ds := range dailySavings {
		totalSaved += int64(ds.Saved)
		dailyBreakdown = append(dailyBreakdown, DailyCost{
			Date:         ds.Date,
			TokensSaved:  ds.Saved,
			CostSaved:    estimator.EstimateCost(ds.Saved),
			CommandCount: ds.Commands,
		})
	}

	// Calculate projections based on average daily savings
	avgDaily := float64(totalSaved) / float64(len(dailySavings))
	monthlyEstimate := estimator.EstimateCost(int(avgDaily * 30))
	yearlyEstimate := monthlyEstimate * 12

	report := &CostReport{
		GeneratedAt:      time.Now(),
		Period:           fmt.Sprintf("last %d days", days),
		TotalTokensSaved: totalSaved,
		EstimatedSavings: estimator.EstimateCost(int(totalSaved)),
		Currency:         estimator.currency,
		Model:            model,
		DailyBreakdown:   dailyBreakdown,
		Projections: CostProjection{
			MonthlyEstimate: monthlyEstimate,
			YearlyEstimate:  yearlyEstimate,
			GrowthRate:      0.0,
		},
	}

	return report, nil
}

// ExportToCSV exports the cost report to CSV.
func (cr *CostReport) ExportToCSV(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Date", "Tokens Saved", "Cost Saved (" + cr.Currency + ")", "Commands"})

	// Write data
	for _, day := range cr.DailyBreakdown {
		writer.Write([]string{
			day.Date,
			fmt.Sprintf("%d", day.TokensSaved),
			fmt.Sprintf("%.4f", day.CostSaved),
			fmt.Sprintf("%d", day.CommandCount),
		})
	}

	// Write summary
	writer.Write([]string{"", "", "", ""})
	writer.Write([]string{"SUMMARY", "", "", ""})
	writer.Write([]string{"Total Tokens Saved", fmt.Sprintf("%d", cr.TotalTokensSaved), "", ""})
	writer.Write([]string{"Total Cost Saved", fmt.Sprintf("%.4f", cr.EstimatedSavings), cr.Currency, ""})
	writer.Write([]string{"Monthly Projection", fmt.Sprintf("%.4f", cr.Projections.MonthlyEstimate), cr.Currency, ""})
	writer.Write([]string{"Yearly Projection", fmt.Sprintf("%.4f", cr.Projections.YearlyEstimate), cr.Currency, ""})

	return nil
}

// ExportToJSON exports the cost report to JSON.
func (cr *CostReport) ExportToJSON(path string) error {
	data, err := json.MarshalIndent(cr, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Format returns a formatted report string.
func (cr *CostReport) Format() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Cost Report (%s)\n", cr.Period)
	fmt.Fprintf(&b, "Generated: %s\n\n", cr.GeneratedAt.Format(time.RFC3339))

	fmt.Fprintf(&b, "Model: %s\n", cr.Model)
	fmt.Fprintf(&b, "Total Tokens Saved: %d\n", cr.TotalTokensSaved)
	fmt.Fprintf(&b, "Estimated Savings: $%.2f %s\n\n", cr.EstimatedSavings, cr.Currency)

	fmt.Fprintf(&b, "Projections:\n")
	fmt.Fprintf(&b, "  Monthly: $%.2f\n", cr.Projections.MonthlyEstimate)
	fmt.Fprintf(&b, "  Yearly: $%.2f\n\n", cr.Projections.YearlyEstimate)

	if len(cr.DailyBreakdown) > 0 {
		fmt.Fprintf(&b, "Daily Breakdown (last 7 days):\n")
		start := len(cr.DailyBreakdown) - 7
		if start < 0 {
			start = 0
		}
		for _, day := range cr.DailyBreakdown[start:] {
			fmt.Fprintf(&b, "  %s: %d tokens - $%.2f\n",
				day.Date, day.TokensSaved, day.CostSaved)
		}
	}

	return b.String()
}

// AlertThreshold represents alert configuration.
type AlertThreshold struct {
	DailyTokenLimit  int64
	DailyCostLimit   float64
	WeeklyTokenLimit int64
	WeeklyCostLimit  float64
}

// CheckAlert checks if any thresholds are exceeded.
func (t *Tracker) CheckAlert(threshold AlertThreshold) ([]Alert, error) {
	var alerts []Alert

	// Check daily limits
	today := time.Now().Format("2006-01-02")
	todayRecord, err := t.GetDailySavings(today, 1)
	if err == nil && len(todayRecord) > 0 {
		tokens := int64(todayRecord[0].Saved)
		if tokens > threshold.DailyTokenLimit {
			alerts = append(alerts, Alert{
				Type:      "daily_tokens",
				Severity:  "warning",
				Message:   fmt.Sprintf("Daily token limit exceeded: %d > %d", tokens, threshold.DailyTokenLimit),
				Timestamp: time.Now(),
			})
		}
	}

	// Check weekly limits (approximate from daily)
	dailySaved, _ := t.TokensSaved24h()
	weeklySaved := dailySaved * 7
	if weeklySaved > threshold.WeeklyTokenLimit {
		alerts = append(alerts, Alert{
			Type:      "weekly_tokens",
			Severity:  "warning",
			Message:   fmt.Sprintf("Weekly token limit exceeded: %d > %d", weeklySaved, threshold.WeeklyTokenLimit),
			Timestamp: time.Now(),
		})
	}

	return alerts, nil
}

// Alert represents a threshold alert.
type Alert struct {
	Type      string
	Severity  string
	Message   string
	Timestamp time.Time
}

// ExportReport exports a report in the specified format.
func (t *Tracker) ExportReport(format, outputPath string, days int) error {
	report, err := t.GenerateCostReport("default", days)
	if err != nil {
		return err
	}

	switch format {
	case "csv":
		return report.ExportToCSV(outputPath)
	case "json":
		return report.ExportToJSON(outputPath)
	case "text":
		return os.WriteFile(outputPath, []byte(report.Format()), 0644)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
