package httpproxy

import "testing"

func TestAdaptiveScaler(t *testing.T) {
	s := NewAdaptiveScaler()

	short := "short input"
	mode := s.GetMode(short)
	if mode != "surface" {
		t.Errorf("Expected surface for short input, got %s", mode)
	}

	long := string(make([]byte, 25000))
	mode = s.GetMode(long)
	if mode != "core" {
		t.Errorf("Expected core for long input, got %s", mode)
	}
}

func TestModelFallbackManager(t *testing.T) {
	m := NewModelFallbackManager("gpt-4o", "claude-3-haiku", "gemini-1.5-flash")

	model := m.GetModel(200)
	if model != "gpt-4o" {
		t.Errorf("Expected primary model, got %s", model)
	}

	model = m.GetModel(429)
	if model == "gpt-4o" {
		t.Error("Expected fallback on 429")
	}

	all := m.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 models, got %d", len(all))
	}
}

func TestOpenTelemetryCollector(t *testing.T) {
	c := NewOpenTelemetryCollector(&OpenTelemetryConfig{
		ServiceName: "tokman",
		Endpoint:    "localhost:4317",
	})

	c.RecordMetric("tokens_processed", 1000, map[string]string{"model": "gpt-4o"})
	output := c.ExportMetrics()
	if output == "" {
		t.Error("Expected non-empty export")
	}
}
