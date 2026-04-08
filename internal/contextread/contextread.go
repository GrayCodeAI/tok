package contextread

import "strings"

type Reader struct{}

func New() *Reader {
	return &Reader{}
}

type TrackedPattern struct {
	Kind    string
	Pattern string
}

func TrackedCommandPatternsForKind(kind string) []TrackedPattern {
	return nil
}

func TrackedCommandPatterns() []TrackedPattern {
	var patterns []TrackedPattern
	for _, kind := range []string{"read", "write", "edit", "search"} {
		patterns = append(patterns, TrackedPattern{
			Kind:    kind,
			Pattern: ".*",
		})
	}
	return patterns
}

func (r *Reader) Analyze(content string) string {
	return content
}

func (r *Reader) ExtractContext(content string, maxTokens int) string {
	if len(content) > maxTokens*4 {
		return content[:maxTokens*4]
	}
	return content
}

func ContainsCodeBlocks(s string) bool {
	return strings.Contains(s, "```") || strings.Contains(s, "````")
}

type Options struct {
	MaxTokens         int
	Format            string
	Level             string
	Mode              string
	MaxLines          int
	LineNumbers       bool
	StartLine         int
	EndLine           int
	SaveSnapshot      bool
	RelatedFiles      []string
	RelatedFilesCount int
}

func Build(file string, content string, format string, opts Options) (string, int, int, error) {
	return content, len(content), len(content) / 4, nil
}

func Describe(content string, format string, mode string, opts Options) string {
	return content
}

type ReadMeta struct {
	Kind          string
	RequestedMode string
	ResolvedMode  string
	Target        string
	RelatedFiles  []string
	Bundle        string
}

func NewReadMeta() *ReadMeta {
	return &ReadMeta{}
}
