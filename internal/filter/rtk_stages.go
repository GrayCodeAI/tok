package filter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PATHShimInjector creates PATH shims to auto-filter subprocesses.
// Inspired by tokf's PATH shim injection.
type PATHShimInjector struct {
	shimDir  string
	commands map[string]string
}

// NewPATHShimInjector creates a new PATH shim injector.
func NewPATHShimInjector(shimDir string) *PATHShimInjector {
	return &PATHShimInjector{
		shimDir:  shimDir,
		commands: make(map[string]string),
	}
}

// Install installs PATH shims for specified commands.
func (psi *PATHShimInjector) Install(commands []string) error {
	os.MkdirAll(psi.shimDir, 0755)

	for _, cmd := range commands {
		realPath, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}
		psi.commands[cmd] = realPath

		shimPath := filepath.Join(psi.shimDir, cmd)
		shimContent := fmt.Sprintf(`#!/bin/sh
exec tokman %s "$@"
`, cmd)
		if err := os.WriteFile(shimPath, []byte(shimContent), 0755); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall removes PATH shims.
func (psi *PATHShimInjector) Uninstall() error {
	for cmd := range psi.commands {
		shimPath := filepath.Join(psi.shimDir, cmd)
		os.Remove(shimPath)
	}
	return os.RemoveAll(psi.shimDir)
}

// UpdatePATH returns the updated PATH with shim directory prepended.
func (psi *PATHShimInjector) UpdatePATH(currentPath string) string {
	return psi.shimDir + ":" + currentPath
}

// ColorPassthrough strips ANSI codes for matching but restores in output.
// Inspired by tokf's color passthrough.
type ColorPassthrough struct {
	stripped string
	codes    []ANSICode
}

// ANSICode represents an ANSI escape sequence.
type ANSICode struct {
	Position int
	Code     string
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripAndStore strips ANSI codes and stores their positions.
func (cp *ColorPassthrough) StripAndStore(content string) string {
	cp.codes = nil
	cp.stripped = ""
	pos := 0
	result := ansiRe.ReplaceAllStringFunc(content, func(match string) string {
		cp.codes = append(cp.codes, ANSICode{Position: pos, Code: match})
		pos++
		return ""
	})
	cp.stripped = result
	return result
}

// RestoreCodes restores ANSI codes to stripped content.
func (cp *ColorPassthrough) RestoreCodes(stripped string) string {
	if len(cp.codes) == 0 {
		return stripped
	}
	var result bytes.Buffer
	codeIdx := 0
	for i, ch := range stripped {
		if codeIdx < len(cp.codes) && cp.codes[codeIdx].Position == i {
			result.WriteString(cp.codes[codeIdx].Code)
			codeIdx++
		}
		result.WriteRune(ch)
	}
	return result.String()
}

// PreferLessMode compares filtered vs piped output and uses smaller.
// Inspired by tokf's prefer-less mode.
func PreferLessMode(original, filtered string) string {
	if len(filtered) < len(original) {
		return filtered
	}
	return original
}

// TaskRunnerWrapping wraps task runner recipes for individual line filtering.
// Inspired by tokf's task runner wrapping.
type TaskRunnerWrapping struct {
	runner    string
	filterCmd string
}

// NewTaskRunnerWrapping creates a new task runner wrapper.
func NewTaskRunnerWrapping(runner, filterCmd string) *TaskRunnerWrapping {
	return &TaskRunnerWrapping{runner: runner, filterCmd: filterCmd}
}

// Wrap wraps a Makefile or Justfile for tokman filtering.
func (trw *TaskRunnerWrapping) Wrap(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ".") {
			result = append(result, line)
			continue
		}

		// Wrap recipe lines
		if strings.HasPrefix(trimmed, "\t") && trw.runner == "make" {
			result = append(result, "\ttokman proxy "+trimmed[1:])
		} else if !strings.HasPrefix(trimmed, "#") && !strings.Contains(trimmed, ":") && trw.runner == "just" {
			result = append(result, "tokman proxy "+trimmed)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
