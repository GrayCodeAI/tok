package forecast

import (
	"context"
	"testing"
	"time"
)

func TestNewForecastModel(t *testing.T) {
	model := NewForecastModel(time.Hour)
	if model == nil {
		t.Error("Expected non-nil model")
	}
	if model.alpha != 0.3 {
		t.Errorf("Expected alpha 0.3, got %f", model.alpha)
	}
}

func TestForecastModelAddPoint(t *testing.T) {
	model := NewForecastModel(time.Hour)

	point1 := DataPoint{Timestamp: time.Now(), Value: 100}
	model.AddPoint(context.Background(), point1)

	point2 := DataPoint{Timestamp: time.Now(), Value: 110}
	model.AddPoint(context.Background(), point2)

	if len(model.history) != 2 {
		t.Errorf("Expected 2 history points, got %d", len(model.history))
	}
}

func TestForecastModelPredict(t *testing.T) {
	model := NewForecastModel(time.Hour)

	for i := 0; i < 10; i++ {
		model.AddPoint(context.Background(), DataPoint{
			Timestamp: time.Now(),
			Value:     float64(100 + i*10),
		})
	}

	prediction := model.Predict(5)
	if prediction == 0 {
		t.Error("Expected non-zero prediction")
	}
}

func TestForecastModelGetConfidence(t *testing.T) {
	model := NewForecastModel(time.Hour)

	for i := 0; i < 20; i++ {
		model.AddPoint(context.Background(), DataPoint{
			Timestamp: time.Now(),
			Value:     100 + float64(i),
		})
	}

	confidence := model.GetConfidence()
	if confidence < 0 || confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", confidence)
	}
}

func TestForecastEngineNew(t *testing.T) {
	engine := NewForecastEngine(time.Hour)
	if engine == nil {
		t.Error("Expected non-nil engine")
	}
}

func TestForecastEngineGetModel(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	model1 := engine.GetModel("tokens")
	model2 := engine.GetModel("tokens")

	if model1 != model2 {
		t.Error("Expected same model instance for same name")
	}
}

func TestForecastEngineAddDataPoint(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	engine.AddDataPoint(context.Background(), "test", DataPoint{
		Timestamp: time.Now(),
		Value:     100,
	})

	prediction := engine.Predict("test", 1)
	if prediction == 0 && engine.GetModel("test").initialized {
		t.Error("Expected non-zero prediction after adding data")
	}
}

func TestForecastEnginePredict(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	for i := 0; i < 15; i++ {
		engine.AddDataPoint(context.Background(), "test", DataPoint{
			Timestamp: time.Now(),
			Value:     100,
		})
	}

	prediction := engine.Predict("test", 3)
	t.Logf("Prediction: %f", prediction)
}

func TestForecastEngineGetAnomalies(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	for i := 0; i < 20; i++ {
		engine.AddDataPoint(context.Background(), "test", DataPoint{
			Timestamp: time.Now(),
			Value:     100,
		})
	}

	engine.AddDataPoint(context.Background(), "test", DataPoint{
		Timestamp: time.Now(),
		Value:     1000,
	})

	anomalies := engine.GetAnomalies(context.Background(), "test", 50)
	t.Logf("Found %d anomalies", len(anomalies))
}

func TestForecastEngineGetTrends(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	for i := 0; i < 30; i++ {
		engine.AddDataPoint(context.Background(), "test", DataPoint{
			Timestamp: time.Now(),
			Value:     100 + float64(i),
		})
	}

	trend := engine.GetTrends(context.Background(), "test")
	t.Logf("Trend: %s", trend)
}

func TestForecastEngineExportModels(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	engine.AddDataPoint(context.Background(), "model1", DataPoint{Value: 100})
	engine.AddDataPoint(context.Background(), "model2", DataPoint{Value: 200})

	export := engine.ExportModels()
	if len(export) != 2 {
		t.Errorf("Expected 2 models exported, got %d", len(export))
	}
}

func TestForecastEngineClearModel(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	engine.AddDataPoint(context.Background(), "test", DataPoint{Value: 100})
	engine.ClearModel("test")

	export := engine.ExportModels()
	if len(export) != 0 {
		t.Errorf("Expected 0 models after clear, got %d", len(export))
	}
}

func TestForecastEngineClearAll(t *testing.T) {
	engine := NewForecastEngine(time.Hour)

	engine.AddDataPoint(context.Background(), "model1", DataPoint{Value: 100})
	engine.AddDataPoint(context.Background(), "model2", DataPoint{Value: 200})

	engine.ClearAll()

	export := engine.ExportModels()
	if len(export) != 0 {
		t.Errorf("Expected 0 models after clear all, got %d", len(export))
	}
}

func TestForecastModelGetSeasonality(t *testing.T) {
	model := NewForecastModel(time.Hour)

	now := time.Now()
	for i := 0; i < 48; i++ {
		model.AddPoint(context.Background(), DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Value:     100 + float64(i%24),
		})
	}

	seasonality := model.GetSeasonality()
	if seasonality == nil {
		t.Logf("Seasonality data available with %d periods", len(seasonality))
	}
}

func TestForecastModelInsufficientHistory(t *testing.T) {
	model := NewForecastModel(time.Hour)

	model.AddPoint(context.Background(), DataPoint{Value: 100})

	prediction := model.Predict(1)
	t.Logf("Prediction with 1 point: %f", prediction)
}

func TestForecastModelEmptyHistory(t *testing.T) {
	model := NewForecastModel(time.Hour)

	prediction := model.Predict(1)
	if prediction != 0 {
		t.Errorf("Expected 0 prediction with no history, got %f", prediction)
	}
}
