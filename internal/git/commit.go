package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GenerateCommitMessage analyzes git diff and generates a terse commit message
func GenerateCommitMessage() (string, error) {
	// Get the diff of staged changes
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}

	stat := string(output)
	if strings.TrimSpace(stat) == "" {
		return "", fmt.Errorf("no staged changes found")
	}

	// Get detailed diff for context
	cmd = exec.Command("git", "diff", "--cached")
	diffOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}

	diff := string(diffOutput)

	// Analyze the diff to determine commit type
	msg := analyzeDiff(diff, stat)
	return msg, nil
}

func analyzeDiff(diff, stat string) string {
	// Simple heuristic-based commit message generation
	// In a real implementation, this might use AI or more sophisticated analysis

	// Count changes
	additions := strings.Count(diff, "\n+")
	deletions := strings.Count(diff, "\n-")
	filesChanged := strings.Count(stat, "|")

	// Determine type based on diff content
	msg := ""

	// Check for specific patterns
	if strings.Contains(diff, "TODO") || strings.Contains(diff, "FIXME") {
		msg = "chore: cleanup TODOs"
	} else if strings.Contains(diff, "test") || strings.Contains(diff, "Test") {
		if additions > deletions {
			msg = "test: add tests"
		} else {
			msg = "test: fix tests"
		}
	} else if strings.Contains(diff, "docs/") || strings.Contains(diff, "README") {
		msg = "docs: update documentation"
	} else if strings.Contains(diff, "go.mod") || strings.Contains(diff, "package.json") {
		msg = "chore: update dependencies"
	} else if strings.Contains(diff, "fix") || strings.Contains(diff, "bug") {
		msg = "fix: resolve issue"
	} else if strings.Contains(diff, "refactor") {
		msg = "refactor: simplify code"
	} else if additions > deletions*2 {
		msg = "feat: add functionality"
	} else if deletions > additions*2 {
		msg = "chore: remove unused code"
	} else {
		msg = "chore: update code"
	}

	// Add scope if multiple files
	if filesChanged > 1 {
		// Try to extract common directory
		lines := strings.Split(stat, "\n")
		if len(lines) > 0 {
			parts := strings.Fields(lines[0])
			if len(parts) > 0 {
				file := parts[0]
				if idx := strings.Index(file, "/"); idx > 0 {
					scope := file[:idx]
					msg = insertScope(msg, scope)
				}
			}
		}
	}

	return msg
}

func insertScope(msg, scope string) string {
	// Insert scope into conventional commit format
	// "type: message" -> "type(scope): message"
	if idx := strings.Index(msg, ":"); idx > 0 {
		return msg[:idx] + "(" + scope + ")" + msg[idx:]
	}
	return msg
}

// GetCommitMessage returns a commit message for a specific file or change
func GetCommitMessage(filepath, changeType string) string {
	switch changeType {
	case "add":
		return fmt.Sprintf("feat: add %s", filepath)
	case "modify":
		return fmt.Sprintf("chore: update %s", filepath)
	case "delete":
		return fmt.Sprintf("chore: remove %s", filepath)
	default:
		return fmt.Sprintf("chore: modify %s", filepath)
	}
}
