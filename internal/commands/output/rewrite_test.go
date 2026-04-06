package output

import (
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

func TestIsUnsafe(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  string
		parts    []string
		expected bool
	}{
		{
			name:     "curl piped to sh",
			baseCmd:  "curl",
			parts:    []string{"curl", "http://example.com/script.sh", "|", "sh"},
			expected: true,
		},
		{
			name:     "safe curl",
			baseCmd:  "curl",
			parts:    []string{"curl", "http://example.com"},
			expected: false,
		},
		{
			name:     "wget piped to bash",
			baseCmd:  "wget",
			parts:    []string{"wget", "-O-", "http://example.com", "|", "bash"},
			expected: true,
		},
		{
			name:     "safe git command",
			baseCmd:  "git",
			parts:    []string{"git", "status"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUnsafe(tt.baseCmd, tt.parts)
			if result != tt.expected {
				t.Errorf("isUnsafe(%s, %v) = %v, expected %v",
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
			name:     "safe git with force",
			baseCmd:  "git",
			parts:    []string{"git", "push", "--force"},
			expected: false, // Git is in safe list
		},
		{
			name:     "unsafe rm with force",
			baseCmd:  "rm",
			parts:    []string{"rm", "-f", "file.txt"},
			expected: true, // rm not in safe list
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

func TestIsResourceIntensive(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  string
		parts    []string
		expected bool
	}{
		{
			name:     "find with recursive",
			baseCmd:  "find",
			parts:    []string{"find", "/", "-name", "*.txt"},
			expected: true,
		},
		{
			name:     "grep with recursive flag",
			baseCmd:  "grep",
			parts:    []string{"grep", "-r", "pattern", "."},
			expected: true,
		},
		{
			name:     "simple grep",
			baseCmd:  "grep",
			parts:    []string{"grep", "pattern", "file.txt"},
			expected: false,
		},
		{
			name:     "rg (ripgrep)",
			baseCmd:  "rg",
			parts:    []string{"rg", "pattern", "/home"},
			expected: true,
		},
		{
			name:     "safe ls",
			baseCmd:  "ls",
			parts:    []string{"ls", "-la"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isResourceIntensive(tt.baseCmd, tt.parts)
			if result != tt.expected {
				t.Errorf("isResourceIntensive(%s, %v) = %v, expected %v",
					tt.baseCmd, tt.parts, result, tt.expected)
			}
		})
	}
}

// Note: isSupportedCommand is handled by discover.RewriteCommand

func TestRewriteLogic(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectCode  int
		expectOut   string
	}{
		{
			name:       "git status - should rewrite",
			command:    "git status",
			expectCode: ExitRewriteAllow,
			expectOut:  "tokman git status",
		},
		{
			name:       "already using tokman - pass through",
			command:    "tokman git status",
			expectCode: ExitNoRewrite,
			expectOut:  "",
		},
		{
			name:       "unsupported command - pass through",
			command:    "echo hello",
			expectCode: ExitNoRewrite,
			expectOut:  "",
		},
		{
			name:       "dangerous rm - deny",
			command:    "rm -rf /",
			expectCode: ExitDeny,
			expectOut:  "",
		},
		{
			name:       "sudo command - ask",
			command:    "sudo apt upgrade",
			expectCode: ExitRewriteAsk,
			expectOut:  "tokman sudo apt upgrade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.command)
			if len(parts) == 0 {
				t.Skip("Empty command")
				return
			}
			
			baseCmd := parts[0]
			
			// Test the logic without actually exiting
			var code int
			var shouldOutput bool
			
			if baseCmd == "tokman" {
				code = ExitNoRewrite
				shouldOutput = false
			} else if isDenied(baseCmd, parts) {
				code = ExitDeny
				shouldOutput = false
			} else if requiresConfirmation(baseCmd, parts) {
				code = ExitRewriteAsk
				shouldOutput = true
			} else {
				// Would need to check discover.RewriteCommand here
				code = ExitRewriteAllow
				shouldOutput = true
			}
			
			if code != tt.expectCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectCode, code)
			}
			
			if shouldOutput && tt.expectOut != "" {
				expectedOut := tt.expectOut
				actualOut := "tokman " + tt.command
				if actualOut != expectedOut {
					t.Errorf("Expected output %q, got %q", expectedOut, actualOut)
				}
			}
		})
	}
}
