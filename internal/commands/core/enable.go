package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable global tok interception",
	Long: `Enable tok to automatically intercept and compress all CLI output.

When enabled, shell hooks will intercept commands and route them through
tok's compression pipeline. Use 'tok disable' to turn off.

Examples:
  tok enable        # Turn on automatic compression
  tok disable       # Turn off automatic compression
  tok status        # Check if tok is enabled`,
	RunE: func(cmd *cobra.Command, args []string) error {
		green := color.New(color.FgGreen).SprintFunc()

		markerPath := getEnabledMarkerPath()
		markerDir := filepath.Dir(markerPath)

		// Ensure directory exists
		if err := os.MkdirAll(markerDir, 0700); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		// Check if already enabled
		if isEnabled() {
			fmt.Printf("%s tok is already enabled\n", green("✓"))
			return nil
		}

		// Create marker file
		if err := os.WriteFile(markerPath, []byte("enabled\n"), 0600); err != nil {
			return fmt.Errorf("error enabling tok: %w", err)
		}

		fmt.Printf("%s tok enabled globally\n", green("✓"))
		fmt.Println()
		fmt.Println("All commands will now be automatically compressed.")
		fmt.Println("Run 'tok disable' to turn off.")
		return nil
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable global tok interception",
	Long: `Disable tok interception. Commands will run normally without compression.

Use 'tok enable' to turn interception back on.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		markerPath := getEnabledMarkerPath()

		if !isEnabled() {
			fmt.Printf("%s tok is already disabled\n", green("✓"))
			return nil
		}

		if err := os.Remove(markerPath); err != nil {
			return fmt.Errorf("error disabling tok: %w", err)
		}

		fmt.Printf("%s tok disabled\n", red("✗"))
		fmt.Println()
		fmt.Println("Commands will run normally without compression.")
		fmt.Println("Run 'tok enable' to turn back on.")
		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(enableCmd) })
	registry.Add(func() { registry.Register(disableCmd) })
}

// getEnabledMarkerPath returns the path to the enabled marker file.
func getEnabledMarkerPath() string {
	return filepath.Join(config.DataPath(), ".enabled")
}

// isEnabled checks if tok is globally enabled.
func isEnabled() bool {
	markerPath := getEnabledMarkerPath()
	_, err := os.Stat(markerPath)
	return err == nil
}
