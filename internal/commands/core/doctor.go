package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/integrity"
	"github.com/GrayCodeAI/tok/internal/session"
	"github.com/GrayCodeAI/tok/internal/telemetry"
	"github.com/GrayCodeAI/tok/internal/tracking"
	"github.com/GrayCodeAI/tok/internal/utils"
)

var (
	doctorFix      bool
	doctorSecurity bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose tok setup issues",
	Long: `Check system configuration, shell hooks, database connectivity,
tokenizer availability, and common setup problems.

Use --security to skip all other checks and print a per-hook integrity
report: for every tok-installed agent hook, show whether its SHA-256
baseline matches the script on disk. Exits non-zero if any hook is
tampered or missing its baseline.`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "attempt to fix detected issues")
	doctorCmd.Flags().BoolVar(&doctorSecurity, "security", false, "per-hook integrity audit only; exits non-zero on tamper")
	registry.Add(func() { registry.Register(doctorCmd) })
}

type checkResult struct {
	Name    string
	Status  string // "ok", "warn", "error"
	Message string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	if doctorSecurity {
		return runDoctorSecurity()
	}

	out.Global().Println("tok doctor — diagnosing setup")
	out.Global().Println("================================")

	results := collectDoctorResults()
	if doctorFix {
		if err := applyDoctorFixes(results); err != nil {
			return err
		}
		results = collectDoctorResults()
	}

	// Print results
	hasError := false
	for _, r := range results {
		icon := "✓"
		switch r.Status {
		case "warn":
			icon = "⚠"
		case "error":
			icon = "✗"
			hasError = true
		}
		out.Global().Printf("  %s %s: %s\n", icon, r.Name, r.Message)
	}

	out.Global().Println()
	if hasError {
		out.Global().Println("Some checks failed. See messages above for fixes.")
		return fmt.Errorf("doctor check failed")
	}
	out.Global().Println("All checks passed!")
	return nil
}

func collectDoctorResults() []checkResult {
	results := []checkResult{
		checkBinary(),
		checkConfigDir(),
		checkDatabase(),
		checkShellHook(),
		checkPath(),
		checkPlatform(),
		checkTokenizer(),
		checkTOMLFilters(),
		checkDiskSpace(),
		checkGoVersion(),
		checkTierSystem(),
	}
	results = append(results, checkSessionStore())
	results = append(results, checkTelemetryStore())
	results = append(results, checkIntegrityBaselines())
	results = append(results, checkDashboardDataQuality())
	results = append(results, checkAgentIntegrations()...)
	return results
}

func applyDoctorFixes(results []checkResult) error {
	configCreated := false
	for _, result := range results {
		if result.Name == "Config Dir" && result.Status == "warn" {
			if err := createDefaulttokConfig(); err != nil {
				return fmt.Errorf("doctor --fix: create config: %w", err)
			}
			configCreated = true
			break
		}
	}
	repairs, err := repairAgentIntegrations()
	if err != nil {
		return fmt.Errorf("doctor --fix: repair integrations: %w", err)
	}
	if !configCreated && repairs == 0 {
		return nil
	}
	return nil
}

func checkBinary() checkResult {
	exe, err := os.Executable()
	if err != nil {
		return checkResult{"Binary", "error", "cannot determine executable path"}
	}
	return checkResult{"Binary", "ok", exe}
}

func checkConfigDir() checkResult {
	configDir := shared.GetConfigDir()
	if configDir == "" {
		return checkResult{"Config Dir", "error", "cannot determine config directory"}
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return checkResult{"Config Dir", "warn", configDir + " (not found — run 'tok init')"}
	}
	return checkResult{"Config Dir", "ok", configDir}
}

func checkDatabase() checkResult {
	dbPath := shared.GetDatabasePath()
	if dbPath == "" {
		return checkResult{"Database", "error", "cannot determine database path"}
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return checkResult{"Database", "warn", dbPath + " (will be created on first use)"}
	}
	tracker, err := shared.OpenTracker()
	if err != nil {
		return checkResult{"Database", "error", dbPath + " (" + err.Error() + ")"}
	}
	defer tracker.Close()
	return checkResult{"Database", "ok", dbPath}
}

func checkShellHook() checkResult {
	agents, err := currentAgentInfos(false)
	if err == nil {
		configured := 0
		partial := 0
		for _, agent := range agents {
			status, _ := describeAgentStatus(agent)
			switch status {
			case "configured":
				configured++
			case "partial", "broken", "legacy":
				partial++
			}
		}
		if configured > 0 {
			message := fmt.Sprintf("%d integration(s) configured", configured)
			if partial > 0 {
				message += fmt.Sprintf(", %d need attention", partial)
			}
			return checkResult{"Agent Hooks", "ok", message}
		}
		if partial > 0 {
			return checkResult{"Agent Hooks", "warn", fmt.Sprintf("%d integration(s) need attention", partial)}
		}
	}

	for _, p := range doctorHookPaths() {
		if _, err := os.Stat(p); err == nil {
			return checkResult{"Agent Hooks", "ok", p}
		}
	}

	return checkResult{"Agent Hooks", "warn", "no agent integrations found — run 'tok init'"}
}

func doctorHookPaths() []string {
	var hookPaths []string
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		hookPaths = append(hookPaths,
			filepath.Join(home, ".claude", "hooks", "tok-rewrite.sh"),
			filepath.Join(home, ".claude", "hooks", "tok.sh"),
		)
	}

	hooksDir := shared.GetHooksPath()
	if hooksDir != "" {
		hookPaths = append(hookPaths, filepath.Join(hooksDir, "tok-rewrite.sh"))
	}

	configDir := shared.GetConfigDir()
	if configDir != "" {
		hookPaths = append(hookPaths, filepath.Join(configDir, "hook.sh"))
	}

	return hookPaths
}

func checkPath() checkResult {
	path, err := exec.LookPath("tok")
	if err != nil {
		return checkResult{"PATH", "warn", "tok not in PATH (may need symlink or PATH update)"}
	}
	return checkResult{"PATH", "ok", path}
}

func checkPlatform() checkResult {
	return checkResult{"Platform", "ok", runtime.GOOS + "/" + runtime.GOARCH + " Go " + runtime.Version()}
}

func checkTokenizer() checkResult {
	// Try to use tiktoken to verify tokenizer is available
	_, err := exec.LookPath("tiktoken")
	if err != nil {
		// tiktoken is embedded in Go binary, so this is OK
		return checkResult{"Tokenizer", "ok", "tiktoken-go (embedded)"}
	}
	return checkResult{"Tokenizer", "ok", "tiktoken available"}
}

func checkTOMLFilters() checkResult {
	srcDir := utils.GetTokSourceDir()
	if srcDir == "" {
		// Installed binary with embedded filters - still functional
		return checkResult{"TOML Filters", "ok", "embedded (installed binary)"}
	}
	builtinDir := filepath.Join(srcDir, "internal", "toml", "builtin")
	if entries, err := os.ReadDir(builtinDir); err == nil {
		count := 0
		for _, e := range entries {
			if !e.IsDir() {
				count++
			}
		}
		return checkResult{"TOML Filters", "ok", fmt.Sprintf("%d built-in filters", count)}
	}
	return checkResult{"TOML Filters", "warn", "built-in filters directory not found"}
}

func checkDiskSpace() checkResult {
	dbPath := shared.GetDatabasePath()
	if dbPath == "" {
		return checkResult{"Disk Space", "warn", "cannot determine database path"}
	}
	if info, err := os.Stat(dbPath); err == nil {
		sizeMB := float64(info.Size()) / 1024 / 1024
		if sizeMB > 100 {
			return checkResult{"Disk Space", "warn", fmt.Sprintf("database is %.1fMB — consider 'tok clean'", sizeMB)}
		}
		return checkResult{"Disk Space", "ok", fmt.Sprintf("database is %.1fMB", sizeMB)}
	}
	return checkResult{"Disk Space", "ok", "no database yet"}
}

func checkGoVersion() checkResult {
	// Check if Go is available for development
	if _, err := exec.LookPath("go"); err == nil {
		return checkResult{"Go", "ok", "available (for development)"}
	}
	return checkResult{"Go", "ok", "not required (prebuilt binary)"}
}

func checkTierSystem() checkResult {
	// Verify the filter system is working
	// Test with sample content
	testInput := "func main() { out.Global().Println(\"hello\") }"
	output, saved := filter.QuickProcessPreset(testInput, filter.ModeMinimal, filter.PresetFast)

	if output != "" && saved >= 0 {
		return checkResult{"Filter System", "ok", "pipeline compression working"}
	}
	return checkResult{"Filter System", "warn", "pipeline may not be compressing"}
}

func checkSessionStore() checkResult {
	manager, err := session.NewSessionManager()
	if err != nil {
		return checkResult{"Sessions", "warn", "session store unavailable"}
	}
	defer manager.Close()

	summary, err := manager.GetSummary()
	if err != nil {
		return checkResult{"Sessions", "warn", "session summary unavailable"}
	}
	if summary.TotalSessions == 0 {
		return checkResult{"Sessions", "ok", "no stored sessions yet"}
	}

	msg := fmt.Sprintf("%d sessions, %d active, %d snapshots", summary.TotalSessions, summary.ActiveSessions, summary.SnapshotCount)
	if summary.TopAgent != "" {
		msg += fmt.Sprintf(", top agent %s", summary.TopAgent)
	}
	return checkResult{"Sessions", "ok", msg}
}

func checkTelemetryStore() checkResult {
	stats, err := telemetry.GetLocalEventStats()
	if err != nil {
		return checkResult{"Telemetry Store", "error", err.Error()}
	}

	switch telemetry.GetConsent() {
	case telemetry.ConsentEnabled:
		if stats.TotalEvents == 0 {
			return checkResult{"Telemetry Store", "ok", "enabled, no local events yet"}
		}
		message := fmt.Sprintf("enabled, %d local events", stats.TotalEvents)
		if stats.LastEventAt != "" {
			message += ", last event " + stats.LastEventAt
		}
		return checkResult{"Telemetry Store", "ok", message}
	case telemetry.ConsentDisabled:
		return checkResult{"Telemetry Store", "ok", "disabled by user"}
	default:
		if stats.TotalEvents > 0 {
			return checkResult{"Telemetry Store", "warn", fmt.Sprintf("%d local events found without explicit consent state", stats.TotalEvents)}
		}
		return checkResult{"Telemetry Store", "ok", "not configured"}
	}
}

func checkIntegrityBaselines() checkResult {
	agents, err := currentAgentInfos(false)
	if err != nil {
		return checkResult{"Integrity Baselines", "warn", err.Error()}
	}

	verified := 0
	warnings := 0
	errors := 0
	for _, agent := range agents {
		if agent.HookDir == "" {
			continue
		}
		hookPath := filepath.Join(agent.HookDir, "tok-rewrite.sh")
		if _, err := os.Stat(hookPath); err != nil {
			continue
		}
		result, err := integrity.VerifyHookAt(hookPath)
		if err != nil {
			errors++
			continue
		}
		switch result.Status {
		case integrity.StatusVerified:
			verified++
		case integrity.StatusOutdated, integrity.StatusNoBaseline:
			warnings++
		case integrity.StatusTampered:
			errors++
		}
	}

	switch {
	case errors > 0:
		return checkResult{"Integrity Baselines", "error", fmt.Sprintf("%d managed hook(s) failed integrity", errors)}
	case warnings > 0:
		return checkResult{"Integrity Baselines", "warn", fmt.Sprintf("%d managed hook(s) need baseline refresh", warnings)}
	case verified > 0:
		return checkResult{"Integrity Baselines", "ok", fmt.Sprintf("%d managed hook(s) verified", verified)}
	default:
		return checkResult{"Integrity Baselines", "ok", "no managed hook scripts present"}
	}
}

func checkDashboardDataQuality() checkResult {
	tracker, err := shared.OpenTracker()
	if err != nil {
		return checkResult{"Dashboard Data", "warn", "tracking database unavailable"}
	}
	defer tracker.Close()

	quality, err := tracker.GetDashboardDataQuality(tracking.DashboardQueryOptions{Days: 30})
	if err != nil {
		return checkResult{"Dashboard Data", "error", err.Error()}
	}
	if quality.TotalCommands == 0 {
		return checkResult{"Dashboard Data", "ok", "no tracked commands yet"}
	}

	var issues []string
	if quality.CommandsMissingAgent > 0 {
		issues = append(issues, fmt.Sprintf("%d missing agent", quality.CommandsMissingAgent))
	}
	if quality.CommandsMissingProvider > 0 {
		issues = append(issues, fmt.Sprintf("%d missing provider", quality.CommandsMissingProvider))
	}
	if quality.CommandsMissingModel > 0 {
		issues = append(issues, fmt.Sprintf("%d missing model", quality.CommandsMissingModel))
	}
	if quality.CommandsMissingSession > 0 {
		issues = append(issues, fmt.Sprintf("%d missing session", quality.CommandsMissingSession))
	}
	if quality.PricingCoverage.FallbackPricingCommands > 0 {
		issues = append(issues, fmt.Sprintf("%d fallback-priced", quality.PricingCoverage.FallbackPricingCommands))
	}

	message := fmt.Sprintf(
		"%d commands, %.1f%% pricing coverage",
		quality.TotalCommands,
		quality.PricingCoverage.CoveragePct(),
	)
	if len(issues) > 0 {
		message += " (" + strings.Join(issues, ", ") + ")"
		return checkResult{"Dashboard Data", "warn", message}
	}
	return checkResult{"Dashboard Data", "ok", message}
}

func checkAgentIntegrations() []checkResult {
	agents, err := currentAgentInfos(false)
	if err != nil {
		return []checkResult{{Name: "Agent Integrations", Status: "warn", Message: err.Error()}}
	}

	results := make([]checkResult, 0, len(agents))
	for _, agent := range agents {
		status, detail := describeAgentStatus(agent)
		message := detail
		if message == "" {
			message = status
		}
		switch status {
		case "configured":
			results = append(results, checkResult{Name: agent.Name, Status: "ok", Message: message})
		case "detected":
			results = append(results, checkResult{Name: agent.Name, Status: "warn", Message: message})
		case "partial", "broken", "legacy":
			results = append(results, checkResult{Name: agent.Name, Status: "warn", Message: message})
		default:
			results = append(results, checkResult{Name: agent.Name, Status: "ok", Message: "not detected"})
		}
	}
	return results
}

func repairAgentIntegrations() (int, error) {
	agents, err := currentAgentInfos(false)
	if err != nil {
		return 0, err
	}

	repaired := 0
	for _, agent := range agents {
		status, detail := describeAgentStatus(agent)
		if !shouldRepairAgentIntegration(status, detail) {
			continue
		}
		if err := setupAgent(agent, installUsesGlobal(agent, false)); err != nil {
			return repaired, fmt.Errorf("%s: %w", agent.Name, err)
		}
		repaired++
	}
	return repaired, nil
}

func shouldRepairAgentIntegration(status, detail string) bool {
	switch status {
	case "partial", "broken", "legacy":
		return true
	case "configured":
		return strings.Contains(detail, "outdated hook") || strings.Contains(detail, "no integrity baseline")
	default:
		return false
	}
}

// runDoctorSecurity prints a per-hook integrity audit across every agent
// tok has installed against. Each agent's hook is shown with its path,
// current integrity status, and on tamper the expected + actual SHA
// prefix so the user can correlate with a diff.
//
// Exits non-zero when any hook is tampered or has no baseline — the
// latter because a hook without a recorded baseline is also a hook we
// cannot prove is legitimate. Verified and NotInstalled are OK.
//
// Scope: this command audits, it does not block runtime. Blocking
// runtime would require every installed hook script to call `tok`'s
// RuntimeCheck before executing, which today only the Claude Code hook
// does. Wiring RuntimeCheck into the other 19 agents' hook templates
// is follow-up work.
func runDoctorSecurity() error {
	agents, err := currentAgentInfos(false)
	if err != nil {
		return err
	}

	out.Global().Println("tok doctor --security — hook integrity audit")
	out.Global().Println("=============================================")

	type row struct {
		agent    string
		hookPath string
		status   string
		detail   string
	}
	var rows []row
	tampered := 0
	noBaseline := 0
	verified := 0

	for _, agent := range agents {
		if agent.HookDir == "" {
			continue
		}
		hookPath := filepath.Join(agent.HookDir, "tok-rewrite.sh")
		if _, err := os.Stat(hookPath); err != nil {
			continue
		}
		result, err := integrity.VerifyHookAt(hookPath)
		if err != nil {
			rows = append(rows, row{agent.Name, hookPath, "ERROR", err.Error()})
			continue
		}
		switch result.Status {
		case integrity.StatusVerified:
			verified++
			rows = append(rows, row{agent.Name, hookPath, "VERIFIED", "hash matches baseline"})
		case integrity.StatusTampered:
			tampered++
			detail := fmt.Sprintf("expected %s…  actual %s…",
				truncateForSecurity(result.Expected, 12), truncateForSecurity(result.Actual, 12))
			rows = append(rows, row{agent.Name, hookPath, "TAMPERED", detail})
		case integrity.StatusNoBaseline:
			noBaseline++
			rows = append(rows, row{agent.Name, hookPath, "NO_BASELINE", "run `tok init --" + agentShortFlag(agent.Name) + "` to record"})
		case integrity.StatusOutdated:
			rows = append(rows, row{agent.Name, hookPath, "OUTDATED", "hook version older than current; rerun tok init"})
		case integrity.StatusOrphanedHash:
			rows = append(rows, row{agent.Name, hookPath, "ORPHANED", "baseline present but hook missing"})
		}
	}

	if len(rows) == 0 {
		out.Global().Println("  no tok-installed hooks found.")
		return nil
	}
	for _, r := range rows {
		icon := "✓"
		switch r.status {
		case "TAMPERED", "ERROR":
			icon = "✗"
		case "NO_BASELINE", "OUTDATED", "ORPHANED":
			icon = "⚠"
		}
		out.Global().Printf("  %s %-20s %s  [%s]\n", icon, r.agent, r.status, r.detail)
		out.Global().Printf("       %s\n", r.hookPath)
	}

	out.Global().Println()
	out.Global().Printf("summary: %d verified, %d no-baseline, %d tampered\n",
		verified, noBaseline, tampered)

	if tampered > 0 || noBaseline > 0 {
		return fmt.Errorf("integrity audit failed: %d tampered, %d without baseline", tampered, noBaseline)
	}
	return nil
}

func truncateForSecurity(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func agentShortFlag(name string) string {
	switch name {
	case "Claude Code":
		return "claude"
	case "Gemini CLI":
		return "gemini"
	case "GitHub Copilot":
		return "copilot"
	case "Qwen Code":
		return "qwen"
	case "Google Antigravity":
		return "antigravity"
	case "Kilo Code":
		return "kilocode"
	case "Roo Code":
		return "roocode"
	default:
		return strings.ToLower(name)
	}
}
