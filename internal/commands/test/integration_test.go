package test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegrationTestRunnerDetectsCargo tests that test-runner correctly detects Cargo projects
func TestIntegrationTestRunnerDetectsCargo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary Cargo project
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create Cargo.toml
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
edition = "2021"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	// Create src directory
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	// Run detection
	runners := DetectedRunners()

	if len(runners) == 0 {
		t.Fatal("Expected to detect Cargo test runner")
	}

	if runners[0].Name != "Cargo" {
		t.Errorf("Expected first runner to be Cargo, got %s", runners[0].Name)
	}

	if runners[0].Command != "cargo" {
		t.Errorf("Expected command to be 'cargo', got %s", runners[0].Command)
	}
}

// TestIntegrationTestRunnerDetectsNodejs tests Node.js project detection
func TestIntegrationTestRunnerDetectsNodejs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create package.json
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "test": "jest"
  }
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Run detection
	runners := DetectedRunners()

	if len(runners) == 0 {
		t.Fatal("Expected to detect npm test runner")
	}

	foundNpm := false
	for _, r := range runners {
		if r.Name == "npm" {
			foundNpm = true
			break
		}
	}

	if !foundNpm {
		t.Errorf("Expected to find npm in runners, got %v", runners)
	}
}

// TestIntegrationTestRunnerDetectsVitest tests Vitest-specific detection
func TestIntegrationTestRunnerDetectsVitest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create package.json
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0"
}
`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)

	// Create vitest.config.ts
	vitestConfig := `import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {}
})
`
	os.WriteFile(filepath.Join(tmpDir, "vitest.config.ts"), []byte(vitestConfig), 0644)

	// Run detection
	runners := DetectedRunners()

	if len(runners) == 0 {
		t.Fatal("Expected to detect test runners")
	}

	// Vitest should be first due to highest priority
	if runners[0].Name != "Vitest" {
		t.Errorf("Expected Vitest to be first (highest priority), got %s", runners[0].Name)
	}
}

// TestIntegrationTestRunnerMultipleDetections tests detection with multiple config files
func TestIntegrationTestRunnerMultipleDetections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create multiple config files
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(`[package]`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module test`), 0644)

	// Run detection
	runners := DetectedRunners()

	// Should detect all three
	if len(runners) < 3 {
		t.Errorf("Expected at least 3 runners, got %d", len(runners))
	}

	// Check that all are detected
	names := make(map[string]bool)
	for _, r := range runners {
		names[r.Name] = true
	}

	if !names["Cargo"] {
		t.Error("Expected to detect Cargo")
	}
	if !names["Go"] {
		t.Error("Expected to detect Go")
	}
	if !names["npm"] {
		t.Error("Expected to detect npm")
	}
}

// TestIntegrationFilterGenericOutput tests the generic test output filter with real scenarios
func TestIntegrationFilterGenericOutput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldContain []string
		shouldExclude []string
		minLines      int
		maxLines      int
	}{
		{
			name: "Rust test output",
			input: `running 10 tests
test test_one ... ok
test test_two ... ok
test test_three ... FAILED
thread 'test_three' panicked at 'assertion failed', src/lib.rs:42:5
note: run with RUST_BACKTRACE=1 environment variable
failures:
    test_three
test result: FAILED. 2 passed; 1 failed; 0 ignored`,
			shouldContain: []string{"FAILED", "test_three", "panicked", "2 passed"},
			shouldExclude: []string{"test_one ... ok", "test_two ... ok"},
			minLines:      1,
			maxLines:      20,
		},
		{
			name: "Jest output",
			input: `PASS src/utils.test.js
  ✓ should add numbers
  ✓ should subtract numbers
FAIL src/api.test.js
  ✕ should fetch data
    Error: Network request failed
Test Suites: 1 failed, 1 passed, 2 total
Tests:       1 failed, 2 passed, 3 total`,
			shouldContain: []string{"FAIL", "should fetch data", "Network request failed"},
			shouldExclude: []string{},
			minLines:      1,
			maxLines:      15,
		},
		{
			name: "Pytest output",
			input: `============================= test session starts ==============================
platform linux -- Python 3.9.0
collected 5 items

test_sample.py ...F.                                                     [100%]

=================================== FAILURES ===================================
________________________________ test_error ________________________________
    def test_error():
>       assert False
E       AssertionError

test_sample.py:5: AssertionError
========================= 1 failed, 4 passed in 0.03s =========================`,
			shouldContain: []string{"FAILURES", "test_error", "AssertionError"},
			shouldExclude: []string{},
			minLines:      3,
			maxLines:      20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGenericTestOutput(tt.input)

			// Check that result is not empty
			if result == "" {
				t.Error("Expected non-empty output")
			}

			// Check that it contains expected strings
			for _, s := range tt.shouldContain {
				if !containsString(result, s) {
					t.Errorf("Expected output to contain %q, got:\n%s", s, result)
				}
			}

			// Check that it excludes certain strings
			for _, s := range tt.shouldExclude {
				if containsString(result, s) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", s, result)
				}
			}

			// Check line count
			lines := 0
			for _, c := range result {
				if c == '\n' {
					lines++
				}
			}
			if lines < tt.minLines {
				t.Errorf("Expected at least %d lines, got %d", tt.minLines, lines)
			}
			if lines > tt.maxLines {
				t.Errorf("Expected at most %d lines, got %d", tt.maxLines, lines)
			}
		})
	}
}

// TestIntegrationDetectedRunnersWithPriority tests priority ordering
func TestIntegrationDetectedRunnersWithPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create a project with multiple test configurations
	// Vitest (priority 110) should win over Jest (priority 80)
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "vitest.config.ts"), []byte(`export default {}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "jest.config.js"), []byte(`module.exports = {}`), 0644)

	runners := DetectedRunners()

	if len(runners) == 0 {
		t.Fatal("Expected to detect runners")
	}

	// Vitest should be first due to highest priority
	if runners[0].Name != "Vitest" {
		t.Errorf("Expected Vitest first (priority 110), got %s", runners[0].Name)
	}

	// Check that both are in the list
	hasVitest := false
	hasJest := false
	for _, r := range runners {
		if r.Name == "Vitest" {
			hasVitest = true
		}
		if r.Name == "Jest" {
			hasJest = true
		}
	}

	if !hasVitest {
		t.Error("Expected Vitest in runners")
	}
	if !hasJest {
		t.Error("Expected Jest in runners")
	}
}
