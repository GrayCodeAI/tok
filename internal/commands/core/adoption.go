package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/discover"
)

var adoptionCmd = &cobra.Command{
	Use:   "adoption",
	Short: "Show tok adoption across Claude Code sessions",
	Long: `Analyze Claude Code session history to show tok adoption statistics.

This command scans Claude Code JSONL session files and calculates:
- Percentage of commands routed through tok
- Token savings by session
- Adoption trends over time

Examples:
  tok adoption              # Show last 10 sessions
  tok adoption --limit 20   # Show last 20 sessions`,
	RunE: runAdoption,
}

var adoptionLimit int

func init() {
	registry.Add(func() { registry.Register(adoptionCmd) })
	adoptionCmd.Flags().IntVarP(&adoptionLimit, "limit", "l", 10, "Maximum sessions to show")
}

// SessionSummary represents a summarized session for display
type SessionSummary struct {
	ID           string
	Date         string
	TotalCmds    int
	tokCmds      int
	OutputTokens int
}

func (s *SessionSummary) AdoptionPct() float64 {
	if s.TotalCmds == 0 {
		return 0
	}
	return float64(s.tokCmds) / float64(s.TotalCmds) * 100
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

func runAdoption(cmd *cobra.Command, args []string) error {
	sessions, err := findClaudeSessions()
	if err != nil {
		return fmt.Errorf("failed to find Claude sessions: %w", err)
	}

	if len(sessions) == 0 {
		out.Global().Println("No Claude Code sessions found in the last 30 days.")
		out.Global().Println("Make sure Claude Code has been used at least once.")
		return nil
	}

	// Sort by modification time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		infoI, _ := os.Stat(sessions[i])
		infoJ, _ := os.Stat(sessions[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Take top N
	if len(sessions) > adoptionLimit {
		sessions = sessions[:adoptionLimit]
	}

	var summaries []SessionSummary

	for _, path := range sessions {
		cmds, err := extractCommandsFromSession(path)
		if err != nil {
			continue
		}

		if len(cmds) == 0 {
			continue
		}

		total, tok := counttokCommands(cmds)

		// Extract session ID from filename
		id := filepath.Base(path)
		id = strings.TrimSuffix(id, filepath.Ext(id))
		if len(id) > 8 {
			id = id[:8]
		}

		// Get relative time
		info, _ := os.Stat(path)
		dateStr := "?"
		if info != nil {
			dateStr = formatRelativeTime(info.ModTime())
		}

		summaries = append(summaries, SessionSummary{
			ID:           id,
			Date:         dateStr,
			TotalCmds:    total,
			tokCmds:      tok,
			OutputTokens: estimateTokens(cmds),
		})
	}

	if len(summaries) == 0 {
		out.Global().Println("No sessions with Bash commands found.")
		return nil
	}

	// Display table
	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint("tok Adoption Overview"))
	out.Global().Println(strings.Repeat("─", 70))
	out.Global().Printf("%-12s %-12s %5s %5s %9s %-7s %8s\n",
		"Session", "Date", "Cmds", "TokM", "Adoption", "", "Output")
	out.Global().Println(strings.Repeat("─", 70))

	var totalCmds, totaltok int

	for _, s := range summaries {
		pct := s.AdoptionPct()
		bar := progressBar(pct, 5)
		totalCmds += s.TotalCmds
		totaltok += s.tokCmds

		out.Global().Printf("%-12s %-12s %5d %5d %8.0f%% %-7s %8s\n",
			s.ID,
			s.Date,
			s.TotalCmds,
			s.tokCmds,
			pct,
			bar,
			formatTokensInt(s.OutputTokens),
		)
	}

	out.Global().Println(strings.Repeat("─", 70))

	avgAdoption := 0.0
	if totalCmds > 0 {
		avgAdoption = float64(totaltok) / float64(totalCmds) * 100
	}
	out.Global().Printf("Average adoption: %.0f%%\n", avgAdoption)
	out.Global().Println()
	out.Global().Println("Tip: Run 'tok discover' to find missed optimization opportunities")

	return nil
}

// findClaudeSessions discovers Claude Code session files
func findClaudeSessions() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Common Claude Code session locations
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
			// Skip subagent files
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

		// Look for assistant messages with tool_use (Bash commands)
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

// counttokCommands counts how many commands are tok-covered
func counttokCommands(commands []string) (total, tok int) {
	for _, cmd := range commands {
		parts := splitCommandChain(cmd)
		for _, part := range parts {
			total++
			// Check if command starts with "tok" or would be rewritten
			if strings.HasPrefix(strings.TrimSpace(part), "tok ") {
				tok++
			} else {
				// Check if discover would rewrite this command
				rewritten, changed := discover.RewriteCommand(part, nil)
				if changed && rewritten != part {
					tok++
				}
			}
		}
	}
	return total, tok
}

// splitCommandChain splits chained commands (&&, ;, ||)
func splitCommandChain(cmd string) []string {
	var parts []string
	separators := []string{" && ", ";", " || "}

	remaining := cmd
	for _, sep := range separators {
		if strings.Contains(remaining, sep) {
			split := strings.Split(remaining, sep)
			for _, s := range split {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
			return parts
		}
	}

	// No chaining found
	return []string{strings.TrimSpace(cmd)}
}

// estimateTokens roughly estimates token count from commands
func estimateTokens(commands []string) int {
	totalChars := 0
	for _, cmd := range commands {
		totalChars += len(cmd)
	}
	// Rough estimate: 4 chars per token
	return totalChars / 4
}

// progressBar creates a simple ASCII progress bar
func progressBar(pct float64, width int) string {
	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return strings.Repeat("@", filled) + strings.Repeat(".", empty)
}

// formatRelativeTime formats a time as relative (Today, Yesterday, 5d ago)
func formatRelativeTime(t time.Time) string {
	diff := time.Since(t)
	days := int(diff.Hours() / 24)

	switch days {
	case 0:
		return "Today"
	case 1:
		return "Yesterday"
	default:
		return fmt.Sprintf("%dd ago", days)
	}
}
