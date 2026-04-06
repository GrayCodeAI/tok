package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMPlugin represents a WebAssembly plugin loaded with wazero.
type WASMPlugin struct {
	runtime wazero.Runtime
	module  api.Module
	name    string
	version string
	path    string
}

// LoadWASM loads a WASM plugin from the specified path.
// The WASM module must export the following functions:
//   - plugin_name() -> string
//   - plugin_version() -> string
//   - plugin_apply(input: string) -> (output: string, tokens_saved: int32)
func LoadWASM(path string) (*WASMPlugin, error) {
	ctx := context.Background()

	// Read WASM file
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read WASM file: %w", err)
	}

	// Create runtime with WASI support
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Compile and instantiate the module
	module, err := r.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithName("tokman_plugin"))
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("instantiate WASM module: %w", err)
	}

	plugin := &WASMPlugin{
		runtime: r,
		module:  module,
		path:    path,
	}

	// Get plugin metadata
	if err := plugin.loadMetadata(); err != nil {
		plugin.Close()
		return nil, fmt.Errorf("load plugin metadata: %w", err)
	}

	return plugin, nil
}

// loadMetadata calls the WASM module's metadata functions
func (p *WASMPlugin) loadMetadata() error {
	ctx := context.Background()

	// Get plugin name
	nameFunc := p.module.ExportedFunction("plugin_name")
	if nameFunc != nil {
		results, err := nameFunc.Call(ctx)
		if err != nil {
			return fmt.Errorf("call plugin_name: %w", err)
		}
		if len(results) > 0 {
			// Read string from memory
			p.name = p.readString(uint32(results[0]))
		}
	}

	// Get plugin version
	versionFunc := p.module.ExportedFunction("plugin_version")
	if versionFunc != nil {
		results, err := versionFunc.Call(ctx)
		if err != nil {
			return fmt.Errorf("call plugin_version: %w", err)
		}
		if len(results) > 0 {
			p.version = p.readString(uint32(results[0]))
		}
	}

	// Set defaults if not provided
	if p.name == "" {
		p.name = "unknown"
	}
	if p.version == "" {
		p.version = "0.0.0"
	}

	return nil
}

// Name returns the plugin's unique identifier.
func (p *WASMPlugin) Name() string {
	return p.name
}

// Version returns the plugin version string.
func (p *WASMPlugin) Version() string {
	return p.version
}

// Apply processes input text using the WASM plugin.
func (p *WASMPlugin) Apply(input string) (output string, tokensSaved int, err error) {
	ctx := context.Background()

	// Get the plugin_apply function
	applyFunc := p.module.ExportedFunction("plugin_apply")
	if applyFunc == nil {
		return "", 0, fmt.Errorf("plugin does not export 'plugin_apply' function")
	}

	// Allocate memory for input string
	inputPtr, err := p.allocateString(input)
	if err != nil {
		return "", 0, fmt.Errorf("allocate input: %w", err)
	}

	// Call plugin_apply(input_ptr, input_len)
	results, err := applyFunc.Call(ctx, uint64(inputPtr), uint64(len(input)))
	if err != nil {
		return "", 0, fmt.Errorf("call plugin_apply: %w", err)
	}

	if len(results) < 2 {
		return "", 0, fmt.Errorf("plugin_apply returned %d values, expected 2", len(results))
	}

	// Read output string from memory
	outputPtr := uint32(results[0])
	output = p.readString(outputPtr)

	// Get tokens saved
	tokensSaved = int(int32(results[1]))

	return output, tokensSaved, nil
}

// Close releases resources held by the plugin.
func (p *WASMPlugin) Close() error {
	if p.runtime != nil {
		return p.runtime.Close(context.Background())
	}
	return nil
}

// allocateString allocates memory in the WASM module for a string.
// Returns the pointer to the allocated memory.
func (p *WASMPlugin) allocateString(s string) (uint32, error) {
	ctx := context.Background()

	// Get the malloc function (if exported)
	mallocFunc := p.module.ExportedFunction("malloc")
	if mallocFunc == nil {
		return 0, fmt.Errorf("WASM module does not export 'malloc' function")
	}

	// Allocate memory
	results, err := mallocFunc.Call(ctx, uint64(len(s)))
	if err != nil {
		return 0, fmt.Errorf("malloc failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("malloc returned no results")
	}

	ptr := uint32(results[0])

	// Write string to memory
	if !p.module.Memory().Write(ptr, []byte(s)) {
		return 0, fmt.Errorf("failed to write string to WASM memory")
	}

	return ptr, nil
}

// readString reads a null-terminated string from WASM memory.
func (p *WASMPlugin) readString(ptr uint32) string {
	if ptr == 0 {
		return ""
	}

	mem := p.module.Memory()
	
	// Find string length (null-terminated)
	length := uint32(0)
	maxLen := uint32(10 * 1024 * 1024) // 10MB max
	
	for length < maxLen {
		b, ok := mem.ReadByte(ptr + length)
		if !ok || b == 0 {
			break
		}
		length++
	}

	if length == 0 {
		return ""
	}

	// Read the string
	bytes, ok := mem.Read(ptr, length)
	if !ok {
		return ""
	}

	return string(bytes)
}

// WASMPluginBuilder helps build WASM plugins with a fluent API.
type WASMPluginBuilder struct {
	path        string
	maxMemory   uint32
	config      wazero.ModuleConfig
	hostFuncs   map[string]interface{}
}

// NewWASMPluginBuilder creates a new WASM plugin builder.
func NewWASMPluginBuilder(path string) *WASMPluginBuilder {
	return &WASMPluginBuilder{
		path:      path,
		maxMemory: 64 * 1024 * 1024, // 64MB default
		hostFuncs: make(map[string]interface{}),
		config:    wazero.NewModuleConfig(),
	}
}

// WithMaxMemory sets the maximum memory for the WASM module.
func (b *WASMPluginBuilder) WithMaxMemory(bytes uint32) *WASMPluginBuilder {
	b.maxMemory = bytes
	return b
}

// WithStdout redirects stdout to the given writer.
func (b *WASMPluginBuilder) WithStdout(w *os.File) *WASMPluginBuilder {
	b.config = b.config.WithStdout(w)
	return b
}

// WithStderr redirects stderr to the given writer.
func (b *WASMPluginBuilder) WithStderr(w *os.File) *WASMPluginBuilder{
	b.config = b.config.WithStderr(w)
	return b
}

// Build loads and builds the WASM plugin.
func (b *WASMPluginBuilder) Build() (*WASMPlugin, error) {
	return LoadWASM(b.path)
}

// ValidateWASMPlugin checks if a WASM file is a valid TokMan plugin.
func ValidateWASMPlugin(path string) error {
	plugin, err := LoadWASM(path)
	if err != nil {
		return fmt.Errorf("load plugin: %w", err)
	}
	defer plugin.Close()

	// Check required exports
	requiredExports := []string{"plugin_apply"}
	for _, name := range requiredExports {
		if plugin.module.ExportedFunction(name) == nil {
			return fmt.Errorf("plugin missing required export: %s", name)
		}
	}

	return nil
}

// WASMPluginInfo returns information about a WASM plugin without loading it.
func WASMPluginInfo(path string) (*PluginInfo, error) {
	plugin, err := LoadWASM(path)
	if err != nil {
		return nil, err
	}
	defer plugin.Close()

	return &PluginInfo{
		Name:    plugin.Name(),
		Version: plugin.Version(),
		Type:    PluginTypeWASM,
		Path:    path,
		Enabled: false,
	}, nil
}
