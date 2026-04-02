package mcptools

import "strings"

type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"`
}

type MCPToolRegistry struct {
	tools map[string]*MCPTool
}

func NewMCPToolRegistry() *MCPToolRegistry {
	reg := &MCPToolRegistry{
		tools: make(map[string]*MCPTool),
	}
	reg.registerBuiltInTools()
	return reg
}

func (r *MCPToolRegistry) registerBuiltInTools() {
	tools := []MCPTool{
		{Name: "ctx_read", Description: "Read file with 7 modes (full, map, signatures, diff, aggressive, entropy, graph)"},
		{Name: "ctx_multi_read", Description: "Read multiple files at once"},
		{Name: "ctx_tree", Description: "Show directory tree structure"},
		{Name: "ctx_shell", Description: "Execute shell command with compression"},
		{Name: "ctx_search", Description: "Search file contents with compression"},
		{Name: "ctx_compress", Description: "Compress text input"},
		{Name: "ctx_smart_read", Description: "Smart file reading with auto-mode"},
		{Name: "ctx_delta", Description: "Show diff between file versions"},
		{Name: "ctx_dedup", Description: "Remove duplicate content"},
		{Name: "ctx_fill", Description: "Fill in missing context"},
		{Name: "ctx_intent", Description: "Detect user intent from context"},
		{Name: "ctx_response", Description: "Generate optimized response"},
		{Name: "ctx_context", Description: "Manage session context"},
		{Name: "ctx_graph", Description: "Build dependency graph"},
		{Name: "ctx_discover", Description: "Discover relevant files"},
		{Name: "ctx_session", Description: "Session management"},
		{Name: "ctx_knowledge", Description: "Persistent knowledge store"},
		{Name: "ctx_agent", Description: "Multi-agent coordination"},
		{Name: "ctx_wrapped", Description: "Generate savings report"},
		{Name: "ctx_benchmark", Description: "Benchmark compression"},
		{Name: "ctx_metrics", Description: "Display usage metrics"},
		{Name: "ctx_analyze", Description: "Analyze codebase"},
		{Name: "ctx_cache", Description: "Manage cache"},
		{Name: "fetch_clean", Description: "Fetch URL and clean HTML to text"},
		{Name: "fetch_clean_batch", Description: "Fetch multiple URLs and clean"},
		{Name: "refine_prompt", Description: "Strip filler words from prompt"},
	}
	for i := range tools {
		r.tools[tools[i].Name] = &tools[i]
	}
}

func (r *MCPToolRegistry) Register(tool *MCPTool) {
	r.tools[tool.Name] = tool
}

func (r *MCPToolRegistry) Get(name string) *MCPTool {
	return r.tools[name]
}

func (r *MCPToolRegistry) List() []*MCPTool {
	var result []*MCPTool
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

func (r *MCPToolRegistry) Search(query string) []*MCPTool {
	var result []*MCPTool
	query = strings.ToLower(query)
	for _, t := range r.tools {
		if strings.Contains(strings.ToLower(t.Name), query) || strings.Contains(strings.ToLower(t.Description), query) {
			result = append(result, t)
		}
	}
	return result
}

func (r *MCPToolRegistry) Count() int {
	return len(r.tools)
}

type EntityProtector struct {
	patterns map[string]string
}

func NewEntityProtector() *EntityProtector {
	return &EntityProtector{
		patterns: map[string]string{
			"ticker":  `\b[A-Z]{2,5}\b`,
			"date":    `\d{4}-\d{2}-\d{2}`,
			"money":   `\$\d+\.?\d*[KMB]?`,
			"percent": `\d+\.?\d*%`,
		},
	}
}

func (p *EntityProtector) Protect(input string) (string, map[string][]string) {
	protected := make(map[string][]string)
	_ = p.patterns
	_ = protected
	return input, protected
}

func (p *EntityProtector) Restore(input string, protected map[string][]string) string {
	return input
}
