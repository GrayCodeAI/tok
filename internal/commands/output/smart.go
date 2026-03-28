package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

func init() {
	registry.Add(func() { registry.Register(smartCmd) })
}

var smartCmd = &cobra.Command{
	Use:   "smart <file>",
	Short: "Generate 2-line technical summary of a file",
	Long: `Generate a concise 2-line technical summary of a source file.

Uses heuristic analysis to extract:
- File purpose/main functionality
- Key exports, functions, or classes

Example:
  tokman smart main.go
  tokman smart src/index.ts`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		summary := generateSmartSummary(filePath)
		fmt.Println(summary)
	},
}

// generateSmartSummary creates a 2-line technical summary
func generateSmartSummary(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	text := string(content)

	// Get file stats
	lines := strings.Count(text, "\n") + 1
	tokens := filter.EstimateTokens(text)

	// Detect language
	lang := detectLanguage(ext)

	// Extract key information
	exports := extractExports(text, ext)
	imports := countImports(text, ext)
	functions := countFunctions(text, ext)

	// Build summary
	line1 := fmt.Sprintf("%s (%d lines, ~%d tokens) - %s",
		filepath.Base(filePath), lines, tokens, lang)

	line2 := fmt.Sprintf("Imports: %d | Exports: %s | Functions: %d",
		imports, exports, functions)

	return line1 + "\n" + line2
}

func detectLanguage(ext string) string {
	langMap := map[string]string{
		".go":    "Go",
		".rs":    "Rust",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".jsx":   "React JSX",
		".tsx":   "React TSX",
		".java":  "Java",
		".kt":    "Kotlin",
		".rb":    "Ruby",
		".php":   "PHP",
		".c":     "C",
		".cpp":   "C++",
		".cs":    "C#",
		".swift": "Swift",
		".sh":    "Shell",
		".md":    "Markdown",
		".json":  "JSON",
		".yaml":  "YAML",
		".toml":  "TOML",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "Unknown"
}

func extractExports(text, ext string) string {
	exports := []string{}

	switch ext {
	case ".go":
		// Find exported functions/types
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "func ") && len(line) > 5 && line[5] >= 'A' && line[5] <= 'Z' {
				// Extract function name
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					name := strings.Split(parts[1], "(")[0]
					exports = append(exports, name)
				}
			}
			if strings.HasPrefix(line, "type ") && len(line) > 5 && line[5] >= 'A' && line[5] <= 'Z' {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					exports = append(exports, parts[1])
				}
			}
		}
	case ".ts", ".tsx":
		// Find exports
		if strings.Contains(text, "export default") {
			exports = append(exports, "default")
		}
		exportConst := strings.Count(text, "export const")
		exportFunc := strings.Count(text, "export function")
		exportClass := strings.Count(text, "export class")
		if exportConst > 0 {
			exports = append(exports, fmt.Sprintf("%d const", exportConst))
		}
		if exportFunc > 0 {
			exports = append(exports, fmt.Sprintf("%d func", exportFunc))
		}
		if exportClass > 0 {
			exports = append(exports, fmt.Sprintf("%d class", exportClass))
		}
	case ".js", ".jsx":
		if strings.Contains(text, "module.exports") || strings.Contains(text, "export ") {
			exports = append(exports, "module")
		}
	case ".py":
		// Find classes and functions at module level
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "class ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					name := strings.Split(parts[1], "(")[0]
					exports = append(exports, name)
				}
			}
		}
	}

	if len(exports) == 0 {
		return "none"
	}
	if len(exports) > 5 {
		return fmt.Sprintf("%s + %d more", strings.Join(exports[:5], ", "), len(exports)-5)
	}
	return strings.Join(exports, ", ")
}

func countImports(text, ext string) int {
	switch ext {
	case ".go":
		return strings.Count(text, "\nimport ") + strings.Count(text, "import (")
	case ".ts", ".tsx", ".js", ".jsx":
		return strings.Count(text, "import ") + strings.Count(text, "require(")
	case ".py":
		return strings.Count(text, "\nimport ") + strings.Count(text, "\nfrom ")
	case ".rs":
		return strings.Count(text, "use ")
	case ".java":
		return strings.Count(text, "import ")
	default:
		return 0
	}
}

func countFunctions(text, ext string) int {
	switch ext {
	case ".go":
		return strings.Count(text, "\nfunc ")
	case ".ts", ".tsx", ".js", ".jsx":
		return strings.Count(text, "function ") + strings.Count(text, "=>")
	case ".py":
		return strings.Count(text, "\ndef ")
	case ".rs":
		return strings.Count(text, "\nfn ")
	case ".java":
		return strings.Count(text, " void ") + strings.Count(text, " static ")
	default:
		return 0
	}
}

func init() {
	if shared.IsVerbose() {
		// Verbose mode shows more detail
	}
}
