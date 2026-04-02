package pathshim

import (
	"os"
	"strings"
)

type PATHShim struct {
	shimDir string
}

func NewPATHShim(shimDir string) *PATHShim {
	if shimDir == "" {
		shimDir = "/tmp/tokman-shims"
	}
	os.MkdirAll(shimDir, 0755)
	return &PATHShim{shimDir: shimDir}
}

func (s *PATHShim) CreateShim(command string) string {
	shimPath := s.shimDir + "/" + command
	shimContent := "#!/bin/bash\ntokman " + command + " \"$@\"\n"
	os.WriteFile(shimPath, []byte(shimContent), 0755)
	return shimPath
}

func (s *PATHShim) GetPATHEntry() string {
	return s.shimDir
}

func (s *PATHShim) Install() {
	currentPATH := os.Getenv("PATH")
	if !strings.Contains(currentPATH, s.shimDir) {
		os.Setenv("PATH", s.shimDir+":"+currentPATH)
	}
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
