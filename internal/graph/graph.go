// Package graph provides graph analysis functionality (stub implementation).
// NOTE: This is a stub package. The full implementation was removed as dead code.
// These stub functions maintain API compatibility.
package graph

import "fmt"

// ProjectGraph represents a project graph (stub).
type ProjectGraph struct {
	Path string
}

// NewProjectGraph creates a new project graph (stub).
func NewProjectGraph(path string) *ProjectGraph {
	return &ProjectGraph{Path: path}
}

// Analyze analyzes code structure (stub).
func (g *ProjectGraph) Analyze(depth string) error {
	return nil
}

// Stats returns graph statistics (stub).
func (g *ProjectGraph) Stats() map[string]interface{} {
	return map[string]interface{}{
		"by_language": map[string]int{},
	}
}

// FindRelatedFiles finds related files (stub).
func (g *ProjectGraph) FindRelatedFiles(file string, count int) []string {
	return []string{}
}

// ImpactAnalysis performs impact analysis (stub).
func (g *ProjectGraph) ImpactAnalysis(file string) string {
	return ""
}

// FormatGraphStats formats graph statistics (stub).
func FormatGraphStats(stats map[string]interface{}) string {
	return fmt.Sprintf("%v", stats)
}

// Visualize creates a visualization (stub).
func Visualize(path string) (string, error) {
	return "", nil
}
