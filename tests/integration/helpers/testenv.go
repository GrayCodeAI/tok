// Package helpers provides utilities for integration tests
package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnvironment provides a clean test environment
type TestEnvironment struct {
	RootDir   string
	ConfigDir string
	DataDir   string
	Cleanup   func()
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	rootDir, err := os.MkdirTemp("", "tokman-integration-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	configDir := filepath.Join(rootDir, "config")
	dataDir := filepath.Join(rootDir, "data")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(rootDir)
		t.Fatalf("Failed to create config directory: %v", err)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		os.RemoveAll(rootDir)
		t.Fatalf("Failed to create data directory: %v", err)
	}

	env := &TestEnvironment{
		RootDir:   rootDir,
		ConfigDir: configDir,
		DataDir:   dataDir,
		Cleanup: func() {
			os.RemoveAll(rootDir)
		},
	}

	t.Cleanup(env.Cleanup)
	return env
}

// SetEnv sets environment variables for the test
func (e *TestEnvironment) SetEnv(key, value string) {
	os.Setenv(key, value)
}

// CreateFile creates a test file with content
func (e *TestEnvironment) CreateFile(t *testing.T, path string, content []byte) string {
	t.Helper()

	fullPath := filepath.Join(e.RootDir, path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", fullPath, err)
	}

	return fullPath
}

// CreateDir creates a test directory
func (e *TestEnvironment) CreateDir(t *testing.T, path string) string {
	t.Helper()

	fullPath := filepath.Join(e.RootDir, path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", fullPath, err)
	}

	return fullPath
}
