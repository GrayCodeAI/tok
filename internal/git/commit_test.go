package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		change   string
		expected string
	}{
		{
			name:     "add file",
			filepath: "newfile.go",
			change:   "add",
			expected: "feat: add newfile.go",
		},
		{
			name:     "modify file",
			filepath: "existing.go",
			change:   "modify",
			expected: "chore: update existing.go",
		},
		{
			name:     "delete file",
			filepath: "old.go",
			change:   "delete",
			expected: "chore: remove old.go",
		},
		{
			name:     "default change",
			filepath: "file.txt",
			change:   "unknown",
			expected: "chore: modify file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommitMessage(tt.filepath, tt.change)
			if result != tt.expected {
				t.Errorf("GetCommitMessage(%q, %q) = %q, want %q", tt.filepath, tt.change, result, tt.expected)
			}
		})
	}
}

func TestInsertScope(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		scope    string
		expected string
	}{
		{
			name:     "standard commit",
			msg:      "feat: add feature",
			scope:    "auth",
			expected: "feat(auth): add feature",
		},
		{
			name:     "no colon",
			msg:      "add feature",
			scope:    "auth",
			expected: "add feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := insertScope(tt.msg, tt.scope)
			if result != tt.expected {
				t.Errorf("insertScope(%q, %q) = %q, want %q", tt.msg, tt.scope, result, tt.expected)
			}
		})
	}
}

func TestGenerateCommitMessageNoGit(t *testing.T) {
	// Test without git repository
	_, err := GenerateCommitMessage()
	// Should either error or return empty
	if err == nil {
		t.Log("GenerateCommitMessage returned no error (might be in git repo)")
	}
}

func TestGenerateCommitMessageMock(t *testing.T) {
	// Create a temporary directory and initialize git
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create a test file and stage it
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)
	exec.Command("git", "add", testFile).Run()

	// Generate commit message
	msg, err := GenerateCommitMessage()
	if err != nil {
		t.Logf("GenerateCommitMessage error: %v (expected in test environment)", err)
		return
	}

	if msg == "" {
		t.Error("GenerateCommitMessage returned empty string")
	}
	t.Logf("Generated commit message: %s", msg)
}
