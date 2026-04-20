package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	flagFileName = ".tok-active"
)

// GetFlagPath returns the path to the tok flag file
func GetFlagPath() string {
	configDir := os.Getenv("TOK_CONFIG_DIR")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), flagFileName)
		}
		configDir = filepath.Join(home, ".config", "tok")
	}
	return filepath.Join(configDir, flagFileName)
}

// IsActive checks if tok mode is currently active
func IsActive() bool {
	flagPath := GetFlagPath()
	_, err := os.Stat(flagPath)
	return err == nil
}

// GetMode returns the current tok mode (lite, full, ultra, etc.)
func GetMode() string {
	flagPath := GetFlagPath()
	data, err := os.ReadFile(flagPath)
	if err != nil {
		return ""
	}
	mode := strings.TrimSpace(string(data))
	if mode == "" {
		return "full"
	}
	return mode
}

// Activate enables tok mode
func Activate(mode string) error {
	if mode == "" {
		mode = "full"
	}

	flagPath := GetFlagPath()
	flagDir := filepath.Dir(flagPath)

	if err := os.MkdirAll(flagDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := os.WriteFile(flagPath, []byte(mode), 0644); err != nil {
		return fmt.Errorf("failed to write flag file: %w", err)
	}

	return nil
}

// Deactivate disables tok mode
func Deactivate() error {
	flagPath := GetFlagPath()
	if err := os.Remove(flagPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove flag file: %w", err)
	}
	return nil
}

// GetStatusLine returns the statusline badge text
func GetStatusLine() string {
	if !IsActive() {
		return ""
	}

	mode := GetMode()
	if mode == "full" || mode == "" {
		return "[TOK]"
	}
	return fmt.Sprintf("[TOK:%s]", strings.ToUpper(mode))
}

// AutoActivateOnStartup checks config and auto-activates if configured
func AutoActivateOnStartup() error {
	// Check env var
	if os.Getenv("TOK_AUTO_ACTIVATE") == "1" {
		mode := os.Getenv("TOK_DEFAULT_MODE")
		if mode == "" {
			mode = "full"
		}
		return Activate(mode)
	}

	// Check config file
	configPath := filepath.Join(filepath.Dir(GetFlagPath()), "config.json")
	// Simple config check - could be expanded
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, could parse for auto_activate setting
		// For now, just check env var
	}

	return nil
}

// ResolveDefaultMode picks an effective mode for commands when user did not pass -mode.
func ResolveDefaultMode() string {
	if IsActive() {
		if mode := GetMode(); mode != "" {
			return mode
		}
	}
	if mode := os.Getenv("TOK_DEFAULT_MODE"); mode != "" {
		return mode
	}
	return "full"
}
