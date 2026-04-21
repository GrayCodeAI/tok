package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

// Report is tok's consolidated optimization audit view.
type Report struct {
	GeneratedAt      time.Time              `json:"generated_at"`
	Days             int                    `json:"days"`
	DriftFingerprint string                 `json:"drift_fingerprint"`
	Summary          Summary                `json:"summary"`
	TurnAnalytics    []TurnMetric           `json:"turn_analytics"`
	CostlyPrompts    []CostlyPrompt         `json:"costly_prompts"`
	IntentProfiles   []IntentProfile        `json:"intent_profiles"`
	AgentBudgets     []AgentBudget          `json:"agent_budgets"`
	BudgetController BudgetControllerReport `json:"budget_controller"`
	AnchorRetention  AnchorRetentionReport  `json:"anchor_retention"`
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

// CheckpointPolicyReport mirrors imported checkpoint-trigger strategy in tok form.
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

// TurnMetric is per-command turn-level analytics.
type TurnMetric struct {
	ID          int64     `json:"id"`
	Command     string    `json:"command"`
	Timestamp   time.Time `json:"timestamp"`
	Original    int64     `json:"original_tokens"`
	Saved       int64     `json:"saved_tokens"`
	Reduction   float64   `json:"reduction_pct"`
	EstimatedUS float64   `json:"estimated_cost_usd"`
}

// CostlyPrompt captures the highest-cost command prompts.
type CostlyPrompt struct {
	Command     string  `json:"command"`
	Count       int64   `json:"count"`
	Original    int64   `json:"original_tokens"`
	Saved       int64   `json:"saved_tokens"`
	EstimatedUS float64 `json:"estimated_cost_usd"`
}

// IntentProfile summarizes compression quality by inferred task intent.
type IntentProfile struct {
	Intent       string  `json:"intent"`
	Commands     int64   `json:"commands"`
	Original     int64   `json:"original_tokens"`
	Saved        int64   `json:"saved_tokens"`
	ReductionPct float64 `json:"reduction_pct"`
}

// AgentBudget allocates usage/savings budgets across agents.
type AgentBudget struct {
	Agent        string  `json:"agent"`
	Commands     int64   `json:"commands"`
	Original     int64   `json:"original_tokens"`
	Saved        int64   `json:"saved_tokens"`
	BudgetShare  float64 `json:"budget_share_pct"`
	ReductionPct float64 `json:"reduction_pct"`
	EstimatedUS  float64 `json:"estimated_cost_usd"`
	SavingsUS    float64 `json:"estimated_savings_usd"`
}

// BudgetControllerReport provides adaptive budget recommendations (LLMLingua/LazyLLM style).
type BudgetControllerReport struct {
	CurrentReductionPct float64            `json:"current_reduction_pct"`
	QualityScore        float64            `json:"quality_score"`
	RecommendedMode     string             `json:"recommended_mode"`
	Bands               map[string]float64 `json:"bands"`
	DecaySchedule       []int              `json:"decay_schedule"`
}

// AnchorRetentionReport scores how well critical anchors are preserved.
type AnchorRetentionReport struct {
	SignalsFound      int64   `json:"signals_found"`
	SignalDensityPct  float64 `json:"signal_density_pct"`
	EstimatedKeepRate float64 `json:"estimated_keep_rate_pct"`
	Grade             string  `json:"grade"`
}

// Snapshot stores a named point-in-time audit for drift/validation comparisons.
type Snapshot struct {
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	Fingerprint string    `json:"fingerprint"`
	Report      Report    `json:"report"`
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
	DriftChanged       bool      `json:"drift_changed"`
	Verdict            string    `json:"verdict"`
}

// GenerateOptions controls optional audit behavior.
type GenerateOptions struct {
	ConfigPath string
}

// GenerateWithOptions builds the audit report with additional runtime options.
func GenerateWithOptions(tracker *tracking.Tracker, days int, opts GenerateOptions) (*Report, error) {
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
	turns, err := queryTurnAnalytics(tracker, window, 50)
	if err != nil {
		return nil, err
	}
	costly, err := queryCostlyPrompts(tracker, window, 10)
	if err != nil {
		return nil, err
	}
	intentProfiles, err := queryIntentProfiles(tracker, window)
	if err != nil {
		return nil, err
	}
	agentBudgets, err := queryAgentBudgets(tracker, window, summary.Original)
	if err != nil {
		return nil, err
	}
	anchorRetention, err := queryAnchorRetention(tracker, window, summary.CommandCount)
	if err != nil {
		return nil, err
	}
	detectorCfg, _ := LoadDetectorConfig(opts.ConfigPath)
	findings, err := detectWaste(tracker, window, summary, detectorCfg)
	if err != nil {
		return nil, err
	}

	quality := scoreQuality(summary, ctxOverhead)
	policy := buildCheckpointPolicy(tracker, window, quality)
	telemetry, telErr := tracker.GetCheckpointTelemetry(days)
	if telErr == nil && telemetry != nil {
		if policy.ObservedBands == nil {
			policy.ObservedBands = map[string]int64{}
		}
		for k, v := range telemetry.ByTrigger {
			policy.ObservedBands["event_"+k] = v
		}
	}
	fingerprint, _ := DriftFingerprint(opts.ConfigPath)
	budgetController := buildBudgetController(summary, quality, intentProfiles)

	recs := buildRecommendations(findings, quality, ctxOverhead, policy)

	return &Report{
		GeneratedAt:      time.Now().UTC(),
		Days:             days,
		DriftFingerprint: fingerprint,
		Summary:          summary,
		TurnAnalytics:    turns,
		CostlyPrompts:    costly,
		IntentProfiles:   intentProfiles,
		AgentBudgets:     agentBudgets,
		BudgetController: budgetController,
		AnchorRetention:  anchorRetention,
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

func queryTurnAnalytics(tr *tracking.Tracker, window string, limit int) ([]TurnMetric, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := tr.Query(`
		SELECT id, command, COALESCE(original_tokens,0), COALESCE(saved_tokens,0),
		       COALESCE(CAST(strftime('%s', timestamp) AS INTEGER), 0)
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		ORDER BY timestamp DESC
		LIMIT ?
	`, window, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TurnMetric, 0, limit)
	for rows.Next() {
		var tm TurnMetric
		var epoch int64
		if err := rows.Scan(&tm.ID, &tm.Command, &tm.Original, &tm.Saved, &epoch); err != nil {
			return nil, err
		}
		if tm.Original > 0 {
			tm.Reduction = (float64(tm.Saved) / float64(tm.Original)) * 100
		}
		tm.EstimatedUS = tokensToUSD(tm.Original)
		tm.Timestamp = time.Unix(epoch, 0).UTC()
		out = append(out, tm)
	}
	return out, rows.Err()
}

func queryCostlyPrompts(tr *tracking.Tracker, window string, limit int) ([]CostlyPrompt, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := tr.Query(`
		SELECT command, COUNT(*) as cnt,
		       COALESCE(SUM(original_tokens),0), COALESCE(SUM(saved_tokens),0)
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY command
		ORDER BY SUM(original_tokens) DESC
		LIMIT ?
	`, window, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CostlyPrompt, 0, limit)
	for rows.Next() {
		var cp CostlyPrompt
		if err := rows.Scan(&cp.Command, &cp.Count, &cp.Original, &cp.Saved); err != nil {
			return nil, err
		}
		cp.EstimatedUS = tokensToUSD(cp.Original)
		out = append(out, cp)
	}
	return out, rows.Err()
}

func queryIntentProfiles(tr *tracking.Tracker, window string) ([]IntentProfile, error) {
	rows, err := tr.Query(`
		SELECT command, COALESCE(SUM(original_tokens),0), COALESCE(SUM(saved_tokens),0), COUNT(*)
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY command
	`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type agg struct {
		commands int64
		original int64
		saved    int64
	}
	byIntent := map[string]*agg{}
	for rows.Next() {
		var command string
		var original, saved, count int64
		if err := rows.Scan(&command, &original, &saved, &count); err != nil {
			return nil, err
		}
		intent := inferIntent(command)
		if _, ok := byIntent[intent]; !ok {
			byIntent[intent] = &agg{}
		}
		byIntent[intent].commands += count
		byIntent[intent].original += original
		byIntent[intent].saved += saved
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]IntentProfile, 0, len(byIntent))
	for intent, a := range byIntent {
		item := IntentProfile{
			Intent:   intent,
			Commands: a.commands,
			Original: a.original,
			Saved:    a.saved,
		}
		if a.original > 0 {
			item.ReductionPct = (float64(a.saved) / float64(a.original)) * 100
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Original > out[j].Original
	})
	return out, nil
}

func queryAgentBudgets(tr *tracking.Tracker, window string, totalOriginal int64) ([]AgentBudget, error) {
	rows, err := tr.Query(`
		SELECT COALESCE(NULLIF(agent_name,''),'unknown') as agent,
		       COUNT(*) as commands,
		       COALESCE(SUM(original_tokens),0) as original_tokens,
		       COALESCE(SUM(saved_tokens),0) as saved_tokens
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY agent
		ORDER BY original_tokens DESC
	`, window)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AgentBudget, 0, 8)
	for rows.Next() {
		var a AgentBudget
		if err := rows.Scan(&a.Agent, &a.Commands, &a.Original, &a.Saved); err != nil {
			return nil, err
		}
		if totalOriginal > 0 {
			a.BudgetShare = float64(a.Original) / float64(totalOriginal) * 100
		}
		if a.Original > 0 {
			a.ReductionPct = float64(a.Saved) / float64(a.Original) * 100
		}
		a.EstimatedUS = tokensToUSD(a.Original)
		a.SavingsUS = tokensToUSD(a.Saved)
		out = append(out, a)
	}
	return out, rows.Err()
}

func queryAnchorRetention(tr *tracking.Tracker, window string, totalCommands int64) (AnchorRetentionReport, error) {
	var signals int64
	err := tr.QueryRow(`
		SELECT COALESCE(COUNT(*),0) FROM commands
		WHERE timestamp >= datetime('now', ?)
		  AND (
			command LIKE '%error%' OR command LIKE '%fail%' OR command LIKE '%panic%' OR
			command LIKE '%fix%' OR command LIKE '%todo%' OR command LIKE '%debug%' OR
			command LIKE '%git diff%' OR command LIKE '%git status%'
		  )
	`, window).Scan(&signals)
	if err != nil {
		return AnchorRetentionReport{}, err
	}
	density := 0.0
	if totalCommands > 0 {
		density = float64(signals) / float64(totalCommands) * 100
	}
	keepRate := 100 - (density * 0.4)
	if keepRate < 50 {
		keepRate = 50
	}
	grade := "A"
	switch {
	case keepRate >= 90:
		grade = "A"
	case keepRate >= 80:
		grade = "B"
	case keepRate >= 70:
		grade = "C"
	default:
		grade = "D"
	}
	return AnchorRetentionReport{
		SignalsFound:      signals,
		SignalDensityPct:  density,
		EstimatedKeepRate: keepRate,
		Grade:             grade,
	}, nil
}

func detectWaste(tr *tracking.Tracker, window string, summary Summary, cfg DetectorConfig) ([]Finding, error) {
	findings := make([]Finding, 0, 12)
	for _, d := range DefaultDetectors() {
		if !cfg.Enabled(d.ID()) {
			continue
		}
		items, err := d.Run(tr, window, summary)
		if err != nil {
			return nil, fmt.Errorf("detector %s: %w", d.ID(), err)
		}
		for _, item := range items {
			if !cfg.AllowedSeverity(d.ID(), item.Severity) {
				continue
			}
			findings = append(findings, item)
		}
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

func buildBudgetController(summary Summary, quality QualityReport, intents []IntentProfile) BudgetControllerReport {
	report := BudgetControllerReport{
		CurrentReductionPct: summary.ReductionPct,
		QualityScore:        quality.Score,
		RecommendedMode:     "balanced",
		Bands: map[string]float64{
			"quality_floor":          75,
			"target_reduction":       60,
			"hard_reduction_ceiling": 90,
		},
	}
	switch {
	case quality.Score < 70:
		report.RecommendedMode = "conservative"
	case summary.ReductionPct < 45:
		report.RecommendedMode = "aggressive"
	}
	// LazyLLM-style depth decay schedule for 10 depth bands.
	base := 1000
	report.DecaySchedule = make([]int, 10)
	for i := range report.DecaySchedule {
		factor := 1.0 - (float64(i) * 0.07)
		if report.RecommendedMode == "conservative" {
			factor += 0.08
		}
		if report.RecommendedMode == "aggressive" {
			factor -= 0.08
		}
		if factor < 0.35 {
			factor = 0.35
		}
		report.DecaySchedule[i] = int(float64(base) * factor)
	}
	// Intent weighting (SWE-pruner style): if debug intent dominates, bias to conservative.
	for _, intent := range intents {
		if intent.Intent == "debug" && intent.Original > summary.Original/3 {
			report.RecommendedMode = "conservative"
			break
		}
	}
	return report
}

func inferIntent(command string) string {
	c := strings.ToLower(strings.TrimSpace(command))
	switch {
	case strings.HasPrefix(c, "go test"), strings.HasPrefix(c, "pytest"), strings.HasPrefix(c, "jest"), strings.HasPrefix(c, "vitest"):
		return "test"
	case strings.HasPrefix(c, "go build"), strings.HasPrefix(c, "npm run build"), strings.HasPrefix(c, "make"), strings.HasPrefix(c, "cargo build"):
		return "build"
	case strings.Contains(c, "debug"), strings.Contains(c, "error"), strings.Contains(c, "fail"), strings.Contains(c, "stack"):
		return "debug"
	case strings.HasPrefix(c, "git diff"), strings.HasPrefix(c, "git log"), strings.HasPrefix(c, "git show"):
		return "review"
	default:
		return "general"
	}
}

// SaveSnapshot writes a named drift-validation snapshot to disk.
func SaveSnapshot(dir, name string, report *Report) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = time.Now().UTC().Format("20060102-150405")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	s := Snapshot{
		Name:        safeID(name),
		CreatedAt:   time.Now().UTC(),
		Fingerprint: report.DriftFingerprint,
		Report:      *report,
	}
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
	driftChanged := base.Fingerprint != candidate.Fingerprint

	verdict := "neutral"
	if deltaReduction >= 3 && deltaQuality >= 0 && deltaParseFailures <= 0 {
		verdict = "improved"
	} else if deltaReduction < -2 || deltaQuality < -3 || deltaParseFailures > 0 {
		verdict = "regressed"
	}
	if driftChanged && verdict == "neutral" {
		verdict = "changed"
	}

	return CompareReport{
		BaseName:           base.Name,
		CandidateName:      candidate.Name,
		ComparedAt:         time.Now().UTC(),
		DeltaSavedTokens:   deltaSaved,
		DeltaReductionPct:  deltaReduction,
		DeltaQualityScore:  deltaQuality,
		DeltaParseFailures: deltaParseFailures,
		DriftChanged:       driftChanged,
		Verdict:            verdict,
	}
}

// DriftFingerprint returns a stable fingerprint of the active config.
func DriftFingerprint(configPath string) (string, error) {
	if strings.TrimSpace(configPath) == "" {
		return "", nil
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	normalized := strings.TrimSpace(string(raw))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:]), nil
}

// RenderHTML writes a lightweight dashboard HTML report.
func RenderHTML(path string, report *Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	const tmpl = `<!doctype html>
<html><head><meta charset="utf-8"><title>tok Audit</title>
<style>
body{font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;background:#f7f8fb;color:#111827}
.card{background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:14px;margin-bottom:14px}
table{width:100%;border-collapse:collapse}th,td{padding:8px;border-bottom:1px solid #eee;text-align:left}
.badge{display:inline-block;padding:2px 8px;border-radius:12px;background:#eef2ff}
</style></head><body>
<h1>tok Audit</h1>
<div class="card"><b>Window:</b> {{.Days}} days | <b>Generated:</b> {{.GeneratedAt}}</div>
<div class="card"><h2>Summary</h2>
<p>Commands: {{.Summary.CommandCount}} | Original: {{.Summary.Original}} | Filtered: {{.Summary.Filtered}} | Saved: {{.Summary.Saved}}</p>
<p>Reduction: {{printf "%.2f" .Summary.ReductionPct}}% | Quality: <span class="badge">{{printf "%.1f" .Quality.Score}} ({{.Quality.Band}})</span></p>
<p>Budget Mode: <span class="badge">{{.BudgetController.RecommendedMode}}</span> | Anchor Retention: <span class="badge">{{.AnchorRetention.Grade}}</span></p>
</div>
<div class="card"><h2>Waste Findings</h2>
<table><tr><th>ID</th><th>Severity</th><th>Description</th><th>Waste Tokens</th></tr>
{{range .WasteFindings}}<tr><td>{{.ID}}</td><td>{{.Severity}}</td><td>{{.Description}}</td><td>{{.EstimatedWaste}}</td></tr>{{end}}
</table></div>
<div class="card"><h2>Top Layers</h2>
<table><tr><th>Layer</th><th>Total Saved</th><th>Avg Saved</th><th>Calls</th></tr>
{{range .TopLayers}}<tr><td>{{.LayerName}}</td><td>{{.TotalSaved}}</td><td>{{printf "%.1f" .AvgSaved}}</td><td>{{.CallCount}}</td></tr>{{end}}
</table></div>
<div class="card"><h2>Costly Prompts</h2>
<table><tr><th>Command</th><th>Count</th><th>Original Tokens</th><th>Estimated Cost (USD)</th></tr>
{{range .CostlyPrompts}}<tr><td>{{.Command}}</td><td>{{.Count}}</td><td>{{.Original}}</td><td>{{printf "%.4f" .EstimatedUS}}</td></tr>{{end}}
</table></div>
<div class="card"><h2>Agent Budgets</h2>
<table><tr><th>Agent</th><th>Share %</th><th>Original</th><th>Saved</th><th>Reduction %</th></tr>
{{range .AgentBudgets}}<tr><td>{{.Agent}}</td><td>{{printf "%.1f" .BudgetShare}}</td><td>{{.Original}}</td><td>{{.Saved}}</td><td>{{printf "%.1f" .ReductionPct}}</td></tr>{{end}}
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
