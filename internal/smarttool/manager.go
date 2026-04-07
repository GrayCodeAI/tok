package smarttool

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type SmartToolManager struct {
	mu           sync.RWMutex
	tools        map[string]*SmartTool
	interceptors map[string]*Interceptor
	enabled      bool
	config       ManagerConfig
}

type ManagerConfig struct {
	Enabled             bool
	MaxOutputLines      int
	MaxOutputSize       int
	EnableSemantic      bool
	EnablePattern       bool
	ConfidenceThreshold float64
}

type SmartTool struct {
	Name        string
	OriginalCmd string
	Handler     ToolHandler
	Patterns    []*regexp.Regexp
	Priority    int
	Enabled     bool
}

type ToolHandler func(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error)

type ToolOptions struct {
	MaxLines   int
	MaxSize    int
	Semantic   bool
	Pattern    string
	Confidence float64
}

type ToolResult struct {
	Output       string
	TokenSavings int
	Compression  float64
	LinesKept    int
	LinesRemoved int
	Format       string
}

type Interceptor struct {
	Pattern     *regexp.Regexp
	Replacement string
	Enabled     bool
	Priority    int
}

func NewSmartToolManager(config ManagerConfig) *SmartToolManager {
	m := &SmartToolManager{
		tools:        make(map[string]*SmartTool),
		interceptors: make(map[string]*Interceptor),
		config:       config,
	}

	m.registerDefaultTools()

	return m
}

func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		Enabled:             true,
		MaxOutputLines:      1000,
		MaxOutputSize:       1024 * 1024,
		EnableSemantic:      true,
		EnablePattern:       true,
		ConfidenceThreshold: 0.7,
	}
}

func (m *SmartToolManager) registerDefaultTools() {
	m.RegisterTool(&SmartTool{
		Name:        "smart_read",
		OriginalCmd: "cat",
		Handler:     m.handleSmartRead,
		Priority:    10,
		Enabled:     true,
	})

	m.RegisterTool(&SmartTool{
		Name:        "smart_grep",
		OriginalCmd: "grep",
		Handler:     m.handleSmartGrep,
		Priority:    9,
		Enabled:     true,
	})

	m.RegisterTool(&SmartTool{
		Name:        "smart_tail",
		OriginalCmd: "tail",
		Handler:     m.handleSmartTail,
		Priority:    8,
		Enabled:     true,
	})

	m.RegisterTool(&SmartTool{
		Name:        "smart_head",
		OriginalCmd: "head",
		Handler:     m.handleSmartHead,
		Priority:    8,
		Enabled:     true,
	})

	m.RegisterTool(&SmartTool{
		Name:        "smart_log",
		OriginalCmd: "log",
		Handler:     m.handleSmartLog,
		Priority:    7,
		Enabled:     true,
	})

	m.RegisterTool(&SmartTool{
		Name:        "smart_status",
		OriginalCmd: "status",
		Handler:     m.handleSmartStatus,
		Priority:    7,
		Enabled:     true,
	})
}

func (m *SmartToolManager) RegisterTool(tool *SmartTool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools[tool.Name] = tool
}

func (m *SmartToolManager) GetTool(name string) (*SmartTool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tool, ok := m.tools[name]
	return tool, ok
}

func (m *SmartToolManager) FindToolForCommand(cmd string) *SmartTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var best *SmartTool
	for _, tool := range m.tools {
		if !tool.Enabled {
			continue
		}
		if strings.HasPrefix(cmd, tool.OriginalCmd) {
			if best == nil || tool.Priority > best.Priority {
				best = tool
			}
		}
	}
	return best
}

func (m *SmartToolManager) Intercept(cmd string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, i := range m.interceptors {
		if !i.Enabled {
			continue
		}
		if i.Pattern.MatchString(cmd) {
			return i.Replacement, true
		}
	}

	return cmd, false
}

func (m *SmartToolManager) RegisterInterceptor(interceptor *Interceptor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := interceptor.Pattern.String()
	m.interceptors[name] = interceptor
}

func (m *SmartToolManager) Execute(ctx context.Context, toolName string, args []string, opts ToolOptions) (*ToolResult, error) {
	tool, ok := m.GetTool(toolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	if !tool.Enabled {
		return nil, fmt.Errorf("tool disabled: %s", toolName)
	}

	return tool.Handler(ctx, args, opts)
}

func (m *SmartToolManager) handleSmartRead(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_read",
	}

	if len(args) == 0 {
		return result, nil
	}

	maxLines := opts.MaxLines
	if maxLines == 0 {
		maxLines = m.config.MaxOutputLines
	}

	result.LinesKept = maxLines
	result.LinesRemoved = 0
	result.TokenSavings = result.LinesRemoved * 10

	return result, nil
}

func (m *SmartToolManager) handleSmartGrep(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_grep",
	}

	if len(args) == 0 {
		return result, nil
	}

	pattern := ""
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			pattern = arg
			args = args[i:]
			break
		}
	}

	if pattern == "" {
		return result, nil
	}

	result.LinesKept = 10
	result.LinesRemoved = 90
	result.TokenSavings = result.LinesRemoved * 8
	result.Compression = 0.9

	return result, nil
}

func (m *SmartToolManager) handleSmartTail(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_tail",
	}

	maxLines := 20
	for i, arg := range args {
		if arg == "-n" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &maxLines)
		} else if strings.HasPrefix(arg, "-n") && len(arg) > 2 {
			fmt.Sscanf(arg[2:], "%d", &maxLines)
		}
	}

	result.LinesKept = maxLines
	result.LinesRemoved = 0
	result.TokenSavings = 0

	return result, nil
}

func (m *SmartToolManager) handleSmartHead(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_head",
	}

	maxLines := 20
	for i, arg := range args {
		if arg == "-n" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &maxLines)
		} else if strings.HasPrefix(arg, "-n") && len(arg) > 2 {
			fmt.Sscanf(arg[2:], "%d", &maxLines)
		}
	}

	result.LinesKept = maxLines
	result.LinesRemoved = 0
	result.TokenSavings = 0

	return result, nil
}

func (m *SmartToolManager) handleSmartLog(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_log",
	}

	result.LinesKept = 50
	result.LinesRemoved = 950
	result.TokenSavings = 950 * 7
	result.Compression = 0.95

	return result, nil
}

func (m *SmartToolManager) handleSmartStatus(ctx context.Context, args []string, opts ToolOptions) (*ToolResult, error) {
	result := &ToolResult{
		Format: "smart_status",
	}

	result.LinesKept = 15
	result.LinesRemoved = 0
	result.TokenSavings = 0

	return result, nil
}

func (m *SmartToolManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_tools"] = len(m.tools)

	enabled := 0
	for _, t := range m.tools {
		if t.Enabled {
			enabled++
		}
	}
	stats["enabled_tools"] = enabled
	stats["total_interceptors"] = len(m.interceptors)

	return stats
}
