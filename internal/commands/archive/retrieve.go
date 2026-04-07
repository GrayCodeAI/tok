package archive

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	retrieveOutput string
	retrieveRaw    bool
	retrieveInfo   bool
)

func init() {
	registry.Add(func() {
		registry.Register(retrieveCmd)
	})
}

var retrieveCmd = &cobra.Command{
	Use:   "retrieve <hash>",
	Short: "Retrieve archived content by hash",
	Long: `Retrieve previously archived content using its SHA-256 hash.

The hash is returned when archiving content with 'tokman archive'.
You can retrieve:
- The filtered content (default)
- The original content (--raw)
- Just metadata (--info)

The content is written to stdout by default, or to a file with --output.`,
	Example: `  # Retrieve by hash
  tokman retrieve e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

  # Retrieve original (unfiltered) content
  tokman retrieve <hash> --raw

  # Save to file
  tokman retrieve <hash> --output=output.txt

  # Show only metadata
  tokman retrieve <hash> --info`,
	Args: cobra.ExactArgs(1),
	RunE: runRetrieve,
}

func init() {
	retrieveCmd.Flags().StringVarP(&retrieveOutput, "output", "o", "", "Output file (default: stdout)")
	retrieveCmd.Flags().BoolVar(&retrieveRaw, "raw", false, "Retrieve original (unfiltered) content")
	retrieveCmd.Flags().BoolVar(&retrieveInfo, "info", false, "Show only metadata")
}

func runRetrieve(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	hash := args[0]

	// Validate hash
	if !archive.IsValidHash(hash) {
		return fmt.Errorf("invalid hash format: expected 64-character hex string")
	}

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

	// Retrieve archive
	entry, err := mgr.Retrieve(ctx, hash)
	if err != nil {
		return fmt.Errorf("failed to retrieve archive: %w", err)
	}

	// Show info only
	if retrieveInfo {
		printArchiveInfo(entry)
		return nil
	}

	// Get content to output
	var content []byte
	if retrieveRaw {
		content = entry.OriginalContent
	} else {
		content = entry.FilteredContent
		if content == nil {
			content = entry.OriginalContent
		}
	}

	// Output content
	if retrieveOutput != "" {
		if err := os.WriteFile(retrieveOutput, content, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("✓ Retrieved to %s\n", retrieveOutput)
	} else {
		// Write to stdout
		_, err := os.Stdout.Write(content)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}

func printArchiveInfo(entry *archive.ArchiveEntry) {
	fmt.Printf("Archive: %s\n", entry.Hash)
	fmt.Printf("  ID: %d\n", entry.ID)
	fmt.Printf("  Command: %s\n", entry.Command)
	if entry.WorkingDirectory != "" {
		fmt.Printf("  Working Directory: %s\n", entry.WorkingDirectory)
	}
	if entry.ProjectPath != "" {
		fmt.Printf("  Project Path: %s\n", entry.ProjectPath)
	}
	if entry.Agent != "" {
		fmt.Printf("  Agent: %s\n", entry.Agent)
	}
	fmt.Printf("  Category: %s\n", entry.Category)
	fmt.Printf("  Original Size: %s\n", formatBytes(entry.OriginalSize))
	fmt.Printf("  Compressed Size: %s\n", formatBytes(entry.CompressedSize))
	fmt.Printf("  Compression Ratio: %.2f%%\n", (1-entry.CompressionRatio())*100)
	fmt.Printf("  Created: %s\n", entry.CreatedAt.Format(time.RFC3339))
	if entry.AccessedAt != nil {
		fmt.Printf("  Last Accessed: %s\n", entry.AccessedAt.Format(time.RFC3339))
	}
	fmt.Printf("  Access Count: %d\n", entry.AccessCount)
	if entry.ExpiresAt != nil {
		fmt.Printf("  Expires: %s\n", entry.ExpiresAt.Format(time.RFC3339))
	}
	if len(entry.Tags) > 0 {
		fmt.Printf("  Tags: %v\n", entry.Tags)
	}
}
