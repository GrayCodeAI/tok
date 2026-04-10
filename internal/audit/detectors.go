package audit

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// Detector evaluates audit data and returns findings.
type Detector interface {
	ID() string
	Run(tr *tracking.Tracker, window string, summary Summary) ([]Finding, error)
}

type detector struct {
	id string
	fn func(tr *tracking.Tracker, window string, summary Summary) ([]Finding, error)
}

func (d detector) ID() string { return d.id }
func (d detector) Run(tr *tracking.Tracker, window string, summary Summary) ([]Finding, error) {
	return d.fn(tr, window, summary)
}

// DefaultDetectors returns built-in waste detectors.
func DefaultDetectors() []Detector {
	return []Detector{
		detector{id: "empty_runs", fn: detectEmptyRuns},
		detector{id: "parse_failures", fn: detectParseFailures},
		detector{id: "low_efficiency", fn: detectLowEfficiency},
		detector{id: "model_routing", fn: detectModelRouting},
	}
}

// DetectorConfig controls detector activation and severity filtering.
type DetectorConfig struct {
	Detectors map[string]DetectorRule
}

// DetectorRule config entry for one detector.
type DetectorRule struct {
	Enabled     bool
	MinSeverity string
}

// Enabled returns whether detector is enabled.
func (c DetectorConfig) Enabled(id string) bool {
	if c.Detectors == nil {
		return true
	}
	r, ok := c.Detectors[id]
	if !ok {
		return true
	}
	return r.Enabled
}

// AllowedSeverity checks per-detector minimum severity.
func (c DetectorConfig) AllowedSeverity(id, severity string) bool {
	if c.Detectors == nil {
		return true
	}
	r, ok := c.Detectors[id]
	if !ok || strings.TrimSpace(r.MinSeverity) == "" {
		return true
	}
	return severityRank(strings.ToLower(severity)) >= severityRank(strings.ToLower(r.MinSeverity))
}

type auditConfigFile struct {
	Audit struct {
		Detectors map[string]struct {
			Enabled     *bool  `toml:"enabled"`
			MinSeverity string `toml:"min_severity"`
		} `toml:"detectors"`
	} `toml:"audit"`
}

// LoadDetectorConfig reads [audit.detectors] config from tokman config file.
func LoadDetectorConfig(configPath string) (DetectorConfig, error) {
	cfg := DetectorConfig{Detectors: map[string]DetectorRule{}}
	if strings.TrimSpace(configPath) == "" {
		return cfg, nil
	}
	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return cfg, fmt.Errorf("stat audit detectors config: %w", err)
		}
		return cfg, nil
	}
	var file auditConfigFile
	if _, err := toml.DecodeFile(configPath, &file); err != nil {
		return cfg, fmt.Errorf("decode audit detectors config: %w", err)
	}
	for id, raw := range file.Audit.Detectors {
		rule := DetectorRule{Enabled: true, MinSeverity: strings.ToLower(strings.TrimSpace(raw.MinSeverity))}
		if raw.Enabled != nil {
			rule.Enabled = *raw.Enabled
		}
		cfg.Detectors[id] = rule
	}
	return cfg, nil
}

func detectEmptyRuns(tr *tracking.Tracker, window string, summary Summary) ([]Finding, error) {
	var emptyCount, emptyWaste int64
	err := tr.QueryRow(`
		SELECT
			COALESCE(COUNT(*), 0),
			COALESCE(SUM(filtered_tokens), 0)
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		  AND original_tokens >= 5000
		  AND filtered_tokens >= CAST(original_tokens * 0.95 AS INTEGER)
	`, window).Scan(&emptyCount, &emptyWaste)
	if err != nil {
		return nil, err
	}
	if emptyCount == 0 {
		return nil, nil
	}
	return []Finding{{
		ID:              "empty_runs",
		Severity:        severityForRatio(float64(emptyCount), float64(max(summary.CommandCount, 1))),
		Description:     fmt.Sprintf("%d high-input runs produced near-zero compression output", emptyCount),
		EstimatedWaste:  emptyWaste,
		EstimatedWasteD: tokensToUSD(emptyWaste),
		Recommendation:  "Enable stricter checkpoint/compaction triggers for long sessions and idle outputs.",
	}}, nil
}

func detectParseFailures(_ *tracking.Tracker, _ string, summary Summary) ([]Finding, error) {
	if summary.ParseFailures <= 0 {
		return nil, nil
	}
	return []Finding{{
		ID:              "parse_failures",
		Severity:        severityForRatio(float64(summary.ParseFailures), float64(max(summary.CommandCount, 1))),
		Description:     fmt.Sprintf("%d parse failures reduced optimizer reliability", summary.ParseFailures),
		EstimatedWaste:  summary.ParseFailures * 1000,
		EstimatedWasteD: tokensToUSD(summary.ParseFailures * 1000),
		Recommendation:  "Harden parser fallback paths for high-frequency command families.",
	}}, nil
}

func detectLowEfficiency(tr *tracking.Tracker, window string, _ Summary) ([]Finding, error) {
	rows, err := tr.Query(`
		SELECT command, COUNT(*) as call_count,
		       COALESCE(SUM(original_tokens), 0) as original_tokens,
		       COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY command
		HAVING original_tokens >= 2000
		ORDER BY (CAST(saved_tokens AS REAL) / NULLIF(original_tokens, 0)) ASC, original_tokens DESC
		LIMIT 5
	`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Finding, 0, 5)
	for rows.Next() {
		var cmd string
		var calls, original, saved int64
		if err := rows.Scan(&cmd, &calls, &original, &saved); err != nil {
			return nil, err
		}
		reduction := 0.0
		if original > 0 {
			reduction = float64(saved) / float64(original)
		}
		if reduction >= 0.15 {
			continue
		}
		waste := int64(float64(original)*0.50) - saved
		if waste < 0 {
			waste = 0
		}
		out = append(out, Finding{
			ID:              "low_efficiency_" + safeID(cmd),
			Severity:        "medium",
			Description:     fmt.Sprintf("Command '%s' is under-optimized (%.1f%% reduction over %d calls)", cmd, reduction*100, calls),
			EstimatedWaste:  waste,
			EstimatedWasteD: tokensToUSD(waste),
			Recommendation:  fmt.Sprintf("Tune layer profile or TOML filter for '%s'.", cmd),
		})
	}
	return out, rows.Err()
}

func detectModelRouting(tr *tracking.Tracker, window string, _ Summary) ([]Finding, error) {
	rows, err := tr.Query(`
		SELECT COALESCE(model_name, ''), COUNT(*) as call_count,
		       COALESCE(SUM(original_tokens), 0) as original_tokens,
		       COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		  AND model_name IS NOT NULL
		  AND model_name != ''
		  AND (command GLOB 'ls*' OR command GLOB 'pwd*' OR command GLOB 'git status*' OR command GLOB 'wc*')
		GROUP BY model_name
		HAVING call_count >= 3
	`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Finding, 0, 3)
	for rows.Next() {
		var model string
		var calls, original, saved int64
		if err := rows.Scan(&model, &calls, &original, &saved); err != nil {
			return nil, err
		}
		m := strings.ToLower(model)
		if !(strings.Contains(m, "opus") || strings.Contains(m, "gpt-5") || strings.Contains(m, "sonnet")) {
			continue
		}
		waste := int64(float64(original)*0.3) - saved
		if waste < 0 {
			waste = 0
		}
		out = append(out, Finding{
			ID:              "model_routing",
			Severity:        "medium",
			Description:     fmt.Sprintf("High-tier model '%s' used for low-complexity command workflows (%d calls)", model, calls),
			EstimatedWaste:  waste,
			EstimatedWasteD: tokensToUSD(waste),
			Recommendation:  "Route repetitive/low-complexity flows to lower-cost model profiles.",
		})
	}
	return out, rows.Err()
}
