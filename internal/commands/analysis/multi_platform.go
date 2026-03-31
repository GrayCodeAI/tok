package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var multiPlatformCmd = &cobra.Command{
	Use:   "multi-platform",
	Short: "Track tokens across multiple AI platforms",
	Long: `Track token usage across 16+ AI clients including Claude Code,
Cursor, Codex, Gemini, OpenCode, and more.

Examples:
  tokman multi-platform scan
  tokman multi-platform scan --clients claude,cursor,gemini`,
	RunE: runMultiPlatform,
}

var mpClients string

// PlatformUsage tracks usage for a single AI platform.
type PlatformUsage struct {
	Name       string  `json:"name"`
	Sessions   int     `json:"sessions"`
	TokensUsed int64   `json:"tokens_used"`
	CostUSD    float64 `json:"cost_usd"`
	Model      string  `json:"model,omitempty"`
	LastActive string  `json:"last_active,omitempty"`
}

// MultiPlatformReport aggregates usage across platforms.
type MultiPlatformReport struct {
	Platforms     []PlatformUsage `json:"platforms"`
	TotalTokens   int64           `json:"total_tokens"`
	TotalCost     float64         `json:"total_cost"`
	TotalSessions int             `json:"total_sessions"`
}

func init() {
	registry.Add(func() { registry.Register(multiPlatformCmd) })
	multiPlatformCmd.Flags().StringVar(&mpClients, "clients", "", "Comma-separated list of clients to scan")
}

func runMultiPlatform(cmd *cobra.Command, args []string) error {
	homeDir, _ := os.UserHomeDir()
	var clients []string
	if mpClients != "" {
		clients = strings.Split(mpClients, ",")
	} else {
		clients = getAllClients()
	}

	var report MultiPlatformReport
	for _, client := range clients {
		usage := scanClient(homeDir, client)
		if usage.Sessions > 0 {
			report.Platforms = append(report.Platforms, usage)
			report.TotalTokens += usage.TokensUsed
			report.TotalCost += usage.CostUSD
			report.TotalSessions += usage.Sessions
		}
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))
	return nil
}

func getAllClients() []string {
	return []string{
		"claude", "cursor", "codex", "gemini", "opencode",
		"amp", "droid", "openclaw", "pi", "kimi",
		"qwen", "roo", "kilo", "mux", "synthetic", "copilot",
	}
}

func scanClient(homeDir, client string) PlatformUsage {
	usage := PlatformUsage{Name: client}

	sessionDirs := findSessionDirs(homeDir, client)
	usage.Sessions = len(sessionDirs)

	for _, dir := range sessionDirs {
		tokens := countTokensInDir(dir)
		usage.TokensUsed += tokens
	}

	// Estimate cost (rough: $10 per 1M tokens for GPT-4 class)
	usage.CostUSD = float64(usage.TokensUsed) / 1000000 * 10

	return usage
}

func findSessionDirs(homeDir, client string) []string {
	var dirs []string
	patterns := map[string]string{
		"claude":   ".claude/projects",
		"cursor":   ".cursor/projects",
		"codex":    ".codex/sessions",
		"gemini":   ".gemini/sessions",
		"opencode": ".opencode/sessions",
	}

	if pattern, ok := patterns[client]; ok {
		fullPath := filepath.Join(homeDir, pattern)
		if entries, err := os.ReadDir(fullPath); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					dirs = append(dirs, filepath.Join(fullPath, e.Name()))
				}
			}
		}
	}

	return dirs
}

func countTokensInDir(dir string) int64 {
	var total int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			total += int64(len(data)) / 4
		}
		return nil
	})
	return total
}

// FormatMultiPlatformReport returns a human-readable report string.
func FormatMultiPlatformReport(report MultiPlatformReport) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total: %d tokens, %.2f USD, %d sessions across %d platforms\n\n",
		report.TotalTokens, report.TotalCost, report.TotalSessions, len(report.Platforms)))
	for _, p := range report.Platforms {
		sb.WriteString(fmt.Sprintf("  %-12s: %8d tokens, %6.2f USD, %3d sessions\n",
			p.Name, p.TokensUsed, p.CostUSD, p.Sessions))
	}
	return sb.String()
}
