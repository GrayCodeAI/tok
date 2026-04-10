package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDetectorConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `
[audit.detectors.empty_runs]
enabled = false
min_severity = "high"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadDetectorConfig(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Enabled("empty_runs") {
		t.Fatal("empty_runs should be disabled")
	}
	if cfg.AllowedSeverity("empty_runs", "medium") {
		t.Fatal("medium severity should be blocked by min_severity=high")
	}
	if !cfg.AllowedSeverity("empty_runs", "high") {
		t.Fatal("high severity should be allowed")
	}
}
