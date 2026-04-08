// Package wasm provides a WebAssembly-based plugin system for custom filters
package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/GrayCodeAI/tokman/internal/config"
)

// Manager manages WASM plugins
type Manager struct {
	runtime   wazero.Runtime
	plugins   map[string]*Plugin
	pluginDir string
	maxMemory uint32
	timeout   time.Duration
}

// Plugin represents a WASM plugin
type Plugin struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	WasmPath    string                 `json:"wasm_path"`
	Checksum    string                 `json:"checksum"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	// Runtime fields
	compiledModule wazero.CompiledModule
	moduleConfig   wazero.ModuleConfig
}

// FilterInput represents input to a WASM filter
type FilterInput struct {
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Options     map[string]interface{} `json:"options"`
}

// FilterOutput represents output from a WASM filter
type FilterOutput struct {
	Content     string  `json:"content"`
	TokensSaved int     `json:"tokens_saved"`
	Compression float64 `json:"compression_ratio"`
	Error       string  `json:"error,omitempty"`
}

// NewManager creates a new WASM plugin manager
func NewManager() (*Manager, error) {
	pluginDir := filepath.Join(config.DataPath(), "plugins", "wasm")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Create WASM runtime
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)

	// Instantiate WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	return &Manager{
		runtime:   runtime,
		plugins:   make(map[string]*Plugin),
		pluginDir: pluginDir,
		maxMemory: 128 * 1024 * 1024, // 128 MB
		timeout:   30 * time.Second,
	}, nil
}

// Close closes the WASM runtime
func (m *Manager) Close(ctx context.Context) error {
	return m.runtime.Close(ctx)
}

// LoadPlugin loads a WASM plugin from disk
func (m *Manager) LoadPlugin(id string) (*Plugin, error) {
	// Check if already loaded
	if plugin, exists := m.plugins[id]; exists {
		return plugin, nil
	}

	// Load plugin metadata
	pluginPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	var plugin Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	// Load and compile WASM module
	wasmPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.wasm", id))
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}

	// Verify checksum
	checksum := calculateChecksum(wasmBytes)
	if checksum != plugin.Checksum {
		return nil, fmt.Errorf("WASM checksum mismatch: expected %s, got %s", plugin.Checksum, checksum)
	}

	// Compile module
	ctx := context.Background()
	compiled, err := m.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	plugin.compiledModule = compiled
	plugin.moduleConfig = wazero.NewModuleConfig().
		WithName(id).
		WithStderr(io.Discard)

	// Cache plugin
	m.plugins[id] = &plugin

	slog.Info("Loaded WASM plugin", "id", id, "name", plugin.Name)
	return &plugin, nil
}

// UnloadPlugin unloads a plugin from memory
func (m *Manager) UnloadPlugin(id string) error {
	plugin, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", id)
	}

	// Close the module if it has a close method
	// Note: wazero doesn't require explicit module closing

	delete(m.plugins, id)
	slog.Info("Unloaded WASM plugin", "id", id)
	return nil
}

// ExecuteFilter executes a WASM filter plugin
func (m *Manager) ExecuteFilter(pluginID string, input FilterInput) (*FilterOutput, error) {
	plugin, err := m.LoadPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	if !plugin.Enabled {
		return nil, fmt.Errorf("plugin is disabled: %s", pluginID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	// Instantiate module
	module, err := m.runtime.InstantiateModule(ctx, plugin.compiledModule, plugin.moduleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}
	defer module.Close(ctx)

	// Serialize input
	inputData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Allocate memory for input
	inputLen := uint64(len(inputData))
	malloc := module.ExportedFunction("malloc")
	if malloc == nil {
		return nil, fmt.Errorf("WASM module missing malloc function")
	}

	inputPtr, err := malloc.Call(ctx, inputLen)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory: %w", err)
	}

	// Write input to WASM memory
	memory := module.Memory()
	if memory == nil {
		return nil, fmt.Errorf("WASM module has no memory")
	}

	if !memory.Write(uint32(inputPtr[0]), inputData) {
		return nil, fmt.Errorf("failed to write input to WASM memory")
	}

	// Call filter function
	filterFn := module.ExportedFunction("filter")
	if filterFn == nil {
		return nil, fmt.Errorf("WASM module missing filter function")
	}

	result, err := filterFn.Call(ctx, inputPtr[0], inputLen)
	if err != nil {
		return nil, fmt.Errorf("filter execution failed: %w", err)
	}

	// Read output from WASM memory
	outputPtr := uint32(result[0])
	outputLen := uint32(result[1])

	outputData, ok := memory.Read(outputPtr, outputLen)
	if !ok {
		return nil, fmt.Errorf("failed to read output from WASM memory")
	}

	// Free allocated memory
	free := module.ExportedFunction("free")
	if free != nil {
		free.Call(ctx, inputPtr[0])
		free.Call(ctx, uint64(outputPtr))
	}

	// Parse output
	var output FilterOutput
	if err := json.Unmarshal(outputData, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	if output.Error != "" {
		return &output, fmt.Errorf("filter error: %s", output.Error)
	}

	return &output, nil
}

// InstallPlugin installs a new plugin
func (m *Manager) InstallPlugin(metadata *Plugin, wasmData []byte) error {
	// Validate metadata
	if metadata.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if metadata.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	// Calculate checksum
	metadata.Checksum = calculateChecksum(wasmData)
	metadata.WasmPath = filepath.Join(m.pluginDir, fmt.Sprintf("%s.wasm", metadata.ID))
	metadata.CreatedAt = time.Now()
	metadata.UpdatedAt = time.Now()
	metadata.Enabled = true

	// Save WASM file
	if err := os.WriteFile(metadata.WasmPath, wasmData, 0644); err != nil {
		return fmt.Errorf("failed to save WASM file: %w", err)
	}

	// Save metadata
	metadataPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.json", metadata.ID))
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	slog.Info("Installed WASM plugin", "id", metadata.ID, "name", metadata.Name)
	return nil
}

// UninstallPlugin removes a plugin
func (m *Manager) UninstallPlugin(id string) error {
	// Unload if loaded
	if _, exists := m.plugins[id]; exists {
		m.UnloadPlugin(id)
	}

	// Remove files
	wasmPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.wasm", id))
	metadataPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.json", id))

	if err := os.Remove(wasmPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove WASM file: %w", err)
	}

	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	slog.Info("Uninstalled WASM plugin", "id", id)
	return nil
}

// EnablePlugin enables a plugin
func (m *Manager) EnablePlugin(id string) error {
	plugin, err := m.LoadPlugin(id)
	if err != nil {
		return err
	}

	plugin.Enabled = true
	plugin.UpdatedAt = time.Now()

	return m.saveMetadata(plugin)
}

// DisablePlugin disables a plugin
func (m *Manager) DisablePlugin(id string) error {
	plugin, err := m.LoadPlugin(id)
	if err != nil {
		return err
	}

	plugin.Enabled = false
	plugin.UpdatedAt = time.Now()

	return m.saveMetadata(plugin)
}

// ListPlugins lists all installed plugins
func (m *Manager) ListPlugins() ([]*Plugin, error) {
	entries, err := os.ReadDir(m.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	var plugins []*Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .json
		plugin, err := m.LoadPlugin(id)
		if err != nil {
			slog.Warn("Failed to load plugin", "id", id, "error", err)
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetPlugin gets a specific plugin
func (m *Manager) GetPlugin(id string) (*Plugin, error) {
	return m.LoadPlugin(id)
}

// UpdatePluginConfig updates plugin configuration
func (m *Manager) UpdatePluginConfig(id string, config map[string]interface{}) error {
	plugin, err := m.LoadPlugin(id)
	if err != nil {
		return err
	}

	plugin.Config = config
	plugin.UpdatedAt = time.Now()

	return m.saveMetadata(plugin)
}

// ValidatePlugin validates a WASM plugin without installing
func (m *Manager) ValidatePlugin(wasmData []byte) error {
	ctx := context.Background()

	// Try to compile
	compiled, err := m.runtime.CompileModule(ctx, wasmData)
	if err != nil {
		return fmt.Errorf("failed to compile WASM: %w", err)
	}

	// Check for required exports
	if compiled.ExportedFunction("malloc") == nil {
		return fmt.Errorf("WASM module missing malloc function")
	}

	if compiled.ExportedFunction("free") == nil {
		return fmt.Errorf("WASM module missing free function")
	}

	if compiled.ExportedFunction("filter") == nil {
		return fmt.Errorf("WASM module missing filter function")
	}

	// Check for memory export
	if compiled.ExportedMemory("memory") == nil {
		return fmt.Errorf("WASM module missing memory export")
	}

	return nil
}

// GetStats returns plugin statistics
func (m *Manager) GetStats() PluginStats {
	stats := PluginStats{
		TotalPlugins: len(m.plugins),
	}

	for _, plugin := range m.plugins {
		if plugin.Enabled {
			stats.EnabledPlugins++
		}
	}

	return stats
}

// PluginStats holds plugin statistics
type PluginStats struct {
	TotalPlugins   int `json:"total_plugins"`
	EnabledPlugins int `json:"enabled_plugins"`
}

// Helper functions

func (m *Manager) saveMetadata(plugin *Plugin) error {
	metadataPath := filepath.Join(m.pluginDir, fmt.Sprintf("%s.json", plugin.ID))
	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
