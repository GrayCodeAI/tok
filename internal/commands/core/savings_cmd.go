package core

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var gainCmd = &cobra.Command{
	Use:   "gain",
	Short: "Show token savings",
	Long:  `Display token savings statistics`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := config.DatabasePath()

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Println("No tracking data found.")
			fmt.Println("Run some commands through TokMan to start tracking token savings!")
			fmt.Printf("\nExample: %s git status\n", os.Args[0])
			return nil
		}

		fmt.Println("═══════════════════════════════════════")
		fmt.Println("           TokMan Savings               ")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("  Database: %s\n", dbPath)
		fmt.Println("  (Run more commands to see savings)")
		fmt.Println("═══════════════════════════════════════")

		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(gainCmd) })
}
