// Package dashboard provides chart data endpoints for cost projections and visualization.
package dashboard

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/GrayCodeAI/tokman/internal/httpmw"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// CostProjection represents a projected cost over time
type CostProjection struct {
	Date        string  `json:"date"`
	TokensSaved int64   `json:"tokens_saved"`
	CostSaved   float64 `json:"cost_saved"`
	Cumulative  float64 `json:"cumulative"`
}

// ContributionCell represents a single cell in the contribution graph
type ContributionCell struct {
	X         int     `json:"x"` // Day of week (0-6)
	Y         int     `json:"y"` // Week number
	Value     int     `json:"value"`
	Intensity float64 `json:"intensity"`
	Date      string  `json:"date,omitempty"`
}

// Contribution3DData represents 3D contribution graph data
type Contribution3DData struct {
	Cells      []ContributionCell3D `json:"cells"`
	Weeks      int                  `json:"weeks"`
	MaxValue   int                  `json:"max_value"`
	TotalValue int                  `json:"total_value"`
	DateRange  DateRange            `json:"date_range"`
}

// ContributionCell3D represents a 3D cell with height
type ContributionCell3D struct {
	X         int     `json:"x"`         // Day of week (0-6)
	Y         int     `json:"y"`         // Week number
	Z         float64 `json:"z"`         // Height based on value
	Value     int     `json:"value"`     // Raw value
	Intensity float64 `json:"intensity"` // 0.0 - 1.0
	Date      string  `json:"date"`
	Weekday   string  `json:"weekday"`
}

// DateRange represents a date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ProjectionRequest contains parameters for cost projections
type ProjectionRequest struct {
	Days       int     `json:"days"`
	Model      string  `json:"model"`
	GrowthRate float64 `json:"growth_rate"` // Monthly growth rate (e.g., 0.1 = 10%)
}

// ProjectionResponse contains all projection data
type ProjectionResponse struct {
	Projections []CostProjection  `json:"projections"`
	Summary     ProjectionSummary `json:"summary"`
	Model       string            `json:"model"`
	CostPer1M   float64           `json:"cost_per_1m"`
}

// ProjectionSummary contains summary statistics
type ProjectionSummary struct {
	TotalTokensSaved int64   `json:"total_tokens_saved"`
	TotalCostSaved   float64 `json:"total_cost_saved"`
	DailyAverage     float64 `json:"daily_average"`
	ProjectedMonthly float64 `json:"projected_monthly"`
	ProjectedYearly  float64 `json:"projected_yearly"`
}

func costProjectionHandler(tracker *tracking.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpmw.JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
			return
		}

		var req ProjectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Default to 30 days with default model
			req.Days = 30
			req.Model = "claude-3.5-sonnet"
			req.GrowthRate = 0.0
		}

		if req.Days <= 0 {
			req.Days = 30
		}
		if req.Days > 365 {
			req.Days = 365
		}

		// Get historical data
		history, err := tracker.GetDailySavings("", req.Days)
		if err != nil {
			httpmw.JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get historical data",
			})
			return
		}

		// Get pricing for model
		costPer1M := 3.0 // Default
		if pricing, ok := modelPricing[req.Model]; ok {
			costPer1M = pricing.input
		} else {
			// Check if it's a known model pattern
			for model, p := range modelPricing {
				if contains(req.Model, model) {
					costPer1M = p.input
					break
				}
			}
		}

		// Calculate historical average
		var totalSaved int64
		for _, h := range history {
			totalSaved += int64(h.Saved)
		}
		dailyAvg := 0.0
		if len(history) > 0 {
			dailyAvg = float64(totalSaved) / float64(len(history))
		}

		// Generate projections
		now := time.Now()
		projections := make([]CostProjection, req.Days)
		var cumulative float64

		for i := 0; i < req.Days; i++ {
			date := now.AddDate(0, 0, i)

			// Apply growth rate
			growthFactor := 1.0
			if req.GrowthRate > 0 {
				months := float64(i) / 30.0
				growthFactor = math.Pow(1.0+req.GrowthRate, months)
			}

			projectedTokens := int64(dailyAvg * growthFactor)
			costSaved := float64(projectedTokens) * costPer1M / 1_000_000
			cumulative += costSaved

			projections[i] = CostProjection{
				Date:        date.Format("2006-01-02"),
				TokensSaved: projectedTokens,
				CostSaved:   costSaved,
				Cumulative:  cumulative,
			}
		}

		// Calculate summary
		monthlyProjection := dailyAvg * 30 * costPer1M / 1_000_000
		yearlyProjection := monthlyProjection * 12

		summary := ProjectionSummary{
			TotalTokensSaved: int64(dailyAvg * float64(req.Days)),
			TotalCostSaved:   cumulative,
			DailyAverage:     dailyAvg,
			ProjectedMonthly: monthlyProjection,
			ProjectedYearly:  yearlyProjection,
		}

		response := ProjectionResponse{
			Projections: projections,
			Summary:     summary,
			Model:       req.Model,
			CostPer1M:   costPer1M,
		}

		httpmw.JSONResponse(w, http.StatusOK, response)
	}
}

func contributionGraphHandler(tracker *tracking.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weeks := 52 // Default to 1 year
		if w := r.URL.Query().Get("weeks"); w != "" {
			if n, err := parseInt(w); err == nil && n > 0 && n <= 104 {
				weeks = n
			}
		}

		// Get command history for the period
		days := weeks * 7
		records, err := tracker.GetDailySavings("", days)
		if err != nil {
			httpmw.JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get command history",
			})
			return
		}

		// Build a map of date -> command count
		dateMap := make(map[string]int)
		for _, r := range records {
			date := r.Date
			// Count based on savings activity (proxy for commands)
			commands := r.Commands
			if commands == 0 {
				commands = r.Saved / 1000 // Estimate: 1000 tokens per command
				if commands == 0 {
					commands = 1
				}
			}
			dateMap[date] += commands
		}

		// Generate contribution cells
		now := time.Now()
		cells := make([]ContributionCell, 0, weeks*7)
		maxValue := 0
		totalValue := 0

		for week := 0; week < weeks; week++ {
			for day := 0; day < 7; day++ {
				daysAgo := (weeks-week-1)*7 + (6 - day)
				date := now.AddDate(0, 0, -daysAgo)
				dateStr := date.Format("2006-01-02")

				value := dateMap[dateStr]
				totalValue += value
				if value > maxValue {
					maxValue = value
				}

				cells = append(cells, ContributionCell{
					X:     day,
					Y:     week,
					Value: value,
					Date:  dateStr,
				})
			}
		}

		// Calculate intensities
		for i := range cells {
			if maxValue > 0 {
				cells[i].Intensity = float64(cells[i].Value) / float64(maxValue)
			}
		}

		startDate := now.AddDate(0, 0, -weeks*7).Format("2006-01-02")
		endDate := now.Format("2006-01-02")

		httpmw.JSONResponse(w, http.StatusOK, map[string]any{
			"cells":       cells,
			"weeks":       weeks,
			"max_value":   maxValue,
			"total_value": totalValue,
			"date_range": DateRange{
				Start: startDate,
				End:   endDate,
			},
		})
	}
}

func contributionGraph3DHandler(tracker *tracking.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weeks := 26 // Default to 6 months
		if w := r.URL.Query().Get("weeks"); w != "" {
			if n, err := parseInt(w); err == nil && n > 0 && n <= 52 {
				weeks = n
			}
		}

		days := weeks * 7
		records, err := tracker.GetDailySavings("", days)
		if err != nil {
			httpmw.JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get command history",
			})
			return
		}

		// Build date map
		dateMap := make(map[string]int)
		for _, rec := range records {
			commands := rec.Commands
			if commands == 0 {
				commands = rec.Saved / 1000
				if commands == 0 {
					commands = 1
				}
			}
			dateMap[rec.Date] += commands
		}

		weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		now := time.Now()
		cells := make([]ContributionCell3D, 0, weeks*7)
		maxValue := 0
		totalValue := 0

		for week := 0; week < weeks; week++ {
			for day := 0; day < 7; day++ {
				daysAgo := (weeks-week-1)*7 + (6 - day)
				date := now.AddDate(0, 0, -daysAgo)
				dateStr := date.Format("2006-01-02")

				value := dateMap[dateStr]
				totalValue += value
				if value > maxValue {
					maxValue = value
				}

				// Calculate Z height (normalized 0-10)
				var z float64
				if maxValue > 0 {
					z = float64(value) / float64(maxValue) * 10
				}

				// Calculate intensity
				intensity := 0.0
				if maxValue > 0 {
					intensity = float64(value) / float64(maxValue)
				}

				cells = append(cells, ContributionCell3D{
					X:         day,
					Y:         week,
					Z:         z,
					Value:     value,
					Intensity: intensity,
					Date:      dateStr,
					Weekday:   weekdays[day],
				})
			}
		}

		startDate := now.AddDate(0, 0, -weeks*7).Format("2006-01-02")
		endDate := now.Format("2006-01-02")

		response := Contribution3DData{
			Cells:      cells,
			Weeks:      weeks,
			MaxValue:   maxValue,
			TotalValue: totalValue,
			DateRange: DateRange{
				Start: startDate,
				End:   endDate,
			},
		}

		httpmw.JSONResponse(w, http.StatusOK, response)
	}
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
