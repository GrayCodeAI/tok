package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var attributionCmd = &cobra.Command{
	Use:   "attribution",
	Short: "Manage commit attribution (Co-Authored-By)",
	Long: `Configure AI attribution on git commits.

When enabled, adds "Co-Authored-By: tok <tok@ai>" to commit messages
to credit the AI assistant for code contributions.

Examples:
  tok attribution                 # Show current settings
  tok attribution enable          # Enable Co-Authored-By
  tok attribution disable         # Disable attribution
  tok attribution set "AI <ai>"   # Set custom author`,
	RunE: runAttribution,
}

var attrEnable, attrDisable bool

func init() {
	registry.Add(func() { registry.Register(attributionCmd) })

	attributionCmd.Flags().BoolVar(&attrEnable, "enable", false, "Enable attribution")
	attributionCmd.Flags().BoolVar(&attrDisable, "disable", false, "Disable attribution")
}

func runAttribution(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	enabled := cfg.Pipeline.EnableAttribution
	author := "tok <tok@ai>" // Default author

	if attrEnable {
		cfg.Pipeline.EnableAttribution = true
		if err := cfg.Save(config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		out.Global().Println(color.GreenString("✓"), "Attribution enabled")
		return nil
	}

	if attrDisable {
		cfg.Pipeline.EnableAttribution = false
		if err := cfg.Save(config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		out.Global().Println(color.GreenString("✓"), "Attribution disabled")
		return nil
	}

	if len(args) > 0 {
		if args[0] == "set" && len(args) > 1 {
			out.Global().Printf("Custom author will be set: %s\n", args[1])
			out.Global().Println("(Custom author support coming soon)")
			return nil
		}
	}

	out.Global().Println("Attribution Settings:")
	if enabled {
		out.Global().Println("  Enabled:", color.GreenString("Yes"))
	} else {
		out.Global().Println("  Enabled:", color.RedString("No"))
	}
	out.Global().Println("  Author:", author)
	out.Global().Println()
	out.Global().Println("Commands:")
	out.Global().Println("  tok attribution enable   - Enable Co-Authored-By")
	out.Global().Println("  tok attribution disable  - Disable attribution")

	return nil
}

func ApplyAttribution(msg string, cfg *config.Config) string {
	if !cfg.Pipeline.EnableAttribution {
		return msg
	}

	author := "tok <tok@ai>"

	coAuthor := fmt.Sprintf("\n\nCo-Authored-By: %s", author)
	if !strings.Contains(msg, coAuthor) {
		msg = msg + coAuthor
	}
	return msg
}

func AmendCommitWithAttribution(cfg *config.Config) error {
	if !cfg.Pipeline.EnableAttribution {
		return nil
	}

	author := "tok <tok@ai>"

	cmd := exec.Command("git", "commit", "--amend", "--no-edit")
	env := os.Environ()
	env = append(env, "GIT_AUTHOR_NAME="+author)
	env = append(env, "GIT_COMMITTER_NAME="+author)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to amend commit: %w\n%s", err, output)
	}
	return nil
}
