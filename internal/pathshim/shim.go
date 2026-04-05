package pathshim

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// validCommandName matches safe command names (alphanumeric, dash, underscore, dot)
// to prevent directory traversal or shell injection via shim names.
var validCommandName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type PATHShim struct {
	shimDir string
}

func NewPATHShim(shimDir string) (*PATHShim, error) {
	if shimDir == "" {
		shimDir = "/tmp/tokman-shims"
	}
	if err := os.MkdirAll(shimDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create shim directory: %w", err)
	}
	return &PATHShim{shimDir: shimDir}, nil
}

func (s *PATHShim) CreateShim(command string) (string, error) {
	if !validCommandName.MatchString(command) {
		return "", fmt.Errorf("invalid command name %q: must contain only alphanumeric characters, dashes, underscores, and dots", command)
	}
	shimPath := filepath.Join(s.shimDir, command)
	shimContent := fmt.Sprintf("#!/bin/bash\ntokman %s \"$@\"\n", command)
	if err := os.WriteFile(shimPath, []byte(shimContent), 0700); err != nil {
		return "", fmt.Errorf("failed to write shim: %w", err)
	}
	return shimPath, nil
}

func (s *PATHShim) GetPATHEntry() string {
	return s.shimDir
}

// InstallPrep returns the updated PATH string with the shim directory prepended.
// Callers should use this to configure child process environments (exec.Cmd.Env)
// rather than mutating the global process environment with os.Setenv.
func (s *PATHShim) InstallPrep() string {
	currentPATH := os.Getenv("PATH")
	if strings.Contains(currentPATH, s.shimDir) {
		return currentPATH
	}
	return s.shimDir + ":" + currentPATH
}

// Install modifies the global process PATH for backward compatibility.
// Deprecated: Use InstallPrep() to get the updated PATH string and pass
// it to child processes via exec.Cmd.Env instead.
func (s *PATHShim) Install() {
	path := s.InstallPrep()
	os.Setenv("PATH", path)
}

type PipeStripper struct {
	commands []string
}

func NewPipeStripper() *PipeStripper {
	return &PipeStripper{
		commands: []string{"head", "tail", "grep", "sort", "uniq", "wc", "cut", "awk", "sed", "tr", "xargs"},
	}
}

func (p *PipeStripper) Strip(command string) string {
	result := command
	for _, cmd := range p.commands {
		pipe := " | " + cmd
		if idx := strings.Index(result, pipe); idx >= 0 {
			result = strings.TrimSpace(result[:idx])
		}
	}
	return result
}

func (p *PipeStripper) HasPipe(command string) bool {
	for _, cmd := range p.commands {
		if strings.Contains(command, "| "+cmd) {
			return true
		}
	}
	return false
}
