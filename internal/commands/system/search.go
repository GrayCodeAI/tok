package system

import (
	"os"
	"path/filepath"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var searchCmd = &cobra.Command{
	Use:   "search <tool>",
	Short: "Search available filters for a command",
	Long:  `Find which TOML filter applies to a given command or tool.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	registry.Add(func() { registry.Register(searchCmd) })
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

	out.Global().Printf("Searching filters for '%s'...\n\n", query)

	// Search built-in filters
	builtinDir := filepath.Join(getTokSourceDir(), "internal", "toml", "builtin")
	found := 0

	if entries, err := os.ReadDir(builtinDir); err == nil {
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".toml") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".toml")
			if strings.Contains(strings.ToLower(name), query) {
				out.Global().Printf("  ✓ %s (built-in)\n", name)
				found++
			}
		}
	}

	// Search user filters
	userDir := config.FiltersDir()
	if entries, err := os.ReadDir(userDir); err == nil {
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".toml") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".toml")
			if strings.Contains(strings.ToLower(name), query) {
				out.Global().Printf("  ✓ %s (user)\n", name)
				found++
			}
		}
	}

	if found == 0 {
		out.Global().Printf("No filters found for '%s'.\n", query)
		out.Global().Println("\nTip: Use 'tok marketplace search <query>' to find community filters.")
		out.Global().Printf("     Or create a custom filter in %s\n", filepath.Join(config.FiltersDir(), "<name>.toml"))
	}

	return nil
}

func getTokSourceDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	dir := filepath.Dir(filepath.Dir(exe))
	builtinDir := filepath.Join(dir, "internal", "toml", "builtin")
	if _, err := os.Stat(builtinDir); err == nil {
		return dir
	}
	dir = filepath.Dir(exe)
	builtinDir = filepath.Join(dir, "internal", "toml", "builtin")
	if _, err := os.Stat(builtinDir); err == nil {
		return dir
	}
	return "."
}
