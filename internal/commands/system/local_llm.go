package system

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var localLlmCmd = &cobra.Command{
	Use:   "local-llm <file>",
	Short: "Heuristic code analysis without external model",
	Long: `Summarizes source files using heuristic analysis — no external model needed.
Extracts imports, functions, structs, and detects patterns to produce
a two-line summary of what the code does.

Example:
  tokman local-llm main.go
  tokman local-llm src/app.py`,
	Args: cobra.ExactArgs(1),
	RunE: runLocalLlm,
}

func init() {
	registry.Add(func() { registry.Register(localLlmCmd) })
}

func runLocalLlm(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ext := filepath.Ext(filePath)
	lang := detectLang(ext)
	summary := analyzeCode(string(content), lang)

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println(cyan(summary.line1))
	fmt.Println(green(summary.line2))

	return nil
}

type codeSummary struct {
	line1 string
	line2 string
}

func analyzeCode(content string, lang string) codeSummary {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	imports := extractImports(content, lang)
	functions := extractFunctions(content, lang)
	structs := extractStructs(content, lang)
	traits := extractTraits(content, lang)
	patterns := detectPatterns(content, lang)

	langName := langDisplayName(lang)

	var mainType string
	if len(structs) > 0 && len(functions) > 0 {
		mainType = langName + " module"
	} else if len(structs) > 0 {
		mainType = langName + " data structures"
	} else if len(functions) > 0 {
		mainType = langName + " functions"
	} else {
		mainType = langName + " code"
	}

	var components []string
	if len(functions) > 0 {
		components = append(components, fmt.Sprintf("%d fn", len(functions)))
	}
	if len(structs) > 0 {
		components = append(components, fmt.Sprintf("%d struct", len(structs)))
	}
	if len(traits) > 0 {
		components = append(components, fmt.Sprintf("%d trait", len(traits)))
	}

	var line1 string
	if len(components) == 0 {
		line1 = fmt.Sprintf("%s (%d lines)", mainType, totalLines)
	} else {
		line1 = fmt.Sprintf("%s (%s) - %d lines", mainType, strings.Join(components, ", "), totalLines)
	}

	var details []string
	if len(imports) > 0 {
		limit := 3
		if len(imports) < limit {
			limit = len(imports)
		}
		details = append(details, fmt.Sprintf("uses: %s", strings.Join(imports[:limit], ", ")))
	}
	if len(patterns) > 0 {
		limit := 3
		if len(patterns) < limit {
			limit = len(patterns)
		}
		details = append(details, fmt.Sprintf("patterns: %s", strings.Join(patterns[:limit], ", ")))
	}
	if len(functions) > 0 && len(details) == 0 {
		limit := 3
		if len(functions) < limit {
			limit = len(functions)
		}
		details = append(details, fmt.Sprintf("defines: %s", strings.Join(functions[:limit], ", ")))
	}

	var line2 string
	if len(details) == 0 {
		line2 = "General purpose code file"
	} else {
		line2 = strings.Join(details, " | ")
	}

	return codeSummary{line1: line1, line2: line2}
}

func langDisplayName(lang string) string {
	names := map[string]string{
		"rust":       "Rust",
		"python":     "Python",
		"javascript": "JavaScript",
		"typescript": "TypeScript",
		"go":         "Go",
		"c":          "C",
		"cpp":        "C++",
		"java":       "Java",
		"ruby":       "Ruby",
		"shell":      "Shell",
		"data":       "Data",
	}
	if name, ok := names[lang]; ok {
		return name
	}
	return "Code"
}

func detectLang(ext string) string {
	switch ext {
	case ".rs":
		return "rust"
	case ".py":
		return "python"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".go":
		return "go"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".sh", ".bash", ".zsh":
		return "shell"
	case ".json", ".yaml", ".yml", ".toml", ".xml", ".csv":
		return "data"
	default:
		return "unknown"
	}
}

func extractImports(content string, lang string) []string {
	var pattern string
	switch lang {
	case "rust":
		pattern = `^use\s+([a-zA-Z_][a-zA-Z0-9_]*(?:::[a-zA-Z_][a-zA-Z0-9_]*)?)`
	case "python":
		pattern = `^(?:from\s+(\S+)|import\s+(\S+))`
	case "javascript", "typescript":
		pattern = `(?:import.*from\s+['"]([^'"]+)['"]|require\(['"]([^'"]+)['"]\))`
	case "go":
		pattern = `^\s*"([^"]+)"$`
	default:
		return nil
	}

	re := regexp.MustCompile(pattern)
	var imports []string
	seen := make(map[string]bool)

	for _, line := range strings.Split(content, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		imp := matches[1]
		if len(matches) > 2 && imp == "" {
			imp = matches[2]
		}
		base := strings.Split(imp, "::")[0]
		base = strings.Trim(base, "\"")
		if !seen[base] && !isStdImport(base, lang) {
			seen[base] = true
			imports = append(imports, base)
		}
		if len(imports) >= 5 {
			break
		}
	}
	return imports
}

func isStdImport(name string, lang string) bool {
	switch lang {
	case "rust":
		return name == "std" || name == "core" || name == "alloc"
	case "python":
		return name == "os" || name == "sys" || name == "re" || name == "json" || name == "typing"
	default:
		return false
	}
}

func extractFunctions(content string, lang string) []string {
	var pattern string
	switch lang {
	case "rust":
		pattern = `(?:pub\s+)?(?:async\s+)?fn\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "python":
		pattern = `def\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "javascript", "typescript":
		pattern = `(?:async\s+)?function\s+([a-zA-Z_][a-zA-Z0-9_]*)|(?:const|let|var)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(?:async\s+)?\(`
	case "go":
		pattern = `func\s+(?:\([^)]+\)\s+)?([a-zA-Z_][a-zA-Z0-9_]*)`
	default:
		return nil
	}

	re := regexp.MustCompile(pattern)
	var functions []string

	for _, line := range strings.Split(content, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		name := matches[1]
		if len(matches) > 2 && name == "" {
			name = matches[2]
		}
		if !strings.HasPrefix(name, "test_") && name != "main" && name != "new" {
			functions = append(functions, name)
		}
		if len(functions) >= 10 {
			break
		}
	}
	return functions
}

func extractStructs(content string, lang string) []string {
	var pattern string
	switch lang {
	case "rust":
		pattern = `(?:pub\s+)?(?:struct|enum)\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "python":
		pattern = `class\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "typescript":
		pattern = `(?:interface|class|type)\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "go":
		pattern = `type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+struct`
	case "java":
		pattern = `(?:public\s+)?class\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	default:
		return nil
	}

	re := regexp.MustCompile(pattern)
	var structs []string
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			structs = append(structs, m[1])
		}
		if len(structs) >= 10 {
			break
		}
	}
	return structs
}

func extractTraits(content string, lang string) []string {
	var pattern string
	switch lang {
	case "rust":
		pattern = `(?:pub\s+)?trait\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	case "typescript":
		pattern = `interface\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	default:
		return nil
	}

	re := regexp.MustCompile(pattern)
	var traits []string
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			traits = append(traits, m[1])
		}
		if len(traits) >= 5 {
			break
		}
	}
	return traits
}

func detectPatterns(content string, lang string) []string {
	var patterns []string

	if strings.Contains(content, "async") && strings.Contains(content, "await") {
		patterns = append(patterns, "async")
	}

	switch lang {
	case "rust":
		if strings.Contains(content, "impl") && strings.Contains(content, "for") {
			patterns = append(patterns, "trait impl")
		}
		if strings.Contains(content, "#[derive") {
			patterns = append(patterns, "derive")
		}
		if strings.Contains(content, "Result<") || strings.Contains(content, "anyhow::") {
			patterns = append(patterns, "error handling")
		}
		if strings.Contains(content, "#[test]") {
			patterns = append(patterns, "tests")
		}
		if strings.Contains(content, "Box<dyn") || strings.Contains(content, "&dyn") {
			patterns = append(patterns, "dyn dispatch")
		}
	case "python":
		if strings.Contains(content, "@dataclass") {
			patterns = append(patterns, "dataclass")
		}
		if strings.Contains(content, "def __init__") {
			patterns = append(patterns, "OOP")
		}
	case "javascript", "typescript":
		if strings.Contains(content, "useState") || strings.Contains(content, "useEffect") {
			patterns = append(patterns, "React hooks")
		}
		if strings.Contains(content, "export default") {
			patterns = append(patterns, "ES modules")
		}
	}

	if len(patterns) > 3 {
		patterns = patterns[:3]
	}
	return patterns
}
