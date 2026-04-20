package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/core"
	"github.com/lakshmanpatel/tok/internal/filter"
)

var (
	cmMode    string
	cmRestore bool
)

var compressMemoryCmd = &cobra.Command{
	Use:   "compress-memory [file]",
	Short: "Compress memory/instruction files for AI agents",
	Long: `Compress prose in memory/instruction files (CLAUDE.md, GEMINI.md, project notes)
so the AI reads fewer tokens every session.

How it works:
  tok compress-memory CLAUDE.md
  → Saves compressed CLAUDE.md, backs up original as CLAUDE.original.md
  → AI reads compressed version (fewer tokens)
  → Human reads/edits the .original.md file

Modes:
  lite         - Keep grammar, drop filler words
  full         - Drop articles, fragments OK (default)
  ultra        - Maximum compression, abbreviations
  wenyan-lite  - Classical register, filler stripped, grammar kept
  wenyan-full  - Fragments + arrow causality + abbreviations
  wenyan-ultra - Extreme abbreviation, no connectives, symbolic

Restore:
  tok compress-memory CLAUDE.md --restore
  → Reverts CLAUDE.md from CLAUDE.original.md`,
	Example: `  tok compress-memory CLAUDE.md
  tok compress-memory docs/notes.md --mode ultra
  tok compress-memory CLAUDE.md --restore`,
	Args: cobra.ExactArgs(1),
	RunE: runCompressMemory,
}

func runCompressMemory(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if cmRestore {
		return restoreMemoryFile(filePath)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	original, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	originalText := string(original)
	originalTokens := core.EstimateTokensPrecise(originalText)

	compressedText := compressProse(originalText, cmMode)
	compressedTokens := core.EstimateTokensPrecise(compressedText)

	saved := originalTokens - compressedTokens
	savingsPct := float64(0)
	if originalTokens > 0 {
		savingsPct = float64(saved) / float64(originalTokens) * 100
	}

	backupPath := absPath + ".original.md"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := os.WriteFile(backupPath, original, 0644); err != nil {
			return fmt.Errorf("cannot create backup: %w", err)
		}
	}

	if err := os.WriteFile(absPath, []byte(compressedText), 0644); err != nil {
		return fmt.Errorf("cannot write compressed file: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(bold("tok compress-memory"))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Printf("  File:       %s\n", cyan(absPath))
	fmt.Printf("  Mode:       %s\n", cmMode)
	fmt.Println()
	fmt.Printf("  Original:   %d tokens\n", originalTokens)
	fmt.Printf("  Compressed: %d tokens\n", compressedTokens)
	fmt.Printf("  Saved:      %s tokens (%s)\n", green(fmt.Sprintf("%d", saved)), green(fmt.Sprintf("%.1f%%", savingsPct)))
	fmt.Println()
	fmt.Printf("  Backup:     %s\n", yellow(backupPath))
	fmt.Println()
	fmt.Printf("%s Edit the .original.md file. AI reads the compressed version.\n", green("✓"))
	fmt.Println()

	return nil
}

func restoreMemoryFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	backupPath := absPath + ".original.md"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	original, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("cannot read backup: %w", err)
	}

	if err := os.WriteFile(absPath, original, 0644); err != nil {
		return fmt.Errorf("cannot restore file: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Restored %s from backup\n", green("✓"), absPath)
	return nil
}

func compressProse(input, mode string) string {
	lines := strings.Split(input, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		if inCodeBlock {
			result = append(result, line)
			continue
		}

		if isProtectedLine(trimmed) {
			result = append(result, line)
			continue
		}

		compressed := compressLine(line, mode)
		result = append(result, compressed)
	}

	return strings.Join(result, "\n")
}

func isProtectedLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}

	protectedPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^#{1,6}\s`),
		regexp.MustCompile(`^[-*+]\s`),
		regexp.MustCompile(`^\d+\.`),
		regexp.MustCompile(`^https?://`),
		regexp.MustCompile(`^[\w./~-]+/\S+\s`),
		regexp.MustCompile(`^import\s`),
		regexp.MustCompile(`^export\s`),
		regexp.MustCompile(`^const\s`),
		regexp.MustCompile(`^let\s`),
		regexp.MustCompile(`^var\s`),
		regexp.MustCompile(`^func\s`),
		regexp.MustCompile(`^class\s`),
		regexp.MustCompile(`^\[.*\]\(.*\)`),
		regexp.MustCompile(`^\|.*\|`),
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`),
		regexp.MustCompile(`^v?\d+\.\d+`),
	}

	for _, p := range protectedPatterns {
		if p.MatchString(trimmed) {
			return true
		}
	}

	return false
}

func compressLine(line, mode string) string {
	indent := leadingWhitespace(line)
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return line
	}

	compressed := trimmed

	switch mode {
	case "lite":
		compressed = compressLite(compressed)
	case "full":
		compressed = compressFull(compressed)
	case "ultra":
		compressed = compressUltra(compressed)
	case "wenyan-lite":
		compressed = filter.WenyanCompress(compressed, filter.WenyanLite)
	case "wenyan-full":
		compressed = filter.WenyanCompress(compressed, filter.WenyanFull)
	case "wenyan-ultra":
		compressed = filter.WenyanCompress(compressed, filter.WenyanUltra)
	default:
		compressed = compressFull(compressed)
	}

	return indent + compressed
}

func compressLite(text string) string {
	fillers := []string{
		"just ", "really ", "basically ", "actually ", "simply ",
		"very ", "quite ", "rather ", "somewhat ", "pretty ",
		"obviously ", "clearly ", "certainly ", "definitely ",
		"please note that ", "it is important to note that ",
		"keep in mind that ", "remember that ",
	}
	for _, f := range fillers {
		text = caseInsensitiveReplace(text, f, "")
	}
	return text
}

func compressFull(text string) string {
	text = compressLite(text)

	articles := []string{" the ", " a ", " an ", " The ", " A ", " An "}
	for _, a := range articles {
		text = strings.ReplaceAll(text, a, " ")
	}

	redundant := map[string]string{
		"in order to":           "to",
		"due to the fact that":  "because",
		"at this point in time": "now",
		"for the purpose of":    "for",
		"in the event that":     "if",
		"with regard to":        "about",
		"with respect to":       "about",
		"on the basis of":       "by",
		"as a result of":        "from",
		"in relation to":        "to",
		"prior to":              "before",
		"subsequent to":         "after",
		"in addition to":        "plus",
		"has the ability to":    "can",
		"is able to":            "can",
		"there is":              "",
		"there are":             "",
		"it is":                 "",
	}
	for k, v := range redundant {
		text = caseInsensitiveReplace(text, k, v)
	}

	text = collapseSpaces(text)
	return text
}

func compressUltra(text string) string {
	text = compressFull(text)

	ultraReplacements := map[string]string{
		"configuration":  "config",
		"implementation": "impl",
		"documentation":  "docs",
		"development":    "dev",
		"environment":    "env",
		"application":    "app",
		"authentication": "auth",
		"authorization":  "authz",
		"information":    "info",
		"directory":      "dir",
		"parameter":      "param",
		"parameters":     "params",
		"arguments":      "args",
		"argument":       "arg",
		"function":       "fn",
		"functions":      "fns",
		"variable":       "var",
		"variables":      "vars",
		"database":       "db",
		"connection":     "conn",
		"connections":    "conns",
		"response":       "resp",
		"responses":      "resps",
		"request":        "req",
		"requests":       "reqs",
		"message":        "msg",
		"messages":       "msgs",
		"error":          "err",
		"errors":         "errs",
		"warning":        "warn",
		"warnings":       "warns",
		"example":        "ex",
		"examples":       "exs",
		"without":        "w/o",
		"with":           "w/",
		"through":        "thru",
		"should":         "shld",
		"would":          "wld",
		"could":          "cld",
		"cannot":         "can't",
		"do not":         "don't",
		"does not":       "doesn't",
		"is not":         "isn't",
		"are not":        "aren't",
		"was not":        "wasn't",
		"were not":       "weren't",
		"have not":       "haven't",
		"has not":        "hasn't",
		"had not":        "hadn't",
		"will not":       "won't",
		"not":            "",
	}
	for k, v := range ultraReplacements {
		text = caseInsensitiveReplace(text, k, v)
	}

	text = collapseSpaces(text)
	return text
}

func caseInsensitiveReplace(text, old, new string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(old))
	return re.ReplaceAllString(text, new)
}

func collapseSpaces(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func leadingWhitespace(line string) string {
	var indent strings.Builder
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent.WriteRune(ch)
		} else {
			break
		}
	}
	return indent.String()
}

func init() {
	registry.Add(func() { registry.Register(compressMemoryCmd) })

	compressMemoryCmd.Flags().StringVar(&cmMode, "mode", "full", "Compression mode: lite|full|ultra|wenyan-lite|wenyan-full|wenyan-ultra")
	compressMemoryCmd.Flags().BoolVar(&cmRestore, "restore", false, "Restore original from backup")

	compressMemoryCmd.Aliases = []string{"md", "compress-md"}
}
