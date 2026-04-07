package archive

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

func init() {
	registry.Add(func() {
		registry.Register(statsCmd)
	})
}

var statsCmd = &cobra.Command{
	Use:     "archive-stats",
	Aliases: []string{"archive-statistics"},
	Short:   "Show archive statistics",
	Long:    `Display statistics about archived content including total size, compression ratios, and access patterns.`,
	RunE:    runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	stats, err := mgr.Stats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("\n%s\n\n", color.New(color.Bold).Sprint("Archive Statistics"))

	fmt.Printf("Total Archives:      %d\n", stats.TotalArchives)
	fmt.Printf("Total Tags:          %d\n", stats.TotalTags)
	fmt.Printf("Total Accesses:      %d\n", stats.TotalAccesses)

	fmt.Println()

	fmt.Printf("Original Size:       %s\n", formatBytes(stats.TotalOriginalSize))
	fmt.Printf("Compressed Size:     %s\n", formatBytes(stats.TotalCompressedSize))
	fmt.Printf("Space Saved:         %s (%.1f%%)\n",
		formatBytes(stats.SpaceSaved()),
		(1-stats.CompressionRatio())*100)

	fmt.Println()

	fmt.Printf("Schema Version:      %d\n", stats.SchemaVersion)
	fmt.Printf("Last Updated:        %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))

	fmt.Println()
	return nil
}
