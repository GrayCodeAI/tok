package contextread

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lakshmanpatel/tok/internal/config"
)

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

var signaturePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*(func|type|var|const|package|import)\b`),
	regexp.MustCompile(`^\s*(class|interface|enum|struct|def|async\s+def)\b`),
	regexp.MustCompile(`^\s*(public|private|protected|export|module)\b`),
	regexp.MustCompile(`^\s*[@#]`),
}

// IsStub reports whether this package is a placeholder implementation.
func IsStub() bool {
	return false
}

// TrackedCommandPatternsForKind returns tracked command patterns for a kind.
func TrackedCommandPatternsForKind(kind string) []string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "read":
		return []string{"tok read", "tok ctx read"}
	case "delta":
		return []string{"tok ctx delta"}
	case "mcp":
		return []string{"tok mcp", "tok proxy"}
	default:
		return nil
	}
}

// TrackedCommandPatterns returns all tracked command patterns.
func TrackedCommandPatterns() []string {
	return []string{
		"tok read",
		"tok ctx read",
		"tok ctx delta",
		"tok mcp",
		"tok proxy",
	}
}

// Build builds context for a file with budget-aware shaping.
func Build(path, content, lang string, opts Options) (string, int, int, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	originalTokens := estimateTokens(normalized)
	mode := normalizeMode(opts.Mode)
	level := normalizeLevel(opts.Level)

	working := normalized
	if opts.StartLine > 0 || opts.EndLine > 0 {
		working = sliceLines(working, opts.StartLine, opts.EndLine)
	}

	switch mode {
	case "delta":
		delta, err := buildDelta(path, working)
		if err != nil {
			return "", 0, 0, err
		}
		working = delta
	case "map":
		working = buildOutline(path, working)
	case "signatures":
		working = extractSignatures(path, working)
	case "graph":
		working = buildGraphView(path, working, opts.RelatedFilesCount)
	case "entropy", "aggressive":
		working = applyAggressiveFilter(working)
	case "full":
		// Preserve content.
	default:
		if level == "none" {
			// Preserve content.
		} else if level == "aggressive" {
			working = applyAggressiveFilter(working)
		} else {
			working = applyMinimalFilter(working)
		}
	}

	if opts.MaxLines > 0 {
		working = trimToLines(working, opts.MaxLines)
	}
	if opts.MaxTokens > 0 {
		working = trimToTokens(working, opts.MaxTokens)
	}
	if opts.LineNumbers {
		working = addLineNumbers(working, opts.StartLine)
	}

	if opts.SaveSnapshot && path != "" && path != "stdin" {
		if err := saveSnapshot(path, normalized); err != nil {
			return "", 0, 0, err
		}
	}

	filteredTokens := estimateTokens(working)
	return working, originalTokens, filteredTokens, nil
}

// Analyze analyzes content for context shape.
func Analyze(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	nonEmpty := 0
	longLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
		if len(line) > 120 {
			longLines++
		}
	}
	return fmt.Sprintf("%d lines, %d non-empty, %d long lines", len(lines), nonEmpty, longLines)
}

// Describe describes a file for context UI.
func Describe(path string) string {
	base := filepath.Base(path)
	ext := strings.TrimPrefix(filepath.Ext(base), ".")
	if ext == "" {
		return base
	}
	return fmt.Sprintf("%s (%s)", base, ext)
}

func normalizeMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto":
		return "auto"
	case "full", "map", "signatures", "aggressive", "entropy", "lines", "delta", "graph":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return "auto"
	}
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "aggressive":
		return "aggressive"
	case "none":
		return "none"
	default:
		return "minimal"
	}
}

func applyMinimalFilter(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	blankRun := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blankRun++
			if blankRun > 1 {
				continue
			}
			out = append(out, "")
			continue
		}
		blankRun = 0
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func applyAggressiveFilter(content string) string {
	lines := strings.Split(applyMinimalFilter(content), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if looksLikeNoise(trimmed) {
			continue
		}
		out = append(out, line)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

func looksLikeNoise(line string) bool {
	prefixes := []string{
		"import ", "from ", "using ", "require(", "console.log(", "fmt.Println(",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func extractSignatures(path, content string) string {
	lines := strings.Split(content, "\n")
	out := []string{fmt.Sprintf("# Signatures for %s", filepath.Base(path))}
	for _, line := range lines {
		for _, pattern := range signaturePatterns {
			if pattern.MatchString(line) {
				out = append(out, strings.TrimRight(line, " \t"))
				break
			}
		}
	}
	if len(out) == 1 {
		return trimToLines(applyMinimalFilter(content), 40)
	}
	return strings.Join(out, "\n")
}

func buildOutline(path, content string) string {
	lines := strings.Split(content, "\n")
	out := []string{
		fmt.Sprintf("# Outline for %s", filepath.Base(path)),
		fmt.Sprintf("Description: %s", Describe(path)),
		"",
	}
	kept := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if isSectionLine(trimmed) {
			out = append(out, "- "+trimmed)
			kept++
		}
	}
	if kept == 0 {
		out = append(out, trimToLines(applyMinimalFilter(content), 25))
	}
	return strings.Join(out, "\n")
}

func isSectionLine(line string) bool {
	if strings.HasPrefix(line, "#") {
		return true
	}
	for _, pattern := range signaturePatterns {
		if pattern.MatchString(line) {
			return true
		}
	}
	return false
}

func buildGraphView(path, content string, relatedCount int) string {
	related := relatedCount
	if related <= 0 {
		related = 3
	}
	return strings.Join([]string{
		fmt.Sprintf("# Graph Context for %s", filepath.Base(path)),
		fmt.Sprintf("Primary: %s", path),
		fmt.Sprintf("Requested related files: %d", related),
		"",
		buildOutline(path, content),
	}, "\n")
}

func trimToLines(content string, maxLines int) string {
	if maxLines <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	head := maxLines
	if head < 1 {
		head = 1
	}
	return strings.Join(lines[:head], "\n")
}

func trimToTokens(content string, maxTokens int) string {
	if maxTokens <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	used := 0
	for _, line := range lines {
		lineTokens := estimateTokens(line)
		if used+lineTokens > maxTokens {
			break
		}
		out = append(out, line)
		used += lineTokens
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func addLineNumbers(content string, startLine int) string {
	lines := strings.Split(content, "\n")
	base := startLine
	if base <= 0 {
		base = 1
	}
	for i, line := range lines {
		lines[i] = fmt.Sprintf("%4d | %s", base+i, line)
	}
	return strings.Join(lines, "\n")
}

func sliceLines(content string, startLine, endLine int) string {
	lines := strings.Split(content, "\n")
	start := 1
	if startLine > 0 {
		start = startLine
	}
	end := len(lines)
	if endLine > 0 && endLine < end {
		end = endLine
	}
	if start > end || start > len(lines) {
		return ""
	}
	return strings.Join(lines[start-1:end], "\n")
}

func buildDelta(path, current string) (string, error) {
	previous, err := loadSnapshot(path)
	if err != nil {
		return "", err
	}
	if previous == "" {
		return strings.Join([]string{
			fmt.Sprintf("# Delta for %s", filepath.Base(path)),
			"No previous snapshot found. Current content follows.",
			"",
			trimToLines(applyMinimalFilter(current), 80),
		}, "\n"), nil
	}

	prevLines := strings.Split(previous, "\n")
	currLines := strings.Split(current, "\n")
	maxLen := len(currLines)
	if len(prevLines) > maxLen {
		maxLen = len(prevLines)
	}

	changes := []string{fmt.Sprintf("# Delta for %s", filepath.Base(path))}
	for i := 0; i < maxLen; i++ {
		var prev, curr string
		if i < len(prevLines) {
			prev = prevLines[i]
		}
		if i < len(currLines) {
			curr = currLines[i]
		}
		switch {
		case prev == curr:
			continue
		case prev == "":
			changes = append(changes, fmt.Sprintf("+ %s", curr))
		case curr == "":
			changes = append(changes, fmt.Sprintf("- %s", prev))
		default:
			changes = append(changes, fmt.Sprintf("~ %s", curr))
		}
	}

	if len(changes) == 1 {
		changes = append(changes, "No changes since last snapshot.")
	}
	return strings.Join(changes, "\n"), nil
}

func estimateTokens(content string) int {
	if content == "" {
		return 0
	}
	return len(content) / 4
}

func snapshotDir() string {
	return filepath.Join(config.DataPath(), "contextread")
}

func snapshotPath(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return filepath.Join(snapshotDir(), hex.EncodeToString(sum[:])+".snap")
}

func saveSnapshot(path, content string) error {
	if err := os.MkdirAll(snapshotDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(snapshotPath(path), []byte(content), 0644)
}

func loadSnapshot(path string) (string, error) {
	data, err := os.ReadFile(snapshotPath(path))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}
