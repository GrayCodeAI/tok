package archive

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	listCategory    string
	listAgent       string
	listProject     string
	listTags        []string
	listLimit       int
	listOffset      int
	listSortBy      string
	listSortOrder   string
	listAfter       string
	listBefore      string
	listOutput      string
	listShowDeleted bool
)

func init() {
	registry.Add(func() {
		registry.Register(listCmd)
	})
}

var listCmd = &cobra.Command{
	Use:     "archive-list",
	Aliases: []string{"archives", "ls-archive"},
	Short:   "List archived content with filters",
	Long: `List all archived content with powerful filtering options.

Filter by:
- Category: command, session, user, system
- Agent: claude, cursor, copilot, etc.
- Project path
- Tags
- Date range
- Sort order

Output formats:
- table (default): Formatted table
- json: JSON output
- csv: CSV format
- simple: Just hashes`,
	Example: `  # List all archives
  tokman archive-list

  # Filter by category
  tokman archive-list --category=command

  # Filter by agent
  tokman archive-list --agent=claude

  # Filter by tags
  tokman archive-list --tags="important,production"

  # Date range
  tokman archive-list --after="2024-01-01" --before="2024-12-31"

  # Pagination
  tokman archive-list --limit=50 --offset=100

  # Sort by access count
  tokman archive-list --sort-by=access_count --sort-order=desc

  # JSON output
  tokman archive-list --output=json`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&listCategory, "category", "", "Filter by category (command, session, user, system)")
	listCmd.Flags().StringVar(&listAgent, "agent", "", "Filter by agent")
	listCmd.Flags().StringVar(&listProject, "project", "", "Filter by project path")
	listCmd.Flags().StringSliceVar(&listTags, "tags", []string{}, "Filter by tags (comma-separated)")
	listCmd.Flags().IntVar(&listLimit, "limit", 100, "Maximum number of results")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "Offset for pagination")
	listCmd.Flags().StringVar(&listSortBy, "sort-by", "created_at", "Sort by field (created_at, accessed_at, size, access_count)")
	listCmd.Flags().StringVar(&listSortOrder, "sort-order", "desc", "Sort order (asc, desc)")
	listCmd.Flags().StringVar(&listAfter, "after", "", "Filter: created after date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listBefore, "before", "", "Filter: created before date (YYYY-MM-DD)")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "table", "Output format (table, json, csv, simple)")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Build options
	opts := buildListOptions()

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

	// List archives
	result, err := mgr.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list archives: %w", err)
	}

	// Output results
	switch listOutput {
	case "json":
		return outputJSON(result)
	case "csv":
		return outputCSV(result)
	case "simple":
		return outputSimple(result)
	default:
		return outputTable(result)
	}
}

func buildListOptions() archive.ArchiveListOptions {
	opts := archive.ArchiveListOptions{
		Limit:     listLimit,
		Offset:    listOffset,
		SortBy:    listSortBy,
		SortOrder: listSortOrder,
	}

	if listCategory != "" {
		opts.Category = archive.ArchiveCategory(listCategory)
	}

	if listAgent != "" {
		opts.Agent = listAgent
	}

	if listProject != "" {
		opts.ProjectPath = listProject
	}

	if len(listTags) > 0 {
		opts.Tags = listTags
	}

	if listAfter != "" {
		t, _ := time.Parse("2006-01-02", listAfter)
		opts.CreatedAfter = &t
	}

	if listBefore != "" {
		t, _ := time.Parse("2006-01-02", listBefore)
		opts.CreatedBefore = &t
	}

	return opts
}

func outputTable(result *archive.ArchiveListResult) error {
	if len(result.Entries) == 0 {
		fmt.Println("No archives found")
		return nil
	}

	// Header
	fmt.Printf("\n%s (%d total, showing %d-%d)\n\n",
		color.New(color.Bold).Sprint("Archives"),
		result.Total,
		listOffset+1,
		listOffset+len(result.Entries))

	// Table header
	fmt.Printf("%-16s %-40s %-12s %-10s %-15s %-15s\n",
		"HASH", "COMMAND", "CATEGORY", "SIZE", "CREATED", "ACCESSED")
	fmt.Println(strings.Repeat("-", 110))

	for _, entry := range result.Entries {
		hash := entry.Hash
		if len(hash) > 12 {
			hash = hash[:12] + "..."
		}

		cmd := truncate(entry.Command, 37)
		size := formatBytes(entry.CompressedSize)
		created := formatTime(entry.CreatedAt)
		accessed := "-"
		if entry.AccessedAt != nil {
			accessed = formatTime(*entry.AccessedAt)
		}

		fmt.Printf("%-16s %-40s %-12s %-10s %-15s %-15s\n",
			hash,
			cmd,
			string(entry.Category),
			size,
			created,
			accessed,
		)
	}

	// Footer
	if result.HasMore {
		fmt.Printf("\n%s Use --offset=%d to see more\n",
			color.YellowString("→"),
			listOffset+listLimit)
	}

	return nil
}

func outputJSON(result *archive.ArchiveListResult) error {
	// Simple JSON output
	fmt.Printf(`{"total":%d,"has_more":%t,"entries":[`, result.Total, result.HasMore)

	for i, entry := range result.Entries {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf(`{"hash":"%s","command":"%s","category":"%s","size":%d,"created_at":"%s"}`,
			entry.Hash,
			escapeJSON(entry.Command),
			entry.Category,
			entry.CompressedSize,
			entry.CreatedAt.Format(time.RFC3339))
	}

	fmt.Println("]}")
	return nil
}

func outputCSV(result *archive.ArchiveListResult) error {
	// Header
	fmt.Println("hash,command,category,original_size,compressed_size,created_at,accessed_at,access_count")

	for _, entry := range result.Entries {
		accessed := ""
		if entry.AccessedAt != nil {
			accessed = entry.AccessedAt.Format(time.RFC3339)
		}

		fmt.Printf("%s,%s,%s,%d,%d,%s,%s,%d\n",
			entry.Hash,
			escapeCSV(entry.Command),
			entry.Category,
			entry.OriginalSize,
			entry.CompressedSize,
			entry.CreatedAt.Format(time.RFC3339),
			accessed,
			entry.AccessCount)
	}

	return nil
}

func outputSimple(result *archive.ArchiveListResult) error {
	for _, entry := range result.Entries {
		fmt.Println(entry.Hash)
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	case duration < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"") {
		s = strings.ReplaceAll(s, `"`, `""`)
		s = fmt.Sprintf(`"%s"`, s)
	}
	return s
}
