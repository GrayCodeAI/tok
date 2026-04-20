package tui

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// OSC-52 is the terminal escape sequence that asks the emulator to
// copy a payload into the system clipboard. It works over ssh, doesn't
// require a native clipboard daemon, and has zero runtime dependencies
// — the tradeoff is that not every terminal honors it (kitty, iTerm2,
// wezterm, alacritty, Windows Terminal do; stock xterm requires an
// explicit `-DISABLE_COPY_TO_CLIPBOARD` off).
//
// We expose yankCmd as a tea.Cmd so the root Update loop can print the
// sequence during the normal render cycle (printing from outside the
// loop would race with tea's terminal writer). Callers emit a toast
// alongside so the user gets visible feedback regardless of whether
// the terminal actually honored the paste.
//
// Format reference: \x1b]52;c;<base64>\x07

// YankCmd returns a tea.Cmd that emits the OSC-52 copy sequence for
// the given payload. Empty payload is a no-op. The command also
// dispatches a toast so users know something happened.
func YankCmd(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	return tea.Batch(
		writeOSC52(text),
		requestToastCmd(ToastSuccess, "yanked "+describeYank(text)),
	)
}

// writeOSC52 is the tea.Cmd that actually writes the escape sequence.
// We write to /dev/tty when we can so the sequence goes to the user's
// terminal even if stdout is captured by tea's renderer buffer.
// Falling back to stdout is safe: tea won't have switched to alt-screen
// during a key handler, so the bytes land on the same stream anyway.
func writeOSC52(payload string) tea.Cmd {
	return func() tea.Msg {
		encoded := base64.StdEncoding.EncodeToString([]byte(payload))
		seq := fmt.Sprintf("\x1b]52;c;%s\x07", encoded)

		if f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
			_, _ = f.WriteString(seq)
			_ = f.Close()
			return nil
		}
		// /dev/tty unavailable (Windows, some sandboxes) — best effort
		// via stderr. Stderr is less likely than stdout to be captured,
		// but still goes to the terminal in most cases.
		_, _ = fmt.Fprint(os.Stderr, seq)
		return nil
	}
}

// describeYank produces a short descriptor for the toast so the user
// can confirm they yanked the intended thing without the toast
// becoming huge.
func describeYank(text string) string {
	first := text
	if idx := strings.IndexAny(text, "\n\r"); idx >= 0 {
		first = text[:idx]
	}
	if len(first) > 40 {
		first = first[:40] + "…"
	}
	return first
}

// RowToTSV renders a Row as tab-separated text suitable for paste into
// a spreadsheet. Cell contents already strip styling because we build
// them from raw domain data, not rendered strings.
func RowToTSV(r Row) string {
	return strings.Join(r.Cells, "\t")
}
