package archive

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var verifyAll bool

func init() {
	registry.Add(func() {
		registry.Register(verifyCmd)
	})
}

var verifyCmd = &cobra.Command{
	Use:     "archive-verify <hash>",
	Short:   "Verify archive integrity by hash",
	Long:    `Verify that an archive's content matches its SHA-256 hash.`,
	Example: `  tokman archive-verify e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
	Args:    cobra.RangeArgs(0, 1),
	RunE:    runVerify,
}

func init() {
	verifyCmd.Flags().BoolVar(&verifyAll, "all", false, "Verify all archives")
}

func runVerify(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if verifyAll {
		// TODO: Implement verify all
		fmt.Println(color.YellowString("Verifying all archives..."))
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("please specify a hash or use --all")
	}

	hash := args[0]

	if !archive.IsValidHash(hash) {
		return fmt.Errorf("invalid hash format")
	}

	valid, err := mgr.Verify(ctx, hash)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if valid {
		fmt.Printf("%s Archive %s is valid\n", color.GreenString("✓"), hash[:16])
	} else {
		fmt.Printf("%s Archive %s is corrupted\n", color.RedString("✗"), hash[:16])
	}

	return nil
}
