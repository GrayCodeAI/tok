package archive

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var cleanupDryRun bool

func init() {
	registry.Add(func() {
		registry.Register(cleanupCmd)
	})
}

var cleanupCmd = &cobra.Command{
	Use:   "archive-cleanup",
	Short: "Clean up expired archives",
	Long:  `Remove archives that have passed their expiration date.`,
	Example: `  tokman archive-cleanup
  tokman archive-cleanup --dry-run`,
	RunE: runCleanup,
}

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupDryRun, "dry-run", false, "Show what would be deleted without actually deleting")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if cleanupDryRun {
		fmt.Println(color.YellowString("Dry run mode - no archives will be deleted"))
		// TODO: Implement dry-run query
		return nil
	}

	deleted, err := mgr.CleanupExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup: %w", err)
	}

	if deleted == 0 {
		fmt.Println("No expired archives found")
	} else {
		fmt.Printf("%s Cleaned up %d expired %s\n",
			color.GreenString("✓"),
			deleted,
			pluralize(int(deleted), "archive", "archives"))
	}

	return nil
}
