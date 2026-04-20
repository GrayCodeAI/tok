package compressor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lakshmanpatel/tok/internal/output"
)

type FileCompressionResult struct {
	OriginalTokens   int
	CompressedTokens int
	SavingsPercent   float64
	BackupPath       string
}

// CompressFile compresses a file and creates a backup
func CompressFile(filename string, mode string) error {
	return CompressFileWithOptions(filename, mode, false)
}

// CompressFileWithOptions compresses a file with optional overwrite behavior.
func CompressFileWithOptions(filename string, mode string, force bool) error {
	result, _, err := CompressFilePreviewWithOptions(filename, mode, force)
	if err != nil {
		return err
	}

	// Read original again for write path to keep behavior simple and explicit.
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	original := string(content)
	compressed, err := compressFileContent(original, mode)
	if err != nil {
		return err
	}

	if err := os.WriteFile(result.BackupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	if err := os.WriteFile(filename, []byte(compressed), 0644); err != nil {
		os.Rename(result.BackupPath, filename)
		return fmt.Errorf("failed to write compressed file: %w", err)
	}

	p := output.Global()
	p.Printf("Compressed: %s\n", filename)
	p.Printf("Backup: %s\n", result.BackupPath)
	p.Printf("Tokens: %d → %d (%.1f%% saved)\n", result.OriginalTokens, result.CompressedTokens, result.SavingsPercent)
	return nil
}

// CompressFilePreview returns what compression would produce without writing.
func CompressFilePreview(filename string, mode string) (FileCompressionResult, string, error) {
	return CompressFilePreviewWithOptions(filename, mode, false)
}

// CompressFilePreviewWithOptions returns preview details with optional overwrite behavior.
func CompressFilePreviewWithOptions(filename string, mode string, force bool) (FileCompressionResult, string, error) {
	var result FileCompressionResult

	// Read original
	content, err := os.ReadFile(filename)
	if err != nil {
		return result, "", fmt.Errorf("failed to read file: %w", err)
	}

	original := string(content)

	// Check if already compressed (has .original backup)
	backupPath := backupPathFor(filename)
	if _, err := os.Stat(backupPath); err == nil && !force {
		return result, "", fmt.Errorf("file already compressed (backup exists: %s)", backupPath)
	}

	// Compress content
	compressed, err := compressFileContent(original, mode)
	if err != nil {
		return result, "", err
	}

	// Calculate savings
	originalTokens := estimateTokens(original)
	compressedTokens := estimateTokens(compressed)
	savings := 0.0
	if originalTokens > 0 {
		savings = float64(originalTokens-compressedTokens) / float64(originalTokens) * 100
	}
	result = FileCompressionResult{
		OriginalTokens:   originalTokens,
		CompressedTokens: compressedTokens,
		SavingsPercent:   savings,
		BackupPath:       backupPath,
	}
	return result, compressed, nil
}

// RestoreFile restores from backup
func RestoreFile(filename string) error {
	backupPath := backupPathFor(filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		legacy := filename + ".original"
		if _, legacyErr := os.Stat(legacy); legacyErr == nil {
			backupPath = legacy
		}
	}

	// Check backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found: %s", backupPath)
	}

	// Read backup
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Restore
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove backup
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	output.Global().Printf("Restored: %s\n", filename)
	return nil
}

func backupPathFor(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".md" {
		return strings.TrimSuffix(filename, ext) + ".original.md"
	}
	return filename + ".original"
}

// compressFileContent compresses file content while preserving structure
func compressFileContent(content, mode string) (string, error) {
	// Split into lines to preserve structure
	lines := strings.Split(content, "\n")
	var result []string

	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Preserve code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// Preserve code blocks, headings, URLs, file paths
		if inCodeBlock ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "http") ||
			strings.HasPrefix(trimmed, "- http") ||
			strings.HasPrefix(trimmed, "- `/") ||
			strings.HasPrefix(trimmed, "- `./") ||
			strings.HasPrefix(trimmed, "`") {
			result = append(result, line)
			continue
		}

		// Compress prose lines
		protected, restore := protectInlineCode(line)
		compressed, err := Compress(protected, mode)
		if err != nil {
			return "", err
		}
		result = append(result, restore(compressed))
	}

	return strings.Join(result, "\n"), nil
}

func protectInlineCode(line string) (string, func(string) string) {
	inline := regexp.MustCompile("`[^`]+`")
	matches := inline.FindAllString(line, -1)
	if len(matches) == 0 {
		return line, func(s string) string { return s }
	}

	protected := line
	for i, m := range matches {
		token := fmt.Sprintf("__TOK_INLINE_%d__", i)
		protected = strings.Replace(protected, m, token, 1)
	}

	return protected, func(s string) string {
		restored := s
		for i, m := range matches {
			token := fmt.Sprintf("__TOK_INLINE_%d__", i)
			restored = strings.ReplaceAll(restored, token, m)
		}
		return restored
	}
}

// estimateTokens roughly estimates token count (very approximate)
func estimateTokens(text string) int {
	// Simple approximation: ~4 chars per token for English
	return len(text) / 4
}

// ValidateFile checks if file can be safely compressed
func ValidateFile(filename string) error {
	if strings.Contains(strings.ToLower(filename), ".original.") || strings.HasSuffix(strings.ToLower(filename), ".original") {
		return fmt.Errorf("backup file detected; refusing to compress already-backed-up file")
	}

	// Check file exists
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check not a directory
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	safeExts := []string{".md", ".txt", ".rst"}
	isSafe := false
	for _, safe := range safeExts {
		if ext == safe {
			isSafe = true
			break
		}
	}

	if !isSafe {
		return fmt.Errorf("unsafe file type: %s (only .md, .txt, .rst allowed)", ext)
	}

	// Check for sensitive patterns
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	sensitivePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)password\s*=`),
		regexp.MustCompile(`(?i)api[_-]?key\s*=`),
		regexp.MustCompile(`(?i)secret\s*=`),
		regexp.MustCompile(`(?i)token\s*=`),
		regexp.MustCompile(`(?i)private[_-]?key`),
	}

	for _, pattern := range sensitivePatterns {
		if pattern.Match(content) {
			return fmt.Errorf("file may contain sensitive data (passwords/keys detected)")
		}
	}

	return nil
}

// BatchCompress compresses multiple files
func BatchCompress(pattern string, mode string) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no files match pattern: %s", pattern)
	}

	p := output.Global()
	for _, file := range matches {
		if err := ValidateFile(file); err != nil {
			p.Errorf("Skipping %s: %v\n", file, err)
			continue
		}

		if err := CompressFile(file, mode); err != nil {
			p.Errorf("Failed %s: %v\n", file, err)
			continue
		}
	}

	return nil
}
