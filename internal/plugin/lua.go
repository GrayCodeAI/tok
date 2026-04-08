package plugin

import (
	"fmt"
	"os"

	"github.com/yuin/gopher-lua"
)

type LuaPlugin struct {
	L       *lua.LState
	name    string
	version string
	path    string
}

func LoadLua(path string) (*LuaPlugin, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lua file: %w", err)
	}

	L := lua.NewState(lua.Options{
		CallStackSize: 1024,
		RegistrySize:  1024,
	})

	defer L.Close()

	if err := L.DoString(string(content)); err != nil {
		return nil, fmt.Errorf("load script: %w", err)
	}

	plugin := &LuaPlugin{
		L:    L,
		path: path,
	}

	if err := plugin.loadMetadata(); err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	return plugin, nil
}

func (p *LuaPlugin) loadMetadata() error {
	fn := p.L.GetGlobal("plugin_name")
	if fn, ok := fn.(lua.LValue); ok {
		p.L.Push(fn)
		if err := p.L.PCall(0, 1, nil); err == nil {
			if ret := p.L.Get(-1); ret != nil && ret.Type() == lua.LTString {
				p.name = ret.String()
			}
		}
	}

	fn = p.L.GetGlobal("plugin_version")
	if fn, ok := fn.(lua.LValue); ok {
		p.L.Push(fn)
		if err := p.L.PCall(0, 1, nil); err == nil {
			if ret := p.L.Get(-1); ret != nil && ret.Type() == lua.LTString {
				p.version = ret.String()
			}
		}
	}

	if p.name == "" {
		p.name = "unknown"
	}
	if p.version == "" {
		p.version = "0.0.0"
	}

	return nil
}

func (p *LuaPlugin) Name() string {
	return p.name
}

func (p *LuaPlugin) Version() string {
	return p.version
}

func (p *LuaPlugin) Apply(input string) (string, int, error) {
	fn := p.L.GetGlobal("plugin_apply")
	if fn == nil || fn.Type() != lua.LTFunction {
		return "", 0, fmt.Errorf("plugin does not define 'plugin_apply' function")
	}

	p.L.Push(fn)
	p.L.Push(lua.LString(input))

	if err := p.L.PCall(1, 2, nil); err != nil {
		return "", 0, fmt.Errorf("apply failed: %w", err)
	}

	output := p.L.Get(-2).String()
	tokensSaved := 0
	if ln, ok := p.L.Get(-1).(lua.LNumber); ok {
		tokensSaved = int(ln)
	}

	return output, tokensSaved, nil
}

func (p *LuaPlugin) Close() error {
	if p.L != nil {
		p.L.Close()
	}
	return nil
}

func ValidateLuaPlugin(path string) error {
	_, err := LoadLua(path)
	return err
}

func LuaPluginInfo(path string) (*PluginInfo, error) {
	plugin, err := LoadLua(path)
	if err != nil {
		return nil, err
	}
	defer plugin.Close()

	return &PluginInfo{
		Name:    plugin.Name(),
		Version: plugin.Version(),
		Type:    PluginTypeLua,
		Path:    path,
	}, nil
}
