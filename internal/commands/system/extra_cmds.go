package system

import (
	"fmt"
	"os"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
)

func init() {
	registry.Add(func() {
		registry.Register(&cobra.Command{
			Use:   "completion [bash|zsh|fish|powershell]",
			Short: "Generate shell completion scripts",
			Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.

Examples:
  tok completion bash   > /etc/bash_completion.d/tok
  tok completion zsh    > "${fpath[1]}/_tok"
  tok completion fish   > ~/.config/fish/completions/tok.fish
  tok completion powershell > $PROFILE`,
			Args: cobra.ExactArgs(1),
			RunE: runCompletion,
		})

		registry.Register(&cobra.Command{
			Use:   "man <output-dir>",
			Short: "Generate man pages for tok",
			Long: `Generate man pages for tok and write them to the specified directory.

Examples:
  tok man /usr/local/share/man/man1
  tok man ./man`,
			Args: cobra.ExactArgs(1),
			RunE: runMan,
		})

		registry.Register(&cobra.Command{
			Use:   "self-update [version]",
			Short: "Update tok to the latest version",
			Long: `Update tok to the latest version using go install.

Examples:
  tok self-update
  tok self-update v0.29.0`,
			Args: cobra.MaximumNArgs(1),
			RunE: runSelfUpdate,
		})
	})
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]
	root := cmd.Root()

	var err error
	switch shell {
	case "bash":
		err = root.GenBashCompletion(os.Stdout)
	case "zsh":
		err = root.GenZshCompletion(os.Stdout)
	case "fish":
		err = root.GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = root.GenPowerShellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}

	if err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", shell, err)
	}
	return nil
}

func runMan(cmd *cobra.Command, args []string) error {
	dir := args[0]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	header := &doc.GenManHeader{
		Title:   "TOK",
		Section: "1",
		Source:  "tok " + cmd.Root().Version,
	}

	if err := doc.GenManTree(cmd.Root(), header, dir); err != nil {
		return fmt.Errorf("failed to generate man pages: %w", err)
	}

	out.Global().Printf("Man pages generated in %s\n", dir)
	return nil
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	version := "latest"
	if len(args) > 0 {
		version = args[0]
	}

	out.Global().Printf("To update tok to %s, run:\n\n", version)
	out.Global().Printf("  go install github.com/lakshmanpatel/tok/cmd/tok@%s\n\n", version)
	out.Global().Println("Or download the latest release from:")
	out.Global().Println("  https://github.com/lakshmanpatel/tok/releases")
	return nil
}
