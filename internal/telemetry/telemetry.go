package telemetry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/integrity"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

const (
	TelemetryEndpoint = "https://api.tok.dev/v1/telemetry"
	ConsentFile       = "telemetry_consent"
	LocalEventsFile   = "events.jsonl"
)

// TelemetryData represents the data sent to telemetry server
type TelemetryData struct {
	// Identity (anonymized)
	DeviceHash string `json:"device_hash"`

	// Environment
	Version       string `json:"version"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	InstallMethod string `json:"install_method,omitempty"`

	// Usage volume (24h)
	CommandCount     int `json:"command_count_24h"`
	TotalCommands    int `json:"total_commands"`
	TokensSaved24h   int `json:"tokens_saved_24h"`
	TokensSaved30d   int `json:"tokens_saved_30d"`
	TokensSavedTotal int `json:"tokens_saved_total"`

	// Quality
	ParseFailures   int      `json:"parse_failures"`
	PassthroughCmds []string `json:"passthrough_cmds,omitempty"` // Top 5, no args

	// Ecosystem
	CommandCategories map[string]int `json:"command_categories,omitempty"`

	// Retention
	DaysSinceFirstUse int `json:"days_since_first_use"`
	ActiveDays30d     int `json:"active_days_30d"`

	// Adoption
	AgentHookType     string `json:"agent_hook_type,omitempty"`
	CustomFilterCount int    `json:"custom_filter_count"`

	// Features
	MetaCommandsUsed []string `json:"meta_commands_used,omitempty"`

	// Economics
	EstimatedUSDSaved float64 `json:"estimated_usd_saved"`

	Timestamp time.Time `json:"timestamp"`
}

type LocalEventStats struct {
	TotalEvents     int            `json:"total_events"`
	ByFeature       map[string]int `json:"by_feature,omitempty"`
	ByCategory      map[string]int `json:"by_category,omitempty"`
	ByDay           map[string]int `json:"by_day,omitempty"`
	TopCommands     []string       `json:"top_commands,omitempty"`
	TopTestRunners  []string       `json:"top_test_runners,omitempty"`
	LastEventAt     string         `json:"last_event_at,omitempty"`
	CommandInvoked  int            `json:"command_invocation_events,omitempty"`
	MetaCommands    int            `json:"meta_command_events,omitempty"`
	OperationalCmds int            `json:"operational_command_events,omitempty"`
}

type telemetryTracker interface {
	GetDashboardSnapshot(opts tracking.DashboardQueryOptions) (*tracking.DashboardSnapshot, error)
	GetParseFailureSummary() (*tracking.ParseFailureSummary, error)
	GetSavings(projectPath string) (*tracking.SavingsSummary, error)
}

// ConsentStatus represents the user's telemetry consent
type ConsentStatus int

const (
	ConsentUnknown ConsentStatus = iota
	ConsentEnabled
	ConsentDisabled
)

// IsEnabled returns true if telemetry is enabled
func IsEnabled() bool {
	// Check env override first (disable takes precedence)
	if os.Getenv("TOK_TELEMETRY_DISABLED") == "1" {
		return false
	}

	// Check consent file
	return GetConsent() == ConsentEnabled
}

// GetConsent returns the current consent status
func GetConsent() ConsentStatus {
	consentPath := getConsentPath()
	data, err := os.ReadFile(consentPath)
	if err != nil {
		return ConsentUnknown
	}

	switch string(data) {
	case "enabled":
		return ConsentEnabled
	case "disabled":
		return ConsentDisabled
	default:
		return ConsentUnknown
	}
}

// SetConsent sets the telemetry consent status
func SetConsent(enabled bool) error {
	consentPath := getConsentPath()

	if enabled {
		return os.WriteFile(consentPath, []byte("enabled"), 0644)
	}
	return os.WriteFile(consentPath, []byte("disabled"), 0644)
}

// ForgetConsent removes consent and all local telemetry data
func ForgetConsent() error {
	consentPath := getConsentPath()
	_ = os.Remove(consentPath)
	_ = os.Remove(localEventsPath())
	_ = os.RemoveAll(filepath.Join(config.DataPath(), "telemetry"))
	return nil
}

// Send sends telemetry data (if enabled)
func Send(data *TelemetryData) error {
	if !IsEnabled() {
		return nil
	}

	// Set device hash if not set
	if data.DeviceHash == "" {
		data.DeviceHash = getDeviceHash()
	}

	data.Timestamp = time.Now()

	// Marshal and send
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Send in background
	go func() {
		var buf bytes.Buffer
		if _, err := buf.Write(jsonData); err != nil {
			return
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(TelemetryEndpoint, "application/json", &buf)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}()

	return nil
}

// CollectDaily collects and sends daily telemetry
func CollectDaily(tracker interface{}) error {
	if !IsEnabled() {
		return nil
	}

	tr, ok := tracker.(telemetryTracker)
	if !ok {
		return fmt.Errorf("unsupported tracker for telemetry collection")
	}

	snapshot, err := tr.GetDashboardSnapshot(tracking.DashboardQueryOptions{
		Days:               30,
		Limit:              5,
		ReductionGoalPct:   40,
		DailyTokenBudget:   100_000,
		WeeklyTokenBudget:  500_000,
		MonthlyTokenBudget: 2_000_000,
	})
	if err != nil {
		return err
	}

	parseSummary, err := tr.GetParseFailureSummary()
	if err != nil {
		return err
	}
	savings, err := tr.GetSavings("")
	if err != nil {
		return err
	}

	data := &TelemetryData{
		Version:           shared.Version,
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		CommandCount:      latestCommandCount(snapshot),
		TotalCommands:     savings.TotalCommands,
		TokensSaved24h:    int(snapshot.Budgets.Daily.SavedTokens),
		TokensSaved30d:    int(snapshot.Overview.TotalSavedTokens),
		TokensSavedTotal:  savings.TotalSaved,
		ParseFailures:     int(parseSummary.Total),
		PassthroughCmds:   breakdownKeys(snapshot.LowSavingsCommands, 5),
		CommandCategories: buildCommandCategories(snapshot.TopCommands),
		DaysSinceFirstUse: snapshot.Lifecycle.DaysSinceFirstUse,
		ActiveDays30d:     snapshot.Lifecycle.ActiveDays30d,
		AgentHookType:     detectHookType(),
		CustomFilterCount: countCustomFilters(),
		MetaCommandsUsed:  nil,
		EstimatedUSDSaved: snapshot.Overview.EstimatedSavingsUSD,
		Timestamp:         time.Now(),
	}

	return Send(data)
}

func getConsentPath() string {
	return filepath.Join(config.DataPath(), ConsentFile)
}

func localEventsPath() string {
	return filepath.Join(config.DataPath(), "telemetry", LocalEventsFile)
}

// EventBatcher batches telemetry events for efficient sending
type EventBatcher struct {
	events     []map[string]interface{}
	mu         sync.Mutex
	batchSize  int
	flushTimer *time.Timer
	flushAfter time.Duration
}

// Global event batcher instance
var globalBatcher = &EventBatcher{
	events:     make([]map[string]interface{}, 0, 100),
	batchSize:  10,
	flushAfter: 30 * time.Second,
}

// AddEvent adds an event to the batch
func (b *EventBatcher) AddEvent(event map[string]interface{}) {
	if !IsEnabled() {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, event)

	// Flush if batch is full
	if len(b.events) >= b.batchSize {
		b.flush()
		return
	}

	// Start or reset flush timer
	if b.flushTimer != nil {
		b.flushTimer.Stop()
	}
	b.flushTimer = time.AfterFunc(b.flushAfter, func() {
		b.Flush()
	})
}

// Flush sends all batched events
func (b *EventBatcher) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.events) == 0 {
		return nil
	}

	return b.flush()
}

// flush sends events without locking (must be called with lock held)
func (b *EventBatcher) flush() error {
	if len(b.events) == 0 {
		return nil
	}

	// Copy events and clear batch
	eventsToSend := make([]map[string]interface{}, len(b.events))
	copy(eventsToSend, b.events)
	b.events = b.events[:0]

	// Stop timer if running
	if b.flushTimer != nil {
		b.flushTimer.Stop()
		b.flushTimer = nil
	}

	// Send asynchronously
	go func() {
		payload := map[string]interface{}{
			"type":      "batch",
			"count":     len(eventsToSend),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"events":    eventsToSend,
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			return
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(TelemetryEndpoint+"/batch", "application/json", &buf)
		if err != nil {
			// Silently fail - telemetry should never block
			return
		}
		defer resp.Body.Close()
	}()

	return nil
}

// GetBatcher returns the global event batcher
func GetBatcher() *EventBatcher {
	return globalBatcher
}

// FlushBatcher flushes all pending telemetry events
func FlushBatcher() error {
	return globalBatcher.Flush()
}

// TrackFeatureUsage tracks usage of specific features (test-runner, quota, etc.)
func TrackFeatureUsage(featureName string, details map[string]interface{}) error {
	if !IsEnabled() {
		return nil
	}

	data := map[string]interface{}{
		"type":      "feature_usage",
		"feature":   featureName,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   shared.Version,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}

	// Merge details
	for k, v := range details {
		data[k] = v
	}

	if err := appendLocalEvent(data); err != nil {
		return err
	}

	// Add to batch instead of sending immediately
	globalBatcher.AddEvent(data)

	return nil
}

// TrackTestRunnerUsage tracks test-runner command usage
func TrackTestRunnerUsage(runnerType string, autoDetected bool) {
	TrackFeatureUsage("test_runner", map[string]interface{}{
		"runner_type":   runnerType,
		"auto_detected": autoDetected,
	})
}

// TrackQuotaUsage tracks gain --quota usage
func TrackQuotaUsage(tier string, usagePct float64) {
	TrackFeatureUsage("quota_check", map[string]interface{}{
		"tier":      tier,
		"usage_pct": usagePct,
	})
}

// TrackRewrite tracks command rewrite events
func TrackRewrite(originalCmd, rewrittenCmd string, testRunner bool) {
	TrackFeatureUsage("command_rewrite", map[string]interface{}{
		"original":    originalCmd,
		"rewritten":   rewrittenCmd,
		"test_runner": testRunner,
	})
}

func getDeviceHash() string {
	// Create a stable, anonymized device identifier
	// Uses hostname + user ID hash
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}

	hash := sha256.Sum256([]byte(hostname + user))
	return hex.EncodeToString(hash[:16]) // First 16 bytes = 32 hex chars
}

// ShowConsentPrompt displays the interactive consent prompt
func ShowConsentPrompt() {
	fmt.Println()
	fmt.Println("📊 tok Telemetry")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Help improve tok by sharing anonymous usage data:")
	fmt.Println()
	fmt.Println("What we collect:")
	fmt.Println("  • Command counts and token savings (aggregate)")
	fmt.Println("  • Which commands are used most (without arguments)")
	fmt.Println("  • Error rates and parse failures")
	fmt.Println("  • Basic environment (OS, version)")
	fmt.Println()
	fmt.Println("What we DON'T collect:")
	fmt.Println("  • Source code or file contents")
	fmt.Println("  • Command arguments or paths")
	fmt.Println("  • Personal information")
	fmt.Println()
	fmt.Print("Enable telemetry? [Y/n]: ")

	var response string
	fmt.Scanln(&response)

	if response == "" || strings.ToLower(response) == "y" {
		SetConsent(true)
		fmt.Println("✓ Telemetry enabled. Thank you!")
	} else {
		SetConsent(false)
		fmt.Println("✓ Telemetry disabled.")
	}
}

func breakdownKeys(items []tracking.DashboardBreakdown, limit int) []string {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Key)
	}
	return out
}

func buildCommandCategories(items []tracking.DashboardBreakdown) map[string]int {
	if len(items) == 0 {
		return nil
	}
	categories := make(map[string]int)
	for _, item := range items {
		base := strings.Fields(item.Key)
		if len(base) == 0 {
			continue
		}
		categories[base[0]] += int(item.Commands)
	}
	return categories
}

func detectHookType() string {
	result, err := integrity.VerifyHook()
	if err == nil && (result.Status == integrity.StatusVerified || result.Status == integrity.StatusOutdated || result.Status == integrity.StatusNoBaseline) {
		return "claude"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "unknown"
	}
	checks := []struct {
		name string
		path string
	}{
		{name: "cursor", path: filepath.Join(home, ".cursor", "hooks", integrity.HookFilename)},
		{name: "windsurf", path: filepath.Join(home, ".windsurf", "hooks", integrity.HookFilename)},
		{name: "gemini", path: filepath.Join(home, ".gemini", "hooks", integrity.HookFilename)},
		{name: "codex", path: filepath.Join(home, ".codex", "hooks", integrity.HookFilename)},
	}
	for _, check := range checks {
		if _, err := os.Stat(check.path); err == nil {
			return check.name
		}
	}
	return "unknown"
}

func countCustomFilters() int {
	filterDir := filepath.Join(config.ConfigDir(), "filters")
	count := 0
	_ = filepath.WalkDir(filterDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".toml") {
			count++
		}
		return nil
	})
	return count
}

func BuildExportSummary(tracker telemetryTracker) map[string]interface{} {
	summary := map[string]interface{}{
		"commands_tracked":   0,
		"tokens_saved_total": 0,
	}

	snapshot, err := tracker.GetDashboardSnapshot(tracking.DashboardQueryOptions{
		Days:               30,
		Limit:              5,
		ReductionGoalPct:   40,
		DailyTokenBudget:   100_000,
		WeeklyTokenBudget:  500_000,
		MonthlyTokenBudget: 2_000_000,
	})
	if err == nil {
		summary["commands_tracked"] = snapshot.Lifecycle.CommandsTotal
		summary["tokens_saved_total"] = snapshot.Overview.TotalSavedTokens
		summary["estimated_usd_saved"] = snapshot.Overview.EstimatedSavingsUSD
		summary["active_days_30d"] = snapshot.Lifecycle.ActiveDays30d
		summary["projects_tracked"] = snapshot.Lifecycle.ProjectsCount
		summary["top_provider_models"] = breakdownKeys(snapshot.TopProviderModels, 5)
		summary["low_savings_commands"] = breakdownKeys(snapshot.LowSavingsCommands, 5)
	}
	if savings, err := tracker.GetSavings(""); err == nil {
		summary["commands_total_lifetime"] = savings.TotalCommands
		summary["tokens_saved_lifetime"] = savings.TotalSaved
	}
	if stats, err := GetLocalEventStats(); err == nil {
		summary["telemetry_events_total"] = stats.TotalEvents
		summary["telemetry_by_feature"] = stats.ByFeature
		summary["telemetry_by_category"] = stats.ByCategory
		summary["telemetry_daily_counts"] = stats.ByDay
		summary["telemetry_top_commands"] = stats.TopCommands
		summary["telemetry_top_test_runners"] = stats.TopTestRunners
		summary["telemetry_last_event_at"] = stats.LastEventAt
	}

	parseSummary, err := tracker.GetParseFailureSummary()
	if err == nil {
		summary["parse_failures_total"] = parseSummary.Total
		summary["parse_recovery_rate_pct"] = parseSummary.RecoveryRate
	}

	keys := make([]string, 0, len(summary))
	for key := range summary {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	summary["keys"] = keys

	return summary
}

func latestCommandCount(snapshot *tracking.DashboardSnapshot) int {
	if snapshot == nil || len(snapshot.DailyTrends) == 0 {
		return 0
	}
	return int(snapshot.DailyTrends[len(snapshot.DailyTrends)-1].Commands)
}

func appendLocalEvent(event map[string]interface{}) error {
	path := localEventsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return nil
}

func RecentLocalEvents(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}
	content, err := os.ReadFile(localEventsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	events := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			events = append(events, event)
		}
	}
	return events, nil
}

func GetLocalEventStats() (*LocalEventStats, error) {
	events, err := RecentLocalEvents(10_000)
	if err != nil {
		return nil, err
	}
	stats := &LocalEventStats{
		ByFeature:  make(map[string]int),
		ByCategory: make(map[string]int),
		ByDay:      make(map[string]int),
	}
	commandCounts := make(map[string]int)
	testRunnerCounts := make(map[string]int)
	for _, event := range events {
		stats.TotalEvents++
		if ts, _ := event["timestamp"].(string); ts != "" {
			if len(ts) >= 10 {
				stats.ByDay[ts[:10]]++
			}
			if ts > stats.LastEventAt {
				stats.LastEventAt = ts
			}
		}
		if feature, _ := event["feature"].(string); feature != "" {
			stats.ByFeature[feature]++
		}
		if category, _ := event["category"].(string); category != "" {
			stats.ByCategory[category]++
			if category == "meta" {
				stats.MetaCommands++
			}
			if category == "operational" {
				stats.OperationalCmds++
			}
		}
		if commandPath, _ := event["command_path"].(string); commandPath != "" {
			commandCounts[commandPath]++
			stats.CommandInvoked++
		}
		if runnerType, _ := event["runner_type"].(string); runnerType != "" {
			testRunnerCounts[runnerType]++
		}
	}
	stats.TopCommands = topKeys(commandCounts, 5)
	stats.TopTestRunners = topKeys(testRunnerCounts, 5)
	return stats, nil
}

func topKeys(counts map[string]int, limit int) []string {
	if len(counts) == 0 || limit <= 0 {
		return nil
	}
	type pair struct {
		key   string
		count int
	}
	items := make([]pair, 0, len(counts))
	for key, count := range counts {
		items = append(items, pair{key: key, count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].key < items[j].key
		}
		return items[i].count > items[j].count
	})
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%s (%d)", item.key, item.count))
	}
	return out
}
