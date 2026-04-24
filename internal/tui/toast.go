package tui

import (
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DefaultToastTTL is how long a toast stays on screen unless the caller
// supplied an explicit TTL.
const DefaultToastTTL = 4 * time.Second

// maxVisibleToasts caps the stack so a storm of events can't push the
// main view off the screen.
const maxVisibleToasts = 4

// toast is one live notification.
type toast struct {
	ID       uint64
	Kind     ToastKind
	Text     string
	ExpireAt time.Time
}

// toastStack holds the currently-visible toasts. It's owned by the root
// model — no locking needed since Bubble Tea funnels every message
// through a single goroutine.
type toastStack struct {
	items  []toast
	nextID uint64
}

// add pushes a new toast onto the stack and returns the tea.Cmd that
// dispatches toastExpireMsg when the TTL elapses.
func (s *toastStack) add(kind ToastKind, text string, ttl time.Duration) tea.Cmd {
	if ttl <= 0 {
		ttl = DefaultToastTTL
	}
	id := atomic.AddUint64(&s.nextID, 1)
	t := toast{
		ID:       id,
		Kind:     kind,
		Text:     text,
		ExpireAt: time.Now().Add(ttl),
	}
	s.items = append(s.items, t)
	// Drop oldest if we're past the cap — keep newest so users see the
	// most recent signal.
	if overflow := len(s.items) - maxVisibleToasts; overflow > 0 {
		s.items = s.items[overflow:]
	}
	return tea.Tick(ttl, func(time.Time) tea.Msg {
		return toastExpireMsg{ID: id}
	})
}

// expire removes a toast by ID; no-op if already gone.
func (s *toastStack) expire(id uint64) {
	out := s.items[:0]
	for _, t := range s.items {
		if t.ID != id {
			out = append(out, t)
		}
	}
	s.items = out
}

// render returns the toast stack as a single multi-line string anchored
// to the caller-supplied width. Caller is responsible for placing it on
// top of the main view (lipgloss.Place / composite render).
func (s *toastStack) render(th theme, width int) string {
	if len(s.items) == 0 {
		return ""
	}
	lines := make([]string, 0, len(s.items))
	for _, t := range s.items {
		lines = append(lines, renderToast(th, t, width))
	}
	return strings.Join(lines, "\n")
}

func renderToast(th theme, t toast, width int) string {
	icon, style := toastStyle(th, t.Kind)
	body := icon + "  " + t.Text
	// Cap width to avoid pushing the main view; leave 8 cols of breathing
	// room on the right. 24 is a sane minimum before truncation.
	boxWidth := width - 8
	if boxWidth < 24 {
		boxWidth = 24
	}
	if lipgloss.Width(body) > boxWidth-4 {
		body = truncate(body, boxWidth-4)
	}
	return style.
		Width(boxWidth).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(body)
}

func toastStyle(th theme, kind ToastKind) (string, lipgloss.Style) {
	switch kind {
	case ToastSuccess:
		c := th.Positive.GetForeground()
		return "✓", lipgloss.NewStyle().Foreground(c).BorderForeground(c)
	case ToastWarning:
		c := th.Warning.GetForeground()
		return "!", lipgloss.NewStyle().Foreground(c).BorderForeground(c)
	case ToastError:
		c := th.Danger.GetForeground()
		return "✗", lipgloss.NewStyle().Foreground(c).BorderForeground(c)
	default:
		c := th.Focus.GetForeground()
		return "·", lipgloss.NewStyle().Foreground(c).BorderForeground(c)
	}
}

// requestToastCmd is a helper for callers outside the model (e.g. the
// action registry) to enqueue a toast via the message loop rather than
// mutating the stack directly.
func requestToastCmd(kind ToastKind, text string) tea.Cmd {
	return func() tea.Msg {
		return toastAddMsg{Kind: kind, Text: text}
	}
}
