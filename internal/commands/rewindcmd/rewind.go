// Package rewindcmd provides CLI commands for the RewindStore.
//
// RewindStore provides zero-loss compression by archiving original
// command outputs before compression. Users can retrieve the full
// uncompressed output at any time using the hash identifier.
package rewindcmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/rewind"
)

var (
	headerColor = color.New(color.FgCyan, color.Bold)
	hashColor   = color.New(color.FgYellow)
	cmdColor    = color.New(color.FgGreen)
	dimColor    = color.New(color.Faint)
	boldColor   = color.New(color.Bold)
	statColor   = color.New(color.FgMagenta)
)

var rewindCmd = &cobra.Command{
	Use:   "rewind",
	Short: "Access original uncompressed command outputs",
	Long: `RewindStore provides zero-loss compression by archiving original
command outputs before compression. You can retrieve the full uncompressed 
output at any time using the hash identifier.

RewindStore is inspired by OMNI's zero-loss architecture.`,
	Example: `  tokman rewind list               # List recent entries
  tokman rewind show abc123        # Show original output for hash
  tokman rewind diff abc123        # Show diff between original and filtered
  tokman rewind delete abc123      # Delete a specific entry
  tokman rewind prune              # Remove old entries
  tokman rewind stats              # Show storage statistics`,
}

var rewindListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent RewindStore entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		entries, err := store.List(limit)
		if err != nil {
			return fmt.Errorf("list entries: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No entries in RewindStore.")
			fmt.Println("RewindStore archives original outputs when commands are filtered.")
			return nil
		}

		headerColor.Println("RewindStore Entries")
		fmt.Println(strings.Repeat("─", 70))
		fmt.Printf("%-16s  %-20s  %8s  %8s  %6s  %s\n",
			"Hash", "Command", "Original", "Filtered", "Saved", "When")
		fmt.Println(strings.Repeat("─", 70))

		for _, e := range entries {
			cmdStr := e.Command
			if e.Args != "" {
				cmdStr += " " + e.Args
			}
			if len(cmdStr) > 20 {
				cmdStr = cmdStr[:17] + "..."
			}

			hashColor.Printf("%-16s", e.Hash)
			fmt.Print("  ")
			cmdColor.Printf("%-20s", cmdStr)
			fmt.Printf("  %8d  %8d  %5.0f%%  %s\n",
				e.OriginalTokens, e.FilteredTokens,
				e.CompressionPct, e.Timestamp.Format("15:04:05"))
		}

		fmt.Println(strings.Repeat("─", 70))
		dimColor.Printf("Showing %d entries. Use --limit N for more.\n", len(entries))

		return nil
	},
}

var rewindShowCmd = &cobra.Command{
	Use:   "show <hash>",
	Short: "Show original uncompressed output for a hash",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash := args[0]
		showFiltered, _ := cmd.Flags().GetBool("filtered")

		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		entry, err := store.Retrieve(hash)
		if err != nil {
			return fmt.Errorf("entry not found: %s", hash)
		}

		if showFiltered {
			headerColor.Printf("Filtered output for %s (%s %s):\n", hash, entry.Command, entry.Args)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println(entry.FilteredOutput)
		} else {
			headerColor.Printf("Original output for %s (%s %s):\n", hash, entry.Command, entry.Args)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println(entry.OriginalOutput)
		}

		dimColor.Printf("\nTokens: %d → %d (%.0f%% saved)\n",
			entry.OriginalTokens, entry.FilteredTokens, entry.CompressionPct)

		return nil
	},
}

var rewindDiffCmd = &cobra.Command{
	Use:   "diff <hash>",
	Short: "Show diff between original and filtered output",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash := args[0]

		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		entry, err := store.Retrieve(hash)
		if err != nil {
			return fmt.Errorf("entry not found: %s", hash)
		}

		headerColor.Printf("Diff for %s (%s %s):\n", hash, entry.Command, entry.Args)
		fmt.Println(strings.Repeat("─", 60))

		// Show original with red
		red := color.New(color.FgRed)
		green := color.New(color.FgGreen)

		origLines := strings.Split(entry.OriginalOutput, "\n")
		filtLines := strings.Split(entry.FilteredOutput, "\n")

		red.Printf("--- Original (%d tokens, %d lines)\n", entry.OriginalTokens, len(origLines))
		green.Printf("+++ Filtered (%d tokens, %d lines)\n", entry.FilteredTokens, len(filtLines))
		fmt.Println(strings.Repeat("─", 60))

		// Simple diff: show what was removed and what remains
		filtSet := make(map[string]bool)
		for _, l := range filtLines {
			filtSet[strings.TrimSpace(l)] = true
		}

		for _, l := range origLines {
			trimmed := strings.TrimSpace(l)
			if trimmed == "" {
				continue
			}
			if filtSet[trimmed] {
				green.Printf("  %s\n", l)
			} else {
				red.Printf("- %s\n", l)
			}
		}

		fmt.Println(strings.Repeat("─", 60))
		statColor.Printf("Savings: %d tokens (%.0f%%)\n", entry.TokensSaved, entry.CompressionPct)

		return nil
	},
}

var rewindDeleteCmd = &cobra.Command{
	Use:   "delete <hash>",
	Short: "Delete a specific entry from RewindStore",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash := args[0]

		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		if err := store.Delete(hash); err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}

		fmt.Printf("Deleted entry: %s\n", hash)
		return nil
	},
}

var rewindPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old entries from RewindStore",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		pruned, err := store.Prune()
		if err != nil {
			return fmt.Errorf("prune failed: %w", err)
		}

		if pruned == 0 {
			fmt.Println("No expired entries to prune.")
		} else {
			fmt.Printf("Pruned %d expired entries.\n", pruned)
		}
		return nil
	},
}

var rewindStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show RewindStore statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStore()
		if err != nil {
			return err
		}
		defer store.Close()

		stats, err := store.GetStats()
		if err != nil {
			return fmt.Errorf("get stats: %w", err)
		}

		headerColor.Println("RewindStore Statistics")
		fmt.Println(strings.Repeat("─", 40))

		fmt.Printf("Total entries:     %d\n", stats.TotalEntries)
		fmt.Printf("Original tokens:   %d\n", stats.TotalOriginal)
		fmt.Printf("Filtered tokens:   %d\n", stats.TotalFiltered)
		statColor.Printf("Tokens saved:      %d\n", stats.TotalSaved)
		fmt.Printf("Avg compression:   %.1f%%\n", stats.AvgCompression)

		// Format database size
		sizeStr := formatBytes(stats.DatabaseSize)
		fmt.Printf("Database size:     %s\n", sizeStr)

		if !stats.OldestEntry.IsZero() {
			fmt.Printf("Oldest entry:      %s\n", stats.OldestEntry.Format("2006-01-02 15:04:05"))
			fmt.Printf("Newest entry:      %s\n", stats.NewestEntry.Format("2006-01-02 15:04:05"))
		}

		fmt.Println(strings.Repeat("─", 40))
		return nil
	},
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func getStore() (*rewind.Store, error) {
	cfg := rewind.DefaultConfig()
	store, err := rewind.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("open RewindStore: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("RewindStore is disabled. Enable in config: [rewind] enabled = true")
	}
	return store, nil
}

func init() {
	// Add subcommands
	rewindCmd.AddCommand(rewindListCmd)
	rewindCmd.AddCommand(rewindShowCmd)
	rewindCmd.AddCommand(rewindDiffCmd)
	rewindCmd.AddCommand(rewindDeleteCmd)
	rewindCmd.AddCommand(rewindPruneCmd)
	rewindCmd.AddCommand(rewindStatsCmd)

	// Flags
	rewindListCmd.Flags().Int("limit", 20, "Number of entries to show")
	rewindShowCmd.Flags().Bool("filtered", false, "Show filtered output instead of original")

	// Register with command registry
	registry.Add(func() { registry.Register(rewindCmd) })
}
