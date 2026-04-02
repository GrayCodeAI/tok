// Package costforecast provides cost forecasting capabilities for TokMan
package costforecast

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Forecaster predicts future costs based on historical data
type Forecaster struct {
	models []Model
	config Config
}

// Config holds forecaster configuration
type Config struct {
	DefaultHorizon    time.Duration
	ConfidenceLevel   float64
	SeasonalityWindow int
	MinDataPoints     int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		DefaultHorizon:    30 * 24 * time.Hour, // 30 days
		ConfidenceLevel:   0.95,
		SeasonalityWindow: 7,
		MinDataPoints:     14,
	}
}

// Model defines a forecasting model interface
type Model interface {
	Name() string
	Predict(historical []DataPoint, horizon int) (Prediction, error)
	Weight() float64
}

// DataPoint represents a historical data point
type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Metadata  map[string]string
}

// Prediction holds forecast results
type Prediction struct {
	Model         string
	Horizon       int
	PointForecast []float64
	LowerBound    []float64
	UpperBound    []float64
	Confidence    float64
	Accuracy      float64
	Trend         TrendDirection
}

// TrendDirection indicates the forecast trend
type TrendDirection string

const (
	TrendUp       TrendDirection = "up"
	TrendDown     TrendDirection = "down"
	TrendStable   TrendDirection = "stable"
	TrendVolatile TrendDirection = "volatile"
)

// ForecastReport contains comprehensive forecast results
type ForecastReport struct {
	GeneratedAt      time.Time
	Horizon          time.Duration
	Forecasts        map[string]Prediction
	EnsembleForecast Prediction
	Aggregated       AggregatedForecast
	Recommendations  []Recommendation
}

// AggregatedForecast combines all forecasts
type AggregatedForecast struct {
	DailyAverage     float64
	MonthlyTotal     float64
	YearlyProjection float64
	GrowthRate       float64
	Seasonality      SeasonalityPattern
	RiskLevel        RiskLevel
}

// SeasonalityPattern captures seasonal trends
type SeasonalityPattern struct {
	HasSeasonality bool
	Period         time.Duration
	PeakDays       []time.Weekday
	LowDays        []time.Weekday
	Strength       float64
}

// RiskLevel indicates forecast risk
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// Recommendation provides actionable insights
type Recommendation struct {
	Type        string
	Priority    int
	Title       string
	Description string
	Impact      float64
	Action      string
}

// NewForecaster creates a cost forecaster
func NewForecaster(config Config) *Forecaster {
	return &Forecaster{
		config: config,
		models: []Model{
			&LinearRegressionModel{weight: 0.3},
			&MovingAverageModel{window: 7, weight: 0.2},
			&ExponentialSmoothingModel{alpha: 0.3, weight: 0.25},
			&SeasonalDecompositionModel{period: 7, weight: 0.25},
		},
	}
}

// Forecast generates a cost forecast
func (f *Forecaster) Forecast(historical []DataPoint, horizon time.Duration) (*ForecastReport, error) {
	if len(historical) < f.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points: need at least %d, got %d", f.config.MinDataPoints, len(historical))
	}

	// Sort historical data by timestamp
	sort.Slice(historical, func(i, j int) bool {
		return historical[i].Timestamp.Before(historical[j].Timestamp)
	})

	days := int(horizon.Hours() / 24)
	if days == 0 {
		days = 30
	}

	report := &ForecastReport{
		GeneratedAt: time.Now(),
		Horizon:     horizon,
		Forecasts:   make(map[string]Prediction),
	}

	// Generate predictions from each model
	predictions := make([]Prediction, 0, len(f.models))
	for _, model := range f.models {
		pred, err := model.Predict(historical, days)
		if err != nil {
			continue
		}
		report.Forecasts[model.Name()] = pred
		predictions = append(predictions, pred)
	}

	// Create ensemble forecast
	report.EnsembleForecast = f.createEnsemble(predictions)

	// Calculate aggregated metrics
	report.Aggregated = f.calculateAggregated(report.EnsembleForecast, historical)

	// Generate recommendations
	report.Recommendations = f.generateRecommendations(report, historical)

	return report, nil
}

func (f *Forecaster) createEnsemble(predictions []Prediction) Prediction {
	if len(predictions) == 0 {
		return Prediction{}
	}

	horizon := len(predictions[0].PointForecast)
	ensemble := Prediction{
		Model:         "ensemble",
		Horizon:       horizon,
		PointForecast: make([]float64, horizon),
		LowerBound:    make([]float64, horizon),
		UpperBound:    make([]float64, horizon),
		Confidence:    f.config.ConfidenceLevel,
	}

	// Weighted average of all predictions
	for i := 0; i < horizon; i++ {
		var weightedSum, totalWeight float64

		for _, pred := range predictions {
			if i < len(pred.PointForecast) {
				// Extract model weight from model name - simplified
				weight := 0.25 // default equal weighting
				weightedSum += pred.PointForecast[i] * weight
				totalWeight += weight
			}
		}

		if totalWeight > 0 {
			ensemble.PointForecast[i] = weightedSum / totalWeight
		}

		// Calculate confidence intervals
		ensemble.LowerBound[i] = ensemble.PointForecast[i] * 0.9
		ensemble.UpperBound[i] = ensemble.PointForecast[i] * 1.1
	}

	// Determine trend
	if len(ensemble.PointForecast) >= 2 {
		first := ensemble.PointForecast[0]
		last := ensemble.PointForecast[len(ensemble.PointForecast)-1]
		change := (last - first) / first

		switch {
		case change > 0.2:
			ensemble.Trend = TrendUp
		case change < -0.2:
			ensemble.Trend = TrendDown
		case math.Abs(change) <= 0.1:
			ensemble.Trend = TrendStable
		default:
			ensemble.Trend = TrendVolatile
		}
	}

	return ensemble
}

func (f *Forecaster) calculateAggregated(ensemble Prediction, historical []DataPoint) AggregatedForecast {
	agg := AggregatedForecast{}

	// Calculate daily average from forecast
	if len(ensemble.PointForecast) > 0 {
		sum := 0.0
		for _, v := range ensemble.PointForecast {
			sum += v
		}
		agg.DailyAverage = sum / float64(len(ensemble.PointForecast))
	}

	agg.MonthlyTotal = agg.DailyAverage * 30
	agg.YearlyProjection = agg.DailyAverage * 365

	// Calculate growth rate
	if len(historical) >= 2 {
		first := historical[0].Value
		last := historical[len(historical)-1].Value
		days := historical[len(historical)-1].Timestamp.Sub(historical[0].Timestamp).Hours() / 24
		if days > 0 {
			agg.GrowthRate = (math.Pow(last/first, 1/days) - 1) * 100
		}
	}

	// Detect seasonality
	agg.Seasonality = f.detectSeasonality(historical)

	// Determine risk level
	agg.RiskLevel = f.calculateRiskLevel(ensemble, historical)

	return agg
}

func (f *Forecaster) detectSeasonality(historical []DataPoint) SeasonalityPattern {
	if len(historical) < 14 {
		return SeasonalityPattern{HasSeasonality: false}
	}

	// Group by day of week
	dayTotals := make(map[time.Weekday]float64)
	dayCounts := make(map[time.Weekday]int)

	for _, dp := range historical {
		day := dp.Timestamp.Weekday()
		dayTotals[day] += dp.Value
		dayCounts[day]++
	}

	// Calculate averages
	dayAverages := make(map[time.Weekday]float64)
	for day, total := range dayTotals {
		if dayCounts[day] > 0 {
			dayAverages[day] = total / float64(dayCounts[day])
		}
	}

	// Find peaks and lows
	var peakDays, lowDays []time.Weekday
	var maxAvg, minAvg float64

	for _, avg := range dayAverages {
		if avg > maxAvg {
			maxAvg = avg
		}
		if minAvg == 0 || avg < minAvg {
			minAvg = avg
		}
	}

	threshold := (maxAvg + minAvg) / 2
	for day, avg := range dayAverages {
		if avg > threshold*1.1 {
			peakDays = append(peakDays, day)
		} else if avg < threshold*0.9 {
			lowDays = append(lowDays, day)
		}
	}

	// Calculate seasonality strength
	strength := 0.0
	if maxAvg > 0 {
		strength = (maxAvg - minAvg) / maxAvg
	}

	return SeasonalityPattern{
		HasSeasonality: strength > 0.1,
		Period:         7 * 24 * time.Hour,
		PeakDays:       peakDays,
		LowDays:        lowDays,
		Strength:       strength,
	}
}

func (f *Forecaster) calculateRiskLevel(ensemble Prediction, historical []DataPoint) RiskLevel {
	if len(ensemble.PointForecast) == 0 || len(historical) == 0 {
		return RiskMedium
	}

	// Calculate coefficient of variation
	mean := 0.0
	for _, v := range ensemble.PointForecast {
		mean += v
	}
	mean /= float64(len(ensemble.PointForecast))

	variance := 0.0
	for _, v := range ensemble.PointForecast {
		variance += math.Pow(v-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(ensemble.PointForecast)))

	cv := stdDev / mean

	switch {
	case cv > 0.5:
		return RiskCritical
	case cv > 0.3:
		return RiskHigh
	case cv > 0.15:
		return RiskMedium
	default:
		return RiskLow
	}
}

func (f *Forecaster) generateRecommendations(report *ForecastReport, historical []DataPoint) []Recommendation {
	recommendations := make([]Recommendation, 0)

	// Budget recommendation
	if report.Aggregated.GrowthRate > 10 {
		recommendations = append(recommendations, Recommendation{
			Type:        "budget",
			Priority:    1,
			Title:       "Increase Budget Allocation",
			Description: fmt.Sprintf("Projected %.1f%% cost growth. Consider increasing budget.", report.Aggregated.GrowthRate),
			Impact:      report.Aggregated.GrowthRate,
			Action:      "Review and increase monthly budget allocation",
		})
	}

	// Seasonality recommendation
	if report.Aggregated.Seasonality.HasSeasonality {
		recommendations = append(recommendations, Recommendation{
			Type:        "optimization",
			Priority:    2,
			Title:       "Leverage Seasonal Patterns",
			Description: "Detected weekly seasonality. Schedule heavy processing on low-cost days.",
			Impact:      report.Aggregated.Seasonality.Strength * 100,
			Action:      fmt.Sprintf("Schedule batch jobs on %v", report.Aggregated.Seasonality.LowDays),
		})
	}

	// Risk recommendation
	if report.Aggregated.RiskLevel == RiskHigh || report.Aggregated.RiskLevel == RiskCritical {
		recommendations = append(recommendations, Recommendation{
			Type:        "risk",
			Priority:    1,
			Title:       "High Cost Volatility Detected",
			Description: "Forecast shows high uncertainty. Implement cost controls.",
			Impact:      100,
			Action:      "Set up automated budget alerts and spending limits",
		})
	}

	return recommendations
}

// Model implementations

// LinearRegressionModel uses simple linear regression
type LinearRegressionModel struct {
	weight float64
}

func (m *LinearRegressionModel) Name() string    { return "linear_regression" }
func (m *LinearRegressionModel) Weight() float64 { return m.weight }

func (m *LinearRegressionModel) Predict(historical []DataPoint, horizon int) (Prediction, error) {
	n := float64(len(historical))
	if n < 2 {
		return Prediction{}, fmt.Errorf("need at least 2 data points")
	}

	// Calculate means
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, dp := range historical {
		x := float64(i)
		y := dp.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	meanX := sumX / n
	meanY := sumY / n

	// Calculate slope and intercept
	slope := (sumXY - n*meanX*meanY) / (sumX2 - n*meanX*meanX)
	intercept := meanY - slope*meanX

	// Generate predictions
	forecast := make([]float64, horizon)
	for i := 0; i < horizon; i++ {
		x := float64(len(historical) + i)
		forecast[i] = intercept + slope*x
	}

	return Prediction{
		Model:         m.Name(),
		Horizon:       horizon,
		PointForecast: forecast,
		Confidence:    0.95,
	}, nil
}

// MovingAverageModel uses moving average
type MovingAverageModel struct {
	window int
	weight float64
}

func (m *MovingAverageModel) Name() string    { return "moving_average" }
func (m *MovingAverageModel) Weight() float64 { return m.weight }

func (m *MovingAverageModel) Predict(historical []DataPoint, horizon int) (Prediction, error) {
	if len(historical) < m.window {
		return Prediction{}, fmt.Errorf("insufficient data for window size %d", m.window)
	}

	// Calculate average of last window
	sum := 0.0
	for i := len(historical) - m.window; i < len(historical); i++ {
		sum += historical[i].Value
	}
	avg := sum / float64(m.window)

	// Generate flat forecast
	forecast := make([]float64, horizon)
	for i := range forecast {
		forecast[i] = avg
	}

	return Prediction{
		Model:         m.Name(),
		Horizon:       horizon,
		PointForecast: forecast,
	}, nil
}

// ExponentialSmoothingModel uses exponential smoothing
type ExponentialSmoothingModel struct {
	alpha  float64
	weight float64
}

func (m *ExponentialSmoothingModel) Name() string    { return "exponential_smoothing" }
func (m *ExponentialSmoothingModel) Weight() float64 { return m.weight }

func (m *ExponentialSmoothingModel) Predict(historical []DataPoint, horizon int) (Prediction, error) {
	if len(historical) == 0 {
		return Prediction{}, fmt.Errorf("no historical data")
	}

	// Initialize with first value
	forecast := historical[0].Value

	// Apply exponential smoothing
	for i := 1; i < len(historical); i++ {
		forecast = m.alpha*historical[i].Value + (1-m.alpha)*forecast
	}

	// Generate predictions
	pred := make([]float64, horizon)
	for i := range pred {
		pred[i] = forecast
	}

	return Prediction{
		Model:         m.Name(),
		Horizon:       horizon,
		PointForecast: pred,
	}, nil
}

// SeasonalDecompositionModel detects and uses seasonality
type SeasonalDecompositionModel struct {
	period int
	weight float64
}

func (m *SeasonalDecompositionModel) Name() string    { return "seasonal_decomposition" }
func (m *SeasonalDecompositionModel) Weight() float64 { return m.weight }

func (m *SeasonalDecompositionModel) Predict(historical []DataPoint, horizon int) (Prediction, error) {
	if len(historical) < m.period*2 {
		return Prediction{}, fmt.Errorf("need at least %d data points for seasonality", m.period*2)
	}

	// Calculate seasonal indices
	indices := make([]float64, m.period)
	for i := 0; i < m.period; i++ {
		sum := 0.0
		count := 0
		for j := i; j < len(historical); j += m.period {
			sum += historical[j].Value
			count++
		}
		if count > 0 {
			indices[i] = sum / float64(count)
		}
	}

	// Calculate trend
	trendSum := 0.0
	for i := len(historical) - m.period; i < len(historical); i++ {
		trendSum += historical[i].Value
	}
	trend := trendSum / float64(m.period)

	// Generate predictions with seasonality
	forecast := make([]float64, horizon)
	for i := 0; i < horizon; i++ {
		seasonalIdx := (len(historical) + i) % m.period
		forecast[i] = trend * (indices[seasonalIdx] / trend)
	}

	return Prediction{
		Model:         m.Name(),
		Horizon:       horizon,
		PointForecast: forecast,
	}, nil
}
