package forecast

import (
	"context"
	"math"
	"sync"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

type ForecastModel struct {
	mu          sync.RWMutex
	history     []DataPoint
	period      time.Duration
	alpha       float64
	beta        float64
	level       float64
	trend       float64
	initialized bool
}

func NewForecastModel(period time.Duration) *ForecastModel {
	return &ForecastModel{
		history: make([]DataPoint, 0),
		period:  period,
		alpha:   0.3,
		beta:    0.1,
	}
}

func (m *ForecastModel) AddPoint(ctx context.Context, point DataPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, point)
	if len(m.history) > 1000 {
		m.history = m.history[len(m.history)-1000:]
	}

	if !m.initialized && len(m.history) >= 2 {
		m.level = m.history[0].Value
		m.trend = m.history[1].Value - m.history[0].Value
		m.initialized = true
	}

	if m.initialized {
		level := m.alpha*point.Value + (1-m.alpha)*(m.level+m.trend)
		trend := m.beta*(level-m.level) + (1-m.beta)*m.trend
		m.level = level
		m.trend = trend
	}
}

func (m *ForecastModel) Predict(steps int) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		if len(m.history) == 0 {
			return 0
		}
		return m.history[len(m.history)-1].Value
	}

	return m.level + float64(steps)*m.trend
}

func (m *ForecastModel) GetConfidence() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) < 10 {
		return 0
	}

	var sum float64
	for _, p := range m.history {
		sum += math.Abs(p.Value - m.level)
	}
	avgDev := sum / float64(len(m.history))

	if avgDev == 0 {
		return 1.0
	}

	confidence := 1.0 - (avgDev / (math.Abs(m.level) + 1))
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

func (m *ForecastModel) GetSeasonality() map[int]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) < 24 {
		return nil
	}

	periodPoints := make(map[int][]float64)
	for i, p := range m.history {
		hour := i % 24
		periodPoints[hour] = append(periodPoints[hour], p.Value-m.level)
	}

	seasonality := make(map[int]float64)
	for hour, values := range periodPoints {
		if len(values) > 0 {
			var sum float64
			for _, v := range values {
				sum += v
			}
			seasonality[hour] = sum / float64(len(values))
		}
	}

	return seasonality
}

type ForecastEngine struct {
	mu     sync.RWMutex
	models map[string]*ForecastModel
	window time.Duration
}

func NewForecastEngine(window time.Duration) *ForecastEngine {
	if window == 0 {
		window = 24 * time.Hour
	}

	return &ForecastEngine{
		models: make(map[string]*ForecastModel),
		window: window,
	}
}

func (e *ForecastEngine) GetModel(name string) *ForecastModel {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if model, ok := e.models[name]; ok {
		return model
	}

	model := NewForecastModel(e.window)
	e.models[name] = model
	return model
}

func (e *ForecastEngine) AddDataPoint(ctx context.Context, name string, point DataPoint) {
	model := e.GetModel(name)
	model.AddPoint(ctx, point)
}

func (e *ForecastEngine) Predict(name string, steps int) float64 {
	model := e.GetModel(name)
	return model.Predict(steps)
}

func (e *ForecastEngine) GetAnomalies(ctx context.Context, name string, threshold float64) []DataPoint {
	e.mu.RLock()
	model, ok := e.models[name]
	e.mu.RUnlock()

	if !ok {
		return nil
	}

	model.mu.RLock()
	history := make([]DataPoint, len(model.history))
	copy(history, model.history)
	model.mu.RUnlock()

	var anomalies []DataPoint
	for _, p := range history {
		predicted := model.Predict(1)
		deviation := math.Abs(p.Value - predicted)
		if deviation > threshold {
			anomalies = append(anomalies, p)
		}
	}

	return anomalies
}

func (e *ForecastEngine) GetTrends(ctx context.Context, name string) string {
	model := e.GetModel(name)

	model.mu.RLock()
	history := make([]DataPoint, len(model.history))
	copy(history, model.history)
	model.mu.RUnlock()

	if len(history) < 2 {
		return "insufficient_data"
	}

	recentAvg := 0.0
	recentCount := 0
	for i := len(history) - 10; i < len(history); i++ {
		if i >= 0 {
			recentAvg += history[i].Value
			recentCount++
		}
	}
	if recentCount > 0 {
		recentAvg /= float64(recentCount)
	}

	olderAvg := 0.0
	olderCount := 0
	for i := 0; i < len(history)-10 && i < len(history); i++ {
		olderAvg += history[i].Value
		olderCount++
	}
	if olderCount > 0 {
		olderAvg /= float64(olderCount)
	}

	change := (recentAvg - olderAvg) / (olderAvg + 1)

	if change > 0.2 {
		return "increasing"
	} else if change < -0.2 {
		return "decreasing"
	}
	return "stable"
}

func (e *ForecastEngine) ExportModels() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]interface{})
	for name, model := range e.models {
		model.mu.RLock()
		result[name] = map[string]interface{}{
			"history_len": len(model.history),
			"initialized": model.initialized,
			"confidence":  model.GetConfidence(),
			"last_value":  model.level,
		}
		model.mu.RUnlock()
	}

	return result
}

func (e *ForecastEngine) ClearModel(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.models, name)
}

func (e *ForecastEngine) ClearAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.models = make(map[string]*ForecastModel)
}
