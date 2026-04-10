package audit

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// Report is TokMan's consolidated optimization audit view.
type Report struct {
	GeneratedAt      time.Time              `json:"generated_at"`
	Days             int                    `json:"days"`
	Summary          Summary                `json:"summary"`
	WasteFindings    []Finding              `json:"waste_findings"`
	ContextOverhead  []ContextComponent     `json:"context_overhead"`
	Quality          QualityReport          `json:"quality"`
	CheckpointPolicy CheckpointPolicyReport `json:"checkpoint_policy"`
	TopLayers        []LayerSummary         `json:"top_layers"`
	Recommendations  []string               `json:"recommendations"`
}

// Summary contains top-line metrics for the selected window.
type Summary struct {
	CommandCount  int64   `json:"command_count"`
	Original      int64   `json:"original_tokens"`
	Filtered      int64   `json:"filtered_tokens"`
	Saved         int64   `json:"saved_tokens"`
	ReductionPct  float64 `json:"reduction_pct"`
	AvgExecMs     float64 `json:"avg_exec_ms"`
	ParseFailures int64   `json:"parse_failures"`
}

// Finding is a detected waste/risk pattern.
type Finding struct {
	ID              string  `json:"id"`
	Severity        string  `json:"severity"`
	Description     string  `json:"description"`
	EstimatedWaste  int64   `json:"estimated_waste_tokens"`
	EstimatedWasteD float64 `json:"estimated_waste_usd"`
	Recommendation  string  `json:"recommendation"`
}

// ContextComponent captures context metadata overhead by source kind.
type ContextComponent struct {
	Kind             string  `json:"kind"`
	CommandCount     int64   `json:"command_count"`
	OriginalTokens   int64   `json:"original_tokens"`
	SavedTokens      int64   `json:"saved_tokens"`
	AvgRelatedFiles  float64 `json:"avg_related_files"`
	BundledRatioPct  float64 `json:"bundled_ratio_pct"`
	OverheadRatioPct float64 `json:"overhead_ratio_pct"`
}

// QualitySignal is a weighted quality indicator.
type QualitySignal struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// QualityReport is an aggregate health score for token optimization quality.
type QualityReport struct {
	Score   float64         `json:"score"`
	Band    string          `json:"band"`
	Signals []QualitySignal `json:"signals"`
}

// CheckpointPolicyReport mirrors imported checkpoint-trigger strategy in TokMan form.
type CheckpointPolicyReport struct {
	RecommendedTriggers []string         `json:"recommended_triggers"`
	ObservedBands       map[string]int64 `json:"observed_bands"`
	Notes               []string         `json:"notes"`
}

// LayerSummary reports per-layer effectiveness in window.
type LayerSummary struct {
	LayerName  string  `json:"layer_name"`
	TotalSaved int64   `json:"total_saved"`
	AvgSaved   float64 `json:"avg_saved"`
	CallCount  int64   `json:"call_count"`
}

// Snapshot stores a named point-in-time audit for drift/validation comparisons.
type Snapshot struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Report    Report    `json:"report"`
}

// CompareReport shows the deltas between two snapshots.
type CompareReport struct {
	BaseName           string    `json:"base_name"`
	CandidateName      string    `json:"candidate_name"`
	ComparedAt         time.Time `json:"compared_at"`
	DeltaSavedTokens   int64     `json:"delta_saved_tokens"`
	DeltaReductionPct  float64   `json:"delta_reduction_pct"`
	DeltaQualityScore  float64   `json:"delta_quality_score"`
	DeltaParseFailures int64     `json:"delta_parse_failures"`
	Verdict            string    `json:"verdict"`
}

// Generate builds the audit report for the last N days.
func Generate(tracker *tracking.Tracker, days int) (*Report, error) {
	if days <= 0 {
		days = 30
	}
	window := fmt.Sprintf("-%d day", days)

	summary, err := querySummary(tracker, window)
	if err != nil {
		return nil, err
	}
	ctxOverhead, err := queryContextOverhead(tracker, window, summary.Original)
	if err != nil {
		return nil, err
	}
	layers, err := queryTopLayers(tracker, window)
	if err != nil {
		return nil, err
	}
	findings, err := detectWaste(tracker, window, summary)
	if err != nil {
		return nil, err
	}

	quality := scoreQuality(summary, ctxOverhead)
	policy := buildCheckpointPolicy(tracker, window, quality)

	recs := buildRecommendations(findings, quality, ctxOverhead, policy)

	return &Report{
		GeneratedAt:      time.Now().UTC(),
		Days:             days,
		Summary:          summary,
		WasteFindings:    findings,
		ContextOverhead:  ctxOverhead,
		Quality:          quality,
		CheckpointPolicy: policy,
		TopLayers:        layers,
		Recommendations:  recs,
	}, nil
}

func querySummary(tr *tracking.Tracker, window string) (Summary, error) {
	var s Summary
	err := tr.QueryRow(`
		SELECT
			COUNT(*) as command_count,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens,
			COALESCE(AVG(exec_time_ms), 0) as avg_exec_ms,
			COALESCE(SUM(CASE WHEN parse_success = 0 THEN 1 ELSE 0 END), 0) as parse_failures
		FROM commands
		WHERE timestamp >= datetime('now', ?)
	`, window).Scan(&s.CommandCount, &s.Original, &s.Filtered, &s.Saved, &s.AvgExecMs, &s.ParseFailures)
	if err != nil {
		return s, fmt.Errorf("query summary: %w", err)
	}
	if s.Original > 0 {
		s.ReductionPct = float64(s.Saved) / float64(s.Original) * 100
	}
	return s, nil
}

func queryContextOverhead(tr *tracking.Tracker, window string, totalOriginal int64) ([]ContextComponent, error) {
	rows, err := tr.Query(`
		SELECT
			COALESCE(NULLIF(context_kind, ''), 'none') as kind,
			COUNT(*) as command_count,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens,
			COALESCE(AVG(context_related_files), 0) as avg_related_files,
			COALESCE(AVG(CASE WHEN context_bundle = 1 THEN 100.0 ELSE 0 END), 0) as bundled_ratio
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY kind
		ORDER BY original_tokens DESC
	`, window)
	if err != nil {
		return nil, fmt.Errorf("query context overhead: %w", err)
	}
	defer rows.Close()

	var out []ContextComponent
	for rows.Next() {
		var c ContextComponent
		if err := rows.Scan(&c.Kind, &c.CommandCount, &c.OriginalTokens, &c.SavedTokens, &c.AvgRelatedFiles, &c.BundledRatioPct); err != nil {
			return nil, err
		}
		if totalOriginal > 0 {
			c.OverheadRatioPct = float64(c.OriginalTokens) / float64(totalOriginal) * 100
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func queryTopLayers(tr *tracking.Tracker, window string) ([]LayerSummary, error) {
	rows, err := tr.Query(`
		SELECT
			ls.layer_name,
			COALESCE(SUM(ls.tokens_saved), 0) as total_saved,
			COALESCE(AVG(ls.tokens_saved), 0) as avg_saved,
			COUNT(*) as call_count
		FROM layer_stats ls
		JOIN commands c ON c.id = ls.command_id
		WHERE c.timestamp >= datetime('now', ?)
		GROUP BY ls.layer_name
		ORDER BY total_saved DESC
		LIMIT 10
	`, window)
	if err != nil {
		return nil, fmt.Errorf("query top layers: %w", err)
	}
	defer rows.Close()

	var out []LayerSummary
	for rows.Next() {
		var l LayerSummary
		if err := rows.Scan(&l.LayerName, &l.TotalSaved, &l.AvgSaved, &l.CallCount); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func detectWaste(tr *tracking.Tracker, window string, summary Summary) ([]Finding, error) {
	findings := make([]Finding, 0, 8)

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
		return nil, fmt.Errorf("detect empty runs: %w", err)
	}
	if emptyCount > 0 {
		findings = append(findings, Finding{
			ID:              "empty_runs",
			Severity:        severityForRatio(float64(emptyCount), float64(max(summary.CommandCount, 1))),
			Description:     fmt.Sprintf("%d high-input runs produced near-zero compression output", emptyCount),
			EstimatedWaste:  emptyWaste,
			EstimatedWasteD: tokensToUSD(emptyWaste),
			Recommendation:  "Enable stricter checkpoint/compaction triggers for long sessions and idle outputs.",
		})
	}

	if summary.ParseFailures > 0 {
		findings = append(findings, Finding{
			ID:              "parse_failures",
			Severity:        severityForRatio(float64(summary.ParseFailures), float64(max(summary.CommandCount, 1))),
			Description:     fmt.Sprintf("%d parse failures reduced optimizer reliability", summary.ParseFailures),
			EstimatedWaste:  summary.ParseFailures * 1000,
			EstimatedWasteD: tokensToUSD(summary.ParseFailures * 1000),
			Recommendation:  "Harden parser fallback paths for high-frequency command families.",
		})
	}

	rows, err := tr.Query(`
		SELECT
			command,
			COUNT(*) as call_count,
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
		return nil, fmt.Errorf("detect low-efficiency commands: %w", err)
	}
	defer rows.Close()

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
		if reduction < 0.15 {
			waste := int64(float64(original)*0.50) - saved
			if waste < 0 {
				waste = 0
			}
			findings = append(findings, Finding{
				ID:              "low_efficiency_" + safeID(cmd),
				Severity:        "medium",
				Description:     fmt.Sprintf("Command '%s' is under-optimized (%.1f%% reduction over %d calls)", cmd, reduction*100, calls),
				EstimatedWaste:  waste,
				EstimatedWasteD: tokensToUSD(waste),
				Recommendation:  fmt.Sprintf("Tune layer profile or TOML filter for '%s'.", cmd),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	modelRows, err := tr.Query(`
		SELECT
			COALESCE(model_name, ''),
			COUNT(*) as call_count,
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
		return nil, fmt.Errorf("detect model routing: %w", err)
	}
	defer modelRows.Close()
	for modelRows.Next() {
		var model string
		var calls, original, saved int64
		if err := modelRows.Scan(&model, &calls, &original, &saved); err != nil {
			return nil, err
		}
		m := strings.ToLower(model)
		if strings.Contains(m, "opus") || strings.Contains(m, "gpt-5") || strings.Contains(m, "sonnet") {
			waste := int64(float64(original)*0.3) - saved
			if waste < 0 {
				waste = 0
			}
			findings = append(findings, Finding{
				ID:              "model_routing",
				Severity:        "medium",
				Description:     fmt.Sprintf("High-tier model '%s' used for low-complexity command workflows (%d calls)", model, calls),
				EstimatedWaste:  waste,
				EstimatedWasteD: tokensToUSD(waste),
				Recommendation:  "Route repetitive/low-complexity flows to lower-cost model profiles.",
			})
		}
	}
	if err := modelRows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity == findings[j].Severity {
			return findings[i].EstimatedWaste > findings[j].EstimatedWaste
		}
		return severityRank(findings[i].Severity) > severityRank(findings[j].Severity)
	})

	if len(findings) > 12 {
		findings = findings[:12]
	}
	return findings, nil
}

func scoreQuality(summary Summary, context []ContextComponent) QualityReport {
	if summary.CommandCount == 0 {
		return QualityReport{Score: 0, Band: "insufficient-data"}
	}

	parseSuccessPct := 100.0
	if summary.CommandCount > 0 {
		parseSuccessPct = 100 - (float64(summary.ParseFailures)/float64(summary.CommandCount))*100
	}

	latencyScore := 100.0
	switch {
	case summary.AvgExecMs <= 200:
		latencyScore = 100
	case summary.AvgExecMs <= 500:
		latencyScore = 85
	case summary.AvgExecMs <= 1000:
		latencyScore = 70
	case summary.AvgExecMs <= 2000:
		latencyScore = 50
	default:
		latencyScore = 30
	}

	contextOverhead := 0.0
	for _, c := range context {
		if c.Kind != "none" {
			contextOverhead += c.OverheadRatioPct
		}
	}
	contextScore := 100.0
	switch {
	case contextOverhead < 20:
		contextScore = 100
	case contextOverhead < 35:
		contextScore = 80
	case contextOverhead < 50:
		contextScore = 60
	default:
		contextScore = 40
	}

	signals := []QualitySignal{
		{Name: "Compression Efficiency", Weight: 0.40, Score: clamp(summary.ReductionPct, 0, 100), Description: fmt.Sprintf("Average reduction %.1f%%", summary.ReductionPct)},
		{Name: "Parse Reliability", Weight: 0.25, Score: clamp(parseSuccessPct, 0, 100), Description: fmt.Sprintf("Parse success %.1f%%", parseSuccessPct)},
		{Name: "Latency Efficiency", Weight: 0.20, Score: latencyScore, Description: fmt.Sprintf("Average execution %.0fms", summary.AvgExecMs)},
		{Name: "Context Discipline", Weight: 0.15, Score: contextScore, Description: fmt.Sprintf("Context overhead %.1f%%", contextOverhead)},
	}

	score := 0.0
	for _, s := range signals {
		score += s.Score * s.Weight
	}

	band := "poor"
	switch {
	case score >= 85:
		band = "excellent"
	case score >= 70:
		band = "good"
	case score >= 55:
		band = "fair"
	}

	return QualityReport{Score: score, Band: band, Signals: signals}
}

func buildCheckpointPolicy(tr *tracking.Tracker, window string, quality QualityReport) CheckpointPolicyReport {
	bands := map[string]int64{}
	for _, threshold := range []int{20000, 50000, 100000} {
		var count int64
		_ = tr.QueryRow(
			"SELECT COALESCE(COUNT(*),0) FROM commands WHERE timestamp >= datetime('now', ?) AND original_tokens >= ?",
			window, threshold,
		).Scan(&count)
		bands[fmt.Sprintf("progressive_%d", threshold)] = count
	}

	triggers := make([]string, 0, 8)
	notes := make([]string, 0, 8)
	if bands["progressive_20000"] > 0 {
		triggers = append(triggers, "progressive-20")
		notes = append(notes, "Capture checkpoint at ~20K-token sessions before context quality decays.")
	}
	if bands["progressive_50000"] > 0 {
		triggers = append(triggers, "progressive-50")
		notes = append(notes, "Add deeper compaction checkpoint at ~50K-token sessions.")
	}
	if quality.Score < 80 {
		triggers = append(triggers, "quality-80")
		notes = append(notes, "Trigger checkpoint when quality falls below 80.")
	}
	if quality.Score < 70 {
		triggers = append(triggers, "quality-70")
		notes = append(notes, "Escalate to aggressive compaction under quality < 70.")
	}
	if len(triggers) == 0 {
		triggers = append(triggers, "stable-no-trigger")
		notes = append(notes, "Current window is stable; monitor for threshold crossings.")
	}

	return CheckpointPolicyReport{
		RecommendedTriggers: uniqSorted(triggers),
		ObservedBands:       bands,
		Notes:               notes,
	}
}

func buildRecommendations(findings []Finding, quality QualityReport, context []ContextComponent, policy CheckpointPolicyReport) []string {
	recs := []string{}
	if quality.Score < 70 {
		recs = append(recs, "Enable stricter compaction + budget gating for long sessions by default.")
	}
	if len(findings) > 0 {
		recs = append(recs, "Prioritize top waste findings by estimated waste tokens and patch command-specific filters.")
	}
	for _, c := range context {
		if c.Kind != "none" && c.OverheadRatioPct >= 25 {
			recs = append(recs, fmt.Sprintf("Reduce '%s' context overhead (%.1f%% of window).", c.Kind, c.OverheadRatioPct))
		}
	}
	if len(policy.RecommendedTriggers) > 0 && policy.RecommendedTriggers[0] != "stable-no-trigger" {
		recs = append(recs, "Adopt checkpoint triggers from policy recommendations to prevent late-session degradation.")
	}
	if len(recs) == 0 {
		recs = append(recs, "Optimization posture is healthy for the selected window.")
	}
	return uniqSorted(recs)
}

// SaveSnapshot writes a named drift-validation snapshot to disk.
func SaveSnapshot(dir, name string, report *Report) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = time.Now().UTC().Format("20060102-150405")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	s := Snapshot{Name: safeID(name), CreatedAt: time.Now().UTC(), Report: *report}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, s.Name+".json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// LoadSnapshot loads a stored audit snapshot.
func LoadSnapshot(path string) (*Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Compare computes high-level deltas between snapshots.
func Compare(base, candidate *Snapshot) CompareReport {
	deltaSaved := candidate.Report.Summary.Saved - base.Report.Summary.Saved
	deltaReduction := candidate.Report.Summary.ReductionPct - base.Report.Summary.ReductionPct
	deltaQuality := candidate.Report.Quality.Score - base.Report.Quality.Score
	deltaParseFailures := candidate.Report.Summary.ParseFailures - base.Report.Summary.ParseFailures

	verdict := "neutral"
	if deltaReduction >= 3 && deltaQuality >= 0 && deltaParseFailures <= 0 {
		verdict = "improved"
	} else if deltaReduction < -2 || deltaQuality < -3 || deltaParseFailures > 0 {
		verdict = "regressed"
	}

	return CompareReport{
		BaseName:           base.Name,
		CandidateName:      candidate.Name,
		ComparedAt:         time.Now().UTC(),
		DeltaSavedTokens:   deltaSaved,
		DeltaReductionPct:  deltaReduction,
		DeltaQualityScore:  deltaQuality,
		DeltaParseFailures: deltaParseFailures,
		Verdict:            verdict,
	}
}

// RenderHTML writes a lightweight dashboard HTML report.
func RenderHTML(path string, report *Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	const tmpl = `<!doctype html>
<html><head><meta charset="utf-8"><title>TokMan Audit</title>
<style>
body{font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;background:#f7f8fb;color:#111827}
.card{background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:14px;margin-bottom:14px}
table{width:100%;border-collapse:collapse}th,td{padding:8px;border-bottom:1px solid #eee;text-align:left}
.badge{display:inline-block;padding:2px 8px;border-radius:12px;background:#eef2ff}
</style></head><body>
<h1>TokMan Audit</h1>
<div class="card"><b>Window:</b> {{.Days}} days | <b>Generated:</b> {{.GeneratedAt}}</div>
<div class="card"><h2>Summary</h2>
<p>Commands: {{.Summary.CommandCount}} | Original: {{.Summary.Original}} | Filtered: {{.Summary.Filtered}} | Saved: {{.Summary.Saved}}</p>
<p>Reduction: {{printf "%.2f" .Summary.ReductionPct}}% | Quality: <span class="badge">{{printf "%.1f" .Quality.Score}} ({{.Quality.Band}})</span></p>
</div>
<div class="card"><h2>Waste Findings</h2>
<table><tr><th>ID</th><th>Severity</th><th>Description</th><th>Waste Tokens</th></tr>
{{range .WasteFindings}}<tr><td>{{.ID}}</td><td>{{.Severity}}</td><td>{{.Description}}</td><td>{{.EstimatedWaste}}</td></tr>{{end}}
</table></div>
<div class="card"><h2>Top Layers</h2>
<table><tr><th>Layer</th><th>Total Saved</th><th>Avg Saved</th><th>Calls</th></tr>
{{range .TopLayers}}<tr><td>{{.LayerName}}</td><td>{{.TotalSaved}}</td><td>{{printf "%.1f" .AvgSaved}}</td><td>{{.CallCount}}</td></tr>{{end}}
</table></div>
<div class="card"><h2>Recommendations</h2><ul>{{range .Recommendations}}<li>{{.}}</li>{{end}}</ul></div>
</body></html>`
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return template.Must(template.New("audit").Parse(tmpl)).Execute(f, report)
}

func tokensToUSD(tokens int64) float64 {
	// Conservative blended estimate used for ranking only: $3 / 1M tokens.
	return (float64(tokens) / 1_000_000.0) * 3.0
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func safeID(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "snapshot"
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "snapshot"
	}
	return out
}

func severityRank(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func severityForRatio(part, whole float64) string {
	if whole <= 0 {
		return "low"
	}
	r := part / whole
	switch {
	case r >= 0.35:
		return "high"
	case r >= 0.15:
		return "medium"
	default:
		return "low"
	}
}

func uniqSorted(in []string) []string {
	if len(in) == 0 {
		return in
	}
	m := make(map[string]struct{}, len(in))
	for _, item := range in {
		if strings.TrimSpace(item) == "" {
			continue
		}
		m[item] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
