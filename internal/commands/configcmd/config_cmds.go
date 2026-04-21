package configcmd

import (
	"fmt"
	"os"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/config"
)

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file",
	RunE:  runConfigInit,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := shared.GetConfig()
	if err != nil {
		cfg = config.Defaults()
	}
	out.Global().Println("Current Configuration:")
	out.Global().Println("=====================")
	out.Global().Printf("  Pipeline:\n")
	out.Global().Printf("    max_context_tokens: %d\n", cfg.Pipeline.MaxContextTokens)
	out.Global().Printf("    default_budget: %d\n", cfg.Pipeline.DefaultBudget)
	out.Global().Printf("    entropy_threshold: %.2f\n", cfg.Pipeline.EntropyThreshold)
	out.Global().Printf("  Filter:\n")
	out.Global().Printf("    mode: %s\n", cfg.Filter.Mode)
	out.Global().Printf("    max_width: %d\n", cfg.Filter.MaxWidth)
	out.Global().Printf("  Tracking:\n")
	out.Global().Printf("    enabled: %v\n", cfg.Tracking.Enabled)
	out.Global().Printf("    database_path: %s\n", cfg.Tracking.DatabasePath)
	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configDir := effectiveConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	configPath := effectiveConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		out.Global().Printf("Config already exists at %s\n", configPath)
		return nil
	}

	defaultConfig := `# tok Configuration
[pipeline]
max_context_tokens = 100000
default_budget = 0
entropy_threshold = 2.0

[filter]
mode = "minimal"
max_width = 0

[tracking]
enabled = true
retention_days = 90
`
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return err
	}

	out.Global().Printf("Created config at %s\n", configPath)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	configDir := effectiveConfigDir()
	configPath := effectiveConfigPath()

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	var lines []string
	if data, err := os.ReadFile(configPath); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	parts := strings.SplitN(key, ".", 2)
	section := ""
	field := key
	if len(parts) == 2 {
		section = parts[0]
		field = parts[1]
	}

	found := false
	inSection := section == ""
	newLines := make([]string, 0, len(lines)+2)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			secName := strings.TrimPrefix(strings.TrimSuffix(trimmed, "]"), "[")
			inSection = secName == section
			newLines = append(newLines, line)
			continue
		}

		if inSection && strings.Contains(trimmed, "=") {
			kv := strings.SplitN(trimmed, "=", 2)
			existingKey := strings.TrimSpace(kv[0])
			if existingKey == field {
				newLines = append(newLines, fmt.Sprintf("%s = %s", field, value))
				found = true
				continue
			}
		}
		newLines = append(newLines, line)
	}

	if !found {
		if section != "" {
			sectionExists := false
			for _, line := range newLines {
				if strings.TrimSpace(line) == fmt.Sprintf("[%s]", section) {
					sectionExists = true
					break
				}
			}
			if !sectionExists {
				newLines = append(newLines, "", fmt.Sprintf("[%s]", section))
			}
			if len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) != "" {
				newLines = append(newLines, "")
			}
			newLines = append(newLines, fmt.Sprintf("%s = %s", field, value))
		} else {
			if len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) != "" {
				newLines = append(newLines, "")
			}
			newLines = append(newLines, fmt.Sprintf("%s = %s", field, value))
		}
	}

	content := strings.Join(newLines, "\n")
	// #nosec G703 -- configPath is resolved via shared.GetConfigPath and points to tok config.
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	out.Global().Printf("Set %s = %s\n", key, value)
	out.Global().Printf("Config: %s\n", configPath)
	return nil
}
