package discover

import (
	"testing"
)

func TestRewrite(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Git commands
		{
			name:     "git status",
			input:    "git status",
			expected: "tokman git status",
		},
		{
			name:     "git status with args",
			input:    "git status --short",
			expected: "tokman git status --short",
		},
		{
			name:     "git diff",
			input:    "git diff",
			expected: "tokman git diff",
		},
		{
			name:     "git log",
			input:    "git log",
			expected: "tokman git log",
		},
		{
			name:     "git log with args",
			input:    "git log -10",
			expected: "tokman git log -10",
		},
		// LS commands
		{
			name:     "ls",
			input:    "ls",
			expected: "tokman ls",
		},
		{
			name:     "ls -la",
			input:    "ls -la",
			expected: "tokman ls -la",
		},
		{
			name:     "ls with path",
			input:    "ls /home/user",
			expected: "tokman ls /home/user",
		},
		// Go commands
		{
			name:     "go test",
			input:    "go test",
			expected: "tokman go test",
		},
		{
			name:     "go test with args",
			input:    "go test ./...",
			expected: "tokman go test ./...",
		},
		{
			name:     "go build",
			input:    "go build",
			expected: "tokman go build",
		},
		// File commands (now supported!)
		{
			name:     "cat file",
			input:    "cat file.txt",
			expected: "tokman read file.txt",
		},
		{
			name:     "rg pattern",
			input:    "rg \"fn main\"",
			expected: "tokman grep \"fn main\"",
		},
		{
			name:     "find command",
			input:    "find . -name foo",
			expected: "tokman find . -name foo",
		},
		// Partial/unsupported
		{
			name:     "partial match not rewritten",
			input:    "git clone",
			expected: "git clone",
		},
		{
			name:     "cd ignored",
			input:    "cd /tmp",
			expected: "cd /tmp",
		},
		{
			name:     "echo ignored",
			input:    "echo hello",
			expected: "echo hello",
		},
		// Compound commands
		{
			name:     "compound and",
			input:    "git add . && cargo test",
			expected: "tokman git add . && tokman cargo test",
		},
		{
			name:     "compound pipe",
			input:    "git log -10 | grep feat",
			expected: "tokman git log -10 | grep feat",
		},
		// Docker
		{
			name:     "docker ps",
			input:    "docker ps",
			expected: "tokman docker ps",
		},
		// GitHub CLI
		{
			name:     "gh pr list",
			input:    "gh pr list",
			expected: "tokman gh pr list",
		},
		// head -N special case (TokMan parity)
		{
			name:     "head -N file",
			input:    "head -20 src/main.rs",
			expected: "tokman read src/main.rs --max-lines 20",
		},
		{
			name:     "head --lines=N file",
			input:    "head --lines=50 src/lib.rs",
			expected: "tokman read src/lib.rs --max-lines 50",
		},
		{
			name:     "head file (no numeric flag)",
			input:    "head src/main.rs",
			expected: "tokman read src/main.rs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rewrite(tt.input)
			if result != tt.expected {
				t.Errorf("Rewrite() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestShouldRewrite(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "git status",
			input:    "git status",
			expected: true,
		},
		{
			name:     "git status with args",
			input:    "git status --short",
			expected: true,
		},
		{
			name:     "cat file (now supported)",
			input:    "cat file.txt",
			expected: true,
		},
		{
			name:     "partial git command",
			input:    "git clone",
			expected: false,
		},
		{
			name:     "ls",
			input:    "ls",
			expected: true,
		},
		{
			name:     "go test",
			input:    "go test",
			expected: true,
		},
		{
			name:     "cd ignored",
			input:    "cd /tmp",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRewrite(tt.input)
			if result != tt.expected {
				t.Errorf("ShouldRewrite() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetMapping(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFound bool
		wantCmd   string
	}{
		{
			name:      "git status",
			input:     "git status",
			wantFound: true,
			wantCmd:   "tokman git",
		},
		{
			name:      "cd ignored",
			input:     "cd /tmp",
			wantFound: false,
		},
		{
			name:      "cat file (now supported)",
			input:     "cat file.txt",
			wantFound: true,
			wantCmd:   "tokman read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, found := GetMapping(tt.input)
			if found != tt.wantFound {
				t.Errorf("GetMapping() found = %v, want %v", found, tt.wantFound)
			}
			if found && mapping.TokManCmd != tt.wantCmd {
				t.Errorf("GetMapping() TokManCmd = %q, want %q", mapping.TokManCmd, tt.wantCmd)
			}
		})
	}
}

func TestListRewrites(t *testing.T) {
	rewrites := ListRewrites()

	if len(rewrites) == 0 {
		t.Error("ListRewrites() returned empty list")
	}

	// Check all returned rewrites are enabled
	for _, r := range rewrites {
		if !r.Enabled {
			t.Errorf("ListRewrites() returned disabled mapping: %s", r.Original)
		}
	}

	// Check expected commands are present (using first prefix from each rule)
	expected := []string{"git", "gh", "cargo", "ls", "go"}
	for _, exp := range expected {
		found := false
		for _, r := range rewrites {
			if r.Original == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListRewrites() missing expected command: %s", exp)
		}
	}
}

func TestClassifyCommand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSupport  bool
		wantCategory string
	}{
		{
			name:         "git status",
			input:        "git status",
			wantSupport:  true,
			wantCategory: "Git",
		},
		{
			name:         "cargo test",
			input:        "cargo test",
			wantSupport:  true,
			wantCategory: "Cargo",
		},
		{
			name:         "docker ps",
			input:        "docker ps",
			wantSupport:  true,
			wantCategory: "Infra",
		},
		{
			name:        "cd ignored",
			input:       "cd /tmp",
			wantSupport: false,
		},
		{
			name:        "terraform unsupported",
			input:       "terraform plan",
			wantSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if class.Supported != tt.wantSupport {
				t.Errorf("ClassifyCommand() Supported = %v, want %v", class.Supported, tt.wantSupport)
			}
			if tt.wantSupport && class.Category != tt.wantCategory {
				t.Errorf("ClassifyCommand() Category = %q, want %q", class.Category, tt.wantCategory)
			}
		})
	}
}

func TestRewriteCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantChanged bool
		wantResult  string
	}{
		{
			name:        "git status",
			input:       "git status",
			wantChanged: true,
			wantResult:  "tokman git status",
		},
		{
			name:        "no rewrite for cd",
			input:       "cd /tmp",
			wantChanged: false,
			wantResult:  "cd /tmp",
		},
		{
			name:        "compound and",
			input:       "git add . && cargo test",
			wantChanged: true,
			wantResult:  "tokman git add . && tokman cargo test",
		},
		{
			name:        "already tokman",
			input:       "tokman git status",
			wantChanged: false,
			wantResult:  "tokman git status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := RewriteCommand(tt.input, nil)
			if changed != tt.wantChanged {
				t.Errorf("RewriteCommand() changed = %v, want %v", changed, tt.wantChanged)
			}
			if result != tt.wantResult {
				t.Errorf("RewriteCommand() = %q, want %q", result, tt.wantResult)
			}
		})
	}
}
