// Package contextread provides context reading capabilities (stub implementation).
// NOTE: This is a stub package. The full implementation was removed as dead code.
// These stub functions return empty/zero values to maintain API compatibility.
package contextread

import "strings"

// Options defines context reading options.
type Options struct {
	Mode              string
	Level             string
	MaxLines          int
	MaxTokens         int
	LineNumbers       bool
	StartLine         int
	EndLine           int
	SaveSnapshot      bool
	RelatedFilesCount int
}

// TrackedCommandPatternsForKind returns tracked command patterns for a kind (stub).
func TrackedCommandPatternsForKind(kind string) []string {
	return nil
}

// TrackedCommandPatterns returns all tracked command patterns (stub).
func TrackedCommandPatterns() []string {
	return nil
}

// Build builds context for a file (stub - returns input unchanged).
func Build(path, content, lang string, opts Options) (string, int, int, error) {
	return content, len(content) / 4, len(content) / 4, nil
}

// Analyze analyzes content for context (stub - returns input unchanged).
func Analyze(content string) string {
	return content
}

// Describe describes a file (stub - returns filename).
func Describe(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}
