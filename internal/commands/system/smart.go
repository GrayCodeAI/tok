package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var smartFileCmd = &cobra.Command{
	Use:   "smart <file>",
	Short: "Generate 2-line heuristic summary of a file",
	Long: `Generate a compact 2-line summary of a file using heuristics.

This command analyzes file content and produces a brief summary:
- For code: shows structure (functions, classes, imports)
- For configs: shows key settings
- For docs: shows headings and key points

Examples:
  tok smart main.go           # Go file summary
  tok smart package.json      # NPM package summary
  tok smart README.md         # Documentation summary
  tok smart Cargo.toml        # Rust project summary`,
	Args: cobra.ExactArgs(1),
	RunE: runSmart,
}

func init() {
	registry.Add(func() { registry.Register(smartFileCmd) })
}

func runSmart(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Open and read file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	timer := tracking.Start()

	// Detect file type
	ext := strings.ToLower(filepath.Ext(filePath))
	language := detectLanguage(ext, filePath)

	// Generate summary
	summary := generateSummary(file, language, ext)

	// Print with styling
	out.Global().Println()
	out.Global().Printf("%s %s\n", color.New(color.Bold).Sprint("📄"), color.New(color.Bold).Sprint(filePath))
	out.Global().Println(color.GreenString("┌─" + strings.Repeat("─", 58)))
	out.Global().Printf("│ %s\n", summary.Line1)
	out.Global().Printf("│ %s\n", summary.Line2)
	out.Global().Println(color.GreenString("└─" + strings.Repeat("─", 58)))
	out.Global().Println()

	// Track usage
	timer.Track("smart "+filePath, "tok smart", 100, 50)

	return nil
}

// Summary represents a 2-line file summary
type Summary struct {
	Line1 string
	Line2 string
}

func generateSummary(file *os.File, language, ext string) Summary {
	scanner := bufio.NewScanner(file)

	switch language {
	case "go":
		return summarizeGo(scanner)
	case "rust":
		return summarizeRust(scanner)
	case "javascript", "typescript":
		return summarizeJS(scanner)
	case "python":
		return summarizePython(scanner)
	case "json":
		return summarizeJSON(scanner, ext)
	case "yaml", "yml", "toml":
		return summarizeConfig(scanner, language)
	case "markdown":
		return summarizeMarkdown(scanner)
	default:
		return summarizeGeneric(scanner)
	}
}

func summarizeGo(scanner *bufio.Scanner) Summary {
	var imports []string
	var structs, funcs int
	var packageName string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "package ") {
			packageName = strings.TrimSpace(strings.TrimPrefix(line, "package"))
		}

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "\t\"") {
			// Extract import path
			if idx := strings.Index(line, "\""); idx != -1 {
				end := strings.Index(line[idx+1:], "\"")
				if end != -1 {
					imp := line[idx+1 : idx+1+end]
					parts := strings.Split(imp, "/")
					if len(parts) > 0 {
						imports = append(imports, parts[len(parts)-1])
					}
				}
			}
		}

		if strings.HasPrefix(line, "type ") && strings.Contains(line, " struct") {
			structs++
		}

		if strings.HasPrefix(line, "func ") {
			funcs++
		}
	}

	line1 := fmt.Sprintf("📦 %s", packageName)
	if len(imports) > 0 {
		impStr := strings.Join(unique(imports[:min(3, len(imports))]), ", ")
		if len(imports) > 3 {
			impStr += fmt.Sprintf(" +%d more", len(imports)-3)
		}
		line1 += fmt.Sprintf(" | imports: %s", impStr)
	}

	line2 := fmt.Sprintf("🔧 %d function", funcs)
	if funcs != 1 {
		line2 += "s"
	}
	if structs > 0 {
		line2 += fmt.Sprintf(" | %d struct", structs)
		if structs != 1 {
			line2 += "s"
		}
	}

	return Summary{Line1: line1, Line2: line2}
}

func summarizeRust(scanner *bufio.Scanner) Summary {
	var structs, impls, funcs int
	var crateName string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "mod ") || strings.HasPrefix(line, "pub mod ") {
			if crateName == "" {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					crateName = parts[len(parts)-1]
				}
			}
		}

		if strings.HasPrefix(line, "struct ") || strings.HasPrefix(line, "pub struct ") {
			structs++
		}

		if strings.HasPrefix(line, "impl ") {
			impls++
		}

		if strings.HasPrefix(line, "fn ") || strings.HasPrefix(line, "pub fn ") {
			funcs++
		}
	}

	line1 := "📦 Rust module"
	if crateName != "" {
		line1 = fmt.Sprintf("📦 %s", crateName)
	}

	line2 := fmt.Sprintf("🔧 %d fn", funcs)
	if structs > 0 {
		line2 += fmt.Sprintf(" | %d structs", structs)
	}
	if impls > 0 {
		line2 += fmt.Sprintf(" | %d impls", impls)
	}

	return Summary{Line1: line1, Line2: line2}
}

func summarizeJS(scanner *bufio.Scanner) Summary {
	var imports, exports, funcs, classes int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "require(") {
			imports++
		}

		if strings.HasPrefix(line, "export ") || strings.HasPrefix(line, "module.exports") {
			exports++
		}

		if strings.HasPrefix(line, "function ") || strings.HasPrefix(line, "const ") && strings.Contains(line, "= (") {
			funcs++
		}

		if strings.HasPrefix(line, "class ") {
			classes++
		}
	}

	line1 := fmt.Sprintf("📦 JS module | %d import", imports)
	if imports != 1 {
		line1 += "s"
	}

	line2 := fmt.Sprintf("🔧 %d fn", funcs)
	if classes > 0 {
		line2 += fmt.Sprintf(" | %d class", classes)
		if classes != 1 {
			line2 += "es"
		}
	}
	if exports > 0 {
		line2 += fmt.Sprintf(" | %d export", exports)
		if exports != 1 {
			line2 += "s"
		}
	}

	return Summary{Line1: line1, Line2: line2}
}

func summarizePython(scanner *bufio.Scanner) Summary {
	var imports, funcs, classes int
	var hasMain bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			imports++
		}

		if strings.HasPrefix(line, "def ") && !strings.HasPrefix(line, "def _") {
			funcs++
		}

		if strings.HasPrefix(line, "class ") {
			classes++
		}

		if strings.Contains(line, `if __name__ == "__main__"`) {
			hasMain = true
		}
	}

	line1 := fmt.Sprintf("🐍 Python | %d import", imports)
	if imports != 1 {
		line1 += "s"
	}

	line2 := fmt.Sprintf("🔧 %d fn", funcs)
	if classes > 0 {
		line2 += fmt.Sprintf(" | %d class", classes)
		if classes != 1 {
			line2 += "es"
		}
	}
	if hasMain {
		line2 += " | executable"
	}

	return Summary{Line1: line1, Line2: line2}
}

func summarizeJSON(scanner *bufio.Scanner, ext string) Summary {
	content := ""
	for scanner.Scan() {
		content += scanner.Text()
	}

	// Count top-level keys
	keys := 0
	if ext == ".json" {
		// Simple counting of top-level keys
		for _, c := range content {
			if c == '"' {
				keys++
			}
		}
		keys = keys / 2 // Each key has opening and closing quote
	}

	size := len(content)
	sizeStr := fmt.Sprintf("%d bytes", size)
	if size > 1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
	}

	line1 := fmt.Sprintf("📄 JSON document | %s", sizeStr)
	line2 := fmt.Sprintf("🔧 ~%d top-level keys", keys/2)

	return Summary{Line1: line1, Line2: line2}
}

func summarizeConfig(scanner *bufio.Scanner, lang string) Summary {
	var sections []string
	var keyCount int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sections = append(sections, line[1:len(line)-1])
		}

		if strings.Contains(line, "=") || strings.Contains(line, ": ") {
			keyCount++
		}
	}

	line1 := fmt.Sprintf("⚙️  %s config", strings.ToUpper(lang))
	if len(sections) > 0 {
		line1 += fmt.Sprintf(" | %d sections", len(sections))
	}

	line2 := fmt.Sprintf("🔧 %d configuration keys", keyCount)

	return Summary{Line1: line1, Line2: line2}
}

func summarizeMarkdown(scanner *bufio.Scanner) Summary {
	var headings []string
	var codeBlocks, links int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# ") {
			headings = append(headings, strings.TrimSpace(line[2:]))
		} else if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimSpace(line[3:]))
		}

		if strings.HasPrefix(line, "```") {
			codeBlocks++
		}

		if strings.Contains(line, "](") {
			links++
		}
	}

	title := "Documentation"
	if len(headings) > 0 {
		title = headings[0]
		if len(title) > 40 {
			title = title[:37] + "..."
		}
	}

	line1 := fmt.Sprintf("📖 %s", title)

	line2 := fmt.Sprintf("📋 %d section", len(headings))
	if len(headings) != 1 {
		line2 += "s"
	}
	if codeBlocks > 0 {
		line2 += fmt.Sprintf(" | %d code blocks", codeBlocks/2)
	}
	if links > 0 {
		line2 += fmt.Sprintf(" | %d links", links)
	}

	return Summary{Line1: line1, Line2: line2}
}

func summarizeGeneric(scanner *bufio.Scanner) Summary {
	var lines, nonEmpty, words int

	for scanner.Scan() {
		lines++
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			nonEmpty++
			words += len(strings.Fields(line))
		}
	}

	line1 := fmt.Sprintf("📄 Text file | %d lines", lines)
	line2 := fmt.Sprintf("📝 %d words", words)

	return Summary{Line1: line1, Line2: line2}
}

func detectLanguage(ext, filePath string) string {
	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".js", ".mjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".md", ".markdown":
		return "markdown"
	default:
		// Check filename for common files
		base := filepath.Base(filePath)
		switch base {
		case "Makefile", "makefile":
			return "makefile"
		case "Dockerfile":
			return "dockerfile"
		case "Cargo.toml":
			return "toml"
		case "package.json":
			return "json"
		case "go.mod":
			return "go"
		}
		return "generic"
	}
}

func unique(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// isLetter checks if a rune is a letter
func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}
