package archive

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	searchCategory string
	searchAgent    string
	searchTags     []string
	searchLimit    int
)

func init() {
	registry.Add(func() {
		registry.Register(searchCmd)
	})
}

var searchCmd = &cobra.Command{
	Use:   "archive-search <query>",
	Short: "Search archives by content and metadata",
	Long: `Search through archived content and metadata.

Search supports:
- Content search: find text within archived content
- Command search: find by command name
- Path search: find by working directory
- Tag search: find by tags
- Combined search: mix multiple criteria

The query is matched against:
- Archive content (if --content flag used)
- Command names
- Working directory paths
- Project paths
- Tags
- Metadata fields`,
	Example: `  # Search by content
  tokman archive-search "error message" --content

  # Search by command
  tokman archive-search "git status"

  # Search with tags
  tokman archive-search "production" --tags="important"

  # Search by category
  tokman archive-search "deploy" --category=command

  # Combined search
  tokman archive-search "build" --category=command --tags="ci"`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&searchCategory, "category", "", "Filter by category")
	searchCmd.Flags().StringVar(&searchAgent, "agent", "", "Filter by agent")
	searchCmd.Flags().StringSliceVar(&searchTags, "tags", []string{}, "Filter by tags")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "Maximum results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	query := args[0]

	// Create archive manager
	cfg := archive.DefaultArchiveConfig()
	mgr, err := archive.NewArchiveManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	// Initialize
	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize archive manager: %w", err)
	}

	// Build search options
	opts := archive.ArchiveListOptions{
		Limit:  searchLimit,
		Query:  query,
		SortBy: "created_at",
	}

	if searchCategory != "" {
		opts.Category = archive.ArchiveCategory(searchCategory)
	}

	// Search archives
	result, err := mgr.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to search archives: %w", err)
	}

	// Filter by tags if specified
	if len(searchTags) > 0 {
		result.Entries = filterByTags(result.Entries, searchTags)
	}

	// Output results
	if len(result.Entries) == 0 {
		fmt.Println("No archives found matching your search")
		return nil
	}

	fmt.Printf("\n%s Found %d %s for '%s'\n\n",
		color.GreenString("✓"),
		len(result.Entries),
		pluralize(len(result.Entries), "result", "results"),
		color.CyanString(query))

	for i, entry := range result.Entries {
		printSearchResult(i+1, entry)
	}

	return nil
}

func filterByTags(entries []archive.ArchiveEntry, requiredTags []string) []archive.ArchiveEntry {
	var filtered []archive.ArchiveEntry

	for _, entry := range entries {
		hasAllTags := true
		for _, requiredTag := range requiredTags {
			found := false
			for _, entryTag := range entry.Tags {
				if strings.EqualFold(entryTag, requiredTag) {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func printSearchResult(index int, entry archive.ArchiveEntry) {
	hash := entry.Hash
	if len(hash) > 16 {
		hash = hash[:16] + "..."
	}

	cmd := truncate(entry.Command, 50)

	fmt.Printf("%d. %s\n", index, color.YellowString(hash))
	fmt.Printf("   Command: %s\n", cmd)
	fmt.Printf("   Category: %s", entry.Category)

	if entry.Agent != "" {
		fmt.Printf(" | Agent: %s", entry.Agent)
	}

	fmt.Printf(" | Size: %s\n", formatBytes(entry.CompressedSize))

	if len(entry.Tags) > 0 {
		fmt.Printf("   Tags: %s\n", strings.Join(entry.Tags, ", "))
	}

	fmt.Printf("   Created: %s\n", entry.CreatedAt.Format("2006-01-02 15:04"))

	if entry.AccessedAt != nil {
		fmt.Printf(" | Accessed: %dx", entry.AccessCount)
	}

	fmt.Println("\n")
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
