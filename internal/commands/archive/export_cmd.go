package archive

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	exportFormat      string
	exportOutput      string
	exportCompression bool
	exportRaw         bool
)

func init() {
	registry.Add(func() {
		registry.Register(exportCmd)
	})
}

var exportCmd = &cobra.Command{
	Use:   "archive-export <hash>",
	Short: "Export archives to file",
	Long: `Export one or more archives to a file for backup or sharing.

Supports formats:
- json: Single JSON file (default)
- tar: TAR archive with metadata and content

Can export single archive or all archives matching filters.`,
	Example: `  # Export single archive
  tokman archive-export <hash> --output=backup.json

  # Export as TAR
  tokman archive-export <hash> --format=tar --output=backup.tar.gz

  # Export all archives
  tokman archive-export --all --output=all-archives.json`,
	RunE: runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Export format (json, tar)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (required)")
	exportCmd.Flags().BoolVar(&exportCompression, "compress", true, "Compress output")
	exportCmd.Flags().BoolVar(&exportRaw, "raw", true, "Include raw content")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if exportOutput == "" {
		return fmt.Errorf("--output flag is required")
	}

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	exporter := archive.NewExporter(mgr)
	opts := archive.ExportOptions{
		Format:      archive.ExportFormat(exportFormat),
		Compression: exportCompression,
		IncludeRaw:  exportRaw,
	}

	var data []byte
	var filename string

	if len(args) > 0 {
		// Export single archive
		hash := args[0]
		data, filename, err = exporter.Export(ctx, hash, opts)
	} else {
		// Export all
		data, filename, err = exporter.ExportAll(ctx, opts, archive.DefaultListOptions())
	}

	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	// Write to file
	if err := os.WriteFile(exportOutput, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("%s Exported to %s (%s)\n",
		color.GreenString("✓"),
		exportOutput,
		formatBytes(int64(len(data))))
	fmt.Printf("  Filename: %s\n", filename)

	return nil
}
