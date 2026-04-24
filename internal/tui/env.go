package tui

import (
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// Environment captures terminal capabilities the TUI adapts to at
// startup. Values are frozen for the life of the process — we don't
// react to mid-session resizes of $LANG because terminals don't
// typically change encoding under us.
type Environment struct {
	IsStdoutTTY bool
	IsStdinTTY  bool
	UTF8        bool
}

// DetectEnvironment inspects stdin/stdout and $LANG/$LC_* to determine
// what the TUI can safely render. Calls `term.IsTerminal` on both
// streams so the tea.Program doesn't spray escape codes into a pipe
// or a file redirect.
func DetectEnvironment() Environment {
	return Environment{
		IsStdoutTTY: term.IsTerminal(int(os.Stdout.Fd())),
		IsStdinTTY:  term.IsTerminal(int(os.Stdin.Fd())),
		UTF8:        detectUTF8(),
	}
}

// detectUTF8 reports whether the current locale advertises UTF-8.
// Glyphs that rely on Unicode (Braille charts, block sparklines,
// calendar cells) degrade to ASCII when this is false.
func detectUTF8() bool {
	for _, v := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if s := os.Getenv(v); s != "" {
			return strings.Contains(strings.ToLower(s), "utf-8") ||
				strings.Contains(strings.ToLower(s), "utf8")
		}
	}
	// No locale envs set. Modern macOS and Linux terminals are UTF-8 by
	// default; Windows consoles are not. Default true on unix-like systems
	// so Braille charts work out of the box without users needing to set
	// LANG.
	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "openbsd", "netbsd":
		return true
	default:
		return false
	}
}
