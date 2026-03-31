package integration

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestRubyCommands tests Ruby ecosystem command wrappers
func TestRubyCommands(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	tests := []struct {
		name          string
		args          []string
		skipIfMissing string
	}{
		{"rake_help", []string{"rake", "--help"}, "rake"},
		{"rspec_help", []string{"rspec", "--help"}, "rspec"},
		{"rubocop_help", []string{"rubocop", "--help"}, "rubocop"},
		{"bundle_help", []string{"bundle", "--help"}, "bundle"},
		{"rails_help", []string{"rails", "--help"}, "rails"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfMissing != "" {
				if _, err := exec.LookPath(tt.skipIfMissing); err != nil {
					t.Skipf("%s not installed, skipping", tt.skipIfMissing)
				}
			}

			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			// Command may fail if tool not installed, but should not crash
			t.Logf("Output length: %d bytes", len(output))
			if err != nil {
				t.Logf("Command returned error (expected if tool missing): %v", err)
			}
		})
	}
}

// TestInfrastructureCommands tests infrastructure tool wrappers
func TestInfrastructureCommands(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	tests := []struct {
		name          string
		args          []string
		skipIfMissing string
	}{
		{"terraform_help", []string{"terraform", "--help"}, "terraform"},
		{"helm_help", []string{"helm", "--help"}, "helm"},
		{"ansible_help", []string{"ansible", "--help"}, "ansible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfMissing != "" {
				if _, err := exec.LookPath(tt.skipIfMissing); err != nil {
					t.Skipf("%s not installed, skipping", tt.skipIfMissing)
				}
			}

			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			t.Logf("Output length: %d bytes", len(output))
			if err != nil {
				t.Logf("Command returned error (expected if tool missing): %v", err)
			}
		})
	}
}

// TestBuildToolCommands tests build system wrappers
func TestBuildToolCommands(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	tests := []struct {
		name          string
		args          []string
		skipIfMissing string
	}{
		{"gradle_help", []string{"gradle", "--help"}, "gradle"},
		{"mvn_help", []string{"mvn", "--help"}, "mvn"},
		{"make_help", []string{"make", "--help"}, "make"},
		{"mix_help", []string{"mix", "--help"}, "mix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfMissing != "" {
				if _, err := exec.LookPath(tt.skipIfMissing); err != nil {
					t.Skipf("%s not installed, skipping", tt.skipIfMissing)
				}
			}

			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			t.Logf("Output length: %d bytes", len(output))
			if err != nil {
				t.Logf("Command returned error (expected if tool missing): %v", err)
			}
		})
	}
}

// TestSystemUtilityCommands tests new system utilities
func TestSystemUtilityCommands(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	tests := []struct {
		name string
		args []string
	}{
		{"df", []string{"df", "-h"}},
		{"jq_help", []string{"jq", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			t.Logf("Output length: %d bytes", len(output))
			if err != nil {
				t.Logf("Command returned error: %v", err)
			}
		})
	}
}

// TestTOMLFilterIntegration tests the TOML filter system end-to-end
func TestTOMLFilterIntegration(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Test filter commands
	tests := []struct {
		name string
		args []string
	}{
		{"filter_list", []string{"filter", "list"}},
		{"filter_validate", []string{"filter", "validate", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			t.Logf("Output length: %d bytes", len(output))
			t.Logf("Output: %s", truncate(string(output), 500))
			if err != nil {
				t.Logf("Command returned error: %v", err)
			}
		})
	}
}

// TestSessionDiscoveryIntegration tests session discovery commands
func TestSessionDiscoveryIntegration(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Test session and discover commands
	tests := []struct {
		name string
		args []string
	}{
		{"session", []string{"session"}},
		{"discover", []string{"discover"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			t.Logf("Output length: %d bytes", len(output))
			t.Logf("Output: %s", truncate(string(output), 500))
			if err != nil {
				t.Logf("Command returned error: %v", err)
			}
		})
	}
}
