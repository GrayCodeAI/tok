package visual

import (
	"fmt"
	"strings"
	"time"
)

const (
	ESC      = "\x1b"
	CSI      = ESC + "["
	Reset    = CSI + "0m"
	Bold     = CSI + "1m"
	FgRed    = CSI + "31m"
	FgGreen  = CSI + "32m"
	FgYellow = CSI + "33m"
	FgBlue   = CSI + "34m"
	FgCyan   = CSI + "36m"
)

var SpinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

var ProgressChars = []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}

type Spinner struct {
	pos    int
	frames []rune
	active bool
}

func NewSpinner() *Spinner {
	return &Spinner{
		frames: SpinnerChars,
		active: true,
	}
}

func (s *Spinner) Next() string {
	if !s.active {
		return ""
	}
	frame := s.frames[s.pos%len(s.frames)]
	s.pos++
	return fmt.Sprintf("%c%s", frame, CSI+"D")
}

func (s *Spinner) Stop() {
	s.active = false
}

type ProgressBar struct {
	width   int
	percent float64
	filled  []rune
	empty   []rune
}

func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		width:  width,
		filled: ProgressChars[4:],
		empty:  []rune{'-'},
	}
}

func (p *ProgressBar) SetPercent(percent float64) {
	p.percent = percent
}

func (p *ProgressBar) String() string {
	filledWidth := int(float64(p.width) * p.percent)
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < p.width; i++ {
		if i < filledWidth {
			b.WriteRune(p.filled[i%len(p.filled)])
		} else {
			b.WriteRune(p.empty[0])
		}
	}
	b.WriteString("]")
	b.WriteString(fmt.Sprintf(" %.0f%%", p.percent*100))
	return b.String()
}

type AnimatedText struct {
	text   string
	frames []string
	pos    int
	delay  time.Duration
}

func NewAnimatedText(text string, delayMs int) *AnimatedText {
	frames := generateFrames(text)
	return &AnimatedText{
		text:   text,
		frames: frames,
		delay:  time.Duration(delayMs) * time.Millisecond,
	}
}

func generateFrames(text string) []string {
	var frames []string
	for i := 1; i <= len(text); i++ {
		frames = append(frames, text[:i])
	}
	return frames
}

func (a *AnimatedText) Next() string {
	if a.pos >= len(a.frames) {
		return a.text
	}
	frame := a.frames[a.pos]
	a.pos++
	return frame
}

func (a *AnimatedText) Reset() {
	a.pos = 0
}

func CursorShow() string {
	return CSI + "?25h"
}

func CursorHide() string {
	return CSI + "?25l"
}

func CursorUp(n int) string {
	return fmt.Sprintf("%s%dA", CSI, n)
}

func CursorDown(n int) string {
	return fmt.Sprintf("%s%dB", CSI, n)
}

func CursorForward(n int) string {
	return fmt.Sprintf("%s%dC", CSI, n)
}

func CursorBack(n int) string {
	return fmt.Sprintf("%s%dD", CSI, n)
}

func ClearLine() string {
	return CSI + "2K"
}

func ClearScreen() string {
	return CSI + "2J"
}
