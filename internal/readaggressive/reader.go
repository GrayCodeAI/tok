package readaggressive

import (
	"os"
	"strings"
)

type AggressiveReader struct {
	maxTokens int
}

func NewAggressiveReader(maxTokens int) *AggressiveReader {
	if maxTokens == 0 {
		maxTokens = 2000
	}
	return &AggressiveReader{maxTokens: maxTokens}
}

func (r *AggressiveReader) ReadAggressive(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	var signatures []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if r.isSignature(trimmed) {
			signatures = append(signatures, line)
		}
	}

	result := strings.Join(signatures, "\n")
	tokens := len(result) / 4
	if tokens > r.maxTokens {
		result = result[:r.maxTokens*4] + "..."
	}
	return result
}

func (r *AggressiveReader) isSignature(line string) bool {
	patterns := []string{
		"func ", "function ", "def ", "class ", "interface ", "type ",
		"export ", "import ", "from ", "require(", "pub fn ",
		"struct ", "enum ", "trait ", "impl ", "mod ",
		"public ", "private ", "protected ", "static ",
		"#include", "package ", "module ", "namespace ",
		"@", "describe(", "it(", "test(", "func Test",
		"const ", "var ", "let ",
	}
	for _, p := range patterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

func (r *AggressiveReader) TokenCount(content string) int {
	return len(content) / 4
}

func (r *AggressiveReader) WithinBudget(content string) bool {
	return r.TokenCount(content) <= r.maxTokens
}

type CompactTree struct {
	MaxDepth int
}

func NewCompactTree(maxDepth int) *CompactTree {
	if maxDepth == 0 {
		maxDepth = 3
	}
	return &CompactTree{MaxDepth: maxDepth}
}

func (t *CompactTree) Render(dir string, maxFiles int) string {
	var sb strings.Builder
	sb.WriteString(dir + "/\n")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return sb.String()
	}
	count := 0
	for _, entry := range entries {
		if count >= maxFiles {
			sb.WriteString("  ...\n")
			break
		}
		prefix := "  ├── "
		if entry.IsDir() {
			prefix = "  ├── [d] "
		}
		sb.WriteString(prefix + entry.Name() + "\n")
		count++
	}
	return sb.String()
}
