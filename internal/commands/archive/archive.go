package archive

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var (
	archiveCategory string
	archiveTags     []string
	archiveExpire   string
	archiveAgent    string
)

func init() {
	registry.Add(func() {
		registry.Register(archiveCmd)
	})
}

var archiveCmd = &cobra.Command{
	Use:   "archive [file]",
	Short: "Archive content for later retrieval",
	Long: `Archive content to the RewindStore for later retrieval by hash.

Archives can be created from:
- A file: tokman archive /path/to/file
- Stdin: echo "content" | tokman archive
- Command output: tokman archive --command="ls -la"

Archived content can be retrieved using 'tokman retrieve <hash>'.
Each archive is identified by a SHA-256 hash and includes metadata
such as timestamp, command, and tags.`,
	Example: `  # Archive a file
  tokman archive output.txt

  # Archive with tags
  tokman archive output.txt --tags="important,production"

  # Archive from stdin
  cat output.txt | tokman archive --tags="cat-result"

  # Archive with expiration (30 days)
  tokman archive output.txt --expire="720h"

  # Archive with specific category
  tokman archive output.txt --category=session`,
	RunE: runArchive,
}

func init() {
	archiveCmd.Flags().StringVar(&archiveCategory, "category", "command", "Archive category (command, session, user, system)")
	archiveCmd.Flags().StringSliceVar(&archiveTags, "tags", []string{}, "Tags for the archive (comma-separated)")
	archiveCmd.Flags().StringVar(&archiveExpire, "expire", "2160h", "Expiration duration (default: 90 days)")
	archiveCmd.Flags().StringVar(&archiveAgent, "agent", "", "Agent that created this archive")
}

func runArchive(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse expiration
	expireDuration, err := time.ParseDuration(archiveExpire)
	if err != nil {
		return fmt.Errorf("invalid expiration duration: %w", err)
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

	// Read content
	var content []byte
	var source string

	if len(args) > 0 {
		// Read from file
		content, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		source = args[0]
	} else {
		// Read from stdin
		content, err = readStdin()
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		source = "stdin"
	}

	// Create archive entry
	entry := archive.NewArchiveEntry(content, source).
		WithCategory(archive.ArchiveCategory(archiveCategory)).
		WithTags(archiveTags...).
		WithExpiration(expireDuration)

	if archiveAgent != "" {
		entry.WithAgent(archiveAgent)
	}

	// Get working directory
	wd, _ := os.Getwd()
	entry.WithWorkingDirectory(wd)

	// Get project path
	entry.ProjectPath = config.ProjectPath()

	// Archive it
	hash, err := mgr.Archive(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to archive: %w", err)
	}

	// Print result
	fmt.Printf("✓ Archived %s\n", source)
	fmt.Printf("  Hash: %s\n", hash)
	fmt.Printf("  Size: %s (original: %s)\n",
		formatBytes(entry.CompressedSize),
		formatBytes(entry.OriginalSize))
	if len(archiveTags) > 0 {
		fmt.Printf("  Tags: %v\n", archiveTags)
	}

	return nil
}

func readStdin() ([]byte, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if info.Mode()&os.ModeNamedPipe == 0 {
		return nil, fmt.Errorf("no input provided (use pipe or specify a file)")
	}

	return os.ReadFile("/dev/stdin")
}
