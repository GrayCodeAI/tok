package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ProjectGraph represents the dependency graph of a project.
// Inspired by lean-ctx's ctx_graph.
type ProjectGraph struct {
	mu      sync.RWMutex
	nodes   map[string]*Node
	edges   map[string][]string
	rootDir string
}

// Node represents a file in the project graph.
type Node struct {
	Path       string
	Language   string
	Size       int64
	Imports    []string
	ImportedBy []string
	Tags       []string
}

// NewProjectGraph creates a new project graph.
func NewProjectGraph(rootDir string) *ProjectGraph {
	return &ProjectGraph{
		nodes:   make(map[string]*Node),
		edges:   make(map[string][]string),
		rootDir: rootDir,
	}
}

// Analyze scans the project and builds the dependency graph.
func (g *ProjectGraph) Analyze() error {
	return filepath.Walk(g.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		lang := detectLanguage(path)
		if lang == "" {
			return nil
		}

		relPath, _ := filepath.Rel(g.rootDir, path)
		node := &Node{
			Path:     relPath,
			Language: lang,
			Size:     info.Size(),
		}

		imports := extractImports(path, lang)
		node.Imports = imports

		g.mu.Lock()
		g.nodes[relPath] = node
		for _, imp := range imports {
			g.edges[relPath] = append(g.edges[relPath], imp)
		}
		g.mu.Unlock()

		return nil
	})
}

// FindRelatedFiles finds files related to the given file through dependencies.
func (g *ProjectGraph) FindRelatedFiles(path string, maxResults int) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	related := make(map[string]int)

	// Direct imports
	for _, imp := range g.edges[path] {
		related[imp] = 10
	}

	// Files that import this file
	for file, imports := range g.edges {
		for _, imp := range imports {
			if imp == path {
				related[file] = 8
			}
		}
	}

	// Same directory files
	dir := filepath.Dir(path)
	for file := range g.nodes {
		if filepath.Dir(file) == dir && file != path {
			related[file] = max(related[file], 3)
		}
	}

	// Sort by score
	type kv struct {
		path  string
		score int
	}
	var pairs []kv
	for k, v := range related {
		pairs = append(pairs, kv{k, v})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].score < pairs[j].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	var results []string
	for i := 0; i < len(pairs) && i < maxResults; i++ {
		results = append(results, pairs[i].path)
	}
	return results
}

// ImpactAnalysis finds all files affected by a change to the given file.
func (g *ProjectGraph) ImpactAnalysis(path string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	affected := make(map[string]bool)
	queue := []string{path}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for file, imports := range g.edges {
			for _, imp := range imports {
				if imp == current && !affected[file] {
					affected[file] = true
					queue = append(queue, file)
				}
			}
		}
	}

	var results []string
	for f := range affected {
		results = append(results, f)
	}
	return results
}

// Stats returns project statistics.
func (g *ProjectGraph) Stats() map[string]any {
	g.mu.RLock()
	defer g.mu.RUnlock()

	byLang := make(map[string]int)
	totalSize := int64(0)
	totalFiles := len(g.nodes)

	for _, node := range g.nodes {
		byLang[node.Language]++
		totalSize += node.Size
	}

	return map[string]any{
		"total_files": totalFiles,
		"total_size":  totalSize,
		"by_language": byLang,
		"total_edges": len(g.edges),
	}
}

func shouldSkipDir(path string) bool {
	skipDirs := []string{
		"node_modules", ".git", "vendor", "__pycache__",
		".tox", ".venv", "dist", "build", "target",
		".next", ".nuxt", ".svelte-kit",
	}
	base := filepath.Base(path)
	for _, d := range skipDirs {
		if base == d {
			return true
		}
	}
	return false
}

func detectLanguage(path string) string {
	ext := filepath.Ext(path)
	langMap := map[string]string{
		".go":    "go",
		".rs":    "rust",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".jsx":   "javascript",
		".rb":    "ruby",
		".java":  "java",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".cs":    "csharp",
		".php":   "php",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
		".toml":  "toml",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".md":    "markdown",
	}
	return langMap[ext]
}

func extractImports(path string, lang string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	var imports []string

	switch lang {
	case "go":
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") {
				if strings.Contains(line, "\"") {
					start := strings.Index(line, "\"")
					end := strings.LastIndex(line, "\"")
					if start < end {
						imports = append(imports, line[start+1:end])
					}
				}
			}
		}
	case "python":
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					imports = append(imports, parts[1])
				}
			}
		}
	case "javascript", "typescript":
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "require(") {
				if strings.Contains(line, "from '") {
					start := strings.Index(line, "from '") + 6
					end := strings.Index(line[start:], "'")
					if end > 0 {
						imports = append(imports, line[start:start+end])
					}
				} else if strings.Contains(line, "from \"") {
					start := strings.Index(line, "from \"") + 6
					end := strings.Index(line[start:], "\"")
					if end > 0 {
						imports = append(imports, line[start:start+end])
					}
				}
			}
		}
	}

	return imports
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FormatGraphStats returns a human-readable stats string.
func FormatGraphStats(stats map[string]any) string {
	return fmt.Sprintf("Graph: %d files, %d edges, %d languages",
		stats["total_files"], stats["total_edges"], len(stats["by_language"].(map[string]int)))
}
