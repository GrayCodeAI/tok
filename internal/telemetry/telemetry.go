package telemetry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/config"
)

const (
	TelemetryEndpoint = "https://api.tokman.dev/v1/telemetry"
	ConsentFile       = "telemetry_consent"
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
	CommandCount      int `json:"command_count_24h"`
	TotalCommands     int `json:"total_commands"`
	TokensSaved24h    int `json:"tokens_saved_24h"`
	TokensSaved30d    int `json:"tokens_saved_30d"`
	TokensSavedTotal  int `json:"tokens_saved_total"`

	// Quality
	ParseFailures      int      `json:"parse_failures"`
	PassthroughCmds    []string `json:"passthrough_cmds,omitempty"` // Top 5, no args

	// Ecosystem
	CommandCategories map[string]int `json:"command_categories,omitempty"`

	// Retention
	DaysSinceFirstUse int `json:"days_since_first_use"`
	ActiveDays30d     int `json:"active_days_30d"`

	// Adoption
	AgentHookType    string `json:"agent_hook_type,omitempty"`
	CustomFilterCount int   `json:"custom_filter_count"`

	// Features
	MetaCommandsUsed []string `json:"meta_commands_used,omitempty"`

	// Economics
	EstimatedUSDSaved float64 `json:"estimated_usd_saved"`

	Timestamp time.Time `json:"timestamp"`
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
	if os.Getenv("TOKMAN_TELEMETRY_DISABLED") == "1" {
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
	os.Remove(consentPath)
	// Note: Server-side deletion would require additional API call
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
		if err := json.NewEncoder(&buf).Encode(jsonData); err != nil {
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

	data := &TelemetryData{
		Version:   shared.Version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now(),
	}

	// TODO: Fill in actual data from tracker
	// This would require tracker interface methods

	return Send(data)
}

func getConsentPath() string {
	return filepath.Join(config.DataPath(), ConsentFile)
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
		"tier":        tier,
		"usage_pct":   usagePct,
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
	fmt.Println("📊 TokMan Telemetry")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Help improve TokMan by sharing anonymous usage data:")
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
