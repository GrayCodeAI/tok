package filtercmd

import (
	"fmt"
	"os"
	"path/filepath"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var filterCreateMatch string
var filterCreateDesc string

var filterCreateCmd = &cobra.Command{
	Use:   "filter-create <name>",
	Short: "Create a new TOML filter template",
	Long:  `Generate a TOML filter template for a command.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFilterCreate,
}

func init() {
	filterCreateCmd.Flags().StringVarP(&filterCreateMatch, "match", "m", "", "command pattern to match")
	filterCreateCmd.Flags().StringVarP(&filterCreateDesc, "description", "d", "", "filter description")
	registry.Add(func() { registry.Register(filterCreateCmd) })
}

func runFilterCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	match := filterCreateMatch
	desc := filterCreateDesc

	if match == "" {
		match = name
	}
	if desc == "" {
		desc = name + " output filter"
	}

	template := fmt.Sprintf(`[[rules]]
match = "%s *"
description = "%s"

[rules.filter]
# Keep error/warning lines
keep_lines_matching = [
    ".*error.*",
    ".*warning.*",
    ".*Error.*",
    ".*Warning.*",
]

# Strip verbose output
strip_lines_matching = [
    "^\\s*$",
    "^\\s*\\[.*\\].*",
]

# Max output lines
max_lines = 100

[rules.filter.on_empty]
return = "%s: completed successfully"
`, match, desc, name)

	filterDir := config.FiltersDir()
	if err := os.MkdirAll(filterDir, 0700); err != nil {
		out.Global().Errorf("warning: failed to create directory: %v\n", err)
	}

	filterPath := filepath.Join(filterDir, name+".toml")
	if _, err := os.Stat(filterPath); err == nil {
		return fmt.Errorf("filter '%s' already exists at %s", name, filterPath)
	}

	if err := os.WriteFile(filterPath, []byte(template), 0600); err != nil {
		return err
	}

	out.Global().Printf("Created filter: %s\n", filterPath)
	out.Global().Println("\nEdit the filter to customize:")
	out.Global().Printf("  %s\n", filterPath)
	out.Global().Println("\nTest it with:")
	out.Global().Printf("  tok filter-test %s\n", name)

	return nil
}
