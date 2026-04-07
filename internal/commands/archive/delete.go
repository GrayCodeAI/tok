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
		registry.Register(deleteCmd)
	})
}

var deleteCmd = &cobra.Command{
	Use:     "archive-delete <hash>",
	Aliases: []string{"rm-archive"},
	Short:   "Delete an archive by hash",
	Long:    `Delete an archived entry permanently by its SHA-256 hash.`,
	Example: `  tokman archive-delete e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	hash := args[0]

	if !archive.IsValidHash(hash) {
		return fmt.Errorf("invalid hash format")
	}

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if err := mgr.Delete(ctx, hash); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Printf("%s Archive %s deleted\n", color.GreenString("✓"), hash[:16])
	return nil
}
