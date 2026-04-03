// Package luau provides debug functionality for Lua scripts.
package luau

import (
	"fmt"
	"strings"
	"time"
)

// DebugInfo contains debugging information.
type DebugInfo struct {
	Script        string
	Input         string
	Output        string
	ExecutionTime time.Duration
	MemoryUsed    int
	Instructions  int
	Errors        []DebugError
	LineNumbers   map[int]int // line -> execution count
}

// DebugError represents a debug error.
type DebugError struct {
	Line    int
	Column  int
	Message string
}

// Debugger provides debugging capabilities.
type Debugger struct {
	enabled     bool
	breakpoints map[int]bool
	trace       bool
}

// NewDebugger creates a new debugger.
func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[int]bool),
	}
}

// Enable enables debugging.
func (d *Debugger) Enable() {
	d.enabled = true
}

// Disable disables debugging.
func (d *Debugger) Disable() {
	d.enabled = false
}

// SetTrace enables/disables tracing.
func (d *Debugger) SetTrace(enabled bool) {
	d.trace = enabled
}

// SetBreakpoint sets a breakpoint at a line.
func (d *Debugger) SetBreakpoint(line int) {
	d.breakpoints[line] = true
}

// ClearBreakpoint clears a breakpoint.
func (d *Debugger) ClearBreakpoint(line int) {
	delete(d.breakpoints, line)
}

// TraceExecution traces script execution.
func (d *Debugger) TraceExecution(vm *VM, script string, input string) (*DebugInfo, error) {
	start := time.Now()

	info := &DebugInfo{
		Script:      script,
		Input:       input,
		LineNumbers: make(map[int]int),
	}

	// Wrap script with tracing
	tracedScript := d.instrumentScript(script)

	// Execute with tracing
	output, err := vm.ExecuteFilter(tracedScript, input)

	info.ExecutionTime = time.Since(start)
	info.Output = output

	if err != nil {
		info.Errors = append(info.Errors, DebugError{
			Message: err.Error(),
		})
	}

	return info, err
}

// instrumentScript adds instrumentation to a script.
func (d *Debugger) instrumentScript(script string) string {
	lines := strings.Split(script, "\n")
	var instrumented []string

	for i, line := range lines {
		lineNum := i + 1

		// Add trace call at beginning of line
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			instrumented = append(instrumented, fmt.Sprintf("tokman.trace(%d)", lineNum))
		}

		instrumented = append(instrumented, line)
	}

	return strings.Join(instrumented, "\n")
}

// ProfileResult contains profiling results.
type ProfileResult struct {
	TotalTime  time.Duration
	AvgTime    time.Duration
	Calls      int
	MemoryPeak int
	Hotspots   []Hotspot
}

// Hotspot represents a performance hotspot.
type Hotspot struct {
	Line  int
	Code  string
	Time  time.Duration
	Calls int
}

// Profile profiles script execution.
func Profile(vm *VM, script string, iterations int) (*ProfileResult, error) {
	result := &ProfileResult{
		Calls: iterations,
	}

	// Warmup
	vm.ExecuteFilter(script, "warmup")

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := vm.ExecuteFilter(script, "test content")
		if err != nil {
			return nil, err
		}
	}
	result.TotalTime = time.Since(start)
	result.AvgTime = result.TotalTime / time.Duration(iterations)

	return result, nil
}

// ValidateScript checks a script for common issues.
func ValidateScript(script string) []ValidationIssue {
	var issues []ValidationIssue

	lines := strings.Split(script, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Check for infinite loops
		if strings.Contains(trimmed, "while true") {
			issues = append(issues, ValidationIssue{
				Line:     lineNum,
				Severity: "warning",
				Message:  "Potential infinite loop: while true",
			})
		}

		// Check for missing OUTPUT
		if i == len(lines)-1 && !strings.Contains(script, "OUTPUT") {
			issues = append(issues, ValidationIssue{
				Line:     0,
				Severity: "error",
				Message:  "Script does not set OUTPUT variable",
			})
		}

		// Check for dangerous patterns
		dangerous := []string{
			"os.execute",
			"io.popen",
			"loadstring",
			"dofile",
			"loadfile",
		}

		for _, pattern := range dangerous {
			if strings.Contains(trimmed, pattern) {
				issues = append(issues, ValidationIssue{
					Line:     lineNum,
					Severity: "error",
					Message:  fmt.Sprintf("Dangerous function used: %s", pattern),
				})
			}
		}
	}

	return issues
}

// ValidationIssue represents a validation issue.
type ValidationIssue struct {
	Line     int
	Severity string // error, warning, info
	Message  string
}

// FormatDebugOutput formats debug info for display.
func FormatDebugOutput(info *DebugInfo) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Debug Output ===\n")
	fmt.Fprintf(&b, "Execution time: %v\n", info.ExecutionTime)
	fmt.Fprintf(&b, "Input size: %d bytes\n", len(info.Input))
	fmt.Fprintf(&b, "Output size: %d bytes\n", len(info.Output))

	if len(info.Errors) > 0 {
		fmt.Fprintln(&b, "\nErrors:")
		for _, err := range info.Errors {
			fmt.Fprintf(&b, "  Line %d: %s\n", err.Line, err.Message)
		}
	}

	return b.String()
}
