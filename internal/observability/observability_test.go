package observability

import "testing"

func TestStructuredLogger(t *testing.T) {
	l := NewStructuredLogger()

	l.Debug("test message", map[string]interface{}{"key": "value"})
	l.Info("info message", nil)
	l.Error("error message", nil)

	entries := l.GetEntries()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	formatted := l.Format(entries[0])
	if formatted == "" {
		t.Error("Expected non-empty formatted output")
	}
}

func TestDistributedTracer(t *testing.T) {
	tr := NewDistributedTracer()

	span := tr.StartSpan("s1", "test-span", "")
	tr.EndSpan(span, "success")

	spans := tr.GetSpans()
	if len(spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(spans))
	}
	if spans[0].Status != "success" {
		t.Errorf("Expected success status, got %s", spans[0].Status)
	}
}

func TestMetricCollector(t *testing.T) {
	m := NewMetricCollector()

	m.Inc("requests_total", 1)
	m.Inc("requests_total", 5)
	m.Set("cpu_usage", 75.5)
	m.Observe("latency", 0.5)

	if m.GetCounter("requests_total") != 6 {
		t.Errorf("Expected 6, got %d", m.GetCounter("requests_total"))
	}
	if m.GetGauge("cpu_usage") != 75.5 {
		t.Errorf("Expected 75.5, got %f", m.GetGauge("cpu_usage"))
	}

	prom := m.ExportPrometheus()
	if prom == "" {
		t.Error("Expected non-empty Prometheus export")
	}
}

func TestSLAReporter(t *testing.T) {
	s := NewSLAReporter(99.9, 100)

	s.Update(99.95, 50)
	if !s.IsHealthy() {
		t.Error("Should be healthy")
	}

	s.Update(95.0, 200)
	if s.IsHealthy() {
		t.Error("Should not be healthy")
	}

	report := s.Report()
	if report["healthy"].(bool) {
		t.Error("Expected unhealthy in report")
	}
}
