package configcmd

import (
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or create configuration file",
	Long:  "",
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		create, _ := cmd.Flags().GetBool("create")

		if create {
			path, err := createDefaultConfig()
			if err != nil {
				return fmt.Errorf("error creating config: %w", err)
			}
			out.Global().Printf("Created: %s\n", path)
			return nil
		}

		showConfig()
		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(configCmd) })
	configCmd.Long = fmt.Sprintf(`Display the current tok configuration or create a default config file.

The configuration file is stored at %s and controls:
- Token tracking behavior
- Output filtering settings
- Shell hook exclusions`, effectiveConfigPath())

	configCmd.Flags().Bool("create", false, "Create default config file")
}

func createDefaultConfig() (string, error) {
	configPath := effectiveConfigPath()

	cfg := config.Defaults()
	if err := cfg.Save(configPath); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return configPath, nil
}

func showConfig() error {
	configPath := effectiveConfigPath()
	out.Global().Printf("Config: %s\n\n", configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		out.Global().Println("(default config, file not created)")
		out.Global().Println()
		cfg := config.Defaults()
		printConfig(cfg)
		return nil
	}

	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	printConfig(cfg)
	return nil
}

func printConfig(cfg *config.Config) {
	out.Global().Println("[tracking]")
	out.Global().Printf("enabled = %v\n", cfg.Tracking.Enabled)
	if cfg.Tracking.DatabasePath != "" {
		out.Global().Printf("database_path = %q\n", cfg.Tracking.DatabasePath)
	}
	out.Global().Printf("telemetry = %v\n", cfg.Tracking.Telemetry)
	out.Global().Println()

	out.Global().Println("[filter]")
	out.Global().Printf("mode = %q\n", cfg.Filter.Mode)
	out.Global().Printf("noise_dirs = %v\n", cfg.Filter.NoiseDirs)
	out.Global().Println()

	out.Global().Println("[hooks]")
	if len(cfg.Hooks.ExcludedCommands) > 0 {
		out.Global().Printf("excluded_commands = %v\n", cfg.Hooks.ExcludedCommands)
	} else {
		out.Global().Println("excluded_commands = []")
	}
}
