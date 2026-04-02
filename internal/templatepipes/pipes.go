package templatepipes

import (
	"strings"
)

type PipeStage func(input string, args []string) string

type PipeEngine struct {
	stages map[string]PipeStage
}

func NewPipeEngine() *PipeEngine {
	engine := &PipeEngine{
		stages: make(map[string]PipeStage),
	}
	engine.registerBuiltInStages()
	return engine
}

func (e *PipeEngine) registerBuiltInStages() {
	e.stages["trim"] = func(input string, args []string) string {
		return strings.TrimSpace(input)
	}

	e.stages["upper"] = func(input string, args []string) string {
		return strings.ToUpper(input)
	}

	e.stages["lower"] = func(input string, args []string) string {
		return strings.ToLower(input)
	}

	e.stages["replace"] = func(input string, args []string) string {
		if len(args) >= 2 {
			return strings.ReplaceAll(input, args[0], args[1])
		}
		return input
	}

	e.stages["split"] = func(input string, args []string) string {
		delimiter := "\n"
		if len(args) > 0 {
			delimiter = args[0]
		}
		parts := strings.Split(input, delimiter)
		if len(parts) > 10 {
			return strings.Join(parts[:10], delimiter)
		}
		return input
	}

	e.stages["head"] = func(input string, args []string) string {
		n := 10
		if len(args) > 0 {
			for _, c := range args[0] {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				}
			}
		}
		lines := strings.Split(input, "\n")
		if len(lines) > n {
			return strings.Join(lines[:n], "\n")
		}
		return input
	}

	e.stages["tail"] = func(input string, args []string) string {
		n := 10
		if len(args) > 0 {
			for _, c := range args[0] {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				}
			}
		}
		lines := strings.Split(input, "\n")
		if len(lines) > n {
			return strings.Join(lines[len(lines)-n:], "\n")
		}
		return input
	}

	e.stages["dedup"] = func(input string, args []string) string {
		lines := strings.Split(input, "\n")
		seen := make(map[string]bool)
		var result []string
		for _, line := range lines {
			if !seen[line] {
				seen[line] = true
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}

	e.stages["collapse"] = func(input string, args []string) string {
		result := input
		for strings.Contains(result, "\n\n\n") {
			result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
		}
		for strings.Contains(result, "  ") {
			result = strings.ReplaceAll(result, "  ", " ")
		}
		return result
	}

	e.stages["truncate"] = func(input string, args []string) string {
		maxLen := 1000
		if len(args) > 0 {
			for _, c := range args[0] {
				if c >= '0' && c <= '9' {
					maxLen = maxLen*10 + int(c-'0')
				}
			}
		}
		if len(input) > maxLen {
			return input[:maxLen] + "..."
		}
		return input
	}

	e.stages["strip_ansi"] = func(input string, args []string) string {
		result := strings.Builder{}
		inEscape := false
		for _, c := range input {
			if c == '\033' {
				inEscape = true
				continue
			}
			if inEscape {
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '~' || c == ';' {
					inEscape = false
				}
				continue
			}
			result.WriteRune(c)
		}
		return result.String()
	}

	e.stages["strip_comments"] = func(input string, args []string) string {
		lines := strings.Split(input, "\n")
		var filtered []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
				continue
			}
			filtered = append(filtered, line)
		}
		return strings.Join(filtered, "\n")
	}

	e.stages["count"] = func(input string, args []string) string {
		lines := strings.Split(strings.TrimSpace(input), "\n")
		return string(rune(len(lines))) + " lines"
	}

	e.stages["sort"] = func(input string, args []string) string {
		lines := strings.Split(input, "\n")
		for i := 0; i < len(lines); i++ {
			for j := i + 1; j < len(lines); j++ {
				if lines[j] < lines[i] {
					lines[i], lines[j] = lines[j], lines[i]
				}
			}
		}
		return strings.Join(lines, "\n")
	}

	e.stages["reverse"] = func(input string, args []string) string {
		lines := strings.Split(input, "\n")
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
		return strings.Join(lines, "\n")
	}
}

func (e *PipeEngine) Execute(input, pipeline string) string {
	stages := strings.Split(pipeline, "|")
	result := input

	for _, stage := range stages {
		stage = strings.TrimSpace(stage)
		parts := strings.SplitN(stage, " ", 2)
		name := parts[0]
		var args []string
		if len(parts) > 1 {
			args = strings.Fields(parts[1])
		}

		if fn, ok := e.stages[name]; ok {
			result = fn(result, args)
		}
	}

	return result
}

func (e *PipeEngine) RegisterStage(name string, fn PipeStage) {
	e.stages[name] = fn
}

func (e *PipeEngine) ListStages() []string {
	var names []string
	for name := range e.stages {
		names = append(names, name)
	}
	return names
}
