package system

import (
	"bufio"
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/discover"
)

var (
	discoverProject string
	discoverLimit   int
	discoverAll     bool
	discoverSince   int
	discoverFormat  string
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover missed token savings from Claude Code history",
	Long: `Analyze Claude Code session history to find commands that could have
used tok wrappers for token savings.

Scans Claude Code JSONL session files to identify commands that weren't
rewritten and estimates potential savings.

Examples:
  tok discover                 # Scan current project
  tok discover --all           # Scan all projects
  tok discover --since 7       # Last 7 days only
  tok discover --format json   # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiscoverEnhanced()
	},
}

func init() {
	registry.Add(func() { registry.Register(discoverCmd) })

	discoverCmd.Flags().StringVarP(&discoverProject, "project", "p", "", "Filter by project path (substring match)")
	discoverCmd.Flags().IntVarP(&discoverLimit, "limit", "l", 15, "Max commands per section")
	discoverCmd.Flags().BoolVarP(&discoverAll, "all", "a", false, "Scan all projects (default: current project only)")
	discoverCmd.Flags().IntVarP(&discoverSince, "since", "s", 30, "Limit to sessions from last N days")
	discoverCmd.Flags().StringVarP(&discoverFormat, "format", "f", "text", "Output format: text, json")
}

// ClaudeJSONLMessage represents the structure of Claude Code JSONL files
type ClaudeJSONLMessage struct {
	Type    string        `json:"type"`
	Message ClaudeMessage `json:"message"`
}

type ClaudeMessage struct {
	Role    string          `json:"role"`
	Content []ClaudeContent `json:"content"`
}

type ClaudeContent struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
}

// DiscoveredCommand represents a discovered command pattern
type DiscoveredCommand struct {
	Command      string `json:"command"`
	Count        int    `json:"count"`
	Category     string `json:"category"`
	tokEquiv     string
	EstSavings   float64 `json:"estimated_savings_pct"`
	TokensSaved  int     `json:"tokens_saved,omitempty"`
	SupportLevel string  `json:"support_level,omitempty"`
}

// DiscoverResult represents the discovery results
type DiscoverResult struct {
	SessionsScanned   int                 `json:"sessions_scanned"`
	TotalCommands     int                 `json:"total_commands"`
	Alreadytok        int                 `json:"already_tok"`
	SupportedMissed   []DiscoveredCommand `json:"supported_missed"`
	PassthroughMissed []DiscoveredCommand `json:"passthrough_missed,omitempty"`
	Unsupported       []DiscoveredCommand `json:"unsupported,omitempty"`
	ParseErrors       int                 `json:"parse_errors"`
	tokBypassCount    int
	tokBypassCmds     []string
}

// runDiscoverEnhanced scans Claude Code JSONL files for missed tok usage.
func runDiscoverEnhanced() error {
	sessions, err := findClaudeSessions()
	if err != nil {
		return fmt.Errorf("failed to find Claude sessions: %w", err)
	}

	// Filter by project if specified
	if !discoverAll && discoverProject != "" {
		var filtered []string
		for _, s := range sessions {
			if strings.Contains(s, discoverProject) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Filter by time
	cutoff := time.Now().Add(-time.Duration(discoverSince) * 24 * time.Hour)
	var recentSessions []string
	for _, s := range sessions {
		info, err := os.Stat(s)
		if err != nil {
			continue
		}
		if info.ModTime().After(cutoff) {
			recentSessions = append(recentSessions, s)
		}
	}
	sessions = recentSessions

	// Categorize commands
	result := &DiscoverResult{
		SessionsScanned:   len(sessions),
		SupportedMissed:   []DiscoveredCommand{},
		PassthroughMissed: []DiscoveredCommand{},
		Unsupported:       []DiscoveredCommand{},
	}

	supportedMap := make(map[string]*DiscoveredCommand)
	passthroughMap := make(map[string]*DiscoveredCommand)
	unsupportedMap := make(map[string]*DiscoveredCommand)
	tokBypassMap := make(map[string]int)

	for _, sessionPath := range sessions {
		commands, err := extractCommandsFromSession(sessionPath)
		if err != nil {
			result.ParseErrors++
			continue
		}

		for _, cmd := range commands {
			parts := splitCommandChain(cmd)
			for _, part := range parts {
				result.TotalCommands++

				// Check for TOK_DISABLED bypass
				if hasDisabledPrefix(part) {
					actualCmd := stripDisabledPrefix(part)
					if isSupportedCommand(actualCmd) {
						result.tokBypassCount++
						tokBypassMap[actualCmd]++
					}
					continue
				}

				// Check if already using tok
				if strings.HasPrefix(strings.TrimSpace(part), "tok ") {
					result.Alreadytok++
					continue
				}

				rewritten, supportLevel := discover.ClassifyCommand(part)
				switch supportLevel {
				case discover.SupportOptimized:
					slug := getCommandSlug(part)
					cat := categorizeCommand(slug)
					savings := estimateSavings(cat)

					if entry, ok := supportedMap[slug]; ok {
						entry.Count++
						entry.TokensSaved += estimateTokens(part) * int(savings) / 100
					} else {
						supportedMap[slug] = &DiscoveredCommand{
							Command:      part,
							Count:        1,
							Category:     cat,
							tokEquiv:     rewritten,
							EstSavings:   savings,
							TokensSaved:  estimateTokens(part) * int(savings) / 100,
							SupportLevel: string(discover.SupportOptimized),
						}
					}
				case discover.SupportPassthrough:
					slug := getCommandSlug(part)
					cat := categorizeCommand(slug)
					if entry, ok := passthroughMap[slug]; ok {
						entry.Count++
					} else {
						passthroughMap[slug] = &DiscoveredCommand{
							Command:      part,
							Count:        1,
							Category:     cat,
							tokEquiv:     rewritten,
							SupportLevel: string(discover.SupportPassthrough),
						}
					}
				default:
					slug := getCommandSlug(part)
					if entry, ok := unsupportedMap[slug]; ok {
						entry.Count++
					} else {
						unsupportedMap[slug] = &DiscoveredCommand{
							Command:      part,
							Count:        1,
							Category:     "unknown",
							SupportLevel: string(discover.SupportUnsupported),
						}
					}
				}
			}
		}
	}

	// Convert maps to slices
	for _, v := range supportedMap {
		result.SupportedMissed = append(result.SupportedMissed, *v)
	}
	for _, v := range passthroughMap {
		result.PassthroughMissed = append(result.PassthroughMissed, *v)
	}
	for _, v := range unsupportedMap {
		result.Unsupported = append(result.Unsupported, *v)
	}

	// Sort by count
	sort.Slice(result.SupportedMissed, func(i, j int) bool {
		return result.SupportedMissed[i].Count > result.SupportedMissed[j].Count
	})
	sort.Slice(result.Unsupported, func(i, j int) bool {
		return result.Unsupported[i].Count > result.Unsupported[j].Count
	})
	sort.Slice(result.PassthroughMissed, func(i, j int) bool {
		return result.PassthroughMissed[i].Count > result.PassthroughMissed[j].Count
	})

	// Get top bypass commands
	for cmd, count := range tokBypassMap {
		result.tokBypassCmds = append(result.tokBypassCmds, fmt.Sprintf("%s (%dx)", cmd, count))
	}
	sort.Strings(result.tokBypassCmds)

	// Limit results
	if len(result.SupportedMissed) > discoverLimit {
		result.SupportedMissed = result.SupportedMissed[:discoverLimit]
	}
	if len(result.PassthroughMissed) > discoverLimit {
		result.PassthroughMissed = result.PassthroughMissed[:discoverLimit]
	}
	if len(result.Unsupported) > discoverLimit {
		result.Unsupported = result.Unsupported[:discoverLimit]
	}

	// Output
	if discoverFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	return printDiscoverText(result)
}

// findClaudeSessions discovers Claude Code session files
func findClaudeSessions() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	possibleDirs := []string{
		filepath.Join(home, ".claude", "sessions"),
		filepath.Join(home, "Library", "Application Support", "Claude", "sessions"),
		filepath.Join(home, ".local", "share", "claude", "sessions"),
	}

	var sessions []string
	for _, dir := range possibleDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !strings.Contains(f, "subagent") {
				sessions = append(sessions, f)
			}
		}
	}

	return sessions, nil
}

// extractCommandsFromSession extracts Bash commands from a Claude Code JSONL file
func extractCommandsFromSession(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var msg ClaudeJSONLMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.Type == "assistant" && msg.Message.Role == "assistant" {
			for _, content := range msg.Message.Content {
				if content.Type == "tool_use" && content.Name == "Bash" {
					if cmd, ok := content.Input["command"].(string); ok {
						commands = append(commands, cmd)
					}
				}
			}
		}
	}

	return commands, scanner.Err()
}

// Helper functions for discover
func splitCommandChain(cmd string) []string {
	separators := []string{" && ", ";", " || "}
	for _, sep := range separators {
		if strings.Contains(cmd, sep) {
			var parts []string
			for _, s := range strings.Split(cmd, sep) {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
			return parts
		}
	}
	return []string{strings.TrimSpace(cmd)}
}

func hasDisabledPrefix(cmd string) bool {
	return strings.HasPrefix(strings.TrimSpace(cmd), "TOK_DISABLED=1 ") ||
		strings.HasPrefix(strings.TrimSpace(cmd), "TOK_DISABLED=1")
}

func stripDisabledPrefix(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.TrimPrefix(cmd, "TOK_DISABLED=1 ")
	cmd = strings.TrimPrefix(cmd, "TOK_DISABLED=1")
	return strings.TrimSpace(cmd)
}

func isSupportedCommand(cmd string) bool {
	rewritten, changed := discover.RewriteCommand(cmd, nil)
	return changed && rewritten != cmd
}

func getCommandSlug(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return cmd
}

func categorizeCommand(slug string) string {
	parts := strings.Fields(slug)
	if len(parts) == 0 {
		return "unknown"
	}
	base := parts[0]

	categories := map[string]string{
		"git":     "git",
		"gh":      "git",
		"cargo":   "rust",
		"npm":     "js",
		"pnpm":    "js",
		"npx":     "js",
		"go":      "go",
		"docker":  "container",
		"kubectl": "container",
		"aws":     "cloud",
		"pytest":  "python",
		"ruff":    "python",
		"ls":      "system",
		"tree":    "system",
		"find":    "system",
	}

	if cat, ok := categories[base]; ok {
		return cat
	}
	return "other"
}

func estimateSavings(category string) float64 {
	savings := map[string]float64{
		"git":       80,
		"rust":      90,
		"js":        85,
		"go":        90,
		"container": 80,
		"cloud":     75,
		"python":    90,
		"system":    80,
	}
	if s, ok := savings[category]; ok {
		return s
	}
	return 70
}

func estimateTokens(cmd string) int {
	return len(cmd) / 4
}

func printDiscoverText(result *DiscoverResult) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	out.Global().Println()
	out.Global().Printf("%s\n", yellow("🔍 tok Discovery Report"))
	out.Global().Println(strings.Repeat("═", 60))
	out.Global().Printf("Sessions scanned: %d\n", result.SessionsScanned)
	out.Global().Printf("Total commands:   %d\n", result.TotalCommands)
	out.Global().Printf("Already tok:   %d (%.0f%%)\n", result.Alreadytok,
		float64(result.Alreadytok)/float64(result.TotalCommands)*100)
	out.Global().Println()

	if len(result.SupportedMissed) > 0 {
		out.Global().Printf("%s\n", cyan("Missed Opportunities"))
		out.Global().Println(strings.Repeat("─", 60))
		out.Global().Printf("%-24s %4s %8s %10s %6s\n", "Command", "Cnt", "Category", "Est.Saved", "Save%")
		out.Global().Println(strings.Repeat("─", 60))
		for _, cmd := range result.SupportedMissed {
			out.Global().Printf("%-24s %4d %8s %10s %5.0f%%\n",
				shared.Truncate(cmd.Command, 24),
				cmd.Count,
				cmd.Category,
				formatTokensInt(cmd.TokensSaved),
				cmd.EstSavings,
			)
		}
		out.Global().Println()
	}

	if len(result.PassthroughMissed) > 0 {
		out.Global().Printf("%s\n", cyan("Passthrough Coverage"))
		out.Global().Println(strings.Repeat("─", 60))
		out.Global().Printf("%-24s %4s %12s %12s\n", "Command", "Cnt", "Category", "Equivalent")
		out.Global().Println(strings.Repeat("─", 60))
		for _, cmd := range result.PassthroughMissed {
			out.Global().Printf("%-24s %4d %12s %12s\n",
				shared.Truncate(cmd.Command, 24),
				cmd.Count,
				cmd.Category,
				shared.Truncate(cmd.tokEquiv, 12),
			)
		}
		out.Global().Println()
	}

	if result.tokBypassCount > 0 {
		out.Global().Printf("%s %d\n", yellow("⚠️  TOK_DISABLED bypasses detected:"), result.tokBypassCount)
		for _, cmd := range result.tokBypassCmds {
			out.Global().Printf("   %s\n", cmd)
		}
		out.Global().Println()
	}

	if len(result.Unsupported) > 0 && shared.Verbose > 0 {
		out.Global().Printf("%s\n", cyan("Unsupported Commands"))
		out.Global().Println(strings.Repeat("─", 60))
		for _, cmd := range result.Unsupported[:min(len(result.Unsupported), 5)] {
			out.Global().Printf("  %-30s  %3dx\n", shared.Truncate(cmd.Command, 30), cmd.Count)
		}
		out.Global().Println()
	}

	if len(result.SupportedMissed) == 0 && len(result.PassthroughMissed) == 0 && result.tokBypassCount == 0 {
		out.Global().Printf("%s\n", green("✓ All commands are optimized!"))
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatTokensInt formats token count for display
func formatTokensInt(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// LegacyDiscoverResult for backward compatibility with old code
type LegacyDiscoverResult struct {
	Project         string              `json:"project,omitempty"`
	TotalCommands   int                 `json:"total_commands"`
	MissedSavings   int                 `json:"missed_savings"`
	Opportunities   []DiscoveredCommand `json:"opportunities"`
	UnsupportedCmds []DiscoveredCommand `json:"unsupported_commands,omitempty"`
}

// LegacyDiscoveredCommand for backward compatibility
type LegacyDiscoveredCommand struct {
	Command     string  `json:"command"`
	Count       int     `json:"count"`
	Category    string  `json:"category"`
	CouldSave   bool    `json:"could_save"`
	SavingsPct  float64 `json:"savings_percent,omitempty"`
	TokensSaved int     `json:"tokens_saved,omitempty"`
	Example     string  `json:"example,omitempty"`
}

func runDiscover() error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	projectFilter := discoverProject
	if projectFilter == "" && !discoverAll {
		projectFilter = shared.GetProjectPath()
	}

	// Initialize tracker to get historical data
	tracker, err := shared.OpenTracker()
	if err != nil {
		return fmt.Errorf("failed to initialize tracker: %w", err)
	}
	defer tracker.Close()

	// Get command stats from tracker
	stats, err := tracker.GetCommandStats(projectFilter)
	if err != nil {
		return fmt.Errorf("failed to get command stats: %w", err)
	}

	// Analyze commands for missed opportunities
	result := LegacyDiscoverResult{
		Project:       projectFilter,
		Opportunities: []DiscoveredCommand{},
	}

	// Known tok wrappers for analysis
	tokWrappers := map[string]string{
		"git":     "Git",
		"gh":      "GitHub",
		"cargo":   "Cargo",
		"docker":  "Infra",
		"kubectl": "Infra",
		"npm":     "PackageManager",
		"pnpm":    "PackageManager",
		"npx":     "PackageManager",
		"go":      "Go",
		"pytest":  "Python",
		"ruff":    "Python",
		"mypy":    "Build",
		"tsc":     "Build",
		"vitest":  "Tests",
		"curl":    "Network",
		"psql":    "Infra",
		"aws":     "Infra",
		"ls":      "Files",
		"find":    "Files",
		"tree":    "Files",
		"grep":    "Files",
	}

	unsupportedCommands := []DiscoveredCommand{}
	totalCommands := 0

	for _, stat := range stats {
		totalCommands += stat.ExecutionCount

		// Check if this command was already using tok
		if strings.HasPrefix(stat.Command, "tok ") {
			continue // Already optimized
		}

		// Extract base command
		parts := strings.Fields(stat.Command)
		if len(parts) == 0 {
			continue
		}
		baseCmd := parts[0]

		// Check if it's a known wrapper opportunity
		if category, ok := tokWrappers[baseCmd]; ok {
			result.Opportunities = append(result.Opportunities, DiscoveredCommand{
				Command:     stat.Command,
				Count:       stat.ExecutionCount,
				Category:    category,
				EstSavings:  stat.ReductionPct,
				TokensSaved: stat.TotalSaved,
			})
		} else if stat.TotalSaved == 0 && stat.ExecutionCount >= 3 {
			// Unsupported command that's frequently used
			unsupportedCommands = append(unsupportedCommands, DiscoveredCommand{
				Command:  stat.Command,
				Count:    stat.ExecutionCount,
				Category: "Unknown",
			})
		}
	}

	result.TotalCommands = totalCommands
	result.MissedSavings = len(result.Opportunities)
	result.UnsupportedCmds = unsupportedCommands

	// Sort by count (descending)
	sort.Slice(result.Opportunities, func(i, j int) bool {
		return result.Opportunities[i].Count > result.Opportunities[j].Count
	})
	sort.Slice(result.UnsupportedCmds, func(i, j int) bool {
		return result.UnsupportedCmds[i].Count > result.UnsupportedCmds[j].Count
	})

	// Limit results
	if len(result.Opportunities) > discoverLimit {
		result.Opportunities = result.Opportunities[:discoverLimit]
	}
	if len(result.UnsupportedCmds) > discoverLimit {
		result.UnsupportedCmds = result.UnsupportedCmds[:discoverLimit]
	}

	// Output results
	if discoverFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	// Text output
	out.Global().Println()
	out.Global().Printf("%s\n", green("🔍 tok Discovery Report"))
	out.Global().Println("════════════════════════════════════════════════════")
	out.Global().Println()

	if projectFilter != "" {
		out.Global().Printf("  Project: %s\n", cyan(projectFilter))
	}
	out.Global().Printf("  Commands analyzed: %d\n", totalCommands)
	out.Global().Println()

	if len(result.Opportunities) > 0 {
		out.Global().Printf("  %s\n", yellow("Missed Opportunities (could use tok):"))
		out.Global().Println("  ─────────────────────────────────────────")
		for _, opp := range result.Opportunities {
			pct := ""
			if opp.EstSavings > 0 {
				pct = fmt.Sprintf("  %4.1f%% saved", opp.EstSavings)
			}
			out.Global().Printf("    %-30s  %3dx  [%s]%s\n", shared.Truncate(opp.Command, 30), opp.Count, opp.Category, pct)
		}
		out.Global().Println()
	}

	if len(result.UnsupportedCmds) > 0 && shared.Verbose > 0 {
		out.Global().Printf("  %s\n", cyan("Unsupported Commands (frequent but no tok wrapper):"))
		out.Global().Println("  ─────────────────────────────────────────")
		for _, cmd := range result.UnsupportedCmds {
			out.Global().Printf("    %-30s  %3dx\n", shared.Truncate(cmd.Command, 30), cmd.Count)
		}
		out.Global().Println()
	}

	if len(result.Opportunities) == 0 {
		out.Global().Printf("  %s\n", green("✓ All commands are already optimized!"))
		out.Global().Println()
	}

	return nil
}
