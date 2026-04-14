package core

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

// untrustCmd represents the untrust command
var untrustCmd = &cobra.Command{
	Use:   "untrust",
	Short: "Revoke trust for project-local TOML filters",
	Long: `Remove the trust entry for .tokman/filters.toml in the current directory.

This will prevent the project-local filters from being applied until
they are trusted again with 'tokman trust'.

Example:
  tokman untrust  # Revoke trust for current directory`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
	},
	RunE: runUntrust,
}

func init() {
	registry.Add(func() { registry.Register(untrustCmd) })
}

func runUntrust(cmd *cobra.Command, args []string) error {
	filterPath := ".tokman/filters.toml"

	// Try to untrust (file may not exist, but trust entry might)
	removed, err := UntrustFilter(filterPath)
	if err != nil {
		return fmt.Errorf("failed to revoke trust: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if removed {
		fmt.Printf("%s Trust revoked for .tokman/filters.toml\n", green("✓"))
		fmt.Println("Project-local filters will no longer be applied.")
	} else {
		fmt.Printf("%s No trust entry found for current directory.\n", yellow("!"))
	}

	return nil
}
