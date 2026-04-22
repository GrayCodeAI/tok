package ml

import (
	"testing"
)

func TestMLCompressor(t *testing.T) {
	m := NewMLCompressor()
	input := "hello world"
	out := m.Compress(input)
	if out != input {
		t.Errorf("expected %q, got %q", input, out)
	}
}

func TestMLQualityPredictor(t *testing.T) {
	m := &MLQualityPredictor{}
	score := m.Predict("test")
	if score != 0.7 {
		t.Errorf("expected 0.7, got %f", score)
	}
}

func TestMLLayerSelector(t *testing.T) {
	m := &MLLayerSelector{}
	layers := m.SelectLayers("test")
	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}
}

func TestMLContentClassifier(t *testing.T) {
	m := &MLContentClassifier{}
	cls := m.Classify("test")
	if cls != "text" {
		t.Errorf("expected 'text', got %q", cls)
	}
}

func TestFeatureFlags(t *testing.T) {
	ff := NewFeatureFlags()
	if ff.IsEnabled("foo") {
		t.Error("expected disabled flag to return false")
	}
}

func TestCanaryDeployment(t *testing.T) {
	cd := &CanaryDeployment{percentage: 10}
	if !cd.ShouldRoute() {
		t.Error("expected ShouldRoute to return true from stub")
	}
}

func TestAutoRollback(t *testing.T) {
	ar := &AutoRollback{threshold: 0.5}
	if !ar.ShouldRollback(0.6) {
		t.Error("expected rollback when error rate exceeds threshold")
	}
	if ar.ShouldRollback(0.4) {
		t.Error("expected no rollback when error rate is below threshold")
	}
}

func TestRegressionDetector(t *testing.T) {
	rd := &RegressionDetector{baseline: 100}
	if !rd.Detect(80) {
		t.Error("expected detection when current is 20% below baseline")
	}
	if rd.Detect(95) {
		t.Error("expected no detection when current is within 10% of baseline")
	}
}
