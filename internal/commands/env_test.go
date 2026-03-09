package commands

import (
	"testing"
)

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"API key", "API_KEY", true},
		{"secret", "SECRET_TOKEN", true},
		{"password", "PASSWORD", true},
		{"auth", "AUTHORIZATION", true},
		{"credential", "AWS_CREDENTIALS", true},
		{"private", "PRIVATE_KEY", true},
		{"access key", "ACCESS_KEY", true},
		{"apikey", "APIKEY", true},
		{"jwt", "JWT_SECRET", true},
		{"pass", "DB_PASS", true},
		{"normal var", "PATH", false},
		{"normal var 2", "HOME", false},
		{"normal var 3", "EDITOR", false},
		{"case insensitive", "api_key", true},
		{"case insensitive 2", "Secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitive(tt.varName)
			if result != tt.expected {
				t.Errorf("isSensitive(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"short value", "abc", "****"},
		{"4 chars", "abcd", "****"},
		{"normal value", "sk-1234567890abcdef", "sk****ef"},
		{"long value", "sk-proj-abcdefghijklmnopqrstuvwxyz1234567890", "sk****90"},
		{"empty", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSensitive(tt.value)
			if result != tt.expected {
				t.Errorf("maskSensitive(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsLangVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"RUST", "RUST_BACKTRACE", true},
		{"CARGO", "CARGO_HOME", true},
		{"PYTHON", "PYTHONPATH", true},
		{"PIP", "PIP_INDEX_URL", true},
		{"NODE", "NODE_ENV", true},
		{"NPM", "NPM_CONFIG_REGISTRY", true},
		{"GO", "GOPATH", true},
		{"GOROOT", "GOROOT", true},
		{"JAVA", "JAVA_HOME", true},
		{"normal var", "PATH", false},
		{"normal var 2", "HOME", false},
		{"case insensitive", "python_version", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLangVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isLangVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestIsCloudVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"AWS", "AWS_ACCESS_KEY_ID", true},
		{"AZURE", "AZURE_SUBSCRIPTION_ID", true},
		{"GCP", "GCP_PROJECT_ID", true},
		{"DOCKER", "DOCKER_HOST", true},
		{"KUBERNETES", "KUBERNETES_SERVICE_HOST", true},
		{"TERRAFORM", "TERRAFORM_VERSION", true},
		{"normal var", "PATH", false},
		{"case insensitive", "aws_region", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCloudVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isCloudVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestIsToolVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"EDITOR", "EDITOR", true},
		{"SHELL", "SHELL", true},
		{"TERM", "TERM", true},
		{"GIT", "GIT_EDITOR", true},
		{"SSH", "SSH_AUTH_SOCK", true},
		{"HOMEBREW", "HOMEBREW_PREFIX", true},
		{"CLAUDE", "CLAUDE_API_KEY", true},
		{"normal var", "PATH", false},
		{"case insensitive", "editor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isToolVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isToolVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestIsInterestingVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"HOME", "HOME", true},
		{"USER", "USER", true},
		{"LANG", "LANG", true},
		{"LC_ALL", "LC_ALL", true},
		{"TZ", "TZ", true},
		{"PWD", "PWD", true},
		{"OLDPWD", "OLDPWD", true},
		{"normal var", "PATH", false},
		{"normal var 2", "EDITOR", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInterestingVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isInterestingVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}
