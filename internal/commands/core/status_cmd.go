package core

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show tokman status",
	Long:  `Display tokman status and configuration`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TokMan: Enabled")
		fmt.Printf("Project: %s\n", config.ProjectPath())
		fmt.Printf("Config: %s\n", config.ConfigPath())

		if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
			fmt.Println("Config: Not found (run 'tokman config --create')")
		} else {
			fmt.Println("Config: Found")
		}

		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(statusCmd) })
}
