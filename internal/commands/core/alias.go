package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage command aliases for common filter+command combos",
	Long: `Create shorthand aliases for frequently used tok commands.

Examples:
  tok alias set gs "git status"
  tok alias set dl "docker logs --tail 50"
  tok alias list`,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	RunE:  runAliasList,
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <name> <command...>",
	Short: "Create or update an alias",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runAliasSet,
}

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an alias",
	Args:  cobra.ExactArgs(1),
	RunE:  runAliasRemove,
}

func init() {
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)
	registry.Add(func() { registry.Register(aliasCmd) })
}

func getAliasPath() string {
	dir := config.ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		out.Global().Errorf("warning: failed to create directory: %v\n", err)
	}
	return filepath.Join(dir, "aliases.txt")
}

func loadAliases() map[string]string {
	aliases := make(map[string]string)
	path := getAliasPath()
	if path == "" {
		return aliases
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return aliases
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			aliases[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return aliases
}

func saveAliases(aliases map[string]string) error {
	path := getAliasPath()
	if path == "" {
		return fmt.Errorf("cannot determine config path")
	}
	var lines []string
	for k, v := range aliases {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

func runAliasList(cmd *cobra.Command, args []string) error {
	aliases := loadAliases()
	if len(aliases) == 0 {
		out.Global().Println("No aliases configured.")
		out.Global().Println("Create one with: tok alias set <name> <command>")
		return nil
	}
	out.Global().Println("Aliases:")
	for name, command := range aliases {
		out.Global().Printf("  %s → %s\n", name, command)
	}
	return nil
}

func runAliasSet(cmd *cobra.Command, args []string) error {
	name := args[0]
	command := strings.Join(args[1:], " ")
	aliases := loadAliases()
	aliases[name] = command
	if err := saveAliases(aliases); err != nil {
		return err
	}
	out.Global().Printf("Alias '%s' → '%s' created.\n", name, command)
	return nil
}

func runAliasRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	aliases := loadAliases()
	if _, ok := aliases[name]; !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}
	delete(aliases, name)
	if err := saveAliases(aliases); err != nil {
		return err
	}
	out.Global().Printf("Alias '%s' removed.\n", name)
	return nil
}
