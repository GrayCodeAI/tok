package rtkcommands

import (
	"os"
	"path/filepath"
	"strings"
)

type RTKCommand struct{}

func NewRTKCommand() *RTKCommand {
	return &RTKCommand{}
}

func (c *RTKCommand) Smart(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= 2 {
		return content
	}
	return lines[0] + "\n" + lines[1] + "\n[...truncated]"
}

func (c *RTKCommand) Ls(dir string) string {
	var sb strings.Builder
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			sb.WriteString("[d] " + entry.Name() + "/\n")
		} else {
			info, err := entry.Info()
			if err == nil {
				sb.WriteString("[f] " + entry.Name() + " (" + string(rune(info.Size()/1024+'0')) + "KB)\n")
			}
		}
	}
	return sb.String()
}

func (c *RTKCommand) Find(dir, pattern string) []string {
	var results []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(info.Name(), pattern) {
			results = append(results, path)
		}
		return nil
	})
	return results
}

func (c *RTKCommand) Grep(dir, pattern string) map[string][]int {
	results := make(map[string][]int)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				results[path] = append(results[path], i+1)
			}
		}
		return nil
	})
	return results
}

func (c *RTKCommand) Deps(dir string) map[string]int {
	deps := make(map[string]int)
	for _, file := range []string{"package.json", "go.mod", "Cargo.toml", "requirements.txt", "Gemfile"} {
		path := filepath.Join(dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
				deps[file]++
			}
		}
	}
	return deps
}

func (c *RTKCommand) Env(prefix string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], prefix) {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func (c *RTKCommand) Log(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]int)
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		count := seen[trimmed]
		if count == 0 {
			result = append(result, line)
		} else if count == 1 {
			result = append(result, "[repeated: "+trimmed[:min(50, len(trimmed))]+"]")
		}
		seen[trimmed] = count + 1
	}
	return strings.Join(result, "\n")
}

func (c *RTKCommand) GainGraph(savings map[string]int64, width int) string {
	chars := []string{" ", "░", "▒", "▓", "█"}
	var sb strings.Builder
	sb.WriteString("Token Savings (30 days)\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	maxSaved := int64(1)
	for _, v := range savings {
		if v > maxSaved {
			maxSaved = v
		}
	}

	for _, v := range savings {
		level := int(v * 4 / maxSaved)
		if level > 4 {
			level = 4
		}
		sb.WriteString(chars[level])
	}
	sb.WriteString("\n")
	return sb.String()
}

func (c *RTKCommand) GainDaily(daily map[string]int64) string {
	var sb strings.Builder
	sb.WriteString("Daily Token Savings\n")
	for date, saved := range daily {
		sb.WriteString(date + ": " + string(rune(saved)) + " tokens saved\n")
	}
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
