// Package output provides a centralized output abstraction for tok CLI.
// It handles JSON mode, quiet mode, TTY detection, and testable io.Writer injection.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

// Printer is the centralized output abstraction for tok CLI.
type Printer struct {
	mu       sync.Mutex
	stdout   io.Writer
	stderr   io.Writer
	quiet    bool
	jsonMode bool
	isTTY    bool
}

// New creates a new Printer writing to os.Stdout and os.Stderr.
func New() *Printer {
	return &Printer{
		stdout: os.Stdout,
		stderr: os.Stderr,
		isTTY:  term.IsTerminal(int(os.Stdout.Fd())),
	}
}

// NewTest creates a Printer for testing with controlled writers.
func NewTest(stdout, stderr io.Writer) *Printer {
	return &Printer{
		stdout: stdout,
		stderr: stderr,
		isTTY:  false,
	}
}

// SetQuiet enables/disables non-essential output.
func (p *Printer) SetQuiet(quiet bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.quiet = quiet
}

// SetJSON enables/disables JSON output mode.
func (p *Printer) SetJSON(jsonMode bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.jsonMode = jsonMode
}

// Print writes to stdout (suppressed in quiet mode).
func (p *Printer) Print(v ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.quiet {
		return
	}
	fmt.Fprint(p.stdout, v...)
}

// Println writes to stdout with newline (suppressed in quiet mode).
func (p *Printer) Println(v ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.quiet {
		return
	}
	fmt.Fprintln(p.stdout, v...)
}

// Printf writes formatted output to stdout (suppressed in quiet mode).
func (p *Printer) Printf(format string, v ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.quiet {
		return
	}
	fmt.Fprintf(p.stdout, format, v...)
}

// Error writes to stderr (never suppressed).
func (p *Printer) Error(v ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.stderr, v...)
}

// Errorf writes formatted error to stderr (never suppressed).
func (p *Printer) Errorf(format string, v ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.stderr, format, v...)
}

// JSON outputs a value as JSON (respects quiet mode).
func (p *Printer) JSON(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.quiet {
		return nil
	}
	enc := json.NewEncoder(p.stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// IsTTY returns whether stdout is a terminal.
func (p *Printer) IsTTY() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isTTY
}

// Stdout returns the stdout writer (for piping to external commands).
func (p *Printer) Stdout() io.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stdout
}

// Stderr returns the stderr writer.
func (p *Printer) Stderr() io.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stderr
}

// Global printer instance for backward compatibility.
var global = New()

// Global returns the global Printer instance.
func Global() *Printer { return global }

// SetGlobal replaces the global Printer instance (for testing).
// Returns the previous global printer so it can be restored.
func SetGlobal(p *Printer) *Printer {
	old := global
	global = p
	return old
}
