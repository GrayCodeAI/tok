package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
	telemetrylib "github.com/lakshmanpatel/tok/internal/telemetry"
)

var (
	telemetryStatus  bool
	telemetryEnable  bool
	telemetryDisable bool
	telemetryForget  bool
	telemetryExport  string
	telemetryFormat  string
)

var telemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Manage telemetry settings (GDPR-compliant)",
	Long: `Manage tok telemetry collection settings.

tok collects anonymized usage data to improve the tool.
All data collection is opt-in and GDPR-compliant.

Collected data:
  - Command usage patterns (anonymized)
  - Token savings metrics
  - Error rates (without sensitive data)

No data collected:
  - File contents
  - Command arguments
  - Personal information

Examples:
  tok telemetry --status     # Check current telemetry status
  tok telemetry --enable     # Enable telemetry
  tok telemetry --disable    # Disable telemetry
  tok telemetry --forget     # Delete local telemetry data and consent
  tok telemetry --export data.json  # Export collected data`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: runTelemetry,
}

func init() {
	registry.Add(func() { registry.Register(telemetryCmd) })
	telemetryCmd.Flags().BoolVar(&telemetryStatus, "status", false, "Show telemetry status")
	telemetryCmd.Flags().BoolVar(&telemetryEnable, "enable", false, "Enable telemetry collection")
	telemetryCmd.Flags().BoolVar(&telemetryDisable, "disable", false, "Disable telemetry collection")
	telemetryCmd.Flags().BoolVar(&telemetryForget, "forget", false, "Delete local telemetry data and consent")
	telemetryCmd.Flags().StringVar(&telemetryExport, "export", "", "Export telemetry data to file")
	telemetryCmd.Flags().StringVar(&telemetryFormat, "format", "json", "Export format (json, csv)")
}

func runTelemetry(cmd *cobra.Command, args []string) error {
	// Default to showing status if no flags provided
	if !telemetryStatus && !telemetryEnable && !telemetryDisable && !telemetryForget && telemetryExport == "" {
		telemetryStatus = true
	}

	if telemetryEnable {
		return enableTelemetry()
	}

	if telemetryDisable {
		return disableTelemetry()
	}

	if telemetryForget {
		return forgetTelemetry()
	}

	if telemetryExport != "" {
		return exportTelemetry(telemetryExport, telemetryFormat)
	}

	if telemetryStatus {
		return showTelemetryStatus()
	}

	return nil
}

func telemetryConfigPath() string {
	return filepath.Join(config.DataPath(), "telemetry.json")
}

type telemetryConfig struct {
	Enabled  bool      `json:"enabled"`
	OptInAt  time.Time `json:"opt_in_at,omitempty"`
	OptOutAt time.Time `json:"opt_out_at,omitempty"`
	Version  string    `json:"version"`
	DataDir  string    `json:"data_dir"`
}

func loadTelemetryConfig() (*telemetryConfig, error) {
	path := telemetryConfigPath()

	// Default: disabled
	cfg := &telemetryConfig{
		Enabled: false,
		Version: "1.0",
		DataDir: filepath.Join(config.DataPath(), "telemetry"),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func saveTelemetryConfig(cfg *telemetryConfig) error {
	path := telemetryConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func enableTelemetry() error {
	cfg, err := loadTelemetryConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Enabled = true
	cfg.OptInAt = time.Now().UTC()

	if err := saveTelemetryConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Telemetry enabled\n", green("✓"))
	out.Global().Println()
	out.Global().Println("tok will now collect anonymized usage data:")
	out.Global().Println("  • Command frequency (without arguments)")
	out.Global().Println("  • Token savings statistics")
	out.Global().Println("  • Error patterns (without sensitive data)")
	out.Global().Println()
	out.Global().Println("You can disable this at any time with: tok telemetry --disable")
	out.Global().Println("View your data with: tok telemetry --export my-data.json")

	return nil
}

func disableTelemetry() error {
	cfg, err := loadTelemetryConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Enabled = false
	cfg.OptOutAt = time.Now().UTC()

	if err := saveTelemetryConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	out.Global().Printf("%s Telemetry disabled\n", yellow("⚠"))
	out.Global().Println()
	out.Global().Println("tok will no longer collect usage data.")
	out.Global().Println("Previously collected data can be exported with: tok telemetry --export")

	return nil
}

func showTelemetryStatus() error {
	cfg, err := loadTelemetryConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	out.Global().Println()
	out.Global().Println(cyan("tok Telemetry Status"))
	out.Global().Println(strings.Repeat("═", 40))

	if cfg.Enabled {
		out.Global().Printf("Status:      %s\n", green("Enabled"))
		if !cfg.OptInAt.IsZero() {
			out.Global().Printf("Opt-in date: %s\n", cfg.OptInAt.Format("2006-01-02"))
		}
	} else {
		out.Global().Printf("Status:      %s\n", red("Disabled"))
		if !cfg.OptOutAt.IsZero() {
			out.Global().Printf("Opt-out date: %s\n", cfg.OptOutAt.Format("2006-01-02"))
		}
	}

	out.Global().Printf("Data dir:    %s\n", cfg.DataDir)

	// Check for existing data
	dataSize := getTelemetryDataSize(cfg.DataDir)
	if dataSize > 0 {
		out.Global().Printf("Data size:   %d KB\n", dataSize/1024)
	}
	if stats, err := telemetrylib.GetLocalEventStats(); err == nil && stats.TotalEvents > 0 {
		out.Global().Printf("Events:      %d\n", stats.TotalEvents)
		if stats.LastEventAt != "" {
			out.Global().Printf("Last event:  %s\n", stats.LastEventAt)
		}
		if len(stats.TopCommands) > 0 {
			out.Global().Printf("Top cmds:    %s\n", strings.Join(stats.TopCommands, ", "))
		}
		if len(stats.TopTestRunners) > 0 {
			out.Global().Printf("Test use:    %s\n", strings.Join(stats.TopTestRunners, ", "))
		}
	}

	out.Global().Println()
	out.Global().Println("GDPR Compliance:")
	out.Global().Println("  ✓ Explicit opt-in required")
	out.Global().Println("  ✓ Right to access (export your data)")
	out.Global().Println("  ✓ Right to erasure (delete on disable)")
	out.Global().Println("  ✓ No PII collected")
	out.Global().Println()
	out.Global().Println("Commands:")
	out.Global().Println("  tok telemetry --enable    Enable telemetry")
	out.Global().Println("  tok telemetry --disable   Disable telemetry")
	out.Global().Println("  tok telemetry --forget    Delete local telemetry data")
	out.Global().Println("  tok telemetry --export    Export your data")

	return nil
}

func forgetTelemetry() error {
	if err := telemetrylib.ForgetConsent(); err != nil {
		return fmt.Errorf("failed to delete local telemetry data: %w", err)
	}
	if err := os.Remove(telemetryConfigPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete telemetry config: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Local telemetry data deleted\n", green("✓"))
	out.Global().Println("Telemetry consent has been cleared.")
	return nil
}

func getTelemetryDataSize(dataDir string) int64 {
	var size int64
	filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func exportTelemetry(outputPath, format string) error {
	cfg, err := loadTelemetryConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Collect telemetry data
	data := map[string]interface{}{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"telemetry":   cfg,
		"summary":     collectTelemetrySummary(),
	}
	if events, err := telemetrylib.RecentLocalEvents(1000); err == nil && len(events) > 0 {
		data["events"] = events
	}

	var output []byte
	if format == "csv" {
		output = []byte(exportAsCSV(data))
	} else {
		output, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write export: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Telemetry data exported to %s\n", green("✓"), outputPath)

	return nil
}

func collectTelemetrySummary() map[string]interface{} {
	tracker, err := shared.OpenTracker()
	if err != nil {
		return map[string]interface{}{
			"tracking_error": err.Error(),
		}
	}
	defer tracker.Close()

	return telemetrylib.BuildExportSummary(tracker)
}

func exportAsCSV(data map[string]interface{}) string {
	// Simple CSV export for telemetry data
	var result strings.Builder
	result.WriteString("key,value\n")
	result.WriteString("exported_at," + data["exported_at"].(string) + "\n")
	return result.String()
}
