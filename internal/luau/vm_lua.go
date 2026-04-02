//go:build lua
// +build lua

// Package luau provides full Lua scripting support with gopher-lua.
// Build with -tags=lua to enable.
package luau

import (
	"fmt"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// VM provides a Lua virtual machine using gopher-lua.
type VM struct {
	L       *lua.LState
	filters map[string]*lua.LFunction
	mu      sync.RWMutex
	sandbox *Sandbox
}

// FilterAPI provides functions callable from Lua filters.
type FilterAPI struct {
	vm *VM
}

// Sandbox provides security restrictions.
type Sandbox struct {
	MaxMemory       int
	MaxInstructions int
	AllowedFuncs    []string
	BlockedFuncs    []string
}

// DefaultSandbox returns a secure default sandbox configuration.
func DefaultSandbox() *Sandbox {
	return &Sandbox{
		MaxMemory:       64 * 1024 * 1024, // 64MB
		MaxInstructions: 1000000,
		AllowedFuncs: []string{
			"string", "table", "math", "pairs", "ipairs",
			"tonumber", "tostring", "type", "next",
		},
		BlockedFuncs: []string{
			"os.execute", "os.exit", "io.open", "io.popen",
			"load", "loadfile", "dofile", "require",
		},
	}
}

// NewVM creates a new Lua VM with sandboxing.
func NewVM() *VM {
	vm := &VM{
		filters: make(map[string]*lua.LFunction),
		sandbox: DefaultSandbox(),
	}

	vm.L = lua.NewState(lua.Options{
		SkipOpenLibs: true, // We'll open only safe libs
	})

	vm.setupSandbox()
	return vm
}

// setupSandbox configures the Lua sandbox environment.
func (vm *VM) setupSandbox() {
	// Open safe standard libraries
	for _, lib := range vm.sandbox.AllowedFuncs {
		switch lib {
		case "string":
			vm.L.OpenLibs(lua.StringLib)
		case "table":
			vm.L.OpenLibs(lua.TableLib)
		case "math":
			vm.L.OpenLibs(lua.MathLib)
		}
	}

	// Block dangerous functions
	for _, blocked := range vm.sandbox.BlockedFuncs {
		parts := strings.Split(blocked, ".")
		if len(parts) == 2 {
			mod := vm.L.GetGlobal(parts[0])
			if mod.Type() == lua.LTTable {
				vm.L.SetField(mod, parts[1], lua.LNil)
			}
		} else {
			vm.L.SetGlobal(blocked, lua.LNil)
		}
	}

	// Set up INPUT global and helper functions
	vm.L.SetGlobal("INPUT", lua.LString(""))

	// Add custom filter API
	api := &FilterAPI{vm: vm}
	vm.L.SetGlobal("FilterAPI", luar.New(vm.L, api))

	// Add gsub helper that works on global INPUT
	vm.L.DoString(`
		function gsub(pattern, replacement)
			local result, count = string.gsub(INPUT, pattern, replacement)
			INPUT = result
			return count
		end

		function match(pattern)
			return string.match(INPUT, pattern)
		end

		function find(pattern)
			return string.find(INPUT, pattern)
		end

		function lines()
			local result = {}
			for line in string.gmatch(INPUT, "[^\r\n]+") do
				table.insert(result, line)
			end
			return result
		end

		function join(sep, tbl)
			return table.concat(tbl, sep or "\n")
		end
	`)
}

// Close closes the Lua VM.
func (vm *VM) Close() {
	if vm.L != nil {
		vm.L.Close()
	}
}

// ExecuteFilter executes a Lua filter script on content.
func (vm *VM) ExecuteFilter(script string, content string) (string, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Set INPUT
	vm.L.SetGlobal("INPUT", lua.LString(content))

	// Execute the filter script
	if err := vm.L.DoString(script); err != nil {
		return content, fmt.Errorf("filter execution error: %w", err)
	}

	// Get result from OUTPUT or INPUT
	output := vm.L.GetGlobal("OUTPUT")
	if output.Type() == lua.LTString {
		return output.String(), nil
	}

	// Fall back to INPUT if OUTPUT not set
	input := vm.L.GetGlobal("INPUT")
	return input.String(), nil
}

// LoadFilter loads a named filter from a script.
func (vm *VM) LoadFilter(name, script string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Compile the script to check for errors
	fn, err := vm.L.LoadString(script)
	if err != nil {
		return fmt.Errorf("failed to load filter %s: %w", name, err)
	}

	vm.filters[name] = fn
	return nil
}

// RunFilter runs a previously loaded filter.
func (vm *VM) RunFilter(name string, content string) (string, error) {
	vm.mu.RLock()
	fn, ok := vm.filters[name]
	vm.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("filter not found: %s", name)
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Set up environment
	vm.L.SetGlobal("INPUT", lua.LString(content))

	// Run the compiled function
	vm.L.Push(fn)
	if err := vm.L.PCall(0, 0, nil); err != nil {
		return content, fmt.Errorf("filter execution error: %w", err)
	}

	// Get result
	output := vm.L.GetGlobal("OUTPUT")
	if output.Type() == lua.LTString {
		return output.String(), nil
	}

	input := vm.L.GetGlobal("INPUT")
	return input.String(), nil
}

// Pool provides a pool of reusable VMs.
type Pool struct {
	vms  chan *VM
	size int
	mu   sync.Mutex
}

// NewPool creates a new VM pool.
func NewPool(size int) *Pool {
	pool := &Pool{
		vms:  make(chan *VM, size),
		size: size,
	}

	for i := 0; i < size; i++ {
		pool.vms <- NewVM()
	}

	return pool
}

// Get gets a VM from the pool.
func (p *Pool) Get() *VM {
	select {
	case vm := <-p.vms:
		return vm
	default:
		return NewVM()
	}
}

// Put returns a VM to the pool.
func (p *Pool) Put(vm *VM) {
	// Reset the VM state
	if vm.L != nil {
		vm.L.SetGlobal("INPUT", lua.LString(""))
		vm.L.SetGlobal("OUTPUT", lua.LNil)
	}

	select {
	case p.vms <- vm:
	default:
		vm.Close()
	}
}

// Close closes all VMs in the pool.
func (p *Pool) Close() {
	close(p.vms)
	for vm := range p.vms {
		vm.Close()
	}
}

// FilterAPI methods exposed to Lua

// Log logs a message from Lua filter.
func (api *FilterAPI) Log(msg string) {
	// In production, this would use proper logging
	fmt.Printf("[Lua Filter] %s\n", msg)
}

// MatchGroups returns all capture groups from a pattern match.
func (api *FilterAPI) MatchGroups(pattern, text string) []string {
	re := api.vm.L.GetGlobal("string") // Access string library
	if re.Type() != lua.LTTable {
		return nil
	}

	matchFn := api.vm.L.GetField(re, "match")
	api.vm.L.Push(matchFn)
	api.vm.L.Push(lua.LString(text))
	api.vm.L.Push(lua.LString(pattern))
	api.vm.L.Call(2, lua.MultRet)

	var results []string
	n := api.vm.L.GetTop()
	for i := 1; i <= n; i++ {
		v := api.vm.L.Get(i)
		if v.Type() == lua.LTString {
			results = append(results, v.String())
		}
	}
	api.vm.L.SetTop(0) // Clear stack

	return results
}
