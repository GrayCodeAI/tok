package core

import (
	"os"
	"path/filepath"
)

// GetTokmanSourceDir returns the TokMan source directory for locating built-in filters.
// Returns empty string for installed binaries (filters loaded from embedded filesystem).
func GetTokmanSourceDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	exeDir := filepath.Dir(exe)

	// For installed binaries (go install), embedded filters are in the binary itself.
	// Try common development layout: ./bin/tokman -> project root is two levels up
	for _, dir := range []string{
		filepath.Dir(exeDir),               // parent of exe dir (e.g., project root if exe in ./bin/)
		filepath.Dir(filepath.Dir(exeDir)), // grandparent
		exeDir,                             // same directory as binary
	} {
		builtinDir := filepath.Join(dir, "internal", "toml", "builtin")
		if _, err := os.Stat(builtinDir); err == nil {
			return dir
		}
	}

	// For installed binaries, return empty string to signal that
	// built-in filters should be loaded from the embedded filesystem
	return ""
}
