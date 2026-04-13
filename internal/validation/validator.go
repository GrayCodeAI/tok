package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	MaxInputSize    = 10 * 1024 * 1024 // 10MB
	MaxCommandArgs  = 1000
	MaxPathLength   = 4096
	MaxConfigSize   = 1 * 1024 * 1024 // 1MB
)

// ValidateInputSize checks input size limits
func ValidateInputSize(input string) error {
	if len(input) > MaxInputSize {
		return fmt.Errorf("input exceeds maximum size of %d bytes", MaxInputSize)
	}
	return nil
}

// ValidateCommandArgs checks command argument limits
func ValidateCommandArgs(args []string) error {
	if len(args) > MaxCommandArgs {
		return fmt.Errorf("too many arguments: %d (max %d)", len(args), MaxCommandArgs)
	}
	for i, arg := range args {
		if len(arg) > MaxPathLength {
			return fmt.Errorf("argument %d exceeds max length", i)
		}
	}
	return nil
}

// SanitizePath prevents path traversal attacks
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	
	// Clean and resolve path
	cleaned := filepath.Clean(path)
	
	// Check for path traversal
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}
	
	// Convert to absolute path
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	
	return abs, nil
}

// ValidateConfigPath ensures config paths are safe
func ValidateConfigPath(path string) error {
	sanitized, err := SanitizePath(path)
	if err != nil {
		return err
	}
	
	// Ensure path is within allowed directories
	homeDir := filepath.Clean(filepath.Join(sanitized, "..", ".."))
	if !strings.HasPrefix(sanitized, homeDir) {
		return fmt.Errorf("config path outside allowed directory")
	}
	
	return nil
}
