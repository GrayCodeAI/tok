package core

import (
	"fmt"
	"os"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/cache"
	"github.com/GrayCodeAI/tok/internal/commands/registry"
)

func init() {
	registry.Add(func() {
		registry.Register(cacheCmd)
	})
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheTopCmd)
	cacheCmd.AddCommand(cacheCleanupCmd)
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the query cache",
	Long: `Manage tok's persistent query cache for instant command retrieval.

The cache stores filtered command outputs keyed by command fingerprint.
This enables 56s → 1s speedup on repeated commands in the same git state.

Subcommands:
  stats     Show cache statistics
  clear     Clear all cached entries
  top       Show most frequently accessed queries
  cleanup   Remove old entries
`,
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Run: func(cmd *cobra.Command, args []string) {
		qc, err := cache.NewQueryCache("")
		if err != nil {
			out.Global().Errorf("Error opening cache: %v\n", err)
			os.Exit(1)
		}
		defer qc.Close()

		stats, err := qc.Stats()
		if err != nil {
			out.Global().Errorf("Error getting stats: %v\n", err)
			os.Exit(1)
		}

		hits, misses := qc.GetRuntimeStats()

		// Print stats
		green := color.New(color.FgGreen).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		out.Global().Println("Query Cache Statistics")
		out.Global().Println("=========================")
		out.Global().Printf("Total Entries: %s\n", green(fmt.Sprintf("%d", stats.TotalEntries)))
		out.Global().Printf("Cache Hits:    %s\n", green(fmt.Sprintf("%d", hits)))
		out.Global().Printf("Cache Misses:  %s\n", yellow(fmt.Sprintf("%d", misses)))
		out.Global().Printf("Hit Rate:      %s\n", blue(fmt.Sprintf("%.1f%%", stats.HitRate*100)))
		out.Global().Printf("Total Saved:   %s tokens\n", green(fmt.Sprintf("%d", stats.TotalSaved)))
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached entries",
	Run: func(cmd *cobra.Command, args []string) {
		qc, err := cache.NewQueryCache("")
		if err != nil {
			out.Global().Errorf("Error opening cache: %v\n", err)
			os.Exit(1)
		}
		defer qc.Close()

		stats, err := qc.Stats()
		if err != nil {
			out.Global().Errorf("Error getting stats: %v\n", err)
			return
		}
		count := stats.TotalEntries

		// Clear all
		qc.Invalidate(func(e *cache.CacheEntry) bool {
			return true
		})

		green := color.New(color.FgGreen).SprintFunc()
		out.Global().Printf("✓ Cleared %s cached entries\n", green(fmt.Sprintf("%d", count)))
	},
}

var cacheTopCmd = &cobra.Command{
	Use:   "top [n]",
	Short: "Show top N most accessed queries",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		limit := 10
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}

		qc, err := cache.NewQueryCache("")
		if err != nil {
			out.Global().Errorf("Error opening cache: %v\n", err)
			os.Exit(1)
		}
		defer qc.Close()

		entries, err := qc.GetTopQueries(limit)
		if err != nil {
			out.Global().Errorf("Error getting top queries: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			out.Global().Println("No cached queries found.")
			return
		}

		out.Global().Printf("Top %d Most Accessed Queries\n", len(entries))
		out.Global().Println("================================")

		for i, entry := range entries {
			out.Global().Printf("\n%d. %s %s\n", i+1, entry.Command, entry.Args)
			out.Global().Printf("   Hits: %d | Saved: %d tokens (%.1f%%)\n",
				entry.HitCount,
				entry.OriginalTokens-entry.FilteredTokens,
				entry.CompressionRatio*100)
			out.Global().Printf("   Last: %s\n", entry.AccessedAt.Format("2006-01-02 15:04"))
		}
	},
}

var cacheCleanupCmd = &cobra.Command{
	Use:   "cleanup [days]",
	Short: "Remove entries older than N days",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		days := 30
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &days)
		}

		qc, err := cache.NewQueryCache("")
		if err != nil {
			out.Global().Errorf("Error opening cache: %v\n", err)
			os.Exit(1)
		}
		defer qc.Close()

		statsBefore, err := qc.Stats()
		if err != nil {
			out.Global().Errorf("Error getting stats: %v\n", err)
			return
		}

		// Cleanup
		maxAge := time.Duration(days) * 24 * time.Hour
		err = qc.Cleanup(maxAge)
		if err != nil {
			out.Global().Errorf("Error cleaning up: %v\n", err)
			os.Exit(1)
		}

		statsAfter, err := qc.Stats()
		if err != nil {
			out.Global().Errorf("Error getting stats: %v\n", err)
			return
		}
		removed := statsBefore.TotalEntries - statsAfter.TotalEntries

		green := color.New(color.FgGreen).SprintFunc()
		out.Global().Printf("✓ Removed %s entries older than %d days\n",
			green(fmt.Sprintf("%d", removed)), days)
		out.Global().Printf("  Remaining: %d entries\n", statsAfter.TotalEntries)
	},
}
