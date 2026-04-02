package costintel

import (
	"math"
	"sync"
	"time"
)

type SpendForecast struct {
	DailyUSD   float64 `json:"daily_usd"`
	WeeklyUSD  float64 `json:"weekly_usd"`
	MonthlyUSD float64 `json:"monthly_usd"`
	YearlyUSD  float64 `json:"yearly_usd"`
	Trend      string  `json:"trend"`
	Confidence float64 `json:"confidence"`
}

type CostTag struct {
	Name   string  `json:"name"`
	Source string  `json:"source"`
	Total  float64 `json:"total"`
}

type CostAlert struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Message  string  `json:"message"`
	Severity int     `json:"severity"`
	USD      float64 `json:"usd"`
}

type CostIntelligence struct {
	costs     map[string]float64
	tags      map[string]*CostTag
	anomalies []CostAlert
	mu        sync.RWMutex
}

func NewCostIntelligence() *CostIntelligence {
	return &CostIntelligence{
		costs: make(map[string]float64),
		tags:  make(map[string]*CostTag),
	}
}

func (c *CostIntelligence) RecordCost(tag, source string, usd float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.costs[tag] += usd

	if _, ok := c.tags[tag]; !ok {
		c.tags[tag] = &CostTag{Name: tag, Source: source}
	}
	c.tags[tag].Total += usd

	c.detectAnomaly(tag, usd)
}

func (c *CostIntelligence) Forecast(dailyRate float64) *SpendForecast {
	c.mu.RLock()
	defer c.mu.RUnlock()

	trend := "stable"
	if dailyRate > 0 {
		trend = "increasing"
	}

	return &SpendForecast{
		DailyUSD:   dailyRate,
		WeeklyUSD:  dailyRate * 7,
		MonthlyUSD: dailyRate * 30,
		YearlyUSD:  dailyRate * 365,
		Trend:      trend,
		Confidence: 0.85,
	}
}

func (c *CostIntelligence) GetBudgetStatus(budgetUSD, spentUSD float64) (bool, float64, string) {
	remaining := budgetUSD - spentUSD
	pctUsed := spentUSD / budgetUSD * 100

	if pctUsed >= 100 {
		return false, 0, "BUDGET_EXCEEDED"
	}
	if pctUsed >= 90 {
		return true, remaining, "BUDGET_WARNING"
	}
	return true, remaining, "BUDGET_OK"
}

func (c *CostIntelligence) detectAnomaly(tag string, usd float64) {
	if len(c.anomalies) > 100 {
		c.anomalies = c.anomalies[1:]
	}

	mean := c.costs[tag] / float64(len(c.costs))
	stddev := 0.0
	for _, v := range c.costs {
		stddev += (v - mean) * (v - mean)
	}
	stddev = math.Sqrt(stddev / float64(len(c.costs)))

	if usd > mean+2*stddev {
		c.anomalies = append(c.anomalies, CostAlert{
			ID:       "anomaly_" + tag,
			Type:     "cost_anomaly",
			Message:  "Cost anomaly detected for " + tag,
			Severity: 7,
			USD:      usd,
		})
	}
}

func (c *CostIntelligence) GetAnomalies() []CostAlert {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]CostAlert{}, c.anomalies...)
}

func (c *CostIntelligence) GetTags() []*CostTag {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*CostTag
	for _, t := range c.tags {
		result = append(result, t)
	}
	return result
}

func (c *CostIntelligence) ExportCSV() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	output := "tag,source,total_usd\n"
	for _, t := range c.tags {
		output += t.Name + "," + t.Source + ","
		output += formatFloat(t.Total) + "\n"
	}
	return output
}

func formatFloat(f float64) string {
	intPart := int(f)
	decPart := int((f - float64(intPart)) * 100)
	return string(rune(intPart+'0')) + "." + string(rune(decPart/10+'0')) + string(rune(decPart%10+'0'))
}

type WeeklyDigest struct {
	WeekStart       time.Time `json:"week_start"`
	WeekEnd         time.Time `json:"week_end"`
	TotalCost       float64   `json:"total_cost"`
	TotalTokens     int64     `json:"total_tokens"`
	TopModel        string    `json:"top_model"`
	TopCommand      string    `json:"top_command"`
	SavingsPct      float64   `json:"savings_pct"`
	Recommendations []string  `json:"recommendations"`
}

func GenerateWeeklyDigest(costs map[string]float64, tokens int64) *WeeklyDigest {
	totalCost := 0.0
	for _, c := range costs {
		totalCost += c
	}

	return &WeeklyDigest{
		WeekStart:   time.Now().AddDate(0, 0, -7),
		WeekEnd:     time.Now(),
		TotalCost:   totalCost,
		TotalTokens: tokens,
		SavingsPct:  0.0,
		Recommendations: []string{
			"Consider using smaller models for simple tasks",
			"Enable aggressive compression mode",
			"Review usage patterns for optimization",
		},
	}
}
