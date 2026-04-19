// Package tee provides raw output recovery for command failures.
//
// The tee system saves unfiltered command output to disk when commands fail,
// allowing LLMs to access the full output without re-executing the command.
// This is especially useful for debugging failed tests or builds.
package tee

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lakshmanpatel/tok/internal/config"
)

const (
	// MinTeeSize is the minimum output size to tee (smaller outputs don't need recovery).
	MinTeeSize = 500

	// DefaultMaxFiles is the default maximum number of tee files to keep.
	DefaultMaxFiles = 20

	// DefaultMaxFileSize is the default maximum file size (1MB).
	DefaultMaxFileSize = 1_048_576
)

// TeeMode controls when tee writes files.
type TeeMode string

const (
	// TeeModeFailures only tees on command failure (default).
	TeeModeFailures TeeMode = "failures"
	// TeeModeAlways always tees output regardless of exit code.
	TeeModeAlways TeeMode = "always"
	// TeeModeNever disables tee entirely.
	TeeModeNever TeeMode = "never"
)

// TeeConfig holds configuration for the tee feature.
type TeeConfig struct {
	Enabled     bool    `toml:"enabled" mapstructure:"enabled"`
	Mode        TeeMode `toml:"mode" mapstructure:"mode"`
	MaxFiles    int     `toml:"max_files" mapstructure:"max_files"`
	MaxFileSize int     `toml:"max_file_size" mapstructure:"max_file_size"`
	Directory   string  `toml:"directory,omitempty" mapstructure:"directory"`
}

// DefaultTeeConfig returns the default tee configuration.
func DefaultTeeConfig() TeeConfig {
	return TeeConfig{
		Enabled:     true,
		Mode:        TeeModeFailures,
		MaxFiles:    DefaultMaxFiles,
		MaxFileSize: DefaultMaxFileSize,
		Directory:   "",
	}
}

// sanitizeSlug sanitizes a command slug for use in filenames.
// Replaces non-alphanumeric chars (except underscore/hyphen) with underscore,
// truncates at 40 chars.
func sanitizeSlug(slug string) string {
	var result strings.Builder
	for _, c := range slug {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result.WriteRune(c)
		} else {
			result.WriteRune('_')
		}
	}
	sanitized := result.String()
	if len(sanitized) > 40 {
		return sanitized[:40]
	}
	return sanitized
}

// getTeeDir returns the tee directory, respecting config and env overrides.
func getTeeDir(cfg *config.Config) (string, error) {
	// Env var override takes precedence
	if dir := os.Getenv("TOKMAN_TEE_DIR"); dir != "" {
		return dir, nil
	}

	// Config override
	if cfg != nil && cfg.Hooks.TeeDir != "" {
		return cfg.Hooks.TeeDir, nil
	}

	// Default: ~/.local/share/tok/tee/
	return filepath.Join(config.DataPath(), "tee"), nil
}

// cleanupOldFiles rotates old tee files, keeping only the last maxFiles.
func cleanupOldFiles(dir string, maxFiles int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter to .log files
	var logFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
			logFiles = append(logFiles, entry)
		}
	}

	if len(logFiles) <= maxFiles {
		return nil
	}

	// Sort by filename (which starts with epoch timestamp = chronological)
	// We need to sort in ascending order to delete oldest first
	for i := 0; i < len(logFiles)-1; i++ {
		for j := i + 1; j < len(logFiles); j++ {
			if logFiles[i].Name() > logFiles[j].Name() {
				logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
			}
		}
	}

	// Remove oldest files
	toRemove := len(logFiles) - maxFiles
	for i := 0; i < toRemove && i < len(logFiles); i++ {
		path := filepath.Join(dir, logFiles[i].Name())
		_ = os.Remove(path)
	}

	return nil
}

// shouldTee checks if tee should write based on config, mode, exit code, and size.
// Returns (shouldTee bool, teeDir string, err error).
func shouldTee(enabled bool, mode TeeMode, rawLen int, exitCode int, teeDir string) (bool, string, error) {
	if !enabled {
		return false, "", nil
	}

	switch mode {
	case TeeModeNever:
		return false, "", nil
	case TeeModeFailures:
		if exitCode == 0 {
			return false, "", nil
		}
	case TeeModeAlways:
		// Always proceed
	}

	if rawLen < MinTeeSize {
		return false, "", nil
	}

	return true, teeDir, nil
}

// writeTeeFile writes raw output to a tee file in the given directory.
// Returns the file path on success.
func writeTeeFile(raw string, commandSlug string, teeDir string, maxFileSize int, maxFiles int) (string, error) {
	if err := os.MkdirAll(teeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tee directory: %w", err)
	}

	slug := sanitizeSlug(commandSlug)
	epoch := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s.log", epoch, slug)
	filepath := filepath.Join(teeDir, filename)

	// Truncate at maxFileSize (find a safe UTF-8 char boundary)
	content := raw
	if len(raw) > maxFileSize {
		// Find the last valid UTF-8 boundary before maxFileSize
		boundary := maxFileSize
		for boundary > 0 && raw[boundary]&0xC0 == 0x80 {
			boundary--
		}
		if boundary == 0 {
			boundary = maxFileSize
		}
		content = raw[:boundary] + fmt.Sprintf("\n\n--- truncated at %d bytes ---", maxFileSize)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write tee file: %w", err)
	}

	// Rotate old files
	_ = cleanupOldFiles(teeDir, maxFiles)

	return filepath, nil
}

// formatHint formats the hint line with ~ shorthand for home directory.
func formatHint(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		rel, err := filepath.Rel(home, path)
		if err == nil {
			return fmt.Sprintf("[full output: ~/%s]", rel)
		}
	}
	return fmt.Sprintf("[full output: %s]", path)
}

// getTeeModeFromEnv returns the tee mode from environment variable or empty string if not set.
func getTeeModeFromEnv() TeeMode {
	mode := os.Getenv("TOKMAN_TEE_MODE")
	switch TeeMode(mode) {
	case TeeModeAlways, TeeModeNever, TeeModeFailures:
		return TeeMode(mode)
	default:
		return ""
	}
}

// TeeRaw writes raw output to a tee file if conditions are met.
// Returns the file path on success, empty string if skipped/failed.
//
// This is the main entry point for tee functionality. It checks:
// - TOKMAN_TEE=0 env override (disables tee)
// - TOKMAN_TEE_MODE env override (failures/always/never)
// - Config enabled flag
// - Tee mode (failures/always/never)
// - Output size >= MinTeeSize
func TeeRaw(raw string, commandSlug string, exitCode int) string {
	// Check TOKMAN_TEE=0 env override (disable)
	if os.Getenv("TOKMAN_TEE") == "0" {
		return ""
	}

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		cfg = &config.Config{}
	}

	teeDir, err := getTeeDir(cfg)
	if err != nil {
		return ""
	}

	// Get tee config (from config file or defaults)
	teeCfg := DefaultTeeConfig()
	if cfg.Hooks.TeeDir != "" {
		teeCfg.Directory = cfg.Hooks.TeeDir
	}

	// Check for env var override for mode
	if envMode := getTeeModeFromEnv(); envMode != "" {
		teeCfg.Mode = envMode
	}

	// Check if we should tee
	should, _, err := shouldTee(teeCfg.Enabled, teeCfg.Mode, len(raw), exitCode, teeDir)
	if err != nil || !should {
		return ""
	}

	path, err := writeTeeFile(raw, commandSlug, teeDir, teeCfg.MaxFileSize, teeCfg.MaxFiles)
	if err != nil {
		return ""
	}

	return path
}

// TeeAndHint writes raw output to a tee file and returns a formatted hint.
// Returns the hint string if file was written, empty string if skipped.
//
// Example output: "[full output: ~/.local/share/tok/tee/1234567890_cargo_test.log]"
func TeeAndHint(raw string, commandSlug string, exitCode int) string {
	path := TeeRaw(raw, commandSlug, exitCode)
	if path == "" {
		return ""
	}
	return formatHint(path)
}

// ForceTeeHint forces tee output regardless of exit code (used when filters truncate).
// Always writes file if size >= MinTeeSize and tee is enabled.
// Returns hint string if file was written, empty string if skipped/disabled.
//
// Used by filters when FilterResult.truncated = true, ensuring
// the LLM has access to full untruncated output via the hint path.
func ForceTeeHint(raw string, commandSlug string) string {
	// Check TOKMAN_TEE=0 env override (disable)
	if os.Getenv("TOKMAN_TEE") == "0" {
		return ""
	}

	// Skip if output too small
	if len(raw) < MinTeeSize {
		return ""
	}

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		cfg = &config.Config{}
	}

	// Get tee directory
	teeDir, err := getTeeDir(cfg)
	if err != nil {
		return ""
	}

	// Check if tee is enabled
	teeCfg := DefaultTeeConfig()
	if !teeCfg.Enabled {
		return ""
	}

	// Force write (ignore mode, respect enabled and size)
	path, err := writeTeeFile(raw, commandSlug, teeDir, teeCfg.MaxFileSize, teeCfg.MaxFiles)
	if err != nil {
		return ""
	}

	return formatHint(path)
}

// GetTeeDir returns the current tee directory path.
// Useful for displaying to users or for testing.
func GetTeeDir() (string, error) {
	cfg, err := config.Load("")
	if err != nil {
		cfg = &config.Config{}
	}
	return getTeeDir(cfg)
}

// TeeFileInfo represents information about a tee file.
type TeeFileInfo struct {
	Path    string
	Name    string
	ModTime time.Time
	Size    int64
}

// ListTeeFiles returns a list of existing tee files with their paths and modification times.
func ListTeeFiles() ([]TeeFileInfo, error) {
	teeDir, err := GetTeeDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(teeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TeeFileInfo{}, nil
		}
		return nil, err
	}

	var files []TeeFileInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(teeDir, entry.Name())
		files = append(files, TeeFileInfo{
			Path:    path,
			Name:    entry.Name(),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	return files, nil
}

// CleanupTeeFiles removes all tee files older than the specified duration.
func CleanupTeeFiles(maxAge time.Duration) (int, error) {
	files, err := ListTeeFiles()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for _, file := range files {
		if file.ModTime.Before(cutoff) {
			if err := os.Remove(file.Path); err == nil {
				removed++
			}
		}
	}

	return removed, nil
}

// WriteAndHint writes content to a tee file and returns a hint message.
// This is the legacy API that maintains backward compatibility.
// For new code, use TeeAndHint instead.
func WriteAndHint(content string, commandSlug string, exitCode int) string {
	return TeeAndHint(content, commandSlug, exitCode)
}
