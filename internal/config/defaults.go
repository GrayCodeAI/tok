package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigPath returns the path to the configuration file.
// Follows XDG Base Directory Specification on Unix.
// Uses %APPDATA% on Windows.
// Falls back to temp directory if no home directory is available.
func ConfigPath() string {
	// Check for explicit override first
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tok", "config.toml")
	}

	// Windows: use %APPDATA%
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "tok", "config.toml")
		}
	}

	// Unix: default to ~/.config
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".config", "tok", "config.toml")
	}

	// Last resort: use temp directory
	return filepath.Join(os.TempDir(), "tok-config", "config.toml")
}

// DataPath returns the path to the data directory.
// Follows XDG Base Directory Specification on Unix.
// Uses %LOCALAPPDATA% on Windows.
// Falls back to temp directory if no home directory is available.
func DataPath() string {
	// Check for explicit override first
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "tok")
	}

	// Windows: use %LOCALAPPDATA%
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "tok")
		}
		// Fallback to APPDATA if LOCALAPPDATA not set
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "tok", "data")
		}
	}

	// Unix: default to ~/.local/share
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".local", "share", "tok")
	}

	// Last resort: use temp directory with tok subdirectory
	return filepath.Join(os.TempDir(), "tok-data")
}

// DatabasePath returns the path to the SQLite database.
func DatabasePath() string {
	if custom := os.Getenv("TOK_DATABASE_PATH"); custom != "" {
		return custom
	}
	return filepath.Join(DataPath(), "tracking.db")
}

// LogPath returns the path to the log file.
func LogPath() string {
	return filepath.Join(DataPath(), "tok.log")
}

// HooksPath returns the path to the hooks directory.
func HooksPath() string {
	return filepath.Join(DataPath(), "hooks")
}

// ProjectPath returns the canonical path for the current working directory.
// Resolves symlinks for accurate project matching.
func ProjectPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	canonical, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return cwd
	}
	return canonical
}

// ConfigDir returns the path to the tok config directory.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tok")
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "tok")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "tok")
	}
	return filepath.Join(home, ".config", "tok")
}

// FiltersDir returns the path to the user filters directory.
func FiltersDir() string {
	return filepath.Join(ConfigDir(), "filters")
}

// FiltersPath returns the path to the filters directory (alias for FiltersDir).
func FiltersPath() string {
	return FiltersDir()
}
