//go:build !lua
// +build !lua

// Package luau provides a stub implementation when Lua is not available.
// Set build tag 'lua' to enable full Lua support with gopher-lua.
package luau

import (
	"fmt"
	"regexp"
	"strings"
)

// VM provides a Lua-like virtual machine (stub implementation).
type VM struct {
	filters map[string]string
}

// FilterAPI provides functions callable from filters.
type FilterAPI struct {
	vm *VM
}

// Sandbox provides security restrictions.
type Sandbox struct {
	MaxMemory       int
	MaxInstructions int
}

// NewVM creates a new VM (stub).
func NewVM() *VM {
	return &VM{
		filters: make(map[string]string),
	}
}

// Close closes the VM.
func (vm *VM) Close() {}

// ExecuteFilter runs a filter script on content (stub - uses Go regex).
func (vm *VM) ExecuteFilter(script string, content string) (string, error) {
	// Parse simple Lua-like patterns from script
	// Supports: INPUT:gsub(pattern, replacement)

	// Find gsub calls
	re := regexp.MustCompile(`INPUT:gsub\("([^"]+)",\s*"([^"]*)"\)`)
	matches := re.FindAllStringSubmatch(script, -1)

	result := content
	for _, match := range matches {
		if len(match) >= 3 {
			pattern := match[1]
			replacement := match[2]

			// Handle Lua patterns (simplified)
			goPattern := luaPatternToGo(pattern)
			re, err := regexp.Compile(goPattern)
			if err != nil {
				return content, fmt.Errorf("invalid pattern: %w", err)
			}
			result = re.ReplaceAllString(result, replacement)
		}
	}

	// Handle simple replacements
	re2 := regexp.MustCompile(`INPUT:gsub\('([^']+)',\s*'([^']*)'\)`)
	matches2 := re2.FindAllStringSubmatch(script, -1)
	for _, match := range matches2 {
		if len(match) >= 3 {
			pattern := match[1]
			replacement := match[2]
			goPattern := luaPatternToGo(pattern)
			re, _ := regexp.Compile(goPattern)
			result = re.ReplaceAllString(result, replacement)
		}
	}

	return result, nil
}

// LoadFilter loads a filter from a string.
func (vm *VM) LoadFilter(name, script string) error {
	vm.filters[name] = script
	return nil
}

// RunFilter runs a loaded filter.
func (vm *VM) RunFilter(name string, content string) (string, error) {
	script, ok := vm.filters[name]
	if !ok {
		return "", fmt.Errorf("filter not found: %s", name)
	}
	return vm.ExecuteFilter(script, content)
}

// luaPatternToGo converts Lua patterns to Go regex (simplified).
func luaPatternToGo(pattern string) string {
	// Handle basic Lua patterns
	result := pattern
	result = strings.ReplaceAll(result, "%%d", "[0-9]")
	result = strings.ReplaceAll(result, "%%a", "[a-zA-Z]")
	result = strings.ReplaceAll(result, "%%w", "[a-zA-Z0-9]")
	result = strings.ReplaceAll(result, "%%s", "[ \\t\\n\\r]")
	result = strings.ReplaceAll(result, "%%.", ".")
	result = strings.ReplaceAll(result, "%%-", "-")
	result = strings.ReplaceAll(result, "%%+", "\\+")
	result = strings.ReplaceAll(result, "%%%%", "%")
	return result
}

// Pool provides a pool of reusable VMs.
type Pool struct {
	vms  chan *VM
	size int
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

// stripANSICodes removes ANSI escape sequences.
func stripANSICodes(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] >= 0x40 && s[i] <= 0x7e {
				i++
			}
			i--
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
