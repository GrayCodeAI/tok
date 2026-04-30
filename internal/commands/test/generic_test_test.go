package test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectedRunners(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedFirst string
		shouldFindAny bool
	}{
		{
			name:          "Cargo project",
			files:         []string{"Cargo.toml"},
			expectedFirst: "Cargo",
			shouldFindAny: true,
		},
		{
			name:          "Go project",
			files:         []string{"go.mod"},
			expectedFirst: "Go",
			shouldFindAny: true,
		},
		{
			name:          "Vitest project",
			files:         []string{"package.json", "vitest.config.ts"},
			expectedFirst: "Vitest",
			shouldFindAny: true,
		},
		{
			name:          "Jest project",
			files:         []string{"package.json", "jest.config.js"},
			expectedFirst: "Jest",
			shouldFindAny: true,
		},
		{
			name:          "npm project (fallback)",
			files:         []string{"package.json"},
			expectedFirst: "npm",
			shouldFindAny: true,
		},
		{
			name:          "pnpm project",
			files:         []string{"package.json", "pnpm-lock.yaml"},
			expectedFirst: "pnpm",
			shouldFindAny: true,
		},
		{
			name:          "Pytest project",
			files:         []string{"pytest.ini"},
			expectedFirst: "Pytest",
			shouldFindAny: true,
		},
		{
			name:          "RSpec project",
			files:         []string{".rspec"},
			expectedFirst: "RSpec",
			shouldFindAny: true,
		},
		{
			name:          "Rake project",
			files:         []string{"Rakefile"},
			expectedFirst: "Rake Test",
			shouldFindAny: true,
		},
		{
			name:          "Playwright project",
			files:         []string{"playwright.config.ts"},
			expectedFirst: "Playwright",
			shouldFindAny: true,
		},
		{
			name:          "Empty directory",
			files:         []string{},
			expectedFirst: "",
			shouldFindAny: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with test files
			tmpDir := t.TempDir()
			t.Chdir(tmpDir)

			// Create test files
			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				if file == "spec" {
					os.Mkdir(path, 0755)
				} else {
					os.WriteFile(path, []byte("test"), 0644)
				}
			}

			runners := DetectedRunners()

			if tt.shouldFindAny {
				if len(runners) == 0 {
					t.Errorf("Expected to find runners, but found none")
					return
				}
				if runners[0].Name != tt.expectedFirst {
					t.Errorf("Expected first runner to be %q, got %q", tt.expectedFirst, runners[0].Name)
				}
			} else {
				if len(runners) > 0 {
					t.Errorf("Expected no runners, but found %v", runners)
				}
			}
		})
	}
}

func TestDetectedRunnersPriority(t *testing.T) {
	// Test that higher priority runners are preferred
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create both Cargo.toml and package.json (Vitest should win due to specificity)
	os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "vitest.config.ts"), []byte("test"), 0644)

	runners := DetectedRunners()

	if len(runners) == 0 {
		t.Fatal("Expected to find runners")
	}

	// Vitest should be first due to priority
	if runners[0].Name != "Vitest" {
		t.Errorf("Expected Vitest to have highest priority, got %q", runners[0].Name)
	}
}

func TestFilterGenericTestOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name: "Test with failures",
			input: `Running tests...
Test suite failed
FAIL: test_one
Error: assertion failed
PASS: test_two
Tests: 2 passed, 1 failed`,
			contains: []string{"Test Failures", "FAIL: test_one", "Error: assertion failed"},
			excludes: []string{},
		},
		{
			name: "All passing tests",
			input: `Running tests...
PASS: test_one
PASS: test_two
Tests: 2 passed, 0 failed`,
			contains: []string{"2 passed"},
			excludes: []string{"Test Failures"},
		},
		{
			name:     "Empty output",
			input:    "",
			contains: []string{},
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGenericTestOutput(tt.input)

			for _, s := range tt.contains {
				if !containsSubstring(result, s) {
					t.Errorf("Expected output to contain %q, got:\n%s", s, result)
				}
			}

			for _, s := range tt.excludes {
				if containsSubstring(result, s) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(substr) == 0 || len(s) == 0 || (len(substr) <= len(s) && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRunnerStruct(t *testing.T) {
	runner := TestRunner{
		Name:        "Test Runner",
		Command:     "test",
		Args:        []string{"--verbose"},
		DetectFiles: []string{"test.config"},
		Priority:    100,
	}

	if runner.Name != "Test Runner" {
		t.Errorf("Expected name to be 'Test Runner', got %q", runner.Name)
	}

	if runner.Command != "test" {
		t.Errorf("Expected command to be 'test', got %q", runner.Command)
	}

	if len(runner.Args) != 1 || runner.Args[0] != "--verbose" {
		t.Errorf("Expected args to be ['--verbose'], got %v", runner.Args)
	}

	if runner.Priority != 100 {
		t.Errorf("Expected priority to be 100, got %d", runner.Priority)
	}
}
