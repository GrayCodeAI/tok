// Package luau provides filter definitions in TOML format.
package luau

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// FilterDef represents a filter definition from TOML.
type FilterDef struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Version     string            `toml:"version"`
	Author      string            `toml:"author"`
	Script      string            `toml:"script"`
	ScriptFile  string            `toml:"script_file"`
	EntryPoint  string            `toml:"entry_point"`
	Config      map[string]interface{} `toml:"config"`
	Enabled     bool              `toml:"enabled"`
	Priority    int               `toml:"priority"`
	Tags        []string          `toml:"tags"`
}

// FilterRegistry manages loaded filters.
type FilterRegistry struct {
	filters map[string]*FilterDef
	vmPool  *Pool
}

// NewFilterRegistry creates a new filter registry.
func NewFilterRegistry(vmPool *Pool) *FilterRegistry {
	return &FilterRegistry{
		filters: make(map[string]*FilterDef),
		vmPool:  vmPool,
	}
}

// LoadFromFile loads a filter from a TOML file.
func (r *FilterRegistry) LoadFromFile(path string) (*FilterDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read filter file: %w", err)
	}

	var def FilterDef
	if err := toml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Load script from file if specified
	if def.ScriptFile != "" {
		scriptPath := filepath.Join(filepath.Dir(path), def.ScriptFile)
		scriptData, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read script file: %w", err)
		}
		def.Script = string(scriptData)
	}

	// Validate
	if def.Name == "" {
		return nil, fmt.Errorf("filter name is required")
	}
	if def.Script == "" {
		return nil, fmt.Errorf("filter script is required")
	}

	// Default to enabled
	if !def.Enabled && def.Enabled == false {
		// Only disable if explicitly set to false
	}

	def.Enabled = true

	// Register
	r.filters[def.Name] = &def

	return &def, nil
}

// LoadFromDir loads all filters from a directory.
func (r *FilterRegistry) LoadFromDir(dir string) ([]*FilterDef, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var loaded []*FilterDef
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !hasExtension(entry.Name(), ".toml") {
			continue
		}

		def, err := r.LoadFromFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue // Skip invalid filters
		}
		loaded = append(loaded, def)
	}

	return loaded, nil
}

// Get gets a filter by name.
func (r *FilterRegistry) Get(name string) (*FilterDef, bool) {
	def, ok := r.filters[name]
	return def, ok
}

// List returns all filter names.
func (r *FilterRegistry) List() []string {
	names := make([]string, 0, len(r.filters))
	for name := range r.filters {
		names = append(names, name)
	}
	return names
}

// Apply applies a filter to content.
func (r *FilterRegistry) Apply(name string, content string) (string, error) {
	def, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("filter not found: %s", name)
	}

	if !def.Enabled {
		return content, nil // Filter disabled, pass through
	}

	// Get VM from pool
	vm := r.vmPool.Get()
	defer r.vmPool.Put(vm)

	// Execute filter
	return vm.ExecuteFilter(def.Script, content)
}

// ApplyAll applies all enabled filters in priority order.
func (r *FilterRegistry) ApplyAll(content string) (string, error) {
	// Sort by priority
	type filterWithPriority struct {
		name     string
		priority int
	}

	var filters []filterWithPriority
	for name, def := range r.filters {
		if def.Enabled {
			filters = append(filters, filterWithPriority{name, def.Priority})
		}
	}

	// Sort (lower priority = earlier)
	for i := 0; i < len(filters); i++ {
		for j := i + 1; j < len(filters); j++ {
			if filters[j].priority < filters[i].priority {
				filters[i], filters[j] = filters[j], filters[i]
			}
		}
	}

	// Apply filters
	result := content
	for _, fp := range filters {
		var err error
		result, err = r.Apply(fp.name, result)
		if err != nil {
			return result, fmt.Errorf("filter %s failed: %w", fp.name, err)
		}
	}

	return result, nil
}

// Example TOML filter definition:
//
// name = "strip_timestamps"
// description = "Remove timestamps from log lines"
// version = "1.0.0"
// author = "user"
//
// script = """
// -- Remove ISO8601 timestamps
// OUTPUT = INPUT:gsub("%d%d%d%d%-%d%d%-%d%dT%d%d:%d%d:%d%d%.?%d*Z?", "[TIMESTAMP]")
// """
//
// enabled = true
// priority = 10
// tags = ["logs", "timestamp"]

// hasExtension checks if filename has the given extension.
func hasExtension(filename, ext string) bool {
	if len(filename) < len(ext) {
		return false
	}
	return filename[len(filename)-len(ext):] == ext
}
