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
		registry.Register(importCmd)
	})
}

var importCmd = &cobra.Command{
	Use:   "archive-import <file>",
	Short: "Import archives from file",
	Long: `Import archives from a previously exported file.

Supports formats:
- .json: JSON export
- .tar: TAR export
- .tar.gz: Compressed TAR export

Duplicate archives (by hash) will be skipped.`,
	Example: `  tokman archive-import backup.json
  tokman archive-import backup.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	filepath := args[0]

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	importer := archive.NewImporter(mgr)

	fmt.Printf("Importing from %s...\n", filepath)

	result, err := importer.ImportFromFile(ctx, filepath)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	fmt.Printf("\n%s Import complete\n\n", color.GreenString("✓"))
	fmt.Printf("  Imported: %d\n", result.Imported)
	fmt.Printf("  Skipped:  %d (duplicates)\n", result.Skipped)
	fmt.Printf("  Errors:   %d\n", result.Errors)
	fmt.Printf("  Total:    %s\n", formatBytes(result.TotalSize))

	return nil
}
