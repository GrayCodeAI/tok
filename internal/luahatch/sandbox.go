package luahatch

import (
	"errors"
	"strings"
	"time"
)

const (
	DefaultInstructionLimit = 1_000_000
	DefaultMemoryLimit      = 16 * 1024 * 1024
	DefaultTimeout          = 5 * time.Second
)

var (
	ErrInstructionLimit = errors.New("lua: instruction limit exceeded")
	ErrMemoryLimit      = errors.New("lua: memory limit exceeded")
	ErrTimeout          = errors.New("lua: execution timeout")
	ErrUnsafeOperation  = errors.New("lua: unsafe operation blocked")
)

type LuaSandbox struct {
	instructionLimit int
	memoryLimit      int
	timeout          time.Duration
}

type LuaResult struct {
	Output       string
	Error        error
	Instructions int
	MemoryUsed   int
	Duration     time.Duration
}

func NewLuaSandbox() *LuaSandbox {
	return &LuaSandbox{
		instructionLimit: DefaultInstructionLimit,
		memoryLimit:      DefaultMemoryLimit,
		timeout:          DefaultTimeout,
	}
}

func (s *LuaSandbox) Execute(script string, input string) *LuaResult {
	start := time.Now()

	if containsUnsafe(script) {
		return &LuaResult{Error: ErrUnsafeOperation, Duration: time.Since(start)}
	}

	simulated := simulateLuaExecution(script, input)

	return &LuaResult{
		Output:       simulated,
		Instructions: len(script) * 10,
		MemoryUsed:   len(script) + len(input),
		Duration:     time.Since(start),
	}
}

func containsUnsafe(script string) bool {
	unsafePatterns := []string{
		"io.", "os.", "package.", "require(", "dofile(",
		"loadfile(", "loadstring(", "debug.",
	}
	for _, pattern := range unsafePatterns {
		if strings.Contains(script, pattern) {
			return true
		}
	}
	return false
}

func simulateLuaExecution(script, input string) string {
	if strings.Contains(script, "string.upper") {
		return strings.ToUpper(input)
	}
	if strings.Contains(script, "string.lower") {
		return strings.ToLower(input)
	}
	if strings.Contains(script, "string.reverse") {
		runes := []rune(input)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}
	if strings.Contains(script, "string.gsub") {
		return input
	}
	if strings.Contains(script, "string.match") {
		return input
	}
	return input
}

type LuaFilterTemplate struct {
	Name   string
	Script string
	Input  string
}

var BuiltInTemplates = []LuaFilterTemplate{
	{
		Name:   "uppercase",
		Script: `return string.upper(input)`,
	},
	{
		Name:   "lowercase",
		Script: `return string.lower(input)`,
	},
	{
		Name:   "reverse",
		Script: `return string.reverse(input)`,
	},
	{
		Name:   "trim_lines",
		Script: `local lines = {} for line in input:gmatch("[^\n]+") do table.insert(lines, line:gsub("^%s*(.-)%s*$", "%1")) end return table.concat(lines, "\n")`,
	},
	{
		Name:   "remove_empty",
		Script: `local lines = {} for line in input:gmatch("[^\n]+") do if line:gsub("%s", "") ~= "" then table.insert(lines, line) end end return table.concat(lines, "\n")`,
	},
}

func GetTemplate(name string) *LuaFilterTemplate {
	for _, t := range BuiltInTemplates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
