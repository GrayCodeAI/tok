package costforecast

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewForecaster(t *testing.T) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	if forecaster == nil {
		t.Fatal("expected forecaster to be created")
	}

	if len(forecaster.models) == 0 {
		t.Error("expected models to be initialized")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultHorizon != 30*24*time.Hour {
		t.Errorf("expected 30 day horizon, got %v", config.DefaultHorizon)
	}

	if config.ConfidenceLevel != 0.95 {
		t.Errorf("expected 0.95 confidence, got %.2f", config.ConfidenceLevel)
	}
}

func TestForecast(t *testing.T) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	// Create historical data (need at least 14 points)
	historical := make([]DataPoint, 14)
	for i := 0; i < 14; i++ {
		historical[i] = DataPoint{
			Timestamp: time.Now().Add(time.Duration(i-14) * 24 * time.Hour),
			Value:     float64(100 + i),
		}
	}

	report, err := forecaster.Forecast(historical, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to forecast: %v", err)
	}

	if report == nil {
		t.Fatal("expected report to be generated")
	}

	if len(report.Forecasts) == 0 {
		t.Error("expected forecasts")
	}
}

func TestForecastInsufficientData(t *testing.T) {
	config := DefaultConfig()
	config.MinDataPoints = 14
	forecaster := NewForecaster(config)

	historical := []DataPoint{
		{Timestamp: time.Now(), Value: 100},
	}

	_, err := forecaster.Forecast(historical, 7*24*time.Hour)
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestLinearRegressionModel(t *testing.T) {
	model := &LinearRegressionModel{weight: 0.3}

	historical := []DataPoint{
		{Timestamp: time.Unix(0, 0), Value: 10},
		{Timestamp: time.Unix(1, 0), Value: 20},
		{Timestamp: time.Unix(2, 0), Value: 30},
	}

	prediction, err := model.Predict(historical, 5)
	if err != nil {
		t.Fatalf("failed to predict: %v", err)
	}

	if len(prediction.PointForecast) != 5 {
		t.Errorf("expected 5 forecasts, got %d", len(prediction.PointForecast))
	}
}

func TestMovingAverageModel(t *testing.T) {
	model := &MovingAverageModel{window: 2, weight: 0.2}

	historical := []DataPoint{
		{Timestamp: time.Now(), Value: 100},
		{Timestamp: time.Now(), Value: 110},
		{Timestamp: time.Now(), Value: 120},
	}

	prediction, err := model.Predict(historical, 3)
	if err != nil {
		t.Fatalf("failed to predict: %v", err)
	}

	if len(prediction.PointForecast) != 3 {
		t.Errorf("expected 3 forecasts, got %d", len(prediction.PointForecast))
	}
}

func TestMovingAverageModelInsufficientData(t *testing.T) {
	model := &MovingAverageModel{window: 5, weight: 0.2}

	historical := []DataPoint{
		{Timestamp: time.Now(), Value: 100},
	}

	_, err := model.Predict(historical, 3)
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestExponentialSmoothingModel(t *testing.T) {
	model := &ExponentialSmoothingModel{alpha: 0.3, weight: 0.25}

	historical := []DataPoint{
		{Timestamp: time.Now(), Value: 100},
		{Timestamp: time.Now(), Value: 110},
		{Timestamp: time.Now(), Value: 105},
	}

	prediction, err := model.Predict(historical, 5)
	if err != nil {
		t.Fatalf("failed to predict: %v", err)
	}

	if len(prediction.PointForecast) != 5 {
		t.Errorf("expected 5 forecasts, got %d", len(prediction.PointForecast))
	}
}

func TestSeasonalDecompositionModel(t *testing.T) {
	model := &SeasonalDecompositionModel{period: 7, weight: 0.25}

	// Need at least 2 periods (14 days)
	historical := make([]DataPoint, 14)
	for i := 0; i < 14; i++ {
		historical[i] = DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * 24 * time.Hour),
			Value:     float64(100 + i),
		}
	}

	prediction, err := model.Predict(historical, 7)
	if err != nil {
		t.Fatalf("failed to predict: %v", err)
	}

	if len(prediction.PointForecast) != 7 {
		t.Errorf("expected 7 forecasts, got %d", len(prediction.PointForecast))
	}
}

func TestDetectSeasonality(t *testing.T) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	historical := make([]DataPoint, 21)
	for i := 0; i < 21; i++ {
		historical[i] = DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * 24 * time.Hour),
			Value:     float64(100 + i%7*10), // Weekly pattern
		}
	}

	pattern := forecaster.detectSeasonality(historical)

	if !pattern.HasSeasonality {
		t.Error("expected seasonality to be detected")
	}
}

func TestCalculateRiskLevel(t *testing.T) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	// Create forecast with high variance
	ensemble := Prediction{
		PointForecast: []float64{100, 200, 50, 300, 25},
	}

	historical := []DataPoint{{Timestamp: time.Now(), Value: 100}}

	risk := forecaster.calculateRiskLevel(ensemble, historical)

	if risk != RiskCritical && risk != RiskHigh {
		t.Errorf("expected high risk, got %s", risk)
	}
}

func TestForecastReportExport(t *testing.T) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	historical := make([]DataPoint, 14)
	for i := 0; i < 14; i++ {
		historical[i] = DataPoint{
			Timestamp: time.Now().Add(time.Duration(i-14) * 24 * time.Hour),
			Value:     float64(100 + i),
		}
	}

	report, err := forecaster.Forecast(historical, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to forecast: %v", err)
	}

	// Export to JSON
	data, err := json.Marshal(report)
	if err != nil {
		t.Errorf("failed to export: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkForecast(b *testing.B) {
	config := DefaultConfig()
	forecaster := NewForecaster(config)

	historical := make([]DataPoint, 30)
	for i := 0; i < 30; i++ {
		historical[i] = DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * 24 * time.Hour),
			Value:     float64(100 + i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		forecaster.Forecast(historical, 7*24*time.Hour)
	}
}
