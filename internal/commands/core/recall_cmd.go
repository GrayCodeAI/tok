package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var recallCmd = &cobra.Command{
	Use:   "recall [query]",
	Short: "Search past command history semantically",
	Long: `Search through your command history to find past commands, outputs, and context.

This helps you remember what commands you ran, what worked, and reuse solutions
from previous sessions.

Examples:
  tok recall "git commit"              # Find git commit commands
  tok recall "docker run"             # Find docker commands
  tok recall "npm install"            # Find npm commands
  tok recall --limit 5                # Show last 5 commands
  tok recall --days 7                 # Search last 7 days only`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runRecall,
}

var (
	recallLimit int
	recallDays  int
	recallJSON  bool
)

func init() {
	registry.Add(func() { registry.Register(recallCmd) })

	recallCmd.Flags().IntVarP(&recallLimit, "limit", "n", 10, "Number of results to show")
	recallCmd.Flags().IntVar(&recallDays, "days", 30, "Search within last N days")
	recallCmd.Flags().BoolVar(&recallJSON, "json", false, "Output as JSON")
}

func runRecall(cmd *cobra.Command, args []string) error {
	dbPath := config.DatabasePath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No command history found.")
		fmt.Println("Run some commands through tok to start building history!")
		return nil
	}

	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open tracking database: %w", err)
	}
	defer tracker.Close()

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	commands, err := tracker.GetRecentCommands(config.ProjectPath(), 500)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	var results []recallResult
	cutoff := time.Now().AddDate(0, 0, -recallDays)

	for _, c := range commands {
		if c.Timestamp.Before(cutoff) {
			continue
		}

		score := matchScore(query, c.Command)
		if score > 0 || query == "" {
			results = append(results, recallResult{
				Command:     c.Command,
				Timestamp:   c.Timestamp,
				SavedTokens: c.SavedTokens,
				Filtered:    c.FilteredTokens,
				Score:       score,
			})
		}
	}

	if query != "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Timestamp.After(results[j].Timestamp)
		})
	}

	if len(results) > recallLimit {
		results = results[:recallLimit]
	}

	if recallJSON {
		return printRecallJSON(results)
	}

	if len(results) == 0 {
		fmt.Println("No matching commands found.")
		if query != "" {
			fmt.Printf("Try a different search term or use --days to extend the time range.\n")
		}
		return nil
	}

	fmt.Println(strings.Repeat("─", 50))

	bold := color.New(color.Bold)
	bold.Print("Recent Commands")
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))

	for i, r := range results {
		ts := r.Timestamp.Format("2006-01-02 15:04")
		cmdTrunc := r.Command
		if len(cmdTrunc) > 60 {
			cmdTrunc = cmdTrunc[:60] + "..."
		}

		fmt.Printf("%d. %s\n", i+1, color.CyanString(cmdTrunc))
		fmt.Printf("   %s | saved %d tokens | %d filtered\n", ts, r.SavedTokens, r.Filtered)
	}

	fmt.Println()
	fmt.Printf("Showing %d of %d results\n", len(results), len(results))

	return nil
}

type recallResult struct {
	Command     string
	Timestamp   time.Time
	SavedTokens int
	Filtered    int
	Score       float64
}

func matchScore(query, command string) float64 {
	if query == "" {
		return 1.0
	}

	query = strings.ToLower(query)
	command = strings.ToLower(command)

	if strings.Contains(command, query) {
		return 1.0
	}

	words := strings.Fields(query)
	matches := 0
	for _, w := range words {
		if strings.Contains(command, w) {
			matches++
		}
	}

	if matches == 0 {
		return 0
	}

	return float64(matches) / float64(len(words))
}

func printRecallJSON(results []recallResult) error {
	type jsonResult struct {
		Command     string `json:"command"`
		Timestamp   string `json:"timestamp"`
		SavedTokens int    `json:"saved_tokens"`
		Filtered    int    `json:"filtered_tokens"`
	}

	var jsonResults []jsonResult
	for _, r := range results {
		jsonResults = append(jsonResults, jsonResult{
			Command:     r.Command,
			Timestamp:   r.Timestamp.Format(time.RFC3339),
			SavedTokens: r.SavedTokens,
			Filtered:    r.Filtered,
		})
	}

	data, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
