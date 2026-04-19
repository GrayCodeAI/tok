package output

import (
	"os/exec"
	"strings"
	"testing"
)

func TestIsDenied(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  string
		parts    []string
		expected bool
	}{
		{
			name:     "rm command",
			baseCmd:  "rm",
			parts:    []string{"rm", "-rf", "/tmp/test"},
			expected: true,
		},
		{
			name:     "dd command",
			baseCmd:  "dd",
			parts:    []string{"dd", "if=/dev/zero", "of=/dev/sda"},
			expected: true,
		},
		{
			name:     "safe git command",
			baseCmd:  "git",
			parts:    []string{"git", "status"},
			expected: false,
		},
		{
			name:     "rm -rf / pattern",
			baseCmd:  "rm",
			parts:    []string{"rm", "-rf", "/"},
			expected: true,
		},
		{
			name:     "safe ls command",
			baseCmd:  "ls",
			parts:    []string{"ls", "-la"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDenied(tt.baseCmd, tt.parts)
			if result != tt.expected {
				t.Errorf("isDenied(%s, %v) = %v, expected %v",
					tt.baseCmd, tt.parts, result, tt.expected)
			}
		})
	}
}

func TestRequiresConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  string
		parts    []string
		expected bool
	}{
		{
			name:     "sudo command",
			baseCmd:  "sudo",
			parts:    []string{"sudo", "apt", "upgrade"},
			expected: true,
		},
		{
			name:     "systemctl command",
			baseCmd:  "systemctl",
			parts:    []string{"systemctl", "restart", "nginx"},
			expected: true,
		},
		{
			name:     "su command",
			baseCmd:  "su",
			parts:    []string{"su", "-", "root"},
			expected: true,
		},
		{
			name:     "safe npm command",
			baseCmd:  "npm",
			parts:    []string{"npm", "install"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := requiresConfirmation(tt.baseCmd, tt.parts)
			if result != tt.expected {
				t.Errorf("requiresConfirmation(%s, %v) = %v, expected %v",
					tt.baseCmd, tt.parts, result, tt.expected)
			}
		})
	}
}

// TestRewriteCommandE2E tests the actual tok rewrite command via subprocess
func TestRewriteCommandE2E(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		expectCode int
		expectOut  string
	}{
		{
			name:       "git status rewrites",
			command:    "git status",
			expectCode: 0, // ExitRewriteAllow - will rewrite to tok git status
			expectOut:  "tok git status",
		},
		{
			name:       "dangerous rm denied",
			command:    "rm -rf /",
			expectCode: 2, // ExitDeny
			expectOut:  "",
		},
		{
			name:       "sudo asks",
			command:    "sudo apt upgrade",
			expectCode: 3, // ExitRewriteAsk
			expectOut:  "tok sudo apt upgrade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the binary first
			buildCmd := exec.Command("go", "build", "-o", "/tmp/tok-test", "../../cmd/tok")
			buildCmd.Dir = "../.."
			if err := buildCmd.Run(); err != nil {
				t.Skipf("Could not build tok: %v", err)
				return
			}

			// Run tok rewrite
			cmd := exec.Command("/tmp/tok-test", "rewrite", tt.command)
			out, err := cmd.Output()

			// Get actual exit code
			actualCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					actualCode = exitErr.ExitCode()
				}
			}

			actualOut := strings.TrimSpace(string(out))

			if actualCode != tt.expectCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectCode, actualCode)
			}

			if tt.expectOut != "" && actualOut != tt.expectOut {
				t.Errorf("Expected output %q, got %q", tt.expectOut, actualOut)
			}
		})
	}
}
