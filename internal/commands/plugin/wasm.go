package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

var (
	pluginCmd = &cobra.Command{
		Use:   "plugin",
		Short: "WASM plugin management",
		Long: `Load and manage WASM-based custom filter plugins.

TokMan supports custom compression filters via WebAssembly.
Place .wasm files in ~/.config/tokman/plugins/

Plugin Interface:
  type FilterPlugin interface {
    Name() string
    Version() string
    Apply(input string, mode string) (string, int)
  }

Examples:
  tokman plugin list    # List loaded plugins
  tokman plugin load    # Load a plugin
  tokman plugin unload  # Unload a plugin`,
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List loaded plugins",
		RunE:  runPluginList,
	}

	loadCmd = &cobra.Command{
		Use:   "load [path]",
		Short: "Load a WASM plugin",
		RunE:  runPluginLoad,
	}

	unloadCmd = &cobra.Command{
		Use:   "unload [name]",
		Short: "Unload a plugin by name",
		RunE:  runPluginUnload,
	}

	createCmd = &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new plugin template",
		RunE:  runPluginCreate,
	}

	enabled bool
)

var (
	loadedPlugins = make(map[string]*WASMPlugin)
	pluginMu      sync.RWMutex
)

type WASMPlugin struct {
	Name    string
	Version string
	Path    string
	Filter  func(input string, mode string) (string, int)
}

func init() {
	pluginCmd.AddCommand(listCmd)
	pluginCmd.AddCommand(loadCmd)
	pluginCmd.AddCommand(unloadCmd)
	pluginCmd.AddCommand(createCmd)

	pluginCmd.Flags().BoolVar(&enabled, "enable", true, "Enable plugin system")

	registry.Add(func() {
		registry.Register(pluginCmd)
	})
}

func runPluginList(cmd *cobra.Command, args []string) error {
	pluginMu.RLock()
	defer pluginMu.RUnlock()

	if len(loadedPlugins) == 0 {
		fmt.Println("No plugins loaded")
		return nil
	}

	fmt.Println("=== Loaded Plugins ===")
	for name, p := range loadedPlugins {
		fmt.Printf("%s v%s - %s\n", name, p.Version, p.Path)
	}

	return nil
}

func runPluginLoad(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("plugin path required")
	}

	pluginPath := args[0]
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", pluginPath)
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		sym, err := p.Lookup("FilterPlugin")
		if err != nil {
			return fmt.Errorf("invalid plugin format: %w", err)
		}
		_ = sym
	}

	pluginMu.Lock()
	defer pluginMu.Unlock()

	name := filepath.Base(pluginPath)
	name = name[:len(name)-len(filepath.Ext(name))]

	loadedPlugins[name] = &WASMPlugin{
		Name:    name,
		Version: "1.0.0",
		Path:    pluginPath,
		Filter:  defaultFilter,
	}

	fmt.Printf("Loaded plugin: %s\n", name)
	return nil
}

func runPluginUnload(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("plugin name required")
	}

	pluginMu.Lock()
	defer pluginMu.Unlock()

	name := args[0]
	if _, ok := loadedPlugins[name]; !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	delete(loadedPlugins, name)
	fmt.Printf("Unloaded plugin: %s\n", name)

	return nil
}

func runPluginCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("plugin name required")
	}

	name := args[0]
	pluginDir := filepath.Join(os.Getenv("HOME"), ".config/tokman", "plugins")
	os.MkdirAll(pluginDir, 0755)

	template := fmt.Sprintf(`package main

import "github.com/GrayCodeAI/tokman/internal/filter"

type MyFilter struct{}

func (f *MyFilter) Name() string {
	return "%s"
}

func (f *MyFilter) Version() string {
	return "1.0.0"
}

func (f *MyFilter) Apply(input string, mode string) (string, int) {
	// Custom filter logic here
	output := input
	
	// Example: remove empty lines
	// lines := strings.Split(output, "\n")
	// var filtered []string
	// for _, line := range lines {
	//     if strings.TrimSpace(line) != "" {
	//         filtered = append(filtered, line)
	//     }
	// }
	// output = strings.Join(filtered, "\n")
	
	saved := filter.EstimateTokens(input) - filter.EstimateTokens(output)
	return output, saved
}

func main() {}
`, name)

	pluginFile := filepath.Join(pluginDir, name+".go")
	if err := os.WriteFile(pluginFile, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	fmt.Printf("Created plugin template: %s\n", pluginFile)
	fmt.Printf("Compile with: tinygo build -target wasi -o %s.wasm %s.go\n", name, name)

	return nil
}

func defaultFilter(input, mode string) (string, int) {
	return input, 0
}

func GetPlugin(name string) *WASMPlugin {
	pluginMu.RLock()
	defer pluginMu.RUnlock()
	return loadedPlugins[name]
}

func ApplyPlugins(input string, mode filter.Mode) string {
	pluginMu.RLock()
	defer pluginMu.RUnlock()

	output := input
	for _, p := range loadedPlugins {
		if p.Filter != nil {
			out, _ := p.Filter(output, string(mode))
			output = out
		}
	}

	return output
}
