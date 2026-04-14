package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// smartCmd is an alias for local-llm (matches RTK command name)
var smartCmd = &cobra.Command{
	Use:   "smart <file>",
	Short: "Generate 2-line heuristic summary of a code file (alias for local-llm)",
	Long: `Analyze a source file and produce a 2-line technical summary using heuristics.

Line 1: What the file is (language, component counts, lines)
Line 2: Key details (imports, patterns, main definitions)

No external LLM required - uses pattern matching and static analysis.

Examples:
  tokman smart main.go
  tokman smart src/utils.py`,
	Args: cobra.ExactArgs(1),
	RunE: runLocalLlm,
}

var (
	smartUseLLM        bool
	smartLLMEndpoint   string
	smartModel         string
	smartForceDownload bool
)

func init() {
	registry.Add(func() { registry.Register(localLlmCmd) })
	registry.Add(func() { registry.Register(smartCmd) })

	// Add LLM flags to smart command
	smartCmd.Flags().BoolVar(&smartUseLLM, "llm", false, "Use local LLM for enhanced analysis (requires Ollama or compatible API)")
	smartCmd.Flags().StringVar(&smartLLMEndpoint, "llm-endpoint", "http://localhost:11434", "Local LLM API endpoint (default: Ollama)")
	smartCmd.Flags().StringVar(&smartModel, "model", "codellama", "Model to use for LLM analysis (default: codellama)")
	smartCmd.Flags().BoolVar(&smartForceDownload, "force-download", false, "Force re-download/pull the model before analysis")
}

func runLocalLlm(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Use LLM if requested
	if smartUseLLM {
		return runSmartWithLLM(filePath, content)
	}

	return runHeuristicAnalysis(filePath, content)
}

// runHeuristicAnalysis performs heuristic code analysis without LLM
func runHeuristicAnalysis(filePath string, content []byte) error {
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

// runSmartWithLLM uses a local LLM API (like Ollama) for enhanced code analysis
func runSmartWithLLM(filePath string, content []byte) error {
	// Force download if requested
	if smartForceDownload {
		if err := pullModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to pull model: %v\n", err)
		}
	}

	// Try to use LLM via API call
	llmOutput, err := queryLocalLLM(filePath, content)
	if err != nil {
		// Fall back to heuristic analysis
		fmt.Fprintf(os.Stderr, "LLM query failed (%v), using heuristic analysis\n", err)
		return runHeuristicAnalysis(filePath, content)
	}

	fmt.Println(llmOutput)
	return nil
}

// pullModel pulls the model from Ollama
func pullModel() error {
	fmt.Fprintf(os.Stderr, "Pulling model %s...\n", smartModel)

	reqBody := map[string]interface{}{
		"name": smartModel,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	pullURL := smartLLMEndpoint + "/api/pull"
	resp, err := http.Post(pullURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull returned status %d", resp.StatusCode)
	}

	fmt.Fprintf(os.Stderr, "Model %s ready\n", smartModel)
	return nil
}

// queryLocalLLM sends a request to a local LLM API (Ollama-compatible)
func queryLocalLLM(filePath string, content []byte) (string, error) {
	// Check if endpoint is reachable
	pingURL := smartLLMEndpoint + "/api/tags"
	resp, err := http.Get(pingURL)
	if err != nil {
		return "", fmt.Errorf("LLM endpoint not available: %w", err)
	}
	resp.Body.Close()

	// Prepare the prompt
	ext := filepath.Ext(filePath)
	lang := detectLang(ext)
	prompt := fmt.Sprintf(`Analyze this %s code file and provide a 2-line summary:
Line 1: What the file is (component type, line count)
Line 2: Key functionality or main purpose

File: %s

Code (first 2000 chars):
%s

Provide only the 2-line summary, nothing else.`, lang, filepath.Base(filePath), truncateContent(content, 2000))

	// Prepare request body
	reqBody := map[string]interface{}{
		"model":  smartModel,
		"prompt": prompt,
		"stream": false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Send request to LLM
	generateURL := smartLLMEndpoint + "/api/generate"
	resp, err = http.Post(generateURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Clean up the response
	summary := strings.TrimSpace(result.Response)
	lines := strings.Split(summary, "\n")

	// Ensure we only return 2 lines
	if len(lines) >= 2 {
		return strings.Join(lines[:2], "\n"), nil
	}

	return summary, nil
}

// truncateContent truncates content to max length
func truncateContent(content []byte, maxLen int) string {
	if len(content) > maxLen {
		return string(content[:maxLen]) + "\n... (truncated)"
	}
	return string(content)
}
